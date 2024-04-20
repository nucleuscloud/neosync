package tabledependency

import (
	"fmt"
	"slices"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
)

type SyncColumn struct {
	Exclude []string
	Include []string
}

type DependsOn struct {
	Table   string
	Columns []string
}

type RunConfig struct {
	Table     string
	Columns   *SyncColumn
	DependsOn []*DependsOn
}

type ConstraintColumns struct {
	NullableColumns    []string
	NonNullableColumns []string
}

func GetRunConfigs(dependencies dbschemas.TableDependency, tables []string, subsets map[string]string) []*RunConfig {
	filteredDepsMap := map[string][]string{}        // only include tables that are in tables arg list
	foreignKeyMap := map[string]map[string]string{} // map: table -> foreign key table -> foreign key column
	configs := []*RunConfig{}

	for table, constraints := range dependencies {
		foreignKeyMap[table] = map[string]string{}
		for _, constraint := range constraints.Constraints {
			foreignKeyMap[table][constraint.ForeignKey.Table] = constraint.ForeignKey.Column
			if slices.Contains(tables, table) && slices.Contains(tables, constraint.ForeignKey.Table) {
				filteredDepsMap[table] = append(filteredDepsMap[table], constraint.ForeignKey.Table)
			}
		}
	}

	processed := make(map[string]bool, len(tables))
	for _, t := range tables {
		processed[t] = false
	}

	// create configs for tables in circular dependencies
	circularDeps := findCircularDependencies(filteredDepsMap)
	for _, cycle := range circularDeps {
		cdConfigs := buildCircularDependencyConfigs(cycle, dependencies, subsets, circularDeps)
		for _, cfg := range cdConfigs {
			if processed[cfg.Table] {
				continue
			}
			if len(cfg.NullableColumns) != 0 {
				excludeConfig := &RunConfig{
					Table:     cfg.Table,
					Columns:   &SyncColumn{Exclude: cfg.NullableColumns},
					DependsOn: []*DependsOn{},
				}
				// only add depends on if not in the circular dependency
				for _, dep := range filteredDepsMap[cfg.Table] {
					if !isInCycle(dep, cfg.Cycles) {
						excludeConfig.DependsOn = append(excludeConfig.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[cfg.Table][dep]}})
					}
				}
				configs = append(configs, excludeConfig)
			}

			// create the include config with dependencies
			includeConfig := &RunConfig{
				Table:     cfg.Table,
				DependsOn: []*DependsOn{},
			}
			if len(cfg.NullableColumns) != 0 {
				includeConfig.Columns = &SyncColumn{Include: cfg.NullableColumns}
			}

			dependsOnMap := map[string]struct{}{}
			for _, dep := range filteredDepsMap[cfg.Table] {
				_, ok := dependsOnMap[fmt.Sprintf("%s.%s", dep, foreignKeyMap[cfg.Table][dep])]
				if !ok {
					includeConfig.DependsOn = append(includeConfig.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[cfg.Table][dep]}})
					dependsOnMap[fmt.Sprintf("%s.%s", dep, foreignKeyMap[cfg.Table][dep])] = struct{}{}
				}
			}
			configs = append(configs, includeConfig)
			processed[cfg.Table] = true
		}
	}

	// create configs for non-circular dependent tables
	for table, isProcessed := range processed {
		if isProcessed {
			continue
		}
		config := &RunConfig{
			Table:     table,
			DependsOn: []*DependsOn{},
		}
		for _, dep := range filteredDepsMap[table] {
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
	for _, constraint := range constraints {
		// need to check if in any of the cycles containing table
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

func determineStartTable(tablesWithNullable, tablesWithSubset, cycle []string) string {
	// if only one table has nullable cols use that as start
	if len(tablesWithNullable) == 1 {
		return tablesWithNullable[0]
	}

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
	return tablesWithNullable[0]
}

type circularDependencyConfig struct {
	Table           string
	NullableColumns []string
	Cycles          [][]string
}

func buildCircularDependencyConfigs(cycle []string, dependencies dbschemas.TableDependency, subsets map[string]string, circularDeps [][]string) []*circularDependencyConfig {
	configs := []*circularDependencyConfig{}

	nullablesColsMap := map[string][]string{}
	tablesWithNullable := []string{}
	tablesWithSubset := []string{}
	tableCyclesMap := map[string][][]string{}
	for _, table := range cycle {
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

	// no nullable cols or no subsets then use order as is
	if len(tablesWithNullable) == 0 || len(tablesWithSubset) == 0 {
		for _, table := range cycle {
			configs = append(configs, &circularDependencyConfig{
				Table:           table,
				NullableColumns: nullablesColsMap[table],
				Cycles:          tableCyclesMap[table],
			})
		}
		return configs
	}

	startTable := determineStartTable(tablesWithNullable, tablesWithSubset, cycle)
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
		key := cycleKey(cycle)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, cycle)
		}
	}

	return unique
}

func cycleKey(cycle []string) string {
	if len(cycle) == 0 {
		return ""
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

	key := ""
	for i := 0; i < len(cycle); i++ {
		key += cycle[(startIndex+i)%len(cycle)] + ","
	}

	return key
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

func GetTablesOrderedByDependency(dependencyMap map[string][]string) ([]string, bool, error) {
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
			return nil, false, fmt.Errorf("unable to build table order")
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

	return orderedTables, hasCycles, nil
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
