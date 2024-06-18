package auth_interceptor

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
)

type Interceptor struct {
	isEnabled  bool
	authHeader string
	getToken   GetTokenFunc
}

type GetTokenFunc func(context.Context) (string, error)

func NewInterceptor(isEnabled bool, authHeader string, getToken GetTokenFunc) connect.Interceptor {
	return &Interceptor{isEnabled: isEnabled, authHeader: authHeader, getToken: getToken}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if i.isEnabled {
			token, err := i.getToken(ctx)
			if err != nil {
				return nil, err
			}
			request.Header().Set(i.authHeader, token)
		}
		return next(ctx, request)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		if i.isEnabled {
			token, err := i.getToken(ctx)
			if err != nil {
				err2 := conn.CloseRequest()
				if err2 != nil {
					slog.Info(err2.Error()) // todo
				}
				return conn
			}
			conn.RequestHeader().Set(i.authHeader, token)
		}
		return conn
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if i.isEnabled {
			if conn.RequestHeader().Get(i.authHeader) == "" {
				return connect.NewError(connect.CodeUnauthenticated, errors.New("no token"))
			}
		}
		return next(ctx, conn)
	}
}
