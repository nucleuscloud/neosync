package authlogging_interceptor

import (
	"context"

	"connectrpc.com/connect"
	"github.com/nucleuscloud/neosync/backend/internal/auth/tokenctx"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

type Interceptor struct {
	db *neosyncdb.NeosyncDb
}

func NewInterceptor(db *neosyncdb.NeosyncDb) connect.Interceptor {
	return &Interceptor{db: db}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		return next(setAuthValues(ctx, i.db), request)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(setAuthValues(ctx, i.db), conn)
	}
}

func setAuthValues(ctx context.Context, db *neosyncdb.NeosyncDb) context.Context {
	vals := getAuthValues(ctx, db)
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx).With(vals...)
	return logger_interceptor.SetLoggerContext(ctx, logger)
}

func getAuthValues(ctx context.Context, db *neosyncdb.NeosyncDb) []any {
	tokenCtxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return []any{}
	}
	output := []any{}

	if tokenCtxResp.JwtContextData != nil {
		output = append(output, "authUserId", tokenCtxResp.JwtContextData.AuthUserId)

		user, err := db.Q.GetUserByProviderSub(ctx, db.Db, tokenCtxResp.JwtContextData.AuthUserId)
		if err == nil {
			output = append(output, "userId", neosyncdb.UUIDString(user.ID))
		}
	} else if tokenCtxResp.ApiKeyContextData != nil {
		output = append(output, "apiKeyType", tokenCtxResp.ApiKeyContextData.ApiKeyType)
		if tokenCtxResp.ApiKeyContextData.ApiKey != nil {
			output = append(output,
				"apiKeyId", neosyncdb.UUIDString(tokenCtxResp.ApiKeyContextData.ApiKey.ID),
				"accountId", neosyncdb.UUIDString(tokenCtxResp.ApiKeyContextData.ApiKey.AccountID),
				"userId", neosyncdb.UUIDString(tokenCtxResp.ApiKeyContextData.ApiKey.UserID),
			)
		}
	}
	return output
}
