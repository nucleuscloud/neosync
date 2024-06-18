package sqlmanager_postgres

import (
	context "context"
	"testing"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_generateCreateTableStatement(t *testing.T) {
	type testcase struct {
		schema      string
		table       string
		rows        []*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow
		constraints []*pg_queries.GetTableConstraintsRow
		expected    string
	}
	cases := []testcase{
		{
			schema: "public",
			table:  "users",
			rows: []*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{
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
			rows: []*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{
				{
					ColumnName:      "id",
					DataType:        "integer",
					OrdinalPosition: 1,
					IsNullable:      "NO",
					ColumnDefault:   "nextval('users_id_seq'::regclass)",
					SequenceType:    "SERIAL",
				},
				{
					ColumnName:      "id2",
					DataType:        "smallint",
					OrdinalPosition: 2,
					IsNullable:      "NO",
					ColumnDefault:   "nextval('users_id2_seq'::regclass)",
					SequenceType:    "SERIAL",
				},
				{
					ColumnName:      "id3",
					DataType:        "bigint",
					OrdinalPosition: 3,
					IsNullable:      "NO",
					ColumnDefault:   "nextval('users_id3_seq'::regclass)",
					SequenceType:    "SERIAL",
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

func Test_GetSchemaInitStatements(t *testing.T) {
	pgquerier := pg_queries.NewMockQuerier(t)
	mockpool := pg_queries.NewMockDBTX(t)
	manager := PostgresManager{
		querier: pgquerier,
		pool:    mockpool,
	}
	pgquerier.On("GetCustomSequencesBySchemaAndTables", mock.Anything, mock.Anything, &pg_queries.GetCustomSequencesBySchemaAndTablesParams{Schema: "public", Tables: []string{"users"}}).
		Return([]*pg_queries.GetCustomSequencesBySchemaAndTablesRow{}, nil)
	pgquerier.On("GetCustomFunctionsBySchemaAndTables", mock.Anything, mock.Anything, &pg_queries.GetCustomFunctionsBySchemaAndTablesParams{Schema: "public", Tables: []string{"users"}}).
		Return([]*pg_queries.GetCustomFunctionsBySchemaAndTablesRow{}, nil)
	pgquerier.On("GetDataTypesBySchemaAndTables", mock.Anything, mock.Anything, &pg_queries.GetDataTypesBySchemaAndTablesParams{Schema: "public", Tables: []string{"users"}}).
		Return([]*pg_queries.GetDataTypesBySchemaAndTablesRow{}, nil)

	pgquerier.On("GetCustomTriggersBySchemaAndTables", mock.Anything, mock.Anything, []string{"public.users"}).
		Return([]*pg_queries.GetCustomTriggersBySchemaAndTablesRow{{SchemaName: "public", TableName: "users", TriggerName: "foo_trigger", Definition: "test-trigger-statement"}}, nil)

	pgquerier.On("GetDatabaseTableSchemasBySchemasAndTables", mock.Anything, mockpool, []string{"public.users"}).
		Return(
			[]*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{
				{
					SchemaName:    "public",
					TableName:     "users",
					ColumnName:    "id",
					DataType:      "uuid",
					ColumnDefault: "",
					IsNullable:    "NO",
				},
			},
			nil,
		)

	pgquerier.On("GetTableConstraintsBySchema", mock.Anything, mockpool, []string{"public"}).
		Return(
			[]*pg_queries.GetTableConstraintsBySchemaRow{
				{
					ConstraintName:       "pk_public_users",
					ConstraintType:       "p",
					SchemaName:           "public",
					TableName:            "users",
					ConstraintColumns:    []string{"id"},
					Notnullable:          []bool{true},
					ConstraintDefinition: "PRIMARY KEY(id)",
				},
			},
			nil,
		)

	pgquerier.On("GetIndicesBySchemasAndTables", mock.Anything, mockpool, []string{"public.users"}).
		Return(
			[]*pg_queries.GetIndicesBySchemasAndTablesRow{
				{
					SchemaName:      "public",
					TableName:       "users",
					IndexName:       "foo",
					IndexDefinition: "CREATE INDEX foo ON public.users USING btree (users_id)",
				},
			},
			nil,
		)

	expected := []*sqlmanager_shared.InitSchemaStatements{
		{Label: "data types", Statements: []string{}},
		{Label: "create table", Statements: []string{"CREATE TABLE IF NOT EXISTS \"public\".\"users\" (\"id\" uuid NOT NULL);"}},
		{Label: "non-fk alter table", Statements: []string{"DO $$\nBEGIN\n\tIF NOT EXISTS (\n\t\tSELECT 1\n\t\tFROM pg_constraint\n\t\tWHERE conname = 'pk_public_users'\n\t\tAND connamespace = 'public'::regnamespace\n\t\tAND conrelid = (\n\t\t\tSELECT oid\n\t\t\tFROM pg_class\n\t\t\tWHERE relname = 'users'\n\t\t\tAND relnamespace = 'public'::regnamespace\n\t\t)\n\t) THEN\n\t\tALTER TABLE \"public\".\"users\" ADD CONSTRAINT pk_public_users PRIMARY KEY(id);\n\tEND IF;\nEND $$;"}},
		{Label: "fk alter table", Statements: []string{}},
		{Label: "table index", Statements: []string{"DO $$\nBEGIN\n\tIF NOT EXISTS (\n\t\tSELECT 1\n\t\tFROM pg_class c\n\t\tJOIN pg_namespace n ON n.oid = c.relnamespace\n\t\tWHERE c.relkind = 'i'\n\t\tAND c.relname = 'foo'\n\t\tAND n.nspname = 'public'\n\t) THEN\n\t\tCREATE INDEX foo ON public.users USING btree (users_id);\n\tEND IF;\nEND $$;"}},
		{Label: "table triggers", Statements: []string{"DO $$\nBEGIN\n    IF NOT EXISTS (\n        SELECT 1\n        FROM pg_trigger t\n        JOIN pg_class c ON c.oid = t.tgrelid\n        JOIN pg_namespace n ON n.oid = c.relnamespace\n        WHERE t.tgname = 'foo_trigger'\n        AND c.relname = 'users'\n        AND n.nspname = 'public'\n    ) THEN\n        test-trigger-statement;\n    END IF;\nEND $$;"}},
	}

	actual, err := manager.GetSchemaInitStatements(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: "public", Table: "users"}})
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
