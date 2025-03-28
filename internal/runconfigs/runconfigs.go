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

type DependsOn struct {
	Table   string
	Columns []string
}

type ForeignKey struct {
	Columns          []string
	NotNullable      []bool
	ReferenceSchema  string
	ReferenceTable   string
	ReferenceColumns []string
}

// SubsetPath represents the shortest path from a table to a where-clause root.
// The Path slice is ordered from the table up to the root (i.e. [child, ..., root]).
type SubsetPath struct {
	Subset    string      // subset query
	Root      string      // table key of the root table with a where clause
	JoinSteps []*JoinStep // join steps from the current table to the subset root table
}

// joinStep represents one “edge” (join) in a join chain from one table to a related table.
type JoinStep struct {
	FromKey    string      // table key of the child table
	ToKey      string      // table key of the parent table
	ForeignKey *ForeignKey // foreign key relationship between the parent and child tables
}

type RunConfig struct {
	id               string                        // unique identifier for the run config
	table            sqlmanager_shared.SchemaTable // table to run the query on
	selectColumns    []string                      // columns to select
	insertColumns    []string                      // columns to insert
	dependsOn        []*DependsOn                  // tables that must be run before this one
	runType          RunType                       // type of run (update or insert)
	primaryKeys      []string                      // primary keys for the table
	whereClause      *string                       // subset query
	orderByColumns   []string                      // columns to order by
	splitColumnPaths bool                          // whether to split column paths
	subsetPaths      []*SubsetPath                 // holds one (or more) shortest paths from this table to any table that has a where clause.
}

func NewRunConfig(
	id string,
	table sqlmanager_shared.SchemaTable,
	runtype RunType,
	primaryKeys []string,
	whereClause *string,
	selectCols, insertCols []string,
	dependsOn []*DependsOn,
	splitColumnPaths bool,
) *RunConfig {
	return &RunConfig{
		id:               id,
		table:            table,
		runType:          runtype,
		primaryKeys:      primaryKeys,
		whereClause:      whereClause,
		insertColumns:    insertCols,
		selectColumns:    selectCols,
		dependsOn:        dependsOn,
		splitColumnPaths: splitColumnPaths,
	}
}

func (rc *RunConfig) Id() string {
	return rc.id
}

func (rc *RunConfig) Table() string {
	return rc.table.String()
}

func (rc *RunConfig) SchemaTable() sqlmanager_shared.SchemaTable {
	return rc.table
}

func (rc *RunConfig) SelectColumns() []string {
	result := make([]string, len(rc.selectColumns))
	copy(result, rc.selectColumns)
	return result
}

func (rc *RunConfig) InsertColumns() []string {
	result := make([]string, len(rc.insertColumns))
	copy(result, rc.insertColumns)
	return result
}

func (rc *RunConfig) DependsOn() []*DependsOn {
	return rc.dependsOn
}

func (rc *RunConfig) RunType() RunType {
	return rc.runType
}

func (rc *RunConfig) PrimaryKeys() []string {
	result := make([]string, len(rc.primaryKeys))
	copy(result, rc.primaryKeys)
	return result
}

func (rc *RunConfig) WhereClause() *string {
	if rc.whereClause == nil {
		return nil
	}
	copied := *rc.whereClause
	return &copied
}

func (rc *RunConfig) SubsetPaths() []*SubsetPath {
	if rc.subsetPaths == nil {
		return []*SubsetPath{}
	}
	return rc.subsetPaths
}

func (rc *RunConfig) OrderByColumns() []string {
	result := make([]string, len(rc.orderByColumns))
	copy(result, rc.orderByColumns)
	return result
}

func (rc *RunConfig) SplitColumnPaths() bool {
	return rc.splitColumnPaths
}

func (rc *RunConfig) String() string {
	var sb strings.Builder

	sb.WriteString("RunConfig:\n")
	if rc == nil {
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("  Id: %s\n", rc.id))
	sb.WriteString(fmt.Sprintf("  Table: %s\n", rc.table))
	sb.WriteString(fmt.Sprintf("  RunType: %s\n", rc.runType))
	sb.WriteString(fmt.Sprintf("  PrimaryKeys: %v\n", rc.primaryKeys))

	if rc.whereClause != nil {
		sb.WriteString(fmt.Sprintf("  WhereClause: %s\n", *rc.whereClause))
	} else {
		sb.WriteString("  WhereClause: nil\n")
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

	sb.WriteString("  SubsetPaths:\n")
	if rc.subsetPaths != nil {
		for i, sp := range rc.subsetPaths {
			sb.WriteString(fmt.Sprintf("    [%d] Root: %s, Subset: %s\n", i, sp.Root, sp.Subset))
			sb.WriteString("    JoinSteps:\n")
			for j, js := range sp.JoinSteps {
				sb.WriteString(
					fmt.Sprintf("      [%d] FromKey: %s, ToKey: %s\n", j, js.FromKey, js.ToKey),
				)
				if js.ForeignKey != nil {
					sb.WriteString(
						fmt.Sprintf(
							"        FK: Columns: %v, NotNullable: %v, ReferenceSchema: %s, ReferenceTable: %s, ReferenceColumns: %v\n",
							js.ForeignKey.Columns,
							js.ForeignKey.NotNullable,
							js.ForeignKey.ReferenceSchema,
							js.ForeignKey.ReferenceTable,
							js.ForeignKey.ReferenceColumns,
						),
					)
				}
			}
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

	// filter dependencies to only include tables in tableColumnsMap (jobmappings)
	filteredFks := filterDependencies(dependencyMap, tableColumnsMap)

	tableConfigsBuilder := newTableConfigsBuilder(
		tableColumnsMap,
		primaryKeyMap,
		subsets,
		uniqueIndexesMap,
		uniqueConstraintsMap,
		filteredFks,
	)

	// build configs for each table
	for schematable := range tableColumnsMap {
		schema, table := sqlmanager_shared.SplitTableKey(schematable)
		schematable := sqlmanager_shared.SchemaTable{
			Schema: schema,
			Table:  table,
		}
		cfgs := tableConfigsBuilder.Build(schematable)
		configs = append(configs, cfgs...)
	}

	// check run path
	if !isValidRunOrder(configs) {
		return nil, errors.New(
			"unsupported circular dependency detected. at least one foreign key in circular dependency must be nullable",
		)
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
		if _, exists := configMap[config.Id()]; exists {
			// configs should be unique
			return false
		}
		configMap[config.Id()] = config
	}

	prevTableLen := 0
	for len(configMap) > 0 {
		// prevents looping forever
		if prevTableLen == len(configMap) {
			return false
		}
		prevTableLen = len(configMap)
		for id, config := range configMap {
			if AreConfigDependenciesSatisfied(config.DependsOn(), seenTables) {
				seenTables[config.Table()] = append(
					seenTables[config.Table()],
					config.InsertColumns()...)
				delete(configMap, id)
			}
		}
	}
	return true
}

// AreConfigDependenciesSatisfied checks if all dependencies for a given config have been completed
// completed is a map of table name to a list of completed columns
func AreConfigDependenciesSatisfied(dependsOn []*DependsOn, completed map[string][]string) bool {
	// root table
	if len(dependsOn) == 0 {
		return true
	}
	// check that all columns in dependency has been completed
	for _, dep := range dependsOn {
		completedCols, ok := completed[dep.Table]
		if ok {
			for _, dc := range dep.Columns {
				if !slices.Contains(completedCols, dc) {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}
