package genbenthosconfigs_activity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildSelectQuery(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		schema   string
		table    string
		columns  []string
		where    string
		expected string
	}{
		{
			name:     "postgres select",
			driver:   "postgres",
			schema:   "public",
			table:    "accounts",
			columns:  []string{"id", "name"},
			where:    "",
			expected: `SELECT "id", "name" FROM "public"."accounts";`,
		},
		{
			name:     "postgres select with where",
			driver:   "postgres",
			schema:   "public",
			table:    "accounts",
			columns:  []string{"id", "name"},
			where:    `"id" = 'some-id'`,
			expected: `SELECT "id", "name" FROM "public"."accounts" WHERE "id" = 'some-id';`,
		},
		{
			name:     "postgres select with where prepared",
			driver:   "postgres",
			schema:   "public",
			table:    "accounts",
			columns:  []string{"id", "name"},
			where:    `"id" = $1`,
			expected: `SELECT "id", "name" FROM "public"."accounts" WHERE "id" = $1;`,
		},
		{
			name:     "mysql select",
			driver:   "mysql",
			schema:   "public",
			table:    "accounts",
			columns:  []string{"id", "name"},
			where:    "",
			expected: "SELECT `id`, `name` FROM `public`.`accounts`;",
		},
		{
			name:     "mysql select with where",
			driver:   "mysql",
			schema:   "public",
			table:    "accounts",
			columns:  []string{"id", "name"},
			where:    "`id` = 'some-id'",
			expected: "SELECT `id`, `name` FROM `public`.`accounts` WHERE `id` = 'some-id';",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			where := tt.where
			sql, err := buildSelectQuery(tt.driver, tt.schema, tt.table, tt.columns, &where)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, sql)
		})
	}
}
