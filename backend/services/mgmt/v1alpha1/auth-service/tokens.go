package v1alpha1_authservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
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

func (s *Service) GetAccessToken(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccessTokenRequest],
) (*connect.Response[mgmtv1alpha1.GetAccessTokenResponse], error) {
	return nil, nucleuserrors.NewNotImplemented("method is not yet implemented")
}

func (s *Service) RefreshAccessToken(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RefreshAccessTokenRequest],
) (*connect.Response[mgmtv1alpha1.RefreshAccessTokenResponse], error) {
	return nil, nucleuserrors.NewNotImplemented("method is not yet implemented")
}
