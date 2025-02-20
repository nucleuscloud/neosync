package tabledependency

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func Test_GetTablesOrderedByDependency_CircularDependency(t *testing.T) {
	dependencies := map[string][]string{
		"other.a": {"other.b"},
		"other.b": {"other.c"},
		"other.c": {"other.a"},
	}

	resp, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, resp.HasCycles, true)
	for _, e := range resp.OrderedTables {
		require.Contains(t, []*sqlmanager_shared.SchemaTable{{Schema: "other", Table: "a"}, {Schema: "other", Table: "b"}, {Schema: "other", Table: "c"}}, e)
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
	expected := [][]*sqlmanager_shared.SchemaTable{
		{{Schema: "public", Table: "regions"}, {Schema: "public", Table: "jobs"}},
		{{Schema: "public", Table: "regions"}, {Schema: "public", Table: "jobs"}},
		{{Schema: "public", Table: "countries"}},
		{{Schema: "public", Table: "locations"}},
		{{Schema: "public", Table: "departments"}},
		{{Schema: "public", Table: "employees"}},
		{{Schema: "public", Table: "dependents"}}}

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

	expected := []*sqlmanager_shared.SchemaTable{{Schema: "public", Table: "countries"}, {Schema: "public", Table: "regions"}, {Schema: "public", Table: "jobs"}, {Schema: "public", Table: "locations"}}
	actual, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, actual.HasCycles, false)
	require.Len(t, actual.OrderedTables, len(expected))
	for _, table := range actual.OrderedTables {
		require.Contains(t, expected, table)
	}
	require.Equal(t, &sqlmanager_shared.SchemaTable{Schema: "public", Table: "locations"}, actual.OrderedTables[len(actual.OrderedTables)-1])
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

	expected := []*sqlmanager_shared.SchemaTable{{Schema: "public", Table: "d"}, {Schema: "public", Table: "c"}, {Schema: "public", Table: "b"}, {Schema: "public", Table: "a"}}
	actual, err := GetTablesOrderedByDependency(dependencies)
	require.NoError(t, err)
	require.Equal(t, expected[0], actual.OrderedTables[0])
	require.Equal(t, actual.HasCycles, false)
}
