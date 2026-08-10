package redisclient

import (
	"context"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const scopeName = "github.com/koliader/tellmi-posts/internal/store/redis"

const (
	attrDBSystem    = "db.system"
	attrDBOperation = "db.operation.name"
	attrDBIndex     = "db.redis.database_index"
	attrServerAddr  = "server.address"
	attrServerPort  = "server.port"
)

// otelHook instruments every redis command with an OTel span and a duration
// histogram, using the global tracer/meter providers installed by otel.Init.
type otelHook struct {
	tracer  trace.Tracer
	hist    metric.Float64Histogram
	dbIndex int
	addr    string
}

var _ redis.Hook = (*otelHook)(nil)

func newOTelHook(opts *redis.Options) *otelHook {
	tracer := otel.Tracer(scopeName)
	meter := otel.Meter(scopeName)
	hist, err := meter.Float64Histogram(
		"redis.command.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of redis commands"),
	)
	if err != nil {
		hist = nil
	}
	return &otelHook{
		tracer:  tracer,
		hist:    hist,
		dbIndex: opts.DB,
		addr:    opts.Addr,
	}
}

func (h *otelHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *otelHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		return h.trace(ctx, cmd.Name(), func(c context.Context) error {
			return next(c, cmd)
		})
	}
}

func (h *otelHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return h.trace(ctx, "PIPELINE", func(c context.Context) error {
			return next(c, cmds)
		})
	}
}

func (h *otelHook) trace(ctx context.Context, operation string, fn func(context.Context) error) error {
	host, port := splitAddr(h.addr)
	start := time.Now()

	ctx, span := h.tracer.Start(ctx, "redis.op",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String(attrDBSystem, "redis"),
			attribute.String(attrDBOperation, operation),
			attribute.Int(attrDBIndex, h.dbIndex),
			attribute.String(attrServerAddr, host),
			attribute.String(attrServerPort, port),
		),
	)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	if h.hist != nil {
		h.hist.Record(ctx, time.Since(start).Seconds(),
			metric.WithAttributes(attribute.String(attrDBOperation, operation)))
	}

	return err
}

func splitAddr(addr string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "localhost", "6379"
	}
	return host, port
}
