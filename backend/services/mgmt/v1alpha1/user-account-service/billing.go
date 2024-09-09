package v1alpha1_useraccountservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
)

const (
	AccountTypePersonal = iota
	AccountTypeTeam
)

func (s *Service) GetAccountStatus(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountStatusRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountStatusResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

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
			usedRecordCount, err := s.getUsedRecordCountForMonth(ctx, req.Msg.GetAccountId(), logger)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve used record count for month: %w", err)
			}
			return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
				UsedRecordCount:    usedRecordCount,
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

func (s *Service) getUsedRecordCountForMonth(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (uint64, error) {
	today := time.Now().UTC()

	promWindow := fmt.Sprintf("%dd", today.Day())
	query, err := metrics.GetPromQueryFromMetric(mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED, metrics.MetricLabels{}, promWindow)
	if err != nil {
		return 0, err
	}
	totalCount, err := metrics.GetTotalUsageFromProm(ctx, s.prometheusclient, query, today, logger)
	if err != nil {
		return 0, err
	}
	if totalCount < 0 {
		logger.Warn("the response from prometheus returned a negative count when computing used records")
		totalCount = 0
	}
	return uint64(totalCount), nil
}
