package neosyncdb

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
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

var (
	discardLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
)

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
