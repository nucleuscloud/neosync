package auth_apikey

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	pkg_utils "github.com/nucleuscloud/neosync/backend/pkg/utils"
)

type TokenContextKey struct{}
type TokenContextData struct {
	RawToken   string
	ApiKey     *db_queries.NeosyncApiAccountApiKey
	ApiKeyType apikey.ApiKeyType
}

var (
	InvalidApiKeyErr = errors.New("token is not a valid neosync api key")
	ApiKeyExpiredErr = nucleuserrors.NewUnauthenticated("token is expired")
)

type Queries interface {
	GetAccountApiKeyByKeyValue(ctx context.Context, db db_queries.DBTX, apiKey string) (db_queries.NeosyncApiAccountApiKey, error)
}

type Client struct {
	q                       Queries
	db                      db_queries.DBTX
	allowedWorkerApiKeys    []string
	allowedWorkerProcedures map[string]any
}

func New(
	queries Queries,
	db db_queries.DBTX,
	allowedWorkerApiKeys []string,
	allowedWorkerProcedures []string,
) *Client {
	allowedWorkerProcedureSet := map[string]any{}
	for _, procedure := range allowedWorkerProcedures {
		allowedWorkerProcedureSet[procedure] = struct{}{}
	}
	return &Client{q: queries, db: db, allowedWorkerApiKeys: allowedWorkerApiKeys, allowedWorkerProcedures: allowedWorkerProcedureSet}
}

func (c *Client) InjectTokenCtx(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
	token, err := utils.GetBearerTokenFromHeader(header, "Authorization")
	if err != nil {
		return nil, err
	}

	if apikey.IsValidV1AccountKey(token) {
		hashedKeyValue := pkg_utils.ToSha256(
			token,
		)
		apiKey, err := c.q.GetAccountApiKeyByKeyValue(ctx, c.db, hashedKeyValue)
		if err != nil {
			return nil, err
		}

		if time.Now().After(apiKey.ExpiresAt.Time) {
			return nil, ApiKeyExpiredErr
		}

		return SetTokenData(ctx, &TokenContextData{
			RawToken:   token,
			ApiKey:     &apiKey,
			ApiKeyType: apikey.AccountApiKey,
		}), nil
	} else if apikey.IsValidV1WorkerKey(token) &&
		isApiKeyAllowed(c.allowedWorkerApiKeys, token) &&
		isProcedureAllowed(c.allowedWorkerProcedures, spec.Procedure) {
		return SetTokenData(ctx, &TokenContextData{
			RawToken:   token,
			ApiKey:     nil,
			ApiKeyType: apikey.WorkerApiKey,
		}), nil
	}
	return nil, InvalidApiKeyErr
}

func GetTokenDataFromCtx(ctx context.Context) (*TokenContextData, error) {
	data, ok := ctx.Value(TokenContextKey{}).(*TokenContextData)
	if !ok {
		return nil, nucleuserrors.NewUnauthenticated("ctx does not contain TokenContextData or unable to cast struct")
	}
	return data, nil
}

func SetTokenData(ctx context.Context, data *TokenContextData) context.Context {
	return context.WithValue(ctx, TokenContextKey{}, data)
}

func isApiKeyAllowed(allowedKeys []string, key string) bool {
	for _, allowedKey := range allowedKeys {
		if secureCompare(allowedKey, key) {
			return true
		}
	}
	return false
}

func isProcedureAllowed(allowedProcedures map[string]any, procedure string) bool {
	_, ok := allowedProcedures[procedure]
	return ok
}

func secureCompare(a, b string) bool {
	// Convert strings to byte slices for comparison
	aBytes := []byte(a)
	bBytes := []byte(b)

	// Check length first; if they differ, return false immediately
	if len(aBytes) != len(bBytes) {
		return false
	}

	// Use ConstantTimeCompare for a timing-attack resistant comparison
	return subtle.ConstantTimeCompare(aBytes, bBytes) == 1
}
