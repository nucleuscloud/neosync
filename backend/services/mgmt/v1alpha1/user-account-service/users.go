package v1alpha1_useraccountservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/auth/tokenctx"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/version"
	"github.com/nucleuscloud/neosync/internal/billing"
	"github.com/stripe/stripe-go/v79"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) GetUser(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetUserRequest],
) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
	if !s.cfg.IsAuthEnabled {
		// intentionally ignoring error here because we are in unauth mode anyways
		// but if it's available, let's return the api key's user id
		apiTokenCtxData, _ := auth_apikey.GetTokenDataFromCtx(ctx)
		if apiTokenCtxData != nil {
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
				UserId: neosyncdb.UUIDString(apiTokenCtxData.ApiKey.UserID),
			}), nil
		}
		user, err := s.db.Q.GetAnonymousUser(ctx, s.db.Db)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return nil, nucleuserrors.New(err)
		} else if err != nil && neosyncdb.IsNoRows(err) {
			user, err = s.db.Q.SetAnonymousUser(ctx, s.db.Db)
			if err != nil {
				return nil, err
			}
		}
		return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
			UserId: neosyncdb.UUIDString(user.ID),
		}), nil
	}

	tokenctxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return nil, err
	}

	if tokenctxResp.ApiKeyContextData != nil {
		if tokenctxResp.ApiKeyContextData.ApiKeyType == apikey.AccountApiKey && tokenctxResp.ApiKeyContextData.ApiKey != nil {
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
				UserId: neosyncdb.UUIDString(tokenctxResp.ApiKeyContextData.ApiKey.UserID),
			}), nil
		}
		return nil, nucleuserrors.NewUnauthenticated(fmt.Sprintf("invalid api key type when calling GetUser: %s", tokenctxResp.ApiKeyContextData.ApiKeyType))
	} else if tokenctxResp.JwtContextData != nil {
		user, err := s.db.Q.GetUserAssociationByProviderSub(ctx, s.db.Db, tokenctxResp.JwtContextData.AuthUserId)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return nil, nucleuserrors.New(err)
		} else if err != nil && neosyncdb.IsNoRows(err) {
			return nil, nucleuserrors.NewNotFound("unable to find user")
		}

		return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
			UserId: neosyncdb.UUIDString(user.UserID),
		}), nil
	}
	return nil, nucleuserrors.NewUnauthenticated("unable to find a valid user based on the provided auth credentials")
}

func (s *Service) SetUser(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetUserRequest],
) (*connect.Response[mgmtv1alpha1.SetUserResponse], error) {
	if !s.cfg.IsAuthEnabled {
		// intentionally ignoring error here because we are in unauth mode anyways
		// but if it's available, let's return the api key's user id
		apiTokenCtxData, _ := auth_apikey.GetTokenDataFromCtx(ctx)
		if apiTokenCtxData != nil {
			return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
				UserId: neosyncdb.UUIDString(apiTokenCtxData.ApiKey.UserID),
			}), nil
		}
		user, err := s.db.Q.SetAnonymousUser(ctx, s.db.Db)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
			UserId: neosyncdb.UUIDString(user.ID),
		}), nil
	}

	tokenctxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return nil, err
	}
	if tokenctxResp.ApiKeyContextData != nil {
		return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
			UserId: neosyncdb.UUIDString(tokenctxResp.ApiKeyContextData.ApiKey.UserID),
		}), nil
	} else if tokenctxResp.JwtContextData != nil {
		tokenCtxData, err := authjwt.GetTokenDataFromCtx(ctx)
		if err != nil {
			return nil, nucleuserrors.New(err)
		}

		user, err := s.db.SetUserByAuthSub(ctx, tokenCtxData.AuthUserId)
		if err != nil {
			return nil, nucleuserrors.New(err)
		}

		return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
			UserId: neosyncdb.UUIDString(user.ID),
		}), nil
	}
	return nil, nucleuserrors.NewUnauthenticated("unable to find a valid user based on the provided auth credentials")
}

func (s *Service) GetUserAccounts(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetUserAccountsRequest],
) (*connect.Response[mgmtv1alpha1.GetUserAccountsResponse], error) {
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userId, err := neosyncdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}
	accounts, err := s.db.Q.GetAccountsByUser(ctx, s.db.Db, userId)
	if err != nil {
		return nil, err
	}

	dtoAccounts := []*mgmtv1alpha1.UserAccount{}
	for index := range accounts {
		dtoAccounts = append(dtoAccounts, dtomaps.ToUserAccount(&accounts[index]))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetUserAccountsResponse{
		Accounts: dtoAccounts,
	}), nil
}

func (s *Service) ConvertPersonalToTeamAccount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.ConvertPersonalToTeamAccountRequest],
) (*connect.Response[mgmtv1alpha1.ConvertPersonalToTeamAccountResponse], error) {
	if !s.cfg.IsAuthEnabled {
		return nil, nucleuserrors.NewForbidden("unable to convert personal account to team account as authentication is not enabled")
	}
	if s.cfg.IsNeosyncCloud && s.billingclient == nil {
		return nil, nucleuserrors.NewForbidden("creating team accounts via the API is currently forbidden in Neosync Cloud environments. Please contact us to create a team account.")
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userId, err := neosyncdb.ToUuid(user.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	personalAccountId := req.Msg.GetAccountId()
	if personalAccountId == "" {
		logger.Debug("account id was not provided during personal->team conversion. Attempting to find personal account")
		accounts, err := s.db.Q.GetAccountsByUser(ctx, s.db.Db, userId)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return nil, err
		} else if err != nil && neosyncdb.IsNoRows(err) {
			return nil, nucleuserrors.NewNotFound("user has no accounts")
		}

		for idx := range accounts {
			if accounts[idx].AccountType == int16(neosyncdb.AccountType_Personal) {
				personalAccountId = neosyncdb.UUIDString(accounts[idx].ID)
				logger.Debug("found personal account to convert to team account", "personalAccountId", personalAccountId)
				break
			}
		}
	}

	personalAccountUuid, err := neosyncdb.ToUuid(personalAccountId)
	if err != nil {
		return nil, err
	}
	resp, err := s.db.ConvertPersonalToTeamAccount(ctx, &neosyncdb.ConvertPersonalToTeamAccountRequest{
		UserId:            userId,
		PersonalAccountId: personalAccountUuid,
		TeamName:          req.Msg.GetName(),
	}, logger)
	if err != nil {
		return nil, err
	}

	var checkoutSessionUrl *string
	if s.cfg.IsNeosyncCloud && !resp.TeamAccount.StripeCustomerID.Valid && s.billingclient != nil {
		account, err := s.db.UpsertStripeCustomerId(
			ctx,
			resp.TeamAccount.ID,
			s.getCreateStripeAccountFunction(user.Msg.GetUserId(), logger),
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to upsert stripe customer id after account creation: %w", err)
		}
		session, err := s.generateCheckoutSession(account.StripeCustomerID.String, account.AccountSlug, user.Msg.GetUserId(), logger)
		if err != nil {
			return nil, fmt.Errorf("unable to generate checkout session: %w", err)
		}
		logger.Debug("stripe checkout session created", "id", session.ID)
		checkoutSessionUrl = &session.URL
		resp.TeamAccount = account // update the team account that now includes a stripe customer id
	}

	return connect.NewResponse(&mgmtv1alpha1.ConvertPersonalToTeamAccountResponse{
		AccountId:            neosyncdb.UUIDString(resp.TeamAccount.ID),
		NewPersonalAccountId: neosyncdb.UUIDString(resp.PersonalAccount.ID),
		CheckoutSessionUrl:   checkoutSessionUrl,
	}), nil
}

func (s *Service) SetPersonalAccount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetPersonalAccountRequest],
) (*connect.Response[mgmtv1alpha1.SetPersonalAccountResponse], error) {
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}

	userId, err := neosyncdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	account, err := s.db.SetPersonalAccount(ctx, userId, s.cfg.DefaultMaxAllowedRecords)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetPersonalAccountResponse{
		AccountId: neosyncdb.UUIDString(account.ID),
	}), nil
}

func (s *Service) IsUserInAccount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsUserInAccountRequest],
) (*connect.Response[mgmtv1alpha1.IsUserInAccountResponse], error) {
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}

	userId, err := neosyncdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}
	accountId, err := neosyncdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	apiKeyCount, err := s.db.Q.IsUserInAccountApiKey(ctx, s.db.Db, db_queries.IsUserInAccountApiKeyParams{
		AccountId: accountId,
		UserId:    userId,
	})
	if err != nil {
		return nil, err
	}
	if apiKeyCount > 0 {
		return connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
			Ok: apiKeyCount > 0,
		}), nil
	}
	count, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
		AccountId: accountId,
		UserId:    userId,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: count > 0,
	}), nil
}

func (s *Service) CreateTeamAccount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateTeamAccountRequest],
) (*connect.Response[mgmtv1alpha1.CreateTeamAccountResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	if !s.cfg.IsAuthEnabled {
		return nil, nucleuserrors.NewForbidden("unable to create team account as authentication is not enabled")
	}
	if s.cfg.IsNeosyncCloud && s.billingclient == nil {
		return nil, nucleuserrors.NewForbidden("creating team accounts via the API is currently forbidden in Neosync Cloud environments. Please contact us to create a team account.")
	}

	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userId, err := neosyncdb.ToUuid(user.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.CreateTeamAccount(ctx, userId, req.Msg.GetName(), logger)
	if err != nil {
		return nil, err
	}

	var checkoutSessionUrl *string
	if s.cfg.IsNeosyncCloud && !account.StripeCustomerID.Valid && s.billingclient != nil {
		account, err = s.db.UpsertStripeCustomerId(
			ctx,
			account.ID,
			s.getCreateStripeAccountFunction(user.Msg.GetUserId(), logger),
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to upsert stripe customer id after account creation: %w", err)
		}
		session, err := s.generateCheckoutSession(account.StripeCustomerID.String, account.AccountSlug, user.Msg.GetUserId(), logger)
		if err != nil {
			return nil, fmt.Errorf("unable to generate checkout session: %w", err)
		}
		logger.Debug("stripe checkout session created", "id", session.ID)
		checkoutSessionUrl = &session.URL
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateTeamAccountResponse{
		AccountId:          neosyncdb.UUIDString(account.ID),
		CheckoutSessionUrl: checkoutSessionUrl,
	}), nil
}

func (s *Service) getCreateStripeAccountFunction(userId string, logger *slog.Logger) func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
	return func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
		email := s.getEmailFromToken(ctx, logger)
		if email == nil {
			return "", errors.New("unable to retrieve user email from auth token when creating stripe account")
		}
		customer, err := s.billingclient.NewCustomer(&billing.CustomerRequest{
			Email:     *email,
			Name:      account.AccountSlug,
			AccountId: neosyncdb.UUIDString(account.ID),
			UserId:    userId,
		})
		if err != nil {
			return "", fmt.Errorf("unable to create new stripe customer: %w", err)
		}
		return customer.ID, nil
	}
}

func (s *Service) generateCheckoutSession(customerId, accountSlug, userId string, logger *slog.Logger) (*stripe.CheckoutSession, error) {
	if s.billingclient == nil {
		return nil, errors.New("unable to generate checkout session as stripe client is nil")
	}

	session, err := s.billingclient.NewCheckoutSession(customerId, accountSlug, userId, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new stripe checkout session: %w", err)
	}
	return session, nil
}

func (s *Service) getEmailFromToken(ctx context.Context, logger *slog.Logger) *string {
	tokenctxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve token from ctx when getting email: %w", err).Error())
		return nil
	}
	if tokenctxResp.JwtContextData != nil && tokenctxResp.JwtContextData.Claims != nil {
		return tokenctxResp.JwtContextData.Claims.Email
	}
	logger.Error(errors.New("unable to retrieve email from token ctx").Error())
	return nil
}

func (s *Service) GetTeamAccountMembers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTeamAccountMembersRequest],
) (*connect.Response[mgmtv1alpha1.GetTeamAccountMembersResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeamAccount(ctx, *accountId); err != nil {
		return nil, err
	}

	userIdentities, err := s.db.Q.GetUserIdentitiesByTeamAccount(ctx, s.db.Db, *accountId)
	if err != nil {
		return nil, err
	}

	dtoUsers := make([]*mgmtv1alpha1.AccountUser, len(userIdentities))
	group := new(errgroup.Group)
	for i := range userIdentities {
		i := i
		user := userIdentities[i]
		group.Go(func() error {
			dtoUsers[i] = &mgmtv1alpha1.AccountUser{
				Id: neosyncdb.UUIDString(user.UserID),
			}
			if user.ProviderSub == "" {
				logger.Warn(fmt.Sprintf("unable to find provider sub associated with user id: %q", neosyncdb.UUIDString(user.UserID)))
				return nil
			}
			if user.ProviderSub != "" {
				authuser, err := s.authadminclient.GetUserBySub(ctx, user.ProviderSub)
				if err != nil {
					logger.Warn(fmt.Sprintf("unable to retrieve user by sub: %s", err.Error()))
				} else {
					dtoUsers[i].Email = authuser.Email
					dtoUsers[i].Name = authuser.Name
					dtoUsers[i].Image = authuser.Picture
				}
			}
			return nil
		})
	}
	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetTeamAccountMembersResponse{
		Users: dtoUsers,
	}), nil
}

func (s *Service) RemoveTeamAccountMember(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RemoveTeamAccountMemberRequest],
) (*connect.Response[mgmtv1alpha1.RemoveTeamAccountMemberResponse], error) {
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	if err := s.verifyTeamAccount(ctx, *accountId); err != nil {
		return nil, err
	}
	memberUserId, err := neosyncdb.ToUuid(req.Msg.UserId)
	if err != nil {
		return nil, err
	}
	err = s.db.Q.RemoveAccountUser(ctx, s.db.Db, db_queries.RemoveAccountUserParams{
		AccountId: *accountId,
		UserId:    memberUserId,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.RemoveTeamAccountMemberResponse{}), nil
}

func (s *Service) InviteUserToTeamAccount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.InviteUserToTeamAccountRequest],
) (*connect.Response[mgmtv1alpha1.InviteUserToTeamAccountResponse], error) {
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userId, err := neosyncdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	accountId, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeamAccount(ctx, *accountId); err != nil {
		return nil, err
	}

	tomorrow := time.Now().Add(24 * time.Hour)
	expiresAt, err := neosyncdb.ToTimestamp(tomorrow)
	if err != nil {
		return nil, err
	}

	invite, err := s.db.CreateTeamAccountInvite(ctx, *accountId, userId, req.Msg.Email, expiresAt)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.InviteUserToTeamAccountResponse{
		Invite: dtomaps.ToAccountInviteDto(invite),
	}), nil
}

func (s *Service) GetTeamAccountInvites(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTeamAccountInvitesRequest],
) (*connect.Response[mgmtv1alpha1.GetTeamAccountInvitesResponse], error) {
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeamAccount(ctx, *accountId); err != nil {
		return nil, err
	}

	invites, err := s.db.Q.GetActiveAccountInvites(ctx, s.db.Db, *accountId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetTeamAccountInvitesResponse{
			Invites: []*mgmtv1alpha1.AccountInvite{},
		}), nil
	}

	dtos := []*mgmtv1alpha1.AccountInvite{}
	for index := range invites {
		dtos = append(dtos, dtomaps.ToAccountInviteDto(&invites[index]))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetTeamAccountInvitesResponse{
		Invites: dtos,
	}), nil
}

func (s *Service) RemoveTeamAccountInvite(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RemoveTeamAccountInviteRequest],
) (*connect.Response[mgmtv1alpha1.RemoveTeamAccountInviteResponse], error) {
	inviteId, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	invite, err := s.db.Q.GetAccountInvite(ctx, s.db.Db, inviteId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.RemoveTeamAccountInviteResponse{}), nil
	}
	accountId, err := s.verifyUserInAccount(ctx, neosyncdb.UUIDString(invite.AccountID))
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeamAccount(ctx, *accountId); err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveAccountInvite(ctx, s.db.Db, inviteId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.RemoveTeamAccountInviteResponse{}), nil
	}

	return connect.NewResponse(&mgmtv1alpha1.RemoveTeamAccountInviteResponse{}), nil
}

func (s *Service) AcceptTeamAccountInvite(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.AcceptTeamAccountInviteRequest],
) (*connect.Response[mgmtv1alpha1.AcceptTeamAccountInviteResponse], error) {
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userUuid, err := neosyncdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	tokenctxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return nil, err
	}
	if tokenctxResp.JwtContextData == nil {
		return nil, nucleuserrors.NewUnauthenticated("must be a valid jwt user to accept team account invites")
	}

	var email *string
	if tokenctxResp.JwtContextData.Claims != nil && tokenctxResp.JwtContextData.Claims.Email != nil {
		email = tokenctxResp.JwtContextData.Claims.Email
	} else {
		userinfo, err := s.authclient.GetUserInfo(ctx, tokenctxResp.JwtContextData.RawToken)
		if err != nil {
			return nil, err
		}
		// should we check if email is verified here? maybe in the future
		if userinfo.Email == "" {
			return nil, nucleuserrors.NewInternalError("retrieved user info but email was not present")
		}
		email = &userinfo.Email
	}
	if email == nil {
		return nil, nucleuserrors.NewUnauthenticated("unable to find email to valid to add user to account")
	}

	accountId, err := s.db.ValidateInviteAddUserToAccount(ctx, userUuid, req.Msg.Token, *email)
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeamAccount(ctx, accountId); err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.AcceptTeamAccountInviteResponse{
		Account: dtomaps.ToUserAccount(&account),
	}), nil
}

func (s *Service) verifyTeamAccount(ctx context.Context, accountId pgtype.UUID) error {
	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountId)
	if err != nil {
		return err
	}
	if account.AccountType != 1 {
		return nucleuserrors.NewForbidden("account is not a team account")
	}
	return nil
}
func isWorkerApiKey(ctx context.Context) bool {
	data, err := auth_apikey.GetTokenDataFromCtx(ctx)
	if err != nil {
		return false
	}
	return data.ApiKeyType == apikey.WorkerApiKey
}

func (s *Service) verifyUserInAccount(
	ctx context.Context,
	accountId string,
) (*pgtype.UUID, error) {
	accountUuid, err := neosyncdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}

	if isWorkerApiKey(ctx) {
		return &accountUuid, nil
	}

	resp, err := s.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return nil, err
	}
	if !resp.Msg.Ok {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}

	return &accountUuid, nil
}

func (s *Service) GetSystemInformation(ctx context.Context, req *connect.Request[mgmtv1alpha1.GetSystemInformationRequest]) (*connect.Response[mgmtv1alpha1.GetSystemInformationResponse], error) {
	versionInfo := version.Get()
	builtDate, err := time.Parse(time.RFC3339, versionInfo.BuildDate)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.GetSystemInformationResponse{
		Version:   versionInfo.GitVersion,
		Commit:    versionInfo.GitCommit,
		Compiler:  versionInfo.Compiler,
		Platform:  versionInfo.Platform,
		BuildDate: timestamppb.New(builtDate),
	}), nil
}
