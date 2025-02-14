package tabledependency

import (
	"errors"
	"fmt"
	"slices"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/backend/pkg/utils"
)

type RunType string

const (
	RunTypeUpdate RunType = "update"
	RunTypeInsert RunType = "insert"
)

type ConstraintColumns struct {
	NullableColumns    []string
	NonNullableColumns []string
}

type TableColumn struct {
	Schema  string
	Table   string
	Columns []string
}

type DependsOn struct {
	Table   string
	Columns []string
}

type ForeignKey struct {
	Columns     []string
	NotNullable []bool
	// ReferenceSchema  string  TODO: need to split out schema and table
	ReferenceTable   string
	ReferenceColumns []string
}

type RunConfig struct {
	table            string // schema.table  TODO: should use sqlmanager_shared.SchemaTable
	selectColumns    []string
	insertColumns    []string
	dependsOn        []*DependsOn // this should be a list of config names like "table.insert", rename to dependsOnConfigs
	foreignKeys      []*ForeignKey
	runType          RunType
	primaryKeys      []string
	whereClause      *string
	orderBy          []string
	selectQuery      *string
	splitColumnPaths bool
}

func newRunConfig(
	table string,
	runtype RunType,
	primaryKeys []string,
	whereClause *string,
) *RunConfig {
	return &RunConfig{
		table:       table,
		runType:     runtype,
		primaryKeys: primaryKeys,
		whereClause: whereClause,
	}
}

func NewRunConfig(
	table string,
	runtype RunType,
	primaryKeys []string,
	whereClause *string,
	selectCols, insertCols []string,
	dependsOn []*DependsOn,
	foreignKeys []*ForeignKey,
	splitColumnPaths bool,
) *RunConfig {
	return &RunConfig{
		table:            table,
		runType:          runtype,
		primaryKeys:      primaryKeys,
		whereClause:      whereClause,
		insertColumns:    insertCols,
		selectColumns:    selectCols,
		dependsOn:        dependsOn,
		splitColumnPaths: splitColumnPaths,
		foreignKeys:      foreignKeys,
	}
}

func (rc *RunConfig) Table() string {
	return rc.table
}

func (rc *RunConfig) SelectColumns() []string {
	return rc.selectColumns
}

func (rc *RunConfig) InsertColumns() []string {
	return rc.insertColumns
}

func (rc *RunConfig) DependsOn() []*DependsOn {
	return rc.dependsOn
}

func (rc *RunConfig) RunType() RunType {
	return rc.runType
}

func (rc *RunConfig) PrimaryKeys() []string {
	return rc.primaryKeys
}

func (rc *RunConfig) WhereClause() *string {
	return rc.whereClause
}

func (rc *RunConfig) SelectQuery() *string {
	return rc.selectQuery
}

func (rc *RunConfig) SplitColumnPaths() bool {
	return rc.splitColumnPaths
}

func (rc *RunConfig) ForeignKeys() []*ForeignKey {
	return rc.foreignKeys
}

func (rc *RunConfig) appendSelectColumns(columns ...string) {
	rc.selectColumns = append(rc.selectColumns, columns...)
}

func (rc *RunConfig) appendInsertColumns(columns ...string) {
	rc.insertColumns = append(rc.insertColumns, columns...)
}

func (rc *RunConfig) appendDependsOn(table string, columns []string) {
	rc.dependsOn = append(rc.dependsOn, &DependsOn{
		Table:   table,
		Columns: columns,
	})
}

func (rc *RunConfig) appendForeignKey(fk *ForeignKey) {
	rc.foreignKeys = append(rc.foreignKeys, fk)
}

func (rc *RunConfig) setOrderBy(orderBy []string) {
	rc.orderBy = orderBy
}

func (rc *RunConfig) SetSelectQuery(query *string) {
	rc.selectQuery = query
}

func GetRunConfigs(
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
	subsets map[string]string,
	primaryKeyMap map[string][]string,
	tableColumnsMap map[string][]string,
	uniqueIndexesMap map[string][][]string,
) ([]*RunConfig, error) {
	configs := []*RunConfig{}

	// dedupe table columns
	for table, cols := range tableColumnsMap {
		tableColumnsMap[table] = utils.DedupeSliceOrdered(cols)
	}

	d := buildDependencies(dependencyMap, tableColumnsMap)

	// create map containing all tables to track when each is processed
	processed := make(map[string]bool, len(tableColumnsMap))
	for t := range tableColumnsMap {
		processed[t] = false
	}

	// create configs for tables in circular dependencies
	circularDeps := FindCircularDependencies(d.filteredDeps)
	groupedCycles := groupDependencies(circularDeps)
	for _, group := range groupedCycles {
		if len(group) == 0 {
			continue
		}
		cycleConfigs, err := processCycles(group, tableColumnsMap, primaryKeyMap, subsets, dependencyMap, d.foreignKeyCols)
		if err != nil {
			return nil, fmt.Errorf("unable to process cycles: %w", err)
		}
		// update table processed map
		for _, cfg := range cycleConfigs {
			processed[cfg.Table()] = true
		}
		configs = append(configs, cycleConfigs...)
	}

	insertConfigs := processTables(processed, d.filteredDeps, d.foreignKeys, tableColumnsMap, primaryKeyMap, subsets)
	configs = append(configs, insertConfigs...)

	// filter configs by subset
	if len(subsets) > 0 {
		configs = filterConfigsWithWhereClause(configs)
	}

	// Add foreign keys to configs
	for _, config := range configs {
		fks := dependencyMap[config.Table()]
		for _, fk := range fks {
			foreignKey := &ForeignKey{
				ReferenceTable: fk.ForeignKey.Table,
			}
			for idx, col := range fk.Columns {
				// by checking insert columns, we can skip foreign keys that are not needed for the insert
				if slices.Contains(config.insertColumns, col) {
					foreignKey.Columns = append(foreignKey.Columns, col)
					foreignKey.NotNullable = append(foreignKey.NotNullable, fk.NotNullable[idx])
					foreignKey.ReferenceColumns = append(foreignKey.ReferenceColumns, fk.ForeignKey.Columns[idx])
				}
			}

			if len(foreignKey.Columns) > 0 {
				config.appendForeignKey(foreignKey)
			}
		}
	}

	setOrderBy(configs, primaryKeyMap, uniqueIndexesMap)

	// check run path
	if !isValidRunOrder(configs) {
		return nil, errors.New("unable to build table run order. unsupported circular dependency detected.")
	}

	return configs, nil
}

func setOrderBy(configs []*RunConfig, primaryKeyMap map[string][]string, uniqueIndexes map[string][][]string) {
	for _, config := range configs {
		cols := getOrderByColumns(config, primaryKeyMap, uniqueIndexes)
		config.setOrderBy(cols)
	}
}

// getOrderByColumns returns order by columns for a table, prioritizing primary keys,
// then unique indexes, and finally falling back to sorted select columns.
func getOrderByColumns(config *RunConfig, primaryKeyMap map[string][]string, uniqueIndexes map[string][][]string) []string {
	table := config.Table()
	cols, ok := primaryKeyMap[table]
	if ok {
		return cols
	}

	indexes := uniqueIndexes[table]
	if len(indexes) > 0 {
		return indexes[0]
	}

	selectCols := config.SelectColumns()
	slices.Sort(selectCols)
	return selectCols
}

// removes update configs that have where clause
// breaks circular dependencies and self references when subset is applied
func filterConfigsWithWhereClause(configs []*RunConfig) []*RunConfig {
	result := make([]*RunConfig, 0)
	visited := make(map[string]bool)
	hasWhereClause := make(map[string]bool)

	var isSubset func(*RunConfig) bool
	isSubset = func(config *RunConfig) bool {
		if hasWhereClause[config.Table()] {
			return true
		}

		key := fmt.Sprintf("%s.%s", config.Table(), config.RunType())
		if visited[key] {
			return false
		}
		visited[key] = true

		if config.WhereClause() != nil {
			hasWhereClause[config.Table()] = true
			return true
		}

		for _, dep := range config.DependsOn() {
			for _, c := range configs {
				if c.Table() == dep.Table {
					if isSubset(c) {
						hasWhereClause[config.Table()] = true
						return true
					}
					break
				}
			}
		}

		return false
	}

	for _, config := range configs {
		if isSubset(config) {
			if config.RunType() == RunTypeInsert {
				result = append(result, config)
			}
		} else {
			result = append(result, config)
		}
	}

	return result
}

func processCycles(
	cycles [][]string,
	tableColumnsMap map[string][]string,
	primaryKeyMap map[string][]string,
	subsets map[string]string,
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
	foreignKeyColsMap map[string]map[string]*ConstraintColumns,
) ([]*RunConfig, error) {
	configs := []*RunConfig{}
	processed := map[string]bool{}
	for _, cycle := range cycles {
		for _, table := range cycle {
			processed[table] = false
		}
	}

	// determines tables that should be inserted and updated
	insertUpdateTables, err := DetermineCycleInsertUpdateTables(cycles, subsets, dependencyMap)
	if err != nil {
		return nil, err
	}

	if len(insertUpdateTables) == 0 {
		return nil, fmt.Errorf("unable to determine start of multi circular dependency: %+v", cycles)
	}

	for _, table := range insertUpdateTables {
		if processed[table] {
			continue
		}
		// create insert and update configs for each start table
		cols, colsOk := tableColumnsMap[table]
		if !colsOk {
			return nil, fmt.Errorf("missing column mappings for table: %s", table)
		}
		pks := primaryKeyMap[table]
		where := subsets[table]
		dependencies, dependencyOk := dependencyMap[table]
		if !dependencyOk {
			return nil, fmt.Errorf("missing dependencies for table: %s", table)
		}

		insertConfig := newRunConfig(table, RunTypeInsert, pks, &where)
		updateConfig := newRunConfig(table, RunTypeUpdate, pks, &where)
		updateConfig.appendDependsOn(table, pks)
		updateConfig.appendSelectColumns(pks...)
		deps := foreignKeyColsMap[table]
		// builds depends on slice
		for fkTable, fkCols := range deps {
			if fkTable == table {
				continue
			}
			if isTableInCycles(cycles, fkTable) {
				if len(fkCols.NullableColumns) > 0 {
					updateConfig.appendDependsOn(fkTable, fkCols.NullableColumns)
				}
				if len(fkCols.NonNullableColumns) > 0 {
					insertConfig.appendDependsOn(fkTable, fkCols.NonNullableColumns)
				}
			} else {
				insertCols := slices.Concat(fkCols.NonNullableColumns, fkCols.NullableColumns)
				insertConfig.appendDependsOn(fkTable, insertCols)
			}
		}
		// builds select + insert columns slices
		for _, d := range dependencies {
			if isTableInCycles(cycles, d.ForeignKey.Table) {
				for idx, col := range d.Columns {
					if !d.NotNullable[idx] {
						updateConfig.appendSelectColumns(col)
						updateConfig.appendInsertColumns(col)
					}
				}
			}
		}
		for _, col := range cols {
			if !slices.Contains(updateConfig.InsertColumns(), col) {
				insertConfig.appendInsertColumns(col)
			}
			// select cols in insert config must be all columns due to S3 as possible output
			insertConfig.appendSelectColumns(col)
		}

		processed[table] = true
		configs = append(configs, insertConfig, updateConfig)
	}

	// create insert configs for all other tables in cycles
	for table := range processed {
		if processed[table] {
			// skip. already created configs for start tables
			continue
		}
		cols := tableColumnsMap[table]
		pks := primaryKeyMap[table]
		where := subsets[table]
		config := newRunConfig(table, RunTypeInsert, pks, &where)
		config.appendInsertColumns(cols...)
		config.appendSelectColumns(cols...)
		deps := foreignKeyColsMap[table]
		for fkTable, fkCols := range deps {
			config.appendDependsOn(fkTable, slices.Concat(fkCols.NullableColumns, fkCols.NonNullableColumns))
		}

		configs = append(configs, config)
	}
	return configs, err
}

func checkTableHasCols(tables []string, tablesColMap map[string][]string) bool {
	for _, t := range tables {
		if _, ok := tablesColMap[t]; !ok {
			return false
		}
	}
	return true
}

// create insert configs for non-circular dependent tables
func processTables(
	tableMap map[string]bool,
	dependencyMap map[string][]string,
	foreignKeyMap map[string]map[string][]string,
	tableColumnsMap map[string][]string,
	primaryKeyMap map[string][]string,
	subsets map[string]string,
) []*RunConfig {
	configs := []*RunConfig{}
	for table, isProcessed := range tableMap {
		if isProcessed {
			continue
		}
		cols := tableColumnsMap[table]
		pks := primaryKeyMap[table]
		where := subsets[table]
		config := newRunConfig(table, RunTypeInsert, pks, &where)
		config.appendInsertColumns(cols...)
		config.appendSelectColumns(cols...)
		for _, dep := range dependencyMap[table] {
			config.appendDependsOn(dep, foreignKeyMap[table][dep])
		}
		configs = append(configs, config)
	}
	return configs
}

type OrderedTablesResult struct {
	OrderedTables []*sqlmanager_shared.SchemaTable
	HasCycles     bool
}

func GetTablesOrderedByDependency(dependencyMap map[string][]string) (*OrderedTablesResult, error) {
	hasCycles := false
	cycles := getMultiTableCircularDependencies(dependencyMap)
	if len(cycles) > 0 {
		hasCycles = true
	}

	tableMap := map[string]struct{}{}
	for t := range dependencyMap {
		tableMap[t] = struct{}{}
	}
	orderedTables := []*sqlmanager_shared.SchemaTable{}
	seenTables := map[string]struct{}{}
	for table := range tableMap {
		dep, ok := dependencyMap[table]
		if !ok || len(dep) == 0 {
			s, t := sqlmanager_shared.SplitTableKey(table)
			orderedTables = append(orderedTables, &sqlmanager_shared.SchemaTable{Schema: s, Table: t})
			seenTables[table] = struct{}{}
			delete(tableMap, table)
		}
	}

	prevTableLen := 0
	for len(tableMap) > 0 {
		// prevents looping forever
		if prevTableLen == len(tableMap) {
			return nil, fmt.Errorf("unable to build table order")
		}
		prevTableLen = len(tableMap)
		for table := range tableMap {
			deps := dependencyMap[table]
			if isReady(seenTables, deps, table, cycles) {
				s, t := sqlmanager_shared.SplitTableKey(table)
				orderedTables = append(orderedTables, &sqlmanager_shared.SchemaTable{Schema: s, Table: t})
				seenTables[table] = struct{}{}
				delete(tableMap, table)
			}
		}
	}

	return &OrderedTablesResult{OrderedTables: orderedTables, HasCycles: hasCycles}, nil
}

func isReady(seen map[string]struct{}, deps []string, table string, cycles [][]string) bool {
	// allow circular dependencies
	circularDeps := getTableCirularDependencies(table, cycles)
	circularDepsMap := map[string]struct{}{}
	for _, cycle := range circularDeps {
		for _, t := range cycle {
			circularDepsMap[t] = struct{}{}
		}
	}
	for _, d := range deps {
		_, cdOk := circularDepsMap[d]
		if cdOk {
			return true
		}
		_, ok := seen[d]
		// allow self dependencies
		if !ok && d != table {
			return false
		}
	}
	return true
}

func isValidRunOrder(configs []*RunConfig) bool {
	seenTables := map[string][]string{}

	configMap := map[string]*RunConfig{}
	for _, config := range configs {
		configName := fmt.Sprintf("%s.%s", config.Table(), config.RunType())
		if _, exists := configMap[configName]; exists {
			// configs should be unique
			return false
		}
		configMap[configName] = config
	}

	prevTableLen := 0
	for len(configMap) > 0 {
		// prevents looping forever
		if prevTableLen == len(configMap) {
			return false
		}
		prevTableLen = len(configMap)
		for name, config := range configMap {
			// root table
			if len(config.DependsOn()) == 0 {
				seenTables[config.Table()] = config.InsertColumns()
				delete(configMap, name)
				continue
			}
			// child table
			for _, d := range config.DependsOn() {
				seenCols, seen := seenTables[d.Table]
				isReady := func() bool {
					if !seen {
						return false
					}
					for _, c := range d.Columns {
						if !slices.Contains(seenCols, c) {
							return false
						}
					}
					return true
				}
				if isReady() {
					seenTables[config.Table()] = append(seenTables[config.Table()], config.InsertColumns()...)
					delete(configMap, name)
				}
			}
		}
	}
	return true
}
