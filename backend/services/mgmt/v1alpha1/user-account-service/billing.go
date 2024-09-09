package v1alpha1_useraccountservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
)

const (
	AccountTypePersonal = iota
	AccountTypeTeam
)

func (s *Service) GetAccountStatus(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountStatusRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountStatusResponse], error) {
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, *accountId)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account: %w", err)
	}

	if s.cfg.IsNeosyncCloud {
		if account.AccountType == int16(dtomaps.AccountType_Personal) {
			return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
				UsedRecordCount:    0,
				AllowedRecordCount: nil,
				SubscriptionStatus: 0,
			}), nil
		}
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
}

func (s *Service) IsAccountStatusValid(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsAccountStatusValidRequest],
) (*connect.Response[mgmtv1alpha1.IsAccountStatusValidResponse], error) {
	accountStatusResp, err := s.GetAccountStatus(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountStatusRequest{
		AccountId: req.Msg.GetAccountId(),
	}))
	if err != nil {
		return nil, err
	}
	// todo: sub status will change based on is cloud, or if expired, etc.
	if accountStatusResp.Msg.GetSubscriptionStatus() > 0 || accountStatusResp.Msg.AllowedRecordCount == nil {
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
	}

	isOverSubscribed := accountStatusResp.Msg.GetUsedRecordCount() >= accountStatusResp.Msg.GetAllowedRecordCount()

	return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{
		IsValid: isOverSubscribed,
	}), nil
}
