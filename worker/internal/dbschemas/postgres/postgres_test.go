package dbschemas_postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPostgresTableDependencies(t *testing.T) {
	constraints := []*ForeignKeyConstraint{
		{ConstraintName: "fk_account_user_associations_account_id", SchemaName: "neosync_api", TableName: "account_user_associations", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id"},               //nolint
		{ConstraintName: "fk_account_user_associations_user_id", SchemaName: "neosync_api", TableName: "account_user_associations", ColumnName: "user_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id"},                        //nolint
		{ConstraintName: "fk_connections_accounts_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id"},                                          //nolint
		{ConstraintName: "fk_connections_created_by_users_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "created_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id"},                                  //nolint
		{ConstraintName: "fk_connections_updated_by_users_id", SchemaName: "neosync_api", TableName: "connections", ColumnName: "updated_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id"},                                  //nolint
		{ConstraintName: "fk_jobdstconassoc_conn_id_conn_id", SchemaName: "neosync_api", TableName: "job_destination_connection_associations", ColumnName: "connection_id", ForeignSchemaName: "neosync_api", ForeignTableName: "connections", ForeignColumnName: "id"}, //nolint
		{ConstraintName: "fk_jobdstconassoc_job_id_jobs_id", SchemaName: "neosync_api", TableName: "job_destination_connection_associations", ColumnName: "job_id", ForeignSchemaName: "neosync_api", ForeignTableName: "jobs", ForeignColumnName: "id"},                //nolint
		{ConstraintName: "fk_jobs_accounts_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "account_id", ForeignSchemaName: "neosync_api", ForeignTableName: "accounts", ForeignColumnName: "id"},                                                        //nolint
		{ConstraintName: "fk_jobs_accounts_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "connection_source_id", ForeignSchemaName: "neosync_api", ForeignTableName: "connections", ForeignColumnName: "id"},                                           //nolint
		{ConstraintName: "fk_jobs_created_by_users_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "created_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id"},                                                //nolint
		{ConstraintName: "fk_jobs_updated_by_users_id", SchemaName: "neosync_api", TableName: "jobs", ColumnName: "updated_by_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id"},                                                //nolint
		{ConstraintName: "fk_user_identity_provider_user_id", SchemaName: "neosync_api", TableName: "user_identity_provider_associations", ColumnName: "user_id", ForeignSchemaName: "neosync_api", ForeignTableName: "users", ForeignColumnName: "id"},                 //nolint
	}

	td := GetPostgresTableDependencies(constraints)
	assert.Equal(t, td, TableDependency{
		"neosync_api.account_user_associations":               {"neosync_api.accounts", "neosync_api.users"},
		"neosync_api.connections":                             {"neosync_api.accounts", "neosync_api.users"},
		"neosync_api.job_destination_connection_associations": {"neosync_api.connections", "neosync_api.jobs"},
		"neosync_api.jobs":                                    {"neosync_api.accounts", "neosync_api.connections", "neosync_api.users"},
		"neosync_api.user_identity_provider_associations":     {"neosync_api.users"},
	})
}

func TestGetPostgresTableDependenciesExtraEdgeCases(t *testing.T) {
	constraints := []*ForeignKeyConstraint{
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "account_id"}, //nolint
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "c", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "user_id"},    //nolint
		{ConstraintName: "t2_b_fkey", SchemaName: "neosync_api", TableName: "t2", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t2", ForeignColumnName: "a"},                                   //nolint
		{ConstraintName: "t3_b_fkey", SchemaName: "neosync_api", TableName: "t3", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t4", ForeignColumnName: "a"},                                   //nolint
		{ConstraintName: "t4_b_fkey", SchemaName: "neosync_api", TableName: "t4", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t3", ForeignColumnName: "a"},                                   //nolint
	}

	td := GetPostgresTableDependencies(constraints)
	assert.Equal(t, td, TableDependency{
		"neosync_api.t1": {"neosync_api.account_user_associations"},
		"neosync_api.t2": {"neosync_api.t2"},
		"neosync_api.t3": {"neosync_api.t4"},
		"neosync_api.t4": {"neosync_api.t3"},
	}, "Testing composite foreign keys, table self-referencing, and table cycles")
}

func TestGenerateCreateTableStatement(t *testing.T) {
	result := generateCreateTableStatement(
		"public", "users",
		[]*DatabaseSchema{
			{
				ColumnName:      "id",
				DataType:        "uuid",
				OrdinalPosition: 1,
				IsNullable:      "NO",
				ColumnDefault:   strPtr("gen_random_uuid()"),
			},
			{
				ColumnName:      "created_at",
				DataType:        "timestamp without time zone",
				OrdinalPosition: 2,
				IsNullable:      "NO",
				ColumnDefault:   strPtr("now()"),
			},
			{ // ordinal position intentionally out of order to ensure the algorithm corrects that
				ColumnName:      "extra",
				DataType:        "varchar",
				OrdinalPosition: 4,
				IsNullable:      "YES",
			},
			{
				ColumnName:      "updated_at",
				DataType:        "timestamp",
				OrdinalPosition: 3,
				IsNullable:      "NO",
				ColumnDefault:   strPtr("CURRENT_TIMESTAMP"),
			},
		},
		[]*DatabaseTableConstraint{
			{
				ConstraintName:       "users_pkey",
				ConstraintDefinition: "PRIMARY KEY (id)",
			},
		},
	)

	assert.Equal(
		t,
		result,
		"CREATE TABLE IF NOT EXISTS public.users (id uuid NOT NULL DEFAULT gen_random_uuid(), created_at timestamp without time zone NOT NULL DEFAULT now(), updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, extra varchar NULL, CONSTRAINT users_pkey PRIMARY KEY (id));", //nolint
	)
}

func strPtr(val string) *string {
	return &val
}
