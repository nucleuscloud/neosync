package v1alpha1_authservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

func (s *Service) GetAuthUser(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAuthUserRequest],
) (*connect.Response[mgmtv1alpha1.GetAuthUserResponse], error) {
	switch req.Msg.AuthProvider {
	case mgmtv1alpha1.AuthProvider_AUTH_PROVIDER_AUTH_0:
		authUser, err := s.auth0Mgmt.GetUserById(ctx, req.Msg.UserId)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&mgmtv1alpha1.GetAuthUserResponse{
			User: &mgmtv1alpha1.AuthUser{
				Id:    req.Msg.UserId,
				Name:  authUser.GetName(),
				Email: authUser.GetEmail(),
				Image: authUser.GetPicture(),
			},
		}), nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid auth provider type")
	}
}
