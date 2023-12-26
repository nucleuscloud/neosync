package v1alpha1_connectiondataservice

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

func (s *Service) verifyUserInAccount(
	ctx context.Context,
	accountId string,
) (*pgtype.UUID, error) { // nolint
	resp, err := s.useraccountService.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return nil, err
	}
	if !resp.Msg.Ok {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}

	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}
	return &accountUuid, nil
}
