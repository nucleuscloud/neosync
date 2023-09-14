package auth_jwt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

type ClientConfig struct {
	BaseUrl      string
	ApiAudiences []string
}

type Client struct {
	jwtValidator *validator.Validator
}

func New(
	cfg *ClientConfig,
) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("must provide jwt client cfg")
	}
	issuerUrl, err := url.Parse(cfg.BaseUrl + "/")
	if err != nil {
		return nil, err
	}

	provider := jwks.NewCachingProvider(issuerUrl, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerUrl.String(),
		cfg.ApiAudiences,
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		jwtValidator: jwtValidator,
	}, nil
}

// Validates and returns a parsed access token (if available)
func (j *Client) validateToken(ctx context.Context, accessToken string) (*validator.ValidatedClaims, error) {
	rawParsedToken, err := j.jwtValidator.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, nucleuserrors.NewUnauthenticated("token was not valid")
	}
	return rawParsedToken.(*validator.ValidatedClaims), nil
}

type tokenContextKey struct{}

type TokenContextData struct {
	ParsedToken *validator.ValidatedClaims
	RawToken    string

	Claims *CustomClaims

	AuthUserId       string
	Scopes           []string // Contains Scopes & Permissions
	IsServiceAccount bool
}

func (t *TokenContextData) HasScope(scope string) bool {
	return hasScope(t.Scopes, scope)
}

func hasScope(scopes []string, expectedScope string) bool {
	for _, scope := range scopes {
		if expectedScope == scope {
			return true
		}
	}
	return false
}

// Validates the ctx is authenticated. Stuffs the parsed token onto the context
func (j *Client) InjectTokenCtx(ctx context.Context, header http.Header) (context.Context, error) {
	unparsedToken := header.Get("Authorization")
	if unparsedToken == "" {
		return nil, nucleuserrors.NewUnauthenticated("must provide valid bearer token")
	}
	pieces := strings.Split(unparsedToken, " ")
	if len(pieces) != 2 {
		return nil, nucleuserrors.NewUnauthenticated("token not in proper format")
	}
	if pieces[0] != "Bearer" {
		return nil, nucleuserrors.NewUnauthenticated("must provided bearer token")
	}
	token := pieces[1]

	parsedToken, err := j.validateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.CustomClaims.(*CustomClaims)
	if !ok {
		return nil, nucleuserrors.NewInternalError("unable to cast custom token claims to CustomClaims struct")
	}

	scopes := getCombinedScopesAndPermissions(claims.Scope, claims.Permissions)
	userId := parsedToken.RegisteredClaims.Subject

	newCtx := context.WithValue(ctx, tokenContextKey{}, &TokenContextData{
		ParsedToken: parsedToken,
		RawToken:    token,

		Claims: claims,

		AuthUserId:       userId,
		Scopes:           scopes,
		IsServiceAccount: false,
	})

	return newCtx, nil
}

func getCombinedScopesAndPermissions(scope string, permissions []string) []string {
	scopes := strings.Split(scope, " ")

	scopeSet := map[string]struct{}{}
	for _, scope := range scopes {
		scopeSet[scope] = struct{}{}
	}
	for _, perm := range permissions {
		scopeSet[perm] = struct{}{}
	}

	scopesAndPerms := []string{}
	for scope := range scopeSet {
		scopesAndPerms = append(scopesAndPerms, scope)
	}
	return scopesAndPerms
}

func GetTokenDataFromCtx(ctx context.Context) (*TokenContextData, error) {
	data, ok := ctx.Value(tokenContextKey{}).(*TokenContextData)
	if !ok {
		return nil, nucleuserrors.NewUnauthenticated("ctx does not contain TokenContextData or unable to cast struct")
	}
	return data, nil
}
