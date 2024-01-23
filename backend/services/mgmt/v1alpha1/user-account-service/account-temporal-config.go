package v1alpha1_useraccountservice

import (
	"context"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

func (s *Service) GetAccountTemporalConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountTemporalConfigRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountTemporalConfigResponse], error) {
	if s.cfg.IsNeosyncCloud {
		return nil, nucleuserrors.NewNotImplemented("not enabled in Neosync Cloud")
	}
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userUuid, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}
	accountUuid, err := nucleusdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	count, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
		UserId:    userUuid,
		AccountId: accountUuid,
	})
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nucleuserrors.NewForbidden("user is not in account")
	}

	tc, err := s.temporalClientManager.GetTemporalConfigByAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccountTemporalConfigResponse{
		Config: tc.ToDto(),
	}), nil
}

func (s *Service) SetAccountTemporalConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetAccountTemporalConfigRequest],
) (*connect.Response[mgmtv1alpha1.SetAccountTemporalConfigResponse], error) {
	if s.cfg.IsNeosyncCloud {
		return nil, nucleuserrors.NewNotImplemented("not enabled in Neosync Cloud")
	}
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userUuid, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	accountUuid, err := nucleusdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
		UserId:    userUuid,
		AccountId: accountUuid,
	})
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nucleuserrors.NewForbidden("user is not in account")
	}

	tc := &pg_models.TemporalConfig{}
	tc.FromDto(req.Msg.Config)

	account, err := s.db.Q.UpdateTemporalConfigByAccount(ctx, s.db.Db, db_queries.UpdateTemporalConfigByAccountParams{
		TemporalConfig: tc,
		AccountId:      accountUuid,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetAccountTemporalConfigResponse{
		Config: account.TemporalConfig.ToDto(),
	}), nil
}
