package querybuilder

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/stretchr/testify/require"
)

func Test_BuildSelectQuery(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		table    string
		columns  []string
		where    string
		expected string
	}{
		{
			name:     "postgres select",
			driver:   sqlmanager_shared.PostgresDriver,
			table:    "public.accounts",
			columns:  []string{"id", "name"},
			where:    "",
			expected: `SELECT "id", "name" FROM "public"."accounts";`,
		},
		{
			name:     "postgres select with where",
			driver:   sqlmanager_shared.PostgresDriver,
			table:    "public.accounts",
			columns:  []string{"id", "name"},
			where:    `"id" = 'some-id'`,
			expected: `SELECT "id", "name" FROM "public"."accounts" WHERE "id" = 'some-id';`,
		},
		{
			name:     "postgres select with where prepared",
			driver:   sqlmanager_shared.PostgresDriver,
			table:    "public.accounts",
			columns:  []string{"id", "name"},
			where:    `"id" = $1`,
			expected: `SELECT "id", "name" FROM "public"."accounts" WHERE "id" = $1;`,
		},
		{
			name:     "mysql select",
			driver:   sqlmanager_shared.MysqlDriver,
			table:    "public.accounts",
			columns:  []string{"id", "name"},
			where:    "",
			expected: "SELECT `id`, `name` FROM `public`.`accounts`;",
		},
		{
			name:     "mysql select with where",
			driver:   sqlmanager_shared.MysqlDriver,
			table:    "public.accounts",
			columns:  []string{"id", "name"},
			where:    "`id` = 'some-id'",
			expected: "SELECT `id`, `name` FROM `public`.`accounts` WHERE `id` = 'some-id';",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			where := tt.where
			sql, err := BuildSelectQuery(tt.driver, tt.table, tt.columns, &where)
			require.NoError(t, err)
			require.Equal(t, tt.expected, sql)
		})
	}
}

func Test_BuildUpdateQuery(t *testing.T) {
	tests := []struct {
		name           string
		driver         string
		schema         string
		table          string
		insertColumns  []string
		whereColumns   []string
		columnValueMap map[string]any
		expected       string
	}{
		{"Single Column postgres", "postgres", "public", "users", []string{"name"}, []string{"id"}, map[string]any{"name": "Alice", "id": 1}, `UPDATE "public"."users" SET "name"='Alice' WHERE ("id" = 1)`},
		{"Special characters postgres", "postgres", "public", "users.stage$dev", []string{"name"}, []string{"id"}, map[string]any{"name": "Alice", "id": 1}, `UPDATE "public"."users.stage$dev" SET "name"='Alice' WHERE ("id" = 1)`},
		{"Multiple Primary Keys postgres", "postgres", "public", "users", []string{"name", "email"}, []string{"id", "other"}, map[string]any{"name": "Alice", "id": 1, "email": "alice@fake.com", "other": "blah"}, `UPDATE "public"."users" SET "email"='alice@fake.com',"name"='Alice' WHERE (("id" = 1) AND ("other" = 'blah'))`},
		{"Single Column mysql", "mysql", "public", "users", []string{"name"}, []string{"id"}, map[string]any{"name": "Alice", "id": 1}, "UPDATE `public`.`users` SET `name`='Alice' WHERE (`id` = 1)"},
		{"Multiple Primary Keys mysql", "mysql", "public", "users", []string{"name", "email"}, []string{"id", "other"}, map[string]any{"name": "Alice", "id": 1, "email": "alice@fake.com", "other": "blah"}, "UPDATE `public`.`users` SET `email`='alice@fake.com',`name`='Alice' WHERE ((`id` = 1) AND (`other` = 'blah'))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := BuildUpdateQuery(tt.driver, tt.schema, tt.table, tt.insertColumns, tt.whereColumns, tt.columnValueMap)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_BuildInsertQuery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tests := []struct {
		name                    string
		driver                  string
		schema                  string
		table                   string
		columns                 []string
		columnDataTypes         []string
		values                  [][]any
		onConflictDoNothing     bool
		columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties
		expected                string
		expectedArgs            []any
	}{
		{"Single Column mysql", "mysql", "public", "users", []string{"name"}, []string{}, [][]any{{"Alice"}, {"Bob"}}, false, []*neosync_benthos.ColumnDefaultProperties{}, "INSERT INTO `public`.`users` (`name`) VALUES (?), (?)", []any{"Alice", "Bob"}},
		{"Special characters mysql", "mysql", "public", "users.stage$dev", []string{"name"}, []string{}, [][]any{{"Alice"}, {"Bob"}}, false, []*neosync_benthos.ColumnDefaultProperties{}, "INSERT INTO `public`.`users.stage$dev` (`name`) VALUES (?), (?)", []any{"Alice", "Bob"}},
		{"Multiple Columns mysql", "mysql", "public", "users", []string{"name", "email"}, []string{}, [][]any{{"Alice", "alice@fake.com"}, {"Bob", "bob@fake.com"}}, true, []*neosync_benthos.ColumnDefaultProperties{}, "INSERT IGNORE INTO `public`.`users` (`name`, `email`) VALUES (?, ?), (?, ?)", []any{"Alice", "alice@fake.com", "Bob", "bob@fake.com"}},
		{"Single Column postgres", "postgres", "public", "users", []string{"name"}, []string{}, [][]any{{"Alice"}, {"Bob"}}, false, []*neosync_benthos.ColumnDefaultProperties{}, `INSERT INTO "public"."users" ("name") VALUES ($1), ($2)`, []any{"Alice", "Bob"}},
		{"Multiple Columns postgres", "postgres", "public", "users", []string{"name", "email"}, []string{}, [][]any{{"Alice", "alice@fake.com"}, {"Bob", "bob@fake.com"}}, true, []*neosync_benthos.ColumnDefaultProperties{}, `INSERT INTO "public"."users" ("name", "email") VALUES ($1, $2), ($3, $4) ON CONFLICT DO NOTHING`, []any{"Alice", "alice@fake.com", "Bob", "bob@fake.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, args, err := BuildInsertQuery(logger, tt.driver, tt.schema, tt.table, tt.columns, tt.columnDataTypes, tt.values, &tt.onConflictDoNothing, tt.columnDefaultProperties)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
			require.Equal(t, tt.expectedArgs, args)
		})
	}
}

func Test_BuildInsertQuery_JsonArray(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	driver := sqlmanager_shared.PostgresDriver
	schema := "public"
	table := "test_table"
	columns := []string{"id", "name", "tags"}
	columnDataTypes := []string{"int", "text", "jsonb[]"}
	columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{nil, nil, nil}
	values := [][]any{
		{1, "John", []map[string]any{{"tag": "cool"}, {"tag": "awesome"}}},
		{2, "Jane", []map[string]any{{"tag": "smart"}, {"tag": "clever"}}},
	}
	onConflictDoNothing := false

	query, _, err := BuildInsertQuery(logger, driver, schema, table, columns, columnDataTypes, values, &onConflictDoNothing, columnDefaultProperties)
	require.NoError(t, err)
	expectedQuery := `INSERT INTO "public"."test_table" ("id", "name", "tags") VALUES ($1, $2, ARRAY['{"tag":"cool"}','{"tag":"awesome"}']::jsonb[]), ($3, $4, ARRAY['{"tag":"smart"}','{"tag":"clever"}']::jsonb[])`
	require.Equal(t, expectedQuery, query)
}

func Test_BuildInsertQuery_Json(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	driver := sqlmanager_shared.PostgresDriver
	schema := "public"
	table := "test_table"
	columns := []string{"id", "name", "tags"}
	columnDataTypes := []string{"int", "text", "json"}
	columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{}
	values := [][]any{
		{1, "John", map[string]any{"tag": "cool"}},
		{2, "Jane", map[string]any{"tag": "smart"}},
	}
	onConflictDoNothing := false

	query, args, err := BuildInsertQuery(logger, driver, schema, table, columns, columnDataTypes, values, &onConflictDoNothing, columnDefaultProperties)
	require.NoError(t, err)
	expectedQuery := `INSERT INTO "public"."test_table" ("id", "name", "tags") VALUES ($1, $2, $3), ($4, $5, $6)`
	require.Equal(t, expectedQuery, query)
	require.Equal(t, []any{int64(1), "John", []byte{123, 34, 116, 97, 103, 34, 58, 34, 99, 111, 111, 108, 34, 125}, int64(2), "Jane", []byte{123, 34, 116, 97, 103, 34, 58, 34, 115, 109, 97, 114, 116, 34, 125}}, args)
}

func TestGetGoquVals(t *testing.T) {
	t.Run("Postgres", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		driver := sqlmanager_shared.PostgresDriver
		row := []any{"value1", 42, true, map[string]any{"key": "value"}, []int{1, 2, 3}}
		columnDataTypes := []string{"text", "integer", "boolean", "jsonb", "integer[]"}
		columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{nil, nil, nil, nil, nil}

		result := getGoquVals(logger, driver, row, columnDataTypes, columnDefaultProperties)

		require.Len(t, result, 5)
		require.Equal(t, "value1", result[0])
		require.Equal(t, 42, result[1])
		require.Equal(t, true, result[2])
		require.JSONEq(t, `{"key":"value"}`, string(result[3].([]byte)))
		require.Equal(t, pq.Array([]any{1, 2, 3}), result[4])
	})

	t.Run("Postgres JSON", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		driver := sqlmanager_shared.PostgresDriver
		row := []any{"value1", 42, true, map[string]any{"key": "value"}, []int{1, 2, 3}}
		columnDataTypes := []string{"jsonb", "jsonb", "jsonb", "jsonb", "json"}
		columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{nil, nil, nil, nil, nil}

		result := getGoquVals(logger, driver, row, columnDataTypes, columnDefaultProperties)

		require.Equal(t, goqu.Vals{
			[]byte(`"value1"`),
			[]byte(`42`),
			[]byte(`true`),
			[]byte(`{"key":"value"}`),
			[]byte(`[1,2,3]`),
		}, result)
	})

	t.Run("Postgres Empty Column DataTypes", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		driver := sqlmanager_shared.MysqlDriver
		row := []any{"value1", 42, true, "DEFAULT"}
		columnDataTypes := []string{}
		columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{nil, nil, nil, {HasDefaultTransformer: true}}

		result := getGoquVals(logger, driver, row, columnDataTypes, columnDefaultProperties)

		require.Len(t, result, 4)
		require.Equal(t, "value1", result[0])
		require.Equal(t, 42, result[1])
		require.Equal(t, true, result[2])
		require.Equal(t, goqu.L("DEFAULT"), result[3])
	})

	t.Run("Mysql", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		driver := sqlmanager_shared.MysqlDriver
		row := []any{"value1", 42, true, "DEFAULT"}
		columnDataTypes := []string{}
		columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{nil, nil, nil, {HasDefaultTransformer: true}}

		result := getGoquVals(logger, driver, row, columnDataTypes, columnDefaultProperties)

		require.Len(t, result, 4)
		require.Equal(t, "value1", result[0])
		require.Equal(t, 42, result[1])
		require.Equal(t, true, result[2])
		require.Equal(t, goqu.L("DEFAULT"), result[3])
	})

	t.Run("EmptyRow", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		driver := sqlmanager_shared.PostgresDriver
		row := []any{}
		columnDataTypes := []string{}
		columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{}

		result := getGoquVals(logger, driver, row, columnDataTypes, columnDefaultProperties)

		require.Empty(t, result)
	})

	t.Run("Mismatch length ColumnDataTypes and Row Values", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		driver := sqlmanager_shared.PostgresDriver
		row := []any{"text", 42, true}
		columnDataTypes := []string{"text"}
		columnDefaultProperties := []*neosync_benthos.ColumnDefaultProperties{}

		result := getGoquVals(logger, driver, row, columnDataTypes, columnDefaultProperties)

		require.Len(t, result, 3)
		require.Equal(t, "text", result[0])
		require.Equal(t, 42, result[1])
		require.Equal(t, true, result[2])
	})
}
