package authmw

import (
	"context"
	"net/http"

	authjwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	nucleusdb "github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type NucleusAuthMiddleware struct {
	jwtClient *authjwt.JwtClient
	db        *nucleusdb.NucleusDb
}

func New(
	jwtClient *authjwt.JwtClient,
	db *nucleusdb.NucleusDb,
) *NucleusAuthMiddleware {

	return &NucleusAuthMiddleware{jwtClient: jwtClient, db: db}
}

func (n *NucleusAuthMiddleware) ValidateAndInjectAll(ctx context.Context, header http.Header) (context.Context, error) {
	// ctx, err := n.ValidateAndInjectToken(ctx, header)
	// if err != nil {
	// 	return nil, err
	// }
	// ctx, err = n.ValidateAndInjectAccount(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	return ctx, nil
}

func (n *NucleusAuthMiddleware) ValidateAndInjectToken(ctx context.Context, header http.Header) (context.Context, error) {
	// ctx, err := n.jwtClient.InjectTokenCtx(ctx, header)
	// if err != nil {
	// 	return nil, err
	// }
	return ctx, nil
}

// func (n *NucleusAuthMiddleware) ValidateAndInjectAccount(ctx context.Context) (context.Context, error) {
// 	ctx, err := authvalidate.InjectAuthAccountCtx(ctx, n.db)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ctx, nil
// }
