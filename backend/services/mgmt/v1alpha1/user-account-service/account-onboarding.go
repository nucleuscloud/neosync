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

func (s *Service) GetAccountOnboardingConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountOnboardingConfigRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountOnboardingConfigResponse], error) {
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

	oc, err := s.db.Q.GetAccountOnboardingConfig(ctx, s.db.Db, accountUuid)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccountOnboardingConfigResponse{
		Config: oc.ToDto(),
	}), nil
}

func (s *Service) SetAccountOnboardingConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetAccountOnboardingConfigRequest],
) (*connect.Response[mgmtv1alpha1.SetAccountOnboardingConfigResponse], error) {
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

	tc := &pg_models.AccountOnboardingConfig{}
	if req.Msg.Config != nil {
		tc.FromDto(req.Msg.Config)
	}

	account, err := s.db.Q.UpdateAccountOnboardingConfig(ctx, s.db.Db, db_queries.UpdateAccountOnboardingConfigParams{
		OnboardingConfig: tc,
		AccountId:        accountUuid,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetAccountOnboardingConfigResponse{
		Config: account.OnboardingConfig.ToDto(),
	}), nil
}
