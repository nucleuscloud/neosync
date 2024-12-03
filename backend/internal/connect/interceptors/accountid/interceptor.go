package accountid_interceptor

import (
	"context"

	"connectrpc.com/connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
)

type Interceptor struct{}

var _ connect.Interceptor = (*Interceptor)(nil)

func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

type AccountIdProvider interface {
	GetAccountId() string
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if provider, ok := request.Any().(AccountIdProvider); ok {
			if accountId := provider.GetAccountId(); accountId != "" {
				logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
				ctx = logger_interceptor.SetLoggerContext(
					ctx,
					logger.With("accountId", accountId),
				)
			}
		}
		return next(ctx, request)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	}
}
