package logger_interceptor

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
)

type Interceptor struct {
	logger *slog.Logger
}

func NewInterceptor(logger *slog.Logger) connect.Interceptor {
	return &Interceptor{
		logger: logger,
	}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		newCtx := SetLoggerContext(ctx, clonelogger(i.logger))
		return next(newCtx, request)
	}
}

func (i *Interceptor) WrapStreamingClient(
	next connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *Interceptor) WrapStreamingHandler(
	next connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		newCtx := SetLoggerContext(ctx, clonelogger(i.logger))
		return next(newCtx, conn)
	}
}

func clonelogger(logger *slog.Logger) *slog.Logger {
	c := *logger
	return &c
}
