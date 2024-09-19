package v1alpha1_useraccountservice

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/stripe/stripe-go/v79"
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
	logger = logger.With("accountId", accountId)

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
				SubscriptionStatus: mgmtv1alpha1.BillingStatus_BILLING_STATUS_UNSPECIFIED,
			}), nil
		} else if s.stripeclient != nil {
			if !account.StripeCustomerID.Valid {
				logger.Warn("stripe is enabled but team account does not have stripe customer id")
				return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
					SubscriptionStatus: mgmtv1alpha1.BillingStatus_BILLING_STATUS_UNSPECIFIED,
				}), nil
			}
			logger.Debug("attempting to find active stripe subscription")
			subscription, err := s.findActiveStripeSubscription(account.StripeCustomerID.String)
			if err != nil {
				return nil, err
			}
			if subscription == nil {
				return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
					SubscriptionStatus: mgmtv1alpha1.BillingStatus_BILLING_STATUS_EXPIRED,
				}), nil
			}
			return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
				SubscriptionStatus: mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE,
			}), nil
		}
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
}

func (s *Service) findActiveStripeSubscription(customerId string) (*stripe.Subscription, error) {
	subIter := s.stripeclient.Subscriptions.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(customerId),
	})
	var validSubscription *stripe.Subscription
	now := time.Now().UTC().Unix()

	for subIter.Next() {
		subscription := subIter.Subscription()
		if isSubscriptionActive(subscription.Status) && subscription.CurrentPeriodStart <= now && subscription.CurrentPeriodEnd > now {
			validSubscription = subscription
			break
		}
	}
	// this could be bad, we may want to cache the stripe subscriptions to prevent any issues with stripe outages
	if subIter.Err() != nil {
		return nil, fmt.Errorf("encountered error when retrieving stripe subscriptions: %w", subIter.Err())
	}
	return validSubscription, nil
}

func isSubscriptionActive(status stripe.SubscriptionStatus) bool {
	switch status {
	case stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusTrialing:
		return true
	case stripe.SubscriptionStatusPastDue,
		stripe.SubscriptionStatusIncomplete:
		// You might want to add a grace period for past_due or incomplete statuses
		// This could be based on the number of days past due or other criteria
		return true
	case stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusIncompleteExpired,
		stripe.SubscriptionStatusUnpaid,
		stripe.SubscriptionStatusPaused:
		return false
	default:
		// If an unknown status is encountered, default to inactive for safety
		return false
	}
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

	accountUuid, err := nucleusdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account: %w", err)
	}

	if account.AccountType == int16(dtomaps.AccountType_Personal) {
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

	billingStatus := accountStatusResp.Msg.GetSubscriptionStatus()

	if s.stripeclient == nil || billingStatus == mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE {
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
	}

	reason := "Account is currently in expired state, visit the billing page to activate your subscription."
	return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{
		IsValid: false,
		Reason:  &reason,
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

func (s *Service) GetAccountBillingPortalSession(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountBillingPortalSessionRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountBillingPortalSessionResponse], error) {
	if !s.cfg.IsNeosyncCloud || s.stripeclient == nil {
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("%s is not implemented", strings.TrimPrefix(mgmtv1alpha1connect.UserAccountServiceGetAccountBillingPortalSessionProcedure, "/")))
	}
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, *accountId)
	if err != nil {
		return nil, err
	}
	if !account.StripeCustomerID.Valid {
		return nil, nucleuserrors.NewForbidden("requested account does not have a valid stripe customer id")
	}

	session, err := s.stripeclient.BillingPortalSessions.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(account.StripeCustomerID.String),
		ReturnURL: stripe.String(fmt.Sprintf("%s/%s/settings/billing", s.cfg.AppBaseUrl, account.AccountSlug)),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to generate billing portal session: %w", err)
	}
	return connect.NewResponse(&mgmtv1alpha1.GetAccountBillingPortalSessionResponse{
		PortalSessionUrl: session.URL,
	}), nil
}
