package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/nucleuscloud/neosync/cli/internal/version"
	http_client "github.com/nucleuscloud/neosync/worker/pkg/http/client"
	"github.com/spf13/viper"
)

// Light wrapper for GetAuthEnabled that instantiates an auth client
func IsAuthEnabled(ctx context.Context) (bool, error) {
	httpclient := http_client.NewWithHeaders(version.Get().Headers())
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(httpclient, GetNeosyncUrl())
	return GetAuthEnabled(ctx, authclient)
}

func GetAuthEnabled(
	ctx context.Context,
	authclient mgmtv1alpha1connect.AuthServiceClient,
) (bool, error) {
	isEnabledResp, err := authclient.GetAuthStatus(ctx, connect.NewRequest(&mgmtv1alpha1.GetAuthStatusRequest{}))
	if err != nil {
		return false, err
	}
	return isEnabledResp.Msg.GetIsEnabled(), nil
}

// This variable is replaced at build time
var defaultBaseUrl string = "http://localhost:8080"

// Returns the neosync url found in the environment, otherwise defaults to localhost
func GetNeosyncUrl() string {
	baseurl := viper.GetString("NEOSYNC_API_URL")
	if baseurl == "" {
		return defaultBaseUrl
	}
	return baseurl
}

type HttpClientConfig struct {
	apiKey       *string
	extraHeaders map[string]string
}

type HttpOption func(cfg *HttpClientConfig)

func WithApiKey(apiKey *string) HttpOption {
	return func(cfg *HttpClientConfig) {
		cfg.apiKey = apiKey
	}
}

// If desired, append any extra headers.
// Note: version headers are already appended to the client when calling GetNeosyncHttpClient
func WithExtraHeaders(headers map[string]string) HttpOption {
	return func(cfg *HttpClientConfig) {
		cfg.extraHeaders = headers
	}
}

// Returns an instance of *http.Client that includes the Neosync API Token if one was found in the environment
func GetNeosyncHttpClient(ctx context.Context, logger *slog.Logger, opts ...HttpOption) (*http.Client, error) {
	cfg := &HttpClientConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	isAuthEnabled, err := IsAuthEnabled(ctx)
	if err != nil {
		return nil, err
	}

	headers := http_client.MergeMaps(version.Get().Headers(), cfg.extraHeaders)
	if !isAuthEnabled {
		return http_client.NewWithHeaders(headers), nil
	}
	if cfg.apiKey != nil && *cfg.apiKey != "" {
		headers = http_client.MergeMaps(headers, http_client.GetBearerAuthHeaders(cfg.apiKey))
	} else {
		accessToken, err := getAccessToken(ctx, headers, logger)
		if err != nil {
			return nil, err
		}
		headers = http_client.MergeMaps(headers, http_client.GetBearerAuthHeaders(&accessToken))
	}

	return http_client.NewWithHeaders(headers), nil
}

func getAccessToken(ctx context.Context, headers map[string]string, logger *slog.Logger) (string, error) {
	httpclient := http_client.NewWithHeaders(headers)
	neosyncurl := GetNeosyncUrl()
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(httpclient, neosyncurl)

	accessToken, err := userconfig.GetAccessToken()
	if err != nil {
		return "", err
	}
	authedAuthClient := mgmtv1alpha1connect.NewAuthServiceClient(
		http_client.NewWithHeaders(http_client.MergeMaps(headers, http_client.GetBearerAuthHeaders(&accessToken))),
		neosyncurl,
	)
	logger.Debug("found existing access token, checking if still valid")
	_, err = authedAuthClient.CheckToken(ctx, connect.NewRequest(&mgmtv1alpha1.CheckTokenRequest{}))
	if err != nil {
		if err := userconfig.RemoveAccessToken(); err != nil {
			return "", err
		}
		logger.Debug(fmt.Errorf("access token is no longer valid. attempting to refresh...: %w", err).Error())
		refreshtoken, err := userconfig.GetRefreshToken()
		if err != nil {
			return "", fmt.Errorf("unable to find refresh token: %w", err)
		}
		refreshResp, err := authclient.RefreshCli(ctx, connect.NewRequest(&mgmtv1alpha1.RefreshCliRequest{
			RefreshToken: refreshtoken,
		}))
		if err != nil {
			return "", fmt.Errorf("unable to refresh token, must login again: %w", err)
		}
		err = userconfig.SetAccessToken(refreshResp.Msg.AccessToken.AccessToken)
		if err != nil {
			logger.Warn("unable to write refreshed access token back to user config", "error", err.Error())
		}
		if refreshResp.Msg.AccessToken.RefreshToken != nil {
			err = userconfig.SetRefreshToken(*refreshResp.Msg.AccessToken.RefreshToken)
			if err != nil {
				logger.Warn("unable to write refreshed refresh token back to user config", "error", err.Error())
			}
		}
		return refreshResp.Msg.GetAccessToken().GetAccessToken(), nil
	}
	return accessToken, nil
}
