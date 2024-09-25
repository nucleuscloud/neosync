package sqlserver

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func Test_filterIdentityColumns(t *testing.T) {
	t.Run("Non-MSSQL driver", func(t *testing.T) {
		driver := "mysql"
		identityCols := []string{"id"}
		columnNames := []string{"id", "name", "age"}
		argRows := [][]any{{1, "Alice", 30}, {2, "Bob", 25}}

		gotCols, gotRows := FilterOutSqlServerDefaultIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, columnNames, gotCols, "Columns should remain unchanged for non-MSSQL driver")
		require.Equal(t, argRows, gotRows, "Rows should remain unchanged for non-MSSQL driver")
	})

	t.Run("MSSQL driver with identity columns", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{"id"}
		columnNames := []string{"id", "name", "age"}
		argRows := [][]any{{1, "Alice", 30}, {2, "Bob", 25}}

		gotCols, gotRows := FilterOutSqlServerDefaultIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, []string{"name", "age"}, gotCols, "Identity column should be removed")
		require.Equal(t, [][]any{{"Alice", 30}, {"Bob", 25}}, gotRows, "Identity column values should be removed")
	})

	t.Run("MSSQL driver with DEFAULT value", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{"id"}
		columnNames := []string{"id", "name", "age", "city"}
		argRows := [][]any{{"DEFAULT", "Alice", 30, "DEFAULT"}, {"DEFAULT", "Bob", 25, "DEFAULT"}}

		gotCols, gotRows := FilterOutSqlServerDefaultIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, []string{"name", "age", "city"}, gotCols, "All columns should be present when DEFAULT is used")
		require.Equal(t, [][]any{{"Alice", 30, "DEFAULT"}, {"Bob", 25, "DEFAULT"}}, gotRows, "All rows should remain unchanged when DEFAULT is used")
	})

	t.Run("Empty identity columns", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{}
		columnNames := []string{"id", "name", "age"}
		argRows := [][]any{{1, "Alice", 30}, {2, "Bob", 25}}

		gotCols, gotRows := FilterOutSqlServerDefaultIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, columnNames, gotCols, "Columns should remain unchanged with empty identity columns")
		require.Equal(t, argRows, gotRows, "Rows should remain unchanged with empty identity columns")
	})

	t.Run("Multiple identity columns", func(t *testing.T) {
		driver := sqlmanager_shared.MssqlDriver
		identityCols := []string{"id", "created_at"}
		columnNames := []string{"id", "name", "age", "created_at"}
		argRows := [][]any{{"DEFAULT", "Alice", 30, "DEFAULT"}, {"DEFAULT", "Bob", 25, "DEFAULT"}}

		gotCols, gotRows := FilterOutSqlServerDefaultIdentityColumns(driver, identityCols, columnNames, argRows)

		require.Equal(t, []string{"name", "age"}, gotCols, "Multiple identity columns should be removed")
		require.Equal(t, [][]any{{"Alice", 30}, {"Bob", 25}}, gotRows, "Multiple identity column values should be removed")
	})
}

func Test_GoTypeToSqlServerType(t *testing.T) {
	t.Run("GoTypeToSqlServerType", func(t *testing.T) {
		t.Run("Empty input", func(t *testing.T) {
			input := [][]any{}
			result := GoTypeToSqlServerType(input)
			require.Equal(t, [][]any{}, result)
		})

		t.Run("Single row with no boolean", func(t *testing.T) {
			input := [][]any{{1, "test", 3.14}}
			expected := [][]any{{1, "test", 3.14}}
			result := GoTypeToSqlServerType(input)
			require.Equal(t, expected, result)
		})

		t.Run("Single row with boolean", func(t *testing.T) {
			input := [][]any{{true, false, "test"}}
			expected := [][]any{{1, 0, "test"}}
			result := GoTypeToSqlServerType(input)
			require.Equal(t, expected, result)
		})

		t.Run("Multiple rows with mixed types", func(t *testing.T) {
			input := [][]any{
				{1, true, "test1"},
				{2, false, "test2"},
				{3, true, "test3"},
			}
			expected := [][]any{
				{1, 1, "test1"},
				{2, 0, "test2"},
				{3, 1, "test3"},
			}
			result := GoTypeToSqlServerType(input)
			require.Equal(t, expected, result)
		})
	})
}
