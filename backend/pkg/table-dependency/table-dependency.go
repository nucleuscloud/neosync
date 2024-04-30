package tabledependency

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
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
	Table       string // schema.table
	Columns     []string
	DependsOn   []*DependsOn
	RunType     RunType
	PrimaryKeys []string
	WhereClause *string
}

type ConstraintColumns struct {
	NullableColumns    []string
	NonNullableColumns []string
}

func GetRunConfigs(
	dependencyMap dbschemas.TableDependency,
	tables []string,
	subsets map[string]string,
	primaryKeyMap map[string][]string,
	tableColumnsMap map[string][]string,
) ([]*RunConfig, error) {
	filteredDepsMap := map[string][]string{}                        // only include tables that are in tables arg list
	foreignKeyMap := map[string]map[string]string{}                 // map: table -> foreign key table -> foreign key column
	foreignKeyColsMap := map[string]map[string]*ConstraintColumns{} // map: table -> foreign key table -> ConstraintColumns
	configs := []*RunConfig{}

	for table, constraints := range dependencyMap {
		foreignKeyMap[table] = map[string]string{}
		foreignKeyColsMap[table] = map[string]*ConstraintColumns{}
		for _, constraint := range constraints.Constraints {
			if _, exists := foreignKeyColsMap[table][constraint.ForeignKey.Table]; !exists {
				foreignKeyColsMap[table][constraint.ForeignKey.Table] = &ConstraintColumns{
					NullableColumns:    []string{},
					NonNullableColumns: []string{},
				}
			}
			if constraint.IsNullable {
				foreignKeyColsMap[table][constraint.ForeignKey.Table].NullableColumns = append(foreignKeyColsMap[table][constraint.ForeignKey.Table].NullableColumns, constraint.ForeignKey.Column)
			} else {
				foreignKeyColsMap[table][constraint.ForeignKey.Table].NonNullableColumns = append(foreignKeyColsMap[table][constraint.ForeignKey.Table].NonNullableColumns, constraint.ForeignKey.Column)
			}
			foreignKeyMap[table][constraint.ForeignKey.Table] = constraint.ForeignKey.Column
			if slices.Contains(tables, table) && slices.Contains(tables, constraint.ForeignKey.Table) {
				filteredDepsMap[table] = append(filteredDepsMap[table], constraint.ForeignKey.Table)
			} else if !constraint.IsNullable {
				return nil, fmt.Errorf("found table constraint that is not nullable. missing table: %s", table)
			}
		}
	}

	// create map containing all tables to track when each is processed
	processed := make(map[string]bool, len(tables))
	for _, t := range tables {
		processed[t] = false
	}

	// create configs for tables in circular dependencies
	circularDeps := findCircularDependencies(filteredDepsMap)
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

	// check run path
	if !isValidRunOrder(configs) {
		return nil, errors.New("unable to build table run order. unsupported circular dependency detected.")
	}

	return configs, nil
}

func processCycles(
	cycles [][]string,
	tableColumnsMap map[string][]string,
	primaryKeyMap map[string][]string,
	subsets map[string]string,
	dependencyMap map[string]*dbschemas.TableConstraints,
	foreignKeyColsMap map[string]map[string]*ConstraintColumns,
) ([]*RunConfig, error) {
	configs := []*RunConfig{}
	// determine start table
	startTables, err := determineCycleStarts(cycles, subsets, dependencyMap)
	if err != nil {
		return nil, err
	}
	if len(startTables) == 0 {
		return nil, fmt.Errorf("unable to determine start of multi circular dependency: %+v", cycles)
	}

	for _, startTable := range startTables {
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
			Table:       startTable,
			DependsOn:   []*DependsOn{},
			RunType:     RunTypeInsert,
			Columns:     []string{},
			PrimaryKeys: pks,
			WhereClause: &where,
		}

		updateConfig := &RunConfig{
			Table:       startTable,
			DependsOn:   []*DependsOn{{Table: startTable, Columns: pks}}, // add insert config as dependency to update config
			RunType:     RunTypeUpdate,
			Columns:     []string{},
			PrimaryKeys: pks,
			WhereClause: &where,
		}
		deps := foreignKeyColsMap[startTable]
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
				insertConfig.DependsOn = append(insertConfig.DependsOn, &DependsOn{Table: fkTable, Columns: fkCols.NonNullableColumns})
			}
		}
		for _, d := range dependencies.Constraints {
			if d.IsNullable {
				updateConfig.Columns = append(updateConfig.Columns, d.Column)
			}
		}
		for _, col := range cols {
			if !slices.Contains(updateConfig.Columns, col) {
				insertConfig.Columns = append(insertConfig.Columns, col)
			}
		}
		configs = append(configs, insertConfig, updateConfig)
	}

	allTables := []string{}
	for _, cycle := range cycles {
		allTables = append(allTables, cycle...)
	}
	// create insert configs for all other tables in cycles
	for _, table := range allTables {
		if slices.Contains(startTables, table) {
			// skip. already created configs for start tables
			continue
		}
		cols := tableColumnsMap[table]
		pks := primaryKeyMap[table]
		where := subsets[table]
		config := &RunConfig{
			Table:       table,
			DependsOn:   []*DependsOn{},
			RunType:     RunTypeInsert,
			Columns:     cols,
			PrimaryKeys: pks,
			WhereClause: &where,
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

func determineCycleStarts(
	cycles [][]string,
	subsets map[string]string,
	dependencyMap dbschemas.TableDependency,
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
			if areAllFkColsNullable(dependencies.Constraints, cycle) {
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

func areAllFkColsNullable(dependencies []*dbschemas.ForeignConstraint, cycle []string) bool {
	for _, dep := range dependencies {
		if !slices.Contains(cycle, dep.ForeignKey.Table) {
			continue
		}
		if !dep.IsNullable {
			return false
		}
	}
	return true
}

// create insert configs for non-circular dependent tables
func processTables(
	tableMap map[string]bool,
	dependencyMap map[string][]string,
	foreignKeyMap map[string]map[string]string,
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
			Table:       table,
			DependsOn:   []*DependsOn{},
			RunType:     RunTypeInsert,
			Columns:     cols,
			PrimaryKeys: pks,
			WhereClause: &where,
		}
		for _, dep := range dependencyMap[table] {
			config.DependsOn = append(config.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
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

func findCircularDependencies(dependencies map[string][]string) [][]string {
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
	if len(cycle) == 0 {
		return []string{}
	}
	min := cycle[0]
	for _, node := range cycle {
		if node < min {
			min = node
		}
	}

	startIndex := -1
	for i, node := range cycle {
		if node == min && (startIndex == -1 || cycle[i-1] > cycle[(i+1)%len(cycle)]) {
			startIndex = i
		}
	}

	ordered := []string{}
	for i := 0; i < len(cycle); i++ {
		ordered = append(ordered, cycle[(startIndex+i)%len(cycle)])
	}

	return ordered
}

func getMultiTableCircularDependencies(dependencyMap map[string][]string) [][]string {
	cycles := findCircularDependencies(dependencyMap)
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
		configMap[fmt.Sprintf("%s.%s", config.Table, config.RunType)] = config
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
				seenTables[config.Table] = config.Columns
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
					seenTables[config.Table] = append(seenTables[config.Table], config.Columns...)
					delete(configMap, name)
				}
			}
		}
	}
	return true
}
