package genbenthosconfigs_activity

import (
	"crypto/sha256"
	"errors"
	"fmt"
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
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
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

type SubsetReferenceKey struct {
	Table         string
	Columns       []string
	OriginalTable *string
}
type SubsetColumnConstraint struct {
	Columns     []string
	NotNullable []bool
	ForeignKey  *SubsetReferenceKey
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
		joinConditionTable := j.JoinTable
		if j.Alias != nil && *j.Alias != "" {
			joinConditionTable = *j.Alias
		}
		joinCondition := goqu.Ex{}
		for joinCol, baseCol := range j.JoinColumnsMap {
			joinCondition[buildSqlIdentifier(joinConditionTable, joinCol)] = goqu.I(buildSqlIdentifier(j.BaseTable, baseCol))
		}
		if j.JoinType == innerJoin {
			var joinTable exp.Expression
			joinTable = goqu.I(j.JoinTable)
			if j.Alias != nil && *j.Alias != "" {
				joinTable = goqu.I(j.JoinTable).As(*j.Alias)
			}
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

// composite keys are an AND
// many self referencing fks are an OR
func buildSelectRecursiveQuery(
	driver, schema, table string,
	columns []string,
	dependencies []*selfReferencingCircularDependency,
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
	goquOnOrEx := []exp.Expression{}
	for _, d := range dependencies {
		goquOnAndEx := []exp.Expression{}
		for _, fk := range d.ForeignKeyColumns {
			if len(fk) > 1 {
				for idx, col := range fk {
					goquOnAndEx = append(goquOnAndEx, goqu.Ex{buildSqlIdentifier(schema, table, d.PrimaryKeyColumns[idx]): goqu.I(buildSqlIdentifier(recursiveCteAlias, col))})
				}
				goquOnOrEx = append(goquOnOrEx, goqu.And(goquOnAndEx...))
			} else {
				goquOnOrEx = append(goquOnOrEx, goqu.Ex{buildSqlIdentifier(schema, table, d.PrimaryKeyColumns[0]): goqu.I(buildSqlIdentifier(recursiveCteAlias, fk[0]))})
			}
		}
	}

	recursiveSelect := selectQuery
	recursiveSelect = recursiveSelect.InnerJoin(goqu.I(recursiveCteAlias), goqu.On(goqu.Or(goquOnOrEx...)))

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
	tableDependencies map[string][]*sql_manager.ColumnConstraint,
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
	dependencyMap := map[string][]*tabledependency.RunConfig{}
	for _, cfg := range dependencyConfigs {
		_, ok := dependencyMap[cfg.Table]
		if ok {
			dependencyMap[cfg.Table] = append(dependencyMap[cfg.Table], cfg)
		} else {
			dependencyMap[cfg.Table] = []*tabledependency.RunConfig{cfg}
		}
	}

	subsetConfigs, err := buildTableSubsetQueryConfigs(driver, tableDependencies, tableWhereMap, dependencyMap)
	if err != nil {
		return nil, err
	}

	for table := range dependencyMap {
		tableMapping := groupedMappings[table]
		config := subsetConfigs[table]

		selectCols := []string{}
		for _, m := range tableMapping.Mappings {
			selectCols = append(selectCols, m.Column)
		}
		if len(config.SelfReferencingCircularDependency) != 0 {
			sql, err := buildSelectRecursiveQuery(
				driver,
				tableMapping.Schema,
				tableMapping.Table,
				buildPlainColumns(tableMapping.Mappings),
				config.SelfReferencingCircularDependency,
				config.Joins,
				config.WhereClauses,
			)
			if err != nil {
				return nil, fmt.Errorf("unable to build recursive select query: %w", err)
			}
			queryMap[table] = sql
		} else if len(config.Joins) == 0 {
			where := strings.Join(config.WhereClauses, " AND ")
			sql, err := buildSelectQuery(driver, tableMapping.Schema, tableMapping.Table, selectCols, &where)
			if err != nil {
				return nil, fmt.Errorf("unable to build select query: %w", err)
			}
			queryMap[table] = sql
		} else {
			sql, err := buildSelectJoinQuery(driver, tableMapping.Schema, tableMapping.Table, selectCols, config.Joins, config.WhereClauses)
			if err != nil {
				return nil, fmt.Errorf("unable to build select query with joins: %w", err)
			}
			queryMap[table] = sql
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

type tableSubset struct {
	Joins        []*sqlJoin
	WhereClauses []string
}

// recusively builds join for subset table
func buildSubsetJoins(table string, data map[string][]*SubsetColumnConstraint, whereClauses map[string]string, visited map[string]bool) *tableSubset {
	joins := []*sqlJoin{}
	wheres := []string{}

	if condition, exists := whereClauses[table]; exists {
		wheres = append(wheres, condition)
	}

	if columns, exists := data[table]; exists {
		for _, col := range columns {
			if col.ForeignKey.Table == "" && col.ForeignKey.Columns == nil {
				continue
			}
			// handle aliased table
			var alias *string
			joinTable := col.ForeignKey.Table
			if col.ForeignKey.OriginalTable != nil && *col.ForeignKey.OriginalTable != "" {
				alias = &col.ForeignKey.Table
				joinTable = *col.ForeignKey.OriginalTable
			}

			joinColMap := map[string]string{}
			for idx, c := range col.ForeignKey.Columns {
				joinColMap[c] = col.Columns[idx]
			}
			joins = append(joins, &sqlJoin{
				JoinType:       innerJoin,
				JoinTable:      joinTable,
				BaseTable:      table,
				Alias:          alias,
				JoinColumnsMap: joinColMap,
			})

			if !visited[col.ForeignKey.Table] {
				visited[col.ForeignKey.Table] = true
				sub := buildSubsetJoins(col.ForeignKey.Table, data, whereClauses, visited)
				joins = append(joins, sub.Joins...)
				wheres = append(wheres, sub.WhereClauses...)
			}
		}
	}
	return &tableSubset{
		Joins:        joins,
		WhereClauses: wheres,
	}
}

type subsetQueryConfig struct {
	Joins                             []*sqlJoin
	WhereClauses                      []string
	SelfReferencingCircularDependency []*selfReferencingCircularDependency
}

func buildTableSubsetQueryConfigs(driver string, tableConstraints map[string][]*sql_manager.ColumnConstraint, whereClauses map[string]string, runConfigMap map[string][]*tabledependency.RunConfig) (map[string]*subsetQueryConfig, error) {
	configs := map[string]*subsetQueryConfig{}

	filteredConstraints := filterForeignKeysWithSubset(runConfigMap, tableConstraints, whereClauses)
	subset, err := buildAliasReferences(driver, filteredConstraints, whereClauses)
	if err != nil {
		return nil, err
	}

	for table := range subset.ColumnConstraints {
		visited := map[string]bool{}
		tableSubset := buildSubsetJoins(table, subset.ColumnConstraints, subset.WhereClauses, visited)
		constraints := tableConstraints[table]
		selfRefCd := getSelfReferencingColumns(table, constraints)
		configs[table] = &subsetQueryConfig{
			Joins:                             tableSubset.Joins,
			WhereClauses:                      tableSubset.WhereClauses,
			SelfReferencingCircularDependency: selfRefCd,
		}
	}
	return configs, nil
}

type subsetConstraints struct {
	ColumnConstraints map[string][]*SubsetColumnConstraint
	WhereClauses      map[string]string
}

func buildAliasReferences(driver string, constraints map[string][]*sql_manager.ColumnConstraint, whereClauses map[string]string) (*subsetConstraints, error) {
	updatedConstraints := map[string][]*SubsetColumnConstraint{}
	aliasReference := map[string]string{} // alias name to table name
	updatedWheres := map[string]string{}
	seenTables := map[string]struct{}{}

	for table, where := range whereClauses {
		updatedWheres[table] = where
	}

	for table, colDefs := range constraints {
		if len(colDefs) == 0 {
			updatedConstraints[table] = []*SubsetColumnConstraint{}
			seenTables = map[string]struct{}{}
		} else {
			updatedConstraints[table] = processAliasConstraints(table, colDefs, updatedConstraints, aliasReference, seenTables)
		}
	}

	if err := updateAliasReferences(driver, updatedConstraints, aliasReference, updatedWheres); err != nil {
		return nil, err
	}

	return &subsetConstraints{
		ColumnConstraints: updatedConstraints,
		WhereClauses:      updatedWheres,
	}, nil
}

// creates alias table reference if there is a double reference
func processAliasConstraints(
	table string,
	colDefs []*sql_manager.ColumnConstraint,
	updatedConstraints map[string][]*SubsetColumnConstraint,
	aliasReference map[string]string,
	seenTables map[string]struct{},
) []*SubsetColumnConstraint {
	if _, exists := updatedConstraints[table]; exists {
		return updatedConstraints[table]
	}

	newColDefs := []*SubsetColumnConstraint{}
	for _, colDef := range colDefs {
		if colDef.ForeignKey.Table == table {
			continue // self reference skip
		}

		if _, exists := seenTables[colDef.ForeignKey.Table]; exists {
			// create aliased table
			newTable := fmt.Sprintf("%s_%s", strings.ReplaceAll(colDef.ForeignKey.Table, ".", "_"), strings.Join(colDef.Columns, "_"))
			alias := aliasHash(newTable)
			aliasReference[alias] = colDef.ForeignKey.Table
			newColDefs = append(newColDefs, &SubsetColumnConstraint{
				Columns: colDef.Columns,
				ForeignKey: &SubsetReferenceKey{
					Table:         alias,
					OriginalTable: &colDef.ForeignKey.Table,
					Columns:       colDef.ForeignKey.Columns,
				},
			})
		} else {
			newColDefs = append(newColDefs, &SubsetColumnConstraint{
				Columns: colDef.Columns,
				ForeignKey: &SubsetReferenceKey{
					Table:   colDef.ForeignKey.Table,
					Columns: colDef.ForeignKey.Columns,
				},
			})
			seenTables[colDef.ForeignKey.Table] = struct{}{}
		}
	}

	updatedConstraints[table] = newColDefs
	return newColDefs
}

// follows constraints and updates references to alias tables that were created
func updateAliasReferences(
	driver string,
	updatedConstraints map[string][]*SubsetColumnConstraint,
	aliasReference map[string]string,
	updatedWheres map[string]string,
) error {
	for alias, table := range aliasReference {
		if _, exists := updatedConstraints[alias]; exists {
			continue
		}

		colDefs := updatedConstraints[table]
		newColDefs := []*SubsetColumnConstraint{}

		for _, c := range colDefs {
			newAlias := aliasHash(fmt.Sprintf("%s_%s", alias, strings.ReplaceAll(c.ForeignKey.Table, ".", "_")))
			aliasReference[newAlias] = c.ForeignKey.Table
			newColDefs = append(newColDefs, &SubsetColumnConstraint{
				Columns: c.Columns,
				ForeignKey: &SubsetReferenceKey{
					Columns:       c.ForeignKey.Columns,
					OriginalTable: &c.ForeignKey.Table,
					Table:         newAlias,
				},
			})
		}

		updatedConstraints[alias] = newColDefs

		where := updatedWheres[table]
		if where != "" {
			aliasedWhere, err := qualifyWhereWithTableAlias(driver, where, alias)
			if err != nil {
				return err
			}
			updatedWheres[alias] = aliasedWhere
		}

		delete(aliasReference, alias)
		if err := updateAliasReferences(driver, updatedConstraints, aliasReference, updatedWheres); err != nil {
			return err
		}
	}
	return nil
}

// check if table or any parent table has a where clause
func shouldSubsetTable(table string, data map[string][]*sql_manager.ColumnConstraint, whereClauses map[string]string, visited map[string]bool) bool {
	if _, exists := whereClauses[table]; exists {
		return true
	}
	if visited[table] {
		return false
	}
	visited[table] = true

	if columns, exists := data[table]; exists {
		for _, col := range columns {
			if col.ForeignKey.Table != "" {
				if shouldSubsetTable(col.ForeignKey.Table, data, whereClauses, visited) {
					return true
				}
			}
		}
	}
	return false
}

// filters out any foreign keys that are not involved in the subset (where clauses) and tables not in the map
func filterForeignKeysWithSubset(runConfigMap map[string][]*tabledependency.RunConfig, constraints map[string][]*sql_manager.ColumnConstraint, whereClauses map[string]string) map[string][]*sql_manager.ColumnConstraint {
	tableSubsetMap := map[string]bool{} // map of which tables to subset
	for table := range runConfigMap {
		visited := map[string]bool{}
		tableSubsetMap[table] = shouldSubsetTable(table, constraints, whereClauses, visited)
	}

	filteredConstraints := map[string][]*sql_manager.ColumnConstraint{}
	for table, configs := range runConfigMap {
		filteredConstraints[table] = []*sql_manager.ColumnConstraint{}
		for _, c := range configs {
			if c.RunType == tabledependency.RunTypeInsert && len(c.DependsOn) > 0 {
				if tableConstraints, ok := constraints[table]; ok {
					for _, colDef := range tableConstraints {
						if exists := tableSubsetMap[colDef.ForeignKey.Table]; exists {
							filteredConstraints[table] = append(filteredConstraints[table], colDef)
						}
					}
				}
			}
		}
	}
	return filteredConstraints
}

func aliasHash(input string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
	if len(hash) > 14 {
		hash = hash[:14]
	}
	return hash
}

func qualifyWhereWithTableAlias(driver, where, alias string) (string, error) {
	query := goqu.Dialect(driver).From(goqu.T(alias)).Select("*").Where(goqu.L(where))
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}
	var updatedSql string
	switch driver {
	case sql_manager.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, nil, alias)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sql_manager.PostgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, nil, alias)
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

func qualifyWhereColumnNames(driver, where, schema, table string) (string, error) {
	sql, err := buildSelectQuery(driver, schema, table, []string{"*"}, &where)
	if err != nil {
		return "", err
	}
	var updatedSql string
	switch driver {
	case sql_manager.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, &schema, table)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sql_manager.PostgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, &schema, table)
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
	PrimaryKeyColumns []string
	ForeignKeyColumns [][]string
}

func getSelfReferencingColumns(table string, constraints []*sql_manager.ColumnConstraint) []*selfReferencingCircularDependency {
	result := []*selfReferencingCircularDependency{}
	resultMap := map[string]*selfReferencingCircularDependency{}

	for _, c := range constraints {
		if c.ForeignKey.Table != table {
			continue
		}
		pkHash := neosync_benthos.ToSha256(fmt.Sprintf("%s%s", c.ForeignKey.Table, strings.Join(c.ForeignKey.Columns, "")))
		if _, exists := resultMap[pkHash]; !exists {
			resultMap[pkHash] = &selfReferencingCircularDependency{
				PrimaryKeyColumns: c.ForeignKey.Columns,
				ForeignKeyColumns: [][]string{},
			}
		}
		resultMap[pkHash].ForeignKeyColumns = append(resultMap[pkHash].ForeignKeyColumns, c.Columns)
	}

	for _, r := range resultMap {
		result = append(result, r)
	}

	return result
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

func qualifyPostgresWhereColumnNames(sql string, schema *string, table string) (string, error) {
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

func updatePostgresExpr(schema *string, table string, node *pg_query.Node) {
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
			if schema != nil && val == *schema {
				continue
			}
			if val == table {
				isQualified = true
				break
			}
			colName = &val
		}
		if !isQualified && colName != nil && *colName != "" {
			fields := []*pg_query.Node{}
			if schema != nil && *schema != "" {
				fields = append(fields, pg_query.MakeStrNode(*schema))
			}
			fields = append(fields, []*pg_query.Node{
				pg_query.MakeStrNode(table),
				pg_query.MakeStrNode(*colName),
			}...)
			col.Fields = fields
		}
	}
}

func qualifyMysqlWhereColumnNames(sql string, schema *string, table string) (string, error) {
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
					s := ""
					if schema != nil && *schema != "" {
						s = *schema
					}
					col.Qualifier.Qualifier = sqlparser.NewTableIdent(s)
					col.Qualifier.Name = sqlparser.NewTableIdent(table)
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
