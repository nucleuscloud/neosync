package runconfigs

import (
	"sort"
	"strings"
)

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
func dfsCycles(
	start, current string,
	dependencies map[string][]string,
	recStack map[string]bool,
	path []string,
	result *[][]string,
) {
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
