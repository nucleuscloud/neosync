package v1alpha1_useraccountservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
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
			allowedRecordCount := getAllowedRecordCount(account.MaxAllowedRecords)
			var usedRecordCount uint64
			if allowedRecordCount != nil && *allowedRecordCount > 0 {
				count, err := s.getUsedRecordCountForMonth(ctx, req.Msg.GetAccountId(), logger)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve used record count for month: %w", err)
				}
				usedRecordCount = count
			}
			return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
				UsedRecordCount:    usedRecordCount,
				AllowedRecordCount: allowedRecordCount,
				SubscriptionStatus: 0,
			}), nil
		}
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
}

func getAllowedRecordCount(maxAllowed pgtype.Int8) *uint64 {
	if maxAllowed.Valid {
		val := toUint64(maxAllowed.Int64)
		return &val
	}
	return nil
}

// Returns the int64 as a uint64, or if int64 is negative, returns 0
func toUint64(val int64) uint64 {
	if val >= 0 {
		return uint64(val)
	}
	return 0
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
	if !s.cfg.IsNeosyncCloud {
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
	}

	// todo: sub payments subscription status check
	if accountStatusResp.Msg.AllowedRecordCount == nil {
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
	}

	currentUsed := accountStatusResp.Msg.GetUsedRecordCount()
	allowed := accountStatusResp.Msg.GetAllowedRecordCount()

	if currentUsed >= allowed {
		reason := fmt.Sprintf("Current used record count (%d) exceeds the allowed limit (%d).", currentUsed, allowed)
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{
			IsValid: false,
			Reason:  &reason,
		}), nil
	}

	if req.Msg.RequestedRecordCount == nil {
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
	}

	requested := req.Msg.GetRequestedRecordCount()
	totalUsed := currentUsed + requested
	if totalUsed > allowed {
		reason := fmt.Sprintf("Adding requested record count (%d) would exceed the allowed limit (%d). Current used: %d.", requested, allowed, currentUsed)
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{
			IsValid: false,
			Reason:  &reason,
		}), nil
	}

	return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{
		IsValid: true,
	}), nil
}

func (s *Service) getUsedRecordCountForMonth(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (uint64, error) {
	today := time.Now().UTC()
	promWindow := fmt.Sprintf("%dd", today.Day())

	queryLabels := metrics.MetricLabels{
		metrics.NewNotEqLabel(metrics.IsUpdateConfigLabel, "true"), // we want to always exclude update configs
		metrics.NewEqLabel(metrics.AccountIdLabel, accountId),
	}

	query, err := metrics.GetPromQueryFromMetric(mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED, queryLabels, promWindow)
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
