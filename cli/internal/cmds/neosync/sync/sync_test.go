package sync_cmd

import (
	"math"
	"testing"

	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/assert"
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
			groups := groupConfigsByDependency(tt.configs)
			assert.Len(t, groups, len(tt.expect))
			for i, group := range groups {
				assert.Equal(t, len(group), len(tt.expect[i]))
				for j, cfg := range group {
					assert.Equal(t, cfg.Name, tt.expect[i][j].Name)
					assert.ElementsMatch(t, cfg.DependsOn, tt.expect[i][j].DependsOn)
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
	groups := groupConfigsByDependency(configs)
	assert.Nil(t, groups)
}

func Test_buildPlainInsertArgs(t *testing.T) {
	assert.Empty(t, buildPlainInsertArgs(nil))
	assert.Empty(t, buildPlainInsertArgs([]string{}))
	assert.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), "root = [this.foo, this.bar, this.baz]")
}

func Test_clampInt(t *testing.T) {
	assert.Equal(t, clampInt(0, 1, 2), 1)
	assert.Equal(t, clampInt(1, 1, 2), 1)
	assert.Equal(t, clampInt(2, 1, 2), 2)
	assert.Equal(t, clampInt(3, 1, 2), 2)
	assert.Equal(t, clampInt(1, 1, 1), 1)

	assert.Equal(t, clampInt(1, 3, 2), 3, "low is evaluated first, order is relevant")

}

func Test_computeMaxPgBatchCount(t *testing.T) {
	assert.Equal(t, computeMaxPgBatchCount(65535), 1)
	assert.Equal(t, computeMaxPgBatchCount(65536), 1, "anything over max should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(math.MaxInt), 1, "anything over pgmax should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(1), 65535)
	assert.Equal(t, computeMaxPgBatchCount(0), 65535)
}

func Test_buildPostgresInsertQuery(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		columns  []string
		expected string
	}{
		{"Single Column", "users", []string{"name"}, "INSERT INTO users (name) VALUES ($1);"},
		{"Multiple Columns", "users", []string{"name", "email"}, "INSERT INTO users (name, email) VALUES ($1, $2);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildPostgresInsertQuery(tt.table, tt.columns)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildPostgresUpdateQuery(t *testing.T) {
	tests := []struct {
		name        string
		table       string
		columns     []string
		primaryKeys []string
		expected    string
	}{
		{"Single Column", "users", []string{"name"}, []string{"id"}, "UPDATE users SET name = $1 WHERE id = $2;"},
		{"Multiple Primary Keys", "users", []string{"name", "email"}, []string{"id", "other"}, "UPDATE users SET name = $1, email = $2 WHERE id = $3 AND other = $4;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildPostgresUpdateQuery(tt.table, tt.columns, tt.primaryKeys)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildMysqlInsertQuery(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		columns  []string
		expected string
	}{
		{"Single Column", "users", []string{"name"}, "INSERT INTO users (name) VALUES (?);"},
		{"Multiple Columns", "users", []string{"name", "email"}, "INSERT INTO users (name, email) VALUES (?, ?);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildMysqlInsertQuery(tt.table, tt.columns)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildMysqlUpdateQuery(t *testing.T) {
	tests := []struct {
		name        string
		table       string
		columns     []string
		primaryKeys []string
		expected    string
	}{
		{"Single Column", "users", []string{"name"}, []string{"id"}, "UPDATE users SET name = ? WHERE id = ?;"},
		{"Multiple Primary Keys", "users", []string{"name", "email"}, []string{"id", "other"}, "UPDATE users SET name = ?, email = ? WHERE id = ? AND other = ?;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildMysqlUpdateQuery(tt.table, tt.columns, tt.primaryKeys)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
