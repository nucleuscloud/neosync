package v1alpha1_useraccountservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/internal/billing"
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
		if account.AccountType == int16(neosyncdb.AccountType_Personal) {
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
		} else if s.billingclient != nil {
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
	subIter := s.billingclient.GetSubscriptions(customerId)
	var validSubscription *stripe.Subscription

	for subIter.Next() {
		subscription := subIter.Subscription()
		if isSubscriptionActive(subscription.Status) {
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

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account: %w", err)
	}

	if account.AccountType == int16(neosyncdb.AccountType_Personal) {
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

	if s.billingclient == nil || billingStatus == mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE {
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

func (s *Service) GetAccountBillingCheckoutSession(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountBillingCheckoutSessionRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountBillingCheckoutSessionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	if !s.cfg.IsNeosyncCloud || s.billingclient == nil {
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("%s is not implemented", strings.TrimPrefix(mgmtv1alpha1connect.UserAccountServiceGetAccountBillingCheckoutSessionProcedure, "/")))
	}
	logger = logger.With("accountId", req.Msg.GetAccountId())
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}

	// retrieve the account, creates a customer id if one doesn't already exist
	account, err := s.db.UpsertStripeCustomerId(
		ctx,
		*accountId,
		s.getCreateStripeAccountFunction(user.Msg.GetUserId(), logger),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("was unable to get account and/or upsert stripe customer id: %w", err)
	}
	if !account.StripeCustomerID.Valid {
		return nil, errors.New("stripe customer id does not exist on account after creation attempt")
	}

	session, err := s.generateCheckoutSession(account.StripeCustomerID.String, account.AccountSlug, user.Msg.GetUserId(), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to generate billing checkout session: %w", err)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccountBillingCheckoutSessionResponse{
		CheckoutSessionUrl: session.URL,
	}), nil
}

func (s *Service) GetAccountBillingPortalSession(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountBillingPortalSessionRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountBillingPortalSessionResponse], error) {
	if !s.cfg.IsNeosyncCloud || s.billingclient == nil {
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

	session, err := s.billingclient.NewBillingPortalSession(account.StripeCustomerID.String, account.AccountSlug)
	if err != nil {
		return nil, fmt.Errorf("unable to generate billing portal session: %w", err)
	}
	return connect.NewResponse(&mgmtv1alpha1.GetAccountBillingPortalSessionResponse{
		PortalSessionUrl: session.URL,
	}), nil
}

func (s *Service) GetBillingAccounts(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetBillingAccountsRequest],
) (*connect.Response[mgmtv1alpha1.GetBillingAccountsResponse], error) {
	if s.cfg.IsNeosyncCloud && !isWorkerApiKey(ctx) {
		return nil, nucleuserrors.NewUnauthorized("must provide valid authentication credentials for this endpoint")
	}

	accountIdsToFilter := []pgtype.UUID{}
	for _, accountId := range req.Msg.GetAccountIds() {
		accountUuid, err := neosyncdb.ToUuid(accountId)
		if err != nil {
			return nil, fmt.Errorf("input did not contain entirely valid uuids: %w", err)
		}
		accountIdsToFilter = append(accountIdsToFilter, accountUuid)
	}

	accounts, err := s.db.Q.GetBilledAccounts(ctx, s.db.Db, accountIdsToFilter)
	if err != nil {
		return nil, err
	}

	dtos := make([]*mgmtv1alpha1.UserAccount, 0, len(accounts))
	for idx := range accounts {
		account := accounts[idx]
		dtos = append(dtos, dtomaps.ToUserAccount(&account))
	}
	return connect.NewResponse(&mgmtv1alpha1.GetBillingAccountsResponse{Accounts: dtos}), nil
}

func (s *Service) SetBillingMeterEvent(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetBillingMeterEventRequest],
) (*connect.Response[mgmtv1alpha1.SetBillingMeterEventResponse], error) {
	if s.billingclient == nil {
		return nil, nucleuserrors.NewUnauthorized("billing is not currently enabled")
	}
	if s.cfg.IsNeosyncCloud && !isWorkerApiKey(ctx) {
		return nil, nucleuserrors.NewUnauthorized("must provide valid authentication credentials for this endpoint")
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx).
		With(
			"accountId", req.Msg.GetAccountId(),
			"eventId", req.Msg.GetEventId(),
			"eventName", req.Msg.GetEventName(),
		)

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("account does not exist")
	}
	if !account.StripeCustomerID.Valid {
		return nil, nucleuserrors.NewBadRequest("account is not an active billed customer")
	}

	var ts *int64
	if req.Msg.GetTimestamp() > 0 {
		conv, err := safeUint64ToInt64(req.Msg.GetTimestamp())
		if err != nil {
			return nil, err
		}
		ts = &conv
	}
	_, err = s.billingclient.NewMeterEvent(&billing.MeterEventRequest{
		EventName:  req.Msg.GetEventName(),
		Identifier: req.Msg.GetEventId(),
		Timestamp:  ts,
		CustomerId: account.StripeCustomerID.String,
		Value:      req.Msg.GetValue(),
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			if stripeErr.Type == stripe.ErrorTypeInvalidRequest && strings.Contains(stripeErr.Msg, "An event already exists with identifier") {
				logger.Warn("unable to create new meter event, identifier already exists")
				return connect.NewResponse(&mgmtv1alpha1.SetBillingMeterEventResponse{}), nil
			}
			// todo: handle rate limits from stripe
		}
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetBillingMeterEventResponse{}), nil
}

func safeUint64ToInt64(value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("uint64 value %d overflows int64", value)
	}
	return int64(value), nil
}
