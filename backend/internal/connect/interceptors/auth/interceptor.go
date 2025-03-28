package auth_interceptor

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
)

type Interceptor struct {
	authFunc           AuthFunc
	excludedProcedures map[string]struct{}
}

type AuthFunc func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error)

func NewInterceptor(authFunc AuthFunc) connect.Interceptor {
	return &Interceptor{authFunc: authFunc}
}

func NewInterceptorWithExclude(authFunc AuthFunc, excludedProcedures []string) connect.Interceptor {
	excludedMap := map[string]struct{}{}
	for _, proc := range excludedProcedures {
		excludedMap[proc] = struct{}{}
	}

	return &Interceptor{authFunc: authFunc, excludedProcedures: excludedMap}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if _, ok := i.excludedProcedures[request.Spec().Procedure]; ok {
			return next(ctx, request)
		}
		newCtx, err := i.authFunc(ctx, request.Header(), request.Spec())
		if err != nil {
			return nil, err
		}
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
		if _, ok := i.excludedProcedures[conn.Spec().Procedure]; ok {
			return next(ctx, conn)
		}
		newCtx, err := i.authFunc(ctx, conn.RequestHeader(), conn.Spec())
		if err != nil {
			return err
		}
		return next(newCtx, conn)
	}
}
