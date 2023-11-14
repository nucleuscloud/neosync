package authmw

import (
	"context"
	"errors"
	"net/http"

	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
)

type AuthClient interface {
	InjectTokenCtx(ctx context.Context, header http.Header) (context.Context, error)
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

func (n *AuthMiddleware) InjectTokenCtx(ctx context.Context, header http.Header) (context.Context, error) {
	ctx, err := n.apiKeyClient.InjectTokenCtx(ctx, header)
	if err != nil && !errors.Is(err, auth_apikey.InvalidApiKeyErr) {
		return nil, err
	} else if err != nil && errors.Is(err, auth_apikey.InvalidApiKeyErr) {
		return n.jwtClient.InjectTokenCtx(ctx, header)
	}
	return ctx, nil
}
