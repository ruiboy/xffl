package postgres

import (
	"context"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
)

type queryCountKey struct{}

// WithQueryCounter injects a fresh query counter into ctx.
func WithQueryCounter(ctx context.Context) context.Context {
	return context.WithValue(ctx, queryCountKey{}, new(atomic.Int64))
}

// QueryCount returns the number of DB queries fired against this ctx.
func QueryCount(ctx context.Context) int64 {
	if c, ok := ctx.Value(queryCountKey{}).(*atomic.Int64); ok {
		return c.Load()
	}
	return 0
}

// QueryCountTracer implements pgx.QueryTracer, incrementing the per-request counter.
type QueryCountTracer struct{}

func (QueryCountTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	if c, ok := ctx.Value(queryCountKey{}).(*atomic.Int64); ok {
		c.Add(1)
	}
	return ctx
}

func (QueryCountTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}
