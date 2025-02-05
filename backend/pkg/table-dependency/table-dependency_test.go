package tabledependency

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func Test_GetRunConfigs_NoSubset_SingleCycle(t *testing.T) {
	where := ""
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
			subsets: map[string]string{},
			expect: []*RunConfig{
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
			subsets: map[string]string{},
			expect: []*RunConfig{
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
			},
		},
		{
			name: "Self Referencing Cycle",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"a_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
			tableColsMap: map[string][]string{
				"public.a": {"id", "a_id", "other"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where, []string{"id", "a_id", "other"}, []string{"id", "other"}, []*DependsOn{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where, []string{"id", "a_id"}, []string{"a_id"}, []*DependsOn{
					{Table: "public.a", Columns: []string{"id"}},
				}),
			},
		},
		{
			name: "Double Self Referencing Cycle",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"a_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
					{Columns: []string{"aa_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
			tableColsMap: map[string][]string{
				"public.a": {"id", "a_id", "aa_id", "other"},
			},
			primaryKeyMap: map[string][]string{
				"public.a": {"id"},
			},
			subsets: map[string]string{},
			expect: []*RunConfig{
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &where,
					[]string{"id", "a_id", "aa_id", "other"},
					[]string{"id", "other"},
					[]*DependsOn{}),
				buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &where,
					[]string{"id", "a_id", "aa_id"},
					[]string{"a_id", "aa_id"},
					[]*DependsOn{
						{Table: "public.a", Columns: []string{"id"}},
					}),
			},
		},
		{
			name: "Single Cycle Composite Foreign Keys",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
					{Columns: []string{"cc_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"other_id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
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
			},
		},
		{
			name: "Single Cycle Composite Foreign Keys Nullable",
			dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
				"public.a": {
					{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
				},
				"public.b": {
					{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
					{Columns: []string{"cc_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"other_id"}}},
				},
				"public.c": {
					{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
				},
			},
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			for _, e := range tt.expect {
				acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
				require.NotNil(t, acutalConfig)
				require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
				require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
				require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
				require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
				require.Equal(t, e.WhereClause(), e.WhereClause())
			}
		})
	}
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
			actual, err := GetRunConfigs(tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			for _, e := range tt.expect {
				acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
				require.NotNil(t, acutalConfig)
				require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
				require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
				require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
				require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
				require.Equal(t, e.WhereClause(), e.WhereClause())
			}
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
		// {
		// 	name: "Multi Table Dependencies",
		// 	dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
		// 		"public.a": {
		// 			{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		// 		},
		// 		"public.b": {
		// 			{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		// 			{Columns: []string{"d_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.d", Columns: []string{"id"}}},
		// 		},
		// 		"public.c": {
		// 			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		// 		},
		// 		"public.d": {
		// 			{Columns: []string{"e_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.e", Columns: []string{"id"}}},
		// 		},
		// 		"public.e": {
		// 			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		// 		},
		// 	},
		// 	tableColsMap: map[string][]string{
		// 		"public.a": {"id", "b_id"},
		// 		"public.b": {"id", "c_id", "d_id", "other_id"},
		// 		"public.c": {"id", "a_id"},
		// 		"public.d": {"id", "e_id"},
		// 		"public.e": {"id", "b_id"},
		// 	},
		// 	primaryKeyMap: map[string][]string{
		// 		"public.a": {"id"},
		// 		"public.b": {"id"},
		// 		"public.c": {"id"},
		// 		"public.e": {"id"},
		// 		"public.d": {"id"},
		// 	},
		// 	subsets: map[string]string{},
		// 	expect: []*RunConfig{
		// 		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id", "other_id"}, []string{"id", "other_id"}, []*DependsOn{}),
		// 		buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id"}, []string{"c_id", "d_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.d", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "e_id"}, []string{"id", "e_id"}, []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.e", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
		// 	},
		// },
		// {
		// 	name: "Multi Table Dependencies Complex Foreign Keys",
		// 	dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
		// 		"public.a": {
		// 			{Columns: []string{"b_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		// 		},
		// 		"public.b": {
		// 			{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		// 			{Columns: []string{"d_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.d", Columns: []string{"id"}}},
		// 		},
		// 		"public.c": {
		// 			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		// 		},
		// 		"public.d": {
		// 			{Columns: []string{"e_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.e", Columns: []string{"id"}}},
		// 		},
		// 		"public.e": {
		// 			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		// 		},
		// 	},
		// 	tableColsMap: map[string][]string{
		// 		"public.a": {"id", "b_id"},
		// 		"public.b": {"id", "c_id", "d_id", "other_id"},
		// 		"public.c": {"id", "a_id"},
		// 		"public.d": {"id", "e_id"},
		// 		"public.e": {"id", "b_id"},
		// 	},
		// 	primaryKeyMap: map[string][]string{
		// 		"public.a": {"id"},
		// 		"public.b": {"id"},
		// 		"public.c": {"id"},
		// 		"public.e": {"id"},
		// 		"public.d": {"id"},
		// 	},
		// 	subsets: map[string]string{},
		// 	expect: []*RunConfig{
		// 		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id"}, []*DependsOn{}),
		// 		buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"b_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}, {Table: "public.b", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "d_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "d_id"}, []string{"d_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.d", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "e_id"}, []string{"id", "e_id"}, []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.e", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
		// 	},
		// },
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
		// {
		// 	name: "Multi Table Dependencies Self Referencing Circular Dependency Simple",
		// 	dependencies: map[string][]*sqlmanager_shared.ForeignConstraint{
		// 		"public.a": {
		// 			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		// 		},
		// 		"public.b": {
		// 			{Columns: []string{"c_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.c", Columns: []string{"id"}}},
		// 			{Columns: []string{"bb_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		// 		},
		// 		"public.c": {
		// 			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		// 		},
		// 	},
		// 	tableColsMap: map[string][]string{
		// 		"public.a": {"id", "b_id"},
		// 		"public.b": {"id", "c_id", "bb_id", "other_id"},
		// 		"public.c": {"id", "a_id"},
		// 	},
		// 	primaryKeyMap: map[string][]string{
		// 		"public.a": {"id"},
		// 		"public.b": {"id"},
		// 		"public.c": {"id"},
		// 	},
		// 	subsets: map[string]string{},
		// 	expect: []*RunConfig{
		// 		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "bb_id", "other_id"}, []string{"id", "other_id"}, []*DependsOn{}),
		// 		buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere, []string{"id", "c_id", "bb_id"}, []string{"c_id", "bb_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}, {Table: "public.c", Columns: []string{"id"}}}),
		// 		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "a_id"}, []string{"id", "a_id"}, []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}),
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			for _, e := range tt.expect {
				acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
				require.NotNil(t, acutalConfig)
				require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
				require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
				require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
				require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
				require.Equal(t, e.WhereClause(), e.WhereClause())
			}
		})
	}
}

func Test_GetRunConfigs_NoSubset_NoCycle(t *testing.T) {
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
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
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
				buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id"}, []string{"id"}, []*DependsOn{}),
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{{Table: "public.c", Columns: []string{"id"}}}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
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
				buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "c_id", "other_id"}, []string{"id", "c_id", "other_id"}, []*DependsOn{}),
				buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere, []string{"id", "b_id"}, []string{"id", "b_id"}, []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetRunConfigs(tt.dependencies, tt.subsets, tt.primaryKeyMap, tt.tableColsMap)
			require.NoError(t, err)
			for _, e := range tt.expect {
				acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
				require.NotNil(t, acutalConfig)
				require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
				require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
				require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
				require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
				require.Equal(t, e.WhereClause(), e.WhereClause())
			}
		})
	}
}

func Test_GetRunConfigs_CompositeKey(t *testing.T) {
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
			[]*DependsOn{{Table: "public.department", Columns: []string{"department_id"}}}),
		buildRunConfig("public.department", RunTypeInsert, []string{"department_id"}, &emptyWhere,
			[]string{"department_id", "department_name", "location"},
			[]string{"department_id", "department_name", "location"},
			[]*DependsOn{}),
		buildRunConfig("public.projects", RunTypeInsert, []string{"project_id"}, &emptyWhere,
			[]string{"project_id", "project_name", "start_date", "end_date", "responsible_employee_id", "responsible_department_id"},
			[]string{"project_id", "project_name", "start_date", "end_date", "responsible_employee_id", "responsible_department_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id", "department_id"}}}),
	}

	actual, err := GetRunConfigs(dependencies, map[string]string{}, primaryKeyMap, tablesColMap)

	require.NoError(t, err)
	for _, e := range expect {
		acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
		require.NotNil(t, acutalConfig)
		require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
		require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
		require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
		require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
		require.Equal(t, e.WhereClause(), e.WhereClause())
	}
}

func Test_GetRunConfigs_HumanResources(t *testing.T) {
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
			[]*DependsOn{}),
		buildRunConfig("public.countries", RunTypeInsert, []string{"country_id"}, &emptyWhere,
			[]string{"country_id", "country_name", "region_id"},
			[]string{"country_id", "country_name", "region_id"},
			[]*DependsOn{{Table: "public.regions", Columns: []string{"region_id"}}}),
		buildRunConfig("public.locations", RunTypeInsert, []string{"location_id"}, &emptyWhere,
			[]string{"location_id", "street_address", "country_id"},
			[]string{"location_id", "street_address", "country_id"},
			[]*DependsOn{{Table: "public.countries", Columns: []string{"country_id"}}}),
		buildRunConfig("public.jobs", RunTypeInsert, []string{"job_id"}, &emptyWhere,
			[]string{"job_id", "job_title"},
			[]string{"job_id", "job_title"},
			[]*DependsOn{}),
		buildRunConfig("public.departments", RunTypeInsert, []string{"department_id"}, &emptyWhere,
			[]string{"department_id", "department_name", "location_id"},
			[]string{"department_id", "department_name", "location_id"},
			[]*DependsOn{{Table: "public.locations", Columns: []string{"location_id"}}}),
		buildRunConfig("public.employees", RunTypeInsert, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "email", "name", "job_id", "manager_id", "department_id"},
			[]string{"employee_id", "email", "name", "job_id", "department_id"},
			[]*DependsOn{{Table: "public.departments", Columns: []string{"department_id"}}, {Table: "public.jobs", Columns: []string{"job_id"}}}),
		buildRunConfig("public.dependents", RunTypeInsert, []string{"dependent_id"}, &emptyWhere,
			[]string{"dependent_id", "name", "employee_id"},
			[]string{"dependent_id", "name", "employee_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}}),
		buildRunConfig("public.employees", RunTypeUpdate, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "manager_id"},
			[]string{"manager_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}}),
	}

	actual, err := GetRunConfigs(dependencies, map[string]string{}, primaryKeyMap, tablesColMap)
	require.NoError(t, err)
	for _, e := range expect {
		acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
		require.NotNil(t, acutalConfig)
		require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
		require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
		require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
		require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
		require.Equal(t, e.WhereClause(), e.WhereClause())
	}
}

func Test_GetRunConfigs_SingleTable_WithFks(t *testing.T) {
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
			[]*DependsOn{}),
		buildRunConfig("public.employees", RunTypeUpdate, []string{"employee_id"}, &emptyWhere,
			[]string{"employee_id", "manager_id"},
			[]string{"manager_id"},
			[]*DependsOn{{Table: "public.employees", Columns: []string{"employee_id"}}}),
	}

	actual, err := GetRunConfigs(dependencies, map[string]string{}, primaryKeyMap, tablesColMap)
	require.NoError(t, err)
	for _, e := range expect {
		acutalConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
		require.NotNil(t, acutalConfig)
		require.ElementsMatch(t, e.InsertColumns(), acutalConfig.InsertColumns())
		require.ElementsMatch(t, e.SelectColumns(), acutalConfig.SelectColumns())
		require.ElementsMatch(t, e.DependsOn(), acutalConfig.DependsOn())
		require.ElementsMatch(t, e.PrimaryKeys(), acutalConfig.PrimaryKeys())
		require.Equal(t, e.WhereClause(), e.WhereClause())
	}
}

func Test_GetRunConfigs_Complex_CircularDependency(t *testing.T) {
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
			[]*DependsOn{}),
		buildRunConfig("public.table_4", RunTypeUpdate, []string{"id_4"}, &emptyWhere,
			[]string{"id_4", "prev_id_4", "next_id_4"},
			[]string{"prev_id_4", "next_id_4"},
			[]*DependsOn{
				{Table: "public.table_4", Columns: []string{"id_4"}},
				{Table: "public.table_3", Columns: []string{"id_3"}},
				{Table: "public.table_1", Columns: []string{"id_1"}},
			}),
		buildRunConfig("public.table_2", RunTypeInsert, []string{"id_2"}, &emptyWhere,
			[]string{"id_2", "name_2", "address_2", "prev_id_2", "next_id_2"},
			[]string{"id_2", "name_2", "address_2"},
			[]*DependsOn{}),
		buildRunConfig("public.table_2", RunTypeUpdate, []string{"id_2"}, &emptyWhere,
			[]string{"id_2", "prev_id_2", "next_id_2"},
			[]string{"prev_id_2", "next_id_2"},
			[]*DependsOn{
				{Table: "public.table_2", Columns: []string{"id_2"}},
				{Table: "public.table_1", Columns: []string{"id_1"}},
				{Table: "public.table_3", Columns: []string{"id_3"}},
			}),
		buildRunConfig("public.table_1", RunTypeInsert, []string{"id_1"}, &emptyWhere,
			[]string{"id_1", "name_1", "address_1", "prev_id_1", "next_id_1"},
			[]string{"id_1", "name_1", "address_1", "prev_id_1", "next_id_1"},
			[]*DependsOn{
				{Table: "public.table_4", Columns: []string{"id_4"}},
				{Table: "public.table_2", Columns: []string{"id_2"}},
			}),
		buildRunConfig("public.table_3", RunTypeInsert, []string{"id_3"}, &emptyWhere,
			[]string{"id_3", "name_3", "address_3", "prev_id_3", "next_id_3"},
			[]string{"id_3", "name_3", "address_3", "prev_id_3", "next_id_3"},
			[]*DependsOn{
				{Table: "public.table_2", Columns: []string{"id_2"}},
				{Table: "public.table_4", Columns: []string{"id_4"}},
			}),
	}

	actual, err := GetRunConfigs(dependencies, map[string]string{}, primaryKeyMap, tablesColMap)
	require.NoError(t, err)
	for _, e := range expect {
		actualConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual) // Adjust getConfigByTableAndType according to your actual utility functions
		require.NotNil(t, actualConfig)
		require.ElementsMatch(t, e.InsertColumns(), actualConfig.InsertColumns())
		require.ElementsMatch(t, e.SelectColumns(), actualConfig.SelectColumns())
		require.ElementsMatch(t, e.DependsOn(), actualConfig.DependsOn())
		require.ElementsMatch(t, e.PrimaryKeys(), actualConfig.PrimaryKeys())
		require.Equal(t, e.WhereClause(), e.WhereClause())
	}
}

func Test_GetRunConfigs_Multiple_CircularDependency(t *testing.T) {
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
		// First insert a with just its primary key and no foreign keys
		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"id"},
			[]*DependsOn{}),
		// Then update a's nullable foreign keys after dependencies exist
		buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"c_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.c", Columns: []string{"id"}},
			}),
		// Insert b with required a_id reference
		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "a_id", "ac_id"},
			[]string{"id", "a_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
			}),
		// Update b's nullable ac_id after all dependencies exist
		buildRunConfig("public.b", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "ac_id"},
			[]string{"ac_id"},
			[]*DependsOn{
				{Table: "public.b", Columns: []string{"id"}},
				{Table: "public.a", Columns: []string{"c_id"}},
			}),
		// Insert c with required b_id reference
		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "b_id", "acb_id"},
			[]string{"id", "b_id"},
			[]*DependsOn{
				{Table: "public.b", Columns: []string{"id"}},
			}),
		// Update c's nullable acb_id after all dependencies exist
		buildRunConfig("public.c", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "acb_id"},
			[]string{"acb_id"},
			[]*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"ac_id"}},
			}),
	}

	actual, err := GetRunConfigs(dependencies, map[string]string{}, primaryKeyMap, tablesColMap)
	require.NoError(t, err)
	for _, e := range expect {
		actualConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
		require.NotNil(t, actualConfig, "expected config for table %s and type %s not found", e.Table(), e.RunType())
		require.ElementsMatch(t, e.InsertColumns(), actualConfig.InsertColumns(), "insert columns mismatch for table %s", e.Table())
		require.ElementsMatch(t, e.SelectColumns(), actualConfig.SelectColumns(), "select columns mismatch for table %s", e.Table())
		require.ElementsMatch(t, e.DependsOn(), actualConfig.DependsOn(), "depends on mismatch for table %s", e.Table())
		require.ElementsMatch(t, e.PrimaryKeys(), actualConfig.PrimaryKeys(), "primary keys mismatch for table %s", e.Table())
		require.Equal(t, e.WhereClause(), e.WhereClause(), "where clause mismatch for table %s", e.Table())
	}
}

func Test_GetRunConfigs_CircularDependency_MultipleFksPerTable(t *testing.T) {
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
		// First insert a with just its primary key and no foreign keys
		buildRunConfig("public.a", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"id"},
			[]*DependsOn{}),
		// Then update a's nullable foreign keys after dependencies exist
		buildRunConfig("public.a", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "c_id"},
			[]string{"c_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id"}},
				{Table: "public.c", Columns: []string{"id"}},
			}),
		// Insert b with required a_id reference
		buildRunConfig("public.b", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "a_id", "ac_id"},
			[]string{"id", "a_id", "ac_id"},
			[]*DependsOn{
				{Table: "public.a", Columns: []string{"id", "c_id"}},
			}),
		// Insert c with required b_id reference
		buildRunConfig("public.c", RunTypeInsert, []string{"id"}, &emptyWhere,
			[]string{"id", "b_id", "acb_id"},
			[]string{"id"},
			[]*DependsOn{}),
		// Update c's nullable acb_id after all dependencies exist
		buildRunConfig("public.c", RunTypeUpdate, []string{"id"}, &emptyWhere,
			[]string{"id", "b_id", "acb_id"},
			[]string{"b_id", "acb_id"},
			[]*DependsOn{
				{Table: "public.c", Columns: []string{"id"}},
				{Table: "public.b", Columns: []string{"id", "ac_id"}},
			}),
	}

	actual, err := GetRunConfigs(dependencies, map[string]string{}, primaryKeyMap, tablesColMap)
	require.NoError(t, err)
	for _, e := range expect {
		actualConfig := getConfigByTableAndType(e.Table(), e.RunType(), actual)
		require.NotNil(t, actualConfig, "expected config for table %s and type %s not found", e.Table(), e.RunType())
		require.ElementsMatch(t, e.InsertColumns(), actualConfig.InsertColumns(), "insert columns mismatch for table %s", e.Table())
		require.ElementsMatch(t, e.SelectColumns(), actualConfig.SelectColumns(), "select columns mismatch for table %s", e.Table())
		require.ElementsMatch(t, e.DependsOn(), actualConfig.DependsOn(), "depends on mismatch for table %s", e.Table())
		require.ElementsMatch(t, e.PrimaryKeys(), actualConfig.PrimaryKeys(), "primary keys mismatch for table %s", e.Table())
		require.Equal(t, e.WhereClause(), e.WhereClause(), "where clause mismatch for table %s", e.Table())
	}
}

func Test_GetRunConfigs_CircularDependencyNoneNullable(t *testing.T) {
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.a": {
			{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.b", Columns: []string{"id"}}},
		},
		"public.b": {
			{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.a", Columns: []string{"id"}}},
		},
	}
	_, err := GetRunConfigs(dependencies, map[string]string{}, map[string][]string{}, map[string][]string{"public.a": {}, "public.b": {}})
	require.Error(t, err)
}

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
				buildRunConfig("users", "initial", nil, nil, []string{"id", "name"}, []string{"id", "name"}, nil),
			},
			expected: true,
		},
		{
			name: "multiple nodes with no dependencies",
			configs: []*RunConfig{
				buildRunConfig("users", "initial", nil, nil, []string{"id", "name"}, []string{"id", "name"}, nil),
				buildRunConfig("products", "initial", nil, nil, []string{"id", "name"}, []string{"id", "name"}, nil),
			},
			expected: true,
		},
		{
			name: "simple dependency chain",
			configs: []*RunConfig{
				buildRunConfig("users", "initial", nil, nil, []string{"id"}, []string{"id"}, nil),
				buildRunConfig("orders", "initial", nil, nil, []string{"id", "user_id"}, []string{"user_id"}, []*DependsOn{{Table: "users", Columns: []string{"id"}}}),
			},
			expected: true,
		},
		{
			name: "circular dependency",
			configs: []*RunConfig{
				buildRunConfig("users", "initial", nil, nil, []string{"id"}, []string{"id"}, []*DependsOn{{Table: "orders", Columns: []string{"user_id"}}}),
				buildRunConfig("orders", "initial", nil, nil, []string{"user_id"}, []string{"user_id"}, []*DependsOn{{Table: "users", Columns: []string{"id"}}}),
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

func TestFilterConfigsWithWhereClause(t *testing.T) {
	t.Run("Basic filtering", func(t *testing.T) {
		t.Parallel()
		whereClause := "id > 10"
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table1", RunTypeUpdate, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table2", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table2", RunTypeUpdate, nil, nil, nil, nil, nil),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 3)
		require.Contains(t, filtered, configs[0])
		require.Contains(t, filtered, configs[2])
		require.Contains(t, filtered, configs[3])
	})

	t.Run("Circular dependency", func(t *testing.T) {
		t.Parallel()
		whereClause := "id > 10"
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table1", RunTypeUpdate, nil, &whereClause, nil, nil, []*DependsOn{{Table: "table1"}}),
			buildRunConfig("table2", RunTypeInsert, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}}),
			buildRunConfig("table2", RunTypeUpdate, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}, {Table: "table2"}}),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 2)
		require.Contains(t, filtered, configs[0])
		require.Contains(t, filtered, configs[2])
	})

	t.Run("Self-reference", func(t *testing.T) {
		t.Parallel()
		whereClause := "id > 10"
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table1", RunTypeUpdate, nil, &whereClause, nil, nil, []*DependsOn{{Table: "table1"}}),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 1)
		require.Contains(t, filtered, configs[0])
	})

	t.Run("Complex dependency chain", func(t *testing.T) {
		t.Parallel()
		whereClause := "id > 10"
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table1", RunTypeUpdate, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table2", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table2", RunTypeUpdate, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}}),
			buildRunConfig("table3", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table3", RunTypeUpdate, nil, nil, nil, nil, []*DependsOn{{Table: "table2"}}),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 3)
		require.Contains(t, filtered, configs[0])
		require.Contains(t, filtered, configs[2])
		require.Contains(t, filtered, configs[4])
	})

	t.Run("Mixed where clauses", func(t *testing.T) {
		t.Parallel()
		whereClause1 := "id > 10"
		whereClause2 := "name LIKE 'test%'"
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, &whereClause1, nil, nil, nil),
			buildRunConfig("table1", RunTypeUpdate, nil, &whereClause1, nil, nil, nil),
			buildRunConfig("table2", RunTypeInsert, nil, &whereClause2, nil, nil, nil),
			buildRunConfig("table2", RunTypeUpdate, nil, &whereClause2, nil, nil, nil),
			buildRunConfig("table3", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table3", RunTypeUpdate, nil, nil, nil, nil, nil),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 4)
		require.Contains(t, filtered, configs[0])
		require.Contains(t, filtered, configs[2])
		require.Contains(t, filtered, configs[4])
		require.Contains(t, filtered, configs[5])
	})

	t.Run("All inserts, no updates", func(t *testing.T) {
		t.Parallel()
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table2", RunTypeInsert, nil, nil, nil, nil, nil),
			buildRunConfig("table3", RunTypeInsert, nil, nil, nil, nil, nil),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 3)
		require.Equal(t, configs, filtered)
	})

	t.Run("Complex scenario with multiple dependencies", func(t *testing.T) {
		t.Parallel()
		whereClause := "id > 10"
		configs := []*RunConfig{
			buildRunConfig("table1", RunTypeInsert, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table1", RunTypeUpdate, nil, &whereClause, nil, nil, nil),
			buildRunConfig("table2", RunTypeInsert, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}, {Table: "table3"}}),
			buildRunConfig("table2", RunTypeUpdate, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}, {Table: "table3"}, {Table: "table2"}}),
			buildRunConfig("table3", RunTypeInsert, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}}),
			buildRunConfig("table3", RunTypeUpdate, nil, nil, nil, nil, []*DependsOn{{Table: "table1"}, {Table: "table3"}}),
			buildRunConfig("table4", RunTypeInsert, nil, nil, nil, nil, []*DependsOn{{Table: "table2"}, {Table: "table3"}}),
			buildRunConfig("table4", RunTypeUpdate, nil, nil, nil, nil, []*DependsOn{{Table: "table2"}, {Table: "table3"}, {Table: "table4"}}),
		}

		filtered := filterConfigsWithWhereClause(configs)

		require.Len(t, filtered, 4)
		require.Contains(t, filtered, configs[0])
		require.Contains(t, filtered, configs[2])
		require.Contains(t, filtered, configs[4])
		require.Contains(t, filtered, configs[6])
	})
}

func getConfigByTableAndType(table string, runtype RunType, configs []*RunConfig) *RunConfig {
	for _, c := range configs {
		if c.Table() == table && c.RunType() == runtype {
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
	rc := newRunConfig(table, runtype, pks, where)
	rc.appendInsertColumns(insertCols...)
	rc.appendSelectColumns(selectCols...)
	for _, dp := range dependsOn {
		rc.appendDependsOn(dp.Table, dp.Columns)
	}
	return rc
}
