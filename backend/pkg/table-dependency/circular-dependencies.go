package tabledependency

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

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
		recStack := make(map[string]bool)
		path := []string{}
		dfsCycles(node, node, dependencies, recStack, path, &result)
	}
	return uniqueCycles(result)
}

// finds all possible path variations
func dfsCycles(start, current string, dependencies map[string][]string, recStack map[string]bool, path []string, result *[][]string) {
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
		dfsCycles(start, neighbor, dependencies, recStack, path, result)
	}

	recStack[current] = false
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
