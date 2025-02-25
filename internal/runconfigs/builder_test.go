package runconfigs

// import (
// 	"testing"

// 	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
// 	"github.com/stretchr/testify/assert"
// )

// // TestRunConfigBuilder_Build covers the high-level entry point for building RunConfigs.
// func TestRunConfigBuilder_Build(t *testing.T) {
// 	t.Run("NoForeignKeys_NoCircularDependency", func(t *testing.T) {
// 		builder := newRunConfigBuilder(
// 			"test_table",
// 			[]string{"col1", "col2"},
// 			[]string{"col1"}, // primary key
// 			nil,              // where clause
// 			nil,              // unique indexes
// 			nil,              // unique constraints
// 			nil,              // foreign keys
// 			false,            // isPartOfCircularDependency
// 		)

// 		configs := builder.Build()
// 		assert.Len(t, configs, 1, "Should only have 1 config for a non-circular table with no foreign keys.")

// 		cfg := configs[0]
// 		assert.Equal(t, RunTypeInsert, cfg.runType, "RunType should be Insert.")
// 		assert.Equal(t, "test_table", cfg.table, "Table should match.")
// 		assert.ElementsMatch(t, []string{"col1", "col2"}, cfg.selectColumns, "Select columns should match all input columns.")
// 		assert.ElementsMatch(t, []string{"col1", "col2"}, cfg.insertColumns, "Insert columns should match all input columns.")
// 		assert.ElementsMatch(t, []string{"col1"}, cfg.primaryKeys, "Primary key should match provided PKs.")
// 		assert.Nil(t, cfg.whereClause, "Where clause should be nil if not provided.")
// 		assert.Empty(t, cfg.dependsOn, "No foreign keys -> no dependencies.")
// 		assert.Empty(t, cfg.foreignKeys, "No foreign keys -> no foreignKeys in config.")
// 		assert.ElementsMatch(t, []string{"col1"}, cfg.orderByColumns,
// 			"With a primary key, orderByColumns should be the primary key.")
// 	})

// 	t.Run("ForeignKeys_NoCircularDependency", func(t *testing.T) {
// 		// Example: table has a single NOT NULL foreign key to reference_table
// 		fk := &sqlmanager_shared.ForeignConstraint{
// 			Columns:     []string{"fk_col"},
// 			NotNullable: []bool{true}, // Not-nullable
// 			ForeignKey: &sqlmanager_shared.ForeignKey{
// 				Table:   "reference_table",
// 				Columns: []string{"id"},
// 			},
// 		}

// 		builder := newRunConfigBuilder(
// 			"test_table",
// 			[]string{"col1", "fk_col", "col2"},
// 			[]string{"col1"}, // primary key
// 			nil,              // where clause
// 			nil,              // unique indexes
// 			nil,              // unique constraints
// 			[]*sqlmanager_shared.ForeignConstraint{fk},
// 			false, // isPartOfCircularDependency
// 		)

// 		configs := builder.Build()
// 		assert.Len(t, configs, 1, "Should only have 1 config for a non-circular table.")

// 		cfg := configs[0]
// 		assert.Equal(t, RunTypeInsert, cfg.runType)
// 		assert.Equal(t, "test_table", cfg.table)
// 		assert.ElementsMatch(t, []string{"col1", "fk_col", "col2"}, cfg.selectColumns, "Select columns should be all columns.")
// 		assert.ElementsMatch(t, []string{"col1", "fk_col", "col2"}, cfg.insertColumns, "All columns get inserted at once for non-circular FKs.")
// 		assert.ElementsMatch(t, []string{"col1"}, cfg.primaryKeys)
// 		assert.Equal(t, 1, len(cfg.dependsOn), "Should have 1 dependency because of the single foreign key.")
// 		assert.Equal(t, "reference_table", cfg.dependsOn[0].Table, "Dependency table should match foreign key table.")
// 		assert.ElementsMatch(t, []string{"id"}, cfg.dependsOn[0].Columns, "Dependency columns should match reference columns.")
// 		assert.Len(t, cfg.foreignKeys, 1, "We should have 1 foreign key in the config.")
// 		assert.ElementsMatch(t, []string{"fk_col"}, cfg.foreignKeys[0].Columns)
// 		assert.ElementsMatch(t, []bool{true}, cfg.foreignKeys[0].NotNullable)
// 		assert.ElementsMatch(t, []string{"id"}, cfg.foreignKeys[0].ReferenceColumns)
// 		assert.Equal(t, "reference_table", cfg.foreignKeys[0].ReferenceTable)
// 		assert.ElementsMatch(t, []string{"col1"}, cfg.orderByColumns)
// 	})

// 	t.Run("ForeignKeys_CircularDependency", func(t *testing.T) {
// 		// Example: table has columns col1 (primary key), fk_col1 (NOT NULL to another table)
// 		// and fk_col2 (NULLABLE to another table). This triggers the circular dependency logic.
// 		fk := &sqlmanager_shared.ForeignConstraint{
// 			Columns:     []string{"fk_col1", "fk_col2"},
// 			NotNullable: []bool{true, false}, // first is not-null, second is nullable
// 			ForeignKey: &sqlmanager_shared.ForeignKey{
// 				Table:   "other_table",
// 				Columns: []string{"other_id1", "other_id2"},
// 			},
// 		}

// 		builder := newRunConfigBuilder(
// 			"test_table",
// 			[]string{"col1", "fk_col1", "fk_col2", "some_data"},
// 			[]string{"col1"},                // primary key
// 			nil,                             // where clause
// 			nil,                             // unique indexes
// 			[][]string{{"col1", "fk_col1"}}, // unique constraints example
// 			[]*sqlmanager_shared.ForeignConstraint{fk},
// 			true, // isPartOfCircularDependency
// 		)

// 		configs := builder.Build()
// 		assert.Len(t, configs, 2, "Should have 2 configs: 1 Insert + 1 Update for the nullable FK.")

// 		insertCfg := configs[0]
// 		updateCfg := configs[1]

// 		// Insert Config checks
// 		assert.Equal(t, RunTypeInsert, insertCfg.runType, "First config should be Insert.")
// 		assert.ElementsMatch(t, []string{"col1", "fk_col1", "fk_col2", "some_data"}, insertCfg.selectColumns,
// 			"Insert's select columns should have all for S3 staging.")
// 		// Insert columns should contain the primary key + not-null FK columns, plus any non-FK columns that remain
// 		// but the "fk_col2" is nullable, so it should be handled in the update step.
// 		// However, "some_data" is not part of any FK constraint, so it remains in the insert set.
// 		assert.ElementsMatch(t, []string{"col1", "fk_col1", "some_data"}, insertCfg.insertColumns,
// 			"Insert columns include PKs, non-null FKs, and any non-FK columns, but skip nullable FKs.")
// 		assert.ElementsMatch(t, []string{"col1"}, insertCfg.primaryKeys)
// 		assert.ElementsMatch(t, []string{"col1"}, insertCfg.orderByColumns,
// 			"For circular dependencies, we still prioritize PKs for ordering.")
// 		assert.Equal(t, "other_table", insertCfg.dependsOn[0].Table,
// 			"Insert config depends on the other table for the NOT NULL foreign key column.")
// 		assert.ElementsMatch(t, []string{"other_id1"}, insertCfg.dependsOn[0].Columns,
// 			"Insert config depends on the corresponding reference column for the NOT NULL foreign key.")
// 		// Check foreignKeys in the insert config
// 		assert.Len(t, insertCfg.foreignKeys, 1)
// 		assert.ElementsMatch(t, []string{"fk_col1"}, insertCfg.foreignKeys[0].Columns)
// 		assert.ElementsMatch(t, []bool{true}, insertCfg.foreignKeys[0].NotNullable)
// 		assert.ElementsMatch(t, []string{"other_id1"}, insertCfg.foreignKeys[0].ReferenceColumns)
// 		assert.Equal(t, "other_table", insertCfg.foreignKeys[0].ReferenceTable)

// 		// Update Config checks
// 		assert.Equal(t, RunTypeUpdate, updateCfg.runType, "Second config should be Update for nullable FK.")
// 		assert.ElementsMatch(t, []string{"col1", "fk_col2"}, updateCfg.selectColumns,
// 			"Update config needs primary keys + the columns to update.")
// 		assert.ElementsMatch(t, []string{"fk_col2"}, updateCfg.insertColumns,
// 			"Update config will only update the nullable foreign key column(s).")
// 		assert.ElementsMatch(t, []string{"col1"}, updateCfg.primaryKeys)
// 		assert.ElementsMatch(t, []string{"col1"}, updateCfg.orderByColumns)
// 		// dependsOn includes the foreign table for the FK, and also the same table's PK if needed
// 		assert.Equal(t, "other_table", updateCfg.dependsOn[0].Table)
// 		assert.ElementsMatch(t, []string{"other_id2"}, updateCfg.dependsOn[0].Columns,
// 			"Update depends on the second reference column for the nullable FK.")
// 		assert.Equal(t, "test_table", updateCfg.dependsOn[1].Table,
// 			"Update also depends on the table's primary key (the same table).")
// 		assert.ElementsMatch(t, []string{"col1"}, updateCfg.dependsOn[1].Columns)
// 		// foreignKeys in the update config
// 		assert.Len(t, updateCfg.foreignKeys, 1)
// 		assert.ElementsMatch(t, []string{"fk_col2"}, updateCfg.foreignKeys[0].Columns)
// 		assert.ElementsMatch(t, []bool{false}, updateCfg.foreignKeys[0].NotNullable)
// 		assert.ElementsMatch(t, []string{"other_id2"}, updateCfg.foreignKeys[0].ReferenceColumns)
// 		assert.Equal(t, "other_table", updateCfg.foreignKeys[0].ReferenceTable)
// 	})

// 	t.Run("NoPrimaryKey_NoCircularDependency", func(t *testing.T) {
// 		// If a table has no primary keys but unique constraints or unique indexes,
// 		// getOrderByColumns picks them. Then the rest of the logic is straightforward.
// 		builder := newRunConfigBuilder(
// 			"no_pk_table",
// 			[]string{"col1", "col2"},
// 			[]string{},           // no primary keys
// 			nil,                  // where clause
// 			[][]string{{"col2"}}, // unique indexes
// 			nil,                  // unique constraints
// 			nil,                  // foreign keys
// 			false,                // not part of circular dependency
// 		)
// 		configs := builder.Build()
// 		assert.Len(t, configs, 1)
// 		cfg := configs[0]
// 		assert.Equal(t, RunTypeInsert, cfg.runType)
// 		assert.Equal(t, "no_pk_table", cfg.table)
// 		assert.ElementsMatch(t, []string{"col1", "col2"}, cfg.selectColumns)
// 		assert.ElementsMatch(t, []string{"col1", "col2"}, cfg.insertColumns)
// 		assert.Empty(t, cfg.primaryKeys)
// 		assert.ElementsMatch(t, []string{"col2"}, cfg.orderByColumns,
// 			"When no PK, getOrderByColumns should pick unique indexes if present.")
// 		assert.Empty(t, cfg.dependsOn)
// 		assert.Empty(t, cfg.foreignKeys)
// 	})
// }
