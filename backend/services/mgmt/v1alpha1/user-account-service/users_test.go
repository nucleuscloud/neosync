package v1alpha1_useraccountservice

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	authjwt "github.com/nucleuscloud/neosync/backend/internal/jwt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
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
	m.QuerierMock.On("GetUserAssociationByAuth0Id", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)

	resp, err := m.Service.GetUser(ctx, &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "SetAnonymousUser", context.Background(), mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mockUserId, resp.Msg.GetUserId())
}

func Test_GetUser_InvalidToken(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	_, err := m.Service.GetUser(context.Background(), &connect.Request[mgmtv1alpha1.GetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "GetUserAssociationByAuth0Id", context.Background(), mock.Anything, mock.Anything)
	assert.Error(t, err)
}

func Test_GetUser_AssociationError(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	var nilUserAssociation db_queries.NeosyncApiUserIdentityProviderAssociation

	m.QuerierMock.On("GetUserAssociationByAuth0Id", ctx, mock.Anything, mockAuthProvider).Return(nilUserAssociation, sql.ErrNoRows)

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

func Test_SetUser_InvalidToken(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	_, err := m.Service.SetUser(context.Background(), &connect.Request[mgmtv1alpha1.SetUserRequest]{})

	m.QuerierMock.AssertNotCalled(t, "GetUserByAuth0Id", context.Background(), mock.Anything, mock.Anything)
	assert.Error(t, err)
}

// TODO fix these test @alisha
// // FAILING
// func Test_SetUser_Error(t *testing.T) {
//	m := createServiceMock(t, &Config{IsAuthEnabled: true})

// 	ctx := getAuthenticatedCtx(mockAuthProvider)
// 	var nilUser db_queries.NeosyncApiUser
// 	var tx pgx.Tx
// 	m.QuerierMock.On("GetUserByAuth0Id", ctx, mock.Anything, mock.Anything).Return(nilUser, errors.New("some error"))
// 	mockDbtx.On("Begin", ctx).Return(tx, errors.New("some error"))

// 	_, err := m.Service.SetUser(ctx, &connect.Request[mgmtv1alpha1.SetUserRequest]{})
// 	assert.Error(t, err)
// }

// // FAILING
// func Test_SetUser(t *testing.T) {
// 	m := createServiceMock(t, &Config{IsAuthEnabled: true})

// 	user := getUserMock(mockUserId)
// 	ctx := getAuthenticatedCtx(mockAuthProvider)
// 	var tx pgx.Tx
// 	m.QuerierMock.On("GetUserByAuth0Id", ctx, mock.Anything, mock.Anything).Return(user, errors.New("some error"))
// 	mockDbtx.On("Begin", ctx).Return(tx, nil)

// 	resp, err := m.Service.SetUser(ctx, &connect.Request[mgmtv1alpha1.SetUserRequest]{})
// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.Equal(t, mockUserId, resp.Msg.GetUserId())
// }

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
	m.QuerierMock.On("GetUserAssociationByAuth0Id", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("GetAccountsByUser", ctx, mock.Anything, userUuid).Return(accounts, nil)

	resp, err := m.Service.GetUserAccounts(ctx, &connect.Request[mgmtv1alpha1.GetUserAccountsRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.GetAccounts()))
	assert.Equal(t, mockAccountId, resp.Msg.GetAccounts()[0].Id)
	assert.Equal(t, slug, resp.Msg.GetAccounts()[0].Name)
}

// SetPersonalAccount
// TODO @alisha

// IsUserInAccount
func Test_IsUserInAccount_True(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	ctx := getAuthenticatedCtxMock(mockAuthProvider)
	userAssociation := getUserIdentityProviderAssociationMock(mockUserId, mockAuthProvider)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	m.QuerierMock.On("GetUserAssociationByAuth0Id", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("IsUserInAccount", ctx, mock.Anything, db_queries.IsUserInAccountParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(1), nil)

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
	m.QuerierMock.On("GetUserAssociationByAuth0Id", ctx, mock.Anything, mockAuthProvider).Return(userAssociation, nil)
	m.QuerierMock.On("IsUserInAccount", ctx, mock.Anything, db_queries.IsUserInAccountParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(int64(0), nil)

	resp, err := m.Service.IsUserInAccount(ctx, &connect.Request[mgmtv1alpha1.IsUserInAccountRequest]{Msg: &mgmtv1alpha1.IsUserInAccountRequest{AccountId: mockAccountId}})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, false, resp.Msg.Ok)
}

type serviceMocks struct {
	Service     *Service
	DbtxMock    *nucleusdb.MockDBTX
	QuerierMock *db_queries.MockQuerier
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	service := New(config, nucleusdb.New(mockDbtx, mockQuerier))

	return &serviceMocks{
		Service:     service,
		DbtxMock:    mockDbtx,
		QuerierMock: mockQuerier,
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
		UserID:          idUuid,
		Auth0ProviderID: providerId,
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
