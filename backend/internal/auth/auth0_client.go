package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth0Client struct {
	tokenUrl string

	clientIdSecretMap map[string]string

	// logger logr.Logger
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

type ServiceAccountTokenResponseData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type ServiceAccountTokenResponse struct {
	Result *ServiceAccountTokenResponseData
	Error  *AuthTokenErrorData
}

func New(
	baseUrl string,
	clientIdSecretMap map[string]string,
) (*Auth0Client, error) {
	if clientIdSecretMap == nil {
		clientIdSecretMap = map[string]string{}
	}

	return &Auth0Client{
		tokenUrl: fmt.Sprintf("%s/oauth/token", baseUrl),

		clientIdSecretMap: clientIdSecretMap,
	}, nil
}

func getHttpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func (c *Auth0Client) GetTokenResponse(
	clientId string,
	code string,
	redirectUri string,
) (*AuthTokenResponse, error) {
	if _, ok := c.clientIdSecretMap[clientId]; !ok {
		// c.logger.Error(fmt.Errorf("unknown client id"), "requested client id was not in safelist")
		return nil, fmt.Errorf("unknown client id")
	}
	clientSecret := c.clientIdSecretMap[clientId]

	payload := strings.NewReader(
		fmt.Sprintf(
			"grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
			clientId,
			clientSecret,
			code,
			redirectUri,
		),
	)
	req, err := http.NewRequest("POST", c.tokenUrl, payload)
	if err != nil {
		// c.logger.Error(err, "unable to request oauth authorization code req")
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := getHttpClient().Do(req)
	if err != nil {
		// c.logger.Error(err, "unable to fulfill authorization code req")
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// c.logger.Error(err, "unable to read body from authorization code req")
		return nil, err
	}

	var tokenResponse *AuthTokenResponseData
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		// c.logger.Error(err, "unable to unmarshal token response from refresh token req")
		return nil, err
	}

	if tokenResponse.AccessToken == "" {
		var errorResponse AuthTokenErrorData
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			// c.logger.Error(err, "unable to unmarshal error response from refresh token req")
			return nil, err
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

func (c *Auth0Client) GetRefreshedAccessToken(clientId string, refreshToken string) (*AuthTokenResponse, error) {
	if _, ok := c.clientIdSecretMap[clientId]; !ok {
		// c.logger.Error(fmt.Errorf("unknown client id"), "requested client id was not in safelist")
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth client id")
	}
	clientSecret := c.clientIdSecretMap[clientId]

	payload := strings.NewReader(
		fmt.Sprintf(
			"grant_type=refresh_token&client_id=%s&client_secret=%s&refresh_token=%s", clientId, clientSecret, refreshToken,
		),
	)
	req, err := http.NewRequest("POST", c.tokenUrl, payload)

	if err != nil {
		// c.logger.Error(err, "unable to initiate refresh token req")
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := getHttpClient().Do(req)

	if err != nil {
		// c.logger.Error(err, "unable to fulfill refresh token req")
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		// c.logger.Error(err, "unable to read body from refresh token req")
		return nil, err
	}

	var tokenResponse *AuthTokenResponseData
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		// c.logger.Error(err, "unable to unmarshal token response from refresh token req")
		return nil, err
	}

	if tokenResponse.AccessToken == "" {
		var errorResponse AuthTokenErrorData
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			// c.logger.Error(err, "unable to unmarshal error response from refresh token req")
			return nil, err
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

func (c *Auth0Client) GetServiceAccountTokenResponse(
	clientId string,
	clientSecret string,
	audience string,
) (*ServiceAccountTokenResponse, error) {
	payload := strings.NewReader(
		fmt.Sprintf(
			"grant_type=client_credentials&client_id=%s&client_secret=%s&audience=%s", clientId, clientSecret, audience,
		),
	)

	req, err := http.NewRequest("POST", c.tokenUrl, payload)
	if err != nil {
		// c.logger.Error(err, "unable to request oauth authorization code req")
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := getHttpClient().Do(req)
	if err != nil {
		// c.logger.Error(err, "unable to fulfill authorization code req")
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// c.logger.Error(err, "unable to read body from authorization code req")
		return nil, err
	}

	var tokenResponse ServiceAccountTokenResponseData
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		// c.logger.Error(err, "unable to unmarshal token response from refresh token req")
		return nil, err
	}

	if tokenResponse.AccessToken == "" {
		var errorResponse AuthTokenErrorData
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			// c.logger.Error(err, "unable to unmarshal error response from refresh token req")
			return nil, err
		}
		return &ServiceAccountTokenResponse{
			Result: nil,
			Error:  &errorResponse,
		}, nil
	}
	return &ServiceAccountTokenResponse{
		Result: &tokenResponse,
		Error:  nil,
	}, nil
}
