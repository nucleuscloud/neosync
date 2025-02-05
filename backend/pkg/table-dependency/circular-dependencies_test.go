package tabledependency

import (
	"sort"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
		{
			name: "Nested cycles",
			dependencies: map[string][]string{
				"public.a": {"public.b"},
				"public.b": {"public.c"},
				"public.c": {"public.d", "public.a"},
				"public.d": {"public.b"},
			},
			expect: [][]string{
				{"public.a", "public.b", "public.c"},
				{"public.b", "public.c", "public.d"},
			},
		},
		{
			name: "Multiple overlapping cycles with shared nodes",
			dependencies: map[string][]string{
				"public.a": {"public.b", "public.d"},
				"public.b": {"public.c"},
				"public.c": {"public.a"},
				"public.d": {"public.e"},
				"public.e": {"public.f"},
				"public.f": {"public.d", "public.a"},
			},
			expect: [][]string{
				{"public.a", "public.b", "public.c"},
				{"public.a", "public.d", "public.e", "public.f"},
				{"public.d", "public.e", "public.f"},
			},
		},
		{
			name: "Diamond shape with multiple paths",
			dependencies: map[string][]string{
				"public.a": {"public.b", "public.c"},
				"public.b": {"public.d"},
				"public.c": {"public.d"},
				"public.d": {"public.a"},
			},
			expect: [][]string{
				{"public.a", "public.b", "public.d"},
				{"public.a", "public.c", "public.d"},
			},
		},
		{
			name: "Complex web of dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.b", "public.c"},
				"public.b": {"public.d", "public.e"},
				"public.c": {"public.e", "public.f"},
				"public.d": {"public.g"},
				"public.e": {"public.g", "public.h"},
				"public.f": {"public.h"},
				"public.g": {"public.i"},
				"public.h": {"public.i"},
				"public.i": {"public.a"},
			},
			expect: [][]string{
				{"public.a", "public.b", "public.d", "public.g", "public.i"},
				{"public.a", "public.b", "public.e", "public.g", "public.i"},
				{"public.a", "public.b", "public.e", "public.h", "public.i"},
				{"public.a", "public.c", "public.e", "public.g", "public.i"},
				{"public.a", "public.c", "public.e", "public.h", "public.i"},
				{"public.a", "public.c", "public.f", "public.h", "public.i"},
			},
		},
		{
			name: "Multiple self-references with shared dependencies",
			dependencies: map[string][]string{
				"public.a": {"public.a", "public.b"},
				"public.b": {"public.b", "public.c"},
				"public.c": {"public.a", "public.c"},
			},
			expect: [][]string{
				{"public.a"},
				{"public.b"},
				{"public.c"},
				{"public.a", "public.b", "public.c"},
			},
		},
		{
			name: "Cycle with branching paths",
			dependencies: map[string][]string{
				"public.root": {"public.a1", "public.a2"},
				"public.a1":   {"public.b1"},
				"public.a2":   {"public.b2"},
				"public.b1":   {"public.c"},
				"public.b2":   {"public.c"},
				"public.c":    {"public.root"},
			},
			expect: [][]string{
				{"public.root", "public.a1", "public.b1", "public.c"},
				{"public.root", "public.a2", "public.b2", "public.c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FindCircularDependencies(tt.dependencies)

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

func Test_uniqueCycles(t *testing.T) {
	tests := []struct {
		name   string
		cycles [][]string
		expect [][]string
	}{
		{
			name:   "duplicates",
			cycles: [][]string{{"a", "b", "c", "d"}, {"a", "b"}, {"b", "c"}, {"c", "d"}, {"d", "a"}, {"c", "a", "d", "b"}},
			expect: [][]string{{"a", "b", "c", "d"}, {"a", "b"}, {"b", "c"}, {"c", "d"}, {"d", "a"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := uniqueCycles(tt.cycles)

			require.Len(t, actual, len(tt.expect))
			require.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func Test_determineCycleStart(t *testing.T) {
	tests := []struct {
		name          string
		cycle         []string
		subsets       map[string]string
		dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint
		expected      string
		expectError   bool
	}{
		{
			name:    "basic cycle with no subsets and nullable foreign keys",
			cycle:   []string{"a", "b"},
			subsets: map[string]string{},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{true}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{false}},
				},
			},
			expected:    "b",
			expectError: false,
		},
		{
			name:  "basic cycle with subsets and nullable foreign keys",
			cycle: []string{"a", "b"},
			subsets: map[string]string{
				"b": "where",
			},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{false}},
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
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{true}},
				},
			},
			expected:    "a",
			expectError: false,
		},
		{
			name:          "cycle with missing dependencies",
			cycle:         []string{"a"},
			subsets:       map[string]string{},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{},
			expected:      "",
			expectError:   true,
		},
		{
			name:    "cycle with non-nullable foreign keys",
			cycle:   []string{"a", "b"},
			subsets: map[string]string{},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{true}},
				},
				"table2": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{true}},
				},
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycles := [][]string{tt.cycle}
			actual, err := DetermineCycleInsertUpdateTables(cycles, tt.subsets, tt.dependencyMap)
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
		dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint
		expected      []string
		expectError   bool
	}{
		{
			name:    "multi cycle one starting point no subsets",
			cycles:  [][]string{{"a", "b", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{false}},
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "d"}, NotNullable: []bool{false}},
				},
				"c": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{false}},
				},
				"d": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "e"}, NotNullable: []bool{false}},
				},
				"e": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
			},
			expected:    []string{"b"},
			expectError: false,
		},
		{
			name:    "multi cycle two starting points no subsets",
			cycles:  [][]string{{"a", "b", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{true}},
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "d"}, NotNullable: []bool{true}},
				},
				"c": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{true}},
				},
				"d": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "e"}, NotNullable: []bool{true}},
				},
				"e": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
			},
			expected:    []string{"a", "e"},
			expectError: false,
		},
		{
			name:    "multi cycle two starting points no subsets 2",
			cycles:  [][]string{{"a", "e", "c"}, {"d", "e", "b"}},
			subsets: map[string]string{},
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "e"}, NotNullable: []bool{true}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "d"}, NotNullable: []bool{false}},
				},
				"c": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{true}},
				},
				"d": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "e"}, NotNullable: []bool{true}},
				},
				"e": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "c"}, NotNullable: []bool{false}},
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{true}},
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
			dependencyMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"a": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{false}},
				},
				"b": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{false}},
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "d"}, NotNullable: []bool{false}},
				},
				"c": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "a"}, NotNullable: []bool{true}},
				},
				"d": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "e"}, NotNullable: []bool{true}},
				},
				"e": {
					{ForeignKey: &sqlmanager_shared.ForeignKey{Table: "b"}, NotNullable: []bool{true}},
				},
			},
			expected:    []string{"a", "b"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := DetermineCycleInsertUpdateTables(tt.cycles, tt.subsets, tt.dependencyMap)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tt.expected, actual)
			}
		})
	}
}

func Test_determineSingleCycleMultipleStarts_SplitNullable(t *testing.T) {
	cycles := [][]string{{"public.a", "public.b", "public.c"}}
	subsets := map[string]string{}
	dependencyMap := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		},
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"ac_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"c_id"}}},
		},
		"public.c": {
			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
			{Columns: []string{"acb_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"ac_id"}}},
		},
	}
	expected := []string{"public.a", "public.b", "public.c"}

	actual, err := DetermineCycleInsertUpdateTables(cycles, subsets, dependencyMap)
	require.NoError(t, err)
	require.ElementsMatch(t, expected, actual)
}

func Test_determineSingleCycleMultipleStarts_Complicated(t *testing.T) {
	cycles := [][]string{{"public.a", "public.b", "public.c"}}
	subsets := map[string]string{}
	dependencyMap := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		},
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			{Columns: []string{"ac_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"c_id"}}},
		},
		"public.c": {
			{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
			{Columns: []string{"acb_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"ac_id"}}},
		},
	}
	expected := []string{"public.a", "public.c"}

	actual, err := DetermineCycleInsertUpdateTables(cycles, subsets, dependencyMap)
	require.NoError(t, err)
	require.ElementsMatch(t, expected, actual)
}

func Test_CycleOrder(t *testing.T) {
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
			expected: []string{"a", "a", "b", "c"},
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
