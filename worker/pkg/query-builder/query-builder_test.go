package querybuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/require"
)

func Test_buildSelectQuery(t *testing.T) {
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

func Test_buildSelectJoinQuery(t *testing.T) {
	tests := []struct {
		name         string
		driver       string
		table        string
		columns      []string
		joins        []*sqlJoin
		whereClauses []string
		expected     string
	}{
		{
			name:    "simple",
			driver:  sqlmanager_shared.PostgresDriver,
			table:   "public.a",
			columns: []string{"id", "name", "email"},
			joins: []*sqlJoin{
				{
					JoinType:  innerJoin,
					JoinTable: "public.b",
					BaseTable: "public.a",
					JoinColumnsMap: map[string]string{
						"a_id": "id",
					},
				},
			},
			whereClauses: []string{`"public"."a"."name" = 'alisha'`, `"public"."b"."email" = 'fake@email.com'`},
			expected:     `SELECT "public"."a"."id", "public"."a"."name", "public"."a"."email" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."a_id" = "public"."a"."id") WHERE ("public"."a"."name" = 'alisha' AND "public"."b"."email" = 'fake@email.com');`,
		},
		{
			name:    "multiple joins",
			driver:  sqlmanager_shared.PostgresDriver,
			table:   "public.a",
			columns: []string{"id", "name", "email"},
			joins: []*sqlJoin{
				{
					JoinType:  innerJoin,
					JoinTable: "public.b",
					BaseTable: "public.a",
					JoinColumnsMap: map[string]string{
						"a_id": "id",
					},
				},
				{
					JoinType:  innerJoin,
					JoinTable: "public.c",
					BaseTable: "public.b",
					JoinColumnsMap: map[string]string{
						"b_id": "id",
					},
				},
			},
			whereClauses: []string{`"public"."a"."name" = 'alisha'`, `"public"."b"."id" = 1`},
			expected:     `SELECT "public"."a"."id", "public"."a"."name", "public"."a"."email" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."a_id" = "public"."a"."id") INNER JOIN "public"."c" ON ("public"."c"."b_id" = "public"."b"."id") WHERE ("public"."a"."name" = 'alisha' AND "public"."b"."id" = 1);`,
		},
		{
			name:    "composite foreign key",
			driver:  sqlmanager_shared.PostgresDriver,
			table:   "public.a",
			columns: []string{"id", "name", "email"},
			joins: []*sqlJoin{
				{
					JoinType:  innerJoin,
					JoinTable: "public.b",
					BaseTable: "public.a",
					JoinColumnsMap: map[string]string{
						"a_id":  "id",
						"aa_id": "other_id",
					},
				},
			},
			whereClauses: []string{`"public"."a"."name" = 'alisha'`, `"public"."b"."email" = 'fake@email.com'`},
			expected:     `SELECT "public"."a"."id", "public"."a"."name", "public"."a"."email" FROM "public"."a" INNER JOIN "public"."b" ON (("public"."b"."a_id" = "public"."a"."id") AND ("public"."b"."aa_id" = "public"."a"."other_id")) WHERE ("public"."a"."name" = 'alisha' AND "public"."b"."email" = 'fake@email.com');`,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := BuildSelectJoinQuery(tt.driver, tt.table, tt.columns, tt.joins, tt.whereClauses, map[string]string{})
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func Test_buildSelectRecursiveQuery(t *testing.T) {
	tests := []struct {
		name          string
		driver        string
		table         string
		columns       []string
		columnInfoMap map[string]*sqlmanager_shared.ColumnInfo
		joins         []*sqlJoin
		whereClauses  []string
		dependencies  []*selfReferencingCircularDependency
		primaryKeyCol [][]string
		expected      string
	}{
		{
			name:          "one foreign key no joins",
			driver:        sqlmanager_shared.PostgresDriver,
			table:         "public.employees",
			columns:       []string{"employee_id", "name", "manager_id"},
			columnInfoMap: map[string]*sqlmanager_shared.ColumnInfo{},
			joins:         []*sqlJoin{},
			whereClauses:  []string{`"public"."employees"."name" = 'alisha'`},
			dependencies: []*selfReferencingCircularDependency{
				{PrimaryKeyColumns: []string{"employee_id"}, ForeignKeyColumns: [][]string{{"manager_id"}}},
			},
			expected: `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id" FROM "public"."employees" WHERE "public"."employees"."name" = 'alisha' UNION (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id" FROM "public"."employees" INNER JOIN "related" ON ("public"."employees"."employee_id" = "related"."manager_id"))) SELECT DISTINCT "employee_id", "name", "manager_id" FROM "related";`,
		},
		{
			name:          "json field",
			driver:        sqlmanager_shared.PostgresDriver,
			table:         "public.employees",
			columns:       []string{"employee_id", "name", "manager_id", "additional_info"},
			columnInfoMap: map[string]*sqlmanager_shared.ColumnInfo{"additional_info": {DataType: "json"}},
			joins:         []*sqlJoin{},
			whereClauses:  []string{`"public"."employees"."name" = 'alisha'`},
			dependencies: []*selfReferencingCircularDependency{
				{PrimaryKeyColumns: []string{"employee_id"}, ForeignKeyColumns: [][]string{{"manager_id"}}},
			},
			expected: `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", to_jsonb("public"."employees"."additional_info") AS "additional_info" FROM "public"."employees" WHERE "public"."employees"."name" = 'alisha' UNION (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", to_jsonb("public"."employees"."additional_info") AS "additional_info" FROM "public"."employees" INNER JOIN "related" ON ("public"."employees"."employee_id" = "related"."manager_id"))) SELECT DISTINCT "employee_id", "name", "manager_id", "additional_info" FROM "related";`,
		},
		{
			name:          "json field mysql",
			driver:        sqlmanager_shared.MysqlDriver,
			table:         "public.employees",
			columns:       []string{"employee_id", "name", "manager_id", "additional_info"},
			columnInfoMap: map[string]*sqlmanager_shared.ColumnInfo{"additional_info": {DataType: "json"}},
			joins:         []*sqlJoin{},
			whereClauses:  []string{`"public"."employees"."name" = 'alisha'`},
			dependencies: []*selfReferencingCircularDependency{
				{PrimaryKeyColumns: []string{"employee_id"}, ForeignKeyColumns: [][]string{{"manager_id"}}},
			},
			expected: "WITH RECURSIVE related AS (SELECT `public`.`employees`.`employee_id`, `public`.`employees`.`name`, `public`.`employees`.`manager_id`, `public`.`employees`.`additional_info` FROM `public`.`employees` WHERE \"public\".\"employees\".\"name\" = 'alisha' UNION (SELECT `public`.`employees`.`employee_id`, `public`.`employees`.`name`, `public`.`employees`.`manager_id`, `public`.`employees`.`additional_info` FROM `public`.`employees` INNER JOIN `related` ON (`public`.`employees`.`employee_id` = `related`.`manager_id`))) SELECT DISTINCT `employee_id`, `name`, `manager_id`, `additional_info` FROM `related`;",
		},
		{
			name:          "multiple foreign keys and joins",
			driver:        sqlmanager_shared.PostgresDriver,
			table:         "public.employees",
			columns:       []string{"employee_id", "name", "manager_id", "department_id", "big_boss_id"},
			columnInfoMap: map[string]*sqlmanager_shared.ColumnInfo{},
			joins: []*sqlJoin{
				{
					JoinType:  innerJoin,
					JoinTable: "public.departments",
					BaseTable: "public.employees",
					JoinColumnsMap: map[string]string{
						"id": "department_id",
					},
				},
			},
			whereClauses: []string{`"public"."employees"."name" = 'alisha'`, `"public"."departments"."department_id" = 1`},
			dependencies: []*selfReferencingCircularDependency{
				{PrimaryKeyColumns: []string{"employee_id"}, ForeignKeyColumns: [][]string{{"manager_id"}, {"big_boss_id"}}},
			},
			expected: `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."department_id", "public"."employees"."big_boss_id" FROM "public"."employees" INNER JOIN "public"."departments" ON ("public"."departments"."id" = "public"."employees"."department_id") WHERE ("public"."employees"."name" = 'alisha' AND "public"."departments"."department_id" = 1) UNION (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."department_id", "public"."employees"."big_boss_id" FROM "public"."employees" INNER JOIN "related" ON (("public"."employees"."employee_id" = "related"."manager_id") OR ("public"."employees"."employee_id" = "related"."big_boss_id")))) SELECT DISTINCT "employee_id", "name", "manager_id", "department_id", "big_boss_id" FROM "related";`,
		},
		{
			name:          "composite foreign keys",
			driver:        sqlmanager_shared.PostgresDriver,
			table:         "public.employees",
			columns:       []string{"employee_id", "department_id", "name", "manager_id", "building_id", "division_id"},
			columnInfoMap: map[string]*sqlmanager_shared.ColumnInfo{},
			joins: []*sqlJoin{
				{
					JoinType:  innerJoin,
					JoinTable: "public.departments",
					BaseTable: "public.employees",
					JoinColumnsMap: map[string]string{
						"id":       "building_id",
						"other_id": "another_id",
					},
				},
			},
			whereClauses: []string{`"public"."employees"."name" = 'alisha'`, `"public"."departments"."building_id" = 1`},
			dependencies: []*selfReferencingCircularDependency{
				{PrimaryKeyColumns: []string{"employee_id", "department_id"}, ForeignKeyColumns: [][]string{{"manager_id", "division_id"}}},
			},
			expected: `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."department_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."building_id", "public"."employees"."division_id" FROM "public"."employees" INNER JOIN "public"."departments" ON (("public"."departments"."id" = "public"."employees"."building_id") AND ("public"."departments"."other_id" = "public"."employees"."another_id")) WHERE ("public"."employees"."name" = 'alisha' AND "public"."departments"."building_id" = 1) UNION (SELECT "public"."employees"."employee_id", "public"."employees"."department_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."building_id", "public"."employees"."division_id" FROM "public"."employees" INNER JOIN "related" ON (("public"."employees"."employee_id" = "related"."manager_id") AND ("public"."employees"."department_id" = "related"."division_id")))) SELECT DISTINCT "employee_id", "department_id", "name", "manager_id", "building_id", "division_id" FROM "related";`,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := BuildSelectRecursiveQuery(tt.driver, tt.table, tt.columns, tt.columnInfoMap, tt.dependencies, tt.joins, tt.whereClauses)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func Test_filterForeignKeysWithSubset_partialtables(t *testing.T) {
	var emptyWhere *string
	where := "id = '36f594af-6d53-4a48-a9b7-b889e2df349e'"
	runConfigMap := map[string]*tabledependency.RunConfig{
		"circle.addresses": {
			Table:         "circle.addresses",
			InsertColumns: []string{"id"},
			SelectColumns: []string{"id"},
			DependsOn:     []*tabledependency.DependsOn{},
			RunType:       tabledependency.RunTypeInsert,
			PrimaryKeys:   []string{"id"},
			WhereClause:   &where,
		},

		"circle.customers": {
			Table:         "circle.customers",
			InsertColumns: []string{"id", "address_id"},
			SelectColumns: []string{"id", "address_id"},
			DependsOn:     []*tabledependency.DependsOn{{Table: "circle.addresses", Columns: []string{"id"}}},
			RunType:       tabledependency.RunTypeInsert,
			PrimaryKeys:   []string{"id"},
			WhereClause:   emptyWhere,
		},
	}

	constraints := map[string][]*sqlmanager_shared.ForeignConstraint{
		"circle.addresses": {
			{
				Columns:     []string{"order_id"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "circle.orders", Columns: []string{"id"}},
			},
		},
		"circle.customers": {
			{
				Columns:     []string{"address_id"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "circle.addresses", Columns: []string{"id"}},
			},
		},
	}

	whereClauses := map[string]string{
		"circle.addresses": "id = '36f594af-6d53-4a48-a9b7-b889e2df349e'",
	}

	expected := map[string][]*sqlmanager_shared.ForeignConstraint{
		"circle.addresses": {},
		"circle.customers": {
			{
				Columns:     []string{"address_id"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "circle.addresses", Columns: []string{"id"}},
			},
		},
	}

	actual := filterForeignKeysWithSubset(runConfigMap, constraints, whereClauses)
	require.Equal(t, expected, actual)
}

func Test_qualifyWhereColumnNames_mysql(t *testing.T) {
	tests := []struct {
		name     string
		where    string
		table    string
		expected string
	}{
		{
			name:     "simple",
			where:    "name = 'alisha'",
			table:    "public.a",
			expected: "public.a.name = 'alisha'",
		},
		{
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			table:    "public.a",
			expected: "public.a.name = 'alisha'",
		},
		{
			name:     "multiple",
			where:    "name = 'alisha' and id = 1",
			table:    "public.a",
			expected: "public.a.name = 'alisha' and public.a.id = 1",
		},
		{
			name:     "where subquery",
			where:    "film_id IN(SELECT film_id FROM film_category INNER JOIN category USING(category_id) WHERE name='Action')",
			table:    "public.film",
			expected: "public.film.film_id in (select film_id from film_category join category using (category_id) where name = 'Action')",
		},
		{
			name:     "quoted column names",
			where:    "`name with space` = 'hey' and PascalCase = 'other'",
			table:    "public.order",
			expected: "public.`order`.`name with space` = 'hey' and public.`order`.PascalCase = 'other'",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qualifyWhereColumnNames(sqlmanager_shared.MysqlDriver, tt.where, tt.table)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func Test_qualifyWhereColumnNames_postgres(t *testing.T) {
	tests := []struct {
		name     string
		where    string
		table    string
		expected string
	}{
		{
			name:     "simple",
			where:    "name = 'alisha'",
			table:    "public.a",
			expected: `public.a.name = 'alisha'`,
		},
		{
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			table:    "public.a",
			expected: `public.a.name = 'alisha'`,
		},
		{
			name:     "simple",
			where:    `"bad name" = 'alisha'`,
			table:    "public.order",
			expected: `public."order"."bad name" = 'alisha'`,
		},
		{
			name:     "multiple",
			where:    "name = 'alisha' and id = 1  or age = 2",
			table:    "public.a",
			expected: `(public.a.name = 'alisha' AND public.a.id = 1) OR public.a.age = 2`,
		},
		{
			name:     "where subquery",
			where:    "film_id IN(SELECT film_id FROM film_category INNER JOIN category USING(category_id) WHERE name='Action');",
			table:    "public.film",
			expected: `public.film.film_id IN (SELECT film_id FROM film_category JOIN category USING (category_id) WHERE name = 'Action')`,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qualifyWhereColumnNames(sqlmanager_shared.PostgresDriver, tt.where, tt.table)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func Test_qualifyWhereWithTableAlias(t *testing.T) {
	tests := []struct {
		driver   string
		name     string
		where    string
		alias    string
		expected string
	}{
		{
			driver:   sqlmanager_shared.PostgresDriver,
			name:     "simple",
			where:    "name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sqlmanager_shared.PostgresDriver,
			name:     "hash alias",
			where:    "composite_keys.department.department_id = '1'",
			alias:    "50d89c0f3af602",
			expected: `"50d89c0f3af602".department_id = '1'`,
		},
		{
			driver:   sqlmanager_shared.PostgresDriver,
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sqlmanager_shared.PostgresDriver,
			name:     "multiple",
			where:    "name = 'alisha' and id = 1  or age = 2",
			alias:    "alias",
			expected: `(alias.name = 'alisha' AND alias.id = 1) OR alias.age = 2`,
		},
		{
			driver:   sqlmanager_shared.MysqlDriver,
			name:     "simple",
			where:    "name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sqlmanager_shared.MysqlDriver,
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sqlmanager_shared.MysqlDriver,
			name:     "multiple",
			where:    "name = 'alisha' and id = 1  or age = 2",
			alias:    "alias",
			expected: `alias.name = 'alisha' and alias.id = 1 or alias.age = 2`,
		},
		{
			driver:   sqlmanager_shared.MysqlDriver,
			name:     "hash alias",
			where:    "composite_keys.department.department_id = '1'",
			alias:    "50d89c0f3af602",
			expected: "`50d89c0f3af602`.department_id = '1'",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qualifyWhereWithTableAlias(tt.driver, tt.where, tt.alias)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func TestGetPrimaryToForeignTableMapFromRunConfigs(t *testing.T) {
	tests := []struct {
		name       string
		runConfigs []*tabledependency.RunConfig
		expected   map[string][]string
	}{
		{
			name:       "no configs",
			runConfigs: nil,
			expected:   map[string][]string{},
		},
		{
			name: "single config without dependencies",
			runConfigs: []*tabledependency.RunConfig{
				{Table: "table1", DependsOn: nil},
			},
			expected: map[string][]string{},
		},
		{
			name: "single config with one dependency",
			runConfigs: []*tabledependency.RunConfig{
				{Table: "table1", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
			},
			expected: map[string][]string{
				"table2": {"table1"},
			},
		},
		{
			name: "multiple configs with shared dependencies",
			runConfigs: []*tabledependency.RunConfig{
				{Table: "table1", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
				{Table: "table3", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
			},
			expected: map[string][]string{
				"table2": {"table1", "table3"},
			},
		},
		{
			name: "config with multiple unique dependencies",
			runConfigs: []*tabledependency.RunConfig{
				{Table: "table1", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}, {Table: "table3"}}},
			},
			expected: map[string][]string{
				"table2": {"table1"},
				"table3": {"table1"},
			},
		},
		{
			name: "circular dependencies",
			runConfigs: []*tabledependency.RunConfig{
				{Table: "table1", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
				{Table: "table2", DependsOn: []*tabledependency.DependsOn{{Table: "table1"}}},
			},
			expected: map[string][]string{
				"table2": {"table1"},
				"table1": {"table2"},
			},
		},
		{
			name: "multiple configs with duplicates",
			runConfigs: []*tabledependency.RunConfig{
				{Table: "table1", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
				{Table: "table1", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
				{Table: "table3", DependsOn: []*tabledependency.DependsOn{{Table: "table2"}}},
			},
			expected: map[string][]string{
				"table2": {"table1", "table3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getPrimaryToForeignTableMapFromRunConfigs(tt.runConfigs)
			for table, dependencies := range actual {
				expected, exists := tt.expected[table]
				require.True(t, exists)
				for _, dep := range dependencies {
					require.Contains(t, expected, dep)
				}
			}
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

func Test_BuildSelectQueryMap_Campus(t *testing.T) {
	driver := "postgres"
	_ = driver
	tableDeps := map[string][]*sqlmanager_shared.ForeignConstraint{}

	bits, err := os.ReadFile("./tabledeps.json")
	require.NoError(t, err)
	err = json.Unmarshal(bits, &tableDeps)
	require.NoError(t, err)

	runConfigs := []*tabledependency.RunConfig{}
	bits, err = os.ReadFile("./runconfigs.json")
	require.NoError(t, err)
	err = json.Unmarshal(bits, &runConfigs)
	require.NoError(t, err)

	groupedColInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{}
	bits, err = os.ReadFile("./groupedcolinfo.json")
	require.NoError(t, err)
	err = json.Unmarshal(bits, &groupedColInfo)
	require.NoError(t, err)

	output, err := BuildSelectQueryMap(driver, tableDeps, runConfigs, true, groupedColInfo)
	require.NoError(t, err)

	bits, _ = json.Marshal(output)
	fmt.Println(string(bits))

	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5434/postgres?sslmode=disable")
	require.NoError(t, err)

	for table, selectQueryrunType := range output {
		_ = table
		statement := selectQueryrunType[tabledependency.RunTypeInsert]
		require.NotEmpty(t, statement)
		fmt.Println("===========")
		fmt.Println(statement)
		_, err := db.Exec(context.Background(), statement)
		require.NoError(t, err, "failed on table", "table", table)
	}
}
