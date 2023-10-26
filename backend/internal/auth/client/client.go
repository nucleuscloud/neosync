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
	tokenurl          string
	clientIdSecretMap map[string]string
}

func New(
	tokenurl string,
	clientIdSecretMap map[string]string,
) *Client {
	return &Client{
		tokenurl:          tokenurl,
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

func (c *Client) GetTokenResponse(
	ctx context.Context,
	clientId string,
	code string,
	redirecturi string,
) (*AuthTokenResponse, error) {
	if _, ok := c.clientIdSecretMap[clientId]; !ok {
		return nil, errors.New("unknown client id, requested client was notin safelist")
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
	req, err := http.NewRequestWithContext(ctx, "POST", c.tokenurl, payload)
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
