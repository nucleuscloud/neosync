package nucleusdb

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

const (
	anonymousUserId = "00000000-0000-0000-0000-000000000000"
	mockUserId      = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId   = "5629813e-1a35-4874-922c-9827d85f0378"
	mockTeamName    = "team-name"
)

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
