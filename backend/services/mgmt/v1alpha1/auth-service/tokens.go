package v1alpha1_authservice

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/gogo/status"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"google.golang.org/grpc/codes"
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
		return nil, status.Errorf(codes.Unauthenticated, "Request unauthenticated")
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

func (s *Service) GetAuthorizeUrl(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAuthorizeUrlRequest],
) (*connect.Response[mgmtv1alpha1.GetAuthorizeUrlResponse], error) {
	params := url.Values{}
	params.Add("client_id", s.cfg.CliClientId)
	params.Add("audience", req.Msg.Audience)
	params.Add("scope", req.Msg.Scope)
	params.Add("redirect_uri", req.Msg.RedirectUri)
	params.Add("state", req.Msg.State)
	params.Add("response_type", "code")

	authorizeUrl := fmt.Sprintf("%s?%s", s.cfg.AuthorizeUrl, params.Encode())
	return connect.NewResponse(&mgmtv1alpha1.GetAuthorizeUrlResponse{
		Url: authorizeUrl,
	}), nil
}

// func (s *Service) GetAccessToken(
// 	ctx context.Context,
// 	req *connect.Request[mgmtv1alpha1.GetAccessTokenRequest],
// ) (*connect.Response[mgmtv1alpha1.GetAccessTokenResponse], error) {
// 	return nil, nucleuserrors.NewNotImplemented("method is not yet implemented")
// }

// func (s *Service) RefreshAccessToken(
// 	ctx context.Context,
// 	req *connect.Request[mgmtv1alpha1.RefreshAccessTokenRequest],
// ) (*connect.Response[mgmtv1alpha1.RefreshAccessTokenResponse], error) {
// 	return nil, nucleuserrors.NewNotImplemented("method is not yet implemented")
// }
