package sync_cmd

import (
	"math"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
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
			groups := groupConfigsByDependency(tt.configs)
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
	groups := groupConfigsByDependency(configs)
	require.Nil(t, groups)
}

func Test_buildPlainInsertArgs(t *testing.T) {
	require.Empty(t, buildPlainInsertArgs(nil))
	require.Empty(t, buildPlainInsertArgs([]string{}))
	require.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), `root = [this."foo", this."bar", this."baz"]`)
}

func Test_clampInt(t *testing.T) {
	require.Equal(t, clampInt(0, 1, 2), 1)
	require.Equal(t, clampInt(1, 1, 2), 1)
	require.Equal(t, clampInt(2, 1, 2), 2)
	require.Equal(t, clampInt(3, 1, 2), 2)
	require.Equal(t, clampInt(1, 1, 1), 1)

	require.Equal(t, clampInt(1, 3, 2), 3, "low is evaluated first, order is relevant")
}

func Test_computeMaxPgBatchCount(t *testing.T) {
	require.Equal(t, computeMaxPgBatchCount(65535), 1)
	require.Equal(t, computeMaxPgBatchCount(65536), 1, "anything over max should clamp to 1")
	require.Equal(t, computeMaxPgBatchCount(math.MaxInt), 1, "anything over pgmax should clamp to 1")
	require.Equal(t, computeMaxPgBatchCount(1), 65535)
	require.Equal(t, computeMaxPgBatchCount(0), 65535)
}

func Test_buildPostgresInsertQuery(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		table    string
		columns  []string
		expected string
	}{
		{"Single Column", "public", "users", []string{"name"}, `INSERT INTO "public"."users" ("name") VALUES ($1);`},
		{"Multiple Columns", "public", "users", []string{"name", "email"}, `INSERT INTO "public"."users" ("name", "email") VALUES ($1, $2);`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildPostgresInsertQuery(tt.schema, tt.table, tt.columns)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildPostgresUpdateQuery(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		table       string
		columns     []string
		primaryKeys []string
		expected    string
	}{
		{"Single Column", "public", "users", []string{"name"}, []string{"id"}, `UPDATE "public"."users" SET "name" = $1 WHERE "id" = $2;`},
		{"Multiple Primary Keys", "public", "users", []string{"name", "email"}, []string{"id", "other"}, `UPDATE "public"."users" SET "name" = $1, "email" = $2 WHERE "id" = $3 AND "other" = $4;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildPostgresUpdateQuery(tt.schema, tt.table, tt.columns, tt.primaryKeys)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildMysqlInsertQuery(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		table    string
		columns  []string
		expected string
	}{
		{"Single Column", "public", "users", []string{"name"}, "INSERT INTO `public`.`users` (`name`) VALUES (?);"},
		{"Multiple Columns", "public", "users", []string{"name", "email"}, "INSERT INTO `public`.`users` (`name`, `email`) VALUES (?, ?);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildMysqlInsertQuery(tt.schema, tt.table, tt.columns)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_buildMysqlUpdateQuery(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		table       string
		columns     []string
		primaryKeys []string
		expected    string
	}{
		{"Single Column", "public", "users", []string{"name"}, []string{"id"}, "UPDATE `public`.`users` SET `name` = ? WHERE `id` = ?;"},
		{"Multiple Primary Keys", "public", "users", []string{"name", "email"}, []string{"id", "other"}, "UPDATE `public`.`users` SET `name` = ?, `email` = ? WHERE `id` = ? AND `other` = ?;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildMysqlUpdateQuery(tt.schema, tt.table, tt.columns, tt.primaryKeys)
			require.Equal(t, tt.expected, actual)
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
				TableConstraints:       map[string][]*sql_manager.ForeignConstraint{},
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
				TableConstraints:       map[string][]*sql_manager.ForeignConstraint{},
				TablePrimaryKeys:       map[string]*mgmtv1alpha1.PrimaryConstraint{},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO "public"."users" ("id", "name", "email") VALUES ($1, $2, $3);`,
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
				TableConstraints: map[string][]*sql_manager.ForeignConstraint{
					"public.accounts": {
						{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}},
					},
				},
				TablePrimaryKeys:       map[string]*mgmtv1alpha1.PrimaryConstraint{},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO "public"."users" ("id", "name", "email") VALUES ($1, $2, $3);`,
					ArgsMapping:   `root = [this."id", this."name", this."email"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name", "email"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
				{
					Query:         `INSERT INTO "public"."accounts" ("id", "user_id") VALUES ($1, $2);`,
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
				TableConstraints: map[string][]*sql_manager.ForeignConstraint{
					"public.accounts": {
						{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}},
					},
					"public.users": {
						{Columns: []string{"account_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.accounts", Columns: []string{"id"}}},
					},
				},
				TablePrimaryKeys: map[string]*mgmtv1alpha1.PrimaryConstraint{
					"public.accounts": {Columns: []string{"id"}},
					"public.users":    {Columns: []string{"id"}},
				},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO "public"."users" ("id", "name") VALUES ($1, $2);`,
					ArgsMapping:   `root = [this."id", this."name"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
				{
					Query:         `INSERT INTO "public"."accounts" ("id", "user_id") VALUES ($1, $2);`,
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
					Query:         `UPDATE "public"."users" SET "account_id" = $1 WHERE "id" = $2;`,
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
				TableConstraints: map[string][]*sql_manager.ForeignConstraint{
					"public.users": {
						{Columns: []string{"user_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}},
					},
				},
				TablePrimaryKeys: map[string]*mgmtv1alpha1.PrimaryConstraint{
					"public.users": {Columns: []string{"id"}},
				},
				InitTableStatementsMap: map[string]string{},
			},
			expect: []*syncConfig{
				{
					Query:         `INSERT INTO "public"."users" ("id", "name") VALUES ($1, $2);`,
					ArgsMapping:   `root = [this."id", this."name"]`,
					InitStatement: "",
					Schema:        "public",
					Table:         "users",
					Columns:       []string{"id", "name"},
					DependsOn:     []*tabledependency.DependsOn{},
					Name:          "public.users",
				},
				{
					Query:         `UPDATE "public"."users" SET "user_id" = $1 WHERE "id" = $2;`,
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
			require.Len(t, configs, len(tt.expect))
			for _, actual := range configs {
				for _, c := range tt.expect {
					if c.Name == actual.Name {
						require.Equal(t, c, actual)
					}
				}
			}
		})
	}
}
