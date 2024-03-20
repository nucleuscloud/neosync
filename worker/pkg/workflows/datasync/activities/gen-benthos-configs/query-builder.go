package genbenthosconfigs_activity

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	dbschemas "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

const (
	innerJoin joinType = "INNER"
)

type joinType string

type sqlJoin struct {
	JoinType   joinType
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
	joins []*sqlJoin,
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
	joins []*sqlJoin,
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
	// map of table -> where clause
	tableWhereMap := map[string]string{}
	for t, opts := range sourceTableOpts {
		mappings, ok := groupedMappings[t]
		where := getWhereFromTableOpts(opts)
		if ok && where != nil && *where != "" {
			qualifiedWhere, err := qualifyWhereColumnNames(driver, *where, mappings.Schema, mappings.Table)
			if err != nil {
				return nil, err
			}
			tableWhereMap[t] = qualifiedWhere
		}
	}

	if !subsetByForeignKeyConstraints || len(tableDependencies) == 0 || len(tableWhereMap) == 0 {
		queryMap, err := buildQueryMapNoSubsetConstraints(driver, groupedMappings, sourceTableOpts)
		if err != nil {
			return nil, err
		}
		return queryMap, nil
	}

	queryMap := map[string]string{}
	pksMap := getPrimaryToForeignTableMapFromRunConfigs(dependencyConfigs)
	rootsWithsubsetMaps := []string{}
	rootsNoSubsets := []string{}

	dependencyMap := map[string][]*tabledependency.RunConfig{}
	for _, cfg := range dependencyConfigs {
		if len(cfg.DependsOn) == 0 {
			_, ok := tableWhereMap[cfg.Table]
			if ok {
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

	for _, rootTable := range roots {
		bfsRes := getBfsPathMap(pksMap, rootTable)
		path := bfsRes.Path
		tablePathMap := bfsRes.NodePathMap

		for _, table := range path {
			// check if query already created
			_, seen := queryMap[table]
			if seen {
				break
			}
			tableMapping := groupedMappings[table]
			var where *string
			tableWhere, whereOk := tableWhereMap[table]
			if whereOk {
				where = &tableWhere
			}
			fks := tableDependencies[table]
			runConfigs := dependencyMap[table]
			selfRefCircularDep := getSelfReferencingColumns(table, fks)

			pathToRoot := tablePathMap[table]
			subsetCfg := buildTableSubsetQueryConfig(table, pathToRoot, tableDependencies, tableWhereMap)
			joins := subsetCfg.Joins
			whereClauses := subsetCfg.WhereClauses

			// root table or no subset use standard select
			if table == rootTable || len(whereClauses) == 0 {
				if where != nil && *where != "" {
					whereClauses = append(whereClauses, *where)
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
					// standard select
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

			if where != nil && *where != "" {
				whereClauses = append(whereClauses, *where)
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
				// select with joins
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

func buildQueryMapNoSubsetConstraints(
	driver string,
	groupedMappings map[string]*tableMapping,
	sourceTableOpts map[string]*sqlSourceTableOptions,
) (map[string]string, error) {
	queryMap := map[string]string{}
	for table, tableMapping := range groupedMappings {
		if shared.AreAllColsNull(tableMapping.Mappings) {
			// skipping table as no columns are mapped
			continue
		}

		tableOpt := sourceTableOpts[table]
		where := getWhereFromTableOpts(tableOpt)

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

type subsetQueryConfig struct {
	Joins        []*sqlJoin
	WhereClauses []string
}

func buildTableSubsetQueryConfig(
	table string,
	pathToRoot []string,
	dependencyMap map[string]*dbschemas.TableConstraints,
	tableWhereMap map[string]string,

) *subsetQueryConfig {
	joins := []*sqlJoin{}
	whereClauses := []string{}
	subsetTables := []string{} // keeps track of tables that are being subset
	fks := dependencyMap[table]

	for _, t := range pathToRoot {
		if t == table {
			continue
		}
		dependencies := dependencyMap[t]
		wc, ok := tableWhereMap[t]
		if ok {
			whereClauses = append(whereClauses, wc)
		}

		if len(whereClauses) > 0 {
			subsetTables = append(subsetTables, t)
			// add joins for parent tables up to first subsetted table
			if dependencies != nil {
				for _, c := range dependencies.Constraints {
					if t != c.ForeignKey.Table && slices.Contains(subsetTables, c.ForeignKey.Table) {
						joins = append(joins, &sqlJoin{
							JoinType:   innerJoin,
							BaseTable:  t,
							BaseColumn: c.Column,
							JoinTable:  c.ForeignKey.Table,
							JoinColumn: c.ForeignKey.Column,
						})
					}
				}
			}
			// add join for current table
			if fks != nil {
				for _, c := range fks.Constraints {
					if t == c.ForeignKey.Table {
						joins = append(joins, &sqlJoin{
							JoinType:   innerJoin,
							BaseTable:  table,
							BaseColumn: c.Column,
							JoinTable:  c.ForeignKey.Table,
							JoinColumn: c.ForeignKey.Column,
						})
					}
				}
			}
		}
	}
	// reverse joins so they are constructed in correct order
	reverseSlice(joins)
	return &subsetQueryConfig{
		Joins:        joins,
		WhereClauses: whereClauses,
	}
}

type bfsPaths struct {
	Path        []string
	NodePathMap map[string][]string
}

func getBfsPathMap(graph map[string][]string, start string) *bfsPaths {
	path := []string{}
	var queue []string
	visited := make(map[string]bool)
	// path to root for each node
	nodePathMap := make(map[string][]string)

	queue = append(queue, start)
	visited[start] = true
	// initialize path with itself
	nodePathMap[start] = []string{start}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		path = append(path, node)

		for _, adjacent := range graph[node] {
			if visited[adjacent] {
				continue
			}
			queue = append(queue, adjacent)
			visited[adjacent] = true
			// copy path and append adjacent node
			currentPath := make([]string, len(nodePathMap[node]))
			copy(currentPath, nodePathMap[node])
			currentPath = append(currentPath, adjacent)
			nodePathMap[adjacent] = currentPath
		}
	}

	return &bfsPaths{
		Path:        path,
		NodePathMap: nodePathMap,
	}
}

func qualifyWhereColumnNames(driver, where, schema, table string) (string, error) {
	sqlSelect := fmt.Sprintf("select * from %s where ", buildSqlIdentifier(schema, table))
	sql := fmt.Sprintf("%s%s", sqlSelect, where)
	var updatedSql string
	switch driver {
	case mysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, schema, table)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case postgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, schema, table)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	default:
		return "", errors.New("unsupported sql driver type")
	}
	index := strings.Index(strings.ToLower(updatedSql), "where")
	if index == -1 {
		// "where" not found
		return "", fmt.Errorf("unable to qualify where column names")
	}
	startIndex := index + len("where")

	return strings.TrimSpace(updatedSql[startIndex:]), nil
}

func getPrimaryToForeignTableMapFromRunConfigs(runConfigs []*tabledependency.RunConfig) map[string][]string {
	dpMap := make(map[string][]string)

	for _, cfg := range runConfigs {
		if _, exists := dpMap[cfg.Table]; !exists {
			dpMap[cfg.Table] = []string{}
		}
		for _, dep := range cfg.DependsOn {
			dpMap[dep.Table] = append(dpMap[dep.Table], cfg.Table)
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

func getWhereFromTableOpts(tableOpts *sqlSourceTableOptions) *string {
	var where *string
	if tableOpts != nil && tableOpts.WhereClause != nil {
		where = tableOpts.WhereClause
	}
	return where
}

func reverseSlice[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
