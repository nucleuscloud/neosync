package nucleusdb

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mock "github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

const (
	anonymousUserId = "00000000-0000-0000-0000-000000000000"
	mockUserId      = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId   = "5629813e-1a35-4874-922c-9827d85f0378"
)

// MockTx is a mock type for the Tx interface
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	args := m.Called()
	return args.Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Row)
}

func (m *MockTx) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

// CreateTeamAccount
func Test_GetUser_Anonymous(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	userUuid, _ := ToUuid(mockUserId)
	accountUuid, _ := ToUuid(mockAccountId)
	teamName := "team-name"
	ctx := context.Background()

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetTeamAccountsByUserId", ctx, mockTx, userUuid).Return([]db_queries.NeosyncApiAccount{}, nil)
	querierMock.On("CreateTeamAccount", ctx, mockTx, teamName).Return(db_queries.NeosyncApiAccount{ID: accountUuid}, nil)
	querierMock.On("CreateAccountUserAssociation", ctx, mockTx).Return(nil, nil)

	resp, err := service.CreateTeamAccount(context.Background(), userUuid, teamName)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
