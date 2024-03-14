package genbenthosconfigs_activity

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
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

	return formatSqlQuery(sql), nil
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
		selectColumns[i] = buildSqlIdentifier(schema, table, col)
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
				goqu.On(goqu.Ex{buildSqlIdentifier(j.JoinTable, j.JoinColumn): goqu.I(buildSqlIdentifier(j.BaseTable, j.BaseColumn))}),
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
	return formatSqlQuery(sql), nil
}

func buildSelectRecursiveQuery(
	driver, schema, table string,
	columns []string,
	foreignKeys []string,
	primaryKeyCol string,
	joins []*SqlJoin,
	whereClauses []string,
) (string, error) {
	recursiveCteAlias := "related"
	builder := goqu.Dialect(driver)
	sqltable := goqu.S(schema).Table(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = buildSqlIdentifier(schema, table, col)
	}
	selectQuery := builder.From(sqltable).Select(selectColumns...)

	initialSelect := selectQuery
	// joins
	for _, j := range joins {
		if j == nil {
			continue
		}
		if j.JoinType == innerJoin {
			table := goqu.I(j.JoinTable)
			joinTable := buildSqlIdentifier(j.JoinTable, j.JoinColumn)
			baseTable := buildSqlIdentifier(j.BaseTable, j.BaseColumn)

			initialSelect = initialSelect.InnerJoin(
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
	initialSelect = initialSelect.Where(goqu.And(goquWhere...))

	// inner join on foreign keys
	goquOnEx := []exp.Expression{}
	for _, fk := range foreignKeys {
		goquOnEx = append(goquOnEx, goqu.Ex{buildSqlIdentifier(schema, table, primaryKeyCol): goqu.I(buildSqlIdentifier(recursiveCteAlias, fk))})
	}
	recursiveSelect := selectQuery
	recursiveSelect = recursiveSelect.InnerJoin(goqu.I(recursiveCteAlias), goqu.On(goqu.Or(goquOnEx...)))

	// union
	unionQuery := initialSelect.Union(recursiveSelect)

	selectCols := make([]any, len(columns))
	for i, col := range columns {
		selectCols[i] = col
	}
	recursiveQuery := builder.From(goqu.T(recursiveCteAlias)).WithRecursive(recursiveCteAlias, unionQuery).SelectDistinct(selectCols...)
	sql, _, err := recursiveQuery.ToSQL()
	if err != nil {
		return "", err
	}
	return formatSqlQuery(sql), nil
}

// returns map of schema.table -> select query
func buildSelectQueryMap(
	driver string,
	groupedMappings map[string]*tableMapping,
	sourceTableOpts map[string]*sqlSourceTableOptions,
	tableDependencies map[string]*dbschemas.TableConstraints,
	dependencyConfigs []*tabledependency.RunConfig,
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

	pksMap := getPrimaryToForeignTableMapFromRunConfigs(dependencyConfigs)
	rootsWithsubsetMaps := []string{}
	rootsNoSubsets := []string{}

	dependencyMap := map[string][]*tabledependency.RunConfig{}
	for _, cfg := range dependencyConfigs {
		if len(cfg.DependsOn) == 0 {
			where := getWhereFromTableOpts(cfg.Table, sourceTableOpts)
			if where != nil && *where != "" {
				rootsWithsubsetMaps = append(rootsWithsubsetMaps, cfg.Table)
			} else {
				rootsNoSubsets = append(rootsNoSubsets, cfg.Table)
			}
		}
		_, ok := dependencyMap[cfg.Table]
		if ok {
			dependencyMap[cfg.Table] = append(dependencyMap[cfg.Table], cfg)
		} else {
			dependencyMap[cfg.Table] = []*tabledependency.RunConfig{cfg}
		}
	}

	// process roots with subsets first
	roots := []string{}
	roots = append(roots, rootsWithsubsetMaps...)
	roots = append(roots, rootsNoSubsets...)
	fmt.Printf("pksMap: %+v\n", pksMap)

	fmt.Printf("%+v\n", roots)

	// need to check all tables processed
	for _, rootTable := range roots {
		path := BFS(pksMap, rootTable)
		fmt.Printf("%+v\n", path)

		whereClauses := []string{}
		joins := make([]*SqlJoin, len(path))
		for idx, table := range path {
			fmt.Printf("table: %s \n", table)
			jsonF, _ := json.MarshalIndent(queryMap, "", " ")
			fmt.Printf("\n %s \n", string(jsonF))
			// check if query already created
			_, seen := queryMap[table]
			fmt.Printf("seen: %t \n", seen)

			if seen {
				fmt.Println("break")
				break
			}
			fmt.Println("-------------")
			tableMapping := groupedMappings[table]
			var where *string
			tableOpt := sourceTableOpts[table]
			if tableOpt != nil && tableOpt.WhereClause != nil {
				where = tableOpt.WhereClause
			}
			fks := tableDependencies[table]
			runConfigs := dependencyMap[table]
			selfRefCircularDep := getSelfReferencingColumns(table, fks)

			// root table or no subset use standard select
			if table == rootTable || len(whereClauses) == 0 {
				if where != nil && *where != "" {
					updatedWhere, err := qualifyWhereColumnNames(*where, tableMapping.Schema, tableMapping.Table)
					if err != nil {
						return nil, err
					}
					whereClauses = append(whereClauses, updatedWhere)
				}
				if len(runConfigs) == 2 && selfRefCircularDep != nil {
					// self referencing circular dependency
					query, err := buildSelectRecursiveQuery(
						driver,
						tableMapping.Schema,
						tableMapping.Table,
						buildPlainColumns(tableMapping.Mappings),
						selfRefCircularDep.ForeignKeyColumns,
						selfRefCircularDep.PrimaryKeyColumn,
						joins,
						whereClauses,
					)
					if err != nil {
						return nil, fmt.Errorf("unable to build select query: %w", err)
					}
					queryMap[table] = query
				} else {
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

				continue
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

			if where != nil && *where != "" {
				updatedWhere, err := qualifyWhereColumnNames(*where, tableMapping.Schema, tableMapping.Table)
				if err != nil {
					return nil, err
				}
				whereClauses = append(whereClauses, updatedWhere)
			}

			if len(runConfigs) == 2 && selfRefCircularDep != nil {
				// self referencing circular dependency
				query, err := buildSelectRecursiveQuery(
					driver,
					tableMapping.Schema,
					tableMapping.Table,
					buildPlainColumns(tableMapping.Mappings),
					selfRefCircularDep.ForeignKeyColumns,
					selfRefCircularDep.PrimaryKeyColumn,
					joins,
					whereClauses,
				)
				if err != nil {
					return nil, fmt.Errorf("unable to build select query: %w", err)
				}
				queryMap[table] = query
			} else {
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
	}
	return queryMap, nil
}

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
	sqlSelect := fmt.Sprintf("select * from %s where ", buildSqlIdentifier(schema, table))
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
	fmt.Println(updatedSql)
	updatedSql = strings.Replace(updatedSql, sqlSelect, "", 1)
	return updatedSql, nil
}

func getPrimaryToForeignTableMapFromRunConfigs(runConfigs []*tabledependency.RunConfig) map[string][]string {
	dpMap := map[string][]string{}
	for _, cfg := range runConfigs {
		_, dpOk := dpMap[cfg.Table]
		if !dpOk {
			dpMap[cfg.Table] = []string{}
		}
		for _, dep := range cfg.DependsOn {
			_, dpOk := dpMap[dep.Table]
			if !dpOk {
				dpMap[dep.Table] = []string{cfg.Table}
			} else {
				dpMap[dep.Table] = append(dpMap[dep.Table], cfg.Table)
			}
		}
	}
	return dpMap
}

type selfReferencingCircularDependency struct {
	PrimaryKeyColumn  string
	ForeignKeyColumns []string
}

func getSelfReferencingColumns(table string, tc *dbschemas.TableConstraints) *selfReferencingCircularDependency {
	if tc == nil {
		return nil
	}
	fkCols := []string{}
	var primaryKeyCol string
	for _, fc := range tc.Constraints {
		if fc.ForeignKey.Table == table {
			fkCols = append(fkCols, fc.Column)
			primaryKeyCol = fc.ForeignKey.Column
		}
	}
	if len(fkCols) > 0 {
		return &selfReferencingCircularDependency{
			PrimaryKeyColumn:  primaryKeyCol,
			ForeignKeyColumns: fkCols,
		}
	}
	return nil
}

func formatSqlQuery(sql string) string {
	return fmt.Sprintf("%s;", sql)
}

func buildSqlIdentifier(identifiers ...string) string {
	return strings.Join(identifiers, ".")
}

func getWhereFromTableOpts(table string, sourceTableOpts map[string]*sqlSourceTableOptions) *string {
	var where *string
	tableOpt := sourceTableOpts[table]
	if tableOpt != nil && tableOpt.WhereClause != nil {
		where = tableOpt.WhereClause
	}
	return where
}
