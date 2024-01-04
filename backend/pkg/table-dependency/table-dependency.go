package tabledependency

import (
	"encoding/json"
	"fmt"
	"slices"

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

	circularDeps := findCircularDependencies(depsMap)
	configs := []*SyncConfig{}

	for _, table := range tables {
		inCircularDep, nullableCols := isInCircularDependency(table, circularDeps, dependencies)
		if inCircularDep {
			// Handle circular dependencies
			// First, create the exclude config
			// only add empty exclude if all foreign constraints are nullable
			if len(nullableCols) != 0 {
				excludeConfig := &SyncConfig{
					Table:   table,
					Columns: &SyncColumn{Exclude: nullableCols},
				}

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

// isInCircularDependency checks if a table is in a circular dependency and returns nullable columns if true.
func isInCircularDependency(table string, circularDeps [][]string, dependencies dbschemas.TableDependency) (bool, []string) {
	var nullableCols []string
	for _, cycle := range circularDeps {
		if isInSlice(table, cycle) {
			for _, constraint := range dependencies[table].Constraints {
				if constraint.IsNullable {
					nullableCols = append(nullableCols, constraint.Column)
				}
			}
			return true, nullableCols
		}
	}
	return false, nil
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

func GetSyncConfigsOld(tableDeps dbschemas.TableDependency, tables []string) []*SyncConfig {
	configs := []*SyncConfig{}

	tableDepsMap := dbschemas.BuildDependsOnSlice(tableDeps)
	fmt.Printf("%+v\n\n", tableDepsMap)

	cycles := findCircularDependencies(tableDepsMap)
	// cycles := [][]string{{"public.a", "public.b"}, {"public.b"}}
	fmt.Printf("cycles %+v\n", cycles)

	for _, t := range tables {
		tableCycles := getTableCyles(t, cycles)
		fmt.Printf("table: %s \n", t)
		if len(tableCycles) > 0 {
			for _, tc := range tableCycles {
				deps := tableDeps[t]
				for _, d := range deps.Constraints {
					if slices.Contains(tc, d.ForeignKey.Table) {
						// update this logic to do all nullable at once
						if d.IsNullable {
							configs = append(
								configs,
								&SyncConfig{Table: t, Columns: &SyncColumn{Exclude: []string{d.Column}}, DependsOn: []*DependsOn{}},
								&SyncConfig{Table: t, Columns: &SyncColumn{Include: []string{d.Column}}, DependsOn: []*DependsOn{{Table: d.ForeignKey.Table, Columns: []string{d.ForeignKey.Column}}}},
							)
						} else {
							configs = append(
								configs,
								&SyncConfig{Table: t, DependsOn: []*DependsOn{{Table: d.ForeignKey.Table, Columns: []string{d.ForeignKey.Column}}}},
							)

						}

					}

				}

			}

		} else {
			fmt.Println("No table cycles")
			// no cycles
			c := &SyncConfig{Table: t, DependsOn: []*DependsOn{}}
			if deps, ok := tableDeps[t]; ok {
				fkMap := getForeignKeyColMap(deps.Constraints)
				for t, cols := range fkMap {
					c.DependsOn = append(c.DependsOn, &DependsOn{Table: t, Columns: cols})
				}
			}
			configs = append(configs, c)
		}
	}

	jsonFx, _ := json.MarshalIndent(configs, "", " ")
	fmt.Printf("\n\n  %s \n\n", string(jsonFx))

	return configs
}

func findCircularDependencies(deps map[string][]string) [][]string {
	visited := make(map[string]bool)
	stack := make(map[string]bool)
	var cycles [][]string

	for node := range deps {
		if !visited[node] {
			var path []string
			if cycle := findCycle(node, deps, visited, stack, path); len(cycle) > 0 {
				cycles = append(cycles, cycle)
			}
		}
	}
	fmt.Printf("%+v\n", cycles)

	return cycles
}

func findCycle(node string, deps map[string][]string, visited, stack map[string]bool, path []string) []string {
	visited[node] = true
	stack[node] = true
	path = append(path, node)

	for _, dep := range deps[node] {
		if !visited[dep] {
			if cycle := findCycle(dep, deps, visited, stack, path); len(cycle) > 0 {
				return cycle
			}
		} else if stack[dep] {
			for i, n := range path {
				if n == dep {
					return path[i:]
				}
			}
		}
	}
	stack[node] = false
	return nil
}

/*
  need function to get root benthos configs and dependent benthod configs
*/

/*
  need way to figure out if benthos config can run based on dependencies
*/
