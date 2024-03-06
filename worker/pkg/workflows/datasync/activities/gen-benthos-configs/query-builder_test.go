package genbenthosconfigs_activity

import (
	"fmt"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
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

func Test_buildSelectQueryMap(t *testing.T) {
	whereId := "id = 1"
	tests := []struct {
		name                          string
		driver                        string
		subsetByForeignKeyConstraints bool
		mappings                      map[string]*tableMapping
		sourceTableOpts               map[string]*sqlSourceTableOptions
		tableDependencies             map[string]*dbschemas.TableConstraints
		expected                      map[string]string
	}{
		{
			name:                          "select no subset",
			driver:                        "postgres",
			subsetByForeignKeyConstraints: false,
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
								Source: "generate_default",
							},
						},
						{
							Schema: "public",
							Table:  "users",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: "",
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
								Source: "generate_default",
							},
						},
						{
							Schema: "public",
							Table:  "accounts",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: "",
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
			tableDependencies: map[string]*dbschemas_utils.TableConstraints{},
			expected: map[string]string{
				"public.users":    `SELECT "id", "name" FROM "public"."users" WHERE id = 1;`,
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
								Source: "generate_default",
							},
						},
						{
							Schema: "public",
							Table:  "users",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: "",
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
								Source: "generate_default",
							},
						},
						{
							Schema: "public",
							Table:  "accounts",
							Column: "name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: "",
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
			tableDependencies: map[string]*dbschemas_utils.TableConstraints{},
			expected: map[string]string{
				"public.users":    `SELECT "id", "name" FROM "public"."users" WHERE id = 1;`,
				"public.accounts": `SELECT "id", "name" FROM "public"."accounts";`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			sql, err := buildSelectQueryMap(tt.driver, tt.mappings, tt.sourceTableOpts, tt.tableDependencies, tt.subsetByForeignKeyConstraints)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, sql)
		})
	}
}

func Test_buildSelectQueryMap_PrimaryKeySubset(t *testing.T) {
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
						Source: "generate_default",
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
						Source: "generate_default",
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
						Source: "generate_default",
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
						Source: "generate_default",
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
						Source: "generate_default",
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
						Source: "generate_default",
					},
				},
			},
		},
	}
	sourceTableOpts := map[string]*sqlSourceTableOptions{
		"public.a": {
			WhereClause: &whereId,
		},
		"public.e": {
			WhereClause: &whereId,
		},
	}
	tableDependencies := map[string]*dbschemas.TableConstraints{
		"public.b": {
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
			},
		},
		"public.c": {
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
			},
		},
		"public.e": {
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "d_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
			},
		},
		"public.f": {
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
			},
		},
	}
	expected :=
		map[string]string{
			"public.a": "",
			"public.b": "",
			"public.c": "",
		}
	sql, err := buildSelectQueryMap("postgres", mappings, sourceTableOpts, tableDependencies, true)
	assert.NoError(t, err)
	assert.Equal(t, expected, sql)
}

func Test_getPrimaryToForeignTableMap(t *testing.T) {
	tables := map[string]struct{}{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {},
		"public.locations":   {},
		"public.dependents":  {},
		"public.departments": {},
		"public.employees":   {},
	}
	dependencies := map[string]*dbschemas_utils.TableConstraints{
		"public.countries": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "region_id", IsNullable: false, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.regions", Column: "region_id"}},
		}},
		"public.departments": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "location_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.locations", Column: "location_id"}},
		}},
		"public.dependents": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "dependent_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.employees", Column: "employees_id"}},
		}},
		"public.locations": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "country_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.countries", Column: "country_id"}},
		}},
		"public.employees": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "department_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.departments", Column: "department_id"}},
			{Column: "job_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.jobs", Column: "job_id"}},
			{Column: "manager_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.employees", Column: "employee_id"}},
		}},
	}

	expected := map[string][]string{
		"public.dependents":  {},
		"public.countries":   {"public.locations"},
		"public.departments": {"public.employees"},
		"public.employees":   {"public.dependents", "public.employees"},
		"public.jobs":        {"public.employees"},
		"public.locations":   {"public.departments"},
		"public.regions":     {"public.countries"},
	}
	actual := getPrimaryToForeignTableMap(dependencies, tables)
	assert.Len(t, actual, len(expected))
	for table, deps := range actual {
		assert.Len(t, deps, len(expected[table]))
		assert.ElementsMatch(t, expected[table], deps)
	}
}

func Test_BFS(t *testing.T) {
	tests := []struct {
		name     string
		graph    map[string][]string
		start    string
		expected []string
	}{
		{
			name: "straight path",
			graph: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {"d"},
				"d": {},
			},
			start:    "a",
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name: "multiple paths",
			graph: map[string][]string{
				"a": {"c", "b"},
				"b": {"c"},
			},
			start:    "a",
			expected: []string{"a", "c", "b"},
		},
		{
			name: "cycle",
			graph: map[string][]string{
				"c": {"a"},
				"b": {"c"},
				"a": {"b"},
			},
			start:    "a",
			expected: []string{"a", "b", "c"},
		},
		{
			name: "cross",
			graph: map[string][]string{
				"a": {"c"},
				"b": {"b"},
				"c": {"d", "e"},
				"d": {},
				"e": {},
			},
			start:    "a",
			expected: []string{"a", "c", "d", "e"},
		},
		{
			name: " self reference",
			graph: map[string][]string{
				"a": {"a"},
			},
			start:    "a",
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), tt.name), func(t *testing.T) {
			path := BFS(tt.graph, tt.start)
			assert.Equal(t, tt.expected, path)
		})
	}
}
