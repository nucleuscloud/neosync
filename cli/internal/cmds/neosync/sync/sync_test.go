package sync_cmd

import (
	"testing"

	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/require"
)

func Test_groupConfigsByDependency(t *testing.T) {
	tests := []struct {
		name    string
		configs []*benthosConfigResponse
		expect  [][]*benthosConfigResponse
	}{
		{
			name: "No dependencies",
			configs: []*benthosConfigResponse{
				{Name: "public.users", DependsOn: []*tabledependency.DependsOn{}, Table: "public.users", Columns: []string{"id", "email"}},
				{Name: "public.accounts", DependsOn: []*tabledependency.DependsOn{}, Table: "public.accounts", Columns: []string{"id", "name"}},
			},
			expect: [][]*benthosConfigResponse{
				{
					{Name: "public.users", DependsOn: []*tabledependency.DependsOn{}, Table: "public.users", Columns: []string{"id", "email"}},
					{Name: "public.accounts", DependsOn: []*tabledependency.DependsOn{}, Table: "public.accounts", Columns: []string{"id", "name"}},
				},
			},
		},
		{
			name: "Multiple dependencies",
			configs: []*benthosConfigResponse{
				{Name: "public.users", DependsOn: []*tabledependency.DependsOn{}, Table: "public.users", Columns: []string{"id", "email"}},
				{Name: "public.accounts", DependsOn: []*tabledependency.DependsOn{}, Table: "public.accounts", Columns: []string{"id", "name"}},
				{Name: "public.jobs", DependsOn: []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}}, Table: "public.jobs", Columns: []string{"id", "user_id"}},
				{Name: "public.regions", DependsOn: []*tabledependency.DependsOn{{Table: "public.accounts", Columns: []string{"id"}}}, Table: "public.regions", Columns: []string{"id", "account_id"}},
				{Name: "public.tasks", DependsOn: []*tabledependency.DependsOn{{Table: "public.jobs", Columns: []string{"id"}}}, Table: "public.tasks", Columns: []string{"id", "job_id"}},
			},
			expect: [][]*benthosConfigResponse{
				{
					{Name: "public.users", DependsOn: []*tabledependency.DependsOn{}, Table: "public.users", Columns: []string{"id", "email"}},
					{Name: "public.accounts", DependsOn: []*tabledependency.DependsOn{}, Table: "public.accounts", Columns: []string{"id", "name"}},
				},
				{
					{Name: "public.jobs", DependsOn: []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}}, Table: "public.jobs", Columns: []string{"id", "user_id"}},
					{Name: "public.regions", DependsOn: []*tabledependency.DependsOn{{Table: "public.accounts", Columns: []string{"id"}}}, Table: "public.regions", Columns: []string{"id", "account_id"}},
				},
				{
					{Name: "public.tasks", DependsOn: []*tabledependency.DependsOn{{Table: "public.jobs", Columns: []string{"id"}}}, Table: "public.tasks", Columns: []string{"id", "job_id"}},
				},
			},
		},
		{
			name: "Simple dependencies",
			configs: []*benthosConfigResponse{
				{Name: "public.users", DependsOn: []*tabledependency.DependsOn{}, Table: "public.users", Columns: []string{"id", "email"}},
				{Name: "public.accounts", DependsOn: []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}}, Table: "public.accounts", Columns: []string{"id", "user_id"}},
				{Name: "public.jobs", DependsOn: []*tabledependency.DependsOn{{Table: "public.accounts", Columns: []string{"id"}}}, Table: "public.jobs", Columns: []string{"id", "account_id"}},
				{Name: "public.regions", DependsOn: []*tabledependency.DependsOn{{Table: "public.jobs", Columns: []string{"id"}}}, Table: "public.regions", Columns: []string{"id", "job_id"}},
				{Name: "public.tasks", DependsOn: []*tabledependency.DependsOn{{Table: "public.regions", Columns: []string{"id"}}}, Table: "public.tasks", Columns: []string{"id", "region_id"}},
			},
			expect: [][]*benthosConfigResponse{
				{
					{Name: "public.users", DependsOn: []*tabledependency.DependsOn{}, Table: "public.users", Columns: []string{"id", "email"}},
				},
				{
					{Name: "public.accounts", DependsOn: []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}}, Table: "public.accounts", Columns: []string{"id", "user_id"}},
				},
				{
					{Name: "public.jobs", DependsOn: []*tabledependency.DependsOn{{Table: "public.accounts", Columns: []string{"id"}}}, Table: "public.jobs", Columns: []string{"id", "account_id"}},
				},
				{
					{Name: "public.regions", DependsOn: []*tabledependency.DependsOn{{Table: "public.jobs", Columns: []string{"id"}}}, Table: "public.regions", Columns: []string{"id", "job_id"}},
				},
				{
					{Name: "public.tasks", DependsOn: []*tabledependency.DependsOn{{Table: "public.regions", Columns: []string{"id"}}}, Table: "public.tasks", Columns: []string{"id", "region_id"}},
				},
			},
		},
		{
			name: "Circular dependencies",
			configs: []*benthosConfigResponse{
				{Name: "public.a", DependsOn: []*tabledependency.DependsOn{}, Table: "public.a", Columns: []string{"id"}},
				{Name: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}, Table: "public.b", Columns: []string{"id", "c_id"}},
				{Name: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}, Table: "public.c", Columns: []string{"id", "a_id"}},
				{Name: "public.a.update", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}, Table: "public.a", Columns: []string{"b_id"}},
			},
			expect: [][]*benthosConfigResponse{
				{
					{Name: "public.a", DependsOn: []*tabledependency.DependsOn{}, Table: "public.a", Columns: []string{"id"}},
				},
				{
					{Name: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}, Table: "public.c", Columns: []string{"id", "a_id"}},
				},
				{
					{Name: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}, Table: "public.b", Columns: []string{"id", "c_id"}},
				},
				{
					{Name: "public.a.update", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}, Table: "public.a", Columns: []string{"b_id"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := groupConfigsByDependency(tt.configs, nil)
			require.Len(t, groups, len(tt.expect))
			for i, group := range groups {
				require.Equal(t, len(group), len(tt.expect[i]))
				expectedConfigMap := map[string]*benthosConfigResponse{}
				for _, cfg := range tt.expect[i] {
					expectedConfigMap[cfg.Name] = cfg
				}
				for _, cfg := range group {
					expect := expectedConfigMap[cfg.Name]
					require.NotNil(t, expect)
					require.ElementsMatch(t, cfg.DependsOn, expect.DependsOn)
				}
			}
		})
	}
}

func Test_groupConfigsByDependency_Error(t *testing.T) {
	configs := []*benthosConfigResponse{
		{Name: "public.a", DependsOn: []*tabledependency.DependsOn{{Table: "public.b", Columns: []string{"id"}}}, Table: "public.a", Columns: []string{"id"}},
		{Name: "public.b", DependsOn: []*tabledependency.DependsOn{{Table: "public.c", Columns: []string{"id"}}}, Table: "public.b", Columns: []string{"id", "c_id"}},
		{Name: "public.c", DependsOn: []*tabledependency.DependsOn{{Table: "public.a", Columns: []string{"id"}}}, Table: "public.c", Columns: []string{"id", "a_id"}},
	}
	groups := groupConfigsByDependency(configs, nil)
	require.Nil(t, groups)
}

func Test_buildPlainInsertArgs(t *testing.T) {
	require.Empty(t, buildPlainInsertArgs(nil))
	require.Empty(t, buildPlainInsertArgs([]string{}))
	require.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), `root = [this."foo", this."bar", this."baz"]`)
}
