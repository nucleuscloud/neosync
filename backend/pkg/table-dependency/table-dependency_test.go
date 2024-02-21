package tabledependency

import (
	"sort"
	"testing"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/stretchr/testify/assert"
)

func Test_FindCircularDependencies(t *testing.T) {
	tests := []struct {
		name         string
		dependencies map[string][]string
		expect       [][]string
	}{
		{
			name: "No circular dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.c": {"public.d"},
				"public.d": {"public.e"},
			},
			expect: nil,
		},
		{
			name: "Self circular dependency",
			dependencies: map[string][]string{
				"public.a": {"public.a"},
				"public.b": {},
			},
			expect: [][]string{{"public.a"}},
		},
		{
			name: "Simple circular dependency",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.b": {"public.a"},
			},
			expect: [][]string{{"public.a", "public.b"}},
		},
		{
			name: "Multiple circular dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.b": {"public.c"},
				"public.c": {"public.a"},
				"public.d": {"public.e"},
				"public.e": {"public.d"},
			},
			expect: [][]string{{"public.a", "public.b", "public.c"}, {"public.d", "public.e"}},
		},
		{
			name: "Multiple connected circular dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.b": {"public.c", "public.d"},
				"public.c": {"public.a"},
				"public.d": {"public.e"},
				"public.e": {"public.b"},
			},
			expect: [][]string{{"public.a", "public.b", "public.c"}, {"public.b", "public.d", "public.e"}},
		},
		{
			name: "Both circular dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.b", "public.a"},
				"public.b": {"public.a"},
			},
			expect: [][]string{{"public.a", "public.b"}, {"public.a"}},
		},
		{
			name: "Three circular dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.b": {"public.c"},
				"public.c": {"public.a"},
			},
			expect: [][]string{{"public.a", "public.b", "public.c"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := findCircularDependencies(tt.dependencies)

			for i := range actual {
				sort.Strings(actual[i])
			}
			for i := range tt.expect {
				sort.Strings(tt.expect[i])
			}

			assert.Len(t, tt.expect, len(actual))
			assert.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func Test_GetRunConfigs(t *testing.T) {
	tests := []struct {
		name         string
		dependencies dbschemas.TableDependency
		tables       []string
		expect       []*RunConfig
	}{
		{
			name: "No circular dependencies",
			dependencies: dbschemas.TableDependency{
				"public.countries": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "region_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.regions", Column: "id"}},
					},
				},
				"public.departments": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "location_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.locations", Column: "id"}},
					},
				},
				"public.employees": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "department_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.departments", Column: "id"}},
						{Column: "job_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.jobs", Column: "id"}},
					},
				},
				"public.locations": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "country_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.countries", Column: "id"}},
					},
				},
			},
			tables: []string{"public.jobs", "public.locations", "public.regions", "public.departments", "public.countries", "public.employees"},
			expect: []*RunConfig{
				{Table: "public.regions", DependsOn: []*DependsOn{}},
				{Table: "public.locations", DependsOn: []*DependsOn{{Table: "public.countries", Columns: []string{"id"}}}},
				{Table: "public.employees", DependsOn: []*DependsOn{{Table: "public.departments", Columns: []string{"id"}}, {Table: "public.jobs", Columns: []string{"id"}}}},
				{Table: "public.departments", DependsOn: []*DependsOn{{Table: "public.locations", Columns: []string{"id"}}}},
				{Table: "public.countries", DependsOn: []*DependsOn{{Table: "public.regions", Columns: []string{"id"}}}},
				{Table: "public.jobs", DependsOn: []*DependsOn{}},
			},
		},
		{
			name: "Self Circular Dependency",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a"},
			expect: []*RunConfig{
				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"a_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"a_id"}}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Double Self Circular Dependency",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
						{Column: "aa_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a"},
			expect: []*RunConfig{
				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"a_id", "aa_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"a_id", "aa_id"}}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Two Table Circular Dependency",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b"},
			expect: []*RunConfig{
				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Three Table Circular Dependency",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			expect: []*RunConfig{
				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Multi Table Dependencies",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "d_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
				"public.d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
					},
				},
				"public.e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c", "public.d", "public.e"},
			expect: []*RunConfig{
				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.e", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.d", DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Exclude With Dependency",
			dependencies: dbschemas.TableDependency{
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			expect: []*RunConfig{
				{Table: "public.a", DependsOn: []*DependsOn{}},
				{Table: "public.b", Columns: &SyncColumn{Exclude: []string{"c_id"}}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", Columns: &SyncColumn{Include: []string{"c_id"}}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.a", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Multiple Depends On",
			dependencies: dbschemas.TableDependency{
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			expect: []*RunConfig{
				{Table: "public.a", DependsOn: []*DependsOn{}},
				{Table: "public.c", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.c", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Multi Unconnected Circular Dependencies",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
				"public.d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "e_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
					},
				},
				"public.e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "d_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c", "public.d", "public.e"},
			expect: []*RunConfig{

				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},

				{Table: "public.d", Columns: &SyncColumn{Exclude: []string{"e_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.e", DependsOn: []*DependsOn{{Table: "public.d", Columns: []string{"id"}}}},
				{Table: "public.d", Columns: &SyncColumn{Include: []string{"e_id"}}, DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Subset of tables",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.b", "public.c"},
			expect: []*RunConfig{
				{Table: "public.c", DependsOn: []*DependsOn{}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Two Table Circular Dependency None Nullable",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b"},
			expect: []*RunConfig{
				{Table: "public.a", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetRunConfigs(tt.dependencies, tt.tables)
			assert.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func Test_GetRunConfigs_MultipleExclude(t *testing.T) {
	dependencies := dbschemas.TableDependency{
		"public.a": &dbschemas.TableConstraints{
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
			},
		},
		"public.b": &dbschemas.TableConstraints{
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
				{Column: "d_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
			},
		},
		"public.c": &dbschemas.TableConstraints{
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
			},
		},
		"public.d": &dbschemas.TableConstraints{
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
			},
		},
		"public.e": &dbschemas.TableConstraints{
			Constraints: []*dbschemas.ForeignConstraint{
				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
			},
		},
	}
	tables := []string{"public.a", "public.b", "public.c", "public.d", "public.e"}

	actual := GetRunConfigs(dependencies, tables)
	expect := map[string]string{
		"public.e": "public.b",
		"public.a": "public.b",
		"public.c": "public.a",
		"public.d": "public.e",
	}

	assert.Len(t, actual, 6)
	for _, a := range actual {
		if a.Columns != nil && a.Columns.Exclude != nil {
			assert.Len(t, a.DependsOn, 0)
			assert.ElementsMatch(t, []string{"c_id", "d_id"}, a.Columns.Exclude)
		} else if a.Columns != nil && a.Columns.Include != nil {
			assert.ElementsMatch(t, []string{"c_id", "d_id"}, a.Columns.Include)
			assert.ElementsMatch(t, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}, a.DependsOn)
		} else {
			assert.Len(t, a.DependsOn, 1)
			assert.Equal(t, expect[a.Table], a.DependsOn[0].Table)
		}
	}
}

func Test_GetTablesOrderedByDependency_CircularDependency(t *testing.T) {
	dependencies := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}

	actual := GetTablesOrderedByDependency(dependencies)
	assert.Nil(t, actual)
}

func Test_GetTablesOrderedByDependency_Dependencies(t *testing.T) {
	dependencies := map[string][]string{
		"countries":   {"regions"},
		"departments": {"locations"},
		"dependents":  {"employees"},
		"employees":   {"departments", "jobs", "employees"},
		"locations":   {"countries"},
		"regions":     {},
		"jobs":        {},
	}
	expected := [][]string{{"dependents", "employees", "departments", "locations", "countries", "regions", "jobs"}}

	actual := GetTablesOrderedByDependency(dependencies)
	assert.Len(t, actual, 1)
	for idx, table := range actual {
		assert.Equal(t, expected[idx], table)
	}
}

func Test_GetTablesOrderedByDependency_Mixed(t *testing.T) {
	dependencies := map[string][]string{
		"countries": {},
		"locations": {"countries"},
		"regions":   {},
		"jobs":      {},
	}

	expected := [][]string{{"jobs"}, {"locations", "countries"}, {"regions"}}
	actual := GetTablesOrderedByDependency(dependencies)
	assert.Len(t, actual, len(expected))
	for _, group := range expected {
		assert.Contains(t, actual, group)
	}
}

func Test_GetTablesOrderedByDependency_BrokenDependencies_NoLoop(t *testing.T) {
	dependencies := map[string][]string{
		"countries": {},
		"locations": {"countries"},
		"regions":   {"a"},
		"jobs":      {"b"},
	}

	actual := GetTablesOrderedByDependency(dependencies)
	assert.Len(t, actual, 3)
}

func Test_GetTablesOrderedByDependency_NestedDependencies(t *testing.T) {
	dependencies := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"d"},
		"d": {},
	}

	expected := [][]string{{"a", "b", "c", "d"}}
	actual := GetTablesOrderedByDependency(dependencies)
	assert.Len(t, actual, 1)
	assert.Equal(t, expected[0], actual[0])
}
