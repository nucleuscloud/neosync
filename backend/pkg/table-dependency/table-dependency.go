package tabledependency

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
)

/*
		need function to get list of benthos configs to create if there is circular dependency

    // self depedency
    { table: public.a, columns: {exclude: [a_id]}} // exclude nullable
    { table: public.a, columns: {inlcude: [a_id]}} // include nullable

	  { table: public.a, columns: { exclude: [b_id] }},
		{ table: public.b, columns: { exclude: [bb_id]}, dependsOn: { table: a, columns: [id]}},
    // last two can happen at same time
		{ table: public.a, columns: { include: [b_id] }, dependsOn: { table: b, columns: [id]}},
		{ table: public.b, columns: { include: [bb_id] }, dependsOn: { table: b, columns: [id]}},

    more complex

	  { table: public.a, columns: { exclude: [b_id] },
		{ table: public.b, columns: { exclude: [bb_id] }, dependsOn: { table: a, columns: [id]}},
		{ table: public.b, columns: { include: [bb_id] }, dependsOn: { table: b, columns: [id]}},
    { table: public.a, columns: { include: [b_id] }, dependsOn: { table: b, columns: [bb_id]}},
*/

type SyncColumn struct {
	Exclude []string
	Include []string
}

type DependsOn struct {
	Table   string
	Columns []string
}

type SyncConfig struct {
	Table     string
	Columns   *SyncColumn  // rename to sync columns??
	DependsOn []*DependsOn // should this be a map?
}

func getTableCyles(table string, cycles [][]string) [][]string {
	tableCycles := [][]string{}
	for _, c := range cycles {
		for _, t := range c {
			if t == table {
				tableCycles = append(tableCycles, c)
			}
		}
	}
	return tableCycles
}

func getForeignKeyColMap(constraints []*dbschemas.ForeignConstraint) map[string][]string {
	fkMap := map[string][]string{}
	for _, d := range constraints {
		_, okFk := fkMap[d.ForeignKey.Table]
		if okFk {
			fkMap[d.ForeignKey.Table] = append(fkMap[d.ForeignKey.Table], d.ForeignKey.Column)
		} else {
			fkMap[d.ForeignKey.Table] = []string{d.ForeignKey.Column}
		}
	}
	return fkMap
}

/*

  CURRENTLY NOT WORKING
	"table_constraints": {
			"public.a": {
				"constraints": [
					{
						"column": "b_id",
						"isNullable": true,
						"foreignKey": { "table": "public.b", column: "id" }
					}
				]
			},
			"public.b": {
				"constraints": [
					{
						"column": "a_id",
						"isNullable": false,
						"foreignKey": { "table": "public.a", column: "id" }
					},
					{
						"column": "bb_id",
						"isNullable": true,
						"foreignKey": { "table": "public.b", column: "id" }
					},
				]
			},
		}
	}

  // ACTUAL
  { table: public.a, columns: { exclude: [b_id] }},
	{ table: public.a, columns: { include: [b_id] }, dependsOn: { table: b, columns: [id]}},
	{ table: public.b, dependsOn: { table: a, columns: [id]}},
  { table: public.b, columns: { exclude: [bb_id] }},
  { table: public.b, columns: { include: [bb_id] }, dependsOn: { table: b, columns: [id]}},




  // WANT
   { table: public.a, columns: { exclude: [b_id] }},
		{ table: public.b, columns: { exclude: [bb_id]}, dependsOn: { table: a, columns: [id]}},
    // last two can happen at same time
		{ table: public.a, columns: { include: [b_id] }, dependsOn: { table: b, columns: [id]}},
		{ table: public.b, columns: { include: [bb_id] }, dependsOn: { table: b, columns: [id]}},
*/

/*
 how many levels of self reference

 Table A
 id
 a_id -> FK id
 other_id
 otherother_id -> FK id

 dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
						{Column: "aa_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			}

{ table: public.a, columns: { exclude: [a_id, aa_id] }}
{ table: public.a columns: { include: [a_id, aa_id] }}


dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
						{Column: "bb_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
        "public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
           {Column: "a_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			}

{ table: public.b columns: { exclude: [a_id] }}
{ table: public.a }
{ table public.b columns: { include: [a_id]}}


dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
						{Column: "bb_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
        "public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
           {Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
			}

{ table: public.a columns: { exclude: [b_id, bb_id] }}
{ table: public.b }
{ table public.a columns: { include: [b_id, bb_id]}}

*/

/*
		dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
            {Column: "d_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
        "public.d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
					},
				},
         "public.e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},

			},

        {Table: "public.a", Columns: &SyncColumn{Exclude: []string{"b_id"}}, DependsOn: []*DependsOn{}},
				{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
        {Table: "public.e", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
        {Table: "public.d", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},


      dbschemas.TableDependency{
				"public.a": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},
				"public.b": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "c_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.c", Column: "id"}},
            {Column: "d_id", IsNullable: true, ForeignKey: &dbschemas.ForeignKey{Table: "public.d", Column: "id"}},
					},
				},
				"public.c": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "a_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.a", Column: "id"}},
					},
				},
        "public.d": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "e_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.e", Column: "id"}},
					},
				},
         "public.e": &dbschemas.TableConstraints{
					Constraints: []*dbschemas.ForeignConstraint{
						{Column: "b_id", IsNullable: false, ForeignKey: &dbschemas.ForeignKey{Table: "public.b", Column: "id"}},
					},
				},

			},

        {Table: "public.b", Columns: &SyncColumn{Exclude: []string{"c_id", "d_id"}}, DependsOn: []*DependsOn{}},
        {Table: "public.e", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
        {Table: "public.a", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
        {Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
        {Table: "public.d", DependsOn: []*DependsOn{{Table: "public.e", Columns: []string{"id"}}}},
        {Table: "public.b", Columns: &SyncColumn{Include: []string{"c_id", "d_id"}}, DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},


				{Table: "public.c", DependsOn: []*DependsOn{{Table: "public.a", Columns: []string{"id"}}}},
        {Table: "public.d", DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},
				{Table: "public.b", DependsOn: []*DependsOn{{Table: "public.c", Columns: []string{"id"}}, {Table: "public.d", Columns: []string{"id"}}}},
				{Table: "public.a", Columns: &SyncColumn{Include: []string{"b_id"}}, DependsOn: []*DependsOn{{Table: "public.b", Columns: []string{"id"}}}},



*/

/*
what if all are nullable?? what should happen
*/

/*
{
  public.a: { nullableColumns: [], notNullableColumns: []}
}

*/

type ConstraintColumns struct {
	NullableColumns    []string
	NonNullableColumns []string
}

// func getSchemaDependencyMap(deps dbschemas.TableDependency) map[string]map[string]ConstraintColumns {
//   depsMap := map[string]map[string]ConstraintColumns{}
//   for table, constraints := range deps {
//     colMap := map[string]ConstraintColumns{}
//     for _, c := range constraints.Constraints {

//     }

//     depsMap[table] = colMap
//   }
//   return depsMap
// }

func GetSyncConfigs(dependencies dbschemas.TableDependency, tables []string) []*SyncConfig {
	depsMap := make(map[string][]string)
	foreignKeyMap := make(map[string]map[string]string) // Map: table -> foreign key column -> referenced column

	for table, constraints := range dependencies {
		foreignKeyMap[table] = make(map[string]string)
		for _, constraint := range constraints.Constraints {
			depsMap[table] = append(depsMap[table], constraint.ForeignKey.Table)
			foreignKeyMap[table][constraint.ForeignKey.Table] = constraint.ForeignKey.Column
		}
	}

	jsonFg, _ := json.MarshalIndent(depsMap, "", " ")
	fmt.Printf("\n\n  depsMap: %s \n\n", string(jsonFg))

	circularDeps := findCircularDependencies(depsMap)
	jsonFr, _ := json.MarshalIndent(circularDeps, "", " ")
	fmt.Printf("\n\n   circularDeps: %s \n\n", string(jsonFr))
	configs := []*SyncConfig{}

	for _, table := range tables {
		inCircularDep, cycles, nullableCols := isInCircularDependency(table, circularDeps, dependencies)
		if inCircularDep {
			// Handle circular dependencies
			// First, create the exclude config
			// only add empty exclude if all foreign constraints are nullable
			if len(nullableCols) != 0 {
				excludeConfig := &SyncConfig{
					Table:     table,
					Columns:   &SyncColumn{Exclude: nullableCols},
					DependsOn: []*DependsOn{},
				}
				jsonF, _ := json.MarshalIndent(depsMap[table], "", " ")
				fmt.Printf("\n\n   depsMap[table]: %s \n\n", string(jsonF))
				// Only add depends on if not in the circular dependency
				for _, dep := range depsMap[table] {
					if !isDependencyInCycle(dep, cycles) {
						excludeConfig.DependsOn = append(excludeConfig.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
					}
				}
				jsonFt, _ := json.MarshalIndent(excludeConfig, "", " ")
				fmt.Printf("\n\n  excludeConfig: %s \n\n", string(jsonFt))

				configs = append(configs, excludeConfig)
			}

			// Then, create the include config with dependencies
			includeConfig := &SyncConfig{
				Table:     table,
				DependsOn: []*DependsOn{},
			}
			if len(nullableCols) != 0 {
				includeConfig.Columns = &SyncColumn{Include: nullableCols}
			}

			dependsOnMap := map[string]struct{}{}
			for _, dep := range depsMap[table] {
				_, ok := dependsOnMap[fmt.Sprintf("%s.%s", dep, foreignKeyMap[table][dep])]
				if !ok {
					includeConfig.DependsOn = append(includeConfig.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
					dependsOnMap[fmt.Sprintf("%s.%s", dep, foreignKeyMap[table][dep])] = struct{}{}
				}
			}
			configs = append(configs, includeConfig)
		} else {
			// Handle non-circular dependencies
			config := &SyncConfig{
				Table:     table,
				DependsOn: []*DependsOn{},
			}
			for _, dep := range depsMap[table] {
				config.DependsOn = append(config.DependsOn, &DependsOn{Table: dep, Columns: []string{foreignKeyMap[table][dep]}})
			}
			configs = append(configs, config)
		}
	}

	return configs
}

func isDependencyInCycle(dep string, cycles [][]string) bool {
	for _, cycle := range cycles {
		for _, table := range cycle {
			if table == dep {
				return true
			}
		}
	}
	return false

}

// isInCircularDependency checks if a table is in a circular dependency and returns nullable columns if true.
func isInCircularDependency(table string, circularDeps [][]string, dependencies dbschemas.TableDependency) (bool, [][]string, []string) {
	var nullableCols []string
	inCircularDependency := false
	cycles := [][]string{}
	for _, cycle := range circularDeps {
		if isInSlice(table, cycle) {
			fmt.Printf("----  table: %s  ----- \n", table)
			jsonFx, _ := json.MarshalIndent(cycle, "", " ")
			fmt.Printf("\n  %s \n", string(jsonFx))
			for _, constraint := range dependencies[table].Constraints {
				jsonF, _ := json.MarshalIndent(constraint, "", " ")
				fmt.Printf("\n  %s \n", string(jsonF))
				if constraint.IsNullable && slices.Contains(cycle, constraint.ForeignKey.Table) {
					// if constraint.IsNullable {
					nullableCols = append(nullableCols, constraint.Column)
				}
			}
			cycles = append(cycles, cycle)
			inCircularDependency = true
			// return true, cycle, nullableCols
		}
	}
	return inCircularDependency, cycles, nullableCols
}

// isInSlice checks if an item is in a slice.
func isInSlice(item string, slice []string) bool {
	for _, elem := range slice {
		if item == elem {
			return true
		}
	}
	return false
}

func findCircularDependencies(dependencies map[string][]string) [][]string {
	visited := make(map[string]bool)
	var result [][]string

	for node := range dependencies {
		if !visited[node] {
			cycles := dfs(node, dependencies, visited, make(map[string]bool), []string{})
			result = append(result, cycles...)
		}
	}
	return uniqueCycles(result)
}

func dfs(node string, dependencies map[string][]string, visited, recStack map[string]bool, path []string) [][]string {
	if recStack[node] {
		index := findInPath(node, path)
		if index != -1 {
			return [][]string{path[index:]}
		}
		return nil
	}

	if visited[node] {
		return nil
	}

	visited[node] = true
	recStack[node] = true
	path = append(path, node)
	var cycles [][]string

	for _, neighbor := range dependencies[node] {
		foundCycles := dfs(neighbor, dependencies, visited, recStack, path)
		cycles = append(cycles, foundCycles...)
	}

	recStack[node] = false
	return cycles
}

func findInPath(node string, path []string) int {
	for i, n := range path {
		if n == node {
			return i
		}
	}
	return -1
}

func uniqueCycles(cycles [][]string) [][]string {
	seen := make(map[string]bool)
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

// func findCircularDependencies(dependencies map[string][]string) [][]string {
// 	allVisited := make(map[string]bool)
// 	var result [][]string

// 	for node := range dependencies {
// 		if !allVisited[node] {
// 			visited, recStack := make(map[string]bool), make(map[string]bool)
// 			cycle := dfs(node, dependencies, visited, recStack, node, []string{})
// 			if len(cycle) > 0 {
// 				result = append(result, cycle)
// 				for _, n := range cycle {
// 					allVisited[n] = true
// 				}
// 			}
// 		}
// 	}
// 	return result
// }

// func dfs(current string, dependencies map[string][]string, visited, recStack map[string]bool, start string, path []string) []string {
// 	if recStack[current] {
// 		if current == start {
// 			return path
// 		}
// 		return nil
// 	}

// 	if visited[current] {
// 		return nil
// 	}

// 	visited[current] = true
// 	recStack[current] = true
// 	path = append(path, current)

// 	for _, neighbor := range dependencies[current] {
// 		if cycle := dfs(neighbor, dependencies, visited, recStack, start, path); cycle != nil {
// 			return cycle
// 		}
// 	}

// 	recStack[current] = false
// 	return nil
// }

func removeDuplicateCycles(cycles [][]string) [][]string {
	// Normalize each cycle
	for i, cycle := range cycles {
		cycles[i] = normalizeCycle(cycle)
	}

	// Use a map to remove duplicates
	uniqueCycles := make(map[string][]string)
	for _, cycle := range cycles {
		key := generateKey(cycle)
		if _, exists := uniqueCycles[key]; !exists {
			uniqueCycles[key] = cycle
		}
	}

	// Convert the map back to a slice
	var result [][]string
	for _, cycle := range uniqueCycles {
		result = append(result, cycle)
	}

	return result
}

// normalizeCycle normalizes the cycle by rotating it so that its smallest element is first.
func normalizeCycle(cycle []string) []string {
	if len(cycle) == 0 {
		return cycle
	}

	smallestIndex := 0
	for i := range cycle {
		if cycle[i] < cycle[smallestIndex] {
			smallestIndex = i
		}
	}

	// Rotate the slice
	return append(cycle[smallestIndex:], cycle[:smallestIndex]...)
}

// generateKey creates a unique key for a cycle by joining its elements.
func generateKey(cycle []string) string {
	sort.Strings(cycle) // Ensuring consistent order for the key
	return fmt.Sprint(cycle)
}

// func findCircularDependencies(deps map[string][]string) [][]string {
// 	visited := make(map[string]bool)
// 	stack := make(map[string]bool)
// 	var cycles [][]string

// 	for node := range deps {
// 		if !visited[node] {
// 			var path []string
// 			if cycle := findCycle(node, deps, visited, stack, path); len(cycle) > 0 {
// 				cycles = append(cycles, cycle)
// 			}
// 		}
// 	}
// 	fmt.Printf("%+v\n", cycles)

// 	return cycles
// }

// func findCycle(node string, deps map[string][]string, visited, stack map[string]bool, path []string) []string {
// 	visited[node] = true
// 	stack[node] = true
// 	path = append(path, node)

// 	for _, dep := range deps[node] {
// 		if !visited[dep] {
// 			if cycle := findCycle(dep, deps, visited, stack, path); len(cycle) > 0 {
// 				return cycle
// 			}
// 		} else if stack[dep] {
// 			for i, n := range path {
// 				if n == dep {
// 					return path[i:]
// 				}
// 			}
// 		}
// 	}
// 	stack[node] = false
// 	return nil
// }

/*
  need function to get root benthos configs and dependent benthod configs
*/

/*
  need way to figure out if benthos config can run based on dependencies
*/
