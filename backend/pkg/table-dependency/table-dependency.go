package tabledependency

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/backend/pkg/utils"
)

type RunType string

const (
	RunTypeUpdate RunType = "update"
	RunTypeInsert RunType = "insert"
)

type TableColumn struct {
	Schema  string
	Table   string
	Columns []string
}

type DependsOn struct {
	Table   string
	Columns []string
}

type RunConfig struct {
	Table         string // schema.table
	SelectColumns []string
	InsertColumns []string
	DependsOn     []*DependsOn
	RunType       RunType
	PrimaryKeys   []string
	WhereClause   *string
	SelectQuery   *string
	// Used for mutations to handle columns that contain dot notations (for nosql)
	SplitColumnPaths bool
}

type ConstraintColumns struct {
	NullableColumns    []string
	NonNullableColumns []string
}

func GetRunConfigs(
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
	subsets map[string]string,
	primaryKeyMap map[string][]string,
	tableColumnsMap map[string][]string,
) ([]*RunConfig, error) {
	filteredDepsMap := map[string][]string{}                        // only include tables that are in tables arg list
	foreignKeyMap := map[string]map[string][]string{}               // map: table -> foreign key table -> foreign key column
	foreignKeyColsMap := map[string]map[string]*ConstraintColumns{} // map: table -> foreign key table -> ConstraintColumns
	configs := []*RunConfig{}

	jsonF, _ := json.MarshalIndent(subsets, "", " ")
	fmt.Printf("subsets: %s \n", string(jsonF))

	// dedupe table columns
	for table, cols := range tableColumnsMap {
		tableColumnsMap[table] = utils.DedupeSliceOrdered(cols)
	}

	for table, constraints := range dependencyMap {
		foreignKeyMap[table] = map[string][]string{}
		foreignKeyColsMap[table] = map[string]*ConstraintColumns{}
		for _, constraint := range constraints {
			for idx, col := range constraint.ForeignKey.Columns {
				if !checkTableHasCols([]string{table, constraint.ForeignKey.Table}, tableColumnsMap) {
					continue
				}
				if _, exists := foreignKeyColsMap[table][constraint.ForeignKey.Table]; !exists {
					foreignKeyColsMap[table][constraint.ForeignKey.Table] = &ConstraintColumns{
						NullableColumns:    []string{},
						NonNullableColumns: []string{},
					}
				}
				notNullable := constraint.NotNullable[idx]
				if notNullable {
					foreignKeyColsMap[table][constraint.ForeignKey.Table].NonNullableColumns = append(foreignKeyColsMap[table][constraint.ForeignKey.Table].NonNullableColumns, col)
				} else {
					foreignKeyColsMap[table][constraint.ForeignKey.Table].NullableColumns = append(foreignKeyColsMap[table][constraint.ForeignKey.Table].NullableColumns, col)
				}
				foreignKeyMap[table][constraint.ForeignKey.Table] = append(foreignKeyMap[table][constraint.ForeignKey.Table], col)
				filteredDepsMap[table] = append(filteredDepsMap[table], constraint.ForeignKey.Table)
			}
		}
	}

	for table, deps := range filteredDepsMap {
		filteredDepsMap[table] = utils.DedupeSliceOrdered(deps)
	}

	// create map containing all tables to track when each is processed
	processed := make(map[string]bool, len(tableColumnsMap))
	for t := range tableColumnsMap {
		processed[t] = false
	}

	// create configs for tables in circular dependencies
	circularDeps := FindCircularDependencies(filteredDepsMap)
	groupedCycles := groupDependencies(circularDeps)
	for _, group := range groupedCycles {
		if len(group) == 0 {
			continue
		}
		cycleConfigs, err := processCycles(group, tableColumnsMap, primaryKeyMap, subsets, dependencyMap, foreignKeyColsMap)
		if err != nil {
			return nil, err
		}
		// update table processed map
		for _, cfg := range cycleConfigs {
			processed[cfg.Table] = true
		}
		configs = append(configs, cycleConfigs...)
	}

	insertConfigs := processTables(processed, filteredDepsMap, foreignKeyMap, tableColumnsMap, primaryKeyMap, subsets)
	configs = append(configs, insertConfigs...)

	// filter configs by subset
	if len(subsets) > 0 {
		fmt.Println()
		fmt.Println("filtering configs by subset")
		fmt.Println()
		configs = filterConfigsWithWhereClause(configs)
	}

	// check run path
	if !isValidRunOrder(configs) {
		return nil, errors.New("unable to build table run order. unsupported circular dependency detected.")
	}

	return configs, nil
}

// removes update configs that have where clause
// breaks circular dependencies and self references when subset is applied
func filterConfigsWithWhereClause(configs []*RunConfig) []*RunConfig {
	result := make([]*RunConfig, 0)
	visited := make(map[string]bool)
	hasWhereClause := make(map[string]bool)

	var checkConfig func(*RunConfig) bool
	checkConfig = func(config *RunConfig) bool {
		if hasWhereClause[config.Table] {
			return true
		}

		key := fmt.Sprintf("%s.%s", config.Table, config.RunType)
		if visited[key] {
			return false
		}
		visited[key] = true

		if config.WhereClause != nil {
			hasWhereClause[key] = true
			return true
		}

		for _, dep := range config.DependsOn {
			for _, c := range configs {
				if c.Table == dep.Table {
					if checkConfig(c) {
						hasWhereClause[key] = true
						return true
					}
					break
				}
			}
		}

		return false
	}

	for _, config := range configs {
		if checkConfig(config) {
			if config.RunType == RunTypeInsert {
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
	// determine start table
	startTables, err := DetermineCycleStarts(cycles, subsets, dependencyMap)
	if err != nil {
		return nil, err
	}

	if len(startTables) == 0 {
		return nil, fmt.Errorf("unable to determine start of multi circular dependency: %+v", cycles)
	}

	for _, startTable := range startTables {
		if processed[startTable] {
			continue
		}
		// create insert and update configs for each start table
		cols, colsOk := tableColumnsMap[startTable]
		if !colsOk {
			return nil, fmt.Errorf("missing column mappings for table: %s", startTable)
		}
		pks := primaryKeyMap[startTable]
		where := subsets[startTable]
		dependencies, dependencyOk := dependencyMap[startTable]
		if !dependencyOk {
			return nil, fmt.Errorf("missing dependencies for table: %s", startTable)
		}

		insertConfig := &RunConfig{
			Table:         startTable,
			DependsOn:     []*DependsOn{},
			RunType:       RunTypeInsert,
			SelectColumns: []string{},
			InsertColumns: []string{},
			PrimaryKeys:   pks,
			WhereClause:   &where,
		}

		updateConfig := &RunConfig{
			Table:         startTable,
			DependsOn:     []*DependsOn{{Table: startTable, Columns: pks}}, // add insert config as dependency to update config
			RunType:       RunTypeUpdate,
			SelectColumns: []string{},
			InsertColumns: []string{},
			PrimaryKeys:   pks,
			WhereClause:   &where,
		}
		updateConfig.SelectColumns = append(updateConfig.SelectColumns, pks...)
		deps := foreignKeyColsMap[startTable]
		// builds depends on slice
		for fkTable, fkCols := range deps {
			if fkTable == startTable {
				continue
			}
			if isTableInCycles(cycles, fkTable) {
				if len(fkCols.NullableColumns) > 0 {
					updateConfig.DependsOn = append(updateConfig.DependsOn, &DependsOn{Table: fkTable, Columns: fkCols.NullableColumns})
				} else {
					insertConfig.DependsOn = append(insertConfig.DependsOn, &DependsOn{Table: fkTable, Columns: fkCols.NonNullableColumns})
				}
			} else {
				insertCols := fkCols.NonNullableColumns
				insertCols = append(insertCols, fkCols.NullableColumns...)
				insertConfig.DependsOn = append(insertConfig.DependsOn, &DependsOn{Table: fkTable, Columns: insertCols})
			}
		}
		// builds select + insert columns slices
		for _, d := range dependencies {
			if isTableInCycles(cycles, d.ForeignKey.Table) {
				for idx, col := range d.Columns {
					if !d.NotNullable[idx] {
						updateConfig.SelectColumns = append(updateConfig.SelectColumns, col)
						updateConfig.InsertColumns = append(updateConfig.InsertColumns, col)
					}
				}
			}
		}
		for _, col := range cols {
			if !slices.Contains(updateConfig.InsertColumns, col) {
				insertConfig.InsertColumns = append(insertConfig.InsertColumns, col)
			}
			// select cols in insert config must be all columns due to S3 as possible output
			insertConfig.SelectColumns = append(insertConfig.SelectColumns, col)
		}
		processed[startTable] = true
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
		config := &RunConfig{
			Table:         table,
			DependsOn:     []*DependsOn{},
			RunType:       RunTypeInsert,
			SelectColumns: cols,
			InsertColumns: cols,
			PrimaryKeys:   pks,
			WhereClause:   &where,
		}
		deps := foreignKeyColsMap[table]
		for fkTable, fkCols := range deps {
			config.DependsOn = append(config.DependsOn, &DependsOn{Table: fkTable, Columns: slices.Concat(fkCols.NullableColumns, fkCols.NonNullableColumns)})
		}
		configs = append(configs, config)
	}
	return configs, err
}

func isTableInCycles(cycles [][]string, table string) bool {
	for _, cycle := range cycles {
		for _, t := range cycle {
			if table == t {
				return true
			}
		}
	}
	return false
}

func DetermineCycleStarts(
	cycles [][]string,
	subsets map[string]string,
	dependencyMap map[string][]*sqlmanager_shared.ForeignConstraint,
) ([]string, error) {
	tableRankMap := map[string]int{}
	possibleStarts := [][]string{}

	// FK columns must be nullable to be a starting point
	// filters out tables where foreign keys are not nullable
	for _, cycle := range cycles {
		filteredCycle := []string{}
		for _, table := range cycle {
			dependencies, ok := dependencyMap[table]
			if !ok {
				return nil, fmt.Errorf("missing dependencies for table: %s", table)
			}
			// FK columns must be nullable to be a starting point
			if areAllFkColsNullable(dependencies, cycle) {
				filteredCycle = append(filteredCycle, table)
			}
		}
		possibleStarts = append(possibleStarts, filteredCycle)
	}

	// rank each table
	for _, cycle := range possibleStarts {
		for _, table := range cycle {
			rank := 1
			currRank, seen := tableRankMap[table]
			if seen {
				// intersect table
				rank++
			}
			_, hasSubset := subsets[table]
			if hasSubset {
				rank += 2
			}
			tableRankMap[table] = rank + currRank
		}
	}

	startingTables := map[string]struct{}{}
	// for each cycle choose highest rank
	for _, cycle := range possibleStarts {
		var start *string
		rank := 0
		for _, table := range cycle {
			table := table
			tableRank := tableRankMap[table]
			if tableRank > rank {
				start = &table
				rank = tableRank
			}
		}
		if start != nil && *start != "" {
			startingTables[*start] = struct{}{}
		}
	}
	results := []string{}
	for t := range startingTables {
		results = append(results, t)
	}
	return results, nil
}

func checkTableHasCols(tables []string, tablesColMap map[string][]string) bool {
	for _, t := range tables {
		if _, ok := tablesColMap[t]; !ok {
			return false
		}
	}
	return true
}

func areAllFkColsNullable(dependencies []*sqlmanager_shared.ForeignConstraint, cycle []string) bool {
	for _, dep := range dependencies {
		if !slices.Contains(cycle, dep.ForeignKey.Table) {
			continue
		}
		for _, notNullable := range dep.NotNullable {
			if notNullable {
				return false
			}
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
		config := &RunConfig{
			Table:         table,
			DependsOn:     []*DependsOn{},
			RunType:       RunTypeInsert,
			InsertColumns: cols,
			SelectColumns: cols,
			PrimaryKeys:   pks,
			WhereClause:   &where,
		}
		for _, dep := range dependencyMap[table] {
			config.DependsOn = append(config.DependsOn, &DependsOn{Table: dep, Columns: foreignKeyMap[table][dep]})
		}
		configs = append(configs, config)
	}
	return configs
}

// returns all cycles table is in
func getTableCirularDependencies(table string, circularDeps [][]string) [][]string {
	cycles := [][]string{}
	for _, cycle := range circularDeps {
		if slices.Contains(cycle, table) {
			cycles = append(cycles, cycle)
		}
	}
	return cycles
}

func FindCircularDependencies(dependencies map[string][]string) [][]string {
	var result [][]string

	for node := range dependencies {
		visited, recStack := make(map[string]bool), make(map[string]bool)
		dfsCycles(node, node, dependencies, visited, recStack, []string{}, &result)
	}
	return uniqueCycles(result)
}

// finds all possible path variations
func dfsCycles(start, current string, dependencies map[string][]string, visited, recStack map[string]bool, path []string, result *[][]string) {
	if recStack[current] {
		if current == start {
			// make copy to prevent reference issues
			cycle := make([]string, len(path))
			copy(cycle, path)
			*result = append(*result, cycle)
		}
		return
	}

	recStack[current] = true
	path = append(path, current)

	for _, neighbor := range dependencies[current] {
		if !visited[neighbor] {
			dfsCycles(start, neighbor, dependencies, visited, recStack, path, result)
		}
	}

	recStack[current] = false
	if start == current {
		visited[current] = true
	}
}

func uniqueCycles(cycles [][]string) [][]string {
	seen := map[string]bool{}
	var unique [][]string

	for _, cycle := range cycles {
		key := buildCycleKey(cycle)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, cycle)
		}
	}

	return unique
}

func buildCycleKey(cycle []string) string {
	order := cycleOrder(cycle)
	return strings.Join(order, ",")
}

func cycleOrder(cycle []string) []string {
	sortedCycle := make([]string, len(cycle))
	copy(sortedCycle, cycle)
	sort.Strings(sortedCycle)
	return sortedCycle
}

func getMultiTableCircularDependencies(dependencyMap map[string][]string) [][]string {
	cycles := FindCircularDependencies(dependencyMap)
	multiTableCycles := [][]string{}
	for _, c := range cycles {
		if len(c) > 1 {
			multiTableCycles = append(multiTableCycles, c)
		}
	}
	return multiTableCycles
}

type OrderedTablesResult struct {
	OrderedTables []string
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
	orderedTables := []string{}
	seenTables := map[string]struct{}{}
	for table := range tableMap {
		dep, ok := dependencyMap[table]
		if !ok || len(dep) == 0 {
			orderedTables = append(orderedTables, table)
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
				orderedTables = append(orderedTables, table)
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

/*
Example
input := [][]string{
	{"a", "b", "c"},
	{"f", "d", "g"},
	{"b", "e", "d"},
	{"m", "i", "l"},
	{"x", "y", "z"},
}
output := [][][]string{
	{{"a", "b", "c"}, {"f", "d", "g"}, {"b", "e", "d"}},
  {{"m", "i", "l"}},
	{{"x", "y", "z"}},
}
*/
// union all
func groupDependencies(dependencies [][]string) [][][]string {
	parent := make(map[string]string)
	rank := make(map[string]int)

	// init union-find structure
	for _, group := range dependencies {
		for _, item := range group {
			if _, ok := parent[item]; !ok {
				parent[item] = item
				rank[item] = 0
			}
		}
	}

	// find root
	var find func(x string) string
	find = func(x string) string {
		if parent[x] != x {
			parent[x] = find(parent[x]) // path compression
		}
		return parent[x]
	}

	// union two sets
	union := func(x, y string) {
		rootX := find(x)
		rootY := find(y)
		if rootX != rootY {
			if rank[rootX] > rank[rootY] {
				parent[rootY] = rootX
			} else if rank[rootX] < rank[rootY] {
				parent[rootX] = rootY
			} else {
				parent[rootY] = rootX
				rank[rootX]++
			}
		}
	}

	// union all
	for _, group := range dependencies {
		base := group[0]
		for _, item := range group[1:] {
			union(base, item)
		}
	}

	// group by root
	groupsMap := make(map[string][]string)
	for item := range parent {
		root := find(item)
		groupsMap[root] = append(groupsMap[root], item)
	}

	groupLists := make(map[string][][]string)
	for root := range groupsMap {
		groupLists[root] = [][]string{}
	}

	for _, dependency := range dependencies {
		root := find(dependency[0])
		groupLists[root] = append(groupLists[root], dependency)
	}

	result := [][][]string{}
	for _, groups := range groupLists {
		result = append(result, groups)
	}

	return result
}

func isValidRunOrder(configs []*RunConfig) bool {
	seenTables := map[string][]string{}

	configMap := map[string]*RunConfig{}
	for _, config := range configs {
		configName := fmt.Sprintf("%s.%s", config.Table, config.RunType)
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
			if len(config.DependsOn) == 0 {
				seenTables[config.Table] = config.InsertColumns
				delete(configMap, name)
				continue
			}
			// child table
			for _, d := range config.DependsOn {
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
					seenTables[config.Table] = append(seenTables[config.Table], config.InsertColumns...)
					delete(configMap, name)
				}
			}
		}
	}
	return true
}
