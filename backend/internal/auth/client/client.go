package auth_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Interface interface {
	GetTokenResponse(ctx context.Context, clientId string, code string, redirecturi string) (*AuthTokenResponse, error)
	GetRefreshedAccessToken(ctx context.Context, clientId string, refreshToken string) (*AuthTokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
	GetTokenEndpoint(ctx context.Context) (string, error)
	GetAuthorizationEndpoint(ctx context.Context) (string, error)
}

type Client struct {
	authBaseUrl       string
	clientIdSecretMap map[string]string
}

func New(
	authBaseUrl string,
	clientIdSecretMap map[string]string,
) *Client {
	return &Client{
		authBaseUrl:       authBaseUrl,
		clientIdSecretMap: clientIdSecretMap,
	}
}

type AuthTokenResponse struct {
	Result *AuthTokenResponseData
	Error  *AuthTokenErrorData
}

type AuthTokenResponseData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IdToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
}

type AuthTokenErrorData struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type UserInfo struct {
	Sub           string `json:"sub"`
	Nickname      string `json:"nickname"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	UpdatedAt     string `json:"updated_at"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func getHttpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

// Uses Authorization Flow defined here: https://auth0.com/docs/api/authentication#authorization-code-flow47
func (c *Client) GetTokenResponse(
	ctx context.Context,
	clientId string,
	code string,
	redirecturi string,
) (*AuthTokenResponse, error) {
	if _, ok := c.clientIdSecretMap[clientId]; !ok {
		return nil, errors.New("unknown client id, requested client was not in safelist")
	}

	clientSecret := c.clientIdSecretMap[clientId]
	payload := strings.NewReader(
		fmt.Sprintf(
			"grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
			clientId,
			clientSecret,
			code,
			redirecturi,
		),
	)
	tokenurl, err := c.GetTokenEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenurl, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to request oauth authorization code: %w", err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := getHttpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfill authorization code request: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to ready body from authorization code response: %w", err)
	}

	var tokenResponse *AuthTokenResponseData
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		return nil, fmt.Errorf("unable to decode token response data from body: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		var errorResponse AuthTokenErrorData
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, fmt.Errorf("unable to decode error response data from body: %w", err)
		}
		return &AuthTokenResponse{
			Result: nil,
			Error:  &errorResponse,
		}, nil
	}
	return &AuthTokenResponse{
		Result: tokenResponse,
		Error:  nil,
	}, nil
}

func (c *Client) GetRefreshedAccessToken(
	ctx context.Context,
	clientId string,
	refreshToken string,
) (*AuthTokenResponse, error) {
	if _, ok := c.clientIdSecretMap[clientId]; !ok {
		return nil, errors.New("unknown client id, requested client was not in safelist")
	}

	clientSecret := c.clientIdSecretMap[clientId]
	payload := strings.NewReader(
		fmt.Sprintf(
			"grant_type=refresh_token&client_id=%s&client_secret=%s&refresh_token=%s", clientId, clientSecret, refreshToken,
		),
	)
	tokenurl, err := c.GetTokenEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenurl, payload)

	if err != nil {
		return nil, fmt.Errorf("unable to initiate refresh token request: %w", err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := getHttpClient().Do(req)

	if err != nil {
		return nil, fmt.Errorf("unable to fulfill refresh token request: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("unable to read body from refresh token request: %w", err)
	}

	var tokenResponse *AuthTokenResponseData
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal token response from refresh token request: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		var errorResponse AuthTokenErrorData
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal error response from refresh token request: %w", err)
		}
		return &AuthTokenResponse{
			Result: nil,
			Error:  &errorResponse,
		}, nil
	}
	return &AuthTokenResponse{
		Result: tokenResponse,
		Error:  nil,
	}, nil
}

func (c *Client) GetUserInfo(
	ctx context.Context,
	accessToken string,
) (*UserInfo, error) {
	userinfourl, err := c.getUserInfoEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userinfourl, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := getHttpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to request user info: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body from user info request: %w", err)
	}

	var userinfo *UserInfo
	err = json.Unmarshal(body, &userinfo)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal user info into struct: %w", err)
	}
	return userinfo, nil
}

func (c *Client) GetTokenEndpoint(ctx context.Context) (string, error) {
	config, err := c.getOpenIdConfiguration(ctx)
	if err != nil {
		return "", err
	}
	if config.TokenEndpoint == "" {
		return "", errors.New("unable to find token endpoint")
	}
	return config.TokenEndpoint, nil
}

func (c *Client) GetAuthorizationEndpoint(ctx context.Context) (string, error) {
	config, err := c.getOpenIdConfiguration(ctx)
	if err != nil {
		return "", err
	}
	if config.AuthorizationEndpoint == "" {
		return "", errors.New("unable to find authorization endpoint")
	}
	return config.AuthorizationEndpoint, nil
}

func (c *Client) getUserInfoEndpoint(ctx context.Context) (string, error) {
	config, err := c.getOpenIdConfiguration(ctx)
	if err != nil {
		return "", err
	}
	if config.UserinfoEndpoint == "" {
		return "", errors.New("unable to find userinfo endpoint")
	}
	return config.UserinfoEndpoint, nil
}

type openIdConfiguration struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

func (c *Client) getOpenIdConfiguration(ctx context.Context) (*openIdConfiguration, error) {
	configUrl := fmt.Sprintf("%s/.well-known/openid-configuration", strings.TrimSuffix(c.authBaseUrl, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configUrl, http.NoBody)
	if err != nil {
		return nil, err
	}

	res, err := getHttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var bodyResp *openIdConfiguration
	err = json.Unmarshal(body, &bodyResp)
	if err != nil {
		return nil, err
	}
	return bodyResp, nil
}
