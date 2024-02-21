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

func GetRunConfigs(dependencies dbschemas.TableDependency, tables []string) []*RunConfig {
	depsMap := map[string][]string{}
	filteredDepsMap := map[string][]string{}        // only include tables that are in tables arg list
	foreignKeyMap := map[string]map[string]string{} // map: table -> foreign key table -> foreign key column

	for table, constraints := range dependencies {
		foreignKeyMap[table] = map[string]string{}
		for _, constraint := range constraints.Constraints {
			depsMap[table] = append(depsMap[table], constraint.ForeignKey.Table)
			foreignKeyMap[table][constraint.ForeignKey.Table] = constraint.ForeignKey.Column
			if slices.Contains(tables, table) && slices.Contains(tables, constraint.ForeignKey.Table) {
				filteredDepsMap[table] = append(filteredDepsMap[table], constraint.ForeignKey.Table)
			}
		}
	}

	circularDeps := findCircularDependencies(filteredDepsMap)
	configs := []*RunConfig{}
	for _, table := range tables {
		cd := isInCircularDependency(table, circularDeps, dependencies)
		if cd.InCircularDependency {
			// handle circular dependencies
			// create the exclude config
			// only add empty exclude if all foreign constraints are nullable
			if len(cd.NullableColumns) != 0 {
				excludeConfig := &RunConfig{
					Table:     table,
					Columns:   &SyncColumn{Exclude: cd.NullableColumns},
					DependsOn: []*DependsOn{},
				}
				// only add depends on if not in the circular dependency
				for _, dep := range filteredDepsMap[table] {
					if !isInCycle(dep, cd.Cycles) {
						excludeConfig.DependsOn = append(excludeConfig.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
					}
				}
				configs = append(configs, excludeConfig)
			}

			// create the include config with dependencies
			includeConfig := &RunConfig{
				Table:     table,
				DependsOn: []*DependsOn{},
			}
			if len(cd.NullableColumns) != 0 {
				includeConfig.Columns = &SyncColumn{Include: cd.NullableColumns}
			}

			dependsOnMap := map[string]struct{}{}
			for _, dep := range filteredDepsMap[table] {
				_, ok := dependsOnMap[fmt.Sprintf("%s.%s", dep, foreignKeyMap[table][dep])]
				if !ok {
					includeConfig.DependsOn = append(includeConfig.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
					dependsOnMap[fmt.Sprintf("%s.%s", dep, foreignKeyMap[table][dep])] = struct{}{}
				}
			}
			configs = append(configs, includeConfig)
		} else {
			// handle non-circular dependencies
			config := &RunConfig{
				Table:     table,
				DependsOn: []*DependsOn{},
			}
			for _, dep := range filteredDepsMap[table] {
				config.DependsOn = append(config.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
			}
			configs = append(configs, config)
		}
	}

	return configs
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

type circularDependencyConfig struct {
	InCircularDependency bool
	NullableColumns      []string
	Cycles               [][]string
}

// checks if a table is in a circular dependency and returns nullable columns + cycle if true.
func isInCircularDependency(table string, circularDeps [][]string, dependencies dbschemas.TableDependency) *circularDependencyConfig {
	var nullableCols []string
	inCircularDependency := false
	cycles := [][]string{}
	for _, cycle := range circularDeps {
		if slices.Contains(cycle, table) {
			for _, constraint := range dependencies[table].Constraints {
				if constraint.IsNullable && slices.Contains(cycle, constraint.ForeignKey.Table) {
					nullableCols = append(nullableCols, constraint.Column)
				}
			}
			cycles = append(cycles, cycle)
			inCircularDependency = true
		}
	}
	return &circularDependencyConfig{
		InCircularDependency: inCircularDependency,
		NullableColumns:      nullableCols,
		Cycles:               cycles,
	}
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

// finds roots and creats children map
func findRootsAndChildren(graph map[string][]string) (map[string]struct{}, map[string][]string) {
	roots := map[string]struct{}{}
	children := map[string][]string{}
	for parent, childs := range graph {
		if _, exists := children[parent]; !exists {
			children[parent] = []string{}
		}
		for _, child := range childs {
			children[child] = append(children[child], parent)
			roots[child] = struct{}{}
		}
		roots[parent] = struct{}{}
	}
	return roots, children
}

// performs a depth-first search on the dependencies
func dfs(node string, visited map[string]bool, graph, children map[string][]string, path *[]string) {
	visited[node] = true
	*path = append(*path, node)
	for _, v := range graph[node] {
		if !visited[v] {
			dfs(v, visited, graph, children, path)
		}
	}
}

// finds all connected tables in dependency map
func findTrees(graph map[string][]string) [][]string {
	roots, children := findRootsAndChildren(graph)
	visited := map[string]bool{}
	var results [][]string
	for node := range roots {
		if !visited[node] && len(children[node]) == 0 {
			var path []string
			dfs(node, visited, graph, children, &path)
			results = append(results, path)
		}
	}
	return results
}

// takes foreign key dependency map and returns groups of tables trees
func GetTablesOrderedByDependency(dependencies map[string][]string) [][]string {
	return findTrees(dependencies)
}
