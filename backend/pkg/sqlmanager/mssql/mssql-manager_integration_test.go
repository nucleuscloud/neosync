package sqlmanager_mssql

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetDatabaseSchema() {
	manager := NewManager(s.source.querier, s.source.testDb, func() {})

	var expectedIdentityGeneration = "IDENTITY(1,1)"
	expectedSubset := []*sqlmanager_shared.DatabaseSchemaRow{
		{
			TableSchema:            "sqlmanagermssql3",
			TableName:              "users",
			ColumnName:             "id",
			DataType:               "int",
			ColumnDefault:          "",
			IsNullable:             false,
			CharacterMaximumLength: -1,
			NumericPrecision:       10,
			NumericScale:           0,
			OrdinalPosition:        1,
			GeneratedType:          nil,
			IdentityGeneration:     &expectedIdentityGeneration,
		},
	}

	actual, err := manager.GetDatabaseSchema(s.ctx)
	require.NoError(s.T(), err)
	containsSubset(s.T(), actual, expectedSubset)
}

func (s *IntegrationTestSuite) Test_GetSchemaColumnMap() {
	manager := NewManager(s.source.querier, s.source.testDb, func() {})

	actual, err := manager.GetSchemaColumnMap(s.ctx)
	require.NoError(s.T(), err)

	usersKey := fmt.Sprintf("%s.%s", "sqlmanagermssql3", "users")

	usersMap, ok := actual[usersKey]
	require.True(s.T(), ok, fmt.Sprintf("%s map should exist", usersKey))
	require.NotEmpty(s.T(), usersMap)
	_, ok = usersMap["id"]
	require.True(s.T(), ok, "users map should have id column")
}

func (s *IntegrationTestSuite) Test_GetTableConstraintsBySchema() {
	manager := NewManager(s.source.querier, s.source.testDb, func() {})

	expected := &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{
			"sqlmanagermssql2.child1": {
				{Columns: []string{"parent_id1", "parent_id2"}, NotNullable: []bool{false, false}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermssql2.parent1",
					Columns: []string{"id1", "id2"},
				}},
			},

			"sqlmanagermssql2.TableA": {
				{Columns: []string{"IdB1", "IdB2"}, NotNullable: []bool{false, false}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermssql2.TableB",
					Columns: []string{"IdB1", "IdB2"},
				}},
			},
			"sqlmanagermssql2.TableB": {
				{Columns: []string{"IdA1", "IdA2"}, NotNullable: []bool{true, true}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermssql2.TableA",
					Columns: []string{"IdA1", "IdA2"},
				}},
			},
		},
		PrimaryKeyConstraints: map[string][]string{
			"sqlmanagermssql2.parent1": {"id1", "id2"},
			"sqlmanagermssql2.child1":  {"id"},

			"sqlmanagermssql2.TableA": {"IdA1", "IdA2"},
			"sqlmanagermssql2.TableB": {"IdB1", "IdB2"},

			"sqlmanagermssql2.defaults_table": {"id"},
		},
	}

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{"sqlmanagermssql2"})
	require.NoError(s.T(), err)
	require.Equal(s.T(), expected.ForeignKeyConstraints, actual.ForeignKeyConstraints)
	require.Equal(s.T(), expected.PrimaryKeyConstraints, actual.PrimaryKeyConstraints)
}

// func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap_BasicCircular() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"
// 	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{schema})
// 	require.NoError(s.T(), err)
// 	require.NotEmpty(s.T(), actual)

// 	constraints, ok := actual.ForeignKeyConstraints[s.buildTable(schema, "t1")]
// 	require.True(s.T(), ok, "t1 should be in map")
// 	require.NotEmpty(s.T(), constraints, "should contain t1 constraints")
// 	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
// 		{
// 			Columns:     []string{"b"},
// 			NotNullable: []bool{false},
// 			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t1"), Columns: []string{"a"}},
// 		},
// 	})

// 	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t2")]
// 	require.True(s.T(), ok, "t2 should be in map")
// 	require.NotEmpty(s.T(), constraints, "should contain t2 constraints")
// 	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
// 		{
// 			Columns:     []string{"b"},
// 			NotNullable: []bool{false},
// 			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t3"), Columns: []string{"a"}},
// 		},
// 	})

// 	constraints, ok = actual.ForeignKeyConstraints[s.buildTable(schema, "t3")]
// 	require.True(s.T(), ok, "t3 should be in map")
// 	require.NotEmpty(s.T(), constraints, "should contain t3 constraints")
// 	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
// 		{
// 			Columns:     []string{"b"},
// 			NotNullable: []bool{false},
// 			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable(schema, "t2"), Columns: []string{"a"}},
// 		},
// 	})
// }

func (s *IntegrationTestSuite) Test_GetRolePermissionsMap() {
	manager := NewManager(s.source.querier, s.source.testDb, func() {})
	schema := "sqlmanagermssql3"

	actual, err := manager.GetRolePermissionsMap(context.Background())
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	usersKey := buildTable(schema, "users")

	usersRecord, ok := actual[usersKey]
	require.True(s.T(), ok, "map should have users perms")
	require.Contains(s.T(), usersRecord, "INSERT")
	require.Contains(s.T(), usersRecord, "UPDATE")
	require.Contains(s.T(), usersRecord, "SELECT")
	require.Contains(s.T(), usersRecord, "DELETE")
}

// func (s *IntegrationTestSuite) Test_GetCreateTableStatement() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	actual, err := manager.GetCreateTableStatement(context.Background(), schema, "users")
// 	require.NoError(s.T(), err)
// 	require.NotEmpty(s.T(), actual)
// 	_, err = s.target.pool.ExecContext(context.Background(), actual)
// 	require.NoError(s.T(), err)
// }

// func (s *IntegrationTestSuite) Test_GetTableInitStatements() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	actual, err := manager.GetTableInitStatements(
// 		context.Background(),
// 		[]*sqlmanager_shared.SchemaTable{
// 			{Schema: schema, Table: "parent1"},
// 			{Schema: schema, Table: "child1"},
// 			{Schema: schema, Table: "order"},
// 		},
// 	)

// 	require.NoError(s.T(), err)
// 	require.NotEmpty(s.T(), actual)
// 	for _, stmt := range actual {
// 		_, err = s.target.pool.ExecContext(context.Background(), stmt.CreateTableStatement)
// 		require.NoError(s.T(), err)
// 	}
// 	for _, stmt := range actual {
// 		for _, index := range stmt.IndexStatements {
// 			_, err = s.target.pool.ExecContext(context.Background(), index)
// 			require.NoError(s.T(), err)
// 		}
// 	}
// 	for _, stmt := range actual {
// 		for _, alter := range stmt.AlterTableStatements {
// 			_, err = s.target.pool.ExecContext(context.Background(), alter.Statement)
// 			require.NoError(s.T(), err)
// 		}
// 	}
// }

// func (s *IntegrationTestSuite) Test_Exec() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	err := manager.Exec(context.Background(), fmt.Sprintf("SELECT 1 FROM %s.%s", schema, "users"))
// 	require.NoError(s.T(), err)
// }

// func (s *IntegrationTestSuite) Test_BatchExec() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	stmt := fmt.Sprintf("SELECT 1 FROM %s.%s;", schema, "users")
// 	err := manager.BatchExec(context.Background(), 2, []string{stmt, stmt, stmt}, &sqlmanager_shared.BatchExecOpts{})
// 	require.NoError(s.T(), err)
// }

// func (s *IntegrationTestSuite) Test_BatchExec_With_Prefix() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	stmt := fmt.Sprintf("SELECT 1 FROM %s.%s;", schema, "users")
// 	err := manager.BatchExec(context.Background(), 2, []string{stmt}, &sqlmanager_shared.BatchExecOpts{
// 		Prefix: &stmt,
// 	})
// 	require.NoError(s.T(), err)
// }

// func (s *IntegrationTestSuite) Test_GetSchemaInitStatements() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	statements, err := manager.GetSchemaInitStatements(context.Background(), []*sqlmanager_shared.SchemaTable{
// 		{Schema: schema, Table: "parent1"},
// 		{Schema: schema, Table: "child1"},
// 		{Schema: schema, Table: "custom_table"},
// 		{Schema: schema, Table: "unique_emails_and_usernames"},
// 		{Schema: schema, Table: "t1"},
// 		{Schema: schema, Table: "t2"},
// 		{Schema: schema, Table: "t3"},
// 		{Schema: schema, Table: "t4"},
// 		{Schema: schema, Table: "t5"},
// 		{Schema: schema, Table: "employee_log"},
// 		{Schema: schema, Table: "users"},
// 	})
// 	require.NoError(s.T(), err)
// 	require.NotEmpty(s.T(), statements)
// 	lableStmtMap := map[string][]string{}
// 	for _, st := range statements {
// 		lableStmtMap[st.Label] = append(lableStmtMap[st.Label], st.Statements...)
// 	}
// 	for _, stmt := range lableStmtMap["create table"] {
// 		_, err = s.target.pool.ExecContext(context.Background(), stmt)
// 		require.NoError(s.T(), err)
// 	}
// 	for _, stmt := range lableStmtMap["table triggers"] {
// 		_, err = s.target.pool.ExecContext(context.Background(), stmt)
// 		require.NoError(s.T(), err)
// 	}
// 	for _, stmt := range lableStmtMap["table index"] {
// 		_, err = s.target.pool.ExecContext(context.Background(), stmt)
// 		require.NoError(s.T(), err)
// 	}
// 	for _, stmt := range lableStmtMap["non-fk alter table"] {
// 		_, err = s.target.pool.ExecContext(context.Background(), stmt)
// 		require.NoError(s.T(), err)
// 	}
// 	for _, stmt := range lableStmtMap["fk alter table"] {
// 		_, err = s.target.pool.ExecContext(context.Background(), stmt)
// 		require.NoError(s.T(), err)
// 	}
// }

// func (s *IntegrationTestSuite) Test_GetSchemaInitStatements_customtable() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	statements, err := manager.GetSchemaInitStatements(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "custom_table"}})
// 	require.NoError(s.T(), err)
// 	require.NotEmpty(s.T(), statements)
// }

// func (s *IntegrationTestSuite) Test_GetSchemaTableTriggers_customtable() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	triggers, err := manager.GetSchemaTableTriggers(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "employee_log"}})
// 	require.NoError(s.T(), err)
// 	require.NotEmpty(s.T(), triggers)
// }

// func (s *IntegrationTestSuite) Test_GetSchemaTableDataTypes_customtable() {
// 	manager := NewManager(s.source.querier, s.source.pool, func() {})
// 	schema := "sqlmanagermysql3"

// 	resp, err := manager.GetSchemaTableDataTypes(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "custom_table"}})
// 	require.NoError(s.T(), err)
// 	require.NotNil(s.T(), resp)
// 	require.NotEmptyf(s.T(), resp.GetStatements(), "statements")
// 	require.NotEmptyf(s.T(), resp.Functions, "functions")
// }

func containsSubset[T any](t testing.TB, array, subset []T) {
	t.Helper()
	for _, elem := range subset {
		require.Contains(t, array, elem)
	}
}

type testColumnProperties struct {
	needsOverride bool
	needsReset    bool
}

func (s *IntegrationTestSuite) Test_GetMssqlColumnOverrideAndResetProperties() {
	manager := NewManager(s.source.querier, s.source.testDb, func() {})

	colInfoMap, err := manager.GetSchemaColumnMap(context.Background())
	require.NoError(s.T(), err)

	testDefaultTable := colInfoMap["testdb.sqlmanagermssql2.defaults_table"]

	var expectedProperties = map[string]testColumnProperties{
		"description":       {needsOverride: false, needsReset: false},
		"registration_date": {needsOverride: false, needsReset: false},
		"score":             {needsOverride: false, needsReset: false},
		"status":            {needsOverride: false, needsReset: false},
		"id":                {needsOverride: true, needsReset: true},
		"last_login":        {needsOverride: false, needsReset: false},
		"age":               {needsOverride: false, needsReset: false},
		"is_active":         {needsOverride: false, needsReset: false},
		"created_at":        {needsOverride: false, needsReset: false},
		"uuid":              {needsOverride: false, needsReset: false},
	}

	for col, colInfo := range testDefaultTable {
		needsOverride, needsReset := GetMssqlColumnOverrideAndResetProperties(colInfo)
		expected, ok := expectedProperties[col]
		require.Truef(s.T(), ok, "Missing expected column %q", col)
		require.Equalf(s.T(), expected.needsOverride, needsOverride, "Incorrect needsOverride value for column %q", col)
		require.Equalf(s.T(), expected.needsReset, needsReset, "Incorrect needsReset value for column %q", col)
	}
}
