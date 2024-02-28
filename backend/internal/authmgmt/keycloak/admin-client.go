package keycloak

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
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
)

var _ authmgmt.Interface = &AdminClient{}

type AdminClient struct {
	domain        string
	tokenprovider TokenProvider
	logger        *slog.Logger
}

// there is more here but leaving it out as it's not relevant right now
type user struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Enabled       bool   `json:"enabled"`
	Firstname     string `json:"firstName"`
	Lastname      string `json:"lastName"`
}

func (u *user) Name() string {
	return fmt.Sprintf("%s %s", u.Firstname, u.Lastname)
}

type TokenProvider interface {
	GetToken(context.Context) (string, error)
}

type ClientCredentialsTokenProvider struct {
	tokenurl       string
	clientId       string
	clientSecret   string
	tokenExpBuffer time.Duration
	logger         *slog.Logger

	rw          *sync.RWMutex
	accessToken *string
	expiresAt   *time.Time
}

func NewCCTokenProvider(tokenurl, clientId, clientSecret string, tokenExpirationBuffer time.Duration, logger *slog.Logger) *ClientCredentialsTokenProvider {
	return &ClientCredentialsTokenProvider{
		tokenurl:       tokenurl,
		clientId:       clientId,
		clientSecret:   clientSecret,
		tokenExpBuffer: tokenExpirationBuffer,
		logger:         logger,
		rw:             &sync.RWMutex{},
	}
}

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

	tokenResp, err := c.getToken(ctx, c.clientId, c.clientSecret)
	if err != nil {
		return "", err
	}
	if tokenResp.Error != nil {
		return "", fmt.Errorf("%s: %s", tokenResp.Error.Error, tokenResp.Error.ErrorDescription)
	}
	c.accessToken = &tokenResp.Result.AccessToken
	newExpiresAt := time.Now().Add(time.Duration(tokenResp.Result.ExpiresIn))
	c.expiresAt = &newExpiresAt
	return *c.accessToken, nil
}

func isTokenValid(token *string, expiresAt *time.Time, expBuffer time.Duration) bool {
	return token != nil && expiresAt != nil && time.Now().Add(-expBuffer).Before(*expiresAt)
}

func (c *ClientCredentialsTokenProvider) getToken(ctx context.Context, clientId, clientSecret string) (*auth_client.AuthTokenResponse, error) {
	values := url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{clientId},
		"client_secret": []string{clientSecret},
	}
	payload := strings.NewReader(
		values.Encode(),
	)
	req, err := http.NewRequestWithContext(ctx, "POST", c.tokenurl, payload)
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

func New(
	domain string,
	provider TokenProvider,
	logger *slog.Logger,
) (*AdminClient, error) {
	return &AdminClient{domain: domain, tokenprovider: provider, logger: logger}, nil
}

func (c *AdminClient) GetUserBySub(ctx context.Context, sub string) (*authmgmt.User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/users/%s", c.domain, sub), http.NoBody)
	if err != nil {
		return nil, err
	}
	token, err := c.tokenprovider.GetToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 399 {
		c.logger.Error("unable to retrieve user", "body", string(body), "statuscode", res.StatusCode)
		return nil, fmt.Errorf("received unsuccessful status code when retrieving keycloak user. code: %d", res.StatusCode)
	}

	var kcuser *user
	err = json.Unmarshal(body, &kcuser)
	if err != nil {
		return nil, err
	}
	return &authmgmt.User{
		Name:    kcuser.Name(),
		Email:   kcuser.Email,
		Picture: "",
	}, nil
}
