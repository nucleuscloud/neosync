package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
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
		return "", err
	}
	return fmt.Sprintf("Bearer %s", token), nil
}

func getToken(ctx context.Context) (string, error) {
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(http.DefaultClient, serverconfig.GetApiBaseUrl())

	issuerResp, err := authclient.GetCliIssuer(ctx, connect.NewRequest(&mgmtv1alpha1.GetCliIssuerRequest{}))
	if err != nil {
		return "", err
	}

	jwtvalidator, err := getJwtValidator(issuerResp.Msg.IssuerUrl, issuerResp.Msg.Audience)
	if err != nil {
		return "", err
	}

	accessToken, err := userconfig.GetAccessToken()
	if err != nil {
		return "", err
	}
	_, err = jwtvalidator.ValidateToken(ctx, accessToken)
	if err != nil {
		err = userconfig.RemoveAccessToken()
		if err != nil {
			return "", err
		}
		slog.Info("access token is no longer valid. attempting to refresh...")
		refreshtoken, err := userconfig.GetRefreshToken()
		if err != nil {
			slog.Info("unable to find refresh token")
			return "", err
		}
		refreshResp, err := authclient.RefreshCli(ctx, connect.NewRequest(&mgmtv1alpha1.RefreshCliRequest{
			RefreshToken: refreshtoken,
		}))
		if err != nil {
			slog.Info("unable to refresh token")
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

func getJwtValidator(issuerurl, audience string) (*validator.Validator, error) {
	issuerUrl, err := url.Parse(issuerurl)
	if err != nil {
		return nil, err
	}
	provider := jwks.NewProvider(issuerUrl)

	jwtvalidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerUrl.String(),
		[]string{audience},
		validator.WithCustomClaims(func() validator.CustomClaims { return nil }),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, err
	}
	return jwtvalidator, nil
}

func IsAuthEnabled(ctx context.Context) (bool, error) {
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(http.DefaultClient, serverconfig.GetApiBaseUrl())
	isEnabledResp, err := authclient.GetAuthStatus(ctx, connect.NewRequest(&mgmtv1alpha1.GetAuthStatusRequest{}))
	if err != nil {
		return false, err
	}
	return isEnabledResp.Msg.IsEnabled, nil
}
