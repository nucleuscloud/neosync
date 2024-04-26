package tabledependency

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
)

type RunType string

const (
	Update RunType = "update"
	Insert RunType = "insert"
)

// type SyncColumn struct {
// 	Exclude []string
// 	Include []string
// }

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
	filteredDepsMap := map[string][]string{} // only include tables that are in tables arg list
	unfilteredDepsMap := map[string][]string{}
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
			unfilteredDepsMap[table] = append(unfilteredDepsMap[table], constraint.ForeignKey.Table)
			if slices.Contains(tables, table) && slices.Contains(tables, constraint.ForeignKey.Table) {
				filteredDepsMap[table] = append(filteredDepsMap[table], constraint.ForeignKey.Table)
			}
		}
	}

	// create map containing all tables to track when each is processed
	processed := make(map[string]bool, len(tables))
	for _, t := range tables {
		processed[t] = false
	}

	// how to handle self referencing
	// create configs for tables in circular dependencies
	circularDeps := findCircularDependencies(unfilteredDepsMap)
	groupedCycles := groupDependencies(circularDeps)
	for _, group := range groupedCycles {
		if len(group) == 0 {
			continue
		}
		if len(group) == 1 {
			cycleConfigs, err := processSimplCycle(group[0], tableColumnsMap, primaryKeyMap, subsets, dependencyMap, foreignKeyColsMap)
			if err != nil {
				return nil, err
			}
			// update table processed map
			for _, cfg := range cycleConfigs {
				processed[cfg.Table] = true
			}
			configs = append(configs, cycleConfigs...)
		} else {

		}
	}

	insertConfigs := processTables(processed, unfilteredDepsMap, foreignKeyMap, tableColumnsMap, primaryKeyMap, subsets)
	configs = append(configs, insertConfigs...)

	// filter out tables from table list
	// check path

	return configs, nil
}

func processSimplCycle(
	cycle []string,
	tableColumnsMap map[string][]string,
	primaryKeyMap map[string][]string,
	subsets map[string]string,
	dependencyMap map[string]*dbschemas.TableConstraints,
	foreignKeyColsMap map[string]map[string]*ConstraintColumns,
) ([]*RunConfig, error) {
	configs := []*RunConfig{}
	// determine start table
	startTable, err := determineCycleStart(cycle, subsets, dependencyMap)
	if err != nil {
		return nil, err
	}

	// create insert and update configs for start table
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
		RunType:     Insert,
		Columns:     []string{},
		PrimaryKeys: pks,
		WhereClause: &where,
	}

	updateConfig := &RunConfig{
		Table:       startTable,
		DependsOn:   []*DependsOn{{Table: startTable, Columns: pks}},
		RunType:     Update,
		Columns:     []string{},
		PrimaryKeys: pks,
		WhereClause: &where,
	}
	deps := foreignKeyColsMap[startTable]
	for fkTable, fkCols := range deps {
		if slices.Contains(cycle, fkTable) {
			updateConfig.DependsOn = append(updateConfig.DependsOn, &DependsOn{Table: fkTable, Columns: fkCols.NullableColumns})
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

	// create insert configs for all other tables in cycle
	for _, table := range cycle {
		if table == startTable {
			continue
		}
		cols := tableColumnsMap[table]
		pks := primaryKeyMap[table]
		where := subsets[table]
		config := &RunConfig{
			Table:       table,
			DependsOn:   []*DependsOn{},
			RunType:     Insert,
			Columns:     cols,
			PrimaryKeys: pks,
			WhereClause: &where,
		}
		deps := foreignKeyColsMap[startTable]
		for fkTable, fkCols := range deps {
			config.DependsOn = append(config.DependsOn, &DependsOn{Table: fkTable, Columns: slices.Concat(fkCols.NullableColumns, fkCols.NonNullableColumns)})
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func determineCycleStart(
	cycle []string,
	subsets map[string]string,
	dependencyMap dbschemas.TableDependency,
) (string, error) {
	var start *string
	rank := 0
	for _, table := range cycle {
		table := table
		newRank := 0
		_, hasSubset := subsets[table]
		if hasSubset {
			newRank++
		}
		dependencies, ok := dependencyMap[table]
		if !ok {
			return "", fmt.Errorf("missing dependencies for table: %s", table)
		}
		if areAllFkColsNullable(dependencies.Constraints, cycle) {
			newRank++
		}
		if newRank > rank {
			start = &table
		}
	}
	if start == nil || *start == "" {
		return "", fmt.Errorf("unable to find start point for cycle: %+v", cycle)
	}

	return *start, nil
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
			RunType:     Insert,
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

type tableFkNullableCols struct {
	AreAllFkColsNullable bool
	NullableCols         []string
}

func getFkNullableCols(cycles [][]string, constraints []*dbschemas.ForeignConstraint) *tableFkNullableCols {
	nullableCols := []string{}
	allFkAreNullable := true
	fmt.Println("getFkNullableCols -----------")
	jsonF, _ := json.MarshalIndent(cycles, "", " ")
	fmt.Printf("\n cycles: %s \n", string(jsonF))
	jsonF, _ = json.MarshalIndent(constraints, "", " ")
	fmt.Printf("\n constraints: %s \n", string(jsonF))

	for _, constraint := range constraints {
		for _, cycle := range cycles {
			if slices.Contains(cycle, constraint.ForeignKey.Table) {
				if constraint.IsNullable {
					nullableCols = append(nullableCols, constraint.Column)
				} else {
					allFkAreNullable = false
				}
			}
		}
	}
	jsonF, _ = json.MarshalIndent(&tableFkNullableCols{
		AreAllFkColsNullable: allFkAreNullable,
		NullableCols:         nullableCols,
	}, "", " ")
	fmt.Printf("\n response: %s \n", string(jsonF))
	fmt.Println("------------------")
	return &tableFkNullableCols{
		AreAllFkColsNullable: allFkAreNullable,
		NullableCols:         nullableCols,
	}
}

func findOverlap(slice1, slice2 []string) []string {
	elemMap := make(map[string]bool)
	for _, item := range slice1 {
		elemMap[item] = true
	}

	var overlap []string
	for _, item := range slice2 {
		if _, found := elemMap[item]; found {
			overlap = append(overlap, item)
		}
	}

	return overlap
}

// tables with the highest nullable cols
func determineStartTable(tablesWithNullable, tablesWithSubset, cycle []string) string {
	// if only one table has nullable cols use that as start
	if len(tablesWithNullable) == 1 {
		return tablesWithNullable[0]
	}

	// if more than one table has nullable cols choose table with most nullables
	// if all have equal nullable cols count then choose one with priority to table with subset

	// if all are nullable with subsets use first table in cycle order
	if len(tablesWithNullable) == len(cycle) && len(tablesWithSubset) == len(cycle) {
		return cycle[0]
	}

	// find nullable with subset as start
	nullableSubsetOverlap := findOverlap(tablesWithNullable, tablesWithSubset)
	if len(nullableSubsetOverlap) > 0 {
		return nullableSubsetOverlap[0]
	}

	// use first tables with nullable cols
	ordered := cycleOrder(tablesWithNullable)
	return ordered[0]
}

type circularDependencyConfig struct {
	Table           string
	NullableColumns []string
	Cycles          [][]string
}

func buildCircularDependencyConfigs(cycle []string, dependencies dbschemas.TableDependency, subsets map[string]string, circularDeps [][]string) ([]*circularDependencyConfig, error) {
	configs := []*circularDependencyConfig{}

	nullablesColsMap := map[string][]string{}
	tablesWithNullable := []string{}
	tablesWithSubset := []string{}
	tableCyclesMap := map[string][][]string{}
	for _, table := range cycle {
		fmt.Printf("\n table: %s \n", table)
		cycles := getTableCirularDependencies(table, circularDeps)
		_, isSubset := subsets[table]
		nc := getFkNullableCols(cycles, dependencies[table].Constraints)
		nullablesColsMap[table] = nc.NullableCols
		tableCyclesMap[table] = cycles
		if nc.AreAllFkColsNullable {
			tablesWithNullable = append(tablesWithNullable, table)
		}
		if isSubset {
			tablesWithSubset = append(tablesWithSubset, table)
		}
	}

	if len(tablesWithNullable) == 0 {
		return nil, fmt.Errorf("found circular dependency with no nullable columns: %+v", cycle)
	}

	startTable := determineStartTable(tablesWithNullable, tablesWithSubset, cycle)
	fmt.Printf("\n start table: %s \n", startTable)
	for _, table := range cycle {
		nullableCols := []string{}
		if table == startTable {
			nullableCols = nullablesColsMap[table]
		}
		configs = append(configs, &circularDependencyConfig{
			Table:           table,
			NullableColumns: nullableCols,
			Cycles:          tableCyclesMap[table],
		})
	}
	return configs, nil
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

func isInCycle(dep string, cycles [][]string) bool {
	for _, cycle := range cycles {
		for _, table := range cycle {
			if table == dep {
				return true
			}
		}
	}
	return false
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
		order := cycleOrder(cycle)
		key := strings.Join(order, ",")
		if !seen[key] {
			seen[key] = true
			unique = append(unique, cycle)
		}
	}

	return unique
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
