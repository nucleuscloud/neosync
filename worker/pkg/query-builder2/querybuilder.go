package querybuilder2

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/doug-martin/goqu/v9"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
	"github.com/xwb1989/sqlparser"
)

type TableInfo struct {
	Schema      string
	Name        string
	Columns     []string
	PrimaryKeys []string
	ForeignKeys []ForeignKey
}

type ForeignKey struct {
	Columns          []string
	ReferenceSchema  string
	ReferenceTable   string
	ReferenceColumns []string
}

type WhereCondition struct {
	Condition string
	Args      []any
}

type QueryBuilder struct {
	tables          map[string]*TableInfo
	whereConditions map[string][]WhereCondition
	defaultSchema   string
	// visitedTables                 map[string]bool
	driver                        string
	subsetByForeignKeyConstraints bool
}

func NewQueryBuilder(defaultSchema, driver string, subsetByForeignKeyConstraints bool) *QueryBuilder {
	return &QueryBuilder{
		tables:          make(map[string]*TableInfo),
		whereConditions: make(map[string][]WhereCondition),
		defaultSchema:   defaultSchema,
		// visitedTables:                 make(map[string]bool),
		driver:                        driver,
		subsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
	}
}

func (qb *QueryBuilder) AddTable(table *TableInfo) {
	key := qb.getTableKey(table.Schema, table.Name)
	qb.tables[key] = table
}

func (qb *QueryBuilder) getRequiredColumns(table *TableInfo) []string {
	columns := make([]string, 0)
	// Add primary keys
	columns = append(columns, table.PrimaryKeys...)
	// Add foreign key columns
	for _, fk := range table.ForeignKeys {
		columns = append(columns, fk.Columns...)
	}
	// Add columns used in WHERE conditions
	if conditions, ok := qb.whereConditions[qb.getTableKey(table.Schema, table.Name)]; ok {
		for _, cond := range conditions {
			parts := strings.Fields(cond.Condition)
			if len(parts) > 0 {
				columns = append(columns, parts[0])
			}
		}
	}
	// Remove duplicates
	return uniqueStrings(columns)
}

func (qb *QueryBuilder) AddWhereCondition(schema, tableName, condition string, args ...any) {
	key := qb.getTableKey(schema, tableName)
	qb.whereConditions[key] = append(qb.whereConditions[key], WhereCondition{Condition: condition, Args: args})
}

func (qb *QueryBuilder) BuildQuery(schema, tableName string) (string, []any, error) {
	key := qb.getTableKey(schema, tableName)
	table, ok := qb.tables[key]
	if !ok {
		return "", nil, fmt.Errorf("table not found: %s", key)
	}
	return qb.buildQueryRecursive(schema, tableName, nil, table.Columns, map[string]int{}, map[string]bool{})
}

func (qb *QueryBuilder) isSelfReferencing(table *TableInfo) bool {
	for _, fk := range table.ForeignKeys {
		if fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
			return true
		}
	}
	return false
}

func (qb *QueryBuilder) buildQueryRecursive(
	schema, tableName string, parentTable *TableInfo,
	columnsToInclude []string, joinCount map[string]int,
	visitedTables map[string]bool,
) (string, []any, error) {
	key := qb.getTableKey(schema, tableName)
	if visitedTables[key] {
		return "", nil, nil // Avoid circular dependencies
	}
	visitedTables[key] = true
	defer delete(visitedTables, key) // Remove from visited after processing

	table, ok := qb.tables[key]
	if !ok {
		return "", nil, fmt.Errorf("table not found: %s", key)
	}

	if len(columnsToInclude) == 0 {
		columnsToInclude = qb.getRequiredColumns(table)
	}
	if len(columnsToInclude) == 0 {
		columnsToInclude = table.Columns
	}
	if len(columnsToInclude) == 0 {
		columnsToInclude = []string{"*"}
	}

	var query squirrel.SelectBuilder
	var allArgs []any

	if qb.isSelfReferencing(table) && qb.subsetByForeignKeyConstraints && parentTable == nil {
		whereConditions := []string{}
		whereArgs := []any{}
		for _, cond := range qb.whereConditions[key] {
			whereConditions = append(whereConditions, cond.Condition)
			whereArgs = append(whereArgs, cond.Args...)
		}
		whereClause := strings.Join(whereConditions, " AND ")
		cteQuery := qb.buildRecursiveCTE(table, &whereClause)

		// Create a new SelectBuilder with the CTE

		// Use the CTE select as a subquery
		// squirrel.SelectBuilder{}.From(cteQuery)
		query = squirrel.Select("*").From(fmt.Sprintf("(%s) as recursive_cte", cteQuery))
		allArgs = append(allArgs, whereArgs...)
	} else {
		query = squirrel.Select(qb.getQualifiedColumns(table, columnsToInclude)...).From(qb.getQualifiedTableName(table))

		// Add WHERE conditions for this table
		if conditions, ok := qb.whereConditions[key]; ok {
			for _, cond := range conditions {
				qualifiedCondition := qb.qualifyWhereCondition(table, cond.Condition)
				query = query.Where(qualifiedCondition, cond.Args...)
				allArgs = append(allArgs, cond.Args...)
			}
		}

		// Only join and apply subsetting if subsetByForeignKeyConstraints is true
		if qb.subsetByForeignKeyConstraints {
			// Recursively build and join queries for related tables
			for _, fk := range table.ForeignKeys {
				if fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
					continue // Skip self-referencing foreign keys here
				}
				subQuery, subArgs, err := qb.buildQueryRecursive(fk.ReferenceSchema, fk.ReferenceTable, table, fk.ReferenceColumns, joinCount, visitedTables)
				if err != nil {
					return "", nil, err
				}
				if subQuery != "" {
					joinCount[fk.ReferenceTable]++
					subQueryAlias := fmt.Sprintf("%s_%s_%d", fk.ReferenceSchema, fk.ReferenceTable, joinCount[fk.ReferenceTable])
					joinCondition := qb.buildJoinCondition(table, fk, joinCount[fk.ReferenceTable])
					query = query.JoinClause(fmt.Sprintf("INNER JOIN (%s) AS %s ON %s",
						subQuery,
						quoteIdentifier(qb.driver, subQueryAlias),
						joinCondition))
					allArgs = append(allArgs, subArgs...)
				}
			}
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return "", nil, err
	}

	return sql, append(allArgs, args...), nil
}

func (qb *QueryBuilder) qualifyWhereCondition(table *TableInfo, condition, driver string) (string, error) {
	query := goqu.Dialect(driver).From(goqu.T(table.Name)).Select("*").Where(goqu.L(condition))
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", fmt.Errorf("unable to build where condition: %w", err)
	}

	var updatedSql string
	switch driver {
	case sqlmanager_shared.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, &table.Schema, table.Name)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sqlmanager_shared.PostgresDriver:
		tree, err := pgquery.Parse(sql)
		if err != nil {
			return "", fmt.Errorf("unable to parse where condition for postgres query: %w", err)
		}
		for _, stmt := range tree.GetStmts() {
			selectStmt := stmt.GetStmt().GetSelectStmt()
			if selectStmt.WhereClause != nil {
				updatePostgresExpr(&table.Schema, table.Name, selectStmt.WhereClause)
			}
		}
		sql, err := pgquery.Deparse(tree)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	default:
		return "", fmt.Errorf("unsupported driver %q when qualifying where condition", driver)
	}
	index := strings.Index(strings.ToLower(updatedSql), "where")
	if index == -1 {
		// "where" not found
		return "", fmt.Errorf("unable to qualify where column names")
	}
	startIndex := index + len("where")
	return strings.TrimSpace(updatedSql[startIndex:]), nil
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

func (qb *QueryBuilder) getQualifiedTableName(table *TableInfo) string {
	if table.Schema == "" || table.Schema == qb.defaultSchema {
		return quoteIdentifier(qb.driver, table.Name)
	}
	return fmt.Sprintf("%s.%s",
		quoteIdentifier(qb.driver, table.Schema),
		quoteIdentifier(qb.driver, table.Name))
}

func (qb *QueryBuilder) getQualifiedColumns(table *TableInfo, columnsToInclude []string) []string {
	qualifiedColumns := make([]string, len(columnsToInclude))
	for i, col := range columnsToInclude {
		qualifiedColumns[i] = fmt.Sprintf("%s.%s",
			qb.getQualifiedTableName(table),
			quoteIdentifier(qb.driver, col))
	}
	return qualifiedColumns
}

func (qb *QueryBuilder) buildJoinCondition(table *TableInfo, fk ForeignKey, joinCount int) string {
	conditions := make([]string, len(fk.Columns))
	for i := range fk.Columns {
		conditions[i] = fmt.Sprintf("%s.%s = %s.%s",
			qb.getQualifiedTableName(table),
			quoteIdentifier(qb.driver, fk.Columns[i]),
			quoteIdentifier(qb.driver, fmt.Sprintf("%s_%s_%d", fk.ReferenceSchema, fk.ReferenceTable, joinCount)),
			quoteIdentifier(qb.driver, fk.ReferenceColumns[i]))
	}
	return strings.Join(conditions, " AND ")
}

func (qb *QueryBuilder) buildSubsetCondition(childTable, parentTable *TableInfo, parentCondition string) string {
	for _, fk := range childTable.ForeignKeys {
		if fk.ReferenceSchema == parentTable.Schema && fk.ReferenceTable == parentTable.Name {
			parts := strings.Fields(parentCondition)
			if len(parts) > 2 && (parts[1] == "=" || parts[1] == "IN") {
				return fmt.Sprintf("%s.%s %s %s",
					qb.getQualifiedTableName(childTable),
					quoteIdentifier(qb.driver, fk.Columns[0]),
					parts[1],
					strings.Join(parts[2:], " "))
			}
		}
	}
	return ""
}

func (qb *QueryBuilder) getTableKey(schema, tableName string) string {
	if schema == "" {
		schema = qb.defaultSchema
	}
	return fmt.Sprintf("%s.%s", schema, tableName)
}

func (qb *QueryBuilder) buildRecursiveCTE(table *TableInfo, whereCondition *string) string {
	columns := qb.getQualifiedColumns(table, table.Columns)

	baseCase := fmt.Sprintf(`
        SELECT %s
        FROM %s`,
		strings.Join(columns, ", "),
		qb.getQualifiedTableName(table))
	if whereCondition != nil && *whereCondition != "" {
		baseCase = fmt.Sprintf("%s\nWHERE %s", baseCase, *whereCondition)
	}

	recursiveCase := fmt.Sprintf(`
        SELECT %s
        FROM %s AS b
        INNER JOIN hierarchy h ON `,
		strings.Join(qb.getAliasedColumns("b", table.Columns), ", "),
		qb.getQualifiedTableName(table))

	joinConditions := []string{}
	for _, fk := range table.ForeignKeys {
		if fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
			for i, col := range fk.Columns {
				joinConditions = append(
					joinConditions,
					fmt.Sprintf("b.%s = h.%s",
						quoteIdentifier(qb.driver, fk.ReferenceColumns[i]),
						quoteIdentifier(qb.driver, col),
					),
				)
			}
		}
	}
	recursiveCase += strings.Join(joinConditions, " OR ")

	return fmt.Sprintf(`
    WITH RECURSIVE hierarchy AS (
        %s
        UNION ALL
        %s
    )
    SELECT DISTINCT * FROM hierarchy`, baseCase, recursiveCase)
}

func (qb *QueryBuilder) getAliasedColumns(alias string, columns []string) []string {
	aliasedColumns := make([]string, len(columns))
	for i, col := range columns {
		aliasedColumns[i] = fmt.Sprintf("%s.%s", alias, quoteIdentifier(qb.driver, col))
	}
	return aliasedColumns
}

func quoteIdentifier(driver, identifier string) string {
	switch driver {
	case sqlmanager_shared.PostgresDriver:
		return fmt.Sprintf("\"%s\"", strings.Replace(identifier, "\"", "\"\"", -1))
	case sqlmanager_shared.MysqlDriver:
		return fmt.Sprintf("`%s`", strings.Replace(identifier, "`", "``", -1))
	default:
		return identifier
	}
}
