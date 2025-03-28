package auth_jwt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
)

type ClientConfig struct {
	// Standard Issuer Url. Used for building the JWKS Provider
	BackendIssuerUrl string
	// Optionally provide a frontend Issuer Url. Falls back to BackendIssuerUrl if not provided.
	// This should be equivalent to what will be present in the "iss" claim of the JWT token.
	// This may be different depending auth provider or if running thorugh a reverse proxy
	FrontendIssuerUrl  *string
	ApiAudiences       []string
	SignatureAlgorithm validator.SignatureAlgorithm
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
	issuerUrl, err := url.Parse(cfg.BackendIssuerUrl)
	if err != nil {
		return nil, err
	}
	provider := jwks.NewCachingProvider(issuerUrl, 5*time.Minute)

	expectedIss := cfg.BackendIssuerUrl
	if cfg.FrontendIssuerUrl != nil {
		expectedIss = *cfg.FrontendIssuerUrl
	}

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		cfg.SignatureAlgorithm,
		expectedIss,
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
func (j *Client) validateToken(
	ctx context.Context,
	accessToken string,
) (*validator.ValidatedClaims, error) {
	rawParsedToken, err := j.jwtValidator.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, nucleuserrors.NewUnauthenticated(err.Error())
	}
	validatedClaims, ok := rawParsedToken.(*validator.ValidatedClaims)
	if !ok {
		return nil, nucleuserrors.NewInternalError(
			"unable to convert token claims what was expected",
		)
	}
	return validatedClaims, nil
}

type TokenContextKey struct{}

type TokenContextData struct {
	ParsedToken *validator.ValidatedClaims
	RawToken    string

	Claims *CustomClaims

	AuthUserId string
	Scopes     []string // Contains Scopes & Permissions
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
func (j *Client) InjectTokenCtx(
	ctx context.Context,
	header http.Header,
	spec connect.Spec,
) (context.Context, error) {
	token, err := utils.GetBearerTokenFromHeader(header, "Authorization")
	if err != nil {
		return nil, err
	}

	parsedToken, err := j.validateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.CustomClaims.(*CustomClaims)
	if !ok {
		return nil, nucleuserrors.NewInternalError(
			"unable to cast custom token claims to CustomClaims struct",
		)
	}

	scopes := getCombinedScopesAndPermissions(claims.Scope, claims.Permissions)
	userId := parsedToken.RegisteredClaims.Subject

	return SetTokenData(ctx, &TokenContextData{
		ParsedToken: parsedToken,
		RawToken:    token,

		Claims: claims,

		AuthUserId: userId,
		Scopes:     scopes,
	}), nil
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
	val := ctx.Value(TokenContextKey{})
	data, ok := val.(*TokenContextData)
	if !ok {
		return nil, nucleuserrors.NewUnauthenticated(
			fmt.Sprintf("ctx does not contain TokenContextData or unable to cast struct: %T", val),
		)
	}
	return data, nil
}

func SetTokenData(ctx context.Context, data *TokenContextData) context.Context {
	return context.WithValue(ctx, TokenContextKey{}, data)
}
