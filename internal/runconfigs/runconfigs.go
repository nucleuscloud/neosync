package runconfigs

import (
	"errors"
	"fmt"
	"slices"
	"strings"

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
	orderByColumns   []string // columns to order by
	selectQuery      *string
	splitColumnPaths bool
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

func (rc *RunConfig) SetSelectQuery(query *string) {
	rc.selectQuery = query
}

func (rc *RunConfig) String() string {
	var sb strings.Builder

	sb.WriteString("RunConfig:\n")
	if rc == nil {
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("  Table: %s\n", rc.table))
	sb.WriteString(fmt.Sprintf("  RunType: %s\n", rc.runType))
	sb.WriteString(fmt.Sprintf("  PrimaryKeys: %v\n", rc.primaryKeys))

	if rc.whereClause != nil {
		sb.WriteString(fmt.Sprintf("  WhereClause: %s\n", *rc.whereClause))
	} else {
		sb.WriteString("  WhereClause: nil\n")
	}

	if rc.selectQuery != nil {
		sb.WriteString(fmt.Sprintf("  SelectQuery: %s\n", *rc.selectQuery))
	} else {
		sb.WriteString("  SelectQuery: nil\n")
	}

	sb.WriteString(fmt.Sprintf("  SelectColumns: %v\n", rc.selectColumns))
	sb.WriteString(fmt.Sprintf("  InsertColumns: %v\n", rc.insertColumns))
	sb.WriteString(fmt.Sprintf("  OrderByColumns: %v\n", rc.orderByColumns))
	sb.WriteString(fmt.Sprintf("  SplitColumnPaths: %v\n", rc.splitColumnPaths))

	sb.WriteString("  DependsOn:\n")
	if rc.dependsOn != nil {
		for i, d := range rc.dependsOn {
			sb.WriteString(fmt.Sprintf("    [%d] Table: %s, Columns: %v\n", i, d.Table, d.Columns))
		}
	} else {
		sb.WriteString("    nil\n")
	}

	sb.WriteString("  ForeignKeys:\n")
	if rc.foreignKeys != nil {
		for i, fk := range rc.foreignKeys {
			sb.WriteString(fmt.Sprintf("    [%d] Columns: %v, NotNullable: %v, ReferenceTable: %s, ReferenceColumns: %v\n",
				i, fk.Columns, fk.NotNullable, fk.ReferenceTable, fk.ReferenceColumns))
		}
	} else {
		sb.WriteString("    nil\n")
	}

	return sb.String()
}

func BuildRunConfigs(
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
	subsets map[string]string,
	primaryKeyMap map[string][]string,
	tableColumnsMap map[string][]string,
	uniqueIndexesMap map[string][][]string,
	uniqueConstraintsMap map[string][][]string,
) ([]*RunConfig, error) {
	configs := []*RunConfig{}

	// dedupe table columns
	for table, cols := range tableColumnsMap {
		tableColumnsMap[table] = utils.DedupeSliceOrdered(cols)
	}

	// filter dependencies to only include tables are in tableColumnsMap (jobmappings)
	filteredFks := filterDependencies(dependencyMap, tableColumnsMap)

	// find circular dependencies
	graph := buildDependencyGraph(filteredFks)
	circularDeps := FindCircularDependencies(graph)
	circularTables := circularDependencyTables(circularDeps)

	// build configs for each table
	for table, columns := range tableColumnsMap {
		var where *string
		if subset, ok := subsets[table]; ok {
			where = &subset
		}
		builder := newRunConfigBuilder(table, columns, primaryKeyMap[table], where, uniqueIndexesMap[table], uniqueConstraintsMap[table], filteredFks[table], circularTables[table])
		cfgs := builder.Build()
		configs = append(configs, cfgs...)
	}

	// check run path
	if !isValidRunOrder(configs) {
		return nil, errors.New("unable to build table run order. unsupported circular dependency detected.")
	}

	return configs, nil
}

func circularDependencyTables(circularDeps [][]string) map[string]bool {
	circularTables := make(map[string]bool)
	for _, cycle := range circularDeps {
		for _, table := range cycle {
			circularTables[table] = true
		}
	}
	return circularTables
}

// filter dependencies to only include tables are in tableColumnsMap (jobmappings)
func filterDependencies(
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
	tableColumnsMap map[string][]string,
) map[string][]*sqlmanager_shared.ForeignConstraint {
	filtered := make(map[string][]*sqlmanager_shared.ForeignConstraint)

	for table, constraints := range dependencyMap {
		for _, constraint := range constraints {
			fkTable := constraint.ForeignKey.Table
			if checkTableHasCols([]string{table, fkTable}, tableColumnsMap) {
				filtered[table] = append(filtered[table], constraint)
			}
		}
	}

	return filtered
}

func checkTableHasCols(tables []string, tablesColMap map[string][]string) bool {
	for _, t := range tables {
		if _, ok := tablesColMap[t]; !ok {
			return false
		}
	}
	return true
}

func isValidRunOrder(configs []*RunConfig) bool {
	seenTables := map[string][]string{}

	configMap := map[string]*RunConfig{}
	for _, config := range configs {
		configName := fmt.Sprintf("%s.%s.%s", config.Table(), config.RunType(), strings.Join(config.InsertColumns(), ","))
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
