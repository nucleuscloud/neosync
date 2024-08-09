package v1alpha1_useraccountservice

import (
	"context"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

func (s *Service) GetAccountOnboardingConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountOnboardingConfigRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountOnboardingConfigResponse], error) {
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	oc, err := s.db.Q.GetAccountOnboardingConfig(ctx, s.db.Db, *accountId)
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
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	inputCfg := req.Msg.GetConfig()
	if inputCfg == nil {
		inputCfg = &mgmtv1alpha1.AccountOnboardingConfig{}
	}

	onboardingConfigModel := &pg_models.AccountOnboardingConfig{}
	onboardingConfigModel.FromDto(inputCfg)

	account, err := s.db.Q.UpdateAccountOnboardingConfig(ctx, s.db.Db, db_queries.UpdateAccountOnboardingConfigParams{
		OnboardingConfig: onboardingConfigModel,
		AccountId:        *accountId,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetAccountOnboardingConfigResponse{
		Config: account.OnboardingConfig.ToDto(),
	}), nil
}
