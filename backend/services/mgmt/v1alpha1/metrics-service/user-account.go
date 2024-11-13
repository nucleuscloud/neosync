package v1alpha1_metricsservice

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

//nolint:unparam
func (s *Service) verifyUserInAccount(
	ctx context.Context,
	accountId string,
) (*pgtype.UUID, error) {
	accountUuid, err := neosyncdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}

	if isWorkerApiKey(ctx) {
		return &accountUuid, nil
	}

	resp, err := s.useraccountservice.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return nil, err
	}
	if !resp.Msg.Ok {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}

	return &accountUuid, nil
}

func isWorkerApiKey(ctx context.Context) bool {
	data, err := auth_apikey.GetTokenDataFromCtx(ctx)
	if err != nil {
		return false
	}
	return data.ApiKeyType == apikey.WorkerApiKey
}

// func (s *Service) getUserUuid(
// 	ctx context.Context,
// ) (*pgtype.UUID, error) {
// 	user, err := s.useraccountservice.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
// 	if err != nil {
// 		return nil, err
// 	}
// 	userUuid, err := neosyncdb.ToUuid(user.Msg.UserId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &userUuid, nil
// }
