package sqlmanager_mysql

import (
	context "context"
	"database/sql"
	"fmt"
	"testing"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GetDatabaseSchema_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	mysqlquerier.On("GetDatabaseSchema", mock.Anything, mockPool).Return(
		[]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema:            "public",
				TableName:              "users",
				ColumnName:             "id",
				DataType:               "varchar",
				ColumnDefault:          "",
				IsNullable:             "NO",
				CharacterMaximumLength: sql.NullInt64{Int64: int64(-1)},
				NumericScale:           sql.NullInt64{Int64: int64(-1)},
				OrdinalPosition:        4,
			},
			{
				TableSchema:            "public",
				TableName:              "orders",
				ColumnName:             "buyer_id",
				DataType:               "integer",
				ColumnDefault:          "",
				IsNullable:             "NO",
				CharacterMaximumLength: sql.NullInt64{Int64: int64(32)},
				NumericScale:           sql.NullInt64{Int64: int64(0)},
				OrdinalPosition:        5,
			},
		}, nil,
	)

	expected := []*sqlmanager_shared.DatabaseSchemaRow{
		{
			TableSchema:   "public",
			TableName:     "users",
			ColumnName:    "id",
			DataType:      "varchar",
			ColumnDefault: "",
			IsNullable:    "NO",
		},
		{
			TableSchema:   "public",
			TableName:     "orders",
			ColumnName:    "buyer_id",
			DataType:      "integer",
			ColumnDefault: "",
			IsNullable:    "NO",
		},
	}

	actual, err := manager.GetDatabaseSchema(context.Background())
	require.NoError(t, err)
	require.ElementsMatch(t, expected, actual)
}

func Test_GetForeignKeyConstraintsMap_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	constraints := []*mysql_queries.GetForeignKeyConstraintsRow{
		{ConstraintName: "fk_account_user_associations_account_id", SchemaName: "neosync_api", TableName: "account_user_associations", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_account_user_associations_user_id", SchemaName: "neosync_api", TableName: "account_user_associations", ColumnName: "user_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_connections_accounts_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_connections_created_by_users_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "created_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "YES"},
		{ConstraintName: "fk_connections_updated_by_users_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "updated_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobdstconassoc_conn_id_conn_id", SchemaName: "neosync_api", TableName: "job_destination_connection_associations", ColumnName: "connection_id", ForeignSchemaName: "neosync_api", ForeignTableName: "connections", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobdstconassoc_job_id_jobs_id", SchemaName: "neosync_api", TableName: "job_destination_connection_associations", ColumnName: "job_id", ForeignSchemaName: "neosync_api", ForeignTableName: "jobs", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobs_accounts_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobs_conn_accounts_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "connection_source_id", ForeignSchemaName: "neosync_api", ForeignTableName: "connections", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobs_created_by_users_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "created_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "YES"},
		{ConstraintName: "fk_jobs_updated_by_users_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "updated_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_user_identity_provider_user_id", SchemaName: "neosync_api", TableName: "user_identity_provider_associations", ColumnName: "user_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
	}
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mockPool, "neosync_api").Return(constraints, nil)

	expected := map[string][]*sqlmanager_shared.ForeignConstraint{
		"neosync_api.account_user_associations": {
			{Columns: []string{"account_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.accounts", Columns: []string{"id"}}},
			{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.users", Columns: []string{"id"}}},
		},
		"neosync_api.connections": {
			{Columns: []string{"account_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.accounts", Columns: []string{"id"}}},
			{Columns: []string{"created_by_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.users", Columns: []string{"id"}}},
			{Columns: []string{"updated_by_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.users", Columns: []string{"id"}}},
		},
		"neosync_api.job_destination_connection_associations": {
			{Columns: []string{"connection_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.connections", Columns: []string{"id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.jobs", Columns: []string{"id"}}},
		},
		"neosync_api.jobs": {
			{Columns: []string{"account_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.accounts", Columns: []string{"id"}}},
			{Columns: []string{"connection_source_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.connections", Columns: []string{"id"}}},
			{Columns: []string{"created_by_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.users", Columns: []string{"id"}}},
			{Columns: []string{"updated_by_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.users", Columns: []string{"id"}}},
		},
		"neosync_api.user_identity_provider_associations": {
			{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.users", Columns: []string{"id"}}},
		},
	}

	actual, err := manager.GetForeignKeyConstraintsMap(context.Background(), []string{"neosync_api"})
	require.NoError(t, err)
	for table, fks := range expected {
		acutalFks := actual[table]
		require.ElementsMatch(t, fks, acutalFks)
	}
}

func Test_GetForeignKeyConstraintsMap_ExtraEdgeCases_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	constraints := []*mysql_queries.GetForeignKeyConstraintsRow{
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "account_id", IsNullable: "NO"},
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "c", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "user_id", IsNullable: "NO"},
		{ConstraintName: "t2_b_fkey", SchemaName: "neosync_api", TableName: "t2", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t2", ForeignColumnName: "a", IsNullable: "NO"},
		{ConstraintName: "t3_b_fkey", SchemaName: "neosync_api", TableName: "t3", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t4", ForeignColumnName: "a", IsNullable: "NO"},
		{ConstraintName: "t4_b_fkey", SchemaName: "neosync_api", TableName: "t4", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t3", ForeignColumnName: "a", IsNullable: "NO"},
	}
	expected := map[string][]*sqlmanager_shared.ForeignConstraint{
		"neosync_api.t1": {
			{Columns: []string{"b", "c"}, NotNullable: []bool{true, true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.account_user_associations", Columns: []string{"account_id", "user_id"}}},
		},
		"neosync_api.t2": {
			{Columns: []string{"b"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.t2", Columns: []string{"a"}}},
		},
		"neosync_api.t3": {
			{Columns: []string{"b"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.t4", Columns: []string{"a"}}},
		},
		"neosync_api.t4": {
			{Columns: []string{"b"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "neosync_api.t3", Columns: []string{"a"}}},
		},
	}

	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mockPool, "neosync_api").Return(constraints, nil)

	actual, err := manager.GetForeignKeyConstraintsMap(context.Background(), []string{"neosync_api"})
	require.NoError(t, err)
	for table, fks := range expected {
		acutalFks := actual[table]
		require.ElementsMatch(t, fks, acutalFks)
	}
}

func Test_GetPrimaryKeyConstraintsMap_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	schemas := []string{"public"}

	mysqlquerier.On("GetPrimaryKeyConstraints", mock.Anything, mockPool, "public").Return(
		[]*mysql_queries.GetPrimaryKeyConstraintsRow{
			{
				SchemaName:     "public",
				TableName:      "users",
				ColumnName:     "id",
				ConstraintName: "users-id",
			},
			{
				SchemaName:     "public",
				TableName:      "accounts",
				ColumnName:     "id",
				ConstraintName: "accounts-id",
			},
			{
				SchemaName:     "public",
				TableName:      "orders",
				ColumnName:     "id",
				ConstraintName: "orders-id",
			},
			{
				SchemaName:     "public",
				TableName:      "composite",
				ColumnName:     "id",
				ConstraintName: "composite-id",
			},
			{
				SchemaName:     "public",
				TableName:      "composite",
				ColumnName:     "other_id",
				ConstraintName: "composite-id",
			},
		}, nil,
	)

	expected := map[string][]string{
		"public.users":     {"id"},
		"public.accounts":  {"id"},
		"public.orders":    {"id"},
		"public.composite": {"id", "other_id"},
	}

	actual, err := manager.GetPrimaryKeyConstraintsMap(context.Background(), schemas)
	require.NoError(t, err)
	for table, expect := range expected {
		require.ElementsMatch(t, expect, actual[table])
	}
}

func Test_GetUniqueConstraintsMap_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	schemas := []string{"public"}

	mysqlquerier.On("GetUniqueConstraints", mock.Anything, mockPool, "public").Return(
		[]*mysql_queries.GetUniqueConstraintsRow{
			{
				SchemaName:     "public",
				TableName:      "person",
				ColumnName:     "name",
				ConstraintName: "person-name-email",
			},
			{
				SchemaName:     "public",
				TableName:      "person",
				ColumnName:     "email",
				ConstraintName: "person-name-email",
			},
			{
				SchemaName:     "public",
				TableName:      "region",
				ColumnName:     "code",
				ConstraintName: "region-code",
			},
			{
				SchemaName:     "public",
				TableName:      "region",
				ColumnName:     "name",
				ConstraintName: "region-name",
			},
		}, nil,
	)

	expected := map[string][][]string{
		"public.person": {{"name", "email"}},
		"public.region": {{"code"}, {"name"}},
	}

	actual, err := manager.GetUniqueConstraintsMap(context.Background(), schemas)

	require.NoError(t, err)
	for table, cols := range expected {
		actualCols := actual[table]
		require.Len(t, actualCols, len(cols))
		for _, col := range cols {
			require.Contains(t, actualCols, col)
		}
	}
}

func Test_GetRolePermissionsMap_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	mysqlquerier.On("GetMysqlRolePermissions", mock.Anything, mockPool, "postgres").Return(
		[]*mysql_queries.GetMysqlRolePermissionsRow{
			{TableSchema: "public", TableName: "users", PrivilegeType: "INSERT"},
			{TableSchema: "public", TableName: "users", PrivilegeType: "UPDATE"},
			{TableSchema: "person", TableName: "users", PrivilegeType: "DELETE"},
			{TableSchema: "other", TableName: "accounts", PrivilegeType: "INSERT"},
		}, nil,
	)

	expected := map[string][]string{
		"public.users":   {"INSERT", "UPDATE"},
		"person.users":   {"DELETE"},
		"other.accounts": {"INSERT"},
	}

	actual, err := manager.GetRolePermissionsMap(context.Background(), "postgres")
	require.NoError(t, err)
	for table, expect := range expected {
		require.ElementsMatch(t, expect, actual[table])
	}
}

func Test_BatchExec_Mysql(t *testing.T) {
	prefix := sqlmanager_shared.DisableForeignKeyChecks
	tests := []struct {
		name          string
		batchSize     int
		statements    []string
		opts          *sqlmanager_shared.BatchExecOpts
		expectedCalls []string
	}{
		{
			name:          "multiple batches",
			batchSize:     2,
			statements:    []string{"CREATE TABLE users;", "CREATE TABLE accounts;", "CREATE TABLE departments;"},
			expectedCalls: []string{"CREATE TABLE users; CREATE TABLE accounts;", "CREATE TABLE departments;"},
		},
		{
			name:          "single statement",
			batchSize:     2,
			statements:    []string{"TRUNCATE TABLE users;"},
			expectedCalls: []string{"TRUNCATE TABLE users;"},
		},
		{
			name:          "multiple batches prefix",
			batchSize:     2,
			statements:    []string{"CREATE TABLE users;", "CREATE TABLE accounts;", "CREATE TABLE departments;"},
			expectedCalls: []string{fmt.Sprintf("%s %s", prefix, "CREATE TABLE users; CREATE TABLE accounts;"), fmt.Sprintf("%s %s", prefix, "CREATE TABLE departments;")},
			opts:          &sqlmanager_shared.BatchExecOpts{Prefix: &prefix},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlquerier := mysql_queries.NewMockQuerier(t)
			mockPool := mysql_queries.NewMockDBTX(t)
			manager := MysqlManager{
				querier: mysqlquerier,
				pool:    mockPool,
			}

			for _, call := range tt.expectedCalls {
				mockPool.On("ExecContext", mock.Anything, call).Return(nil, nil)
			}

			err := manager.BatchExec(context.Background(), tt.batchSize, tt.statements, tt.opts)
			require.NoError(t, err)
		})
	}
}

func Test_Exec_Mysql(t *testing.T) {
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	mockPool := mysql_queries.NewMockDBTX(t)
	manager := MysqlManager{
		querier: mysqlquerier,
		pool:    mockPool,
	}

	stmt := "TRUNCATE TABLE users;"
	mockPool.On("ExecContext", mock.Anything, stmt).Return(nil, nil)

	err := manager.Exec(context.Background(), stmt)
	require.NoError(t, err)
}

func Test_EscapeMysqlColumns(t *testing.T) {
	require.Empty(t, EscapeMysqlColumns(nil))
	require.Equal(
		t,
		EscapeMysqlColumns([]string{"foo", "bar", "baz"}),
		[]string{"`foo`", "`bar`", "`baz`"},
	)
}

func Test_BuildMysqlTruncateStatement(t *testing.T) {
	actual, err := BuildMysqlTruncateStatement("public", "users")
	require.NoError(t, err)
	require.Equal(
		t,
		`TRUNCATE "public"."users";`,
		actual,
	)
}
