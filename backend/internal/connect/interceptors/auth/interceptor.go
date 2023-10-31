package auth_interceptor

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
)

type Interceptor struct {
	authFunc AuthFunc
}

type AuthFunc func(ctx context.Context, header http.Header) (context.Context, error)

func NewInterceptor(authFunc AuthFunc) connect.Interceptor {
	return &Interceptor{authFunc: authFunc}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		newCtx, err := i.authFunc(ctx, request.Header())
		if err != nil {
			return nil, err
		}
		return next(newCtx, request)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		newCtx, err := i.authFunc(ctx, conn.RequestHeader())
		if err != nil {
			return err
		}
		return next(newCtx, conn)
	}
}
