package v1alpha1_useraccountservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/auth0/go-auth0/management"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/auth0"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	anonymousUserId  = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider = "test-provider"
	mockUserId       = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId    = "5629813e-1a35-4874-922c-9827d85f0378"
)

// GetUser
func Test_GetUser_Anonymous(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	user := getUserMock(anonymousUserId)

	m.QuerierMock.On("GetAnonymousUser", context.Background(), mock.Anything).Return(user, nil)

	resp, err := m.Service.GetUser(context.Background(), &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "SetAnonymousUser", context.Background(), mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, anonymousUserId, resp.Msg.GetUserId())
}

func Test_GetUser_Anonymous_Error(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	var nilUser db_queries.NeosyncApiUser

	m.QuerierMock.On("GetAnonymousUser", context.Background(), mock.Anything).Return(nilUser, errors.New("some error"))

	resp, err := m.Service.GetUser(context.Background(), &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "SetAnonymousUser", context.Background(), mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_Service_GetUser_AuthDisabled_ApiKey(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	uuidStr := uuid.NewString()
	useruuid, _ := nucleusdb.ToUuid(uuidStr)
	ctx := context.WithValue(
		context.Background(),
		auth_apikey.TokenContextKey{},
		&auth_apikey.TokenContextData{
			RawToken: uuidStr,
			ApiKey: &db_queries.NeosyncApiAccountApiKey{
				UserID: useruuid,
			},
		},
	)

	resp, err := m.Service.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Msg.UserId, uuidStr)
}

func Test_GetUser_SetAnonymous(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	user := getUserMock(anonymousUserId)
	var nilUser db_queries.NeosyncApiUser

	m.QuerierMock.On("GetAnonymousUser", context.Background(), mock.Anything).Return(nilUser, sql.ErrNoRows)
	m.QuerierMock.On("SetAnonymousUser", context.Background(), mock.Anything).Return(user, nil)

	resp, err := m.Service.GetUser(context.Background(), &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, anonymousUserId, resp.Msg.GetUserId())
}

func Test_GetUser(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)

	resp, err := m.Service.GetUser(ctx, &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "SetAnonymousUser", context.Background(), mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockUserId, resp.Msg.GetUserId())
}

func Test_GetUser_InvalidToken(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	_, err := m.Service.GetUser(context.Background(), &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "GetUserAssociationByProviderSub", context.Background(), mock.Anything, mock.Anything)
	assert.Error(t, err)
}

func Test_GetUser_AssociationError(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	var nilUserAssociation db_queries.NeosyncApiUserIdentityProviderAssociation

	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(nilUserAssociation, sql.ErrNoRows)

	_, err := m.Service.GetUser(ctx, &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	assert.Error(t, err)
}

// SetUser
func Test_SetUser_Anonymous_Error(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	var nilUser db_queries.NeosyncApiUser

	m.QuerierMock.On("SetAnonymousUser", context.Background(), mock.Anything).Return(nilUser, errors.New("some error"))

	resp, err := m.Service.SetUser(context.Background(), &connect.Request[mgmtv1alpha1.SetUserRequest]{})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_SetUser_Anonymous(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	user := getUserMock(anonymousUserId)

	m.QuerierMock.On("SetAnonymousUser", context.Background(), mock.Anything).Return(user, nil)

	resp, err := m.Service.SetUser(context.Background(), &connect.Request[mgmtv1alpha1.SetUserRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, anonymousUserId, resp.Msg.GetUserId())
}

func Test_Service_SetUser_AuthDisabled_ApiKey(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: false})

	uuidStr := uuid.NewString()
	useruuid, _ := nucleusdb.ToUuid(uuidStr)
	ctx := context.WithValue(
		context.Background(),
		auth_apikey.TokenContextKey{},
		&auth_apikey.TokenContextData{
			RawToken: uuidStr,
			ApiKey: &db_queries.NeosyncApiAccountApiKey{
				UserID: useruuid,
			},
		},
	)

	resp, err := m.Service.SetUser(ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Msg.UserId, uuidStr)
}

func Test_SetUser_InvalidToken(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	_, err := m.Service.SetUser(context.Background(), &connect.Request[mgmtv1alpha1.SetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "GetUserByProviderSub", context.Background(), mock.Anything, mock.Anything)
	assert.Error(t, err)
}

// GetUserAccounts
func Test_GetUserAccounts(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	slug := "slug"
	accounts := []db_queries.NeosyncApiAccount{
		getUserAccountMock(mockAccountId, slug, 0),
	}
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("GetAccountsByUser", ctx, mock.Anything, userUuid).Return(accounts, nil)

	resp, err := m.Service.GetUserAccounts(ctx, &connect.Request[mgmtv1alpha1.GetUserAccountsRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.GetAccounts()))
	assert.Equal(t, mockAccountId, resp.Msg.GetAccounts()[0].Id)
	assert.Equal(t, slug, resp.Msg.GetAccounts()[0].Name)
}

func Test_SetPersonalAccount(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.DbtxMock.On("Begin", ctx).Return(mockTx, nil)
	m.QuerierMock.On("GetPersonalAccountByUserId", ctx, mockTx, userUuid).Return(db_queries.NeosyncApiAccount{ID: accountUuid}, nil)
	m.QuerierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := m.Service.SetPersonalAccount(ctx, &connect.Request[mgmtv1alpha1.SetPersonalAccountRequest]{
		Msg: &mgmtv1alpha1.SetPersonalAccountRequest{},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.AccountId)
}

// IsUserInAccount
func Test_IsUserInAccount_True(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("IsUserInAccount", ctx, mock.Anything, db_queries.IsUserInAccountParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(1), nil)
	m.QuerierMock.On("IsUserInAccountApiKey", ctx, mock.Anything, db_queries.IsUserInAccountApiKeyParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(0), nil)

	resp, err := m.Service.IsUserInAccount(ctx, &connect.Request[mgmtv1alpha1.IsUserInAccountRequest]{Msg: &mgmtv1alpha1.IsUserInAccountRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.Ok)
}

func Test_IsUserInAccount_False(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("IsUserInAccount", ctx, mock.Anything, db_queries.IsUserInAccountParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(0), nil)
	m.QuerierMock.On("IsUserInAccountApiKey", ctx, mock.Anything, db_queries.IsUserInAccountApiKeyParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(0), nil)

	resp, err := m.Service.IsUserInAccount(ctx, &connect.Request[mgmtv1alpha1.IsUserInAccountRequest]{Msg: &mgmtv1alpha1.IsUserInAccountRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.Ok)
}

func Test_IsUserInAccount_ApiKey(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("IsUserInAccountApiKey", ctx, mock.Anything, db_queries.IsUserInAccountApiKeyParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(1), nil)

	resp, err := m.Service.IsUserInAccount(ctx, &connect.Request[mgmtv1alpha1.IsUserInAccountRequest]{Msg: &mgmtv1alpha1.IsUserInAccountRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Msg.Ok)
}

func Test_CreateTeamAccount(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	mockTeamName := "team-name"
	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.DbtxMock.On("Begin", ctx).Return(mockTx, nil)
	m.QuerierMock.On("GetAccountsByUser", ctx, mockTx, userUuid).Return([]db_queries.NeosyncApiAccount{{AccountSlug: "other"}}, nil)
	m.QuerierMock.On("CreateTeamAccount", ctx, mockTx, mockTeamName).Return(db_queries.NeosyncApiAccount{ID: accountUuid, AccountSlug: mockTeamName}, nil)
	m.QuerierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := m.Service.CreateTeamAccount(ctx, &connect.Request[mgmtv1alpha1.CreateTeamAccountRequest]{Msg: &mgmtv1alpha1.CreateTeamAccountRequest{Name: mockTeamName}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.AccountId)
}

// GetTeamAccountMembers
func Test_GetTeamAccountMembers(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	authProviderId := "auth-provider-id"
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	m.QuerierMock.On("GetUserIdentitiesByTeamAccount", ctx, mock.Anything, accountUuid).Return(
		[]db_queries.NeosyncApiUserIdentityProviderAssociation{{UserID: userUuid, ProviderSub: authProviderId}},
		nil,
	)
	m.Auth0MgmtMock.On("GetUserById", ctx, authProviderId).
		Return(
			&management.User{
				ID: &authProviderId,
			}, nil,
		)
	resp, err := m.Service.GetTeamAccountMembers(ctx, &connect.Request[mgmtv1alpha1.GetTeamAccountMembersRequest]{Msg: &mgmtv1alpha1.GetTeamAccountMembersRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.Users))
	assert.Equal(t, mockUserId, resp.Msg.Users[0].Id)
}

func Test_GetTeamAccountMembers_NoAuthUser(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	authProviderId := "auth-provider-id"
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	m.QuerierMock.On("GetUserIdentitiesByTeamAccount", ctx, mock.Anything, accountUuid).Return(
		[]db_queries.NeosyncApiUserIdentityProviderAssociation{{UserID: userUuid, ProviderSub: authProviderId}},
		nil,
	)
	m.Auth0MgmtMock.On("GetUserById", ctx, authProviderId).
		Return(
			nil, fmt.Errorf("bad news"),
		)
	resp, err := m.Service.GetTeamAccountMembers(ctx, &connect.Request[mgmtv1alpha1.GetTeamAccountMembersRequest]{Msg: &mgmtv1alpha1.GetTeamAccountMembersRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.Users))
	assert.Equal(t, mockUserId, resp.Msg.Users[0].Id)
}

// RemoveTeamAccountMember
func Test_RemoveTeamAccountMember(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, true)
	m.QuerierMock.On("RemoveAccountUser", ctx, mock.Anything, db_queries.RemoveAccountUserParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(nil)

	resp, err := m.Service.RemoveTeamAccountMember(ctx, &connect.Request[mgmtv1alpha1.RemoveTeamAccountMemberRequest]{Msg: &mgmtv1alpha1.RemoveTeamAccountMemberRequest{AccountId: mockAccountId, UserId: mockUserId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_RemoveTeamAccountMember_NoRows(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, true)
	m.QuerierMock.On("RemoveAccountUser", ctx, mock.Anything, db_queries.RemoveAccountUserParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(sql.ErrNoRows)

	resp, err := m.Service.RemoveTeamAccountMember(ctx, &connect.Request[mgmtv1alpha1.RemoveTeamAccountMemberRequest]{Msg: &mgmtv1alpha1.RemoveTeamAccountMemberRequest{AccountId: mockAccountId, UserId: mockUserId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// InviteUserToTeamAccount
func Test_InviteUserToTeamAccount_Unauthorized(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, false)

	resp, err := m.Service.InviteUserToTeamAccount(ctx, &connect.Request[mgmtv1alpha1.InviteUserToTeamAccountRequest]{Msg: &mgmtv1alpha1.InviteUserToTeamAccountRequest{AccountId: mockAccountId, Email: "test-email"}})

	m.QuerierMock.AssertNotCalled(t, "UpdateActiveAccountInvitesToExpired", mock.Anything, mock.Anything, mock.Anything)
	m.QuerierMock.AssertNotCalled(t, "CreateAccountInvite", mock.Anything, mock.Anything, mock.Anything)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_InviteUserToTeamAccount_NotTeamAccount(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, false)

	resp, err := m.Service.InviteUserToTeamAccount(ctx, &connect.Request[mgmtv1alpha1.InviteUserToTeamAccountRequest]{Msg: &mgmtv1alpha1.InviteUserToTeamAccountRequest{AccountId: mockAccountId, Email: "test-email"}})

	m.QuerierMock.AssertNotCalled(t, "UpdateActiveAccountInvitesToExpired", mock.Anything, mock.Anything, mock.Anything)
	m.QuerierMock.AssertNotCalled(t, "CreateAccountInvite", mock.Anything, mock.Anything, mock.Anything)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// GetTeamAccountInvites
func Test_GetTeamAccountInvites(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, true)
	m.QuerierMock.On("GetActiveAccountInvites", ctx, mock.Anything, accountUuid).Return([]db_queries.NeosyncApiAccountInvite{
		{AccountID: accountUuid, SenderUserID: userUuid},
	}, nil)

	resp, err := m.Service.GetTeamAccountInvites(ctx, &connect.Request[mgmtv1alpha1.GetTeamAccountInvitesRequest]{Msg: &mgmtv1alpha1.GetTeamAccountInvitesRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.Invites))
}

func Test_GetTeamAccountInvites_NoRows(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	var nilInvites []db_queries.NeosyncApiAccountInvite

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, true)
	m.QuerierMock.On("GetActiveAccountInvites", ctx, mock.Anything, accountUuid).Return(nilInvites, sql.ErrNoRows)

	resp, err := m.Service.GetTeamAccountInvites(ctx, &connect.Request[mgmtv1alpha1.GetTeamAccountInvitesRequest]{Msg: &mgmtv1alpha1.GetTeamAccountInvitesRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Msg.Invites))
}

// RemoveTeamAccountInvite
func Test_RemoveTeamAccountInvite(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	inviteId := uuid.NewString()
	inviteUuid, _ := nucleusdb.ToUuid(inviteId)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)

	mockVerifyUserInAccount(ctx, m.QuerierMock, accountUuid, userUuid, true)
	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, true)
	m.QuerierMock.On("GetAccountInvite", ctx, mock.Anything, inviteUuid).Return(db_queries.NeosyncApiAccountInvite{
		AccountID: accountUuid, SenderUserID: userUuid, ID: inviteUuid,
	}, nil)
	m.QuerierMock.On("RemoveAccountInvite", ctx, mock.Anything, inviteUuid).Return(nil)

	resp, err := m.Service.RemoveTeamAccountInvite(ctx, &connect.Request[mgmtv1alpha1.RemoveTeamAccountInviteRequest]{Msg: &mgmtv1alpha1.RemoveTeamAccountInviteRequest{Id: inviteId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_RemoveTeamAccountInvite_NotFound(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	ctx := getAuthenticatedCtxMock(mockAuthProvider)

	var nilInvite db_queries.NeosyncApiAccountInvite
	inviteId := uuid.NewString()
	inviteUuid, _ := nucleusdb.ToUuid(inviteId)

	m.QuerierMock.On("GetAccountInvite", ctx, mock.Anything, inviteUuid).Return(nilInvite, sql.ErrNoRows)

	resp, err := m.Service.RemoveTeamAccountInvite(ctx, &connect.Request[mgmtv1alpha1.RemoveTeamAccountInviteRequest]{Msg: &mgmtv1alpha1.RemoveTeamAccountInviteRequest{Id: inviteId}})

	m.QuerierMock.AssertNotCalled(t, "RemoveAccountInvite", mock.Anything, mock.Anything, mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// AcceptTeamAccountInvite

func Test_AcceptTeamAccountInvite(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	token := uuid.NewString()
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	inviteUuid, _ := nucleusdb.ToUuid(uuid.NewString())
	email := "fakeemail@fakefake.com"
	expiresAt := pgtype.Timestamp{
		Time: time.Now().Add(1 * time.Hour),
	}

	mockVerifyTeamAccount(ctx, m.QuerierMock, accountUuid, true)
	m.QuerierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("GetUserIdentityByUserId", ctx, mock.Anything, userUuid).Return(userAssociation, nil)
	m.DbtxMock.On("Begin", ctx).Return(mockTx, nil)
	m.Auth0MgmtMock.On("GetUserById", ctx, userAssociation.ProviderSub).
		Return(
			&management.User{
				ID:    &userAssociation.ProviderSub,
				Email: &email,
			}, nil,
		)
	m.QuerierMock.On("GetAccountInviteByToken", ctx, mockTx, token).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        email,
		ExpiresAt:    expiresAt,
		ID:           inviteUuid,
		Accepted:     pgtype.Bool{Bool: false},
	}, nil)
	m.QuerierMock.On("UpdateAccountInviteToAccepted", ctx, mockTx, inviteUuid).Return(db_queries.NeosyncApiAccountInvite{}, nil)
	m.QuerierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, sql.ErrNoRows)
	m.QuerierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)
	m.QuerierMock.On("GetAccount", ctx, mock.Anything, accountUuid).Return(db_queries.NeosyncApiAccount{
		ID:          accountUuid,
		AccountType: int16(1),
	}, nil)

	resp, err := m.Service.AcceptTeamAccountInvite(ctx, &connect.Request[mgmtv1alpha1.AcceptTeamAccountInviteRequest]{
		Msg: &mgmtv1alpha1.AcceptTeamAccountInviteRequest{
			Token: token,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockAccountId, resp.Msg.Account.Id)
}

type serviceMocks struct {
	Service                   *Service
	DbtxMock                  *nucleusdb.MockDBTX
	QuerierMock               *db_queries.MockQuerier
	Auth0MgmtMock             *auth0.MockAuth0MgmtClientInterface
	TemporalClientManagerMock *clientmanager.MockTemporalClientManagerClient
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockAuth0MgmtClient := auth0.NewMockAuth0MgmtClientInterface(t)
	temporalClientManager := clientmanager.NewMockTemporalClientManagerClient(t)

	service := New(config, nucleusdb.New(mockDbtx, mockQuerier), mockAuth0MgmtClient, temporalClientManager)

	return &serviceMocks{
		Service:                   service,
		DbtxMock:                  mockDbtx,
		QuerierMock:               mockQuerier,
		Auth0MgmtMock:             mockAuth0MgmtClient,
		TemporalClientManagerMock: temporalClientManager,
	}
}

func getUserMock(userId string) db_queries.NeosyncApiUser {
	idUuid, _ := nucleusdb.ToUuid(userId)
	return db_queries.NeosyncApiUser{ID: idUuid}
}

//nolint:all
func getAuthenticatedCtxMock(authProviderId string) context.Context {
	data := &authjwt.TokenContextData{AuthUserId: authProviderId}
	return context.WithValue(context.Background(), authjwt.TokenContextKey{}, data)
}

//nolint:all
func getUserIdentityProviderAssociationMock(userId, providerId string) db_queries.NeosyncApiUserIdentityProviderAssociation {
	idUuid, _ := nucleusdb.ToUuid(userId)
	return db_queries.NeosyncApiUserIdentityProviderAssociation{
		UserID:      idUuid,
		ProviderSub: providerId,
	}
}

func getUserAccountMock(accountId, slug string, accountType int16) db_queries.NeosyncApiAccount {
	idUuid, _ := nucleusdb.ToUuid(accountId)
	return db_queries.NeosyncApiAccount{
		ID:          idUuid,
		AccountSlug: slug,
		AccountType: accountType,
	}
}

func mockVerifyUserInAccount(ctx context.Context, querierMock *db_queries.MockQuerier, accountUuid, userUuid pgtype.UUID, isInAccount bool) {
	var inAccount = 0
	if isInAccount {
		inAccount = 1
	}
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	querierMock.On("GetUserAssociationByProviderSub", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	querierMock.On("IsUserInAccount", ctx, mock.Anything, db_queries.IsUserInAccountParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(inAccount), nil)
	querierMock.On("IsUserInAccountApiKey", ctx, mock.Anything, db_queries.IsUserInAccountApiKeyParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(0), nil)
}

func mockVerifyTeamAccount(ctx context.Context, querierMock *db_queries.MockQuerier, accountUuid pgtype.UUID, isTeamAccount bool) {
	var accountType = 0
	if isTeamAccount {
		accountType = 1
	}
	querierMock.On("GetAccount", ctx, mock.Anything, accountUuid).Return(db_queries.NeosyncApiAccount{
		ID:          accountUuid,
		AccountType: int16(accountType),
	}, nil)
}
