package tabledependency

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func Test_buildDependencies(t *testing.T) {
	t.Run("Simple dependency map", func(t *testing.T) {
		dependencyMap := map[string][]*sqlmanager_shared.ForeignConstraint{
			"table1": {
				{
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "table2",
						Columns: []string{"col1"},
					},
					NotNullable: []bool{false},
				},
			},
		}
		tableColumnsMap := map[string][]string{
			"table1": {"col1", "col2"},
			"table2": {"col1", "col2"},
		}

		result := buildDependencies(dependencyMap, tableColumnsMap)

		expectedFilteredDeps := map[string][]string{
			"table1": {"table2"},
		}
		expectedForeignKeys := map[string]map[string][]string{
			"table1": {"table2": {"col1"}},
		}
		expectedForeignKeyCols := map[string]map[string]*ConstraintColumns{
			"table1": {
				"table2": {
					NullableColumns:    []string{"col1"},
					NonNullableColumns: []string{},
				},
			},
		}

		require.Equal(t, expectedFilteredDeps, result.filteredDeps, "FilteredDeps mismatch")
		require.Equal(t, expectedForeignKeys, result.foreignKeys, "ForeignKeys mismatch")
		require.Equal(t, expectedForeignKeyCols, result.foreignKeyCols, "ForeignKeyCols mismatch")
	})

	t.Run("Multiple dependencies", func(t *testing.T) {
		dependencyMap := map[string][]*sqlmanager_shared.ForeignConstraint{
			"table1": {
				{
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "table2",
						Columns: []string{"col1", "col2"},
					},
					NotNullable: []bool{true, false},
				},
				{
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "table3",
						Columns: []string{"col3"},
					},
					NotNullable: []bool{true},
				},
			},
		}
		tableColumnsMap := map[string][]string{
			"table1": {"col1", "col2", "col3"},
			"table2": {"col1", "col2"},
			"table3": {"col3"},
		}

		result := buildDependencies(dependencyMap, tableColumnsMap)

		expectedFilteredDeps := map[string][]string{
			"table1": {"table2", "table3"},
		}
		expectedForeignKeys := map[string]map[string][]string{
			"table1": {
				"table2": {"col1", "col2"},
				"table3": {"col3"},
			},
		}
		expectedForeignKeyCols := map[string]map[string]*ConstraintColumns{
			"table1": {
				"table2": {
					NullableColumns:    []string{"col2"},
					NonNullableColumns: []string{"col1"},
				},
				"table3": {
					NullableColumns:    []string{},
					NonNullableColumns: []string{"col3"},
				},
			},
		}

		require.Equal(t, expectedFilteredDeps, result.filteredDeps, "FilteredDeps mismatch")
		require.Equal(t, expectedForeignKeys, result.foreignKeys, "ForeignKeys mismatch")
		require.Equal(t, expectedForeignKeyCols, result.foreignKeyCols, "ForeignKeyCols mismatch")
	})

	t.Run("Empty dependency map", func(t *testing.T) {
		dependencyMap := map[string][]*sqlmanager_shared.ForeignConstraint{}
		tableColumnsMap := map[string][]string{}

		result := buildDependencies(dependencyMap, tableColumnsMap)

		require.Empty(t, result.filteredDeps, "Expected empty FilteredDeps")
		require.Empty(t, result.foreignKeys, "Expected empty ForeignKeys")
		require.Empty(t, result.foreignKeyCols, "Expected empty ForeignKeyCols")
	})

	t.Run("Missing table in tableColumnsMap", func(t *testing.T) {
		dependencyMap := map[string][]*sqlmanager_shared.ForeignConstraint{
			"table1": {
				{
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "table2",
						Columns: []string{"col1"},
					},
					NotNullable: []bool{false},
				},
			},
		}
		tableColumnsMap := map[string][]string{
			"table1": {"col1", "col2"},
			// table2 is missing
		}

		result := buildDependencies(dependencyMap, tableColumnsMap)

		require.Empty(t, result.filteredDeps, "Expected empty FilteredDeps")
		require.Len(t, result.foreignKeys, 1, "Expected ForeignKeys with 1 entry")
		require.Len(t, result.foreignKeyCols, 1, "Expected ForeignKeyCols with 1 entry")
	})
}

func Test_processTableConstraints(t *testing.T) {
	deps := &dependencies{
		filteredDeps:   make(map[string][]string),
		foreignKeys:    make(map[string]map[string][]string),
		foreignKeyCols: make(map[string]map[string]*ConstraintColumns),
	}
	table := "table1"
	constraints := []*sqlmanager_shared.ForeignConstraint{
		{
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Table:   "table2",
				Columns: []string{"col1"},
			},
			NotNullable: []bool{false},
		},
	}
	tableColumnsMap := map[string][]string{
		"table1": {"col1", "col2"},
		"table2": {"col1", "col2"},
	}

	processTableConstraints(deps, table, constraints, tableColumnsMap)

	require.Len(t, deps.foreignKeys[table], 1, "Expected 1 foreign key entry for table1")
	require.Len(t, deps.foreignKeyCols[table], 1, "Expected 1 foreign key column entry for table1")
}

func Test_updateForeignKeyMaps(t *testing.T) {
	deps := &dependencies{
		foreignKeys: map[string]map[string][]string{
			"table1": {"table2": []string{}},
		},
		foreignKeyCols: map[string]map[string]*ConstraintColumns{
			"table1": {
				"table2": &ConstraintColumns{
					NullableColumns:    []string{},
					NonNullableColumns: []string{},
				},
			},
		},
	}
	table := "table1"
	fkTable := "table2"
	col := "col1"

	t.Run("Nullable column", func(t *testing.T) {
		updateForeignKeyMaps(deps, table, fkTable, col, false)

		require.Len(t, deps.foreignKeys[table][fkTable], 1, "Expected 1 foreign key column")
		require.Len(t, deps.foreignKeyCols[table][fkTable].NullableColumns, 1, "Expected 1 nullable column")
	})

	t.Run("Non-nullable column", func(t *testing.T) {
		updateForeignKeyMaps(deps, table, fkTable, "col2", true)

		require.Len(t, deps.foreignKeys[table][fkTable], 2, "Expected 2 foreign key columns")
		require.Len(t, deps.foreignKeyCols[table][fkTable].NonNullableColumns, 1, "Expected 1 non-nullable column")
	})
}

func Test_deduplicateFilteredDeps(t *testing.T) {
	deps := &dependencies{
		filteredDeps: map[string][]string{
			"table1": {"table2", "table3", "table2"},
		},
	}

	deduplicateFilteredDeps(deps)

	require.Len(t, deps.filteredDeps["table1"], 2, "Expected 2 unique dependencies")
	require.Equal(t, []string{"table2", "table3"}, deps.filteredDeps["table1"], "Expected [table2, table3]")
}
