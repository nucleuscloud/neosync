package clientcredtokenprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
)

type tokenProvider interface {
	GetToken(context.Context) (*auth_client.AuthTokenResponse, error)
}

type tokenProviderClient struct {
	tokenurl     string
	clientId     string
	clientSecret string
}

func (c *tokenProviderClient) GetToken(
	ctx context.Context,
) (*auth_client.AuthTokenResponse, error) {
	values := url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{c.clientId},
		"client_secret": []string{c.clientSecret},
	}
	payload := strings.NewReader(
		values.Encode(),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenurl, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to request oauth authorization code: %w", err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfill authorization code request: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to ready body from authorization code response: %w", err)
	}

	var tokenResponse *auth_client.AuthTokenResponseData
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		return nil, fmt.Errorf("unable to decode token response data from body: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		var errorResponse auth_client.AuthTokenErrorData
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, fmt.Errorf("unable to decode error response data from body: %w", err)
		}
		return &auth_client.AuthTokenResponse{
			Result: nil,
			Error:  &errorResponse,
		}, nil
	}
	return &auth_client.AuthTokenResponse{
		Result: tokenResponse,
		Error:  nil,
	}, nil
}

// Caches the access token on the struct. Retrieves a new one once it has expired, within the token exp buffer limit
type ClientCredentialsTokenProvider struct {
	tokenprovider  tokenProvider
	tokenExpBuffer time.Duration
	logger         *slog.Logger

	rw          sync.RWMutex
	accessToken *string
	expiresAt   *time.Time
}

func New(
	tokenurl, clientId, clientSecret string,
	tokenExpirationBuffer time.Duration,
	logger *slog.Logger,
) *ClientCredentialsTokenProvider {
	return &ClientCredentialsTokenProvider{
		tokenprovider: &tokenProviderClient{
			tokenurl:     tokenurl,
			clientId:     clientId,
			clientSecret: clientSecret,
		},
		tokenExpBuffer: tokenExpirationBuffer,
		logger:         logger,
	}
}

// Returns an always valid access token
func (c *ClientCredentialsTokenProvider) GetToken(ctx context.Context) (string, error) {
	c.rw.RLock()
	at := c.accessToken
	expiresAt := c.expiresAt
	c.rw.RUnlock()
	if isTokenValid(at, expiresAt, c.tokenExpBuffer) {
		return *at, nil
	}

	c.rw.Lock()
	defer c.rw.Unlock()
	at = c.accessToken
	expiresAt = c.expiresAt
	if isTokenValid(at, expiresAt, c.tokenExpBuffer) {
		return *at, nil
	}
	c.accessToken = nil
	c.expiresAt = nil

	tokenResp, err := c.tokenprovider.GetToken(ctx)
	if err != nil {
		return "", err
	}
	if tokenResp.Error != nil {
		return "", fmt.Errorf("%s: %s", tokenResp.Error.Error, tokenResp.Error.ErrorDescription)
	}
	c.accessToken = &tokenResp.Result.AccessToken
	newExpiresAt := time.Now().Add(time.Duration(tokenResp.Result.ExpiresIn) * time.Second)
	c.expiresAt = &newExpiresAt
	return *c.accessToken, nil
}

func isTokenValid(token *string, expiresAt *time.Time, expBuffer time.Duration) bool {
	return token != nil && expiresAt != nil && time.Now().Add(expBuffer).Before(*expiresAt)
}
