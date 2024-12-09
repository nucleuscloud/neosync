package querybuilder

import (
	"fmt"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
	tests := []struct {
		name                string
		driver              string
		schema              string
		table               string
		records             []map[string]any
		onConflictDoNothing bool
		expected            string
		expectedArgs        []any
	}{
		{
			name:                "Single Column mysql",
			driver:              "mysql",
			schema:              "public",
			table:               "users",
			records:             []map[string]any{{"name": "Alice"}, {"name": "Bob"}},
			onConflictDoNothing: false,
			expected:            "INSERT INTO `public`.`users` (`name`) VALUES (?), (?)",
			expectedArgs:        []any{"Alice", "Bob"},
		},
		{
			name:                "Special characters mysql",
			driver:              "mysql",
			schema:              "public",
			table:               "users.stage$dev",
			records:             []map[string]any{{"name": "Alice"}, {"name": "Bob"}},
			onConflictDoNothing: false,
			expected:            "INSERT INTO `public`.`users.stage$dev` (`name`) VALUES (?), (?)",
			expectedArgs:        []any{"Alice", "Bob"},
		},
		{
			name:                "Multiple Columns mysql",
			driver:              "mysql",
			schema:              "public",
			table:               "users",
			records:             []map[string]any{{"name": "Alice", "email": "alice@fake.com"}, {"name": "Bob", "email": "bob@fake.com"}},
			onConflictDoNothing: true,
			expected:            "INSERT IGNORE INTO `public`.`users` (`email`, `name`) VALUES (?, ?), (?, ?)",
			expectedArgs:        []any{"alice@fake.com", "Alice", "bob@fake.com", "Bob"},
		},
		{
			name:                "Single Column postgres",
			driver:              "postgres",
			schema:              "public",
			table:               "users",
			records:             []map[string]any{{"name": "Alice"}, {"name": "Bob"}},
			onConflictDoNothing: false,
			expected:            `INSERT INTO "public"."users" ("name") VALUES ($1), ($2)`,
			expectedArgs:        []any{"Alice", "Bob"},
		},
		{
			name:                "Multiple Columns postgres",
			driver:              "postgres",
			schema:              "public",
			table:               "users",
			records:             []map[string]any{{"name": "Alice", "email": "alice@fake.com"}, {"name": "Bob", "email": "bob@fake.com"}},
			onConflictDoNothing: true,
			expected:            `INSERT INTO "public"."users" ("email", "name") VALUES ($1, $2), ($3, $4) ON CONFLICT DO NOTHING`,
			expectedArgs:        []any{"alice@fake.com", "Alice", "bob@fake.com", "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goquRows := toGoquRecords(tt.records)
			actual, args, err := BuildInsertQuery(tt.driver, tt.schema, tt.table, goquRows, &tt.onConflictDoNothing)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
			require.Equal(t, tt.expectedArgs, args)
		})
	}
}
