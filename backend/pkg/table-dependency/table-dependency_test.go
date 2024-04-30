package tabledependency

import (
	"sort"
	"testing"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/stretchr/testify/require"
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
		{
			name: "Three circular dependencies + self referencing",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.b": {"public.c", "public.b"},
				"public.c": {"public.a"},
			},
			expect: [][]string{{"public.a", "public.b", "public.c"}, {"public.b"}},
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

			require.Len(t, tt.expect, len(actual))
			require.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func Test_determineCycleStart(t *testing.T) {
	tests := []struct {
		name          string
		cycle         []string
		subsets       map[string]string
		dependencyMap dbschemas.TableDependency
		expected      string
		expectError   bool
	}{
		{
			name:    "basic cycle with no subsets and nullable foreign keys",
			cycle:   []string{"a", "b"},
			subsets: map[string]string{},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: true},
					},
				},
			},
			expected:    "a",
			expectError: false,
		},
		{
			name:  "basic cycle with subsets and nullable foreign keys",
			cycle: []string{"a", "b"},
			subsets: map[string]string{
				"b": "where",
			},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: true},
					},
				},
			},
			expected:    "b",
			expectError: false,
		},
		{
			name:  "basic cycle with subsets and not nullable foreign keys",
			cycle: []string{"a", "b"},
			subsets: map[string]string{
				"b": "where",
			},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: false},
					},
				},
			},
			expected:    "a",
			expectError: false,
		},
		{
			name:          "cycle with missing dependencies",
			cycle:         []string{"a"},
			subsets:       map[string]string{},
			dependencyMap: dbschemas.TableDependency{},
			expected:      "",
			expectError:   true,
		},
		{
			name:    "cycle with non-nullable foreign keys",
			cycle:   []string{"a", "b"},
			subsets: map[string]string{},
			dependencyMap: dbschemas.TableDependency{
				"table1": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: false},
					},
				},
				"table2": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: false},
					},
				},
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycles := [][]string{tt.cycle}
			actual, err := determineCycleStarts(cycles, tt.subsets, tt.dependencyMap)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual[0])
			}
		})
	}
}

func Test_determineMultiCycleStart(t *testing.T) {
	tests := []struct {
		name          string
		cycles        [][]string
		subsets       map[string]string
		dependencyMap dbschemas.TableDependency
		expected      []string
		expectError   bool
	}{
		{
			name:    "multi cycle one starting point no subsets",
			cycles:  [][]string{{"a", "b", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: true},
						{ForeignKey: &dbschemas.ForeignKey{Table: "d"}, IsNullable: true},
					},
				},
				"c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: true},
					},
				},
				"d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "e"}, IsNullable: true},
					},
				},
				"e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
			},
			expected:    []string{"b"},
			expectError: false,
		},
		{
			name:    "multi cycle two starting points no subsets",
			cycles:  [][]string{{"a", "b", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: false},
						{ForeignKey: &dbschemas.ForeignKey{Table: "d"}, IsNullable: false},
					},
				},
				"c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: false},
					},
				},
				"d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "e"}, IsNullable: false},
					},
				},
				"e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
			},
			expected:    []string{"a", "e"},
			expectError: false,
		},
		{
			name:    "multi cycle two starting points no subsets 2",
			cycles:  [][]string{{"a", "e", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "e"}, IsNullable: false},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "d"}, IsNullable: true},
					},
				},
				"c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: false},
					},
				},
				"d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "e"}, IsNullable: false},
					},
				},
				"e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "c"}, IsNullable: true},
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: false},
					},
				},
			},
			expected:    []string{"b", "e"},
			expectError: false,
		},
		{
			name:   "multi cycle two starting point subsets",
			cycles: [][]string{{"a", "b", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{
				"a": "where",
			},
			dependencyMap: dbschemas.TableDependency{
				"a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: true},
					},
				},
				"b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: true},
						{ForeignKey: &dbschemas.ForeignKey{Table: "d"}, IsNullable: true},
					},
				},
				"c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "a"}, IsNullable: false},
					},
				},
				"d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "e"}, IsNullable: false},
					},
				},
				"e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{ForeignKey: &dbschemas.ForeignKey{Table: "b"}, IsNullable: false},
					},
				},
			},
			expected:    []string{"a", "b"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := determineCycleStarts(tt.cycles, tt.subsets, tt.dependencyMap)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tt.expected, actual)
			}
		})
	}
}

func Test_GetRunConfigs_NoSubset_SingleCycle(t *testing.T) {
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
		{
			name: "Single Cycle",
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
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id"},
				"public.c": {"id", "a_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "c_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Single Cycle Non Cycle Start",
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
				"public.x": {"id"},
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.x", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "x_id"}, DependsOn: []*DependsOn{{Table: "public.x", Columns: []string{"id"}}}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "c_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Self Referencing Cycle",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "a_id", "other"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "other"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Double Self Referencing Cycle",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
						{Column: "aa_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "a_id", "aa_id", "other"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "other"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"a_id", "aa_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Single Cycle Composite Foreign Keys",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "cc_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "other_id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "cc_id"},
				"public.c": {"id", "other_id", "a_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id", "other_id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id", "other_id"}, WhereClause: &where, Columns: []string{"id", "other_id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "c_id", "cc_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id", "other_id"}}}},
			},
		},
		{
			name: "Single Cycle Composite Foreign Keys Nullable",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "cc_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "other_id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "cc_id"},
				"public.c": {"id", "other_id", "a_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id", "other_id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "b_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id", "other_id"}, WhereClause: &where, Columns: []string{"id", "other_id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.b", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"c_id", "cc_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id", "other_id"}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.tables, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			require.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func Test_GetRunConfigs_Subset_SingleCycle(t *testing.T) {
	where := "where"
	emptyWhere := ""
	tests := []struct {
		name          string
		dependencies  dbschemas.TableDependency
		tables        []string
		subsets       map[string]string
		tableColsMap  map[string][]string
		primaryKeyMap map[string][]string
		expect        []*RunConfig
	}{
		{
			name: "Single Cycle",
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
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id"},
				"public.c": {"id", "a_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{
				"public.b": where,
			},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id", "c_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Single Cycle Non Cycle Start",
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
				"public.x": {"id"},
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{
				"public.x": "where",
			},
			expect: []*RunConfig{
				{Table: "public.x", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &where, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "x_id"}, DependsOn: []*DependsOn{{Table: "public.x", Columns: []string{"id"}}}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "c_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.tables, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			require.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func Test_GetRunConfigs_NoSubset_MultiCycle(t *testing.T) {
	emptyWhere := ""
	tests := []struct {
		name          string
		dependencies  dbschemas.TableDependency
		tables        []string
		subsets       map[string]string
		tableColsMap  map[string][]string
		primaryKeyMap map[string][]string
		expect        []*RunConfig
	}{
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
			},
			tables: []string{"public.a", "public.b", "public.c", "public.d", "public.e"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "d_id", "other_id"},
				"public.c": {"id", "a_id"},
				"public.d": {"id", "e_id"},
				"public.e": {"id", "b_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
				"public.e": {"id"},
				"public.d": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "other_id"}, DependsOn: []*DependsOn{}},
				{Table: "public.b", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"c_id", "d_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "b_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.d", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "e_id"}, DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
				{Table: "public.e", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "b_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Multi Table Dependencies Complex Foreign Keys",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
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
			},
			tables: []string{"public.a", "public.b", "public.c", "public.d", "public.e"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "d_id", "other_id"},
				"public.c": {"id", "a_id"},
				"public.d": {"id", "e_id"},
				"public.e": {"id", "b_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
				"public.e": {"id"},
				"public.d": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "c_id", "other_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"d_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
				{Table: "public.d", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "e_id"}, DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
				{Table: "public.e", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "b_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Multi Table Dependencies Self Referencing Circular Dependency Complex",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "bb_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "bb_id", "other_id"},
				"public.c": {"id", "a_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id"}, DependsOn: []*DependsOn{}},
				{Table: "public.a", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"b_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "c_id", "other_id"}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"bb_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
		{
			name: "Multi Table Dependencies Self Referencing Circular Dependency Simple",
			dependencies: dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
						{Column: "bb_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			},
			tables: []string{"public.a", "public.b", "public.c"},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "bb_id", "other_id"},
				"public.c": {"id", "a_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				{Table: "public.a", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "b_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "other_id"}, DependsOn: []*DependsOn{}},
				{Table: "public.b", RunType: RunTypeUpdate, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"c_id", "bb_id"}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}}},
				{Table: "public.c", RunType: RunTypeInsert, PrimaryKeys: []string{"id"}, WhereClause: &emptyWhere, Columns: []string{"id", "a_id"}, DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.tables, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			for _, e := range tt.expect {
				acutalConfig := getConfigByTableAndType(e.Table, e.RunType, actual)
				require.NotNil(t, acutalConfig)
				require.ElementsMatch(t, e.Columns, acutalConfig.Columns)
				require.ElementsMatch(t, e.DependsOn, acutalConfig.DependsOn)
				require.ElementsMatch(t, e.PrimaryKeys, acutalConfig.PrimaryKeys)
				require.Equal(t, e.WhereClause, e.WhereClause)
			}
		})
	}
}

func getConfigByTableAndType(table string, runtype RunType, configs []*RunConfig) *RunConfig {
	for _, c := range configs {
		if c.Table == table && c.RunType == runtype {
			return c
		}
	}
	return nil
}

func Test_GetRunConfigs_CircularDependencyNoneNullable(t *testing.T) {
	dependencies := dbschemas.TableDependency{
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
	}
	tables := []string{"public.a", "public.b"}
	_, err := GetRunConfigs(dependencies, tables, map[string]string{}, map[string][]string{}, map[string][]string{})
	require.Error(t, err)
}

func Test_GetTablesOrderedByDependency_CircularDependency(t *testing.T) {
	dependencies := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}

	resp, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, resp.HasCycles, true)
	for _, e := range resp.OrderedTables {
		require.Contains(t, []string{"a", "b", "c"}, e)
	}
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
	expected := [][]string{{"regions", "jobs"}, {"regions", "jobs"}, {"countries"}, {"locations"}, {"departments"}, {"employees"}, {"dependents"}}

	actual, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, actual.HasCycles, false)

	for idx, table := range actual.OrderedTables {
		require.Contains(t, expected[idx], table)
	}
}

func Test_GetTablesOrderedByDependency_Mixed(t *testing.T) {
	dependencies := map[string][]string{
		"countries": {},
		"locations": {"countries"},
		"regions":   {},
		"jobs":      {},
	}

	expected := []string{"countries", "regions", "jobs", "locations"}
	actual, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, actual.HasCycles, false)
	require.Len(t, actual.OrderedTables, len(expected))
	for _, table := range actual.OrderedTables {
		require.Contains(t, expected, table)
	}
	require.Equal(t, "locations", actual.OrderedTables[len(actual.OrderedTables)-1])
}

func Test_GetTablesOrderedByDependency_BrokenDependencies_NoLoop(t *testing.T) {
	dependencies := map[string][]string{
		"countries": {},
		"locations": {"countries"},
		"regions":   {"a"},
		"jobs":      {"b"},
	}

	_, err := GetTablesOrderedByDependency(dependencies)
	require.Error(t, err)
}

func Test_GetTablesOrderedByDependency_NestedDependencies(t *testing.T) {
	dependencies := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"d"},
		"d": {},
	}

	expected := []string{"d", "c", "b", "a"}
	actual, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, expected[0], actual.OrderedTables[0])
	require.Equal(t, actual.HasCycles, false)
}

func TestCycleOrder(t *testing.T) {
	tests := []struct {
		name     string
		cycle    []string
		expected []string
	}{
		{
			name:     "Single element",
			cycle:    []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "Already sorted",
			cycle:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Needs sorting",
			cycle:    []string{"b", "c", "a"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Duplicate minimums",
			cycle:    []string{"c", "a", "b", "a"},
			expected: []string{"a", "b", "a", "c"},
		},
		{
			name:     "All elements are same",
			cycle:    []string{"a", "a", "a"},
			expected: []string{"a", "a", "a"},
		},
		{
			name:     "Empty slice",
			cycle:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cycleOrder(tt.cycle)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_isValidRunOrder(t *testing.T) {
	tests := []struct {
		name     string
		configs  []*RunConfig
		expected bool
	}{
		{
			name:     "empty configuration",
			configs:  []*RunConfig{},
			expected: true,
		},
		{
			name: "single node with no dependencies",
			configs: []*RunConfig{
				{Table: "users", RunType: "initial", Columns: []string{"id", "name"}, DependsOn: nil},
			},
			expected: true,
		},
		{
			name: "multiple nodes with no dependencies",
			configs: []*RunConfig{
				{Table: "users", RunType: "initial", Columns: []string{"id", "name"}, DependsOn: nil},
				{Table: "products", RunType: "initial", Columns: []string{"id", "name"}, DependsOn: nil},
			},
			expected: true,
		},
		{
			name: "simple dependency chain",
			configs: []*RunConfig{
				{Table: "users", RunType: "initial", Columns: []string{"id"}, DependsOn: nil},
				{Table: "orders", RunType: "initial", Columns: []string{"user_id"}, DependsOn: []*DependsOn{{Table: "users", Columns: []string{"id"}}}},
			},
			expected: true,
		},
		{
			name: "circular dependency",
			configs: []*RunConfig{
				{Table: "users", RunType: "initial", Columns: []string{"id"}, DependsOn: []*DependsOn{{Table: "orders", Columns: []string{"user_id"}}}},
				{Table: "orders", RunType: "initial", Columns: []string{"user_id"}, DependsOn: []*DependsOn{{Table: "users", Columns: []string{"id"}}}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidRunOrder(tt.configs)
			require.Equal(t, tt.expected, actual)
		})
	}
}
