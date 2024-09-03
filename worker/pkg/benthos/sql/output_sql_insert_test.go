package neosync_benthos_sql

import (
	"context"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_SqlInsertOutputEmptyShutdown(t *testing.T) {
	conf := `
driver: postgres
dsn: foo
schema: bar
table: baz
args_mapping: 'root = [this.id]'
`
	spec := sqlInsertOutputSpec()
	env := service.NewEnvironment()

	insertConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	insertOutput, err := newInsertOutput(insertConfig, service.MockResources(), nil, false)
	require.NoError(t, err)
	require.NoError(t, insertOutput.Close(context.Background()))
}

func Test_filterIdentityColumns(t *testing.T) {
	t.Run("Non-MSSQL driver", func(t *testing.T) {
		driver := "mysql"
		identityCols := []string{"id"}
		columnNames := []string{"id", "name", "age"}
		argRows := [][]any{{1, "Alice", 30}, {2, "Bob", 25}}

		gotCols, gotRows := filterIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, columnNames, gotCols, "Columns should remain unchanged for non-MSSQL driver")
		require.Equal(t, argRows, gotRows, "Rows should remain unchanged for non-MSSQL driver")
	})

	t.Run("MSSQL driver with identity columns", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{"id"}
		columnNames := []string{"id", "name", "age"}
		argRows := [][]any{{1, "Alice", 30}, {2, "Bob", 25}}

		gotCols, gotRows := filterIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, []string{"id", "name", "age"}, gotCols, "Identity column should be removed")
		require.Equal(t, [][]any{{1, "Alice", 30}, {2, "Bob", 25}}, gotRows, "Identity column values should be removed")
	})

	t.Run("MSSQL driver with DEFAULT value", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{"id"}
		columnNames := []string{"id", "name", "age", "city"}
		argRows := [][]any{{"DEFAULT", "Alice", 30, "DEFAULT"}, {"DEFAULT", "Bob", 25, "DEFAULT"}}

		gotCols, gotRows := filterIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, []string{"name", "age", "city"}, gotCols, "All columns should be present when DEFAULT is used")
		require.Equal(t, [][]any{{"Alice", 30, "DEFAULT"}, {"Bob", 25, "DEFAULT"}}, gotRows, "All rows should remain unchanged when DEFAULT is used")
	})

	t.Run("Empty identity columns", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{}
		columnNames := []string{"id", "name", "age"}
		argRows := [][]any{{1, "Alice", 30}, {2, "Bob", 25}}

		gotCols, gotRows := filterIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, columnNames, gotCols, "Columns should remain unchanged with empty identity columns")
		require.Equal(t, argRows, gotRows, "Rows should remain unchanged with empty identity columns")
	})

	t.Run("Multiple identity columns", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{"id", "created_at"}
		columnNames := []string{"id", "name", "age", "created_at"}
		argRows := [][]any{{"DEFAULT", "Alice", 30, "DEFAULT"}, {"DEFAULT", "Bob", 25, "DEFAULT"}}

		gotCols, gotRows := filterIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, []string{"name", "age"}, gotCols, "Multiple identity columns should be removed")
		require.Equal(t, [][]any{{"Alice", 30}, {"Bob", 25}}, gotRows, "Multiple identity column values should be removed")
	})
}
