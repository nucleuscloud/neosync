package sqlmanager_mysql

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetTableConstraintsBySchema() {
	manager := MysqlManager{querier: s.querier, pool: s.containers.Source.DB}

	expected := &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{
			"sqlmanagermysql.container": {
				{Columns: []string{"container_status_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermysql.container_status",
					Columns: []string{"id"},
				}},
			},
			"sqlmanagermysql2.container": {
				{Columns: []string{"container_status_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermysql2.container_status",
					Columns: []string{"id"},
				}},
			},
		},
		PrimaryKeyConstraints: map[string][]string{
			"sqlmanagermysql.container":         {"id"},
			"sqlmanagermysql.container_status":  {"id"},
			"sqlmanagermysql2.container":        {"id"},
			"sqlmanagermysql2.container_status": {"id"},
		},
	}

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{"sqlmanagermysql", "sqlmanagermysql2"})
	require.NoError(s.T(), err)
	require.Equal(s.T(), expected.ForeignKeyConstraints, actual.ForeignKeyConstraints)
	require.Equal(s.T(), expected.PrimaryKeyConstraints, actual.PrimaryKeyConstraints)
}

func (s *IntegrationTestSuite) Test_GetSchemaColumnMap() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"
	actual, err := manager.GetSchemaColumnMap(context.Background())
	require.NoError(s.T(), err)

	usersKey := s.buildTable(schema, "users")

	usersMap, ok := actual[usersKey]
	require.True(s.T(), ok, fmt.Sprintf("%s map should exist", usersKey))
	require.NotEmpty(s.T(), usersMap)
	_, ok = usersMap["id"]
	require.True(s.T(), ok, "users map should have id column")
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"
	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	constraints, ok := actual.ForeignKeyConstraints[s.buildTable(schema, "child1")]
	require.True(s.T(), ok)
	require.NotEmpty(s.T(), constraints)

	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"parent_id"},
			NotNullable: []bool{false},
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Table:   fmt.Sprintf("%s.parent1", "sqlmanagermysql3"),
				Columns: []string{"id"},
			},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t1")]
	require.True(s.T(), ok, "t1 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t1 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t1"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t2")]
	require.True(s.T(), ok, "t2 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t2 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t3"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t3")]
	require.True(s.T(), ok, "t3 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t3 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t2"), Columns: []string{"a"}},
		},
	})
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap_BasicCircular() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"
	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	constraints, ok := actual.ForeignKeyConstraints[s.buildTable(schema, "t1")]
	require.True(s.T(), ok, "t1 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t1 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t1"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t2")]
	require.True(s.T(), ok, "t2 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t2 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t3"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t3")]
	require.True(s.T(), ok, "t3 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t3 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t2"), Columns: []string{"a"}},
		},
	})
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap_Composite() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	constraints, ok := actual.ForeignKeyConstraints[s.buildTable(schema, "t5")]
	require.True(s.T(), ok, "t5 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t5 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"x", "y"},
			NotNullable: []bool{true, true},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t4"), Columns: []string{"a", "b"}},
		},
	})
}

func (s *IntegrationTestSuite) Test_GetPrimaryKeyConstraintsMap() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	require.NotEmpty(s.T(), actual.PrimaryKeyConstraints)

	pkeys, ok := actual.PrimaryKeyConstraints[s.buildTable(schema, "users")]
	require.True(s.T(), ok, " had no entries")
	require.ElementsMatch(s.T(), []string{"id"}, pkeys)

	pkeys, ok = actual.PrimaryKeyConstraints[s.buildTable(schema, "t4")]
	require.True(s.T(), ok, "t4 had no entries")
	require.ElementsMatch(s.T(), []string{"a", "b"}, pkeys)
}

func (s *IntegrationTestSuite) Test_GetUniqueConstraintsMap() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	require.NotEmpty(s.T(), actual.UniqueConstraints)

	uniques, ok := actual.UniqueConstraints[s.buildTable(schema, "unique_emails")]
	require.True(s.T(), ok, "unique_emails had no entries")
	require.ElementsMatch(s.T(), [][]string{{"email"}}, uniques)
}

func (s *IntegrationTestSuite) Test_GetUniqueConstraintsMap_Composite() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	require.NotEmpty(s.T(), actual.UniqueConstraints)

	uniques, ok := actual.UniqueConstraints[s.buildTable(schema, "unique_emails_and_usernames")]
	require.True(s.T(), ok, "unique_emails_and_usernames had no entries")
	require.Len(s.T(), uniques, 1)
	entry := uniques[0]
	require.ElementsMatch(s.T(), []string{"email", "username"}, entry)
}

func (s *IntegrationTestSuite) Test_GetRolePermissionsMap() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetRolePermissionsMap(context.Background())
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	usersKey := s.buildTable(schema, "users")

	usersRecord, ok := actual[usersKey]
	require.True(s.T(), ok, "map should have users perms")
	require.Contains(s.T(), usersRecord, "INSERT")
	require.Contains(s.T(), usersRecord, "UPDATE")
	require.Contains(s.T(), usersRecord, "SELECT")
	require.Contains(s.T(), usersRecord, "DELETE")
	require.Contains(s.T(), usersRecord, "TRIGGER")
}

func (s *IntegrationTestSuite) Test_GetCreateTableStatement() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetCreateTableStatement(context.Background(), schema, "users")
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	_, err = s.containers.Target.DB.ExecContext(context.Background(), actual)
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) Test_GetTableInitStatements() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	actual, err := manager.GetTableInitStatements(
		context.Background(),
		[]*sqlmanager_shared.SchemaTable{
			{Schema: schema, Table: "parent1"},
			{Schema: schema, Table: "child1"},
			{Schema: schema, Table: "order"},
		},
	)

	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	for _, stmt := range actual {
		_, err = s.containers.Target.DB.ExecContext(context.Background(), stmt.CreateTableStatement)
		require.NoError(s.T(), err)
	}
	for _, stmt := range actual {
		for _, index := range stmt.IndexStatements {
			_, err = s.containers.Target.DB.ExecContext(context.Background(), index)
			require.NoError(s.T(), err)
		}
	}
	for _, stmt := range actual {
		for _, alter := range stmt.AlterTableStatements {
			_, err = s.containers.Target.DB.ExecContext(context.Background(), alter.Statement)
			require.NoError(s.T(), err)
		}
	}
}

func (s *IntegrationTestSuite) Test_Exec() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	err := manager.Exec(context.Background(), fmt.Sprintf("SELECT 1 FROM %s.%s", schema, "users"))
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) Test_BatchExec() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	stmt := fmt.Sprintf("SELECT 1 FROM %s.%s;", schema, "users")
	err := manager.BatchExec(context.Background(), 2, []string{stmt, stmt, stmt}, &sqlmanager_shared.BatchExecOpts{})
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) Test_BatchExec_With_Prefix() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	stmt := fmt.Sprintf("SELECT 1 FROM %s.%s;", schema, "users")
	err := manager.BatchExec(context.Background(), 2, []string{stmt}, &sqlmanager_shared.BatchExecOpts{
		Prefix: &stmt,
	})
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) Test_GetSchemaInitStatements() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	statements, err := manager.GetSchemaInitStatements(context.Background(), []*sqlmanager_shared.SchemaTable{
		{Schema: schema, Table: "parent1"},
		{Schema: schema, Table: "child1"},
		{Schema: schema, Table: "custom_table"},
		{Schema: schema, Table: "unique_emails_and_usernames"},
		{Schema: schema, Table: "t1"},
		{Schema: schema, Table: "t2"},
		{Schema: schema, Table: "t3"},
		{Schema: schema, Table: "t4"},
		{Schema: schema, Table: "t5"},
		{Schema: schema, Table: "employee_log"},
		{Schema: schema, Table: "users"},
	})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), statements)
	lableStmtMap := map[string][]string{}
	for _, st := range statements {
		lableStmtMap[st.Label] = append(lableStmtMap[st.Label], st.Statements...)
	}
	for _, stmt := range lableStmtMap["create table"] {
		_, err = s.containers.Target.DB.ExecContext(context.Background(), stmt)
		require.NoError(s.T(), err)
	}
	for _, stmt := range lableStmtMap["table triggers"] {
		_, err = s.containers.Target.DB.ExecContext(context.Background(), stmt)
		require.NoError(s.T(), err)
	}
	for _, stmt := range lableStmtMap["table index"] {
		_, err = s.containers.Target.DB.ExecContext(context.Background(), stmt)
		require.NoError(s.T(), err)
	}
	for _, stmt := range lableStmtMap["non-fk alter table"] {
		_, err = s.containers.Target.DB.ExecContext(context.Background(), stmt)
		require.NoError(s.T(), err)
	}
	for _, stmt := range lableStmtMap["fk alter table"] {
		_, err = s.containers.Target.DB.ExecContext(context.Background(), stmt)
		require.NoError(s.T(), err)
	}
}

func (s *IntegrationTestSuite) Test_GetSchemaInitStatements_customtable() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	statements, err := manager.GetSchemaInitStatements(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "custom_table"}})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), statements)
}

func (s *IntegrationTestSuite) Test_GetSchemaTableTriggers_customtable() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	triggers, err := manager.GetSchemaTableTriggers(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "employee_log"}})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), triggers)
}

func (s *IntegrationTestSuite) Test_GetSchemaTableDataTypes_customtable() {
	manager := NewManager(s.querier, s.containers.Source.DB, func() {})
	schema := "sqlmanagermysql3"

	resp, err := manager.GetSchemaTableDataTypes(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "custom_table"}})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.NotEmptyf(s.T(), resp.GetStatements(), "statements")
	require.NotEmptyf(s.T(), resp.Functions, "functions")
}

func containsSubset[T any](t testing.TB, array, subset []T) {
	t.Helper()
	for _, elem := range subset {
		require.Contains(t, array, elem)
	}
}
