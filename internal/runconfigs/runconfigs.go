package runconfigs

import (
	"errors"
	"fmt"
	"slices"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
	orderByColumns   []string // columns to order by
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

func (rc *RunConfig) OrderByColumns() []string {
	return rc.orderByColumns
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

func (rc *RunConfig) setOrderByColumns(orderBy []string) {
	rc.orderByColumns = orderBy
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
	uniqueConstraintsMap map[string][][]string,
) ([]*RunConfig, error) {
	configs := []*RunConfig{}

	// // dedupe table columns
	// for table, cols := range tableColumnsMap {
	// 	tableColumnsMap[table] = utils.DedupeSliceOrdered(cols)
	// }

	configs = make([]*RunConfig, 0)
	for table, cols := range tableColumnsMap {
		primaryKeys := primaryKeyMap[table]
		var where *string
		subset, ok := subsets[table]
		if ok {
			where = &subset
		}
		insertConfig := newRunConfig(table, RunTypeInsert, primaryKeys, where)
		insertConfig.appendSelectColumns(cols...) // this should be set select columns
		colsMap := map[string]bool{}
		for _, c := range cols {
			colsMap[c] = true
		}
		dependencies, ok := dependencyMap[table]
		if ok {
			for _, d := range dependencies {
				insertCols := []string{}
				updateCols := []string{}
				for idx, c := range d.Columns {
					notNullable := d.NotNullable[idx]
					if notNullable {
						insertCols = append(insertCols, c)
					} else {
						updateCols = append(updateCols, c)
					}
				}
				insertConfig.appendInsertColumns(insertCols...)
				insertConfig.appendUpdateColumns(updateCols...)
			}
		}

		for c, ok := range colsMap {
			if ok {
				insertConfig.appendInsertColumns(c)
			}
		}

	}

	// filter configs by subset
	if len(subsets) > 0 {
		configs = filterConfigsWithWhereClause(configs)
	}

	setOrderByColumns(configs, primaryKeyMap, uniqueIndexesMap, uniqueConstraintsMap)

	// check run path
	if !isValidRunOrder(configs) {
		return nil, errors.New("unable to build table run order. unsupported circular dependency detected.")
	}

	return configs, nil
}

func isFkNullable(fk *sqlmanager_shared.ForeignConstraint) bool {
	for _, nullable := range fk.NotNullable {
		if !nullable {
			return false
		}
	}
	return true
}

func setOrderByColumns(configs []*RunConfig, primaryKeyMap map[string][]string, uniqueIndexes, uniqueConstraints map[string][][]string) {
	for _, config := range configs {
		cols := getOrderByColumns(config, primaryKeyMap, uniqueIndexes, uniqueConstraints)
		config.setOrderByColumns(cols)
	}
}

// getOrderByColumns returns order by columns for a table, prioritizing primary keys,
// then unique indexes, and finally falling back to sorted select columns.
func getOrderByColumns(config *RunConfig, primaryKeyMap map[string][]string, uniqueIndexes, uniqueConstraints map[string][][]string) []string {
	table := config.Table()
	cols, ok := primaryKeyMap[table]
	if ok {
		return cols
	}

	constraints := uniqueConstraints[table]
	if len(constraints) > 0 {
		return constraints[0]
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
