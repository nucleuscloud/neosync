package v1alpha1_useraccountservice

import (
	"context"
	"time"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

func (s *Service) GetUser(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetUserRequest],
) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {

	if !s.cfg.IsAuthEnabled {
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

	tokenCtxData, err := authjwt.GetTokenDataFromCtx(ctx)
	if err != nil {
		return nil, nucleuserrors.New(err)
	}

	user, err := s.db.Q.GetUserAssociationByAuth0Id(ctx, s.db.Db, tokenCtxData.AuthUserId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.New(err)
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find user")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: nucleusdb.UUIDString(user.UserID),
	}), nil
}

func (s *Service) SetUser(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetUserRequest],
) (*connect.Response[mgmtv1alpha1.SetUserResponse], error) {
	if !s.cfg.IsAuthEnabled {
		user, err := s.db.Q.SetAnonymousUser(ctx, s.db.Db)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
			UserId: nucleusdb.UUIDString(user.ID),
		}), nil
	}

	tokenCtxData, err := authjwt.GetTokenDataFromCtx(ctx)
	if err != nil {
		return nil, nucleuserrors.New(err)
	}

	user, err := s.db.SetUserByAuth0Id(ctx, tokenCtxData.AuthUserId)
	if err != nil {
		return nil, nucleuserrors.New(err)
	}

	return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{
		UserId: nucleusdb.UUIDString(user.ID),
	}), nil
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
	for _, account := range accounts {
		dtoAccounts = append(dtoAccounts, &mgmtv1alpha1.UserAccount{
			Id:   nucleusdb.UUIDString(account.ID),
			Name: account.AccountSlug,
			Type: dtomaps.ToAccountTypeDto(account.AccountType),
		})
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
	count, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
		AccountId: accountId,
		UserId:    userId,
	})
	if count == 0 {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}
	users, err := s.db.Q.GetUsersByTeamAccount(ctx, s.db.Db, accountId)
	if err != nil {
		return nil, err
	}

	dtoUsers := []*mgmtv1alpha1.AccountUser{}
	for _, user := range users {
		dtoUsers = append(dtoUsers, &mgmtv1alpha1.AccountUser{
			Id: nucleusdb.UUIDString(user.ID),
		})
	}

	return connect.NewResponse(&mgmtv1alpha1.GetTeamAccountMembersResponse{
		Users: dtoUsers,
	}), nil
}

func (s *Service) RemoveTeamAccountMember(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RemoveTeamAccountMemberRequest],
) (*connect.Response[mgmtv1alpha1.RemoveTeamAccountMemberResponse], error) {
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
	count, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
		AccountId: accountId,
		UserId:    userId,
	})
	if count == 0 {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}
	memberUserId, err := nucleusdb.ToUuid(req.Msg.UserId)
	if err != nil {
		return nil, err
	}
	err = s.db.Q.RemoveTeamAccountUser(ctx, s.db.Db, db_queries.RemoveTeamAccountUserParams{
		AccountId: accountId,
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
	accountId, err := nucleusdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	count, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
		AccountId: accountId,
		UserId:    userId,
	})
	if count == 0 {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	expiresAt, err := nucleusdb.ToTimestamp(tomorrow)
	if err != nil {
		return nil, err
	}

	invite, err := s.db.Q.CreateTeamAccountInvite(ctx, s.db.Db, db_queries.CreateTeamAccountInviteParams{
		AccountID:    accountId,
		SenderUserID: userId,
		Email:        req.Msg.Email,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.InviteUserToTeamAccountResponse{
		Invite: dtomaps.ToAccountInviteDto(&invite),
	}), nil
}
