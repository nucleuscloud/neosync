package authmw

import (
	"context"
	"net/http"
)

type JwtClient interface {
	InjectTokenCtx(ctx context.Context, header http.Header) (context.Context, error)
}

type AuthMiddleware struct {
	jwtClient JwtClient
	// db        *nucleusdb.NucleusDb
}

func New(
	jwtClient JwtClient,
	// db *nucleusdb.NucleusDb,
) *AuthMiddleware {

	return &AuthMiddleware{jwtClient: jwtClient}
}

func (n *AuthMiddleware) ValidateAndInjectAll(ctx context.Context, header http.Header) (context.Context, error) {
	ctx, err := n.ValidateAndInjectJwtToken(ctx, header)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (n *AuthMiddleware) ValidateAndInjectJwtToken(ctx context.Context, header http.Header) (context.Context, error) {
	ctx, err := n.jwtClient.InjectTokenCtx(ctx, header)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}
