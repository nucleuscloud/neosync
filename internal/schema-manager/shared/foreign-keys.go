package schemamanager_shared

import (
	"fmt"
	"log/slog"
	"sort"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

func BuildOrderedForeignKeyConstraintsToDrop(
	logger *slog.Logger,
	diff *SchemaDifferences,
) []*sqlmanager_shared.ForeignKeyConstraint {
	allFks := []*sqlmanager_shared.ForeignKeyConstraint{}
	if diff.ExistsInDestination != nil {
		allFks = append(allFks, diff.ExistsInDestination.ForeignKeyConstraints...)
	}
	if diff.ExistsInBoth != nil && diff.ExistsInBoth.Different != nil {
		allFks = append(allFks, diff.ExistsInBoth.Different.ForeignKeyConstraints...)
	}
	if len(allFks) == 0 {
		return nil
	}

	// Separate self-referencing constraints (they form a cycle of length 1)
	// from the "normal" constraints that can be placed into a DAG.
	var selfRefFKs, normalFKs []*sqlmanager_shared.ForeignKeyConstraint
	for _, fk := range allFks {
		if fk.ReferencedSchema == fk.ReferencingSchema &&
			fk.ReferencedTable == fk.ReferencingTable {
			selfRefFKs = append(selfRefFKs, fk)
		} else {
			normalFKs = append(normalFKs, fk)
		}
	}

	// A small helper to build a unique "schema.table" key
	key := func(schema, table string) string {
		return fmt.Sprintf("%s.%s", schema, table)
	}

	// Build adjacency list for topological sort:
	// We treat "ReferencedTable -> ReferencingTable" as an edge:
	//    parent --> child
	// so we can drop the child's FK after the parent in topological order.
	parentToChildren := make(map[string]map[string]bool)
	inDegree := make(map[string]int) // track how many parents each table has

	// Initialize adjacency map for any table we see
	addTableIfMissing := func(tableKey string) {
		if _, ok := parentToChildren[tableKey]; !ok {
			parentToChildren[tableKey] = make(map[string]bool)
		}
		if _, ok := inDegree[tableKey]; !ok {
			inDegree[tableKey] = 0
		}
	}

	// Populate graph edges
	for _, fk := range normalFKs {
		p := key(fk.ReferencedSchema, fk.ReferencedTable)   // "parent"
		c := key(fk.ReferencingSchema, fk.ReferencingTable) // "child"
		addTableIfMissing(p)
		addTableIfMissing(c)
		// Only add edge once
		if !parentToChildren[p][c] {
			parentToChildren[p][c] = true
			inDegree[c]++
		}
	}

	// Kahn’s Algorithm:
	// 1. Find all tables with inDegree = 0 -> put them in the queue
	// 2. Pop one, append it to topological order, decrement inDegree of its children
	// 3. When a child's inDegree hits 0, enqueue it
	// 4. Repeat until queue is empty
	var queue []string
	for t, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, t)
		}
	}

	var topoOrder []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		topoOrder = append(topoOrder, current)

		for child := range parentToChildren[current] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	// If topological order doesn't include all tables, we have a cycle among the remainder.
	// Either forcibly drop them or return an error. Here, we forcibly handle them (like your code).
	hadCycle := (len(topoOrder) < len(inDegree))
	if hadCycle {
		logger.Warn(
			"Cycle detected among foreign keys. Forcibly dropping all remaining constraints.",
		)
	}

	// If no cycle, we can produce a stable drop order for "normal" FKs:
	// We'll drop constraints in the reverse of the topological order of the *ReferencingTable*.
	// Because a child is guaranteed to appear after its parent in topoOrder,
	// dropping from child->parent means we won't block on references.

	// Map each table to its topological index; if there's a cycle, fallback to index = 0
	tableIndex := make(map[string]int)
	for i, tbl := range topoOrder {
		tableIndex[tbl] = i
	}

	// Sort the normalFKs by the *descending* index of the referencing table
	// so that "leaf" tables in the DAG appear first in the final slice.
	sort.Slice(normalFKs, func(i, j int) bool {
		fi := key(normalFKs[i].ReferencingSchema, normalFKs[i].ReferencingTable)
		fj := key(normalFKs[j].ReferencingSchema, normalFKs[j].ReferencingTable)
		return tableIndex[fi] > tableIndex[fj]
	})

	// If there's a cycle, everything might be index=0, so the stable ordering is somewhat moot—
	// but we'll just drop them all anyway.
	// Combine self-ref constraints first (they can be dropped at any time),
	// then the normal constraints in the sorted order.
	var finalOrdered []*sqlmanager_shared.ForeignKeyConstraint

	// Self-referencing FKs first
	finalOrdered = append(finalOrdered, selfRefFKs...)
	// Then normal FKs
	finalOrdered = append(finalOrdered, normalFKs...)

	return finalOrdered
}
