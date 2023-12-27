package v1alpha1_useraccountservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/auth/tokenctx"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/internal/version"
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
				UserId: nucleusdb.UUIDString(apiTokenCtxData.ApiKey.UserID),
			}), nil
		}
		user, err := s.db.Q.GetAnonymousUser(ctx, s.db.Db)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, nucleuserrors.New(err)
		} else if err != nil && nucleusdb.IsNoRows(err) {
			user, err = s.db.Q.SetAnonymousUser(ctx, s.db.Db)
			if err != nil {
				return nil, err
			}
		}
		return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
			UserId: nucleusdb.UUIDString(user.ID),
		}), nil
	}

	tokenctxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return nil, err
	}

	if tokenctxResp.ApiKeyContextData != nil {
		return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
			UserId: nucleusdb.UUIDString(tokenctxResp.ApiKeyContextData.ApiKey.UserID),
		}), nil
	} else if tokenctxResp.JwtContextData != nil {
		user, err := s.db.Q.GetUserAssociationByProviderSub(ctx, s.db.Db, tokenctxResp.JwtContextData.AuthUserId)
		if err != nil && !nucleusdb.IsNoRows(err) {
			return nil, nucleuserrors.New(err)
		} else if err != nil && nucleusdb.IsNoRows(err) {
			return nil, nucleuserrors.NewNotFound("unable to find user")
		}

		return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
			UserId: nucleusdb.UUIDString(user.UserID),
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
				UserId: nucleusdb.UUIDString(apiTokenCtxData.ApiKey.UserID),
			}), nil
		}
		user, err := s.db.Q.SetAnonymousUser(ctx, s.db.Db)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
			UserId: nucleusdb.UUIDString(user.ID),
		}), nil
	}

	tokenctxResp, err := tokenctx.GetTokenCtx(ctx)
	if err != nil {
		return nil, err
	}
	if tokenctxResp.ApiKeyContextData != nil {
		return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
			UserId: nucleusdb.UUIDString(tokenctxResp.ApiKeyContextData.ApiKey.UserID),
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
			UserId: nucleusdb.UUIDString(user.ID),
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
	userId, err := nucleusdb.ToUuid(user.Msg.UserId)
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

	return connect.NewResponse(&mgmtv1alpha1.ConvertPersonalToTeamAccountResponse{}), nil
}

func (s *Service) SetPersonalAccount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetPersonalAccountRequest],
) (*connect.Response[mgmtv1alpha1.SetPersonalAccountResponse], error) {
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}

	userId, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	account, err := s.db.SetPersonalAccount(ctx, userId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetPersonalAccountResponse{
		AccountId: nucleusdb.UUIDString(account.ID),
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

	userId, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}
	accountId, err := nucleusdb.ToUuid(req.Msg.AccountId)
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
	user, err := s.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userId, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	account, err := s.db.CreateTeamAccount(ctx, userId, req.Msg.Name)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateTeamAccountResponse{
		AccountId: nucleusdb.UUIDString(account.ID),
	}), nil
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
			if user.ProviderSub != "" {
				authUser, err := s.auth0MgmtClient.GetUserById(ctx, user.ProviderSub)
				if err != nil {
					// if unable to get auth user still return user id
					dtoUsers[i] = &mgmtv1alpha1.AccountUser{
						Id: nucleusdb.UUIDString(user.UserID),
					}
					return err
				}
				dtoUsers[i] = &mgmtv1alpha1.AccountUser{
					Id:    nucleusdb.UUIDString(user.UserID),
					Name:  authUser.GetName(),
					Email: authUser.GetEmail(),
					Image: authUser.GetPicture(),
				}
			} else {
				return errors.New("unsupported or no provider id for user identity")
			}
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		logger.Error(fmt.Errorf("unable to hydrate auth user: %w", err).Error())
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
	memberUserId, err := nucleusdb.ToUuid(req.Msg.UserId)
	if err != nil {
		return nil, err
	}
	err = s.db.Q.RemoveAccountUser(ctx, s.db.Db, db_queries.RemoveAccountUserParams{
		AccountId: *accountId,
		UserId:    memberUserId,
	})
	if err != nil && !nucleusdb.IsNoRows(err) {
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
	userId, err := nucleusdb.ToUuid(user.Msg.UserId)
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
	expiresAt, err := nucleusdb.ToTimestamp(tomorrow)
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
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && nucleusdb.IsNoRows(err) {
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
	inviteId, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	invite, err := s.db.Q.GetAccountInvite(ctx, s.db.Db, inviteId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.RemoveTeamAccountInviteResponse{}), nil
	}
	accountId, err := s.verifyUserInAccount(ctx, nucleusdb.UUIDString(invite.AccountID))
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeamAccount(ctx, *accountId); err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveAccountInvite(ctx, s.db.Db, inviteId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && nucleusdb.IsNoRows(err) {
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
	userUuid, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}

	userIdentity, err := s.db.Q.GetUserIdentityByUserId(ctx, s.db.Db, userUuid)
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
		// check auth user email to invite email
		// todo: remove the need for this by adding the email to the auth0 jwt, if it's not already there.
		authUser, err := s.auth0MgmtClient.GetUserById(ctx, userIdentity.ProviderSub)
		if err != nil {
			return nil, err
		}
		authUserEmail := authUser.GetEmail()
		email = &authUserEmail
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

func (s *Service) verifyUserInAccount(
	ctx context.Context,
	accountId string,
) (*pgtype.UUID, error) {
	resp, err := s.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return nil, err
	}
	if !resp.Msg.Ok {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}

	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return nil, err
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
