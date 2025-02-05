package sqlmanager_mssql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmssql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/sqlserver"
	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
)

type testColumnProperties struct {
	needsOverride bool
	needsReset    bool
}

func Test_MssqlManager(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	t.Log("Running integration tests for Mssql Manager")
	t.Parallel()

	ctx := context.Background()
	containers, err := tcmssql.NewMssqlTestSyncContainer(ctx, []tcmssql.Option{}, []tcmssql.Option{})
	if err != nil {
		t.Fatal(err)
	}
	source := containers.Source
	target := containers.Target
	t.Log("Successfully created source and target mssql test containers")

	err = setup(ctx, containers)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Successfully setup source and target databases")
	manager := mssql.NewManager(mssql_queries.New(), source.DB, func() {})

	t.Run("GetDatabaseSchema", func(t *testing.T) {
		t.Parallel()
		var expectedIdentityGeneration = "IDENTITY(1,1)"
		var expectedIdentitySeed = 1
		var expectedIdentityIncrement = 1
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
				ColumnDefaultType:      nil,
				IdentitySeed:           &expectedIdentitySeed,
				IdentityIncrement:      &expectedIdentityIncrement,
			},
		}

		actual, err := manager.GetDatabaseSchema(ctx)
		require.NoError(t, err)
		containsSubset(t, actual, expectedSubset)
	})

	t.Run("GetSchemaColumnMap", func(t *testing.T) {
		t.Parallel()
		actual, err := manager.GetSchemaColumnMap(ctx)
		require.NoError(t, err)

		usersKey := fmt.Sprintf("%s.%s", "sqlmanagermssql3", "users")

		usersMap, ok := actual[usersKey]
		require.True(t, ok, fmt.Sprintf("%s map should exist", usersKey))
		require.NotEmpty(t, usersMap)
		_, ok = usersMap["id"]
		require.True(t, ok, "users map should have id column")
	})

	t.Run("GetTableConstraintsBySchema", func(t *testing.T) {
		t.Parallel()
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

		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{"sqlmanagermssql2"})
		require.NoError(t, err)
		require.Equal(t, expected.ForeignKeyConstraints, actual.ForeignKeyConstraints)
		require.Equal(t, expected.PrimaryKeyConstraints, actual.PrimaryKeyConstraints)
	})

	t.Run("GetRolePermissionMap", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermssql3"

		actual, err := manager.GetRolePermissionsMap(context.Background())
		require.NoError(t, err)
		require.NotEmpty(t, actual)

		usersKey := buildTable(schema, "users")

		usersRecord, ok := actual[usersKey]
		require.True(t, ok, "map should have users perms")
		require.Contains(t, usersRecord, "INSERT")
		require.Contains(t, usersRecord, "UPDATE")
		require.Contains(t, usersRecord, "SELECT")
		require.Contains(t, usersRecord, "DELETE")
	})

	t.Run("GetMssqlColumnOverrideAndResetProperties", func(t *testing.T) {
		t.Parallel()
		colInfoMap, err := manager.GetSchemaColumnMap(context.Background())
		require.NoError(t, err)

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
			needsOverride, needsReset := mssql.GetMssqlColumnOverrideAndResetProperties(colInfo)
			expected, ok := expectedProperties[col]
			require.Truef(t, ok, "Missing expected column %q", col)
			require.Equalf(t, expected.needsOverride, needsOverride, "Incorrect needsOverride value for column %q", col)
			require.Equalf(t, expected.needsReset, needsReset, "Incorrect needsReset value for column %q", col)
		}
	})

	t.Run("GetSchemaInitStatements", func(t *testing.T) {
		t.Parallel()
		schema := "mssqlinit"
		tables := []string{"Invoices", "Customers", "Orders", "Products", "OrderItems"}

		schematables := []*sqlmanager_shared.SchemaTable{}
		for _, t := range tables {
			schematables = append(schematables, &sqlmanager_shared.SchemaTable{Schema: schema, Table: t})
		}

		statements, err := manager.GetSchemaInitStatements(ctx, schematables)
		require.NoErrorf(t, err, "failed to get schema init statements")
		require.NotEmpty(t, statements)
		lableStmtMap := map[string][]string{}
		for _, st := range statements {
			lableStmtMap[st.Label] = append(lableStmtMap[st.Label], st.Statements...)
		}
		for _, stmt := range lableStmtMap["data types"] {
			_, err = target.DB.ExecContext(ctx, stmt)
			require.NoErrorf(t, err, "failed to create data type in target db: %s", stmt)
		}
		for _, stmt := range lableStmtMap["create table"] {
			_, err = target.DB.ExecContext(ctx, stmt)
			require.NoErrorf(t, err, "failed to create tables in target db: %s", stmt)
		}
		for _, stmt := range lableStmtMap["table triggers"] {
			_, err = target.DB.ExecContext(ctx, stmt)
			require.NoErrorf(t, err, "failed to create table triggers in target db: %s", stmt)
		}
		for _, stmt := range lableStmtMap["table index"] {
			_, err = target.DB.ExecContext(ctx, stmt)
			require.NoErrorf(t, err, "failed to create table indexes in target db: %s", stmt)
		}
		for _, stmt := range lableStmtMap["non-fk alter table"] {
			_, err = target.DB.ExecContext(ctx, stmt)
			require.NoErrorf(t, err, "failed to create non-fk constraints in target db: %s", stmt)
		}
		for _, stmt := range lableStmtMap["fk alter table"] {
			_, err = target.DB.ExecContext(ctx, stmt)
			require.NoErrorf(t, err, "failed to create fk constraints in target db: %s", stmt)
		}
	})

	t.Log("Finished running mssql manager integration tests")
	t.Cleanup(func() {
		t.Log("Cleaning up source and target mssql containers")
		err := source.TearDown(ctx)
		if err != nil {
			t.Fatal(err)
		}
		err = target.TearDown(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
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

func containsSubset[T any](t testing.TB, array, subset []T) {
	for idx, elem := range subset {
		require.Contains(t, array, elem, "array does not contain expected subset element at index %d", idx)
	}
}

func buildTable(schema, tableName string) string {
	return fmt.Sprintf("%s.%s", schema, tableName)
}

type Schema struct {
	Name string
}

func setup(ctx context.Context, containers *tcmssql.MssqlTestSyncContainer) error {
	baseDir := "testdata"

	sourceSetupContents, err := readSqlFiles(filepath.Join(baseDir, "source-setup"))
	if err != nil {
		return fmt.Errorf("unable to read source setup files: %w", err)
	}

	destSetupContents, err := readSqlFiles(filepath.Join(baseDir, "dest-setup"))
	if err != nil {
		return fmt.Errorf("unable to read dest setup files: %w", err)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		for i, stmt := range sourceSetupContents {
			_, err := containers.Source.DB.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing source setup statement %d: %w", i+1, err)
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		for i, stmt := range destSetupContents {
			_, err := containers.Target.DB.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing dest setup statement: %d: %w", i+1, err)
			}
		}
		return nil
	})

	err = errgrp.Wait()
	if err != nil {
		return err
	}

	return nil
}

func readSqlFiles(dir string) ([]string, error) {
	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %v", dir, err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Prepare a slice to store results
	sqlContents := make([]string, len(sqlFiles))

	// Use errgroup for concurrent file reading
	var eg errgroup.Group
	var mu sync.Mutex

	for i, file := range sqlFiles {
		i, file := i, file
		eg.Go(func() error {
			content, err := os.ReadFile(filepath.Join(dir, file))
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", file, err)
			}

			mu.Lock()
			sqlContents[i] = string(content)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all goroutines to complete and check for errors
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return sqlContents, nil
}
