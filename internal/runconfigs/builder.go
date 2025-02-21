package runconfigs

import (
	"fmt"
	"slices"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

// RunConfigBuilder is responsible for generating RunConfigs that define how to process table data.
// It handles two main scenarios:
// 1. Tables without circular dependencies - generates a single INSERT config
// 2. Tables with circular dependencies - generates multiple configs to handle the cycle:
//   - Initial INSERT with non-nullable foreign key columns
//   - UPDATE configs for each nullable foreign key reference
//
// This allows for properly ordered data synchronization while maintaining referential integrity.

type runConfigBuilder struct {
	table                      string
	primaryKeys                []string
	whereClause                *string
	columns                    []string
	uniqueIndexes              [][]string
	uniqueConstraints          [][]string
	foreignKeys                []*sqlmanager_shared.ForeignConstraint
	isPartOfCircularDependency bool
}

func newRunConfigBuilder(
	table string,
	columns []string,
	primaryKeys []string,
	whereClause *string,
	uniqueIndexes [][]string,
	uniqueConstraints [][]string,
	foreignKeys []*sqlmanager_shared.ForeignConstraint,
	isPartOfCircularDependency bool,
) *runConfigBuilder {
	return &runConfigBuilder{
		table:                      table,
		primaryKeys:                primaryKeys,
		whereClause:                whereClause,
		columns:                    columns,
		uniqueIndexes:              uniqueIndexes,
		uniqueConstraints:          uniqueConstraints,
		foreignKeys:                foreignKeys,
		isPartOfCircularDependency: isPartOfCircularDependency,
	}
}

func (b *runConfigBuilder) Build() []*RunConfig {
	if b.isPartOfCircularDependency {
		return b.buildCircularDependencyConfigs()
	} else {
		return []*RunConfig{b.buildInsertConfig()}
	}
}

func (b *runConfigBuilder) buildInsertConfig() *RunConfig {
	config := &RunConfig{
		id:             fmt.Sprintf("%s.%s", b.table, RunTypeInsert),
		table:          b.table,
		runType:        RunTypeInsert,
		selectColumns:  b.columns,
		insertColumns:  b.columns,
		primaryKeys:    b.primaryKeys,
		whereClause:    b.whereClause,
		orderByColumns: b.getOrderByColumns(b.columns),
		dependsOn:      b.getDependsOn(),
		foreignKeys:    b.getForeignKeys(b.columns),
	}
	return config
}

func (b *runConfigBuilder) buildCircularDependencyConfigs() []*RunConfig {
	var configs []*RunConfig

	var where *string
	if b.whereClause != nil {
		where = b.whereClause
	}

	orderByColumns := b.getOrderByColumns(b.columns)
	insertConfig := &RunConfig{
		id:             fmt.Sprintf("%s.%s", b.table, RunTypeInsert),
		table:          b.table,
		runType:        RunTypeInsert,
		selectColumns:  b.columns, // select cols in insert config must be all columns due to S3 as possible output
		insertColumns:  b.primaryKeys,
		primaryKeys:    b.primaryKeys,
		whereClause:    where,
		orderByColumns: orderByColumns,
		dependsOn:      []*DependsOn{},
	}

	// Track which columns still need to be inserted (that aren’t handled by constraints).
	remainingColumns := make(map[string]bool, len(b.columns))
	for _, col := range b.columns {
		if slices.Contains(b.primaryKeys, col) {
			continue
		}
		remainingColumns[col] = true
	}

	updateConfigCount := 0
	// build update configs for any nullable foreign keys
	for _, fc := range b.foreignKeys {
		if fc == nil || fc.ForeignKey == nil {
			continue
		}

		insertCols, insertFkCols, updateCols, updateFkCols := []string{}, []string{}, []string{}, []string{}

		// Classify each constrained column into insert vs. update groups
		// based on whether the column is NOT NULL.
		for i, col := range fc.Columns {
			// Mark this column as handled in constraints (so we don’t insert it again later).
			remainingColumns[col] = false

			if fc.NotNullable[i] {
				insertCols = append(insertCols, col)
				insertFkCols = append(insertFkCols, fc.ForeignKey.Columns[i])
			} else {
				updateCols = append(updateCols, col)
				updateFkCols = append(updateFkCols, fc.ForeignKey.Columns[i])
			}
		}

		// For NOT NULL constraints, we can safely insert them now (but they depend on the referenced table).
		if len(insertCols) > 0 {
			insertConfig.insertColumns = append(insertConfig.insertColumns, insertCols...)
			insertConfig.dependsOn = append(insertConfig.dependsOn, &DependsOn{
				Table:   fc.ForeignKey.Table,
				Columns: insertFkCols,
			})
		}

		// For columns that can be null, we do them after the main insert (Update).
		if len(updateCols) > 0 {
			updateConfigCount++
			updateConfig := b.buildUpdateConfig(fc, updateCols, updateFkCols, where, orderByColumns, updateConfigCount)
			configs = append(configs, updateConfig)
		}
	}

	// Handle any columns that do not appear in any constraints.
	for col, stillNeeded := range remainingColumns {
		if stillNeeded {
			insertConfig.insertColumns = append(insertConfig.insertColumns, col)
		}
	}

	insertConfig.foreignKeys = b.getForeignKeys(insertConfig.insertColumns)

	// Insert config should be at the front, then any update configs follow.
	configs = append([]*RunConfig{insertConfig}, configs...)
	return configs
}

func (b *runConfigBuilder) buildUpdateConfig(
	fc *sqlmanager_shared.ForeignConstraint,
	updateCols []string,
	updateFkCols []string,
	where *string,
	orderByColumns []string,
	count int,
) *RunConfig {
	dependsOn := []*DependsOn{
		{
			Table:   fc.ForeignKey.Table,
			Columns: updateFkCols,
		},
	}

	// if the foreign key table is not the same as the table, we need to add a depends on for the primary keys
	if fc.ForeignKey.Table != b.table {
		dependsOn = append(dependsOn, &DependsOn{
			Table:   b.table,
			Columns: b.primaryKeys,
		})
	}

	selectColumns := slices.Concat(b.primaryKeys, updateCols)
	return &RunConfig{
		id:             fmt.Sprintf("%s.%s.%d", b.table, RunTypeUpdate, count),
		table:          b.table,
		runType:        RunTypeUpdate,
		selectColumns:  selectColumns,
		insertColumns:  updateCols,
		primaryKeys:    b.primaryKeys,
		whereClause:    where,
		orderByColumns: orderByColumns,
		dependsOn:      dependsOn,
		foreignKeys:    b.getForeignKeys(updateCols),
	}
}

func (b *runConfigBuilder) getDependsOn() []*DependsOn {
	dependsOn := []*DependsOn{}
	for _, fk := range b.foreignKeys {
		dependsOn = append(dependsOn, &DependsOn{
			Table:   fk.ForeignKey.Table,
			Columns: fk.ForeignKey.Columns,
		})
	}
	return dependsOn
}

// getForeignKeys returns foreign keys that are needed for the insert columns
func (b *runConfigBuilder) getForeignKeys(insertColumns []string) []*ForeignKey {
	fmt.Println(insertColumns)
	foreignKeys := []*ForeignKey{}
	for _, fk := range b.foreignKeys {
		foreignKey := &ForeignKey{
			ReferenceTable: fk.ForeignKey.Table,
		}
		for idx, col := range fk.Columns {
			// by checking insert columns, we can skip foreign keys that are not needed for the insert
			// if slices.Contains(insertColumns, col) {
			foreignKey.Columns = append(foreignKey.Columns, col)
			foreignKey.NotNullable = append(foreignKey.NotNullable, fk.NotNullable[idx])
			foreignKey.ReferenceColumns = append(foreignKey.ReferenceColumns, fk.ForeignKey.Columns[idx])
			// }
		}

		if len(foreignKey.Columns) > 0 {
			foreignKeys = append(foreignKeys, foreignKey)
		}
	}
	return foreignKeys
}

// getOrderByColumns returns order by columns for a table, prioritizing primary keys,
// then unique indexes, and finally falling back to sorted select columns.
func (b *runConfigBuilder) getOrderByColumns(selectColumns []string) []string {
	if len(b.primaryKeys) > 0 {
		return b.primaryKeys
	}

	if len(b.uniqueConstraints) > 0 {
		return b.uniqueConstraints[0]
	}

	if len(b.uniqueIndexes) > 0 {
		return b.uniqueIndexes[0]
	}

	slices.Sort(selectColumns)
	return selectColumns
}
