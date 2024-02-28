package sync_cmd

import (
	"math"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
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
				expectedConfigMap := map[string]*benthosConfigResponse{}
				for _, cfg := range tt.expect[i] {
					expectedConfigMap[cfg.Name] = cfg
				}
				for _, cfg := range group {
					expect := expectedConfigMap[cfg.Name]
					assert.NotNil(t, expect)
					assert.ElementsMatch(t, cfg.DependsOn, expect.DependsOn)
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
	assert.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), `root = [this."foo", this."bar", this."baz"]`)
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
		{"Single Column", "users", []string{"name"}, `INSERT INTO users ("name") VALUES ($1);`},
		{"Multiple Columns", "users", []string{"name", "email"}, `INSERT INTO users ("name", "email") VALUES ($1, $2);`},
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
		{"Single Column", "users", []string{"name"}, []string{"id"}, `UPDATE users SET "name" = $1 WHERE "id" = $2;`},
		{"Multiple Primary Keys", "users", []string{"name", "email"}, []string{"id", "other"}, `UPDATE users SET "name" = $1, "email" = $2 WHERE "id" = $3 AND "other" = $4;`},
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
		{"Single Column", "users", []string{"name"}, "INSERT INTO users (`name`) VALUES (?);"},
		{"Multiple Columns", "users", []string{"name", "email"}, "INSERT INTO users (`name`, `email`) VALUES (?, ?);"},
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
		{"Single Column", "users", []string{"name"}, []string{"id"}, "UPDATE users SET `name` = ? WHERE `id` = ?;"},
		{"Multiple Primary Keys", "users", []string{"name", "email"}, []string{"id", "other"}, "UPDATE users SET `name` = ?, `email` = ? WHERE `id` = ? AND `other` = ?;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildMysqlUpdateQuery(tt.table, tt.columns, tt.primaryKeys)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildSyncConfigs_postgres(t *testing.T) {
	tests := []struct {
		name   string
		config *schemaConfig
		expect []*syncConfig
	}{
		{
			name: "Empty Schema",
			config: &schemaConfig{
				Schemas:                []*mgmtv1alpha1.DatabaseColumn{},
				TableConstraints:       map[string]*dbschemas_utils.TableConstraints{},
				TablePrimaryKeys:       map[string]*mgmtv1alpha1.PrimaryConstraint{},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{},
		},
		{
			name: "Single Table",
			config: &schemaConfig{
				Schemas: []*mgmtv1alpha1.DatabaseColumn{
					{Schema: "public", Table: "users", Column: "id", DataType: ""},
					{Schema: "public", Table: "users", Column: "name", DataType: ""},
					{Schema: "public", Table: "users", Column: "email", DataType: ""},
				},
				TableConstraints:       map[string]*dbschemas_utils.TableConstraints{},
				TablePrimaryKeys:       map[string]*mgmtv1alpha1.PrimaryConstraint{},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO public.users ("id", "name", "email") VALUES ($1, $2, $3);`,
					ArgsMapping:   `root = [this."id", this."name", this."email"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name", "email"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
			},
		},
		{
			name: "Multiple Tables",
			config: &schemaConfig{
				Schemas: []*mgmtv1alpha1.DatabaseColumn{
					{Schema: "public", Table: "users", Column: "id", DataType: ""},
					{Schema: "public", Table: "users", Column: "name", DataType: ""},
					{Schema: "public", Table: "users", Column: "email", DataType: ""},
					{Schema: "public", Table: "accounts", Column: "id", DataType: ""},
					{Schema: "public", Table: "accounts", Column: "user_id", DataType: ""},
				},
				TableConstraints: map[string]*dbschemas_utils.TableConstraints{
					"public.accounts": {Constraints: []*dbschemas_utils.ForeignConstraint{
						{Column: "user_id", IsNullable: false, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.users", Column: "id"}},
					}},
				},
				TablePrimaryKeys:       map[string]*mgmtv1alpha1.PrimaryConstraint{},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO public.users ("id", "name", "email") VALUES ($1, $2, $3);`,
					ArgsMapping:   `root = [this."id", this."name", this."email"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name", "email"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
				{
					Query:         `INSERT INTO public.accounts ("id", "user_id") VALUES ($1, $2);`,
					ArgsMapping:   `root = [this."id", this."user_id"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "accounts",
					Columns:       []string{"id", "user_id"},
					DependsOn: []*tabledependency.DependsOn{{
						Table: "public.users", Columns: []string{"id"},
					}},
					Name: "public.accounts",
				},
			},
		},
		{
			name: "Circular Dependent Tables",
			config: &schemaConfig{
				Schemas: []*mgmtv1alpha1.DatabaseColumn{
					{Schema: "public", Table: "users", Column: "id", DataType: ""},
					{Schema: "public", Table: "users", Column: "name", DataType: ""},
					{Schema: "public", Table: "users", Column: "account_id", DataType: ""},
					{Schema: "public", Table: "accounts", Column: "id", DataType: ""},
					{Schema: "public", Table: "accounts", Column: "user_id", DataType: ""},
				},
				TableConstraints: map[string]*dbschemas_utils.TableConstraints{
					"public.accounts": {Constraints: []*dbschemas_utils.ForeignConstraint{
						{Column: "user_id", IsNullable: false, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.users", Column: "id"}},
					}},
					"public.users": {Constraints: []*dbschemas_utils.ForeignConstraint{
						{Column: "account_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.accounts", Column: "id"}},
					}},
				},
				TablePrimaryKeys: map[string]*mgmtv1alpha1.PrimaryConstraint{
					"public.accounts": {Columns: []string{"id"}},
					"public.users":    {Columns: []string{"id"}},
				},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO public.users ("id", "name") VALUES ($1, $2);`,
					ArgsMapping:   `root = [this."id", this."name"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
				{
					Query:         `INSERT INTO public.accounts ("id", "user_id") VALUES ($1, $2);`,
					ArgsMapping:   `root = [this."id", this."user_id"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "accounts",
					Columns:       []string{"id", "user_id"},
					DependsOn: []*tabledependency.DependsOn{{
						Table: "public.users", Columns: []string{"id"},
					}},
					Name: "public.accounts",
				},
				{
					Query:         `UPDATE public.users SET "account_id" = $1 WHERE "id" = $2;`,
					ArgsMapping:   `root = [this."account_id", this."id"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"account_id"},
					DependsOn: []*tabledependency.DependsOn{
						{Table: "public.users", Columns: []string{"id"}},
					},
					Name: "public.users.updated",
				},
			},
		},
		{
			name: "Self Circular Dependency",
			config: &schemaConfig{
				Schemas: []*mgmtv1alpha1.DatabaseColumn{
					{Schema: "public", Table: "users", Column: "id", DataType: ""},
					{Schema: "public", Table: "users", Column: "name", DataType: ""},
					{Schema: "public", Table: "users", Column: "user_id", DataType: ""},
				},
				TableConstraints: map[string]*dbschemas_utils.TableConstraints{
					"public.users": {Constraints: []*dbschemas_utils.ForeignConstraint{
						{Column: "user_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.users", Column: "id"}},
					}},
				},
				TablePrimaryKeys: map[string]*mgmtv1alpha1.PrimaryConstraint{
					"public.users": {Columns: []string{"id"}},
				},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO public.users ("id", "name") VALUES ($1, $2);`,
					ArgsMapping:   `root = [this."id", this."name"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
				{
					Query:         `UPDATE public.users SET "user_id" = $1 WHERE "id" = $2;`,
					ArgsMapping:   `root = [this."user_id", this."id"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"user_id"},
					DependsOn: []*tabledependency.DependsOn{
						{Table: "public.users", Columns: []string{"id"}},
					},
					Name: "public.users.updated",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := buildSyncConfigs(tt.config, buildPostgresInsertQuery, buildPostgresUpdateQuery)
			assert.Len(t, configs, len(tt.expect))
			for _, actual := range configs {
				for _, c := range tt.expect {
					if c.Name == actual.Name {
						assert.Equal(t, actual, c)
					}
				}
			}
		})
	}
}
