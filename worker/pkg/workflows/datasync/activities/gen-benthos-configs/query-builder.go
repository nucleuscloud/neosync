package genbenthosconfigs_activity

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
	"github.com/xwb1989/sqlparser"

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
	JoinType       joinType
	JoinTable      string
	BaseTable      string
	Alias          *string
	JoinColumnsMap map[string]string // map of joinColumn to baseColumn
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
		joinCondition := goqu.Ex{}
		for joinCol, baseCol := range j.JoinColumnsMap {
			joinCondition[buildSqlIdentifier(j.JoinTable, joinCol)] = goqu.I(buildSqlIdentifier(j.BaseTable, baseCol))
		}
		if j.JoinType == innerJoin {
			joinTable := goqu.I(j.JoinTable)
			query = query.InnerJoin(
				joinTable,
				goqu.On(joinCondition),
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
	var builder goqu.DialectWrapper
	if driver == sql_manager.MysqlDriver {
		opts := goqu.DefaultDialectOptions()
		opts.QuoteRune = '`'
		opts.SupportsWithCTERecursive = true
		opts.SupportsWithCTE = true
		dialectName := "custom-mysql-dialect"
		goqu.RegisterDialect(dialectName, opts)
		builder = goqu.Dialect(dialectName)
	} else {
		builder = goqu.Dialect(driver)
	}

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
		joinCondition := goqu.Ex{}
		for joinCol, baseCol := range j.JoinColumnsMap {
			joinCondition[buildSqlIdentifier(j.JoinTable, joinCol)] = goqu.I(buildSqlIdentifier(j.BaseTable, baseCol))
		}
		if j.JoinType == innerJoin {
			table := goqu.I(j.JoinTable)
			initialSelect = initialSelect.InnerJoin(
				table,
				goqu.On(joinCondition),
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
	tableDependencies map[string][]*sql_manager.ForeignConstraint,
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
				continue
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

			jsonF, _ := json.MarshalIndent(subsetCfg, "", " ")
			fmt.Printf("\n subsetCfg: %s \n", string(jsonF))

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
	Joins                             []*sqlJoin
	WhereClauses                      []string
	SelfReferencingCircularDependency *selfReferencingCircularDependency
}

func buildTableSubsetQueryConfig(
	table string,
	pathToRoot []string,
	dependencyMap map[string][]*sql_manager.ForeignConstraint,
	tableWhereMap map[string]string,
) *subsetQueryConfig {
	fmt.Println("------- buildTableSubsetQueryConfig")
	fmt.Printf("table: %s \n", table)

	reversePathToRoot := reverseSlice(pathToRoot)
	jsonF, _ := json.MarshalIndent(reversePathToRoot, "", " ")
	fmt.Printf("\n reversePathToRoot: %s \n", string(jsonF))

	joins := []*sqlJoin{}
	whereClauses := []string{}
	subsetTables := []string{} // keeps track of tables that are being subset
	fks := dependencyMap[table]
	jsonF, _ = json.MarshalIndent(fks, "", " ")
	fmt.Printf("\n fks: %s \n", string(jsonF))

	fksTableMap := buildFkTableMapOld(fks)
	jsonF, _ = json.MarshalIndent(fksTableMap, "", " ")
	fmt.Printf("\n fksTableMap: %s \n", string(jsonF))
	jsonF, _ = json.MarshalIndent(dependencyMap, "", " ")
	fmt.Printf("\n dependencyMap: %s \n", string(jsonF))
	for _, t := range pathToRoot {
		if t == table {
			continue
		}
		dependencies := dependencyMap[t]

		wc, ok := tableWhereMap[t]
		if ok {
			whereClauses = append(whereClauses, wc)
		}

		// pathToRoot: [
		// 	"public.expense_report",
		// 	"public.department",
		// 	"public.company"
		// 	]

		// SELECT er.*
		// FROM public.expense_report er
		// JOIN public.department dsrc ON tsrc.department_id = dsrc.id
		// JOIN public.department ddest ON tdest.department_id = ddest.id
		// join public.company c on c.id = dsr.company_id
		// join public.company cc on c.id == ddest.company_id
		// WHERE dsrc.company_id = 1 AND ddest.company_id = 1;

		if len(whereClauses) > 0 {
			subsetTables = append(subsetTables, t)
			// add joins for parent tables up to first subsetted table
			depsTableMap := buildFkTableMapOld(dependencies)
			jsonF, _ := json.MarshalIndent(depsTableMap, "", " ")
			fmt.Printf("\n depsTableMap: %s \n", string(jsonF))
			for fkTable, colsMap := range depsTableMap {
				fmt.Println("tables in depsTableMap")
				fmt.Println(colsMap)
				if t != fkTable && slices.Contains(subsetTables, fkTable) {
					joins = append(joins, &sqlJoin{
						JoinType:  innerJoin,
						BaseTable: t,
						JoinTable: fkTable,
						// JoinColumnsMap: colsMap,
						JoinColumnsMap: map[string]string{"id": "id"},
					})
				}
			}
			// add join for current table
			for fkTable, colsMap := range fksTableMap {
				fmt.Println("current table fksTableMap")
				fmt.Println(colsMap)
				if t == fkTable {
					joins = append(joins, &sqlJoin{
						JoinType:  innerJoin,
						BaseTable: table,
						JoinTable: fkTable,
						// JoinColumnsMap: colMap,
						JoinColumnsMap: map[string]string{"id": "id"},
					})
				}
			}
		}
	}
	// reverse joins so they are constructed in correct order
	// reverseSlice(joins)

	return &subsetQueryConfig{
		Joins:        joins,
		WhereClauses: whereClauses,
	}
}

// func getDoubleReference() map[string][]string {

// }
func buildFkTableMapOld(fks []*sql_manager.ForeignConstraint) map[string]map[string]string {
	fksTableMap := map[string]map[string]string{} // map of fk table to map of fk column to base table column
	for _, c := range fks {
		if _, exists := fksTableMap[c.ForeignKey.Table]; !exists {
			fksTableMap[c.ForeignKey.Table] = map[string]string{}
		}
		fksTableMap[c.ForeignKey.Table][c.ForeignKey.Column] = c.Column
	}
	return fksTableMap
}

func buildFkTableMap(fks []*sql_manager.ForeignConstraint) map[string]map[string][]string {
	fksTableMap := map[string]map[string][]string{} // map of fk table to map of fk column to base table column
	for _, c := range fks {
		if _, exists := fksTableMap[c.ForeignKey.Table]; !exists {
			fksTableMap[c.ForeignKey.Table] = map[string][]string{}
		}
		fksTableMap[c.ForeignKey.Table][c.ForeignKey.Column] = append(fksTableMap[c.ForeignKey.Table][c.ForeignKey.Column], c.Column)
	}
	return fksTableMap
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
	case sql_manager.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, schema, table)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sql_manager.PostgresDriver:
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
	dpMap := map[string]map[string]struct{}{}

	for _, cfg := range runConfigs {
		if _, exists := dpMap[cfg.Table]; !exists {
			dpMap[cfg.Table] = map[string]struct{}{}
		}
		for _, dep := range cfg.DependsOn {
			if _, exists := dpMap[dep.Table]; !exists {
				dpMap[dep.Table] = map[string]struct{}{}
			}
			dpMap[dep.Table][cfg.Table] = struct{}{}
		}
	}
	tableDependencyMap := map[string][]string{}
	for table, fkTables := range dpMap {
		for t := range fkTables {
			tableDependencyMap[table] = append(tableDependencyMap[table], t)
		}
	}

	return tableDependencyMap
}

type selfReferencingCircularDependency struct {
	PrimaryKeyColumn  string
	ForeignKeyColumns []string
}

func getSelfReferencingColumns(table string, tc []*sql_manager.ForeignConstraint) *selfReferencingCircularDependency {
	if tc == nil {
		return nil
	}
	fkCols := []string{}
	var primaryKeyCol string
	for _, fc := range tc {
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

func reverseSlice[T any](slice []T) []T {
	newSlice := make([]T, len(slice))
	copy(newSlice, slice)

	for i, j := 0, len(newSlice)-1; i < j; i, j = i+1, j-1 {
		newSlice[i], newSlice[j] = newSlice[j], newSlice[i]
	}

	return newSlice
}

func qualifyPostgresWhereColumnNames(sql, schema, table string) (string, error) {
	tree, err := pgquery.Parse(sql)
	if err != nil {
		return "", err
	}

	for _, stmt := range tree.GetStmts() {
		selectStmt := stmt.GetStmt().GetSelectStmt()

		if selectStmt.WhereClause != nil {
			updatePostgresExpr(schema, table, selectStmt.WhereClause)
		}
	}
	updatedSql, err := pgquery.Deparse(tree)
	if err != nil {
		return "", err
	}
	return updatedSql, nil
}

func updatePostgresExpr(schema, table string, node *pg_query.Node) {
	switch expr := node.Node.(type) {
	case *pg_query.Node_SubLink:
		updatePostgresExpr(schema, table, node.GetSubLink().GetTestexpr())
	case *pg_query.Node_BoolExpr:
		for _, arg := range expr.BoolExpr.GetArgs() {
			updatePostgresExpr(schema, table, arg)
		}
	case *pg_query.Node_AExpr:
		updatePostgresExpr(schema, table, expr.AExpr.GetLexpr())
		updatePostgresExpr(schema, table, expr.AExpr.Rexpr)
	case *pg_query.Node_ColumnDef:
	case *pg_query.Node_ColumnRef:
		col := node.GetColumnRef()
		isQualified := false
		var colName *string
		// find col name and check if already has schema + table name
		for _, f := range col.Fields {
			val := f.GetString_().GetSval()
			if val == schema {
				continue
			}
			if val == table {
				isQualified = true
				break
			}
			colName = &val
		}
		if !isQualified && colName != nil && *colName != "" {
			col.Fields = []*pg_query.Node{
				pg_query.MakeStrNode(schema),
				pg_query.MakeStrNode(table),
				pg_query.MakeStrNode(*colName),
			}
		}
	}
}

func qualifyMysqlWhereColumnNames(sql, schema, table string) (string, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", err
	}

	switch stmt := stmt.(type) { //nolint:gocritic
	case *sqlparser.Select:
		err = sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) { //nolint:gocritic
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
		if err != nil {
			return "", err
		}
	}

	return sqlparser.String(stmt), nil
}
