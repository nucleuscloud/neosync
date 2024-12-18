package v1alpha1_authservice

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

func (s *Service) GetAuthStatus(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAuthStatusRequest],
) (*connect.Response[mgmtv1alpha1.GetAuthStatusResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetAuthStatusResponse{
		IsEnabled: s.cfg.IsAuthEnabled,
	}), nil
}

func (s *Service) LoginCli(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.LoginCliRequest],
) (*connect.Response[mgmtv1alpha1.LoginCliResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	resp, err := s.authclient.GetTokenResponse(ctx, s.cfg.CliClientId, req.Msg.Code, req.Msg.RedirectUri)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		logger.Error(
			fmt.Sprintf("Unable to get access token. Title: %s -- Description: %s", resp.Error.Error, resp.Error.ErrorDescription),
		)
		return nil, nucleuserrors.NewUnauthenticated("Request unauthenticated")
	}
	var refreshToken *string
	if resp.Result.RefreshToken != "" {
		refreshToken = &resp.Result.RefreshToken
	}
	var idToken *string
	if resp.Result.IdToken != "" {
		idToken = &resp.Result.IdToken
	}
	return connect.NewResponse(&mgmtv1alpha1.LoginCliResponse{
		AccessToken: &mgmtv1alpha1.AccessToken{
			AccessToken:  resp.Result.AccessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(resp.Result.ExpiresIn),
			Scope:        resp.Result.Scope,
			IdToken:      idToken,
			TokenType:    resp.Result.TokenType,
		},
	}), nil
}

func (s *Service) RefreshCli(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RefreshCliRequest],
) (*connect.Response[mgmtv1alpha1.RefreshCliResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	resp, err := s.authclient.GetRefreshedAccessToken(ctx, s.cfg.CliClientId, req.Msg.RefreshToken)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		logger.Error(
			fmt.Sprintf("Unable to get refreshed token. Title: %s -- Description: %s", resp.Error.Error, resp.Error.ErrorDescription),
		)
		return nil, nucleuserrors.NewUnauthenticated("Unable to refresh access token")
	}
	var refreshToken *string
	if resp.Result.RefreshToken != "" {
		refreshToken = &resp.Result.RefreshToken
	}
	var idToken *string
	if resp.Result.IdToken != "" {
		idToken = &resp.Result.IdToken
	}
	return connect.NewResponse(&mgmtv1alpha1.RefreshCliResponse{
		AccessToken: &mgmtv1alpha1.AccessToken{
			AccessToken:  resp.Result.AccessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(resp.Result.ExpiresIn),
			Scope:        resp.Result.Scope,
			IdToken:      idToken,
			TokenType:    resp.Result.TokenType,
		},
	}), nil
}

func (s *Service) GetAuthorizeUrl(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAuthorizeUrlRequest],
) (*connect.Response[mgmtv1alpha1.GetAuthorizeUrlResponse], error) {
	params := url.Values{}
	params.Add("client_id", s.cfg.CliClientId)
	params.Add("audience", s.cfg.CliAudience)
	params.Add("scope", req.Msg.Scope)
	params.Add("redirect_uri", req.Msg.RedirectUri)
	params.Add("state", req.Msg.State)
	params.Add("response_type", "code")

	authorizeUrl, err := s.authclient.GetAuthorizationEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAuthorizeUrlResponse{
		Url: fmt.Sprintf("%s?%s", authorizeUrl, params.Encode()),
	}), nil
}

func (s *Service) CheckToken(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckTokenRequest],
) (*connect.Response[mgmtv1alpha1.CheckTokenResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.CheckTokenResponse{}), nil
}
