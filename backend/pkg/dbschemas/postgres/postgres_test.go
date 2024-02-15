package dbschemas_postgres

import (
	"testing"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/stretchr/testify/assert"
)

func TestGetPostgresTableDependencies(t *testing.T) {
	constraints := []*pg_queries.GetForeignKeyConstraintsRow{
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

	td := GetPostgresTableDependencies(constraints)
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

func TestGetPostgresTableDependenciesExtraEdgeCases(t *testing.T) {
	constraints := []*pg_queries.GetForeignKeyConstraintsRow{
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "account_id", IsNullable: "NO"},
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ColumnName: "c", ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnName: "user_id", IsNullable: "NO"},
		{ConstraintName: "t2_b_fkey", SchemaName: "neosync_api", TableName: "t2", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t2", ForeignColumnName: "a", IsNullable: "NO"},
		{ConstraintName: "t3_b_fkey", SchemaName: "neosync_api", TableName: "t3", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t4", ForeignColumnName: "a", IsNullable: "NO"},
		{ConstraintName: "t4_b_fkey", SchemaName: "neosync_api", TableName: "t4", ColumnName: "b", ForeignSchemaName: "neosync_api", ForeignTableName: "t3", ForeignColumnName: "a", IsNullable: "NO"},
	}

	td := GetPostgresTableDependencies(constraints)
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

func TestGenerateCreateTableStatement(t *testing.T) {
	type testcase struct {
		schema      string
		table       string
		rows        []*pg_queries.GetDatabaseTableSchemaRow
		constraints []*pg_queries.GetTableConstraintsRow
		expected    string
	}
	cases := []testcase{
		{
			schema: "public",
			table:  "users",
			rows: []*pg_queries.GetDatabaseTableSchemaRow{
				{
					ColumnName:      "id",
					DataType:        "uuid",
					OrdinalPosition: 1,
					IsNullable:      "NO",
					ColumnDefault:   "gen_random_uuid()",
				},
				{
					ColumnName:      "created_at",
					DataType:        "timestamp without time zone",
					OrdinalPosition: 2,
					IsNullable:      "NO",
					ColumnDefault:   "now()",
				},
				{
					ColumnName:      "updated_at",
					DataType:        "timestamp",
					OrdinalPosition: 3,
					IsNullable:      "NO",
					ColumnDefault:   "CURRENT_TIMESTAMP",
				},
				{
					ColumnName:      "extra",
					DataType:        "varchar",
					OrdinalPosition: 5,
					IsNullable:      "YES",
				},
				{
					ColumnName:             "name",
					DataType:               "varchar(40)",
					OrdinalPosition:        6,
					IsNullable:             "YES",
					CharacterMaximumLength: 40,
				},
			},
			constraints: []*pg_queries.GetTableConstraintsRow{
				{
					ConstraintName:       "users_pkey",
					ConstraintDefinition: "PRIMARY KEY (id)",
				},
			},
			expected: `CREATE TABLE IF NOT EXISTS public.users ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "created_at" timestamp without time zone NOT NULL DEFAULT now(), "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "extra" varchar NULL, "name" varchar(40) NULL, CONSTRAINT users_pkey PRIMARY KEY (id));`,
		},
		{
			schema: "public",
			table:  "users",
			rows: []*pg_queries.GetDatabaseTableSchemaRow{
				{
					ColumnName:      "id",
					DataType:        "integer",
					OrdinalPosition: 1,
					IsNullable:      "NO",
					ColumnDefault:   "nextval('users_id_seq'::regclass)",
				},
				{
					ColumnName:      "id2",
					DataType:        "smallint",
					OrdinalPosition: 2,
					IsNullable:      "NO",
					ColumnDefault:   "nextval('users_id2_seq'::regclass)",
				},
				{
					ColumnName:      "id3",
					DataType:        "bigint",
					OrdinalPosition: 3,
					IsNullable:      "NO",
					ColumnDefault:   "nextval('users_id3_seq'::regclass)",
				},
			},
			constraints: []*pg_queries.GetTableConstraintsRow{},
			expected:    `CREATE TABLE IF NOT EXISTS public.users ("id" SERIAL NOT NULL, "id2" SMALLSERIAL NOT NULL, "id3" BIGSERIAL NOT NULL);`,
		},
	}

	for _, testcase := range cases {
		t.Run(t.Name(), func(t *testing.T) {
			actual := generateCreateTableStatement(testcase.schema, testcase.table, testcase.rows, testcase.constraints)
			assert.Equal(t, testcase.expected, actual)
		})
	}
}

func TestGetUniqueSchemaColMappings(t *testing.T) {
	mappings := GetUniqueSchemaColMappings(
		[]*pg_queries.GetDatabaseSchemaRow{
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
