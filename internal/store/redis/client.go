package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

type Config struct {
	Addr     string
	Password string
	DB       int
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(opts)
	client.AddHook(newOTelHook(opts))

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
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
