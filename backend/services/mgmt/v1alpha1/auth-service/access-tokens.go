package v1alpha1_authservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) GetNewAccessToken(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetNewAccessTokenRequest],
) (*connect.Response[mgmtv1alpha1.GetNewAccessTokenResponse], error) {
	if req.Msg.ClientId == "" || req.Msg.RefreshToken == "" {
		return nil, nucleuserrors.NewBadRequest("must provide client id and refresh token")
	}

	// logger, err := loggermiddleware.GetLoggerFromContext(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// logger = logger.WithValues("clientId", req.ClientId)

	res, err := s.authclient.GetRefreshedAccessToken(req.Msg.ClientId, req.Msg.RefreshToken)
	if err != nil {
		// logger.Error(err, "unable to refresh access token")
		return nil, nucleuserrors.New(err)
	} else if res.Error != nil {
		// logger.Error(
		// 	fmt.Errorf("Unable to refresh token. Title: %s -- Description: %s", res.Error.Error, res.Error.ErrorDescription),
		// 	"Unable to get refresh token",
		// )
		return nil, status.Errorf(codes.Unauthenticated, "unable to refresh access token")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetNewAccessTokenResponse{
		AccessToken:  res.Result.AccessToken,
		RefreshToken: res.Result.RefreshToken,
		ExpiresIn:    int64(res.Result.ExpiresIn),
		Scope:        res.Result.Scope,
		IdToken:      res.Result.IdToken,
		TokenType:    res.Result.TokenType,
	}), nil
}

func (s *Service) GetAccessToken(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccessTokenRequest],
) (*connect.Response[mgmtv1alpha1.GetAccessTokenResponse], error) {
	if req.Msg.ClientId == "" || req.Msg.Code == "" || req.Msg.RedirectUri == "" {
		return nil, nucleuserrors.NewBadRequest("must provide client id, code, and redirect uri")
	}

	// logger, err := loggermiddleware.GetLoggerFromContext(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// logger = logger.WithValues("clientId", req.ClientId)

	res, err := s.authclient.GetTokenResponse(req.Msg.ClientId, req.Msg.Code, req.Msg.RedirectUri)
	if err != nil {
		// logger.Error(err, "unable to retrieve access token")
		return nil, nucleuserrors.New(err)
	} else if res.Error != nil {
		// logger.Error(
		// 	fmt.Errorf("Unable to get access token. Title: %s -- Description: %s", res.Error.Error, res.Error.ErrorDescription),
		// 	"Unable to get access token",
		// )
		return nil, status.Errorf(codes.Unauthenticated, "Request unauthenticated")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccessTokenResponse{
		AccessToken:  res.Result.AccessToken,
		RefreshToken: res.Result.RefreshToken,
		ExpiresIn:    int64(res.Result.ExpiresIn),
		Scope:        res.Result.Scope,
		IdToken:      res.Result.IdToken,
		TokenType:    res.Result.TokenType,
	}), nil
}
