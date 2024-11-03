package v1alpha1_useraccountservice

import (
	"context"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

func (s *Service) GetAccountTemporalConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountTemporalConfigRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountTemporalConfigResponse], error) {
	if s.cfg.IsNeosyncCloud {
		return nil, nucleuserrors.NewNotImplemented("not enabled in Neosync Cloud")
	}
	_, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	tc, err := s.temporalConfigProvider.GetConfig(ctx, req.Msg.GetAccountId())
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
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	dtoCfg := req.Msg.GetConfig()
	if dtoCfg == nil {
		dtoCfg = &mgmtv1alpha1.AccountTemporalConfig{}
	}

	tc := &pg_models.TemporalConfig{}
	tc.FromDto(dtoCfg)

	_, err = s.db.Q.UpdateTemporalConfigByAccount(ctx, s.db.Db, db_queries.UpdateTemporalConfigByAccountParams{
		TemporalConfig: tc,
		AccountId:      *accountUuid,
	})
	if err != nil {
		return nil, err
	}

	updatedConfig, err := s.temporalConfigProvider.GetConfig(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetAccountTemporalConfigResponse{
		Config: updatedConfig.ToDto(),
	}), nil
}
