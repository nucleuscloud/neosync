package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
)

const (
	AuthHeader = "Authorization"
)

func GetAuthHeaderTokenFn(
	apiKey *string,
) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		if apiKey != nil && *apiKey != "" {
			return fmt.Sprintf("Bearer %s", *apiKey), nil
		}
		return getAuthHeaderToken(ctx)
	}
}

func getAuthHeaderToken(ctx context.Context) (string, error) {
	token, err := getToken(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get access token, try running neosync login again or provide an API Key: %w", err)
	}
	return fmt.Sprintf("Bearer %s", token), nil
}

func getToken(ctx context.Context) (string, error) {
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(http.DefaultClient, serverconfig.GetApiBaseUrl())

	accessToken, err := userconfig.GetAccessToken()
	if err != nil {
		return "", err
	}
	authedAuthClient := mgmtv1alpha1connect.NewAuthServiceClient(
		newHttpClient(map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		}),
		serverconfig.GetApiBaseUrl(),
	)
	_, err = authedAuthClient.CheckToken(ctx, connect.NewRequest(&mgmtv1alpha1.CheckTokenRequest{}))
	if err != nil {
		if err := userconfig.RemoveAccessToken(); err != nil {
			return "", err
		}
		slog.Info(fmt.Errorf("access token is no longer valid. attempting to refresh...: %w", err).Error())
		refreshtoken, err := userconfig.GetRefreshToken()
		if err != nil {
			slog.Info(fmt.Errorf("unable to find refresh token: %w", err).Error())
			return "", err
		}
		refreshResp, err := authclient.RefreshCli(ctx, connect.NewRequest(&mgmtv1alpha1.RefreshCliRequest{
			RefreshToken: refreshtoken,
		}))
		if err != nil {
			slog.Info(fmt.Errorf("unable to refresh token: %w", err).Error())
			return "", err
		}
		err = userconfig.SetAccessToken(refreshResp.Msg.AccessToken.AccessToken)
		if err != nil {
			slog.Warn("unable to write refreshed access token back to user config", "error", err.Error())
		}
		if refreshResp.Msg.AccessToken.RefreshToken != nil {
			err = userconfig.SetRefreshToken(*refreshResp.Msg.AccessToken.RefreshToken)
			if err != nil {
				slog.Warn("unable to write refreshed refresh token back to user config", "error", err.Error())
			}
		}
		return refreshResp.Msg.AccessToken.AccessToken, nil
	}
	return accessToken, nil
}

func IsAuthEnabled(ctx context.Context) (bool, error) {
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(http.DefaultClient, serverconfig.GetApiBaseUrl())
	isEnabledResp, err := authclient.GetAuthStatus(ctx, connect.NewRequest(&mgmtv1alpha1.GetAuthStatusRequest{}))
	if err != nil {
		return false, err
	}
	return isEnabledResp.Msg.IsEnabled, nil
}

func newHttpClient(
	headers map[string]string,
) *http.Client {
	return &http.Client{
		Transport: &headerTransport{
			Transport: http.DefaultTransport,
			Headers:   headers,
		},
	}
}

type headerTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}
	for key, value := range t.Headers {
		req.Header.Add(key, value)
	}
	return t.Transport.RoundTrip(req)
}
