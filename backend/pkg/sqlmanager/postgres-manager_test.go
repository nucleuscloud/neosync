package sqlmanager

import (
	context "context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GetDatabaseSchema(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	pgquerier.On("GetDatabaseSchema", mock.Anything, mockPool).Return(
		[]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema:            "public",
				TableName:              "users",
				ColumnName:             "id",
				DataType:               "varchar",
				ColumnDefault:          "",
				IsNullable:             "NO",
				CharacterMaximumLength: 220,
				NumericPrecision:       -1,
				NumericScale:           -1,
				OrdinalPosition:        4,
			},
			{
				TableSchema:            "public",
				TableName:              "orders",
				ColumnName:             "buyer_id",
				DataType:               "integer",
				ColumnDefault:          "",
				IsNullable:             "NO",
				CharacterMaximumLength: -1,
				NumericPrecision:       32,
				NumericScale:           0,
				OrdinalPosition:        5,
			},
		}, nil,
	)

	expected := []*DatabaseSchemaRow{
		{
			TableSchema:            "public",
			TableName:              "users",
			ColumnName:             "id",
			DataType:               "varchar",
			ColumnDefault:          "",
			IsNullable:             "NO",
			CharacterMaximumLength: 220,
			NumericPrecision:       -1,
			NumericScale:           -1,
			OrdinalPosition:        4,
		},
		{
			TableSchema:            "public",
			TableName:              "orders",
			ColumnName:             "buyer_id",
			DataType:               "integer",
			ColumnDefault:          "",
			IsNullable:             "NO",
			CharacterMaximumLength: -1,
			NumericPrecision:       32,
			NumericScale:           0,
			OrdinalPosition:        5,
		},
	}

	actual, err := manager.GetDatabaseSchema(context.Background())
	require.NoError(t, err)
	require.ElementsMatch(t, expected, actual)
}

func Test_GetForeignKeyConstraintsMap(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	schemas := []string{"public"}

	pgquerier.On("GetTableConstraintsBySchema", mock.Anything, mockPool, schemas).Return(
		mockTableConstraintsRows(), nil,
	)

	expected := map[string][]*ForeignConstraint{
		"public.orders": {
			{Column: "buyer_id", IsNullable: false, ForeignKey: &ForeignKey{
				Table:  "public.users",
				Column: "id",
			},
			},
			{Column: "account_id", IsNullable: true, ForeignKey: &ForeignKey{
				Table:  "public.accounts",
				Column: "id",
			}},
		},
		"public.users": {
			{Column: "composite_id", IsNullable: false, ForeignKey: &ForeignKey{
				Table:  "public.composite",
				Column: "id",
			},
			},
			{Column: "other_composite_id", IsNullable: false, ForeignKey: &ForeignKey{
				Table:  "public.composite",
				Column: "other_id",
			}},
		},
		"public.accounts": {
			{Column: "user_id", IsNullable: false, ForeignKey: &ForeignKey{
				Table:  "public.users",
				Column: "id",
			},
			},
		}}

	actual, err := manager.GetForeignKeyConstraintsMap(context.Background(), schemas)
	require.NoError(t, err)
	for table, expect := range expected {
		require.ElementsMatch(t, expect, actual[table])
	}
}

func Test_GetForeignKeyConstraintsMap_ExtraEdgeCases(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	schemas := []string{"public"}

	constraints := []*pg_queries.GetTableConstraintsBySchemaRow{
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ConstraintColumns: []string{"b"}, ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnNames: []string{"account_id"}, Notnullable: []bool{true}, ConstraintType: "f"},
		{ConstraintName: "t1_b_c_fkey", SchemaName: "neosync_api", TableName: "t1", ConstraintColumns: []string{"c"}, ForeignSchemaName: "neosync_api", ForeignTableName: "account_user_associations", ForeignColumnNames: []string{"user_id"}, Notnullable: []bool{true}, ConstraintType: "f"},
		{ConstraintName: "t2_b_fkey", SchemaName: "neosync_api", TableName: "t2", ConstraintColumns: []string{"b"}, ForeignSchemaName: "neosync_api", ForeignTableName: "t2", ForeignColumnNames: []string{"a"}, Notnullable: []bool{true}, ConstraintType: "f"},
		{ConstraintName: "t3_b_fkey", SchemaName: "neosync_api", TableName: "t3", ConstraintColumns: []string{"b"}, ForeignSchemaName: "neosync_api", ForeignTableName: "t4", ForeignColumnNames: []string{"a"}, Notnullable: []bool{true}, ConstraintType: "f"},
		{ConstraintName: "t4_b_fkey", SchemaName: "neosync_api", TableName: "t4", ConstraintColumns: []string{"b"}, ForeignSchemaName: "neosync_api", ForeignTableName: "t3", ForeignColumnNames: []string{"a"}, Notnullable: []bool{true}, ConstraintType: "f"},
	}
	pgquerier.On("GetTableConstraintsBySchema", mock.Anything, mockPool, schemas).Return(
		constraints, nil,
	)

	actual, err := manager.GetForeignKeyConstraintsMap(context.Background(), schemas)
	require.NoError(t, err)
	require.Equal(t, actual, map[string][]*ForeignConstraint{
		"neosync_api.t1": {
			{Column: "b", IsNullable: false, ForeignKey: &ForeignKey{Table: "neosync_api.account_user_associations", Column: "account_id"}},
			{Column: "c", IsNullable: false, ForeignKey: &ForeignKey{Table: "neosync_api.account_user_associations", Column: "user_id"}},
		},
		"neosync_api.t2": {
			{Column: "b", IsNullable: false, ForeignKey: &ForeignKey{Table: "neosync_api.t2", Column: "a"}},
		},
		"neosync_api.t3": {
			{Column: "b", IsNullable: false, ForeignKey: &ForeignKey{Table: "neosync_api.t4", Column: "a"}},
		},
		"neosync_api.t4": {
			{Column: "b", IsNullable: false, ForeignKey: &ForeignKey{Table: "neosync_api.t3", Column: "a"}},
		},
	}, "Testing composite foreign keys, table self-referencing, and table cycles")
}

func Test_GetPrimaryKeyConstraintsMap(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	schemas := []string{"public"}

	pgquerier.On("GetTableConstraintsBySchema", mock.Anything, mockPool, schemas).Return(
		mockTableConstraintsRows(), nil,
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

func Test_GetUniqueConstraintsMap(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	schemas := []string{"public"}

	pgquerier.On("GetTableConstraintsBySchema", mock.Anything, mockPool, schemas).Return(
		mockTableConstraintsRows(), nil,
	)

	expected := map[string][][]string{
		"public.person": {{"name", "email"}},
		"public.region": {{"code"}, {"name"}},
	}

	actual, err := manager.GetUniqueConstraintsMap(context.Background(), schemas)

	require.NoError(t, err)
	require.Len(t, actual["public.person"], 1)
	require.Len(t, actual["public.region"], 2)
	require.ElementsMatch(t, expected["public.person"][0], actual["public.person"][0])
	require.ElementsMatch(t, expected["public.region"][0], actual["public.region"][0])
	require.ElementsMatch(t, expected["public.region"][1], actual["public.region"][1])
}

func Test_GetRolePermissionsMap(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	pgquerier.On("GetPostgresRolePermissions", mock.Anything, mockPool, "postgres").Return(
		[]*pg_queries.GetPostgresRolePermissionsRow{
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

func Test_GetCreateTableStatement(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	pgquerier.On("GetDatabaseTableSchema", mock.Anything, mockPool, &pg_queries.GetDatabaseTableSchemaParams{
		Schema: "public",
		Table:  "users",
	}).Return(
		[]*pg_queries.GetDatabaseTableSchemaRow{
			{
				SchemaName:             "public",
				TableName:              "users",
				ColumnName:             "id",
				DataType:               "varchar",
				ColumnDefault:          "",
				IsNullable:             "NO",
				CharacterMaximumLength: 220,
				NumericPrecision:       -1,
				NumericScale:           -1,
				OrdinalPosition:        4,
			},
			{
				SchemaName:             "public",
				TableName:              "users",
				ColumnName:             "age",
				DataType:               "integer",
				ColumnDefault:          "",
				IsNullable:             "YES",
				CharacterMaximumLength: -1,
				NumericPrecision:       32,
				NumericScale:           0,
				OrdinalPosition:        5,
			},
		}, nil,
	)

	pgquerier.On("GetTableConstraints", mock.Anything, mockPool, &pg_queries.GetTableConstraintsParams{
		Schema: "public",
		Table:  "users",
	}).Return(
		[]*pg_queries.GetTableConstraintsRow{
			{
				ConstraintName:       "users_pkey",
				ConstraintDefinition: "PRIMARY KEY (id)",
			},
		},
		nil,
	)

	actual, err := manager.GetCreateTableStatement(context.Background(), "public", "users")
	require.NoError(t, err)
	require.Equal(t, "CREATE TABLE IF NOT EXISTS \"public\".\"users\" (\"id\" varchar NOT NULL, \"age\" integer NULL, CONSTRAINT users_pkey PRIMARY KEY (id));", actual)
}

func Test_GenerateCreateTableStatement(t *testing.T) {
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
			expected: `CREATE TABLE IF NOT EXISTS "public"."users" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "created_at" timestamp without time zone NOT NULL DEFAULT now(), "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "extra" varchar NULL, "name" varchar(40) NULL, CONSTRAINT users_pkey PRIMARY KEY (id));`,
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
			expected:    `CREATE TABLE IF NOT EXISTS "public"."users" ("id" SERIAL NOT NULL, "id2" SMALLSERIAL NOT NULL, "id3" BIGSERIAL NOT NULL);`,
		},
	}

	for _, testcase := range cases {
		t.Run(t.Name(), func(t *testing.T) {
			actual := generateCreateTableStatement(testcase.schema, testcase.table, testcase.rows, testcase.constraints)
			require.Equal(t, testcase.expected, actual)
		})
	}
}

func Test_BatchExec(t *testing.T) {
	prefix := DisableForeignKeyChecks
	tests := []struct {
		name          string
		batchSize     int
		statements    []string
		opts          *BatchExecOpts
		expectedCalls []string
	}{
		{
			name:          "multiple batches",
			batchSize:     2,
			statements:    []string{"CREATE TABLE users;", "CREATE TABLE accounts;", "CREATE TABLE departments;"},
			expectedCalls: []string{"CREATE TABLE users;\nCREATE TABLE accounts;", "CREATE TABLE departments;"},
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
			expectedCalls: []string{fmt.Sprintf("%s %s", prefix, "CREATE TABLE users;\nCREATE TABLE accounts;"), fmt.Sprintf("%s %s", prefix, "CREATE TABLE departments;")},
			opts:          &BatchExecOpts{Prefix: &prefix},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgquerier := pg_queries.NewMockQuerier(t)
			mockPool := pg_queries.NewMockDBTX(t)
			manager := PostgresManager{
				querier: pgquerier,
				pool:    mockPool,
			}

			for _, call := range tt.expectedCalls {
				var cmdtag pgconn.CommandTag
				mockPool.On("Exec", mock.Anything, call).Return(cmdtag, nil)
			}

			err := manager.BatchExec(context.Background(), tt.batchSize, tt.statements, tt.opts)
			require.NoError(t, err)
		})
	}
}

func Test_Exec(t *testing.T) {

	pgquerier := pg_queries.NewMockQuerier(t)
	mockPool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockPool,
	}

	stmt := "TRUNCATE TABLE users;"
	var cmdtag pgconn.CommandTag
	mockPool.On("Exec", mock.Anything, stmt).Return(cmdtag, nil)

	err := manager.Exec(context.Background(), stmt)
	require.NoError(t, err)
}

func Test_EscapePgColumns(t *testing.T) {
	require.Empty(t, EscapePgColumns(nil))
	require.Equal(
		t,
		EscapePgColumns([]string{"foo", "bar", "baz"}),
		[]string{`"foo"`, `"bar"`, `"baz"`},
	)
}

func Test_BuildPgTruncateStatement(t *testing.T) {
	require.Equal(
		t,
		BuildPgTruncateStatement([]string{"foo", "bar", "baz"}),
		"TRUNCATE TABLE foo, bar, baz;",
	)
}

func Test_BuildPgTruncateCascadeStatement(t *testing.T) {
	actual, err := BuildPgTruncateCascadeStatement("public", "users")
	require.NoError(t, err)
	require.Equal(
		t,
		"TRUNCATE \"public\".\"users\" CASCADE;",
		actual,
	)
}

func mockTableConstraintsRows() []*pg_queries.GetTableConstraintsBySchemaRow {
	return []*pg_queries.GetTableConstraintsBySchemaRow{
		{
			ConstraintName:     "fk_users",
			SchemaName:         "public",
			TableName:          "orders",
			ConstraintColumns:  []string{"buyer_id"},
			ForeignSchemaName:  "public",
			ForeignTableName:   "users",
			ForeignColumnNames: []string{"id"},
			ConstraintType:     "f",
			Notnullable:        []bool{true},
		},
		{
			ConstraintName:     "fk_users_composite",
			SchemaName:         "public",
			TableName:          "users",
			ConstraintColumns:  []string{"composite_id", "other_composite_id"},
			ForeignSchemaName:  "public",
			ForeignTableName:   "composite",
			ForeignColumnNames: []string{"id", "other_id"},
			ConstraintType:     "f",
			Notnullable:        []bool{true, true},
		},
		{
			ConstraintName:     "fk_account",
			SchemaName:         "public",
			TableName:          "orders",
			ConstraintColumns:  []string{"account_id"},
			ForeignSchemaName:  "public",
			ForeignTableName:   "accounts",
			ForeignColumnNames: []string{"id"},
			ConstraintType:     "f",
			Notnullable:        []bool{false},
		},
		{
			ConstraintName:     "fk_users",
			SchemaName:         "public",
			TableName:          "accounts",
			ConstraintColumns:  []string{"user_id"},
			ForeignSchemaName:  "public",
			ForeignTableName:   "users",
			ForeignColumnNames: []string{"id"},
			ConstraintType:     "f",
			Notnullable:        []bool{true},
		},
		{
			SchemaName:        "public",
			TableName:         "users",
			ConstraintName:    "users",
			ConstraintColumns: []string{"id"},
			ConstraintType:    "p",
		},
		{
			SchemaName:        "public",
			TableName:         "orders",
			ConstraintName:    "orders",
			ConstraintColumns: []string{"id"},
			ConstraintType:    "p",
		},
		{
			SchemaName:        "public",
			TableName:         "accounts",
			ConstraintName:    "accoutns",
			ConstraintColumns: []string{"id"},
			ConstraintType:    "p",
		},
		{
			SchemaName:        "public",
			TableName:         "composite",
			ConstraintName:    "composite",
			ConstraintColumns: []string{"id", "other_id"},
			ConstraintType:    "p",
		},
		{
			SchemaName:        "public",
			TableName:         "person",
			ConstraintName:    "person",
			ConstraintColumns: []string{"email", "name"},
			ConstraintType:    "u",
		},
		{
			SchemaName:        "public",
			TableName:         "region",
			ConstraintName:    "region",
			ConstraintColumns: []string{"code"},
			ConstraintType:    "u",
		},
		{
			SchemaName:        "public",
			TableName:         "region",
			ConstraintName:    "region",
			ConstraintColumns: []string{"name"},
			ConstraintType:    "u",
		},
	}
}
