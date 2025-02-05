package sqlmanager_postgres

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const schema = "sqlmanagerpostgres@special"
const capitalSchema = "CaPiTaL"

func Test_PostgresManager(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	t.Log("Running integration tests for Postgres Manager")
	t.Parallel()

	ctx := context.Background()
	containers, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
	if err != nil {
		t.Fatal(err)
	}
	source := containers.Source
	target := containers.Target
	t.Log("Successfully created source and target postgres test containers")

	err = setup(ctx, containers)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Successfully setup source and target databases")
	sourceDB, err := sql.Open(sqlmanager_shared.PostgresDriver, source.URL)
	if err != nil {
		t.Fatal(err)
	}
	manager := postgres.NewManager(pg_queries.New(), sourceDB, func() {})

	t.Run("GetDatabaseSchema", func(t *testing.T) {
		t.Parallel()
		expectedSubset := []*sqlmanager_shared.DatabaseSchemaRow{
			{
				TableSchema:            schema,
				TableName:              "users",
				ColumnName:             "id",
				DataType:               "text",
				ColumnDefault:          "",
				IsNullable:             false,
				CharacterMaximumLength: -1,
				NumericPrecision:       -1,
				NumericScale:           -1,
				OrdinalPosition:        1,
			},
		}

		actual, err := manager.GetDatabaseSchema(ctx)
		require.NoError(t, err)
		containsSubset(t, actual, expectedSubset)
	})

	t.Run("GetDatabaseSchema_With_Identity", func(t *testing.T) {
		t.Parallel()
		expectedSubset := []*sqlmanager_shared.DatabaseSchemaRow{
			{
				TableSchema:            schema,
				TableName:              "users_with_identity",
				ColumnName:             "id",
				DataType:               "integer",
				ColumnDefault:          "",
				IsNullable:             false,
				CharacterMaximumLength: -1,
				NumericPrecision:       32,
				NumericScale:           0,
				OrdinalPosition:        1,
				IdentityGeneration:     sqlmanager_shared.Ptr("a"),
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

		usersKey := buildTable(schema, "users")

		usersMap, ok := actual[usersKey]
		require.True(t, ok, fmt.Sprintf("%s map should exist", usersKey))
		require.NotEmpty(t, usersMap)
		_, ok = usersMap["id"]
		require.True(t, ok, "users map should have id column")
	})

	t.Run("GetForeignKeyConstraintsMap", func(t *testing.T) {
		t.Parallel()
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
					Table:   fmt.Sprintf("%s.parent1", schema),
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
		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)

		// Test t1 circular reference
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

		// Test t2->t3->t2 circular reference
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
		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		require.NotEmpty(t, actual.PrimaryKeyConstraints)

		pkeys, ok := actual.PrimaryKeyConstraints[buildTable(schema, "users_with_identity")]
		require.True(t, ok, "users_with_identity had no entries")
		require.ElementsMatch(t, []string{"id"}, pkeys)

		pkeys, ok = actual.PrimaryKeyConstraints[buildTable(schema, "t4")]
		require.True(t, ok, "t4 had no entries")
		require.ElementsMatch(t, []string{"a", "b"}, pkeys)
	})

	t.Run("GetUniqueConstraintsMap", func(t *testing.T) {
		t.Parallel()
		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		require.NotEmpty(t, actual.UniqueConstraints)

		uniques, ok := actual.UniqueConstraints[buildTable(schema, "unique_emails")]
		require.True(t, ok, "unique_emails had no entries")
		require.ElementsMatch(t, [][]string{{"email"}}, uniques)
	})

	t.Run("GetUniqueConstraintsMap_Composite", func(t *testing.T) {
		t.Parallel()
		actual, err := manager.GetTableConstraintsBySchema(ctx, []string{schema})
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
		actual, err := manager.GetRolePermissionsMap(context.Background())
		require.NoError(t, err)
		require.NotEmpty(t, actual)

		usersKey := buildTable(schema, "users")
		usersRecord, ok := actual[usersKey]
		require.True(t, ok, "map should have users perms")
		require.ElementsMatch(t, []string{
			"INSERT", "SELECT", "UPDATE", "DELETE",
			"TRUNCATE", "REFERENCES", "TRIGGER",
		}, usersRecord)
	})

	t.Run("GetCreateTableStatement", func(t *testing.T) {
		t.Parallel()
		actual, err := manager.GetCreateTableStatement(context.Background(), schema, "users")
		require.NoError(t, err)
		require.NotEmpty(t, actual)
		_, err = target.DB.Exec(ctx, actual)
		require.NoError(t, err)
	})

	t.Run("Exec", func(t *testing.T) {
		t.Parallel()
		sql, _, err := goqu.Dialect("postgres").Select("*").From(goqu.T("users").Schema(schema)).ToSQL()
		require.NoError(t, err)

		err = manager.Exec(context.Background(), sql)
		require.NoError(t, err)
	})

	t.Run("BatchExec", func(t *testing.T) {
		t.Parallel()
		sql, _, err := goqu.Dialect("postgres").Select("*").From(goqu.T("users").Schema(schema)).ToSQL()
		require.NoError(t, err)
		sql += ";"

		err = manager.BatchExec(context.Background(), 2, []string{sql, sql, sql}, &sqlmanager_shared.BatchExecOpts{})
		require.NoError(t, err)
	})

	t.Run("BatchExec_With_Prefix", func(t *testing.T) {
		t.Parallel()
		sql, _, err := goqu.Dialect("postgres").Select("*").From(goqu.T("users").Schema(schema)).ToSQL()
		require.NoError(t, err)
		sql += ";"

		err = manager.BatchExec(context.Background(), 2, []string{sql}, &sqlmanager_shared.BatchExecOpts{
			Prefix: &sql,
		})
		require.NoError(t, err)
	})

	t.Run("GetTableRowCount", func(t *testing.T) {
		t.Parallel()
		table := "tablewithcount"

		count, err := manager.GetTableRowCount(context.Background(), schema, table, nil)
		require.NoError(t, err)
		require.Equal(t, count, int64(2))

		where := "id = '1'"
		count, err = manager.GetTableRowCount(context.Background(), schema, table, &where)
		require.NoError(t, err)
		require.Equal(t, count, int64(1))

		where = "id = 'doesnotexist'"
		count, err = manager.GetTableRowCount(context.Background(), schema, table, &where)
		require.NoError(t, err)
		require.Equal(t, count, int64(0))
	})

	t.Run("GetPostgresColumnOverrideAndResetProperties", func(t *testing.T) {
		t.Parallel()
		colInfoMap, err := manager.GetSchemaColumnMap(context.Background())
		require.NoError(t, err)

		testDefaultTable := colInfoMap["testdb.sqlmanagerpostgres2.defaults_table"]

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
			needsOverride, needsReset := postgres.GetPostgresColumnOverrideAndResetProperties(colInfo)
			expected, ok := expectedProperties[col]
			require.Truef(t, ok, "Missing expected column %q", col)
			require.Equalf(t, expected.needsOverride, needsOverride, "Incorrect needsOverride value for column %q", col)
			require.Equalf(t, expected.needsReset, needsReset, "Incorrect needsReset value for column %q", col)
		}
	})

	t.Run("GetSchemaInitStatements", func(t *testing.T) {
		t.Parallel()
		schematables, err := getSchemaTables(ctx, manager)
		require.NoError(t, err)
		require.NotEmpty(t, schematables)

		statements, err := manager.GetSchemaInitStatements(ctx, schematables)
		require.NoError(t, err)
		require.NotEmpty(t, statements)

		allStmts := []string{}
		for _, block := range statements {
			allStmts = append(allStmts, block.Statements...)
		}
		require.NotEmpty(t, allStmts)

		for _, block := range statements {
			t.Logf("executing %d statements for label %q", len(block.Statements), block.Label)
			for _, stmt := range block.Statements {
				_, err = target.DB.Exec(ctx, stmt)
				require.NoError(t, err, "failed to execute %s statement %q", block.Label, stmt)
			}
		}
	})

	t.Run("BuildPgIdentityColumnResetCurrentSql", func(t *testing.T) {
		t.Parallel()
		table := "BadName"

		sql := postgres.BuildPgIdentityColumnResetCurrentSql(capitalSchema, table, "ID")
		require.NoError(t, err)
		_, err = source.DB.Exec(ctx, sql)
		require.NoError(t, err, "failed to execute statement %q", sql)
	})

	t.Log("Finished running postgres manager integration tests")
	t.Cleanup(func() {
		t.Log("Cleaning up source and target postgres containers")
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

func getSchemaTables(ctx context.Context, manager *postgres.PostgresManager) ([]*sqlmanager_shared.SchemaTable, error) {
	cols, err := manager.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	schematableMap := map[string]*sqlmanager_shared.SchemaTable{}
	for _, col := range cols {
		schematable := &sqlmanager_shared.SchemaTable{Schema: col.TableSchema, Table: col.TableName}
		schematableMap[buildTable(col.TableSchema, col.TableName)] = schematable
	}
	return slices.Collect(maps.Values(schematableMap)), nil
}

func setup(ctx context.Context, containers *tcpostgres.PostgresTestSyncContainer) error {
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
		err := containers.Target.RunSqlFiles(errctx, &baseDir, []string{"create-schemas.sql"})
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

type testColumnProperties struct {
	needsOverride bool
	needsReset    bool
}
