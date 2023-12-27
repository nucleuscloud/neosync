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
	tokenurl, err := c.getTokenEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", tokenurl, payload)
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
	tokenurl, err := c.getTokenEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", tokenurl, payload)

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

func (c *Client) getTokenEndpoint(ctx context.Context) (string, error) {
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

type openIdConfiguration struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

func (c *Client) getOpenIdConfiguration(ctx context.Context) (*openIdConfiguration, error) {
	configUrl := fmt.Sprintf("%s/.well-known/openid-configuration", strings.TrimSuffix(c.authBaseUrl, "/"))

	req, err := http.NewRequestWithContext(ctx, "GET", configUrl, http.NoBody)
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
