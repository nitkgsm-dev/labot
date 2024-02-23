package logging

import (
	"context"
	"log/slog"
)

type contextKey string

const slogKey contextKey = "slog"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, slogKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(slogKey).(*slog.Logger); ok {
		return v
	}
	return DefaultBuilder().Build()
}
