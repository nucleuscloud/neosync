package sqlmanager_mysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_MysqlManager(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	t.Log("Running integration tests for Mysql Manager")
	t.Parallel()

	ctx := context.Background()
	containers, err := tcmysql.NewMysqlTestSyncContainer(ctx, []tcmysql.Option{}, []tcmysql.Option{})
	if err != nil {
		t.Fatal(err)
	}
	source := containers.Source
	target := containers.Target
	t.Log("Successfully created source and target mysql test containers")

	err = setup(ctx, containers)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Successfully setup source and target databases")

	sourceDB, err := sql.Open(sqlmanager_shared.MysqlDriver, source.URL)
	if err != nil {
		t.Fatal(err)
	}
	manager := mysql.NewManager(mysql_queries.New(), sourceDB, func() {})

	t.Run("GetTableConstraintsBySchema", func(t *testing.T) {
		t.Parallel()
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
		require.NoError(t, err)
		require.Equal(t, expected.ForeignKeyConstraints, actual.ForeignKeyConstraints)
		require.Equal(t, expected.PrimaryKeyConstraints, actual.PrimaryKeyConstraints)
	})

	t.Run("GetSchemaColumnMap", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"
		actual, err := manager.GetSchemaColumnMap(context.Background())
		require.NoError(t, err)

		usersKey := buildTable(schema, "users")

		usersMap, ok := actual[usersKey]
		require.True(t, ok, fmt.Sprintf("%s map should exist", usersKey))
		require.NotEmpty(t, usersMap)
		_, ok = usersMap["id"]
		require.True(t, ok, "users map should have id column")
	})

	t.Run("GetForeignKeyConstraintsMap", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"
		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)

		constraints, ok := actual.ForeignKeyConstraints[buildTable(schema, "child1")]
		require.True(t, ok)
		require.NotEmpty(t, constraints)

		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"parent_id"},
				NotNullable: []bool{false},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   fmt.Sprintf("%s.parent1", "sqlmanagermysql3"),
					Columns: []string{"id"},
				},
			},
		})

		constraints, ok = actual.ForeignKeyConstraints[buildTable(schema, "t1")]
		require.True(t, ok, "t1 should be in map")
		require.NotEmpty(t, constraints, "should contain t1 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"b"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t1"), Columns: []string{"a"}},
			},
		})

		constraints, ok = actual.ForeignKeyConstraints[buildTable(schema, "t2")]
		require.True(t, ok, "t2 should be in map")
		require.NotEmpty(t, constraints, "should contain t2 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"b"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t3"), Columns: []string{"a"}},
			},
		})

		constraints, ok = actual.ForeignKeyConstraints[buildTable(schema, "t3")]
		require.True(t, ok, "t3 should be in map")
		require.NotEmpty(t, constraints, "should contain t3 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"b"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t2"), Columns: []string{"a"}},
			},
		})
	})

	t.Run("GetForeignKeyConstraintsMap_BasicCircular", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"
		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)

		constraints, ok := actual.ForeignKeyConstraints[buildTable(schema, "t1")]
		require.True(t, ok, "t1 should be in map")
		require.NotEmpty(t, constraints, "should contain t1 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"b"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t1"), Columns: []string{"a"}},
			},
		})

		constraints, ok = actual.ForeignKeyConstraints[buildTable(schema, "t2")]
		require.True(t, ok, "t2 should be in map")
		require.NotEmpty(t, constraints, "should contain t2 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"b"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t3"), Columns: []string{"a"}},
			},
		})

		constraints, ok = actual.ForeignKeyConstraints[buildTable(schema, "t3")]
		require.True(t, ok, "t3 should be in map")
		require.NotEmpty(t, constraints, "should contain t3 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"b"},
				NotNullable: []bool{false},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t2"), Columns: []string{"a"}},
			},
		})
	})

	t.Run("GetForeignKeyConstraintsMap_Composite", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)

		constraints, ok := actual.ForeignKeyConstraints[buildTable(schema, "t5")]
		require.True(t, ok, "t5 should be in map")
		require.NotEmpty(t, constraints, "should contain t5 constraints")
		containsSubset(t, constraints, []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"x", "y"},
				NotNullable: []bool{true, true},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: buildTable(schema, "t4"), Columns: []string{"a", "b"}},
			},
		})
	})

	t.Run("GetPrimaryKeyConstraintsMap", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		require.NotEmpty(t, actual.PrimaryKeyConstraints)

		pkeys, ok := actual.PrimaryKeyConstraints[buildTable(schema, "users")]
		require.True(t, ok, " had no entries")
		require.ElementsMatch(t, []string{"id"}, pkeys)

		pkeys, ok = actual.PrimaryKeyConstraints[buildTable(schema, "t4")]
		require.True(t, ok, "t4 had no entries")
		require.ElementsMatch(t, []string{"a", "b"}, pkeys)
	})

	t.Run("GetUniqueConstraintsMap", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		require.NotEmpty(t, actual.UniqueConstraints)

		uniques, ok := actual.UniqueConstraints[buildTable(schema, "unique_emails")]
		require.True(t, ok, "unique_emails had no entries")
		require.ElementsMatch(t, [][]string{{"email"}}, uniques)
	})

	t.Run("GetUniqueConstraintsMap_Composite", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		require.NotEmpty(t, actual.UniqueConstraints)

		uniques, ok := actual.UniqueConstraints[buildTable(schema, "unique_emails_and_usernames")]
		require.True(t, ok, "unique_emails_and_usernames had no entries")
		require.Len(t, uniques, 1)
		entry := uniques[0]
		require.ElementsMatch(t, []string{"email", "username"}, entry)
	})

	t.Run("GetRolePermissionsMap", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

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
		require.Contains(t, usersRecord, "TRIGGER")
	})

	t.Run("GetCreateTableStatement", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		actual, err := manager.GetCreateTableStatement(context.Background(), schema, "users")
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		_, err = target.DB.ExecContext(context.Background(), actual)
		require.NoError(t, err)
	})

	t.Run("GetTableInitStatements", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql4"

		actual, err := manager.GetTableInitStatements(
			context.Background(),
			[]*sqlmanager_shared.SchemaTable{
				{Schema: schema, Table: "parent1"},
				{Schema: schema, Table: "child1"},
				{Schema: schema, Table: "order"},
				{Schema: schema, Table: "test_mixed_index"},
			},
		)

		require.NoError(t, err)
		require.NotEmpty(t, actual)
		for _, stmt := range actual {
			_, err = target.DB.ExecContext(context.Background(), stmt.CreateTableStatement)
			require.NoError(t, err)
		}
		for _, stmt := range actual {
			t.Logf("statements %+v", stmt.IndexStatements)
			for _, index := range stmt.IndexStatements {
				t.Logf("executing index statement %q", index)
				_, err = target.DB.ExecContext(context.Background(), index)
				require.NoError(t, err)
			}
		}
		for _, stmt := range actual {
			for _, alter := range stmt.AlterTableStatements {
				_, err = target.DB.ExecContext(context.Background(), alter.Statement)
				require.NoError(t, err)
			}
		}
	})

	t.Run("Exec", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		err := manager.Exec(context.Background(), fmt.Sprintf("SELECT 1 FROM %s.%s", schema, "users"))
		require.NoError(t, err)
	})

	t.Run("BatchExec", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		stmt := fmt.Sprintf("SELECT 1 FROM %s.%s;", schema, "users")
		err := manager.BatchExec(context.Background(), 2, []string{stmt, stmt, stmt}, &sqlmanager_shared.BatchExecOpts{})
		require.NoError(t, err)
	})

	t.Run("BatchExec_With_Prefix", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		stmt := fmt.Sprintf("SELECT 1 FROM %s.%s;", schema, "users")
		err := manager.BatchExec(context.Background(), 2, []string{stmt}, &sqlmanager_shared.BatchExecOpts{
			Prefix: &stmt,
		})
		require.NoError(t, err)
	})

	t.Run("GetSchemaInitStatements", func(t *testing.T) {
		t.Parallel()
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
			// {Schema: schema, Table: "users"},
		})
		require.NoError(t, err)
		require.NotEmpty(t, statements)
		for _, block := range statements {
			t.Logf("executing %d statements for label %q", len(block.Statements), block.Label)
			for _, stmt := range block.Statements {
				_, err = target.DB.ExecContext(ctx, stmt)
				require.NoError(t, err, "failed to execute %s statement %q", block.Label, stmt)
			}
		}
	})

	t.Run("GetSchemaInitStatements_customtable", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		statements, err := manager.GetSchemaInitStatements(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "custom_table"}})
		require.NoError(t, err)
		require.NotEmpty(t, statements)
	})

	t.Run("GetSchemaTableTriggers_customtable", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		triggers, err := manager.GetSchemaTableTriggers(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "employee_log"}})
		require.NoError(t, err)
		require.NotEmpty(t, triggers)
	})

	t.Run("GetSchemaTableDataTypes_customtable", func(t *testing.T) {
		t.Parallel()
		schema := "sqlmanagermysql3"

		resp, err := manager.GetSchemaTableDataTypes(context.Background(), []*sqlmanager_shared.SchemaTable{{Schema: schema, Table: "custom_table"}})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmptyf(t, resp.GetStatements(), "statements")
		require.NotEmptyf(t, resp.Functions, "functions")
	})
}

func setup(ctx context.Context, containers *tcmysql.MysqlTestSyncContainer) error {
	baseDir := "testdata"

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		err := containers.Source.RunSqlFiles(errctx, &baseDir, []string{"setup.sql"})
		if err != nil {
			return fmt.Errorf("encountered error when executing source setup statement: %w", err)
		}
		return nil
	})
	errgrp.Go(func() error {
		err := containers.Target.RunSqlFiles(errctx, &baseDir, []string{"init.sql"})
		if err != nil {
			return fmt.Errorf("encountered error when executing dest setup statement: %w", err)
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return err
	}

	return nil
}

func containsSubset[T any](t testing.TB, array, subset []T) {
	t.Helper()
	for _, elem := range subset {
		require.Contains(t, array, elem)
	}
}

func buildTable(schema, tableName string) string {
	return fmt.Sprintf("%s.%s", schema, tableName)
}
