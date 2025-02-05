package v1alpha1_useraccountservice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/internal/billing"
	"github.com/stripe/stripe-go/v81"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// 14 days duration
	trialDuration = 14 * 24 * time.Hour
)

func (s *Service) GetAccountStatus(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountStatusRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountStatusResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	userdataclient := s.UserDataClient()
	user, err := userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceAccount(ctx, userdata.NewIdentifier(req.Msg.GetAccountId()), rbac.AccountAction_View)
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	logger = logger.With("accountId", req.Msg.GetAccountId())
	if !s.cfg.IsNeosyncCloud || s.billingclient == nil {
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{}), nil
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account: %w", err)
	}

	trialStatus := getTrialStatus(account.CreatedAt)

	if account.AccountType == int16(neosyncdb.AccountType_Personal) {
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
			SubscriptionStatus: trialStatus,
		}), nil
	}
	if !account.StripeCustomerID.Valid {
		logger.Warn("stripe is enabled but team account does not have stripe customer id")
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
			SubscriptionStatus: trialStatus,
		}), nil
	}

	logger.Debug("attempting to find active stripe subscription")
	subscriptions, err := s.getStripeSubscriptions(account.StripeCustomerID.String)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("found %d stripe subscriptions for account", len(subscriptions)))
	_, hasActiveSub := findActiveStripeSubscription(subscriptions)
	if hasActiveSub {
		logger.Debug("account has active billing subscription")
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
			SubscriptionStatus: mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE,
		}), nil
	}
	if len(subscriptions) == 0 {
		logger.Debug("account has no subscriptions, returning trial status")
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
			SubscriptionStatus: trialStatus,
		}), nil
	}
	if trialStatus == mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_ACTIVE {
		logger.Debug("account has no active subscriptions but trial is still active")
		return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
			SubscriptionStatus: trialStatus,
		}), nil
	}
	logger.Debug("account has no active subscriptions and trial is expired")
	return connect.NewResponse(&mgmtv1alpha1.GetAccountStatusResponse{
		SubscriptionStatus: mgmtv1alpha1.BillingStatus_BILLING_STATUS_EXPIRED,
	}), nil
}

func getTrialStatus(ts pgtype.Timestamp) mgmtv1alpha1.BillingStatus {
	if !ts.Valid || ts.Time.IsZero() {
		return mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_EXPIRED
	}

	trialEndTime := ts.Time.Add(trialDuration)
	trialActive := time.Now().UTC().Before(trialEndTime)

	if trialActive {
		return mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_ACTIVE
	}
	return mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_EXPIRED
}

func (s *Service) getStripeSubscriptions(customerId string) ([]*stripe.Subscription, error) {
	subIter := s.billingclient.GetSubscriptions(customerId)
	output := []*stripe.Subscription{}
	for subIter.Next() {
		output = append(output, subIter.Subscription())
	}
	if subIter.Err() != nil {
		return nil, fmt.Errorf("encountered error when retrieving stripe subscriptions: %w", subIter.Err())
	}
	return output, nil
}

func findActiveStripeSubscription(subs []*stripe.Subscription) (*stripe.Subscription, bool) {
	for _, sub := range subs {
		if isSubscriptionActive(sub.Status) {
			return sub, true
		}
	}
	return nil, false
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

	if !s.cfg.IsNeosyncCloud || s.billingclient == nil {
		return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
	}

	accountStatus := mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_REASON_UNSPECIFIED
	var description string
	isValid := false

	var trialExpiryDate *timestamppb.Timestamp

	switch accountStatusResp.Msg.GetSubscriptionStatus() {
	case mgmtv1alpha1.BillingStatus_BILLING_STATUS_EXPIRED:
		accountStatus = mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_IN_EXPIRED_STATE
		description = "Account is currently in expired state, visit the billing page to activate your subscription."
	case mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_EXPIRED:
		accountStatus = mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_TRIAL_EXPIRED
		description = "The trial period has ended, visit the billing page to activate your subscription."
	case mgmtv1alpha1.BillingStatus_BILLING_STATUS_TRIAL_ACTIVE:
		accountStatus = mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_ACCOUNT_TRIAL_ACTIVE
		isValid = true

		accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
		if err != nil {
			return nil, err
		}

		acc, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
		if err != nil {
			return nil, err
		}

		expiryTime := acc.CreatedAt.Time.Add(trialDuration)
		trialExpiryDate = timestamppb.New(expiryTime)
	case mgmtv1alpha1.BillingStatus_BILLING_STATUS_ACTIVE:
		accountStatus = mgmtv1alpha1.AccountStatus_ACCOUNT_STATUS_REASON_UNSPECIFIED
		isValid = true
	}
	return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{
		IsValid:        isValid,
		AccountStatus:  accountStatus,
		Reason:         &description,
		TrialExpiresAt: trialExpiryDate,
	}), nil
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
	userdataclient := s.UserDataClient()
	user, err := userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	err = user.EnforceAccount(ctx, userdata.NewIdentifier(req.Msg.GetAccountId()), rbac.AccountAction_Edit)
	if err != nil {
		return nil, err
	}

	// retrieve the account, creates a customer id if one doesn't already exist
	account, err := s.db.UpsertStripeCustomerId(
		ctx,
		accountUuid,
		s.getCreateStripeAccountFunction(user.Id(), logger),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("was unable to get account and/or upsert stripe customer id: %w", err)
	}
	if !account.StripeCustomerID.Valid {
		return nil, errors.New("stripe customer id does not exist on account after creation attempt")
	}

	session, err := s.generateCheckoutSession(account.StripeCustomerID.String, account.AccountSlug, user.Id(), logger)
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
	userdataclient := s.UserDataClient()
	user, err := userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	err = user.EnforceAccount(ctx, userdata.NewIdentifier(req.Msg.GetAccountId()), rbac.AccountAction_Edit)
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
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
	userdataclient := s.UserDataClient()
	user, err := userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if s.cfg.IsNeosyncCloud && !user.IsWorkerApiKey() {
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
	userdataclient := s.UserDataClient()
	user, err := userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if s.cfg.IsNeosyncCloud && !user.IsWorkerApiKey() {
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
