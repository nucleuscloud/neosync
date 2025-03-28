package authmw

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
)

type AuthClient interface {
	InjectTokenCtx(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error)
}

type AuthMiddleware struct {
	jwtClient    AuthClient
	apiKeyClient AuthClient
}

func New(
	jwtClient AuthClient,
	apiKeyClient AuthClient,
) *AuthMiddleware {
	return &AuthMiddleware{jwtClient: jwtClient, apiKeyClient: apiKeyClient}
}

func (n *AuthMiddleware) InjectTokenCtx(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
	apiKeyCtx, err := n.apiKeyClient.InjectTokenCtx(ctx, header, spec)
	if err != nil && !errors.Is(err, auth_apikey.ErrInvalidApiKey) {
		return nil, err
	} else if err != nil && errors.Is(err, auth_apikey.ErrInvalidApiKey) {
		return n.jwtClient.InjectTokenCtx(ctx, header, spec)
	}
	return apiKeyCtx, nil
}
