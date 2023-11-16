package auth_apikey

import (
	"context"
	"errors"
	"net/http"
	"time"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
)

type TokenContextKey struct{}
type TokenContextData struct {
	RawToken string
	ApiKey   *db_queries.NeosyncApiAccountApiKey
}

var (
	InvalidApiKeyErr = errors.New("token is not a valid neosync api key")
	ApiKeyExpiredErr = nucleuserrors.NewUnauthenticated("token is expired")
)

type Queries interface {
	GetAccountApiKeyByKeyValue(ctx context.Context, db db_queries.DBTX, apiKey string) (db_queries.NeosyncApiAccountApiKey, error)
}

type Client struct {
	q  Queries
	db db_queries.DBTX
}

func New(
	queries Queries,
	db db_queries.DBTX,
) *Client {
	return &Client{q: queries, db: db}
}

func (c *Client) InjectTokenCtx(ctx context.Context, header http.Header) (context.Context, error) {
	token, err := utils.GetBearerTokenFromHeader(header, "Authorization")
	if err != nil {
		return nil, err
	}
	if !apikey.IsValidV1AccountKey(token) {
		return nil, InvalidApiKeyErr
	}

	apiKey, err := c.q.GetAccountApiKeyByKeyValue(ctx, c.db, token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(apiKey.ExpiresAt.Time) {
		return nil, ApiKeyExpiredErr
	}

	newctx := context.WithValue(ctx, TokenContextKey{}, &TokenContextData{
		RawToken: token,
		ApiKey:   &apiKey,
	})
	return newctx, err
}

func GetTokenDataFromCtx(ctx context.Context) (*TokenContextData, error) {
	data, ok := ctx.Value(TokenContextKey{}).(*TokenContextData)
	if !ok {
		return nil, nucleuserrors.NewUnauthenticated("ctx does not contain TokenContextData or unable to cast struct")
	}

	return data, nil
}
