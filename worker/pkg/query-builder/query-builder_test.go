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
		table          string
		insertColumns  []string
		whereColumns   []string
		columnValueMap map[string]any
		expected       string
	}{
		{"Single Column postgres", "postgres", "public.users", []string{"name"}, []string{"id"}, map[string]any{"name": "Alice", "id": 1}, `UPDATE "public"."users" SET "name"='Alice' WHERE ("id" = 1)`},
		{"Multiple Primary Keys postgres", "postgres", "public.users", []string{"name", "email"}, []string{"id", "other"}, map[string]any{"name": "Alice", "id": 1, "email": "alice@fake.com", "other": "blah"}, `UPDATE "public"."users" SET "email"='alice@fake.com',"name"='Alice' WHERE (("id" = 1) AND ("other" = 'blah'))`},
		{"Single Column mysql", "mysql", "public.users", []string{"name"}, []string{"id"}, map[string]any{"name": "Alice", "id": 1}, "UPDATE `public`.`users` SET `name`='Alice' WHERE (`id` = 1)"},
		{"Multiple Primary Keys mysql", "mysql", "public.users", []string{"name", "email"}, []string{"id", "other"}, map[string]any{"name": "Alice", "id": 1, "email": "alice@fake.com", "other": "blah"}, "UPDATE `public`.`users` SET `email`='alice@fake.com',`name`='Alice' WHERE ((`id` = 1) AND (`other` = 'blah'))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := BuildUpdateQuery(tt.driver, tt.table, tt.insertColumns, tt.whereColumns, tt.columnValueMap)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_BuildInsertQuery(t *testing.T) {
	tests := []struct {
		name                string
		driver              string
		table               string
		columns             []string
		values              [][]any
		onConflictDoNothing bool
		expected            string
	}{
		{"Single Column mysql", "mysql", "public.users", []string{"name"}, [][]any{{"Alice"}, {"Bob"}}, false, "INSERT INTO `public`.`users` (`name`) VALUES ('Alice'), ('Bob')"},
		{"Multiple Columns mysql", "mysql", "public.users", []string{"name", "email"}, [][]any{{"Alice", "alice@fake.com"}, {"Bob", "bob@fake.com"}}, true, "INSERT IGNORE INTO `public`.`users` (`name`, `email`) VALUES ('Alice', 'alice@fake.com'), ('Bob', 'bob@fake.com')"},
		{"Single Column postgres", "postgres", "public.users", []string{"name"}, [][]any{{"Alice"}, {"Bob"}}, false, `INSERT INTO "public"."users" ("name") VALUES ('Alice'), ('Bob')`},
		{"Multiple Columns postgres", "postgres", "public.users", []string{"name", "email"}, [][]any{{"Alice", "alice@fake.com"}, {"Bob", "bob@fake.com"}}, true, `INSERT INTO "public"."users" ("name", "email") VALUES ('Alice', 'alice@fake.com'), ('Bob', 'bob@fake.com') ON CONFLICT DO NOTHING`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := BuildInsertQuery(tt.driver, tt.table, tt.columns, tt.values, &tt.onConflictDoNothing)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
