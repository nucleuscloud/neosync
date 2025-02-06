package tokenctx

import (
	"context"

	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
)

type TokenCtxResponse struct {
	JwtContextData    *auth_jwt.TokenContextData
	ApiKeyContextData *auth_apikey.TokenContextData
}

// Attempts to get all token contexts that are available. If none, then returns unauth error
func GetTokenCtx(ctx context.Context) (*TokenCtxResponse, error) {
	apikeyData, err := auth_apikey.GetTokenDataFromCtx(ctx)
	if err != nil {
		jwtData, err := auth_jwt.GetTokenDataFromCtx(ctx)
		if err != nil {
			return nil, nucleuserrors.NewUnauthenticated("unable to find any token data in context")
		}
		return &TokenCtxResponse{JwtContextData: jwtData}, nil
	}
	return &TokenCtxResponse{ApiKeyContextData: apikeyData}, nil
}
