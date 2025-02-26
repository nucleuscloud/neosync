package runconfigs

import (
	"fmt"
	"slices"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

// tableConfigsBuilder holds global data about all tables in the job
// and computes shared information like circular dependencies and subset paths
// that will be used by the runConfigBuilder to create individual RunConfig instances.
type tableConfigsBuilder struct {
	primaryKeys          map[string][]string
	whereClauses         map[string]string
	columns              map[string][]string
	uniqueIndexes        map[string][][]string
	uniqueConstraints    map[string][][]string
	foreignKeys          map[string][]*sqlmanager_shared.ForeignConstraint
	circularDependencies map[string]bool
	subsetPaths          map[string][]*SubsetPath
}

func newTableConfigsBuilder(
	columns map[string][]string,
	primaryKeys map[string][]string,
	whereClauses map[string]string,
	uniqueIndexes map[string][][]string,
	uniqueConstraints map[string][][]string,
	foreignKeys map[string][]*sqlmanager_shared.ForeignConstraint,
) *tableConfigsBuilder {
	b := &tableConfigsBuilder{
		columns:           columns,
		primaryKeys:       primaryKeys,
		whereClauses:      whereClauses,
		uniqueIndexes:     uniqueIndexes,
		uniqueConstraints: uniqueConstraints,
		foreignKeys:       foreignKeys,
	}
	// find circular dependencies
	graph := b.buildDependencyGraph()
	circularDeps := FindCircularDependencies(graph)
	b.circularDependencies = circularDependencyTables(circularDeps)

	// compute subset paths
	b.subsetPaths = b.computeAllSubsetPaths()
	return b
}

func (b *tableConfigsBuilder) Build(table sqlmanager_shared.SchemaTable) []*RunConfig {
	tableKey := table.String()
	whereClause := b.whereClauses[tableKey]
	return newRunConfigBuilder(
		table,
		b.columns[tableKey],
		b.primaryKeys[tableKey],
		&whereClause,
		b.uniqueIndexes[tableKey],
		b.uniqueConstraints[tableKey],
		b.foreignKeys[tableKey],
		b.circularDependencies[tableKey],
		b.subsetPaths[tableKey],
	).Build()
}

// buildDependencyGraph builds a dependency graph from a map of table names to their foreign constraints tables
func (b *tableConfigsBuilder) buildDependencyGraph() map[string][]string {
	graph := make(map[string][]string)
	for table, constraints := range b.foreignKeys {
		for _, constraint := range constraints {
			graph[table] = append(graph[table], constraint.ForeignKey.Table)
		}
	}
	return graph
}

// computeAllSubsetPaths computes, for each table, the shortest paths (if any)
// to any table that has a where clause. It returns a map from table name to a slice
// of SubsetPath—one per where-clause root reachable.
func (b *tableConfigsBuilder) computeAllSubsetPaths() map[string][]*SubsetPath {
	// Build the reverse graph: for each parent table, list its child tables.
	reverseGraph := make(map[string][]string)
	for child, constraints := range b.foreignKeys {
		for _, fc := range constraints {
			parent := fc.ForeignKey.Table
			reverseGraph[parent] = append(reverseGraph[parent], child)
		}
	}

	// Global result: table -> list of SubsetPath.
	result := make(map[string][]*SubsetPath)

	// Multi-source BFS starting from every table that has a where clause.
	// The bfsEntry now also carries the joinSteps along the path.
	type bfsEntry struct {
		src       string      // the where-clause root
		current   string      // current table
		joinSteps []*JoinStep // join steps taken from src to current
	}
	queue := []bfsEntry{}
	// visited[src][node] ensures we record only the shortest path per source.
	visited := make(map[string]map[string]bool)

	// Initialize the queue with each where-clause root.
	for root, clause := range b.whereClauses {
		if visited[root] == nil {
			visited[root] = make(map[string]bool)
		}
		visited[root][root] = true
		// For the root itself, record its own path (with no join steps).
		result[root] = append(result[root], &SubsetPath{
			Root:      root,
			Subset:    clause,
			JoinSteps: []*JoinStep{},
		})
		// Enqueue the root so we can traverse to its children.
		queue = append(queue, bfsEntry{src: root, current: root, joinSteps: []*JoinStep{}})
	}

	// Process the BFS queue.
	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]
		// For each child of the current table.
		for _, child := range reverseGraph[entry.current] {
			if visited[entry.src] == nil {
				visited[entry.src] = make(map[string]bool)
			}
			if visited[entry.src][child] {
				continue
			}
			visited[entry.src][child] = true

			// Find a matching foreign constraint between entry.current and child.
			var js *JoinStep
			found := false
			// Iterate over all foreign constraints for the child.
			for _, fc := range b.foreignKeys[child] {
				// We are looking for a foreign key where the parent table matches entry.current.
				if fc.ForeignKey.Table == entry.current {
					// For simplicity, take the first column pair as the join keys.
					if len(fc.ForeignKey.Columns) > 0 && len(fc.Columns) > 0 {
						referenceSchema, referenceTable := sqlmanager_shared.SplitTableKey(fc.ForeignKey.Table)
						js = &JoinStep{
							ToKey:   entry.current,
							FromKey: child,
							// Create a new ForeignKey value from fc.ForeignKey.
							ForeignKey: &ForeignKey{
								Columns:          fc.Columns,
								NotNullable:      fc.NotNullable,
								ReferenceSchema:  referenceSchema,
								ReferenceTable:   referenceTable,
								ReferenceColumns: fc.ForeignKey.Columns,
							},
						}
						found = true
						break
					}
				}
			}
			// If no matching foreign key is found, we still propagate without a joinStep.
			newJoinSteps := make([]*JoinStep, len(entry.joinSteps))
			copy(newJoinSteps, entry.joinSteps)
			if found {
				newJoinSteps = append(newJoinSteps, js)
			}

			revJoinSteps := reverseJoinSteps(newJoinSteps)
			result[child] = append(result[child], &SubsetPath{
				Root:      entry.src,
				Subset:    b.whereClauses[entry.src],
				JoinSteps: revJoinSteps,
			})
			queue = append(queue, bfsEntry{src: entry.src, current: child, joinSteps: newJoinSteps})
		}
	}

	return result
}

// reverseJoinSteps reverses a slice of JoinSteps.
func reverseJoinSteps(steps []*JoinStep) []*JoinStep {
	n := len(steps)
	result := make([]*JoinStep, n)
	for i, v := range steps {
		result[n-1-i] = v
	}
	return result
}

// RunConfigBuilder is responsible for generating RunConfigs that define how to process table data.
// It handles two main scenarios:
// 1. Tables without circular dependencies, subsets - generates a single INSERT config
// 2. Tables with circular dependencies or subsets - generates multiple configs to handle the cycle/subset:
//   - Initial INSERT with non-nullable foreign key columns
//   - UPDATE configs for each nullable foreign key reference
//
// This allows for properly ordered data synchronization while maintaining referential integrity.

type runConfigBuilder struct {
	table                      sqlmanager_shared.SchemaTable
	primaryKeys                []string
	whereClause                *string
	columns                    []string
	uniqueIndexes              [][]string
	uniqueConstraints          [][]string
	foreignKeys                []*sqlmanager_shared.ForeignConstraint
	isPartOfCircularDependency bool
	subsetPaths                []*SubsetPath
}

func newRunConfigBuilder(
	table sqlmanager_shared.SchemaTable,
	columns []string,
	primaryKeys []string,
	whereClause *string,
	uniqueIndexes [][]string,
	uniqueConstraints [][]string,
	foreignKeys []*sqlmanager_shared.ForeignConstraint,
	isPartOfCircularDependency bool,
	subsetPaths []*SubsetPath,
) *runConfigBuilder {
	return &runConfigBuilder{
		table:                      table,
		primaryKeys:                primaryKeys,
		whereClause:                whereClause,
		columns:                    columns,
		uniqueIndexes:              uniqueIndexes,
		uniqueConstraints:          uniqueConstraints,
		foreignKeys:                foreignKeys,
		isPartOfCircularDependency: isPartOfCircularDependency,
		subsetPaths:                subsetPaths,
	}
}

// Build generates a list of RunConfig objects based on the table's properties and relationships.
// Decides whether to build a single INSERT config or multiple configs based on whether the table
// is part of a circular dependency or has subsets.
func (b *runConfigBuilder) Build() []*RunConfig {
	if b.isPartOfCircularDependency || len(b.subsetPaths) > 0 {
		return b.buildConstraintHandlingConfigs()
	} else {
		return []*RunConfig{b.buildInsertConfig()}
	}
}

func (b *runConfigBuilder) buildInsertConfig() *RunConfig {
	config := &RunConfig{
		id:             fmt.Sprintf("%s.%s", b.table, RunTypeInsert),
		table:          b.table,
		runType:        RunTypeInsert,
		selectColumns:  b.columns,
		insertColumns:  b.columns,
		primaryKeys:    b.primaryKeys,
		whereClause:    b.whereClause,
		orderByColumns: b.getOrderByColumns(b.columns),
		dependsOn:      b.getDependsOn(),
		subsetPaths:    b.subsetPaths,
	}
	return config
}

// buildConstraintHandlingConfigs builds multiple configs to handle circular dependencies or subsets.
// It handles circular dependencies and subsets the same way:
// Inserts all non-nullable columns, then builds update configs for any nullable columns that reference other tables.
// This is important to bring over as much subset data as possible.
func (b *runConfigBuilder) buildConstraintHandlingConfigs() []*RunConfig {
	var configs []*RunConfig

	var where *string
	if b.whereClause != nil {
		where = b.whereClause
	}

	orderByColumns := b.getOrderByColumns(b.columns)
	insertConfig := &RunConfig{
		id:             fmt.Sprintf("%s.%s", b.table, RunTypeInsert),
		table:          b.table,
		runType:        RunTypeInsert,
		selectColumns:  b.columns, // Select columns in insert config must be all columns due to S3 as possible output
		insertColumns:  b.primaryKeys,
		primaryKeys:    b.primaryKeys,
		whereClause:    where,
		orderByColumns: orderByColumns,
		dependsOn:      []*DependsOn{},
		subsetPaths:    b.subsetPaths,
	}

	// Track which columns still need to be inserted (that aren’t handled by constraints).
	remainingColumns := make(map[string]bool, len(b.columns))
	for _, col := range b.columns {
		if slices.Contains(b.primaryKeys, col) {
			continue
		}
		remainingColumns[col] = true
	}

	updateConfigCount := 0
	// Build update configs for any nullable foreign keys
	for _, fc := range b.foreignKeys {
		if fc == nil || fc.ForeignKey == nil {
			continue
		}

		insertCols, insertFkCols, updateCols, updateFkCols := []string{}, []string{}, []string{}, []string{}

		// Classify each constrained column into insert vs. update groups
		// based on whether the column is NOT NULL.
		for i, col := range fc.Columns {
			// Mark this column as handled in constraints (so we don’t insert it again later).
			remainingColumns[col] = false

			if fc.NotNullable[i] {
				insertCols = append(insertCols, col)
				insertFkCols = append(insertFkCols, fc.ForeignKey.Columns[i])
			} else {
				updateCols = append(updateCols, col)
				updateFkCols = append(updateFkCols, fc.ForeignKey.Columns[i])
			}
		}

		// For NOT NULL constraints, we can safely insert them now (but they depend on the referenced table).
		if len(insertCols) > 0 {
			insertConfig.insertColumns = append(insertConfig.insertColumns, insertCols...)
			insertConfig.dependsOn = append(insertConfig.dependsOn, &DependsOn{
				Table:   fc.ForeignKey.Table,
				Columns: insertFkCols,
			})
		}

		// For columns that can be null, we do them after the main insert (Update).
		if len(updateCols) > 0 {
			updateConfigCount++
			updateConfig := b.buildUpdateConfig(fc, updateCols, updateFkCols, where, orderByColumns, updateConfigCount)
			configs = append(configs, updateConfig)
		}
	}

	// Handle any columns that do not appear in any constraints.
	for col, stillNeeded := range remainingColumns {
		if stillNeeded {
			insertConfig.insertColumns = append(insertConfig.insertColumns, col)
		}
	}

	// Insert config should be at the front, then any update configs follow.
	configs = append([]*RunConfig{insertConfig}, configs...)
	return configs
}

func (b *runConfigBuilder) buildUpdateConfig(
	fc *sqlmanager_shared.ForeignConstraint,
	updateCols []string,
	updateFkCols []string,
	where *string,
	orderByColumns []string,
	count int,
) *RunConfig {
	// Add foreign key table as a dependency
	dependsOn := []*DependsOn{
		{
			Table:   fc.ForeignKey.Table,
			Columns: updateFkCols,
		},
	}

	// If the foreign key table is not the same as the table, we need to add INSERT config as a dependency
	if fc.ForeignKey.Table != b.table.String() {
		dependsOn = append(dependsOn, &DependsOn{
			Table:   b.table.String(),
			Columns: b.primaryKeys,
		})
	}

	selectColumns := slices.Concat(b.primaryKeys, updateCols)
	return &RunConfig{
		id:             fmt.Sprintf("%s.%s.%d", b.table, RunTypeUpdate, count),
		table:          b.table,
		runType:        RunTypeUpdate,
		selectColumns:  selectColumns,
		insertColumns:  updateCols,
		primaryKeys:    b.primaryKeys,
		whereClause:    where,
		orderByColumns: orderByColumns,
		dependsOn:      dependsOn,
		subsetPaths:    b.subsetPaths,
	}
}

func (b *runConfigBuilder) getDependsOn() []*DependsOn {
	dependsOn := []*DependsOn{}
	for _, fk := range b.foreignKeys {
		dependsOn = append(dependsOn, &DependsOn{
			Table:   fk.ForeignKey.Table,
			Columns: fk.ForeignKey.Columns,
		})
	}
	return dependsOn
}

// getOrderByColumns returns order by columns for a table, prioritizing primary keys,
// then unique constraints, then unique indexes, and finally falling back to sorted select columns.
func (b *runConfigBuilder) getOrderByColumns(selectColumns []string) []string {
	if len(b.primaryKeys) > 0 {
		return b.primaryKeys
	}

	if len(b.uniqueConstraints) > 0 {
		return b.uniqueConstraints[0]
	}

	if len(b.uniqueIndexes) > 0 {
		return b.uniqueIndexes[0]
	}

	sc := slices.Clone(selectColumns)
	slices.Sort(sc)
	return sc
}
