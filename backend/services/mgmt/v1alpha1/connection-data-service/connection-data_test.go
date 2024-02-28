package v1alpha1_connectiondataservice

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	awsmanager "github.com/nucleuscloud/neosync/backend/internal/aws"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	anonymousUserId    = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider   = "test-provider"
	mockUserId         = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId      = "5629813e-1a35-4874-922c-9827d85f0378"
	mockConnectionName = "test-conn"
	mockConnectionId   = "884765c6-1708-488d-b03a-70a02b12c81e"
)

type ConnTypeMock string

const (
	MysqlMock    ConnTypeMock = "mysql"
	PostgresMock ConnTypeMock = "postgres"
	AwsS3Mock    ConnTypeMock = "awsS3"
)

// GetConnectionSchema
func Test_GetConnectionSchema_AwsS3(t *testing.T) {
	m := createServiceMock(t)

	mockJobRunId := "7c54e1ce-3924-477c-bfa8-ab8bd36cfee2-2023-12-21T22:02:35Z"
	mockPrefix := "workflows/7c54e1ce-3924-477c-bfa8-ab8bd36cfee2-2023-12-21T22:02:35Z/activities/public.regions/"
	mockKey := "workflows/7c54e1ce-3924-477c-bfa8-ab8bd36cfee2-2023-12-21T22:02:35Z/activities/public.regions/data/228.txt.gz"
	path := fmt.Sprintf("workflows/%s/activities/", mockJobRunId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, AwsS3Mock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.AwsManagerMock.On("NewS3Client", mock.Anything, mock.Anything).Return(nil, nil)
	isTruncated := false
	m.AwsManagerMock.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything, &s3.ListObjectsV2Input{
		Bucket:            aws.String(connection.ConnectionConfig.GetAwsS3Config().GetBucket()),
		Prefix:            aws.String(path),
		Delimiter:         aws.String("/"),
		ContinuationToken: nil,
	}).Return(&s3.ListObjectsV2Output{
		CommonPrefixes: []types.CommonPrefix{{Prefix: &mockPrefix}},
		IsTruncated:    &isTruncated,
	}, nil)

	m.AwsManagerMock.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything, &s3.ListObjectsV2Input{
		Bucket:  aws.String(connection.ConnectionConfig.GetAwsS3Config().GetBucket()),
		Prefix:  aws.String(fmt.Sprintf("%spublic.regions/data", path)),
		MaxKeys: aws.Int32(1),
	}).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{{Key: &mockKey}},
	}, nil)

	data, _ := gzipData([]byte(`{"region_id":1,"region_name":"Europe"}`))
	m.AwsManagerMock.On("GetObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: aws.Int64(int64(len(`{"region_id":1,"region_name":"Europe"}`))),
		ContentType:   aws.String("application/octet-stream"),
	}, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
			SchemaConfig: &mgmtv1alpha1.ConnectionSchemaConfig{
				Config: &mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config{
					AwsS3Config: &mgmtv1alpha1.AwsS3SchemaConfig{
						Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobRunId{JobRunId: mockJobRunId},
					},
				},
			},
		},
	})

	expected := []*mgmtv1alpha1.DatabaseColumn{
		{Schema: "public", Table: "regions", Column: "region_id"},
		{Schema: "public", Table: "regions", Column: "region_name"},
	}

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, expected, resp.Msg.Schemas)
}

func Test_GetConnectionSchema_Postgres(t *testing.T) {
	m := createServiceMock(t)

	mockColumns := []*pg_queries.GetDatabaseSchemaRow{
		{
			TableSchema: "public",
			TableName:   "users",
			ColumnName:  "id",
			DataType:    "integer",
		},
		{
			TableSchema: "public",
			TableName:   "users",
			ColumnName:  "name",
			DataType:    "character varying",
		}}

	pool, _ := pgxpool.New(context.Background(), "")
	m.PgPoolContainerMock.On("Open", mock.Anything).Return(pool, nil)
	m.PgPoolContainerMock.On("Close")
	m.SqlConnectorMock.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.PgPoolContainerMock, nil)

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.PgQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return(mockColumns, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	expected := []*mgmtv1alpha1.DatabaseColumn{}
	for _, col := range mockColumns {
		expected = append(expected, &mgmtv1alpha1.DatabaseColumn{
			Schema:   col.TableSchema,
			Table:    col.TableName,
			Column:   col.ColumnName,
			DataType: col.DataType,
		})
	}

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, expected, resp.Msg.Schemas)
}

func Test_GetConnectionSchema_Mysql(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	mockColumns := []*mysql_queries.GetDatabaseSchemaRow{
		{
			TableSchema: "public",
			TableName:   "users",
			ColumnName:  "id",
			DataType:    "integer",
		},
		{
			TableSchema: "public",
			TableName:   "users",
			ColumnName:  "name",
			DataType:    "character varying",
		},
	}

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)

	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return(mockColumns, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_NoRows(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{}, sql.ErrNoRows)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, []*mgmtv1alpha1.DatabaseColumn{}, resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{}, errors.New("oh no"))

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// GetConnectionForeignConstraints
func Test_GetConnectionForeignConstraints_Mysql(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)

	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	m.MysqlQueierMock.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				IsNullable:        "NO",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
			},
		}, nil)

	resp, err := m.Service.GetConnectionForeignConstraints(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionForeignConstraintsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionForeignConstraintsRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableConstraints, 1)
	assert.EqualValues(t, map[string]*mgmtv1alpha1.ForeignConstraintTables{
		"public.user_account_associations": {Constraints: []*mgmtv1alpha1.ForeignConstraint{
			{Column: "user_id", IsNullable: false, ForeignKey: &mgmtv1alpha1.ForeignKey{Table: "public.users", Column: "id"}},
		}},
	}, resp.Msg.TableConstraints)
}

func Test_GetConnectionForeignConstraints_Postgres(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	pool, _ := pgxpool.New(context.Background(), "")
	m.PgPoolContainerMock.On("Open", mock.Anything).Return(pool, nil)
	m.PgPoolContainerMock.On("Close")
	m.SqlConnectorMock.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.PgPoolContainerMock, nil)
	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)

	m.PgQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	m.PgQueierMock.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				IsNullable:        "NO",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
			},
		}, nil)

	resp, err := m.Service.GetConnectionForeignConstraints(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionForeignConstraintsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionForeignConstraintsRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableConstraints, 1)
	assert.EqualValues(t, map[string]*mgmtv1alpha1.ForeignConstraintTables{
		"public.user_account_associations": {Constraints: []*mgmtv1alpha1.ForeignConstraint{
			{Column: "user_id", IsNullable: false, ForeignKey: &mgmtv1alpha1.ForeignKey{Table: "public.users", Column: "id"}},
		}},
	}, resp.Msg.TableConstraints)
}

// GetConnectionPrimaryConstraints
func Test_GetConnectionPrimaryConstraints_Mysql(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)

	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	m.MysqlQueierMock.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetPrimaryKeyConstraintsRow{
			{
				ConstraintName: "pk_users_id",
				SchemaName:     "public",
				TableName:      "users",
				ColumnName:     "id",
			},
		}, nil)

	resp, err := m.Service.GetConnectionPrimaryConstraints(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionPrimaryConstraintsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionPrimaryConstraintsRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableConstraints, 1)
	assert.EqualValues(t, map[string]*mgmtv1alpha1.PrimaryConstraint{
		"public.users": {Columns: []string{"id"}},
	}, resp.Msg.TableConstraints)
}

func Test_GetConnectionPrimaryConstraints_Postgres(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	pool, _ := pgxpool.New(context.Background(), "")
	m.PgPoolContainerMock.On("Open", mock.Anything).Return(pool, nil)
	m.PgPoolContainerMock.On("Close")
	m.SqlConnectorMock.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.PgPoolContainerMock, nil)
	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)

	m.PgQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	m.PgQueierMock.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetPrimaryKeyConstraintsRow{
			{
				ConstraintName: "pk_users_id",
				SchemaName:     "public",
				TableName:      "users",
				ColumnName:     "id",
			},
		}, nil)

	resp, err := m.Service.GetConnectionPrimaryConstraints(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionPrimaryConstraintsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionPrimaryConstraintsRequest{
			ConnectionId: mockConnectionId,
		},
	})

	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableConstraints, 1)
	assert.EqualValues(t, map[string]*mgmtv1alpha1.PrimaryConstraint{
		"public.users": {Columns: []string{"id"}},
	}, resp.Msg.TableConstraints)
}

func Test_GetConnectionInitStatements_Mysql_Create(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("users", "CREATE TABLE public.users")
	m.SqlMock.ExpectQuery("SHOW CREATE TABLE public.users;").WillReturnRows(rows)

	resp, err := m.Service.GetConnectionInitStatements(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionInitStatementsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionInitStatementsRequest{
			ConnectionId: mockConnectionId,
			Options: &mgmtv1alpha1.InitStatementOptions{
				InitSchema:           true,
				TruncateBeforeInsert: false,
			},
		},
	})

	expectedInit := "CREATE TABLE IF NOT EXISTS  public.users;"
	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableInitStatements, 1)
	assert.Len(t, resp.Msg.TableTruncateStatements, 0)
	assert.Equal(t, expectedInit, resp.Msg.TableInitStatements["public.users"])
}

func Test_GetConnectionInitStatements_Mysql_Truncate(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.SqlDbContainerMock.On("Open").Return(m.SqlDbMock, nil)
	m.SqlDbContainerMock.On("Close").Return(nil)
	m.SqlConnectorMock.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.SqlDbContainerMock, nil)
	m.MysqlQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)

	resp, err := m.Service.GetConnectionInitStatements(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionInitStatementsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionInitStatementsRequest{
			ConnectionId: mockConnectionId,
			Options: &mgmtv1alpha1.InitStatementOptions{
				InitSchema:           false,
				TruncateBeforeInsert: true,
			},
		},
	})

	expectedTruncate := "TRUNCATE TABLE `public`.`users`;"
	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableInitStatements, 0)
	assert.Len(t, resp.Msg.TableTruncateStatements, 1)
	assert.Equal(t, expectedTruncate, resp.Msg.TableTruncateStatements["public.users"])
}

func Test_GetConnectionInitStatements_Postgres_Create(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	pool, _ := pgxpool.New(context.Background(), "")
	m.PgPoolContainerMock.On("Open", mock.Anything).Return(pool, nil)
	m.PgPoolContainerMock.On("Close")
	m.SqlConnectorMock.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.PgPoolContainerMock, nil)
	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.PgQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	m.PgQueierMock.On("GetDatabaseTableSchema", mock.Anything, mock.Anything, &pg_queries.GetDatabaseTableSchemaParams{
		Schema: "public",
		Table:  "users",
	}).Return([]*pg_queries.GetDatabaseTableSchemaRow{
		{
			ColumnName:      "id",
			DataType:        "uuid",
			OrdinalPosition: 1,
			IsNullable:      "NO",
			ColumnDefault:   "gen_random_uuid()",
		},
		{
			ColumnName:             "name",
			DataType:               "varchar(40)",
			OrdinalPosition:        6,
			IsNullable:             "YES",
			CharacterMaximumLength: 40,
		},
	}, nil)

	m.PgQueierMock.On("GetTableConstraints", mock.Anything, mock.Anything, &pg_queries.GetTableConstraintsParams{
		Schema: "public",
		Table:  "users",
	}).Return([]*pg_queries.GetTableConstraintsRow{
		{
			ConstraintName:       "users_pkey",
			ConstraintDefinition: "PRIMARY KEY (id)",
		},
	}, nil)

	resp, err := m.Service.GetConnectionInitStatements(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionInitStatementsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionInitStatementsRequest{
			ConnectionId: mockConnectionId,
			Options: &mgmtv1alpha1.InitStatementOptions{
				InitSchema:           true,
				TruncateBeforeInsert: false,
				TruncateCascade:      false,
			},
		},
	})

	expectedInit := "CREATE TABLE IF NOT EXISTS \"public\".\"users\" (\"id\" uuid NOT NULL DEFAULT gen_random_uuid(), \"name\" varchar(40) NULL, CONSTRAINT users_pkey PRIMARY KEY (id));"
	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableInitStatements, 1)
	assert.Len(t, resp.Msg.TableTruncateStatements, 0)
	assert.Equal(t, expectedInit, resp.Msg.TableInitStatements["public.users"])
}

func Test_GetConnectionInitStatements_Postgres_Truncate(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	pool, _ := pgxpool.New(context.Background(), "")
	m.PgPoolContainerMock.On("Open", mock.Anything).Return(pool, nil)
	m.PgPoolContainerMock.On("Close")
	m.SqlConnectorMock.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(m.PgPoolContainerMock, nil)
	connection := getConnectionMock(mockAccountId, mockConnectionName, mockConnectionId, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.ConnectionServiceMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: connection,
	}), nil)
	m.PgQueierMock.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)

	resp, err := m.Service.GetConnectionInitStatements(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionInitStatementsRequest]{
		Msg: &mgmtv1alpha1.GetConnectionInitStatementsRequest{
			ConnectionId: mockConnectionId,
			Options: &mgmtv1alpha1.InitStatementOptions{
				InitSchema:           false,
				TruncateBeforeInsert: true,
				TruncateCascade:      true,
			},
		},
	})

	expectedTruncate := "TRUNCATE TABLE \"public\".\"users\" CASCADE;"
	assert.Nil(t, err)
	assert.Len(t, resp.Msg.TableInitStatements, 0)
	assert.Len(t, resp.Msg.TableTruncateStatements, 1)
	assert.Equal(t, expectedTruncate, resp.Msg.TableTruncateStatements["public.users"])
}

type serviceMocks struct {
	Service                *Service
	DbtxMock               *nucleusdb.MockDBTX
	QuerierMock            *db_queries.MockQuerier
	UserAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceMock  *mgmtv1alpha1connect.MockConnectionServiceClient
	JobServiceMock         *mgmtv1alpha1connect.MockJobServiceHandler
	SqlMock                sqlmock.Sqlmock
	SqlDbMock              *sql.DB
	SqlDbContainerMock     *sqlconnect.MockSqlDbContainer
	PgPoolContainerMock    *sqlconnect.MockPgPoolContainer
	PgQueierMock           *pg_queries.MockQuerier
	MysqlQueierMock        *mysql_queries.MockQuerier
	SqlConnectorMock       *sqlconnect.MockSqlConnector
	AwsManagerMock         *awsmanager.MockNeosyncAwsManagerClient
}

func createServiceMock(t *testing.T) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockJobService := mgmtv1alpha1connect.NewMockJobServiceHandler(t)
	mockPgquerier := pg_queries.NewMockQuerier(t)
	mockMysqlquerier := mysql_queries.NewMockQuerier(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	mockAwsManager := awsmanager.NewMockNeosyncAwsManagerClient(t)

	sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	service := New(&Config{}, mockUserAccountService, mockConnectionService, mockJobService, mockAwsManager, mockSqlConnector, mockPgquerier, mockMysqlquerier)

	return &serviceMocks{
		Service:                service,
		DbtxMock:               mockDbtx,
		QuerierMock:            mockQuerier,
		UserAccountServiceMock: mockUserAccountService,
		ConnectionServiceMock:  mockConnectionService,
		JobServiceMock:         mockJobService,
		SqlMock:                sqlMock,
		SqlDbMock:              sqlDbMock,
		SqlDbContainerMock:     sqlconnect.NewMockSqlDbContainer(t),
		PgPoolContainerMock:    sqlconnect.NewMockPgPoolContainer(t),
		PgQueierMock:           mockPgquerier,
		MysqlQueierMock:        mockMysqlquerier,
		SqlConnectorMock:       mockSqlConnector,
		AwsManagerMock:         mockAwsManager,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) { //nolint
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

//nolint:all
func getConnectionMock(accountId, name string, id string, connType ConnTypeMock) *mgmtv1alpha1.Connection {
	timestamp := timestamppb.New(time.Now())
	connection := &mgmtv1alpha1.Connection{
		AccountId:       accountId,
		Name:            name,
		Id:              id,
		CreatedByUserId: mockUserId,
		UpdatedByUserId: mockUserId,
		CreatedAt:       timestamp,
		UpdatedAt:       timestamp,
	}
	if connType == MysqlMock {
		connection.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.MysqlConnection{
							Host:     "host",
							Port:     5432,
							Name:     "database",
							User:     "user",
							Pass:     "topsecret",
							Protocol: "tcp",
						},
					},
				},
			},
		}
	} else if connType == PostgresMock {
		sslMode := "disable"
		connection.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.PostgresConnection{
							Host:    "host",
							Port:    5432,
							Name:    "database",
							User:    "user",
							Pass:    "topsecret",
							SslMode: &sslMode,
						},
					},
				},
			},
		}

	} else if connType == AwsS3Mock {
		connection.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
				AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
					Bucket: "neosync",
				},
			},
		}
	}

	return connection
}

func gzipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Test_isValidTable(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		columns  []*mgmtv1alpha1.DatabaseColumn
		expected bool
	}{
		{
			name:  "table exists",
			table: "users",
			columns: []*mgmtv1alpha1.DatabaseColumn{
				{Table: "users"},
				{Table: "orders"},
			},
			expected: true,
		},
		{
			name:  "table does not exist",
			table: "payments",
			columns: []*mgmtv1alpha1.DatabaseColumn{
				{Table: "users"},
				{Table: "orders"},
			},
			expected: false,
		},
		{
			name:     "empty table name",
			table:    "",
			columns:  []*mgmtv1alpha1.DatabaseColumn{{Table: "users"}, {Table: "orders"}},
			expected: false,
		},
		{
			name:     "empty columns slice",
			table:    "users",
			columns:  []*mgmtv1alpha1.DatabaseColumn{},
			expected: false,
		},
		{
			name:     "nil columns slice",
			table:    "users",
			columns:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidTable(tt.table, tt.columns)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_isValidSchema(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		columns  []*mgmtv1alpha1.DatabaseColumn
		expected bool
	}{
		{
			name:   "Schema exists",
			schema: "users",
			columns: []*mgmtv1alpha1.DatabaseColumn{
				{Schema: "users"},
				{Schema: "orders"},
			},
			expected: true,
		},
		{
			name:   "table does not exist",
			schema: "payments",
			columns: []*mgmtv1alpha1.DatabaseColumn{
				{Schema: "users"},
				{Schema: "orders"},
			},
			expected: false,
		},
		{
			name:     "empty table name",
			schema:   "",
			columns:  []*mgmtv1alpha1.DatabaseColumn{{Schema: "users"}, {Schema: "orders"}},
			expected: false,
		},
		{
			name:     "empty columns slice",
			schema:   "users",
			columns:  []*mgmtv1alpha1.DatabaseColumn{},
			expected: false,
		},
		{
			name:     "nil columns slice",
			schema:   "users",
			columns:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidSchema(tt.schema, tt.columns)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
