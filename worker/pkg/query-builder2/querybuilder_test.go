package querybuilder2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_qualifyWhereColumnNames_mysql(t *testing.T) {
	tests := []struct {
		name     string
		where    string
		schema   string
		table    string
		expected string
	}{
		{
			name:     "simple",
			where:    "name = 'alisha'",
			schema:   "public",
			table:    "a",
			expected: "public.a.name = 'alisha'",
		},
		{
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			schema:   "public",
			table:    "a",
			expected: "public.a.name = 'alisha'",
		},
		{
			name:     "multiple",
			where:    "name = 'alisha' and id = 1",
			schema:   "public",
			table:    "a",
			expected: "public.a.name = 'alisha' and public.a.id = 1",
		},
		{
			name:     "where subquery",
			where:    "film_id IN(SELECT film_id FROM film_category INNER JOIN category USING(category_id) WHERE name='Action')",
			schema:   "public",
			table:    "film",
			expected: "public.film.film_id in (select film_id from film_category join category using (category_id) where name = 'Action')",
		},
		{
			name:     "quoted column names",
			where:    "`name with space` = 'hey' and PascalCase = 'other'",
			schema:   "public",
			table:    "order",
			expected: "public.`order`.`name with space` = 'hey' and public.`order`.PascalCase = 'other'",
		},
	}

	qb := &QueryBuilder{driver: "mysql"}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qb.qualifyWhereCondition(&tt.schema, tt.table, tt.where)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func Test_qualifyWhereColumnNames_postgres(t *testing.T) {
	tests := []struct {
		name     string
		where    string
		schema   string
		table    string
		expected string
	}{
		{
			name:     "simple",
			where:    "name = 'alisha'",
			schema:   "public",
			table:    "a",
			expected: `public.a.name = 'alisha'`,
		},
		{
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			schema:   "public",
			table:    "a",
			expected: `public.a.name = 'alisha'`,
		},
		{
			name:     "simple",
			where:    `"bad name" = 'alisha'`,
			schema:   "public",
			table:    "order",
			expected: `public."order"."bad name" = 'alisha'`,
		},
		{
			name:     "multiple",
			where:    "name = 'alisha' and id = 1  or age = 2",
			schema:   "public",
			table:    "a",
			expected: `(public.a.name = 'alisha' AND public.a.id = 1) OR public.a.age = 2`,
		},
		{
			name:     "where subquery",
			where:    "film_id IN(SELECT film_id FROM film_category INNER JOIN category USING(category_id) WHERE name='Action');",
			schema:   "public",
			table:    "film",
			expected: `public.film.film_id IN (SELECT film_id FROM film_category JOIN category USING (category_id) WHERE name = 'Action')`,
		},
		{
			name:     "where subquery",
			where:    "DATE_PART('year', event_date) = 2023 AND attendees > 100;",
			schema:   "public",
			table:    "film",
			expected: "date_part('year', public.film.event_date) = 2023 AND public.film.attendees > 100",
		},
	}

	qb := &QueryBuilder{driver: "postgres"}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qb.qualifyWhereCondition(&tt.schema, tt.table, tt.where)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}
