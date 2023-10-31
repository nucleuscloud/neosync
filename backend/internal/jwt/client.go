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

type JwtValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (any, error)
}

type Client struct {
	jwtValidator JwtValidator
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
	validatedClaims, ok := rawParsedToken.(*validator.ValidatedClaims)
	if !ok {
		return nil, nucleuserrors.NewInternalError("unable to convert token claims what was expected")
	}
	return validatedClaims, nil
}

type TokenContextKey struct{}

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

func getBearerTokenFromHeader(
	header http.Header,
	key string,
) (string, error) {
	unparsedToken := header.Get(key)
	if unparsedToken == "" {
		return "", nucleuserrors.NewUnauthenticated("must provide valid bearer token")
	}
	pieces := strings.Split(unparsedToken, " ")
	if len(pieces) != 2 {
		return "", nucleuserrors.NewUnauthenticated("token not in proper format")
	}
	if pieces[0] != "Bearer" {
		return "", nucleuserrors.NewUnauthenticated("must provided bearer token")
	}
	token := pieces[1]
	return token, nil
}

// Validates the ctx is authenticated. Stuffs the parsed token onto the context
func (j *Client) InjectTokenCtx(ctx context.Context, header http.Header) (context.Context, error) {
	token, err := getBearerTokenFromHeader(header, "Authorization")
	if err != nil {
		return nil, err
	}

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

	newCtx := context.WithValue(ctx, TokenContextKey{}, &TokenContextData{
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
		if _, ok := scopeSet[perm]; !ok {
			scopes = append(scopes, perm)
			scopeSet[perm] = struct{}{}
		}
	}
	return scopes
}

func GetTokenDataFromCtx(ctx context.Context) (*TokenContextData, error) {
	data, ok := ctx.Value(TokenContextKey{}).(*TokenContextData)
	if !ok {
		return nil, nucleuserrors.NewUnauthenticated("ctx does not contain TokenContextData or unable to cast struct")
	}
	return data, nil
}
