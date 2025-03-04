package runconfigs

import (
	"slices"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BuildRunConfigs_NoSubset_SingleCycle(t *testing.T) {
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
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}, nil),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"id"}},
			}, nil),
			buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}, nil),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
			}, nil),
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
			buildRunConfig("public.x", RunTypeInsert, []string{"id"}, &where, []string{"id"}, []string{"id"}, []*DependsOn{}, nil),
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "b_id", "x_id"}, []string{"id", "x_id"}, []*DependsOn{
				{Table: "public.x", Columns: []string{"id"}},
			}, nil),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"id"}},
			}, nil),
			buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}, nil),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
			}, nil),
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
			buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id", "other"}, []string{"id", "other"}, []*DependsOn{}, nil),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "a_id"}, []string{"a_id"}, []*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}, nil),
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
				[]*DependsOn{},
				nil),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "a_id"},
				[]string{"a_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				},
				nil),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "aa_id"},
				[]string{"aa_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				},
				nil),
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
				[]*DependsOn{},
				nil),
			buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "b_id"},
				[]string{"b_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
					{Table: "public.b", Columns: []string{"id"}},
				},
				nil),
			buildRunConfig("public.c", RunTypeInsert, []string{"id", "other_id"}, &where,
				[]string{"id", "other_id", "a_id"},
				[]string{"id", "other_id", "a_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				},
				nil),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "c_id", "cc_id"},
				[]string{"id", "c_id", "cc_id"},
				[]*DependsOn{
					{Table: "public.c", Columns: []string{"id", "other_id"}},
				},
				nil),
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
				},
				nil),
			buildRunConfig("public.c", RunTypeInsert, []string{"id", "other_id"}, &where,
				[]string{"id", "other_id", "a_id"},
				[]string{"id", "other_id", "a_id"},
				[]*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				},
				nil),
			buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where,
				[]string{"id", "c_id", "cc_id"},
				[]string{"id"},
				[]*DependsOn{},
				nil),
			buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &where,
				[]string{"id", "c_id", "cc_id"},
				[]string{"c_id", "cc_id"},
				[]*DependsOn{
					{Table: "public.b", Columns: []string{"id"}},
					{Table: "public.c", Columns: []string{"id", "other_id"}},
				},
				nil),
		}

		assertRunConfigs(t, dependencies, subsets, primaryKeyMap, tableColsMap, expect)
	})
}

func Test_BuildRunConfigs_Subset_SingleCycle(t *testing.T) {
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{
					{
						Subset: where, Root: "public.b", JoinSteps: []*JoinStep{
							{FromKey: "public.a", ToKey: "public.b", ForeignKey: &ForeignKey{Columns: []string{"b_id"}, NotNullable: []bool{false}, ReferenceSchema: "public", ReferenceTable: "b", ReferenceColumns: []string{"id"}}},
						},
					},
				}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.b", JoinSteps: []*JoinStep{
							{FromKey: "public.a", ToKey: "public.b", ForeignKey: &ForeignKey{Columns: []string{"b_id"}, NotNullable: []bool{false}, ReferenceSchema: "public", ReferenceTable: "b", ReferenceColumns: []string{"id"}}},
						},
					},
				}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.b", JoinSteps: []*JoinStep{
							{FromKey: "public.c", ToKey: "public.a", ForeignKey: &ForeignKey{Columns: []string{"a_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "a", ReferenceColumns: []string{"id"}}},
							{FromKey: "public.a", ToKey: "public.b", ForeignKey: &ForeignKey{Columns: []string{"b_id"}, NotNullable: []bool{false}, ReferenceSchema: "public", ReferenceTable: "b", ReferenceColumns: []string{"id"}}},
						},
					},
				}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &where, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.b", JoinSteps: []*JoinStep{},
					},
				}),
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
				buildRunConfig("public.x", RunTypeInsert, []string{"id"}, &where, []string{"id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{
					{
						Subset: where, Root: "public.x", JoinSteps: []*JoinStep{},
					},
				}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id", "x_id"}, []string{"id", "x_id"}, []*DependsOn{{Table: "public.x", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.x", JoinSteps: []*JoinStep{{FromKey: "public.a", ToKey: "public.x", ForeignKey: &ForeignKey{Columns: []string{"x_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "x", ReferenceColumns: []string{"id"}}}},
					},
				}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.x", JoinSteps: []*JoinStep{{FromKey: "public.a", ToKey: "public.x", ForeignKey: &ForeignKey{Columns: []string{"x_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "x", ReferenceColumns: []string{"id"}}}},
					},
				}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.x", JoinSteps: []*JoinStep{
							{FromKey: "public.c", ToKey: "public.a", ForeignKey: &ForeignKey{Columns: []string{"a_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "a", ReferenceColumns: []string{"id"}}},
							{FromKey: "public.a", ToKey: "public.x", ForeignKey: &ForeignKey{Columns: []string{"x_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "x", ReferenceColumns: []string{"id"}}},
						},
					},
				}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id"}, []string{"id", "c_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{
					{
						Subset: where, Root: "public.x", JoinSteps: []*JoinStep{
							{FromKey: "public.b", ToKey: "public.c", ForeignKey: &ForeignKey{Columns: []string{"c_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "c", ReferenceColumns: []string{"id"}}},
							{FromKey: "public.c", ToKey: "public.a", ForeignKey: &ForeignKey{Columns: []string{"a_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "a", ReferenceColumns: []string{"id"}}},
							{FromKey: "public.a", ToKey: "public.x", ForeignKey: &ForeignKey{Columns: []string{"x_id"}, NotNullable: []bool{true}, ReferenceSchema: "public", ReferenceTable: "x", ReferenceColumns: []string{"id"}}},
						},
					},
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRunConfigs(t, tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap, tt.expect)
		})
	}
}

func Test_BuildRunConfigs_NoSubset_MultiCycle(t *testing.T) {
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
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id", "other_id"}, []string{"id", "other_id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "c_id"}, []string{"c_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "d_id"}, []string{"d_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.d", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "e_id"}, []string{"id", "e_id"}, []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.e", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "d_id"}, []string{"d_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.d", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "e_id"}, []string{"id", "e_id"}, []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.e", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "bb_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "bb_id"}, []string{"bb_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}, []*SubsetPath{}),
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
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "bb_id", "other_id"}, []string{"id", "other_id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "c_id"}, []string{"c_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "bb_id"}, []string{"bb_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}, []*SubsetPath{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRunConfigs(t, tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap, tt.expect)
		})
	}
}

func Test_BuildRunConfigs_NoSubset_NoCycle(t *testing.T) {
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
			name: "Straight dependencies",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
				},
				"public.c": {},
			},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "other_id"},
				"public.c": {"id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
			},
		},
		{
			name: "Duplicate Columns",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
				},
				"public.c": {},
			},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id", "id"},
				"public.b": {"id", "c_id", "other_id", "id"},
				"public.c": {"id", "id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
				"public.c": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id"}, []string{"id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
			},
		},
		{
			name: "Sub Tree",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
				},
				"public.c": {},
			},
			tableColsMap: map[string][]string{
				"public.a": {"id", "b_id"},
				"public.b": {"id", "c_id", "other_id"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
				"public.b": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{}, []*SubsetPath{}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}, []*SubsetPath{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRunConfigs(t, tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap, tt.expect)
		})
	}
}

func Test_BuildRunConfigs_CompositeKey(t *testing.T) {
	emptyWhere := ""
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.department", Columns: []string{"department_id"}}},
		},
		"public.projects": {
			{Columns: []string{"responsible_employee_id", "responsible_department_id"}, NotNullable: []bool{false, false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id", "department_id"}}},
		},
	}
	primaryKeyMap := map[string][]string{
		"public.department": {
			"department_id",
		},
		"public.employees": {
			"employee_id",
			"department_id",
		},
		"public.projects": {
			"project_id",
		},
	}
	tablesColMap := map[string][]string{
		"public.department": {
			"department_id",
			"department_name",
			"location",
		},
		"public.employees": {
			"employee_id",
			"department_id",
			"first_name",
			"last_name",
			"email",
		},
		"public.projects": {
			"project_id",
			"project_name",
			"start_date",
			"end_date",
			"responsible_employee_id",
			"responsible_department_id",
		},
	}

	expect := []*RunConfig{
		buildRunConfig("public.employees", RunTypeInsert, []string{"employee_id", "department_id"}, &emptyWhere,
			[]string{"employee_id", "department_id", "first_name", "last_name", "email"},
			[]string{"employee_id", "department_id", "first_name", "last_name", "email"},
			[]*DependsOn{{Table: "public.department", Columns: []string{"department_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.department", RunTypeInsert, []string{"department_id"}, &emptyWhere,
			[]string{"department_id", "department_name", "location"},
			[]string{"department_id", "department_name", "location"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.projects", RunTypeInsert, []string{"project_id"}, &emptyWhere,
			[]string{"project_id", "project_name", "start_date", "end_date", "responsible_employee_id", "responsible_department_id"},
			[]string{"project_id", "project_name", "start_date", "end_date", "responsible_employee_id", "responsible_department_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id", "department_id"}}}, []*SubsetPath{}),
	}

	assertRunConfigs(t, dependencies, map[string]string{}, primaryKeyMap, tablesColMap, expect)
}

func Test_BuildRunConfigs_HumanResources(t *testing.T) {
	emptyWhere := ""
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}},
		},
		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"employee_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
		"public.employees": {
			{Columns: []string{"job_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
	}
	primaryKeyMap := map[string][]string{
		"public.regions":     {"region_id"},
		"public.countries":   {"country_id"},
		"public.locations":   {"location_id"},
		"public.jobs":        {"job_id"},
		"public.departments": {"department_id"},
		"public.employees":   {"employee_id"},
		"public.dependents":  {"dependent_id"},
	}
	tablesColMap := map[string][]string{
		"public.regions": {
			"region_id",
			"region_name",
		},
		"public.countries": {
			"country_id",
			"country_name",
			"region_id",
		},
		"public.locations": {
			"location_id",
			"street_address",
			"country_id",
		},
		"public.jobs": {
			"job_id",
			"job_title",
		},
		"public.departments": {
			"department_id",
			"department_name",
			"location_id",
		},
		"public.employees": {
			"employee_id",
			"email",
			"name",
			"job_id",
			"manager_id",
			"department_id",
		},
		"public.dependents": {
			"dependent_id",
			"name",
			"employee_id",
		},
	}

	expect := []*RunConfig{
		buildRunConfig("public.regions", RunTypeInsert, []string{"region_id"}, &emptyWhere,
			[]string{"region_id", "region_name"},
			[]string{"region_id", "region_name"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.countries", RunTypeInsert, []string{"country_id"}, &emptyWhere,
			[]string{"country_id", "country_name", "region_id"},
			[]string{"country_id", "country_name", "region_id"},
			[]*DependsOn{{Table: "public.regions", Columns: []string{"region_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.locations", RunTypeInsert, []string{"location_id"}, &emptyWhere,
			[]string{"location_id", "street_address", "country_id"},
			[]string{"location_id", "street_address", "country_id"},
			[]*DependsOn{{Table: "public.countries", Columns: []string{"country_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.jobs", RunTypeInsert, []string{"job_id"}, &emptyWhere,
			[]string{"job_id", "job_title"},
			[]string{"job_id", "job_title"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.departments", RunTypeInsert, []string{"department_id"}, &emptyWhere,
			[]string{"department_id", "department_name", "location_id"},
			[]string{"department_id", "department_name", "location_id"},
			[]*DependsOn{{Table: "public.locations", Columns: []string{"location_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.employees", RunTypeInsert, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "email", "name", "job_id", "manager_id", "department_id"},
			[]string{"employee_id", "email", "name", "job_id"},
			[]*DependsOn{{Table: "public.jobs", Columns: []string{"job_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.dependents", RunTypeInsert, []string{"dependent_id"}, &emptyWhere,
			[]string{"dependent_id", "name", "employee_id"},
			[]string{"dependent_id", "name", "employee_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.employees", RunTypeUpdate, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "manager_id"},
			[]string{"manager_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}}, []*SubsetPath{}),
		buildRunConfig("public.employees", RunTypeUpdate, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "department_id"},
			[]string{"department_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}, {Table: "public.departments", Columns: []string{"department_id"}}}, []*SubsetPath{}),
	}

	assertRunConfigs(t, dependencies, map[string]string{}, primaryKeyMap, tablesColMap, expect)
}

func Test_BuildRunConfigs_SingleTable_WithFks(t *testing.T) {
	emptyWhere := ""
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}},
		},
		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"employee_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
		"public.employees": {
			{Columns: []string{"job_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
	}
	primaryKeyMap := map[string][]string{
		"public.regions":     {"region_id"},
		"public.countries":   {"country_id"},
		"public.locations":   {"location_id"},
		"public.jobs":        {"job_id"},
		"public.departments": {"department_id"},
		"public.employees":   {"employee_id"},
		"public.dependents":  {"dependent_id"},
	}
	tablesColMap := map[string][]string{
		"public.employees": {
			"employee_id",
			"email",
			"name",
			"job_id",
			"manager_id",
			"department_id",
		},
	}

	expect := []*RunConfig{
		buildRunConfig("public.employees", RunTypeInsert, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "email", "name", "job_id", "manager_id", "department_id"},
			[]string{"employee_id", "email", "name", "job_id", "department_id"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.employees", RunTypeUpdate, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "manager_id"},
			[]string{"manager_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}}, []*SubsetPath{}),
	}

	assertRunConfigs(t, dependencies, map[string]string{}, primaryKeyMap, tablesColMap, expect)
}

func Test_BuildRunConfigs_Complex_CircularDependency(t *testing.T) {
	emptyWhere := ""
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.table_1": {
			{Columns: []string{"prev_id_1"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_4", Columns: []string{"id_4"}}},
			{Columns: []string{"next_id_1"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_2", Columns: []string{"id_2"}}},
		},
		"public.table_2": {
			{Columns: []string{"prev_id_2"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_1", Columns: []string{"id_1"}}},
			{Columns: []string{"next_id_2"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_3", Columns: []string{"id_3"}}},
		},
		"public.table_3": {
			{Columns: []string{"prev_id_3"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_2", Columns: []string{"id_2"}}},
			{Columns: []string{"next_id_3"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_4", Columns: []string{"id_4"}}},
		},
		"public.table_4": {
			{Columns: []string{"prev_id_4"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_3", Columns: []string{"id_3"}}},
			{Columns: []string{"next_id_4"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.table_1", Columns: []string{"id_1"}}},
		},
	}
	primaryKeyMap := map[string][]string{
		"public.table_1": {"id_1"},
		"public.table_2": {"id_2"},
		"public.table_3": {"id_3"},
		"public.table_4": {"id_4"},
	}
	tablesColMap := map[string][]string{
		"public.table_1": {"id_1", "name_1", "address_1", "prev_id_1", "next_id_1"},
		"public.table_2": {"id_2", "name_2", "address_2", "prev_id_2", "next_id_2"},
		"public.table_3": {"id_3", "name_3", "address_3", "prev_id_3", "next_id_3"},
		"public.table_4": {"id_4", "name_4", "address_4", "prev_id_4", "next_id_4"},
	}

	expect := []*RunConfig{
		buildRunConfig("public.table_4", RunTypeInsert, []string{"id_4"}, &emptyWhere,
			[]string{"id_4", "name_4", "address_4", "prev_id_4", "next_id_4"},
			[]string{"id_4", "name_4", "address_4"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.table_4", RunTypeUpdate, []string{"id_4"}, &emptyWhere,
			[]string{"id_4", "prev_id_4"},
			[]string{"prev_id_4"},
			[]*DependsOn{
				{Table: "public.table_4", Columns: []string{"id_4"}},
				{Table: "public.table_3", Columns: []string{"id_3"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.table_4", RunTypeUpdate, []string{"id_4"}, &emptyWhere,
			[]string{"id_4", "next_id_4"},
			[]string{"next_id_4"},
			[]*DependsOn{
				{Table: "public.table_4", Columns: []string{"id_4"}},
				{Table: "public.table_1", Columns: []string{"id_1"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.table_2", RunTypeInsert, []string{"id_2"}, &emptyWhere,
			[]string{"id_2", "name_2", "address_2", "prev_id_2", "next_id_2"},
			[]string{"id_2", "name_2", "address_2"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.table_2", RunTypeUpdate, []string{"id_2"}, &emptyWhere,
			[]string{"id_2", "prev_id_2"},
			[]string{"prev_id_2"},
			[]*DependsOn{
				{Table: "public.table_2", Columns: []string{"id_2"}},
				{Table: "public.table_1", Columns: []string{"id_1"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.table_2", RunTypeUpdate, []string{"id_2"}, &emptyWhere,
			[]string{"id_2", "next_id_2"},
			[]string{"next_id_2"},
			[]*DependsOn{
				{Table: "public.table_2", Columns: []string{"id_2"}},
				{Table: "public.table_3", Columns: []string{"id_3"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.table_1", RunTypeInsert, []string{"id_1"}, &emptyWhere,
			[]string{"id_1", "name_1", "address_1", "prev_id_1", "next_id_1"},
			[]string{"id_1", "name_1", "address_1", "prev_id_1", "next_id_1"},
			[]*DependsOn{
				{Table: "public.table_4", Columns: []string{"id_4"}},
				{Table: "public.table_2", Columns: []string{"id_2"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.table_3", RunTypeInsert, []string{"id_3"}, &emptyWhere,
			[]string{"id_3", "name_3", "address_3", "prev_id_3", "next_id_3"},
			[]string{"id_3", "name_3", "address_3", "prev_id_3", "next_id_3"},
			[]*DependsOn{
				{Table: "public.table_2", Columns: []string{"id_2"}},
				{Table: "public.table_4", Columns: []string{"id_4"}},
			}, []*SubsetPath{}),
	}

	assertRunConfigs(t, dependencies, map[string]string{}, primaryKeyMap, tablesColMap, expect)
}

func Test_BuildRunConfigs_Multiple_CircularDependency(t *testing.T) {
	emptyWhere := ""
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
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
	primaryKeyMap := map[string][]string{
		"public.a": {"id"},
		"public.b": {"id"},
		"public.c": {"id"},
	}
	tablesColMap := map[string][]string{
		"public.a": {"id", "c_id"},
		"public.b": {"id", "a_id", "ac_id"},
		"public.c": {"id", "b_id", "acb_id"},
	}

	expect := []*RunConfig{
		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"id"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"c_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.c", Columns: []string{"id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "a_id", "ac_id"},
			[]string{"id", "a_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "ac_id"},
			[]string{"ac_id"},
			[]*DependsOn{
				{Table: "public.b", Columns: []string{"id"}},
				{Table: "public.a", Columns: []string{"c_id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "b_id", "acb_id"},
			[]string{"id", "b_id"},
			[]*DependsOn{
				{Table: "public.b", Columns: []string{"id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.c", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "acb_id"},
			[]string{"acb_id"},
			[]*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"ac_id"}},
			}, []*SubsetPath{}),
	}

	assertRunConfigs(t, dependencies, map[string]string{}, primaryKeyMap, tablesColMap, expect)
}

func Test_BuildRunConfigs_CircularDependency_MultipleFksPerTable(t *testing.T) {
	emptyWhere := ""
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
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
	primaryKeyMap := map[string][]string{
		"public.a": {"id"},
		"public.b": {"id"},
		"public.c": {"id"},
	}
	tablesColMap := map[string][]string{
		"public.a": {"id", "c_id"},
		"public.b": {"id", "a_id", "ac_id"},
		"public.c": {"id", "b_id", "acb_id"},
	}

	expect := []*RunConfig{
		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"id"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"c_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.c", Columns: []string{"id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "a_id", "ac_id"},
			[]string{"id", "a_id", "ac_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.a", Columns: []string{"c_id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "b_id", "acb_id"},
			[]string{"id"},
			[]*DependsOn{}, []*SubsetPath{}),
		buildRunConfig("public.c", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "b_id"},
			[]string{"b_id"},
			[]*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"id"}},
			}, []*SubsetPath{}),
		buildRunConfig("public.c", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "acb_id"},
			[]string{"acb_id"},
			[]*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"ac_id"}},
			}, []*SubsetPath{}),
	}

	assertRunConfigs(t, dependencies, map[string]string{}, primaryKeyMap, tablesColMap, expect)
}

func Test_BuildRunConfigs_CircularDependencyNoneNullable(t *testing.T) {
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		},
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
	}
	_, err := BuildRunConfigs(dependencies, map[string]string{}, map[string][]string{}, map[string][]string{"public.a": {}, "public.b": {}}, map[string][][]string{}, map[string][][]string{})
	require.Error(t, err)
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
				{
					id:            "public.users.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"},
					runType:       RunTypeInsert,
					selectColumns: []string{"id", "name"},
					insertColumns: []string{"id", "name"},
				},
			},
			expected: true,
		},
		{
			name: "multiple nodes with no dependencies",
			configs: []*RunConfig{
				{
					id:            "public.users.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"},
					runType:       RunTypeInsert,
					selectColumns: []string{"id", "name"},
					insertColumns: []string{"id", "name"},
				},
				{
					id:            "public.products.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "products"},
					runType:       RunTypeInsert,
					selectColumns: []string{"id", "name"},
					insertColumns: []string{"id", "name"},
				},
			},
			expected: true,
		},
		{
			name: "simple dependency chain",
			configs: []*RunConfig{
				{
					id:            "public.users.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"},
					runType:       RunTypeInsert,
					primaryKeys:   []string{"id"},
					selectColumns: []string{"id"},
					insertColumns: []string{"id"},
				},
				{
					id:            "public.orders.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "orders"},
					runType:       RunTypeInsert,
					primaryKeys:   []string{"id"},
					selectColumns: []string{"id", "user_id"},
					insertColumns: []string{"user_id"},
					dependsOn:     []*DependsOn{{Table: "public.users", Columns: []string{"id"}}},
				},
			},
			expected: true,
		},
		{
			name: "circular dependency",
			configs: []*RunConfig{
				{
					id:            "public.users.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"},
					runType:       RunTypeInsert,
					primaryKeys:   []string{"id"},
					selectColumns: []string{"id"},
					insertColumns: []string{"id"},
					dependsOn:     []*DependsOn{{Table: "public.orders", Columns: []string{"user_id"}}},
				},
				{
					id:            "public.orders.insert",
					table:         sqlmanager_shared.SchemaTable{Schema: "public", Table: "orders"},
					runType:       RunTypeInsert,
					primaryKeys:   []string{"id"},
					selectColumns: []string{"user_id"},
					insertColumns: []string{"user_id"},
					dependsOn:     []*DependsOn{{Table: "public.users", Columns: []string{"id"}}},
				},
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
	subsetPaths []*SubsetPath,
) *RunConfig {
	schema, table := sqlmanager_shared.SplitTableKey(table)
	schemaTable := sqlmanager_shared.SchemaTable{
		Schema: schema,
		Table:  table,
	}
	rc := &RunConfig{
		table:         schemaTable,
		runType:       runtype,
		selectColumns: selectCols,
		insertColumns: insertCols,
		primaryKeys:   pks,
		whereClause:   where,
		dependsOn:     dependsOn,
		subsetPaths:   subsetPaths,
	}
	return rc
}

func assertRunConfigs(t *testing.T, dependencies map[string][]*sqlmanager_shared.ForeignConstraint, subsets map[string]string, primaryKeyMap map[string][]string, tableColsMap map[string][]string, expect []*RunConfig) {
	actual, err := BuildRunConfigs(dependencies, subsets, primaryKeyMap, tableColsMap, map[string][][]string{}, map[string][][]string{})
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

		// Test subset paths
		assert.Len(t, acutalConfig.SubsetPaths(), len(e.SubsetPaths()),
			"Subset paths count mismatch for table %s (type: %s) - expected %d but got %d",
			e.Table(), e.RunType(), len(e.SubsetPaths()), len(acutalConfig.SubsetPaths()))

		if len(e.SubsetPaths()) > 0 {
			// Compare each subset path
			for _, expectedPath := range e.SubsetPaths() {
				found := false
				for _, actualPath := range acutalConfig.SubsetPaths() {
					if expectedPath.Root == actualPath.Root && expectedPath.Subset == actualPath.Subset {
						assert.Len(t, actualPath.JoinSteps, len(expectedPath.JoinSteps),
							"Join steps count mismatch for table %s (type: %s) in subset path to %s",
							e.Table(), e.RunType(), expectedPath.Root)

						// Compare join steps if lengths match
						if len(actualPath.JoinSteps) == len(expectedPath.JoinSteps) {
							for j, expectedStep := range expectedPath.JoinSteps {
								actualStep := actualPath.JoinSteps[j]
								assert.Equal(t, expectedStep.FromKey, actualStep.FromKey,
									"FromKey mismatch in join step %d for table %s (type: %s)",
									j, e.Table(), e.RunType())
								assert.Equal(t, expectedStep.ToKey, actualStep.ToKey,
									"ToKey mismatch in join step %d for table %s (type: %s)",
									j, e.Table(), e.RunType())
							}
						}
						found = true
						break
					}
				}
				assert.True(t, found, "Could not find matching subset path for Root=%s, Subset=%s in table %s (type: %s)",
					expectedPath.Root, expectedPath.Subset, e.Table(), e.RunType())
			}
		}
	}
}
