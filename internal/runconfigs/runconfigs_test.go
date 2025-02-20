package runconfigs

import (
	"slices"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetRunConfigs_NoSubset_SingleCycle(t *testing.T) {
	where := ""

	t.Run("Single Cycle", func(t *testing.T) {
		dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.a": {
				{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
			},
			"public.b": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
			},
			"public.c": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			},
		}
		tableColsMap := map[string][]string{
			"public.a": {"id", "b_id"},
			"public.b": {"id", "c_id"},
			"public.c": {"id", "a_id"},
		}
		primaryKeyMap := map[string][]string{
			"public.a": {"id"},
			"public.b": {"id"},
			"public.c": {"id"},
		}
		subsets := map[string]string{}
		expect := []*RunConfig{
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"id"}},
			}),
			buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
			}),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})

	t.Run("Single Cycle Non Cycle Start", func(t *testing.T) {
		dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.a": {
				{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				{Columns: []string{"x_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.x", Columns: []string{"id"}}},
			},
			"public.b": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
			},
			"public.c": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			},
		}
		tableColsMap := map[string][]string{
			"public.a": {"id", "b_id", "x_id"},
			"public.b": {"id", "c_id"},
			"public.c": {"id", "a_id"},
			"public.x": {"id"},
		}
		primaryKeyMap := map[string][]string{
			"public.x": {"id"},
			"public.a": {"id"},
			"public.b": {"id"},
			"public.c": {"id"},
		}
		subsets := map[string]string{}
		expect := []*RunConfig{
			buildRunConfig("public.x", RunTypeInsert, []string{"id"}, &where, []string{"id"}, []string{"id"}, []*DependsOn{}),
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "b_id", "x_id"}, []string{"id", "x_id"}, []*DependsOn{
				{Table: "public.x", Columns: []string{"id"}},
			}),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"id"}},
			}),
			buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
			}),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})

	t.Run("Self Referencing Cycle", func(t *testing.T) {
		dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.a": {
				{Columns: []string{"a_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			},
		}
		tableColsMap := map[string][]string{
			"public.a": {"id", "a_id", "other"},
		}
		primaryKeyMap := map[string][]string{
			"public.a": {"id"},
		}
		subsets := map[string]string{}
		expect := []*RunConfig{
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id", "other"}, []string{"id", "other"}, []*DependsOn{}),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "a_id"}, []string{"a_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})

	t.Run("Double Self Referencing Cycle", func(t *testing.T) {
		dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.a": {
				{Columns: []string{"a_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				{Columns: []string{"aa_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			},
		}
		tableColsMap := map[string][]string{
			"public.a": {"id", "a_id", "aa_id", "other"},
		}
		primaryKeyMap := map[string][]string{
			"public.a": {"id"},
		}
		subsets := map[string]string{}
		expect := []*RunConfig{
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "a_id", "aa_id", "other"},
				[]string{"id", "other"},
				[]*DependsOn{}),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "a_id"},
				[]string{"a_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				}),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "aa_id"},
				[]string{"aa_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				}),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})

	t.Run("Single Cycle Composite Foreign Keys", func(t *testing.T) {
		dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.a": {
				{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
			},
			"public.b": {
				{Columns: []string{"c_id", "cc_id"}, NotNullable: []bool{true, true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id", "other_id"}}},
			},
			"public.c": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			},
		}
		tableColsMap := map[string][]string{
			"public.a": {"id", "b_id"},
			"public.b": {"id", "c_id", "cc_id"},
			"public.c": {"id", "other_id", "a_id"},
		}
		primaryKeyMap := map[string][]string{
			"public.a": {"id"},
			"public.b": {"id"},
			"public.c": {"id", "other_id"},
		}
		subsets := map[string]string{}
		expect := []*RunConfig{
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "b_id"},
				[]string{"id"},
				[]*DependsOn{}),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "b_id"},
				[]string{"b_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
					{Table: "public.b", Columns: []string{"id"}},
				}),
			buildRunConfig("public.c", RunTypeInsert, []string{"id", "other_id"}, &where,
				[]string{"id", "other_id", "a_id"},
				[]string{"id", "other_id", "a_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				}),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "c_id", "cc_id"},
				[]string{"id", "c_id", "cc_id"},
				[]*DependsOn{
					{Table: "public.c", Columns: []string{"id", "other_id"}},
				}),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})

	t.Run("Single Cycle Composite Foreign Keys Nullable", func(t *testing.T) {
		dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.a": {
				{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
			},
			"public.b": {
				{Columns: []string{"c_id", "cc_id"}, NotNullable: []bool{false, false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id", "other_id"}}},
			},
			"public.c": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
			},
		}
		tableColsMap := map[string][]string{
			"public.a": {"id", "b_id"},
			"public.b": {"id", "c_id", "cc_id"},
			"public.c": {"id", "other_id", "a_id"},
		}
		primaryKeyMap := map[string][]string{
			"public.a": {"id"},
			"public.b": {"id"},
			"public.c": {"id", "other_id"},
		}
		subsets := map[string]string{}
		expect := []*RunConfig{
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "b_id"},
				[]string{"id", "b_id"},
				[]*DependsOn{
					{Table: "public.b", Columns: []string{"id"}},
				}),
			buildRunConfig("public.c", RunTypeInsert, []string{"id", "other_id"}, &where,
				[]string{"id", "other_id", "a_id"},
				[]string{"id", "other_id", "a_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				}),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "c_id", "cc_id"},
				[]string{"id"},
				[]*DependsOn{}),
			buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "c_id", "cc_id"},
				[]string{"c_id", "cc_id"},
				[]*DependsOn{
					{Table: "public.b", Columns: []string{"id"}},
					{Table: "public.c", Columns: []string{"id", "other_id"}},
				}),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})
}

func Test_GetRunConfigs_Subset_SingleCycle(t *testing.T) {
	where := "where"
	emptyWhere := ""
	tests := []struct {
		name          string
		dependencies  map[string][]*sqlmanager_shared.ForeignConstraint
		subsets       map[string]string
		tableColsMap  map[string][]string
		primaryKeyMap map[string][]string
		expect        []*RunConfig
	}{
		{
			name: "Single Cycle",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
			},
		},
		{
			name: "Single Cycle Non Cycle Start",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
					{Columns: []string{"x_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.x", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
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
				buildRunConfig("public.x", RunTypeInsert, []string{"id"}, &where, []string{"id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id", "x_id"}, []string{"id", "x_id"}, []*DependsOn{{Table: "public.x", Columns: []string{"id"}}}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRunConfigs(t, tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap, tt.expect)
		})
	}
}

func Test_GetRunConfigs_NoSubset_MultiCycle(t *testing.T) {
	emptyWhere := ""
	tests := []struct {
		name          string
		dependencies  map[string][]*sqlmanager_shared.ForeignConstraint
		subsets       map[string]string
		tableColsMap  map[string][]string
		primaryKeyMap map[string][]string
		expect        []*RunConfig
	}{
		{
			name: "Multi Table Dependencies",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
					{Columns: []string{"d_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.d", Columns: []string{"id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
				"public.d": {
					{Columns: []string{"e_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.e", Columns: []string{"id"}}},
				},
				"public.e": {
					{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
			},
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
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id", "other_id"}, []string{"id", "other_id"}, []*DependsOn{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "c_id"}, []string{"c_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "d_id"}, []string{"d_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
				buildRunConfig("public.d", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "e_id"}, []string{"id", "e_id"}, []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}),
				buildRunConfig("public.e", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
			},
		},
		{
			name: "Multi Table Dependencies Complex Foreign Keys",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
					{Columns: []string{"d_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.d", Columns: []string{"id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
				"public.d": {
					{Columns: []string{"e_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.e", Columns: []string{"id"}}},
				},
				"public.e": {
					{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
			},
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "d_id"}, []string{"d_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
				buildRunConfig("public.d", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "e_id"}, []string{"id", "e_id"}, []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}),
				buildRunConfig("public.e", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
			},
		},
		{
			name: "Multi Table Dependencies Self Referencing Circular Dependency Complex",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
					{Columns: []string{"bb_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "bb_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "bb_id"}, []string{"bb_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
			},
		},
		{
			name: "Multi Table Dependencies Self Referencing Circular Dependency Simple",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
					{Columns: []string{"bb_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "bb_id", "other_id"}, []string{"id", "other_id"}, []*DependsOn{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "c_id"}, []string{"c_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "bb_id"}, []string{"bb_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRunConfigs(t, tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap, tt.expect)
		})
	}
}

func getConfigByTableAndType(table string, runtype RunType, insertCols []string, configs []*RunConfig) *RunConfig {
	for _, c := range configs {
		cCols := slices.Clone(c.InsertColumns())
		iCols := slices.Clone(insertCols)
		slices.Sort(cCols)
		slices.Sort(iCols)
		if c.Table() == table && c.RunType() == runtype && slices.Equal(cCols, iCols) {
			return c
		}
	}
	return nil
}

func buildRunConfig(
	table string,
	runtype RunType,
	pks []string,
	where *string,
	selectCols, insertCols []string,
	dependsOn []*DependsOn,
) *RunConfig {
	rc := &RunConfig{
		table:         table,
		runType:       runtype,
		selectColumns: selectCols,
		insertColumns: insertCols,
		primaryKeys:   pks,
		whereClause:   where,
		dependsOn:     dependsOn,
	}
	return rc
}

func assertRunConfigs(t *testing.T, dependencies map[string][]*sqlmanager_shared.ForeignConstraint, subsets map[string]string, primaryKeyMap map[string][]string, tableColsMap map[string][]string, expect []*RunConfig) {
	actual, err := GetRunConfigs(dependencies, subsets, primaryKeyMap, tableColsMap, map[string][][]string{}, map[string][][]string{})
	require.NoError(t, err)
	assert.Len(t, actual, len(expect), "expected %d configs but got %d", len(expect), len(actual))
	for _, e := range expect {
		acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), e.InsertColumns(), actual)
		require.NotNil(t, acutalConfig, "expected config for table %s (type: %s, insert columns: %v) to exist", e.Table(), e.RunType(), e.InsertColumns())
		assert.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns(),
			"Select columns mismatch for table %s (type: %s) - expected %v but got %v",
			e.Table(), e.RunType(), e.SelectColumns(), acutalConfig.SelectColumns())
		assert.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns(),
			"Insert columns mismatch for table %s (type: %s) - expected %v but got %v",
			e.Table(), e.RunType(), e.InsertColumns(), acutalConfig.InsertColumns())
		assert.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn(),
			"Dependencies mismatch for table %s (type: %s)",
			e.Table(), e.RunType())
		assert.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys(),
			"Primary keys mismatch for table %s (type: %s) - expected %v but got %v",
			e.Table(), e.RunType(), e.PrimaryKeys(), acutalConfig.PrimaryKeys())
		assert.Equal(t, e.WhereClause(), e.WhereClause(),
			"Where clause mismatch for table %s (type: %s) - expected %v but got %v",
			e.Table(), e.RunType(), e.WhereClause(), e.WhereClause())
	}
}
