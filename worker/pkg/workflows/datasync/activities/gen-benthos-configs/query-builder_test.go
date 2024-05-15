package genbenthosconfigs_activity

import (
	"encoding/json"
	"fmt"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
			require.Equal(t, tt.expected, sql)
		})
	}
}

// func Test_buildSelectJoinQuery(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		driver       string
// 		schema       string
// 		table        string
// 		columns      []string
// 		joins        []*sqlJoin
// 		whereClauses []string
// 		expected     string
// 	}{
// 		{
// 			name:    "simple",
// 			driver:  "postgres",
// 			schema:  "public",
// 			table:   "a",
// 			columns: []string{"id", "name", "email"},
// 			joins: []*sqlJoin{
// 				{
// 					JoinType:  innerJoin,
// 					JoinTable: "public.b",
// 					BaseTable: "public.a",
// 					JoinColumnsMap: map[string]string{
// 						"a_id": "id",
// 					},
// 				},
// 			},
// 			whereClauses: []string{`"public"."a"."name" = 'alisha'`, `"public"."b"."email" = 'fake@email.com'`},
// 			expected:     `SELECT "public"."a"."id", "public"."a"."name", "public"."a"."email" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."a_id" = "public"."a"."id") WHERE ("public"."a"."name" = 'alisha' AND "public"."b"."email" = 'fake@email.com');`,
// 		},
// 		{
// 			name:    "multiple joins",
// 			driver:  "postgres",
// 			schema:  "public",
// 			table:   "a",
// 			columns: []string{"id", "name", "email"},
// 			joins: []*sqlJoin{
// 				{
// 					JoinType:  innerJoin,
// 					JoinTable: "public.b",
// 					BaseTable: "public.a",
// 					JoinColumnsMap: map[string]string{
// 						"a_id": "id",
// 					},
// 				},
// 				{
// 					JoinType:  innerJoin,
// 					JoinTable: "public.c",
// 					BaseTable: "public.b",
// 					JoinColumnsMap: map[string]string{
// 						"b_id": "id",
// 					},
// 				},
// 			},
// 			whereClauses: []string{`"public"."a"."name" = 'alisha'`, `"public"."b"."id" = 1`},
// 			expected:     `SELECT "public"."a"."id", "public"."a"."name", "public"."a"."email" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."a_id" = "public"."a"."id") INNER JOIN "public"."c" ON ("public"."c"."b_id" = "public"."b"."id") WHERE ("public"."a"."name" = 'alisha' AND "public"."b"."id" = 1);`,
// 		},
// 		{
// 			name:    "composite foreign key",
// 			driver:  "postgres",
// 			schema:  "public",
// 			table:   "a",
// 			columns: []string{"id", "name", "email"},
// 			joins: []*sqlJoin{
// 				{
// 					JoinType:  innerJoin,
// 					JoinTable: "public.b",
// 					BaseTable: "public.a",
// 					JoinColumnsMap: map[string]string{
// 						"a_id":  "id",
// 						"aa_id": "other_id",
// 					},
// 				},
// 			},
// 			whereClauses: []string{`"public"."a"."name" = 'alisha'`, `"public"."b"."email" = 'fake@email.com'`},
// 			expected:     `SELECT "public"."a"."id", "public"."a"."name", "public"."a"."email" FROM "public"."a" INNER JOIN "public"."b" ON (("public"."b"."a_id" = "public"."a"."id") AND ("public"."b"."aa_id" = "public"."a"."other_id")) WHERE ("public"."a"."name" = 'alisha' AND "public"."b"."email" = 'fake@email.com');`,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
// 			response, err := buildSelectJoinQuery(tt.driver, tt.schema, tt.table, tt.columns, tt.joins, tt.whereClauses)
// 			require.NoError(t, err)
// 			require.Equal(t, tt.expected, response)
// 		})
// 	}
// }

// func Test_buildSelectRecursiveQuery(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		driver        string
// 		schema        string
// 		table         string
// 		columns       []string
// 		joins         []*sqlJoi
// 		whereClauses  []string
// 		foreignKeys   []string
// 		primaryKeyCol string
// 		expected      string
// 	}{
// 		{
// 			name:          "one foreign key no joins",
// 			driver:        "postgres",
// 			schema:        "public",
// 			table:         "employees",
// 			columns:       []string{"employee_id", "name", "manager_id"},
// 			joins:         []*sqlJoin{},
// 			whereClauses:  []string{`"public"."employees"."name" = 'alisha'`},
// 			foreignKeys:   []string{"manager_id"},
// 			primaryKeyCol: "employee_id",
// 			expected:      `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id" FROM "public"."employees" WHERE "public"."employees"."name" = 'alisha' UNION (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id" FROM "public"."employees" INNER JOIN "related" ON ("public"."employees"."employee_id" = "related"."manager_id"))) SELECT DISTINCT "employee_id", "name", "manager_id" FROM "related";`,
// 		},
// 		{
// 			name:    "multiple foreign keys and joins",
// 			driver:  "postgres",
// 			schema:  "public",
// 			table:   "employees",
// 			columns: []string{"employee_id", "name", "manager_id", "department_id", "big_boss_id"},
// 			joins: []*sqlJoin{
// 				{
// 					JoinType:  innerJoin,
// 					JoinTable: "public.departments",
// 					BaseTable: "public.employees",
// 					JoinColumnsMap: map[string]string{
// 						"id": "department_id",
// 					},
// 				},
// 			},
// 			whereClauses:  []string{`"public"."employees"."name" = 'alisha'`, `"public"."departments"."department_id" = 1`},
// 			foreignKeys:   []string{"manager_id", "big_boss_id"},
// 			primaryKeyCol: "employee_id",
// 			expected:      `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."department_id", "public"."employees"."big_boss_id" FROM "public"."employees" INNER JOIN "public"."departments" ON ("public"."departments"."id" = "public"."employees"."department_id") WHERE ("public"."employees"."name" = 'alisha' AND "public"."departments"."department_id" = 1) UNION (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."department_id", "public"."employees"."big_boss_id" FROM "public"."employees" INNER JOIN "related" ON (("public"."employees"."employee_id" = "related"."manager_id") OR ("public"."employees"."employee_id" = "related"."big_boss_id")))) SELECT DISTINCT "employee_id", "name", "manager_id", "department_id", "big_boss_id" FROM "related";`,
// 		},
// 		{
// 			name:    "composite foreign keys",
// 			driver:  "postgres",
// 			schema:  "public",
// 			table:   "employees",
// 			columns: []string{"employee_id", "name", "manager_id", "department_id", "big_boss_id"},
// 			joins: []*sqlJoin{
// 				{
// 					JoinType:  innerJoin,
// 					JoinTable: "public.departments",
// 					BaseTable: "public.employees",
// 					JoinColumnsMap: map[string]string{
// 						"id":       "department_id",
// 						"other_id": "another_id",
// 					},
// 				},
// 			},
// 			whereClauses:  []string{`"public"."employees"."name" = 'alisha'`, `"public"."departments"."department_id" = 1`},
// 			foreignKeys:   []string{"manager_id", "big_boss_id"},
// 			primaryKeyCol: "employee_id",
// 			expected:      `WITH RECURSIVE related AS (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."department_id", "public"."employees"."big_boss_id" FROM "public"."employees" INNER JOIN "public"."departments" ON (("public"."departments"."id" = "public"."employees"."department_id") AND ("public"."departments"."other_id" = "public"."employees"."another_id")) WHERE ("public"."employees"."name" = 'alisha' AND "public"."departments"."department_id" = 1) UNION (SELECT "public"."employees"."employee_id", "public"."employees"."name", "public"."employees"."manager_id", "public"."employees"."department_id", "public"."employees"."big_boss_id" FROM "public"."employees" INNER JOIN "related" ON (("public"."employees"."employee_id" = "related"."manager_id") OR ("public"."employees"."employee_id" = "related"."big_boss_id")))) SELECT DISTINCT "employee_id", "name", "manager_id", "department_id", "big_boss_id" FROM "related";`,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
// 			response, err := buildSelectRecursiveQuery(tt.driver, tt.schema, tt.table, tt.columns, tt.foreignKeys, tt.primaryKeyCol, tt.joins, tt.whereClauses)
// 			require.NoError(t, err)
// 			require.Equal(t, tt.expected, response)
// 		})
// 	}
// }

func Test_buildSelectQueryMap(t *testing.T) {
	whereId := "id = 1"
	tests := []struct {
		name                          string
		driver                        string
		subsetByForeignKeyConstraints bool
		mappings                      map[string]*tableMapping
		sourceTableOpts               map[string]*sqlSourceTableOptions
		tableDependencies             map[string][]*sql_manager.ForeignConstraint
		dependencyConfigs             []*tabledependency.RunConfig
		expected                      map[string]string
	}{
		{
			name:                          "select no subset",
			driver:                        "postgres",
			subsetByForeignKeyConstraints: true,
			mappings: map[string]*tableMapping{
				"public.users": {
					Schema: "public",
					Table:  "users",
					Mappings: []*mgmtv1alpha1.JobMapping{
						{
							Schema: "public",
							Table:  "users",
							Column: "id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
							},
						},
						{
							Schema: "public",
							Table:  "users",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
							},
						},
					},
				},
				"public.accounts": {
					Schema: "public",
					Table:  "accounts",
					Mappings: []*mgmtv1alpha1.JobMapping{
						{
							Schema: "public",
							Table:  "accounts",
							Column: "id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
							},
						},
						{
							Schema: "public",
							Table:  "accounts",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
							},
						},
					},
				},
			},
			sourceTableOpts:   map[string]*sqlSourceTableOptions{},
			tableDependencies: map[string][]*sql_manager.ForeignConstraint{},
			dependencyConfigs: []*tabledependency.RunConfig{
				{Table: "public.users", DependsOn: []*tabledependency.DependsOn{}},
				{Table: "public.accounts", DependsOn: []*tabledependency.DependsOn{}},
			},
			expected: map[string]string{
				"public.users":    `SELECT "id", "name" FROM "public"."users";`,
				"public.accounts": `SELECT "id", "name" FROM "public"."accounts";`,
			},
		},
		{
			name:                          "select subset no foreign keys",
			driver:                        "postgres",
			subsetByForeignKeyConstraints: true,
			mappings: map[string]*tableMapping{
				"public.users": {
					Schema: "public",
					Table:  "users",
					Mappings: []*mgmtv1alpha1.JobMapping{
						{
							Schema: "public",
							Table:  "users",
							Column: "id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
							},
						},
						{
							Schema: "public",
							Table:  "users",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
							},
						},
					},
				},
				"public.accounts": {
					Schema: "public",
					Table:  "accounts",
					Mappings: []*mgmtv1alpha1.JobMapping{
						{
							Schema: "public",
							Table:  "accounts",
							Column: "id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
							},
						},
						{
							Schema: "public",
							Table:  "accounts",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
							},
						},
					},
				},
			},
			sourceTableOpts: map[string]*sqlSourceTableOptions{
				"public.users": {
					WhereClause: &whereId,
				},
			},
			tableDependencies: map[string][]*sql_manager.ForeignConstraint{},
			dependencyConfigs: []*tabledependency.RunConfig{
				{Table: "public.users", DependsOn: []*tabledependency.DependsOn{}},
				{Table: "public.accounts", DependsOn: []*tabledependency.DependsOn{}},
			},
			expected: map[string]string{
				"public.users":    `SELECT "id", "name" FROM "public"."users" WHERE id = 1;`,
				"public.accounts": `SELECT "id", "name" FROM "public"."accounts";`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			sql, err := buildSelectQueryMap(tt.driver, tt.mappings, tt.sourceTableOpts, tt.tableDependencies, tt.dependencyConfigs, tt.subsetByForeignKeyConstraints)
			require.NoError(t, err)
			require.Equal(t, tt.expected, sql)
		})
	}
}

func Test_buildSelectQueryMap_SubsetsForeignKeys(t *testing.T) {
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.d": {
			Schema: "public",
			Table:  "d",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "d",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "d",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	bWhere := "name = 'bob'"
	cWhere := "id = 1"
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.b": {WhereClause: &bWhere},
		"public.c": {WhereClause: &cWhere},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
		"public.c": {
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
		"public.d": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.d", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id" FROM "public"."a";`,
			"public.b": `SELECT "id", "name", "a_id" FROM "public"."b" WHERE public.b.name = 'bob';`,
			"public.c": `SELECT "public"."c"."id", "public"."c"."b_id" FROM "public"."c" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE (public.b.name = 'bob' AND public.c.id = 1);`,
			"public.d": `SELECT "public"."d"."id", "public"."d"."c_id" FROM "public"."d" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."d"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE (public.b.name = 'bob' AND public.c.id = 1);`,
		}

	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_SubsetsCompositeForeignKeys(t *testing.T) {
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	aWhere := "name = 'bob'"
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.a": {WhereClause: &aWhere},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "a_name", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "name"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id", "name"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id", "name" FROM "public"."a" WHERE public.a.name = 'bob';`,
			"public.b": `SELECT "public"."b"."id", "public"."b"."a_name", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON (("public"."a"."id" = "public"."b"."a_id") AND ("public"."a"."name" = "public"."b"."a_name")) WHERE public.a.name = 'bob';`,
		}

	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_SubsetsOffForeignKeys(t *testing.T) {
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.d": {
			Schema: "public",
			Table:  "d",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "d",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "d",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	bWhere := "name = 'bob'"
	cWhere := "id = 1"
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.b": {WhereClause: &bWhere},
		"public.c": {WhereClause: &cWhere},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
		"public.c": {
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
		"public.d": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.d", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id" FROM "public"."a";`,
			"public.b": `SELECT "id", "name", "a_id" FROM "public"."b" WHERE name = 'bob';`,
			"public.c": `SELECT "id", "b_id" FROM "public"."c" WHERE id = 1;`,
			"public.d": `SELECT "id", "c_id" FROM "public"."d";`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, false)

	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_CircularDependency(t *testing.T) {
	whereName := "name = 'neo'"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.b": {
			WhereClause: &whereName,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.b": {
			{Column: "a_id", IsNullable: true, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
		"public.c": {
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
		"public.a": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.b", Columns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		{Table: "public.b", Columns: []string{"a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.b": `SELECT "id", "name", "a_id" FROM "public"."b" WHERE public.b.name = 'neo';`,
			"public.c": `SELECT "public"."c"."id", "public"."c"."b_id" FROM "public"."c" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.name = 'neo';`,
			"public.a": `SELECT "public"."a"."id", "public"."a"."c_id" FROM "public"."a" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."a"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.name = 'neo';`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)

	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_MultiplSubsets(t *testing.T) {
	whereId := "id = 1"
	whereName := "name = 'neo'"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.d": {
			Schema: "public",
			Table:  "d",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "d",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.e": {
			Schema: "public",
			Table:  "e",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "e",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "e",
					Column: "d_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.f": {
			Schema: "public",
			Table:  "f",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "f",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "f",
					Column: "e_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.a": {
			WhereClause: &whereId,
		},
		"public.b": {
			WhereClause: &whereName,
		},
		"public.e": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
		"public.c": {
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
		"public.e": {
			{Column: "d_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.d", Column: "id"}},
		},
		"public.f": {
			{Column: "e_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.e", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.d", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.e", DependsOn: []*tabledependency.DependsOn{{Table: "public.d", Columns: []string{"id"}}}},
		{Table: "public.f", DependsOn: []*tabledependency.DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id" FROM "public"."a" WHERE public.a.id = 1;`,
			"public.b": `SELECT "public"."b"."id", "public"."b"."name", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE (public.a.id = 1 AND public.b.name = 'neo');`,
			"public.c": `SELECT "public"."c"."id", "public"."c"."b_id" FROM "public"."c" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE (public.a.id = 1 AND public.b.name = 'neo');`,
			"public.d": `SELECT "id" FROM "public"."d";`,
			"public.e": `SELECT "id", "d_id" FROM "public"."e" WHERE public.e.id = 1;`,
			"public.f": `SELECT "public"."f"."id", "public"."f"."e_id" FROM "public"."f" INNER JOIN "public"."e" ON ("public"."e"."id" = "public"."f"."e_id") WHERE public.e.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_MultipleRoots(t *testing.T) {
	whereId := "id = 1"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.d": {
			Schema: "public",
			Table:  "d",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "d",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "d",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.e": {
			Schema: "public",
			Table:  "e",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "e",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "e",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.b": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.c": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
		"public.d": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
		"public.e": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.d", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		{Table: "public.e", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id" FROM "public"."a";`,
			"public.b": `SELECT "id" FROM "public"."b" WHERE public.b.id = 1;`,
			"public.c": `SELECT "public"."c"."id", "public"."c"."a_id", "public"."c"."b_id" FROM "public"."c" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.id = 1;`,
			"public.d": `SELECT "public"."d"."id", "public"."d"."c_id" FROM "public"."d" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."d"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.id = 1;`,
			"public.e": `SELECT "public"."e"."id", "public"."e"."c_id" FROM "public"."e" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."e"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_MultipleRootsAndWheres(t *testing.T) {
	whereId := "id = 1"
	whereId2 := "id = 2"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "x_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.x": {
			Schema: "public",
			Table:  "x",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "x",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "c",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.d": {
			Schema: "public",
			Table:  "d",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "d",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "d",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.e": {
			Schema: "public",
			Table:  "e",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "e",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "e",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.b": {
			WhereClause: &whereId,
		},
		"public.x": {
			WhereClause: &whereId2,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.c": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
		"public.d": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
		"public.e": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
		"public.a": {
			{Column: "x_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.x", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.x", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{{Table: "public.x", Columns: []string{"id"}}}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.d", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		{Table: "public.e", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id" FROM "public"."a";`,
			"public.b": `SELECT "id" FROM "public"."b" WHERE public.b.id = 1;`,
			"public.c": `SELECT "public"."c"."id", "public"."c"."a_id", "public"."c"."b_id" FROM "public"."c" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.id = 1;`,
			"public.d": `SELECT "public"."d"."id", "public"."d"."c_id" FROM "public"."d" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."d"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.id = 1;`,
			"public.e": `SELECT "public"."e"."id", "public"."e"."c_id" FROM "public"."e" INNER JOIN "public"."c" ON ("public"."c"."id" = "public"."e"."c_id") INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."c"."b_id") WHERE public.b.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	jsonF, _ := json.MarshalIndent(sql, "", " ")
	fmt.Printf("\n sql: %s \n", string(jsonF))
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_DoubleCircularDependencyRoot(t *testing.T) {
	whereId := "id = 1"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "a_a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.a": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.a": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "a_a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", Columns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.a", Columns: []string{"a_id", "aa_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": ` WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" WHERE public.a.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "a_a_id" FROM "related";`,
			"public.b": `SELECT "public"."b"."id", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.a.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_DoubleReference(t *testing.T) {
	whereId := "id = 1"
	mappings := map[string]*tableMapping{
		"public.company": {
			Schema: "public",
			Table:  "company",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "company",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.department": {
			Schema: "public",
			Table:  "department",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "department",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "department",
					Column: "company_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.expense_report": {
			Schema: "public",
			Table:  "expense_report",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "expense_report",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "expense_report",
					Column: "department_source_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "expense_report",
					Column: "department_destination_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.company": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.department": {
			{Column: "company_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.company", Column: "id"}},
		},
		"public.expense_report": {
			{Column: "department_source_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.department", Column: "id"}},
			{Column: "department_destination_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.department", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.company", Columns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.department", Columns: []string{"id", "company_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.company", Columns: []string{"id"}}}},
		{Table: "public.expense_report", Columns: []string{"id", "department_source_id", "department_destination_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.department", Columns: []string{"department_source_id", "department_destination_id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" WHERE public.a.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "a_a_id" FROM "related";`,
			"public.b": `SELECT "public"."b"."id", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.a.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_Other(t *testing.T) {
	whereId := "id = 1"
	mappings := map[string]*tableMapping{
		"public.company": {
			Schema: "public",
			Table:  "company",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "company",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.department": {
			Schema: "public",
			Table:  "department",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "department",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "department",
					Column: "company_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.transaction": {
			Schema: "public",
			Table:  "transaction",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "transaction",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.expense_report": {
			Schema: "public",
			Table:  "expense_report",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "expense_report",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "expense_report",
					Column: "department_source_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "expense_report",
					Column: "transaction_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.company": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.department": {
			{Column: "company_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.company", Column: "id"}},
		},
		"public.expense_report": {
			{Column: "department_source_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.department", Column: "id"}},
			{Column: "transaction_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.transaction", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.company", Columns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.transaction", Columns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.department", Columns: []string{"id", "company_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.company", Columns: []string{"id"}}}},
		{Table: "public.expense_report", Columns: []string{"id", "department_source_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.department", Columns: []string{"department_source_id"}}, {Table: "public.transaction", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" WHERE public.a.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "a_a_id" FROM "related";`,
			"public.b": `SELECT "public"."b"."id", "public"."b"."a_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.a.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_doubleCircularDependencyRoot_mysql(t *testing.T) {
	whereId := "id = 1"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "a_a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.a": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.a": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "a_a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", Columns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.a", Columns: []string{"a_id", "aa_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": "WITH RECURSIVE related AS (SELECT `public`.`a`.`id`, `public`.`a`.`a_id`, `public`.`a`.`a_a_id` FROM `public`.`a` WHERE public.a.id = 1 UNION (SELECT `public`.`a`.`id`, `public`.`a`.`a_id`, `public`.`a`.`a_a_id` FROM `public`.`a` INNER JOIN `related` ON ((`public`.`a`.`id` = `related`.`a_id`) OR (`public`.`a`.`id` = `related`.`a_a_id`)))) SELECT DISTINCT `id`, `a_id`, `a_a_id` FROM `related`;",
			"public.b": "SELECT `public`.`b`.`id`, `public`.`b`.`a_id` FROM `public`.`b` INNER JOIN `public`.`a` ON (`public`.`a`.`id` = `public`.`b`.`a_id`) WHERE public.a.id = 1;",
		}
	sql, err := buildSelectQueryMap("mysql", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_DoubleCircularDependencyChild(t *testing.T) {
	whereId := "id = 1"
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "a_a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "a",
					Column: "b_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.b": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.a": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "a_a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "b_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.b", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", Columns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		{Table: "public.a", Columns: []string{"a_id", "aa_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{}},
	}
	expected :=
		map[string]string{
			"public.a": `WITH RECURSIVE related AS (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id", "public"."a"."b_id" FROM "public"."a" INNER JOIN "public"."b" ON ("public"."b"."id" = "public"."a"."b_id") WHERE public.b.id = 1 UNION (SELECT "public"."a"."id", "public"."a"."a_id", "public"."a"."a_a_id", "public"."a"."b_id" FROM "public"."a" INNER JOIN "related" ON (("public"."a"."id" = "related"."a_id") OR ("public"."a"."id" = "related"."a_a_id")))) SELECT DISTINCT "id", "a_id", "a_a_id", "b_id" FROM "related";`,
			"public.b": `SELECT "id" FROM "public"."b" WHERE public.b.id = 1;`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)
	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_buildSelectQueryMap_shouldContinue(t *testing.T) {
	mappings := map[string]*tableMapping{
		"public.a": {
			Schema: "public",
			Table:  "a",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "a",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.b": {
			Schema: "public",
			Table:  "b",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "b",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "a_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "b",
					Column: "d_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.c": {
			Schema: "public",
			Table:  "c",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "c",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.d": {
			Schema: "public",
			Table:  "d",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "d",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "d",
					Column: "c_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
		"public.e": {
			Schema: "public",
			Table:  "e",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "e",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
				{
					Schema: "public",
					Table:  "e",
					Column: "d_id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					},
				},
			},
		},
	}
	aWhere := "id = 1"
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.a": {WhereClause: &aWhere},
	}
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.b": {
			{Column: "a_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.a", Column: "id"}},
			{Column: "d_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.d", Column: "id"}},
		},
		"public.d": {
			{Column: "c_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.c", Column: "id"}},
		},
		"public.e": {
			{Column: "d_id", IsNullable: false, ForeignKey: &sql_manager.ForeignKey{Table: "public.d", Column: "id"}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "public.a", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
		{Table: "public.c", DependsOn: []*tabledependency.DependsOn{}},
		{Table: "public.d", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		{Table: "public.e", DependsOn: []*tabledependency.DependsOn{{Table: "public.d", Columns: []string{"id"}}}},
	}
	expected :=
		map[string]string{
			"public.a": `SELECT "id" FROM "public"."a" WHERE public.a.id = 1;`,
			"public.b": `SELECT "public"."b"."id", "public"."b"."a_id", "public"."b"."d_id" FROM "public"."b" INNER JOIN "public"."a" ON ("public"."a"."id" = "public"."b"."a_id") WHERE public.a.id = 1;`,
			"public.c": `SELECT "id" FROM "public"."c";`,
			"public.d": `SELECT "id", "c_id" FROM "public"."d";`,
			"public.e": `SELECT "id", "d_id" FROM "public"."e";`,
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, dependencyConfigs, true)

	require.NoError(t, err)
	require.Equal(t, expected, sql)
}

func Test_getBfsPathMap(t *testing.T) {
	tests := []struct {
		name     string
		graph    map[string][]string
		start    string
		expected *bfsPaths
	}{
		{
			name: "straight path",
			graph: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {"d"},
				"d": {},
			},
			start: "a",
			expected: &bfsPaths{
				Path: []string{"a", "b", "c", "d"},
				NodePathMap: map[string][]string{
					"a": {"a"},
					"b": {"a", "b"},
					"c": {"a", "b", "c"},
					"d": {"a", "b", "c", "d"},
				},
			},
		},
		{
			name: "multiple paths",
			graph: map[string][]string{
				"a": {"c", "b"},
				"b": {"c"},
			},
			start: "a",
			expected: &bfsPaths{
				Path: []string{"a", "c", "b"},
				NodePathMap: map[string][]string{
					"a": {"a"},
					"b": {"a", "b"},
					"c": {"a", "c"},
				},
			},
		},
		{
			name: "cycle",
			graph: map[string][]string{
				"c": {"a"},
				"b": {"c"},
				"a": {"b"},
			},
			start: "a",
			expected: &bfsPaths{
				Path: []string{"a", "b", "c"},
				NodePathMap: map[string][]string{
					"a": {"a"},
					"b": {"a", "b"},
					"c": {"a", "b", "c"},
				},
			},
		},
		{
			name: "cross",
			graph: map[string][]string{
				"a": {"c"},
				"b": {"c"},
				"c": {"d", "e"},
				"d": {},
				"e": {},
			},
			start: "a",
			expected: &bfsPaths{
				Path: []string{"a", "c", "d", "e"},
				NodePathMap: map[string][]string{
					"a": {"a"},
					"c": {"a", "c"},
					"d": {"a", "c", "d"},
					"e": {"a", "c", "e"},
				},
			},
		},
		{
			name: "self reference",
			graph: map[string][]string{
				"a": {"a"},
			},
			start: "a",
			expected: &bfsPaths{
				Path: []string{"a"},
				NodePathMap: map[string][]string{
					"a": {"a"},
				},
			},
		},
		{
			name: "multi linear",
			graph: map[string][]string{
				"a": {"b", "c", "d"},
				"b": {"e"},
				"c": {"f"},
				"d": {"g"},
			},
			start: "a",
			expected: &bfsPaths{
				Path: []string{"a", "b", "c", "d", "e", "f", "g"},
				NodePathMap: map[string][]string{
					"a": {"a"},
					"b": {"a", "b"},
					"c": {"a", "c"},
					"d": {"a", "d"},
					"e": {"a", "b", "e"},
					"f": {"a", "c", "f"},
					"g": {"a", "d", "g"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			path := getBfsPathMap(tt.graph, tt.start)
			require.Equal(t, tt.expected, path)
		})
	}
}

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
			schema:   "PublicSchema",
			table:    "order",
			expected: "PublicSchema.`order`.`name with space` = 'hey' and PublicSchema.`order`.PascalCase = 'other'",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qualifyWhereColumnNames(sql_manager.MysqlDriver, tt.where, tt.schema, tt.table)
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
			schema:   "PublicSchema",
			table:    "order",
			expected: `"PublicSchema"."order"."bad name" = 'alisha'`,
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
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			response, err := qualifyWhereColumnNames(sql_manager.PostgresDriver, tt.where, tt.schema, tt.table)
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
			driver:   sql_manager.PostgresDriver,
			name:     "simple",
			where:    "name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sql_manager.PostgresDriver,
			name:     "hash alias",
			where:    "composite_keys.department.department_id = '1'",
			alias:    "50d89c0f3af602",
			expected: `"50d89c0f3af602".department_id = '1'`,
		},
		{
			driver:   sql_manager.PostgresDriver,
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sql_manager.PostgresDriver,
			name:     "multiple",
			where:    "name = 'alisha' and id = 1  or age = 2",
			alias:    "alias",
			expected: `(alias.name = 'alisha' AND alias.id = 1) OR alias.age = 2`,
		},
		{
			driver:   sql_manager.MysqlDriver,
			name:     "simple",
			where:    "name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sql_manager.MysqlDriver,
			name:     "simple",
			where:    "public.a.name = 'alisha'",
			alias:    "alias",
			expected: `alias.name = 'alisha'`,
		},
		{
			driver:   sql_manager.MysqlDriver,
			name:     "multiple",
			where:    "name = 'alisha' and id = 1  or age = 2",
			alias:    "alias",
			expected: `alias.name = 'alisha' and alias.id = 1 or alias.age = 2`,
		},
		{
			driver:   sql_manager.MysqlDriver,
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
