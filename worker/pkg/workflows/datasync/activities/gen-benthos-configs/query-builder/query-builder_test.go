package genbenthosconfigs_querybuilder

import (
	"fmt"
	"testing"

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
			sql, err := buildSelectQuery(tt.driver, tt.table, tt.columns, &where)
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
			response, err := buildSelectJoinQuery(tt.driver, tt.table, tt.columns, tt.joins, tt.whereClauses)
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
			response, err := buildSelectRecursiveQuery(tt.driver, tt.table, tt.columns, tt.columnInfoMap, tt.dependencies, tt.joins, tt.whereClauses)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func Test_BuildSelectQueryMap(t *testing.T) {
	whereId := "id = 1"
	tests := []struct {
		name                          string
		driver                        string
		subsetByForeignKeyConstraints bool
		tableDependencies             map[string][]*sqlmanager_shared.ForeignConstraint
		dependencyConfigs             []*tabledependency.RunConfig
		expected                      map[string]map[tabledependency.RunType]string
	}{
		{
			name:                          "select no subset",
			driver:                        sqlmanager_shared.PostgresDriver,
			subsetByForeignKeyConstraints: true,
			tableDependencies:             map[string][]*sqlmanager_shared.ForeignConstraint{},
			dependencyConfigs: []*tabledependency.RunConfig{
				{Table: "public.users", SelectColumns: []string{"id", "name"}, InsertColumns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}, RunType: tabledependency.RunTypeInsert},
				{Table: "public.accounts", SelectColumns: []string{"id", "name"}, InsertColumns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}, RunType: tabledependency.RunTypeInsert},
			},
			expected: map[string]map[tabledependency.RunType]string{
				"public.users":    {tabledependency.RunTypeInsert: `SELECT "id", "name" FROM "public"."users";`},
				"public.accounts": {tabledependency.RunTypeInsert: `SELECT "id", "name" FROM "public"."accounts";`},
			},
		},
		{
			name:                          "select subset no foreign keys",
			driver:                        sqlmanager_shared.PostgresDriver,
			subsetByForeignKeyConstraints: true,
			tableDependencies:             map[string][]*sqlmanager_shared.ForeignConstraint{},
			dependencyConfigs: []*tabledependency.RunConfig{
				{Table: "public.users", SelectColumns: []string{"id", "name"}, InsertColumns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId, RunType: tabledependency.RunTypeInsert},
				{Table: "public.accounts", SelectColumns: []string{"id", "name"}, InsertColumns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}, RunType: tabledependency.RunTypeInsert},
			},
			expected: map[string]map[tabledependency.RunType]string{
				"public.users":    {tabledependency.RunTypeInsert: `SELECT "id", "name" FROM "public"."users" WHERE id = 1;`},
				"public.accounts": {tabledependency.RunTypeInsert: `SELECT "id", "name" FROM "public"."accounts";`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			sql, err := BuildSelectQueryMap(tt.driver, tt.tableDependencies, tt.dependencyConfigs, tt.subsetByForeignKeyConstraints, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
			require.NoError(t, err)
			require.Equal(t, tt.expected, sql)
		})
	}
}

func Test_BuildSelectQueryMap_SubsetsCompositeForeignKeys(t *testing.T) {
	aWhere := "name = 'bob'"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.b": {
			{Columns: []string{"a_name", "a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"name", "id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "name"}, InsertColumns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &aWhere},
		{Table: "public.b", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "a_name", "a_id"}, InsertColumns: []string{"id", "a_name", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id", "name"}}}},
	}
	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.a": {tabledependency.RunTypeInsert: `SELECT "id", "name" FROM "public"."a" WHERE public.a.name = 'bob';`},
			"public.b": {tabledependency.RunTypeInsert: `SELECT "public"."b"."id", "public"."b"."a_name", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON (("public"."a"."id" = "public"."b"."a_id") AND ("public"."a"."name" = "public"."b"."a_name")) WHERE public.a.name = 'bob';`},
		}

	sql, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_BuildSelectQueryMap_CircularDependency(t *testing.T) {
	whereName := "name = 'neo'"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
		"public.c": {
			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		},
		"public.a": {
			{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "name"}, InsertColumns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereName},
		{Table: "public.c", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "b_id"}, InsertColumns: []string{"id", "b_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		{Table: "public.b", RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.a": {tabledependency.RunTypeInsert: `SELECT "public"."a"."id", "public"."a"."c_id" FROM "public"."a" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."a"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.name = 'neo';`},
			"public.b": {
				tabledependency.RunTypeInsert: `SELECT "public"."b"."id", "public"."b"."a_id", "public"."b"."name" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."a"."c_id") WHERE public.b.name = 'neo';`,
				tabledependency.RunTypeUpdate: `SELECT "public"."b"."id", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."a"."c_id") WHERE public.b.name = 'neo';`,
			},
			"public.c": {tabledependency.RunTypeInsert: `SELECT "public"."c"."id", "public"."c"."b_id" FROM "public"."c" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.b.name = 'neo';`},
		}
	sql, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_BuildSelectQueryMap_circularDependency_additional_table(t *testing.T) {
	whereName := "name = 'neo'"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.addresses": {
			{Columns: []string{"order_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.orders", Columns: []string{"id"}}},
		},
		"public.customers": {
			{Columns: []string{"address_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.addresses", Columns: []string{"id"}}},
		},
		"public.orders": {
			{Columns: []string{"customer_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.customers", Columns: []string{"id"}}},
		},
		"public.payments": {
			{Columns: []string{"customer_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.customers", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.orders", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "customer_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.addresses", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "order_id"}, InsertColumns: []string{"id", "order_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.orders", Columns: []string{"id"}}}, WhereClause: &whereName},
		{Table: "public.customers", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "address_id"}, InsertColumns: []string{"id", "address_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.addresses", Columns: []string{"id"}}}},
		{Table: "public.payments", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "customer_id"}, InsertColumns: []string{"id", "customer_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.customers", Columns: []string{"id"}}}},
		{Table: "public.orders", RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "customer_id"}, InsertColumns: []string{"customer_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.orders", Columns: []string{"id"}}, {Table: "public.customers", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.addresses": {tabledependency.RunTypeInsert: `SELECT "public"."addresses"."id", "public"."addresses"."order_id" FROM "public"."addresses" INNER JOIN "public"."orders" ON ("public"."orders"."id" = "public"."addresses"."order_id") INNER JOIN "public"."customers" ON ("public"."customers"."id" = "public"."orders"."customer_id") WHERE public.addresses.name = 'neo';`},
			"public.customers": {tabledependency.RunTypeInsert: `SELECT "public"."customers"."id", "public"."customers"."address_id" FROM "public"."customers" INNER JOIN "public"."addresses" ON ("public"."addresses"."id" = "public"."customers"."address_id") INNER JOIN "public"."orders" ON ("public"."orders"."id" = "public"."addresses"."order_id") WHERE public.addresses.name = 'neo';`},
			"public.orders": {
				tabledependency.RunTypeInsert: `SELECT "public"."orders"."id", "public"."orders"."customer_id" FROM "public"."orders" INNER JOIN "public"."customers" ON ("public"."customers"."id" = "public"."orders"."customer_id") INNER JOIN "public"."addresses" ON ("public"."addresses"."id" = "public"."customers"."address_id") WHERE public.addresses.name = 'neo';`,
				tabledependency.RunTypeUpdate: `SELECT "public"."orders"."id", "public"."orders"."customer_id" FROM "public"."orders" INNER JOIN "public"."customers" ON ("public"."customers"."id" = "public"."orders"."customer_id") INNER JOIN "public"."addresses" ON ("public"."addresses"."id" = "public"."customers"."address_id") WHERE public.addresses.name = 'neo';`,
			},
			"public.payments": {tabledependency.RunTypeInsert: `SELECT "public"."payments"."id", "public"."payments"."customer_id" FROM "public"."payments" INNER JOIN "public"."customers" ON ("public"."customers"."id" = "public"."payments"."customer_id") INNER JOIN "public"."addresses" ON ("public"."addresses"."id" = "public"."customers"."address_id") INNER JOIN "public"."orders" ON ("public"."orders"."id" = "public"."addresses"."order_id") WHERE public.addresses.name = 'neo';`},
		}
	sql, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_BuildSelectQueryMap_DoubleCircularDependencyRoot(t *testing.T) {
	whereId := "id = 1"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"a_a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "aa_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
		{Table: "public.a", RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "aa_id"}, InsertColumns: []string{"a_id", "aa_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"id", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
	}

	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.a": {
				tabledependency.RunTypeInsert: `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."aa_id" FROM "public"."a" WHERE public.a.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."aa_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "aa_id" FROM "related";`,
				tabledependency.RunTypeUpdate: `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."aa_id" FROM "public"."a" WHERE public.a.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."aa_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "aa_id" FROM "related";`,
			},
			"public.b": {tabledependency.RunTypeInsert: `SELECT "public"."b"."id", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.a.id = 1;`},
		}
	sql, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_BuildSelectQueryMap_doubleCircularDependencyRoot_mysql(t *testing.T) {
	whereId := "id = 1"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"a_a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "a_a_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
		{Table: "public.a", RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "a_a_id"}, InsertColumns: []string{"a_id", "a_a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"id", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
	}

	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.a": {
				tabledependency.RunTypeInsert: "WITH RECURSIVE related AS (SELECT `public`.`a`.`id`, `public`.`a`.`a_id`, `public`.`a`.`a_a_id` FROM `public`.`a` WHERE public.a.id = 1 UNION (SELECT `public`.`a`.`id`, `public`.`a`.`a_id`, `public`.`a`.`a_a_id` FROM `public`.`a` INNER JOIN `related` ON ((`public`.`a`.`id` = `related`.`a_id`) OR (`public`.`a`.`id` = `related`.`a_a_id`)))) SELECT DISTINCT `id`, `a_id`, `a_a_id` FROM `related`;",
				tabledependency.RunTypeUpdate: "WITH RECURSIVE related AS (SELECT `public`.`a`.`id`, `public`.`a`.`a_id`, `public`.`a`.`a_a_id` FROM `public`.`a` WHERE public.a.id = 1 UNION (SELECT `public`.`a`.`id`, `public`.`a`.`a_id`, `public`.`a`.`a_a_id` FROM `public`.`a` INNER JOIN `related` ON ((`public`.`a`.`id` = `related`.`a_id`) OR (`public`.`a`.`id` = `related`.`a_a_id`)))) SELECT DISTINCT `id`, `a_id`, `a_a_id` FROM `related`;",
			},
			"public.b": {tabledependency.RunTypeInsert: "SELECT `public`.`b`.`id`, `public`.`b`.`a_id` FROM `public`.`b` INNER JOIN `public`.`a` ON (`public`.`a`.`id` = `public`.`b`.`a_id`) WHERE public.a.id = 1;"},
		}
	sql, err := BuildSelectQueryMap(sqlmanager_shared.MysqlDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_BuildSelectQueryMap_DoubleCircularDependencyChild(t *testing.T) {
	whereId := "id = 1"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"a_a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "a_a_id", "b_id"}, InsertColumns: []string{"id", "b_id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.a", RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "a_a_id"}, InsertColumns: []string{"a_id", "a_a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"id", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}, WhereClause: &whereId},
	}

	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.a": {
				tabledependency.RunTypeInsert: `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id", "public"."a"."b_id" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."a"."b_id") WHERE public.b.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id", "public"."a"."b_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "a_a_id", "b_id" FROM "related";`,
				tabledependency.RunTypeUpdate: `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."a"."b_id") WHERE public.b.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "a_a_id" FROM "related";`,
			},
			"public.b": {tabledependency.RunTypeInsert: `SELECT "id", "a_id" FROM "public"."b" WHERE public.b.id = 1;`},
		}
	sql, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_BuildSelectQueryMap_shouldContinue(t *testing.T) {
	aWhere := "id = 1"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"d_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.d", Columns: []string{"id"}}},
		},
		"public.d": {
			{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		},
		"public.e": {
			{Columns: []string{"d_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.d", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &aWhere},
		{Table: "public.b", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "a_id", "d_id"}, InsertColumns: []string{"id", "a_id", "d_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
		{Table: "public.c", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.d", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		{Table: "public.e", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "d_id"}, InsertColumns: []string{"id", "d_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.d", Columns: []string{"id"}}}},
	}

	expected :=
		map[string]map[tabledependency.RunType]string{
			"public.a": {tabledependency.RunTypeInsert: `SELECT "id" FROM "public"."a" WHERE public.a.id = 1;`},
			"public.b": {tabledependency.RunTypeInsert: `SELECT "public"."b"."id", "public"."b"."a_id", "public"."b"."d_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.a.id = 1;`},
			"public.c": {tabledependency.RunTypeInsert: `SELECT "id" FROM "public"."c";`},
			"public.d": {tabledependency.RunTypeInsert: `SELECT "id", "c_id" FROM "public"."d";`},
			"public.e": {tabledependency.RunTypeInsert: `SELECT "id", "d_id" FROM "public"."e";`},
		}
	sql, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})

	require.NoError(t, err)
	require.Equal(t, expected, sql)
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
