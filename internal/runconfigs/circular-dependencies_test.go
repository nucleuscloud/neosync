package runconfigs

import (
	"sort"
	"testing"

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
