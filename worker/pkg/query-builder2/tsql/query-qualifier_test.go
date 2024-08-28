package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_QualifyWhereCondition(t *testing.T) {
	tests := []struct {
		name     string
		inputSQL string
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple WHERE clause with single table",
			inputSQL: "SELECT * FROM users WHERE name = 'John'",
			expected: `SELECT * FROM users WHERE "users"."name" = 'John'`,
			wantErr:  false,
		},
		{
			name:     "Simple WHERE table already qualified",
			inputSQL: `SELECT * FROM "users" WHERE name = 'John'`,
			expected: `SELECT * FROM "users" WHERE "users"."name" = 'John'`,
			wantErr:  false,
		},
		{
			name:     "JOIN with qualified columns",
			inputSQL: "SELECT u.id, o.order_id FROM users u JOIN orders o ON u.id = o.user_id WHERE o.amount > 100",
			expected: `SELECT u.id, o.order_id FROM users u JOIN orders o ON u.id = o.user_id WHERE o.amount > 100`,
			wantErr:  false,
		},
		{
			name:     "Multiple conditions in WHERE clause",
			inputSQL: "SELECT id FROM users WHERE name = 'John' AND age > 30",
			expected: `SELECT id FROM users WHERE "users"."name" = 'John' AND "users"."age" > 30`,
			wantErr:  false,
		},
		{
			name:     "Subquery in WHERE clause",
			inputSQL: "SELECT id FROM users WHERE id IN (SELECT user_id FROM orders WHERE amount > 100)",
			expected: `SELECT id FROM users WHERE "users"."id" IN ( SELECT user_id FROM orders WHERE "orders"."amount" > 100 )`,
			wantErr:  false,
		},
		{
			name:     "No WHERE clause",
			inputSQL: "SELECT id FROM users",
			expected: `SELECT id FROM users`,
			wantErr:  false,
		},
		{
			name:     "Complex WHERE clause with OR",
			inputSQL: "SELECT id FROM users WHERE name = 'John' OR age > 30",
			expected: `SELECT id FROM users WHERE "users"."name" = 'John' OR "users"."age" > 30`,
			wantErr:  false,
		},
		{
			name:     "Invalid SQL input",
			inputSQL: "SELECT id FROM WHERE name = 'John'",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := QualifyWhereCondition(tt.inputSQL)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}
