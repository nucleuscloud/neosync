package genbenthosconfigs_activity

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

func buildSelectQuery(
	driver, schema, table string,
	columns []string,
	whereClause *string,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.S(schema).Table(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = col
	}
	query := builder.From(sqltable).Select(selectColumns...)

	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s;", sql), nil
}

type SqlJoin struct {
}

func buildSelectJoinQuery(
	driver, schema, table string,
	columns []string,
	joins []*SqlJoin,
	whereStr *string,
) (string, error) {

}

// returns map of schema.table -> select query
func buildSelectQueryMap(
	driver string,
	groupedMappings map[string]*tableMapping,
	sourceTableOpts map[string]*sqlSourceTableOptions,
	tableDependencies map[string]*dbschemas.TableConstraints,
	subsetByForeignKeyConstraints bool,
) (map[string]string, error) {
	queryMap := map[string]string{}
	if !subsetByForeignKeyConstraints || len(tableDependencies) == 0 {
		for table, tableMapping := range groupedMappings {
			if shared.AreAllColsNull(tableMapping.Mappings) {
				// skipping table as no columns are mapped
				continue
			}

			var where *string
			tableOpt := sourceTableOpts[table]
			if tableOpt != nil && tableOpt.WhereClause != nil {
				where = tableOpt.WhereClause
			}

			query, err := buildSelectQuery(
				driver,
				tableMapping.Schema,
				tableMapping.Table,
				buildPlainColumns(tableMapping.Mappings),
				where,
			)
			if err != nil {
				return nil, fmt.Errorf("unable to build select query: %w", err)
			}
			queryMap[table] = query
		}
		return queryMap, nil
	}
	uniqueTables := map[string]struct{}{}
	for table := range groupedMappings {
		uniqueTables[table] = struct{}{}
	}
	fksMap := getForeignToPrimaryTableMap(tableDependencies, uniqueTables)
	pksMap := getPrimaryToForeignTableMap(tableDependencies, uniqueTables)
	roots := []string{}

	// handle roots first
	for table, deps := range fksMap {
		if len(deps) == 0 {
			roots = append(roots, table)
		}
	}
	for _, rootTable := range roots {
		deps, hasDeps := pksMap[rootTable]
		if !hasDeps || len(deps) == 0 {
			continue
		}
		fmt.Println("--------------")
		path := BFS(pksMap, rootTable)
		for _, table := range path {
			if table == rootTable {
				tableMapping := groupedMappings[table]
				var where *string
				tableOpt := sourceTableOpts[table]
				if tableOpt != nil && tableOpt.WhereClause != nil {
					where = tableOpt.WhereClause
				}
				query, err := buildSelectQuery(
					driver,
					tableMapping.Schema,
					tableMapping.Table,
					buildPlainColumns(tableMapping.Mappings),
					where,
				)
				if err != nil {
					return nil, fmt.Errorf("unable to build select query: %w", err)
				}
				queryMap[table] = query
			}
			// check if table already in query map
			_, seen := queryMap[table]
			if seen {
				continue
			}

		}

	}
	return queryMap, nil
}

/*
NOTES
1. use find replace in where to add table.column syntax

how to find shortest path?
1. find root tables. tables with no foreign keys. what if no root table?
2. for each root table write query
*/

func BFS(graph map[string][]string, start string) []string {
	var queue []string
	path := []string{}
	visited := make(map[string]bool)

	queue = append(queue, start)
	visited[start] = true

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		path = append(path, node)

		for _, adjacent := range graph[node] {
			if !visited[adjacent] {
				queue = append(queue, adjacent)
				visited[adjacent] = true
			}
		}
	}
	return path
}

// todo move to shared
func getForeignToPrimaryTableMap(td map[string]*dbschemas.TableConstraints, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table := range uniqueTables {
		_, dpOk := dpMap[table]
		if !dpOk {
			dpMap[table] = []string{}
		}
		constraints, ok := td[table]
		if !ok {
			continue
		}
		for _, dep := range constraints.Constraints {
			dpMap[table] = append(dpMap[table], dep.ForeignKey.Table)
		}
	}
	return dpMap
}

func getPrimaryToForeignTableMap(td map[string]*dbschemas.TableConstraints, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table := range uniqueTables {
		_, dpOk := dpMap[table]
		if !dpOk {
			dpMap[table] = []string{}
		}
		constraints, ok := td[table]
		if !ok {
			continue
		}
		for _, dep := range constraints.Constraints {
			dpMap[dep.ForeignKey.Table] = append(dpMap[dep.ForeignKey.Table], table)
		}
	}
	return dpMap
}
