package authlogging_interceptor

import (
	"context"

	"connectrpc.com/connect"
	"github.com/nucleuscloud/neosync/backend/internal/auth/tokenctx"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type Interceptor struct {
	db *nucleusdb.NucleusDb
}

func NewInterceptor(db *nucleusdb.NucleusDb) connect.Interceptor {
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

func setAuthValues(ctx context.Context, db *nucleusdb.NucleusDb) context.Context {
	vals := getAuthValues(ctx, db)
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx).With(vals...)
	return logger_interceptor.SetLoggerContext(ctx, logger)
}

func getAuthValues(ctx context.Context, db *nucleusdb.NucleusDb) []any {
	tokenCtxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return []any{}
	}
	output := []any{}

	if tokenCtxResp.JwtContextData != nil {
		output = append(output, "authUserId", tokenCtxResp.JwtContextData.AuthUserId)

		user, err := db.Q.GetUserByProviderSub(ctx, db.Db, tokenCtxResp.JwtContextData.AuthUserId)
		if err == nil {
			output = append(output, "userId", nucleusdb.UUIDString(user.ID))
		}
	} else if tokenCtxResp.ApiKeyContextData != nil {
		output = append(output, "apiKeyType", tokenCtxResp.ApiKeyContextData.ApiKeyType)
		if tokenCtxResp.ApiKeyContextData.ApiKey != nil {
			output = append(output,
				"apiKeyId", nucleusdb.UUIDString(tokenCtxResp.ApiKeyContextData.ApiKey.ID),
				"accountId", nucleusdb.UUIDString(tokenCtxResp.ApiKeyContextData.ApiKey.AccountID),
				"userId", nucleusdb.UUIDString(tokenCtxResp.ApiKeyContextData.ApiKey.UserID),
			)
		}
	}
	return output
}
