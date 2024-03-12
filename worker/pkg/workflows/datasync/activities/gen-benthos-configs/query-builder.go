package genbenthosconfigs_activity

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/xwb1989/sqlparser"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

const (
	innerJoin     JoinType      = "INNER"
	whereOperator WhereOperator = "EQUAL"
)

type WhereOperator string
type JoinType string

type SqlJoin struct {
	JoinType   JoinType
	JoinTable  string
	JoinColumn string
	BaseTable  string
	BaseColumn string
}

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

func buildSelectJoinQuery(
	driver, schema, table string,
	columns []string,
	joins []*SqlJoin,
	whereClauses []string,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.S(schema).Table(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = col
	}
	query := builder.From(sqltable).Select(selectColumns...)
	// joins
	for _, j := range joins {
		if j == nil {
			continue
		}
		if j.JoinType == innerJoin {
			joinTable := goqu.I(j.JoinTable)
			query = query.InnerJoin(
				joinTable,
				goqu.On(goqu.Ex{fmt.Sprintf("%s.%s", j.JoinTable, j.JoinColumn): goqu.I(fmt.Sprintf("%s.%s", j.BaseTable, j.BaseColumn))}),
			)
		}
	}
	// where
	goquWhere := []exp.Expression{}
	for _, w := range whereClauses {
		goquWhere = append(goquWhere, goqu.L(w))
	}
	query = query.Where(goqu.And(goquWhere...))

	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", sql), nil
}

func buildSelectUnionQuery(
	driver, schema, table string,
	columns []string,
	selfReferenceCol string,
	primaryKeyCol string,
	joins []*SqlJoin,
	whereClauses []string,
) (string, error) {
	table1Alias := "t1"
	table2Alias := "t2"
	builder := goqu.Dialect(driver)
	sqltable := goqu.S(schema).Table(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = col
	}
	firstSelect := builder.From(sqltable).Select(selectColumns...)

	// aliased columns
	unionColumns := make([]any, len(columns))
	for i, col := range columns {
		unionColumns[i] = fmt.Sprintf("%s.%s", table2Alias, col)
	}
	secondSelect := builder.From(sqltable).As(table1Alias).Select(unionColumns...).InnerJoin(
		sqltable.As(table2Alias),
		goqu.On(goqu.Ex{fmt.Sprintf("%s.%s", table1Alias, selfReferenceCol): goqu.I(fmt.Sprintf("%s.%s", table2Alias, primaryKeyCol))}),
	)

	// joins
	for _, j := range joins {
		if j == nil {
			continue
		}
		if j.JoinType == innerJoin {
			table := goqu.I(j.JoinTable)
			joinTable := fmt.Sprintf("%s.%s", j.JoinTable, j.JoinColumn)
			baseTable := fmt.Sprintf("%s.%s", j.BaseTable, j.BaseColumn)

			firstSelect = firstSelect.InnerJoin(
				table,
				goqu.On(goqu.Ex{joinTable: goqu.I(baseTable)}),
			)
			secondSelect = secondSelect.InnerJoin(
				table,
				goqu.On(goqu.Ex{joinTable: goqu.I(baseTable)}),
			)
		}
	}

	// where
	goquWhere := []exp.Expression{}
	for _, w := range whereClauses {
		goquWhere = append(goquWhere, goqu.L(w))
	}
	firstSelect = firstSelect.Where(goqu.And(goquWhere...))
	secondSelect.Where(goqu.And(goquWhere...))

	// union
	unionQuery := firstSelect.Union(secondSelect)
	sql, _, err := unionQuery.ToSQL()
	if err != nil {
		return "", err
	}
	return sql, nil
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
	// need to check for circular dependencies
	// need to check all tables processed
	// self referencing circular dependencies need union all
	// multi table circular dependencies need to be subset and run in same order
	// for each circular dependency choose entry point and add to root
	for _, rootTable := range roots {
		deps, hasDeps := pksMap[rootTable]
		if !hasDeps || len(deps) == 0 {
			continue
		}
		path := BFS(pksMap, rootTable)
		fmt.Printf("%+v\n", path)
		whereClauses := []string{}
		joins := make([]*SqlJoin, len(path))
		for idx, table := range path {
			tableMapping := groupedMappings[table]
			var where *string
			tableOpt := sourceTableOpts[table]
			if tableOpt != nil && tableOpt.WhereClause != nil {
				where = tableOpt.WhereClause
			}

			// root table or no subset use standard select
			if table == rootTable || len(whereClauses) == 0 {
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

				if where != nil && *where != "" {
					updatedWhere, err := qualifyWhereColumnNames(*where, tableMapping.Schema, tableMapping.Table)
					if err != nil {
						return nil, err
					}
					whereClauses = append(whereClauses, updatedWhere)
				}

				continue
			}
			// check if table already in query map
			_, seen := queryMap[table]
			if seen {
				// continue
				break
			}
			fks, ok := tableDependencies[table]
			if !ok {
				return nil, fmt.Errorf("no foreign keys found for table %s", table)
			}
			parentTable := path[idx-1]
			for _, fk := range fks.Constraints {
				if fk.ForeignKey.Table == parentTable {
					joins[len(path)-idx] = &SqlJoin{
						JoinType:   innerJoin,
						BaseTable:  table,
						BaseColumn: fk.Column,
						JoinTable:  fk.ForeignKey.Table,
						JoinColumn: fk.ForeignKey.Column,
					}

				}
			}
			// subsetting exists so need to join all child tables

			if where != nil && *where != "" {
				updatedWhere, err := qualifyWhereColumnNames(*where, tableMapping.Schema, tableMapping.Table)
				if err != nil {
					return nil, err
				}
				whereClauses = append(whereClauses, updatedWhere)
			}
			query, err := buildSelectJoinQuery(
				driver,
				tableMapping.Schema,
				tableMapping.Table,
				buildPlainColumns(tableMapping.Mappings),
				joins,
				whereClauses,
			)
			if err != nil {
				return nil, fmt.Errorf("unable to build select query: %w", err)
			}
			queryMap[table] = query
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

func qualifyWhereColumnNames(where, schema, table string) (string, error) {
	sqlSelect := fmt.Sprintf("select * from %s.%s where ", schema, table)
	sql := fmt.Sprintf("%s%s", sqlSelect, where)
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", err
	}
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) {
			case *sqlparser.ComparisonExpr:
				if col, ok := node.Left.(*sqlparser.ColName); ok {
					if col.Qualifier.IsEmpty() {
						col.Qualifier.Qualifier = sqlparser.NewTableIdent(schema)
						col.Qualifier.Name = sqlparser.NewTableIdent(table)
					}
				}
				return false, nil
			}
			return true, nil
		}, stmt)
	}
	updatedSql := sqlparser.String(stmt)
	updatedSql = strings.Replace(updatedSql, sqlSelect, "", 1)
	return updatedSql, nil
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
