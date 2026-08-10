package redisclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	// reconnectBackoffInitial is how long we wait before the first retry when
	// redis is unreachable; it doubles until reconnectBackoffMax, mirroring the
	// postgres (pgxpool) and rabbitmq reconnection behaviour in this service.
	reconnectBackoffInitial = 500 * time.Millisecond
	reconnectBackoffMax     = 10 * time.Second
	// healthCheckInterval is how often a healthy connection is probed so a
	// mid-run outage is noticed without pinging redis constantly.
	healthCheckInterval = 30 * time.Second
)

type Client struct {
	client    *redis.Client
	stopped   chan struct{}
	closeOnce sync.Once
}

type Config struct {
	Addr     string
	Password string
	DB       int
}

// redisLogger routes go-redis's internal diagnostics (e.g. "connection pool:
// failed to dial") into zerolog instead of the standard log package, so they
// follow the service's structured, colored logging. It logs at debug level:
// per-command dial failures during an outage are already surfaced (and
// rate-limited to state transitions) by reconnectLoop, so an unhealthy redis
// never floods the console.
type redisLogger struct{}

func (redisLogger) Printf(_ context.Context, format string, v ...any) {
	log.Debug().Msgf(format, v...)
}

// New creates a redis client without requiring a live connection, matching how
// postgres (pgxpool) is handled: the service never fails to start because redis
// is down. go-redis dials lazily per command and reconnects on its own, and a
// background loop keeps probing redis with exponential backoff so an outage is
// logged and recovery is reported.
func New(ctx context.Context, cfg Config) *Client {
	redis.SetLogger(redisLogger{})

	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(opts)
	client.AddHook(newOTelHook(opts))

	c := &Client{
		client:  client,
		stopped: make(chan struct{}),
	}
	go c.reconnectLoop(ctx)

	return c
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		close(c.stopped)
	})
	return c.client.Close()
}

// reconnectLoop probes redis with a Ping and reconnects (transparently, via the
// client) using exponential backoff when it is unreachable, so a redis outage
// never takes the service down and recovery is logged.
func (c *Client) reconnectLoop(ctx context.Context) {
	backoff := reconnectBackoffInitial
	connected := true

	for {
		err := c.client.Ping(ctx).Err()
		if err != nil {
			if connected {
				log.Warn().Err(err).Msg("redis: connection lost, attempting reconnect")
				connected = false
			}
			select {
			case <-ctx.Done():
				return
			case <-c.stopped:
				return
			case <-time.After(backoff):
			}
			if backoff < reconnectBackoffMax {
				backoff *= 2
			}
			continue
		}

		if !connected {
			log.Info().Msg("redis: reconnected")
		}
		connected = true
		backoff = reconnectBackoffInitial

		select {
		case <-ctx.Done():
			return
		case <-c.stopped:
			return
		case <-time.After(healthCheckInterval):
		}
	}
}

func (c *Client) Set(
	ctx context.Context,
	key string,
	value any,
) error {
	return c.client.Set(ctx, key, value, 0).Err()
}

func (c *Client) Get(
	ctx context.Context,
	key string,
) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Delete(
	ctx context.Context,
	key string,
) error {
	return c.client.Del(ctx, key).Err()
}

// LPush prepends values to the head of the list at key (atomic).
func (c *Client) LPush(
	ctx context.Context,
	key string,
	values ...string,
) error {
	return c.client.LPush(ctx, key, values).Err()
}

// LTrim keeps only the list elements from start to stop inclusive (atomic).
func (c *Client) LTrim(
	ctx context.Context,
	key string,
	start int64,
	stop int64,
) error {
	return c.client.LTrim(ctx, key, start, stop).Err()
}

// LRange returns the list elements from start to stop inclusive.
func (c *Client) LRange(
	ctx context.Context,
	key string,
	start int64,
	stop int64,
) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// LLen returns the length of the list at key (0 when the key is absent).
func (c *Client) LLen(
	ctx context.Context,
	key string,
) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// Pipeline starts a transaction pipeline so a batch of commands executes
// atomically via MULTI/EXEC.
func (c *Client) Pipeline(ctx context.Context) redis.Pipeliner {
	return c.client.TxPipeline()
}

func (c *Client) DeleteByPattern(
	ctx context.Context,
	pattern string,
) error {
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(
			ctx,
			cursor,
			pattern,
			100, // scan up to ~100 keys per iteration
		).Result()
		if err != nil {
			return fmt.Errorf("scan redis keys: %w", err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("delete redis keys: %w", err)
			}
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return nil
}
