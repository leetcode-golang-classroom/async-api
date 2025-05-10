package logger

import (
	"context"
	"log/slog"
	"os"
)

// CtxKey - struct.
type CtxKey struct{}

// CtxWithLogger - create ctx with logger.
func CtxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	if ctxLog, ok := ctx.Value(CtxKey{}).(*slog.Logger); ok && ctxLog == logger {
		return ctx
	}
	return context.WithValue(ctx, CtxKey{}, logger)
}

// FromContext - get slog from ctx.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(CtxKey{}).(*slog.Logger); ok {
		return logger
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
}
