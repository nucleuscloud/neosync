package tabledependency

import (
	"encoding/json"
	"fmt"
	"testing"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/stretchr/testify/assert"
)

// func Test_FindCircularDependencies(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		dependencies map[string][]string
// 		expect       [][]string
// 	}{
// 		{
// 			name: "No circular dependencies",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.b"},
// 				"public.c": {"public.d"},
// 				"public.d": {"public.e"},
// 			},
// 			expect: nil,
// 		},
// 		{
// 			name: "Self circular dependency",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.a"},
// 				"public.b": {},
// 			},
// 			expect: [][]string{{"public.a"}},
// 		},
// 		{
// 			name: "Simple circular dependency",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.b"},
// 				"public.b": {"public.a"},
// 			},
// 			expect: [][]string{{"public.a", "public.b"}},
// 		},
// 		{
// 			name: "Multiple circular dependencies",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.b"},
// 				"public.b": {"public.c"},
// 				"public.c": {"public.a"},
// 				"public.d": {"public.e"},
// 				"public.e": {"public.d"},
// 			},
// 			expect: [][]string{{"public.a", "public.b", "public.c"}, {"public.d", "public.e"}},
// 		},
// 		{
// 			name: "Multiple connected circular dependencies",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.b"},
// 				"public.b": {"public.c", "public.d"},
// 				"public.c": {"public.a"},
// 				"public.d": {"public.e"},
// 				"public.e": {"public.b"},
// 			},
// 			expect: [][]string{{"public.a", "public.b", "public.c"}, {"public.b", "public.d", "public.e"}},
// 		},
// 		{
// 			name: "Both circular dependencies",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.b", "public.a"},
// 				"public.b": {"public.a"},
// 			},
// 			expect: [][]string{{"public.a", "public.b"}, {"public.a"}},
// 		},
// 		{
// 			name: "Three circular dependencies",
// 			dependencies: map[string][]string{
// 				"public.a": {"public.b"},
// 				"public.b": {"public.c"},
// 				"public.c": {"public.a"},
// 			},
// 			expect: [][]string{{"public.a", "public.b", "public.c"}},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actual := findCircularDependencies(tt.dependencies)

// 			for i := range actual {
// 				sort.Strings(actual[i])
// 			}
// 			for i := range tt.expect {
// 				sort.Strings(tt.expect[i])
// 			}

// 			assert.Len(t, tt.expect, len(actual))
// 			assert.ElementsMatch(t, tt.expect, actual)
// 		})
// 	}
// }

func Test_GetRunConfigs_NoSubset(t *testing.T) {

	where := ""
	tests := []struct {
		name          string
		dependencies  dbschemas.TableDependency
		tables        []string
		subsets       map[string]string
		tableColsMap  map[string][]string
		primaryKeyMap map[string][]string
		expect        []*RunConfig
	}{
		// {
		// 	name: "Multi Table Dependencies",
		// 	dependencies: dbschemas.TableDependency{
		// 		"public.a": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
		// 			},
		// 		},
		// 		"public.b": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
		// 				{Column: "d_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
		// 			},
		// 		},
		// 		"public.c": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
		// 			},
		// 		},
		// 		"public.d": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
		// 			},
		// 		},
		// 		"public.e": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
		// 			},
		// 		},
		// 	},
		// 	tables:  []string{"public.a", "public.b", "public.c", "public.d", "public.e"},
		// 	subsets: map[string]string{},
		// 	expect: []*RunConfig{
		// 		{Table: "public.a", DependsOn: []*DependsOn{}},
		// 		{Table: "public.b", DependsOn: []*DependsOn{}},
		// 		{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		// 		{Table: "public.e", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		// 		{Table: "public.d", DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
		// 		{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
		// 		{Table: "public.a", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
		// 	},
		// },
		// {
		// 	name: "Single Cycle",
		// 	dependencies: dbschemas.TableDependency{
		// 		"public.a": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
		// 			},
		// 		},
		// 		"public.b": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
		// 			},
		// 		},
		// 		"public.c": &dbschemas.TableConstraints{
		// 			Constraints: []*dbschemas.ForeignConstraint{
		// 				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
		// 			},
		// 		},
		// 	},
		// 	tables: []string{"public.a", "public.b", "public.c"},
		// 	tableColsMap: map[string][]string{
		// 		"public.a": {"id", "b_id"},
		// 		"public.b": {"id", "c_id"},
		// 		"public.c": {"id", "a_id"},
		// 	},
		// 	primaryKeyMap: map[string][]string{
		// 		"public.a": {"id"},
		// 		"public.b": {"id"},
		// 		"public.c": {"id"},
		// 	},
		// 	subsets: map[string]string{},
		// 	expect: []*RunConfig{
		// 		{Table: "public.a", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
		// 		{Table: "public.a", RunType: Update, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
		// 		{Table: "public.c", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
		// 		{Table: "public.b", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "c_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
		// 	},
		// },
		{
			name: "Single Cycle",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
						{Column: "x_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.x", Column: "id"}},
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
			tables: []string{"public.a", "public.b", "public.c", "public.x"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id", "x_id"},
				"public.b": {"id", "c_id"},
				"public.c": {"id", "a_id"},
				"public.x": {"id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.x", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{{Table: "public.x", Columns: []string{"id"}}}},
				{Table: "public.a", RunType: Update, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: Insert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "c_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.tables, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			jsonF, _ := json.MarshalIndent(actual, "", " ")
			fmt.Printf("\n actual: %s \n", string(jsonF))
			jsonF, _ = json.MarshalIndent(tt.expect, "", " ")
			fmt.Printf("\n expect: %s \n", string(jsonF))
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expect, actual)
		})
	}
}

// func Test_GetRunConfigs_CircularDependencyNoneNullable(t *testing.T) {
// 	dependencies := dbschemas.TableDependency{
// 		"public.a": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 			},
// 		},
// 		"public.b": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 			},
// 		},
// 	}
// 	tables := []string{"public.a", "public.b"}
// 	_, err := GetRunConfigs(dependencies, tables, map[string]string{})
// 	assert.Error(t, err)
// }

// func Test_GetRunConfigs_CircularDependencyAllNullable(t *testing.T) {
// 	dependencies := dbschemas.TableDependency{
// 		"public.c": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 			},
// 		},
// 		"public.a": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 			},
// 		},
// 		"public.b": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
// 			},
// 		},
// 	}
// 	tables := []string{"public.a", "public.b", "public.c"}
// 	expect := []*RunConfig{
// 		{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
// 		{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
// 		{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
// 		{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
// 	}

// 	actual, err := GetRunConfigs(dependencies, tables, map[string]string{})
// 	assert.NoError(t, err)
// 	assert.ElementsMatch(t, expect, actual)
// }

// func Test_GetRunConfigs_MultiCircularDependency(t *testing.T) {
// 	dependencies := dbschemas.TableDependency{
// 		"public.a": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 			},
// 		},
// 		"public.b": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
// 				{Column: "d_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
// 			},
// 		},
// 		"public.c": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 			},
// 		},
// 		"public.d": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
// 			},
// 		},
// 		"public.e": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 			},
// 		},
// 	}
// 	tables := []string{"public.a", "public.b", "public.c", "public.d", "public.e"}
// 	expect := []*RunConfig{
// 		{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
// 		{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
// 		{Table: "public.e", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
// 		{Table: "public.d", DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
// 		{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
// 		{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
// 	}

// 	actual, err := GetRunConfigs(dependencies, tables, map[string]string{})
// 	assert.NoError(t, err)
// 	assert.ElementsMatch(t, expect, actual)
// }

// func Test_GetRunConfigs_MultipleExclude(t *testing.T) {
// 	dependencies := dbschemas.TableDependency{
// 		"public.a": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 			},
// 		},
// 		"public.b": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
// 				{Column: "d_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
// 			},
// 		},
// 		"public.c": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 			},
// 		},
// 		"public.d": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
// 			},
// 		},
// 		"public.e": &dbschemas.TableConstraints{
// 			Constraints: []*dbschemas.ForeignConstraint{
// 				{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 			},
// 		},
// 	}
// 	tables := []string{"public.a", "public.b", "public.c", "public.d", "public.e"}

// 	actual, err := GetRunConfigs(dependencies, tables, map[string]string{})
// 	assert.NoError(t, err)
// 	expect := map[string]string{
// 		"public.e": "public.b",
// 		"public.a": "public.b",
// 		"public.c": "public.a",
// 		"public.d": "public.e",
// 	}

// 	assert.Len(t, actual, 6)
// 	for _, a := range actual {
// 		if a.Columns != nil && a.Columns.Exclude != nil {
// 			assert.Len(t, a.DependsOn, 0)
// 			assert.Equal(t, "public.b", a.Table)
// 			assert.ElementsMatch(t, []string{"c_id", "d_id"}, a.Columns.Exclude)
// 		} else if a.Columns != nil && a.Columns.Include != nil {
// 			assert.Equal(t, "public.b", a.Table)
// 			assert.ElementsMatch(t, []string{"c_id", "d_id"}, a.Columns.Include)
// 			assert.ElementsMatch(t, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}, a.DependsOn)
// 		} else {
// 			assert.Len(t, a.DependsOn, 1)
// 			assert.Equal(t, expect[a.Table], a.DependsOn[0].Table)
// 		}
// 	}
// }

// func Test_GetRunConfigs_Subset(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		dependencies dbschemas.TableDependency
// 		tables       []string
// 		subsets      map[string]string
// 		expect       []*RunConfig
// 	}{
// 		{
// 			name: "No circular dependencies",
// 			dependencies: dbschemas.TableDependency{
// 				"public.countries": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "region_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.regions", Column: "id"}},
// 					},
// 				},
// 				"public.departments": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "location_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.locations", Column: "id"}},
// 					},
// 				},
// 				"public.employees": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "department_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.departments", Column: "id"}},
// 						{Column: "job_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.jobs", Column: "id"}},
// 					},
// 				},
// 				"public.locations": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "country_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.countries", Column: "id"}},
// 					},
// 				},
// 			},
// 			tables: []string{"public.jobs", "public.locations", "public.regions", "public.departments", "public.countries", "public.employees"},
// 			subsets: map[string]string{
// 				"public.jobs":      "id = 1",
// 				"public.locations": "id = 1",
// 			},
// 			expect: []*RunConfig{
// 				{Table: "public.regions", DependsOn: []*DependsOn{}},
// 				{Table: "public.locations", DependsOn: []*DependsOn{{Table: "public.countries", Columns: []string{"id"}}}},
// 				{Table: "public.employees", DependsOn: []*DependsOn{{Table: "public.departments", Columns: []string{"id"}}, {Table: "public.jobs", Columns: []string{"id"}}}},
// 				{Table: "public.departments", DependsOn: []*DependsOn{{Table: "public.locations", Columns: []string{"id"}}}},
// 				{Table: "public.countries", DependsOn: []*DependsOn{{Table: "public.regions", Columns: []string{"id"}}}},
// 				{Table: "public.jobs", DependsOn: []*DependsOn{}},
// 			},
// 		},
// 		{
// 			name: "Self Circular Dependency",
// 			dependencies: dbschemas.TableDependency{
// 				"public.a": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 					},
// 				},
// 			},
// 			tables:  []string{"public.a"},
// 			subsets: map[string]string{"public.a": "id = 1"},
// 			expect: []*RunConfig{
// 				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"a_id"}}, DependsOn: []*DependsOn{}},
// 				{Table: "public.a", Columns: &SyncColumn{Include: []string{"a_id"}}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
// 			},
// 		},
// 		{
// 			name: "Double Self Circular Dependency",
// 			dependencies: dbschemas.TableDependency{
// 				"public.a": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 						{Column: "aa_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 					},
// 				},
// 			},
// 			tables:  []string{"public.a"},
// 			subsets: map[string]string{"public.a": "id = 1"},
// 			expect: []*RunConfig{
// 				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"a_id", "aa_id"}}, DependsOn: []*DependsOn{}},
// 				{Table: "public.a", Columns: &SyncColumn{Include: []string{"a_id", "aa_id"}}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
// 			},
// 		},
// 		{
// 			name: "Two Table Circular Dependency",
// 			dependencies: dbschemas.TableDependency{
// 				"public.a": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 					},
// 				},
// 				"public.b": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 					},
// 				},
// 			},
// 			tables:  []string{"public.a", "public.b"},
// 			subsets: map[string]string{"public.a": "id = 1"},
// 			expect: []*RunConfig{
// 				{Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
// 				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
// 				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
// 			},
// 		},
// 		{
// 			name: "Three Table Circular Dependency",
// 			dependencies: dbschemas.TableDependency{
// 				"public.a": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
// 					},
// 				},
// 				"public.b": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
// 					},
// 				},
// 				"public.c": &dbschemas.TableConstraints{
// 					Constraints: []*dbschemas.ForeignConstraint{
// 						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
// 					},
// 				},
// 			},
// 			tables:  []string{"public.a", "public.b", "public.c"},
// 			subsets: map[string]string{"public.c": "id = 1"},
// 			expect: []*RunConfig{
// 				{Table: "public.c", Columns: &SyncColumn{Exclude: []string{"a_id"}}, DependsOn: []*DependsOn{}},
// 				{Table: "public.a", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
// 				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
// 				{Table: "public.c", Columns: &SyncColumn{Include: []string{"a_id"}}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actual, err := GetRunConfigs(tt.dependencies, tt.tables, tt.subsets)
// 			assert.NoError(t, err)
// 			assert.ElementsMatch(t, tt.expect, actual)
// 		})
// 	}
// }

// func Test_determineStartTable(t *testing.T) {
// 	tests := []struct {
// 		name               string
// 		tablesWithNullable []string
// 		tablesWithSubsets  []string
// 		cycle              []string
// 		expect             string
// 	}{
// 		{
// 			name:               "One table with nullable columns",
// 			tablesWithNullable: []string{"public.b"},
// 			tablesWithSubsets:  []string{"public.a", "public.b"},
// 			cycle:              []string{"public.a", "public.b"},
// 			expect:             "public.b",
// 		},
// 		{
// 			name:               "All tables have nullable columns and subsets",
// 			tablesWithNullable: []string{"public.a", "public.b"},
// 			tablesWithSubsets:  []string{"public.a", "public.b"},
// 			cycle:              []string{"public.a", "public.b"},
// 			expect:             "public.a",
// 		},
// 		{
// 			name:               "One table with nullable columns and subset",
// 			tablesWithNullable: []string{"public.a", "public.c"},
// 			tablesWithSubsets:  []string{"public.c", "public.b"},
// 			cycle:              []string{"public.a", "public.b", "public.c"},
// 			expect:             "public.c",
// 		},
// 		{
// 			name:               "Many tables with nullable columns no overlap with subset tables",
// 			tablesWithNullable: []string{"public.a", "public.c"},
// 			tablesWithSubsets:  []string{"public.b"},
// 			cycle:              []string{"public.a", "public.b", "public.c"},
// 			expect:             "public.a",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actual := determineStartTable(tt.tablesWithNullable, tt.tablesWithSubsets, tt.cycle)
// 			assert.Equal(t, tt.expect, actual)
// 		})
// 	}
// }

// func Test_GetTablesOrderedByDependency_CircularDependency(t *testing.T) {
// 	dependencies := map[string][]string{
// 		"a": {"b"},
// 		"b": {"c"},
// 		"c": {"a"},
// 	}

// 	resp, err := GetTablesOrderedByDependency(dependencies)
// 	assert.NoError(t, err)
// 	assert.Equal(t, resp.HasCycles, true)
// 	for _, e := range resp.OrderedTables {
// 		assert.Contains(t, []string{"a", "b", "c"}, e)
// 	}
// }

// func Test_GetTablesOrderedByDependency_Dependencies(t *testing.T) {
// 	dependencies := map[string][]string{
// 		"countries":   {"regions"},
// 		"departments": {"locations"},
// 		"dependents":  {"employees"},
// 		"employees":   {"departments", "jobs", "employees"},
// 		"locations":   {"countries"},
// 		"regions":     {},
// 		"jobs":        {},
// 	}
// 	expected := [][]string{{"regions", "jobs"}, {"regions", "jobs"}, {"countries"}, {"locations"}, {"departments"}, {"employees"}, {"dependents"}}

// 	actual, err := GetTablesOrderedByDependency(dependencies)
// 	assert.NoError(t, err)
// 	assert.Equal(t, actual.HasCycles, false)

// 	for idx, table := range actual.OrderedTables {
// 		assert.Contains(t, expected[idx], table)
// 	}
// }

// func Test_GetTablesOrderedByDependency_Mixed(t *testing.T) {
// 	dependencies := map[string][]string{
// 		"countries": {},
// 		"locations": {"countries"},
// 		"regions":   {},
// 		"jobs":      {},
// 	}

// 	expected := []string{"countries", "regions", "jobs", "locations"}
// 	actual, err := GetTablesOrderedByDependency(dependencies)
// 	assert.NoError(t, err)
// 	assert.Equal(t, actual.HasCycles, false)
// 	assert.Len(t, actual.OrderedTables, len(expected))
// 	for _, table := range actual.OrderedTables {
// 		assert.Contains(t, expected, table)
// 	}
// 	assert.Equal(t, "locations", actual.OrderedTables[len(actual.OrderedTables)-1])
// }

// func Test_GetTablesOrderedByDependency_BrokenDependencies_NoLoop(t *testing.T) {
// 	dependencies := map[string][]string{
// 		"countries": {},
// 		"locations": {"countries"},
// 		"regions":   {"a"},
// 		"jobs":      {"b"},
// 	}

// 	_, err := GetTablesOrderedByDependency(dependencies)
// 	assert.Error(t, err)
// }

// func Test_GetTablesOrderedByDependency_NestedDependencies(t *testing.T) {
// 	dependencies := map[string][]string{
// 		"a": {"b"},
// 		"b": {"c"},
// 		"c": {"d"},
// 		"d": {},
// 	}

// 	expected := []string{"d", "c", "b", "a"}
// 	actual, err := GetTablesOrderedByDependency(dependencies)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expected[0], actual.OrderedTables[0])
// 	assert.Equal(t, actual.HasCycles, false)
// }

// func TestCycleOrder(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		cycle    []string
// 		expected []string
// 	}{
// 		{
// 			name:     "Single element",
// 			cycle:    []string{"a"},
// 			expected: []string{"a"},
// 		},
// 		{
// 			name:     "Already sorted",
// 			cycle:    []string{"a", "b", "c"},
// 			expected: []string{"a", "b", "c"},
// 		},
// 		{
// 			name:     "Needs sorting",
// 			cycle:    []string{"b", "c", "a"},
// 			expected: []string{"a", "b", "c"},
// 		},
// 		{
// 			name:     "Duplicate minimums",
// 			cycle:    []string{"c", "a", "b", "a"},
// 			expected: []string{"a", "b", "a", "c"},
// 		},
// 		{
// 			name:     "All elements are same",
// 			cycle:    []string{"a", "a", "a"},
// 			expected: []string{"a", "a", "a"},
// 		},
// 		{
// 			name:     "Empty slice",
// 			cycle:    []string{},
// 			expected: []string{},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := cycleOrder(tt.cycle)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }
