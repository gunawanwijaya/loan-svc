package pkg

import (
	"context"
	"log/slog"
)

var Context interface {
	PutSlogLogger(ctx context.Context, logger *slog.Logger) context.Context
	SlogLogger(ctx context.Context) *slog.Logger
} = struct {
	ctxKeySlogLogger
}{}

type ctxKeySlogLogger struct{}

func (key ctxKeySlogLogger) PutSlogLogger(ctx context.Context, val *slog.Logger) context.Context {
	if val == nil {
		panic("ctxKeySlogLogger")
		// return ctx
	}
	return context.WithValue(ctx, key, val)
}
func (key ctxKeySlogLogger) SlogLogger(ctx context.Context) *slog.Logger {
	v, _ := ctx.Value(key).(*slog.Logger)
	return v
}
