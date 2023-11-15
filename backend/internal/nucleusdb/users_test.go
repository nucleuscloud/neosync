package nucleusdb

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

const (
	anonymousUserId = "00000000-0000-0000-0000-000000000000"
	mockUserId      = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId   = "5629813e-1a35-4874-922c-9827d85f0378"
	mockTeamName    = "team-name"
	mockAuth0Id     = "643a8663-6b2e-4d29-a0f0-4a0700ff21ea"
	mockEmail       = "fake@fake.com"
)

// SetUserByAuth0Id
func Test_SetUserByAuth0Id(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetUserByAuth0Id", ctx, mockTx, mockAuth0Id).Return(db_queries.NeosyncApiUser{ID: userUuid}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetUserByAuth0Id(context.Background(), mockAuth0Id)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userUuid, resp.ID)
}

func Test_SetUserByAuth0Id_Association_User(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	ctx := context.Background()
	var nilUser db_queries.NeosyncApiUser

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetUserByAuth0Id", ctx, mockTx, mockAuth0Id).Return(nilUser, sql.ErrNoRows)
	querierMock.On("GetUserAssociationByAuth0Id", ctx, mockTx, mockAuth0Id).Return(db_queries.NeosyncApiUserIdentityProviderAssociation{UserID: userUuid}, nil)
	querierMock.On("GetUser", ctx, mockTx, userUuid).Return(db_queries.NeosyncApiUser{ID: userUuid}, nil)

	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetUserByAuth0Id(context.Background(), mockAuth0Id)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userUuid, resp.ID)
}

func Test_SetUserByAuth0Id_NoAssociation(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	ctx := context.Background()
	var nilUser db_queries.NeosyncApiUser
	var nilAssociation db_queries.NeosyncApiUserIdentityProviderAssociation

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetUserByAuth0Id", ctx, mockTx, mockAuth0Id).Return(nilUser, sql.ErrNoRows)
	querierMock.On("GetUserAssociationByAuth0Id", ctx, mockTx, mockAuth0Id).Return(nilAssociation, sql.ErrNoRows)
	querierMock.On("CreateNonMachineUser", ctx, mockTx).Return(db_queries.NeosyncApiUser{ID: userUuid}, nil)
	querierMock.On("CreateAuth0IdentityProviderAssociation", ctx, mockTx, db_queries.CreateAuth0IdentityProviderAssociationParams{
		UserID:          userUuid,
		Auth0ProviderID: mockAuth0Id,
	}).Return(db_queries.NeosyncApiUserIdentityProviderAssociation{UserID: userUuid,
		Auth0ProviderID: mockAuth0Id}, nil)

	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetUserByAuth0Id(context.Background(), mockAuth0Id)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userUuid, resp.ID)
}

func Test_SetUserByAuth0Id_Association_NoUser(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	ctx := context.Background()
	var nilUser db_queries.NeosyncApiUser

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetUserByAuth0Id", ctx, mockTx, mockAuth0Id).Return(nilUser, sql.ErrNoRows)
	querierMock.On("GetUserAssociationByAuth0Id", ctx, mockTx, mockAuth0Id).Return(db_queries.NeosyncApiUserIdentityProviderAssociation{UserID: userUuid}, nil)
	querierMock.On("GetUser", ctx, mockTx, userUuid).Return(nilUser, sql.ErrNoRows)
	querierMock.On("CreateNonMachineUser", ctx, mockTx).Return(db_queries.NeosyncApiUser{ID: userUuid}, nil)

	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetUserByAuth0Id(context.Background(), mockAuth0Id)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userUuid, resp.ID)
}

func Test_SetUserByAuth0Id_CreateAuth0IdentityProviderAssociation_Error(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	ctx := context.Background()
	var nilUser db_queries.NeosyncApiUser
	var nilAssociation db_queries.NeosyncApiUserIdentityProviderAssociation

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetUserByAuth0Id", ctx, mockTx, mockAuth0Id).Return(nilUser, sql.ErrNoRows)
	querierMock.On("GetUserAssociationByAuth0Id", ctx, mockTx, mockAuth0Id).Return(nilAssociation, sql.ErrNoRows)
	querierMock.On("CreateNonMachineUser", ctx, mockTx).Return(db_queries.NeosyncApiUser{ID: userUuid}, nil)
	querierMock.On("CreateAuth0IdentityProviderAssociation", ctx, mockTx, db_queries.CreateAuth0IdentityProviderAssociationParams{
		UserID:          userUuid,
		Auth0ProviderID: mockAuth0Id,
	}).Return(nilAssociation, errors.New("bad news"))

	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetUserByAuth0Id(context.Background(), mockAuth0Id)

	mockTx.AssertNotCalled(t, "Commit", mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// SetPersonalAccount
func Test_SetPersonalAccount(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetPersonalAccountByUserId", ctx, mockTx, userUuid).Return(db_queries.NeosyncApiAccount{ID: accountUuid}, nil)
	querierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetPersonalAccount(ctx, userUuid)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp.ID)
}

func Test_SetPersonalAccount_CreateUserAssociation(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	ctx := context.Background()
	var nilAssociation db_queries.NeosyncApiAccountUserAssociation

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetPersonalAccountByUserId", ctx, mockTx, userUuid).Return(db_queries.NeosyncApiAccount{ID: accountUuid}, nil)
	querierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(nilAssociation, sql.ErrNoRows)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetPersonalAccount(ctx, userUuid)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp.ID)
}

func Test_SetPersonalAccount_CreateAccount(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	ctx := context.Background()
	var nilAccount db_queries.NeosyncApiAccount

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetPersonalAccountByUserId", ctx, mockTx, userUuid).Return(nilAccount, sql.ErrNoRows)
	querierMock.On("CreatePersonalAccount", ctx, mockTx, "personal").Return(db_queries.NeosyncApiAccount{ID: accountUuid}, nil)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{AccountID: accountUuid, UserID: userUuid}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetPersonalAccount(ctx, userUuid)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp.ID)
}

func Test_SetPersonalAccount_Rollback(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	ctx := context.Background()
	var nilAccount db_queries.NeosyncApiAccount

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetPersonalAccountByUserId", ctx, mockTx, userUuid).Return(nilAccount, sql.ErrNoRows)
	querierMock.On("CreatePersonalAccount", ctx, mockTx, "personal").Return(db_queries.NeosyncApiAccount{ID: accountUuid}, nil)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{AccountID: accountUuid, UserID: userUuid}, errors.New("boo"))
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.SetPersonalAccount(ctx, userUuid)

	mockTx.AssertNotCalled(t, "Commit", ctx)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// CreateTeamAccount
func Test_CreateTeamAccount(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountsByUser", ctx, mockTx, userUuid).Return([]db_queries.NeosyncApiAccount{{AccountSlug: "other"}}, nil)
	querierMock.On("CreateTeamAccount", ctx, mockTx, mockTeamName).Return(db_queries.NeosyncApiAccount{ID: accountUuid, AccountSlug: mockTeamName}, nil)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.CreateTeamAccount(context.Background(), userUuid, mockTeamName)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp.ID)
	assert.Equal(t, mockTeamName, resp.AccountSlug)
}

func Test_CreateTeamAccount_AlreadyExists(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountsByUser", ctx, mockTx, userUuid).Return([]db_queries.NeosyncApiAccount{{AccountSlug: mockTeamName}}, nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.CreateTeamAccount(context.Background(), userUuid, mockTeamName)

	querierMock.AssertNotCalled(t, "CreateTeamAccount", mock.Anything, mock.Anything, mock.Anything)
	querierMock.AssertNotCalled(t, "CreateAccountUserAssociation", mock.Anything, mock.Anything, mock.Anything)
	mockTx.AssertNotCalled(t, "Commit", mock.Anything)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_CreateTeamAccount_NoRows(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	var nilAccounts []db_queries.NeosyncApiAccount
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountsByUser", ctx, mockTx, userUuid).Return(nilAccounts, sql.ErrNoRows)
	querierMock.On("CreateTeamAccount", ctx, mockTx, mockTeamName).Return(db_queries.NeosyncApiAccount{ID: accountUuid, AccountSlug: mockTeamName}, nil)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.CreateTeamAccount(context.Background(), userUuid, mockTeamName)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp.ID)
	assert.Equal(t, mockTeamName, resp.AccountSlug)
}

func Test_CreateTeamAccount_Rollback(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	ctx := context.Background()
	var nilAssociation db_queries.NeosyncApiAccountUserAssociation

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountsByUser", ctx, mockTx, userUuid).Return([]db_queries.NeosyncApiAccount{{AccountSlug: "other"}}, nil)
	querierMock.On("CreateTeamAccount", ctx, mockTx, mockTeamName).Return(db_queries.NeosyncApiAccount{ID: accountUuid, AccountSlug: mockTeamName}, nil)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(nilAssociation, errors.New("sad"))
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.CreateTeamAccount(context.Background(), userUuid, mockTeamName)

	mockTx.AssertCalled(t, "Rollback", ctx)
	mockTx.AssertNotCalled(t, "Commit", mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// CreateTeamAccountInvite
func Test_CreateTeamAccountInvite(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	expiresAt := pgtype.Timestamp{
		Time: time.Now(),
	}
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccount", ctx, mockTx, accountUuid).Return(db_queries.NeosyncApiAccount{AccountType: int16(1)}, nil)
	querierMock.On("UpdateActiveAccountInvitesToExpired", ctx, mockTx, db_queries.UpdateActiveAccountInvitesToExpiredParams{
		AccountId: accountUuid,
		Email:     mockEmail,
	}).Return(db_queries.NeosyncApiAccountInvite{}, nil)
	querierMock.On("CreateAccountInvite", ctx, mockTx, db_queries.CreateAccountInviteParams{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        mockEmail,
		ExpiresAt:    expiresAt,
	}).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        mockEmail,
		ExpiresAt:    expiresAt,
	}, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	resp, err := service.CreateTeamAccountInvite(context.Background(), accountUuid, userUuid, mockEmail, expiresAt)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_CreateTeamAccountInvite_NotTeamAccount(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	expiresAt := pgtype.Timestamp{
		Time: time.Now(),
	}
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccount", ctx, mockTx, accountUuid).Return(db_queries.NeosyncApiAccount{AccountType: int16(0)}, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	resp, err := service.CreateTeamAccountInvite(context.Background(), accountUuid, userUuid, mockEmail, expiresAt)

	querierMock.AssertNotCalled(t, "UpdateActiveAccountInvitesToExpired", mock.Anything, mock.Anything, mock.Anything)
	querierMock.AssertNotCalled(t, "CreateAccountInvite", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ValidateInviteTokenAndAddUserToAccount
func Test_ValidateInviteAddUserToAccount_InvalidEmail(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)
	service := New(dbtxMock, querierMock)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	token := uuid.NewString()
	expiresAt := pgtype.Timestamp{
		Time: time.Now(),
	}
	ctx := context.Background()

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountInviteByToken", ctx, mockTx, token).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        "diff-email",
		ExpiresAt:    expiresAt,
	}, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	_, err := service.ValidateInviteAddUserToAccount(context.Background(), userUuid, token, mockEmail)

	querierMock.AssertNotCalled(t, "CreateAccountUserAssociation", mock.Anything, mock.Anything, mock.Anything)
	querierMock.AssertNotCalled(t, "UpdateAccountInviteToAccepted", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
}

func Test_ValidateInviteAddUserToAccount_UserAlreadyInAccount(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)
	service := New(dbtxMock, querierMock)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	inviteUuid, _ := ToUuid(uuid.NewString())
	token := uuid.NewString()
	expiresAt := pgtype.Timestamp{
		Time: time.Now(),
	}
	ctx := context.Background()

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountInviteByToken", ctx, mockTx, token).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        mockEmail,
		ExpiresAt:    expiresAt,
		ID:           inviteUuid,
	}, nil)
	querierMock.On("UpdateAccountInviteToAccepted", ctx, mockTx, inviteUuid).Return(db_queries.NeosyncApiAccountInvite{}, nil)
	querierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	resp, err := service.ValidateInviteAddUserToAccount(context.Background(), userUuid, token, mockEmail)

	querierMock.AssertNotCalled(t, "CreateAccountUserAssociation", mock.Anything, mock.Anything, mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp)
}

func Test_ValidateInviteAddUserToAccount_InvitedAlreadyAccepted(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)
	service := New(dbtxMock, querierMock)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	inviteUuid, _ := ToUuid(uuid.NewString())
	token := uuid.NewString()
	expiresAt := pgtype.Timestamp{
		Time: time.Now(),
	}
	ctx := context.Background()

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountInviteByToken", ctx, mockTx, token).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        mockEmail,
		ExpiresAt:    expiresAt,
		ID:           inviteUuid,
		Accepted:     pgtype.Bool{Bool: true},
	}, nil)
	querierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, sql.ErrNoRows)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	_, err := service.ValidateInviteAddUserToAccount(context.Background(), userUuid, token, mockEmail)

	querierMock.AssertNotCalled(t, "CreateAccountUserAssociation", mock.Anything, mock.Anything, mock.Anything)
	querierMock.AssertNotCalled(t, "UpdateAccountInviteToAccepted", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
}

func Test_ValidateInviteAddUserToAccount_InvitedExpired(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)
	service := New(dbtxMock, querierMock)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	inviteUuid, _ := ToUuid(uuid.NewString())
	token := uuid.NewString()
	expiresAt := pgtype.Timestamp{
		Time: time.Now().Add(-1 * time.Hour),
	}
	ctx := context.Background()

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountInviteByToken", ctx, mockTx, token).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        mockEmail,
		ExpiresAt:    expiresAt,
		ID:           inviteUuid,
		Accepted:     pgtype.Bool{Bool: false},
	}, nil)
	querierMock.On("UpdateAccountInviteToAccepted", ctx, mockTx, inviteUuid).Return(db_queries.NeosyncApiAccountInvite{}, nil)
	querierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, sql.ErrNoRows)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	_, err := service.ValidateInviteAddUserToAccount(context.Background(), userUuid, token, mockEmail)

	querierMock.AssertNotCalled(t, "CreateAccountUserAssociation", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
}

func Test_ValidateInviteAddUserToAccount(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)
	service := New(dbtxMock, querierMock)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	inviteUuid, _ := ToUuid(uuid.NewString())
	token := uuid.NewString()
	expiresAt := pgtype.Timestamp{
		Time: time.Now().Add(1 * time.Hour),
	}
	ctx := context.Background()

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetAccountInviteByToken", ctx, mockTx, token).Return(db_queries.NeosyncApiAccountInvite{
		AccountID:    accountUuid,
		SenderUserID: userUuid,
		Email:        mockEmail,
		ExpiresAt:    expiresAt,
		ID:           inviteUuid,
		Accepted:     pgtype.Bool{Bool: false},
	}, nil)
	querierMock.On("UpdateAccountInviteToAccepted", ctx, mockTx, inviteUuid).Return(db_queries.NeosyncApiAccountInvite{}, nil)
	querierMock.On("GetAccountUserAssociation", ctx, mockTx, db_queries.GetAccountUserAssociationParams{
		AccountId: accountUuid,
		UserId:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, sql.ErrNoRows)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx, db_queries.CreateAccountUserAssociationParams{
		AccountID: accountUuid,
		UserID:    userUuid,
	}).Return(db_queries.NeosyncApiAccountUserAssociation{}, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)

	resp, err := service.ValidateInviteAddUserToAccount(context.Background(), userUuid, token, mockEmail)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accountUuid, resp)
}
