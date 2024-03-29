package dbschemas_mysql

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"

	"github.com/stretchr/testify/assert"
)

func Test_GetMysqlTableDependencies(t *testing.T) {
	constraints := []*mysql_queries.GetForeignKeyConstraintsRow{
		{ConstraintName: "fk_account_user_associations_account_id", SchemaName: "neosync_api", TableName: "account_user_associations", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_account_user_associations_user_id", SchemaName: "neosync_api", TableName: "account_user_associations", ColumnName: "user_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_connections_accounts_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_connections_created_by_users_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "created_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "YES"},
		{ConstraintName: "fk_connections_updated_by_users_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "updated_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobdstconassoc_conn_id_conn_id", SchemaName: "neosync_api", TableName: "job_destination_connection_associations", ColumnName: "connection_id", ForeignSchemaName: "neosync_api", ForeignTableName: "connections", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobdstconassoc_job_id_jobs_id", SchemaName: "neosync_api", TableName: "job_destination_connection_associations", ColumnName: "job_id", ForeignSchemaName: "neosync_api", ForeignTableName: "jobs", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobs_accounts_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobs_accounts_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "connection_source_id", ForeignSchemaName: "neosync_api", ForeignTableName: "connections", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_jobs_created_by_users_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "created_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "YES"},
		{ConstraintName: "fk_jobs_updated_by_users_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "updated_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
		{ConstraintName: "fk_user_identity_provider_user_id", SchemaName: "neosync_api", TableName: "user_identity_provider_associations", ColumnName: "user_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id", IsNullable: "NO"},
	}

	td := GetMysqlTableDependencies(constraints)
	assert.Equal(t, td, dbschemas.TableDependency{
		"neosync_api.account_user_associations": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "account_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.accounts", Column: "id"}},
			{Column: "user_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.users", Column: "id"}},
		}},
		"neosync_api.connections": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "account_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.accounts", Column: "id"}},
			{Column: "created_by_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.users", Column: "id"}},
			{Column: "updated_by_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.users", Column: "id"}},
		}},
		"neosync_api.job_destination_connection_associations": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "connection_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.connections", Column: "id"}},
			{Column: "job_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.jobs", Column: "id"}},
		}},
		"neosync_api.jobs": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "account_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.accounts", Column: "id"}},
			{Column: "connection_source_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.connections", Column: "id"}},
			{Column: "created_by_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.users", Column: "id"}},
			{Column: "updated_by_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.users", Column: "id"}},
		}},
		"neosync_api.user_identity_provider_associations": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "user_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.users", Column: "id"}},
		}},
	})
}

func Test_GetMysqlTableDependenciesExtraEdgeCases(t *testing.T) {
	constraints := []*mysql_queries.GetForeignKeyConstraintsRow{
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "account_id", IsNullable: "NO"},
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "c", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "user_id", IsNullable: "NO"},
		{ConstraintName: "t2_b_fkey", SchemaName: "neosync_api", TableName: "t2", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t2", ForeignColumnName: "a", IsNullable: "NO"},
		{ConstraintName: "t3_b_fkey", SchemaName: "neosync_api", TableName: "t3", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t4", ForeignColumnName: "a", IsNullable: "NO"},
		{ConstraintName: "t4_b_fkey", SchemaName: "neosync_api", TableName: "t4", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t3", ForeignColumnName: "a", IsNullable: "NO"},
	}

	td := GetMysqlTableDependencies(constraints)
	assert.Equal(t, td, dbschemas.TableDependency{
		"neosync_api.t1": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "b", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.account_user_associations", Column: "account_id"}},
			{Column: "c", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.account_user_associations", Column: "user_id"}},
		}},
		"neosync_api.t2": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "b", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.t2", Column: "a"}},
		}},
		"neosync_api.t3": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "b", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.t4", Column: "a"}},
		}},
		"neosync_api.t4": {Constraints: []*dbschemas.ForeignConstraint{
			{Column: "b", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "neosync_api.t3", Column: "a"}},
		}},
	}, "Testing composite foreign keys, table self-referencing, and table cycles")
}

func TestGetUniqueSchemaColMappings(t *testing.T) {
	mappings := GetUniqueSchemaColMappings(
		[]*mysql_queries.GetDatabaseSchemaRow{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},
			{TableSchema: "public", TableName: "users", ColumnName: "updated_by"},

			{TableSchema: "neosync_api", TableName: "accounts", ColumnName: "id"},
		},
	)
	assert.Contains(t, mappings, "public.users", "job mappings are a subset of the present database schemas")
	assert.Contains(t, mappings, "neosync_api.accounts", "job mappings are a subset of the present database schemas")
	assert.Contains(t, mappings["public.users"], "id", "")
	assert.Contains(t, mappings["public.users"], "created_by", "")
	assert.Contains(t, mappings["public.users"], "updated_by", "")
	assert.Contains(t, mappings["neosync_api.accounts"], "id", "")
}

func Test_BatchMysqlStmts(t *testing.T) {
	initialStatement := "SET FOREIGN_KEY_CHECKS = 0;"
	tests := []struct {
		name             string
		batchSize        int
		statements       []string
		initialStatement *string
		expectedCalls    []string
	}{
		{
			name:             "multiple batches without initialStatement",
			batchSize:        2,
			statements:       []string{"CREATE TABLE users;", "CREATE TABLE accounts;", "CREATE TABLE departments;"},
			initialStatement: nil,
			expectedCalls:    []string{"CREATE TABLE users; CREATE TABLE accounts;", "CREATE TABLE departments;"},
		},
		{
			name:             "multiple batches with initialStatement",
			batchSize:        2,
			statements:       []string{"TRUNCATE TABLE users;", "TRUNCATE TABLE accounts;", "TRUNCATE TABLE departments;"},
			initialStatement: &initialStatement,
			expectedCalls:    []string{fmt.Sprintf("%s %s", initialStatement, "TRUNCATE TABLE users; TRUNCATE TABLE accounts;"), fmt.Sprintf("%s %s", initialStatement, "TRUNCATE TABLE departments;")},
		},
		{
			name:             "single statement",
			batchSize:        2,
			statements:       []string{"TRUNCATE TABLE users;"},
			initialStatement: nil,
			expectedCalls:    []string{"TRUNCATE TABLE users;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			ctx := context.Background()

			for _, call := range tt.expectedCalls {
				sqlMock.ExpectExec(call).WillReturnResult(sqlmock.NewResult(1, 2))
			}

			err = BatchExecStmts(ctx, sqlDbMock, tt.batchSize, tt.statements, tt.initialStatement)
			assert.NoError(t, err)

			if err := sqlMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_EscapeMysqlColumns(t *testing.T) {
	assert.Empty(t, EscapeMysqlColumns(nil))
	assert.Equal(
		t,
		EscapeMysqlColumns([]string{"foo", "bar", "baz"}),
		[]string{"`foo`", "`bar`", "`baz`"},
	)
}
