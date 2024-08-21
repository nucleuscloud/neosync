package querybuilder2

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/doug-martin/goqu/v9/sqlgen"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
	"github.com/xwb1989/sqlparser"
)

const (
	mysqlDialect = "custom-mysql-dialect"
)

func init() {
	goqu.RegisterDialect(mysqlDialect, buildMysqlDialect())
}

func buildMysqlDialect() *sqlgen.SQLDialectOptions {
	opts := goqu.DefaultDialectOptions()
	opts.QuoteRune = '`'
	opts.SupportsWithCTERecursive = true
	opts.SupportsWithCTE = true
	return opts
}

type TableInfo struct {
	Schema      string
	Name        string
	Columns     []string
	PrimaryKeys []string
	ForeignKeys []*ForeignKey
}

type AliasTableInfo struct {
	Name string
}

// Returns table info with just the table name
func newTableInfoAlias(alias string) *AliasTableInfo {
	return &AliasTableInfo{Name: alias}
}
func (t *AliasTableInfo) GetSchema() *string {
	return nil
}
func (t *AliasTableInfo) GetName() string {
	return t.Name
}

type TableIdentity interface {
	GetSchema() *string
	GetName() string
}

func (t *TableInfo) GetSchema() *string {
	if t.Schema == "" {
		return nil
	}
	return &t.Schema
}
func (t *TableInfo) GetName() string {
	return t.Name
}

func (t *TableInfo) GetIdentifierExpression() exp.IdentifierExpression {
	table := goqu.T(t.Name)
	if t.Schema == "" {
		return table
	}
	return table.Schema(t.Schema)
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
	// schema.table -> column -> { column info }
	columnInfo                    map[string]map[string]*sqlmanager_shared.ColumnInfo
	defaultSchema                 string
	driver                        string
	subsetByForeignKeyConstraints bool
}

func NewQueryBuilder(defaultSchema, driver string, subsetByForeignKeyConstraints bool, columnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo) *QueryBuilder {
	return &QueryBuilder{
		tables:                        make(map[string]*TableInfo),
		whereConditions:               make(map[string][]WhereCondition),
		defaultSchema:                 defaultSchema,
		driver:                        driver,
		subsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
		columnInfo:                    columnInfo,
	}
}

func (qb *QueryBuilder) isJsonColumn(schematable, column string) bool {
	tableInfo, ok := qb.columnInfo[schematable]
	if ok {
		colInfo, ok := tableInfo[column]
		if ok {
			return colInfo.DataType == "json"
		}
	}
	return false
}

func (qb *QueryBuilder) AddTable(table *TableInfo) {
	key := qb.getTableKey(table.Schema, table.Name)
	qb.tables[key] = table
}

func (qb *QueryBuilder) getDialect() goqu.DialectWrapper {
	switch qb.driver {
	case sqlmanager_shared.MysqlDriver:
		return goqu.Dialect(mysqlDialect)
	default:
		return goqu.Dialect(qb.driver)
	}
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

func (qb *QueryBuilder) BuildQuery(schema, tableName string) (sqlstatement string, args []any, err error) {
	key := qb.getTableKey(schema, tableName)
	table, ok := qb.tables[key]
	if !ok {
		return "", nil, fmt.Errorf("table not found: %s", key)
	}
	query, err := qb.buildQueryRecursive(schema, tableName, nil, table.Columns, map[string]int{}, map[string]bool{})
	if err != nil {
		return "", nil, fmt.Errorf("unable to build query for %s.%s: %w", schema, tableName, err)
	}
	if query == nil {
		return "", nil, fmt.Errorf("received no error, but query was nil for %s.%s", schema, tableName)
	}
	sql, args, err := query.ToSQL()
	if err != nil {
		return "", nil, fmt.Errorf("unable to convery structured query to string for %s.%s: %w", schema, tableName, err)
	}
	return sql, args, nil
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
) (*goqu.SelectDataset, error) {
	key := qb.getTableKey(schema, tableName)
	if visitedTables[key] {
		return nil, nil // Avoid circular dependencies
	}
	visitedTables[key] = true
	defer delete(visitedTables, key) // Remove from visited after processing

	table, ok := qb.tables[key]
	if !ok {
		return nil, fmt.Errorf("table not found: %s", key)
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

	dialect := qb.getDialect()
	var query *goqu.SelectDataset

	isSelfReferencing := qb.isSelfReferencing(table)

	if isSelfReferencing && qb.subsetByForeignKeyConstraints && len(qb.whereConditions) > 0 && parentTable == nil {
		// Handle self-referencing table with possible additional foreign keys
		cteQuery, err := qb.buildRecursiveCTE(table, qb.whereConditions[key])
		if err != nil {
			return nil, err
		}
		query = dialect.From(cteQuery.As("recursive_cte"))

		// Add joins for other foreign keys
		for _, fk := range table.ForeignKeys {
			if fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
				continue // Skip the self-referencing foreign key
			}
			subQuery, err := qb.buildQueryRecursive(fk.ReferenceSchema, fk.ReferenceTable, table, fk.ReferenceColumns, joinCount, visitedTables)
			if err != nil {
				return nil, err
			}
			if subQuery != nil {
				joinCount[fk.ReferenceTable]++
				subqueryAlias := fmt.Sprintf("%s_%s_%d", fk.ReferenceSchema, fk.ReferenceTable, joinCount[fk.ReferenceTable])
				conditions := make([]goqu.Expression, len(fk.Columns))
				for i := range fk.Columns {
					conditions[i] = goqu.T("recursive_cte").Col(fk.Columns[i]).Eq(
						goqu.T(subqueryAlias).Col(fk.ReferenceColumns[i]),
					)
				}
				query = query.Join(
					goqu.L("(?)", subQuery).As(subqueryAlias),
					goqu.On(goqu.And(conditions...)),
				)
			}
		}
	} else {
		t := table.GetIdentifierExpression()
		cols := make([]exp.Expression, len(columnsToInclude))
		for i := range columnsToInclude {
			cols[i] = t.Col(columnsToInclude[i])
		}
		query = dialect.From(t).Select(toAnySlice(cols)...)

		// Add WHERE conditions for this table
		if conditions, ok := qb.whereConditions[key]; ok {
			for _, cond := range conditions {
				qualifiedCondition, err := qb.qualifyWhereCondition(table, cond.Condition)
				if err != nil {
					return nil, err
				}
				query = query.Where(goqu.L(qualifiedCondition, cond.Args...))
			}
		}

		// Only join and apply subsetting if subsetByForeignKeyConstraints is true
		if qb.subsetByForeignKeyConstraints && len(qb.whereConditions) > 0 {
			// Recursively build and join queries for related tables
			for _, fk := range table.ForeignKeys {
				if isSelfReferencing && fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
					continue // Skip self-referencing foreign keys here
				}
				subQuery, err := qb.buildQueryRecursive(fk.ReferenceSchema, fk.ReferenceTable, table, fk.ReferenceColumns, joinCount, visitedTables)
				if err != nil {
					return nil, err
				}

				if subQuery != nil {
					joinCount[fk.ReferenceTable]++
					subQueryAlias := fmt.Sprintf("%s_%s_%d", fk.ReferenceSchema, fk.ReferenceTable, joinCount[fk.ReferenceTable])
					conditions := make([]goqu.Expression, len(fk.Columns))
					for i := range fk.Columns {
						conditions[i] = t.Col(fk.Columns[i]).Eq(
							goqu.T(subQueryAlias).Col(fk.ReferenceColumns[i]),
						)
					}
					query = query.Join(
						goqu.L("(?)", subQuery).As(subQueryAlias),
						goqu.On(goqu.And(conditions...)),
					)
				}
			}
		}
	}

	return query, nil
}

func (qb *QueryBuilder) qualifyWhereCondition(table TableIdentity, condition string) (string, error) {
	query := qb.getDialect().From(goqu.T(table.GetName())).Select(goqu.Star()).Where(goqu.L(condition))
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", fmt.Errorf("unable to build where condition: %w", err)
	}

	var updatedSql string
	switch qb.driver {
	case sqlmanager_shared.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, table.GetSchema(), table.GetName())
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sqlmanager_shared.PostgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, table.GetSchema(), table.GetName())
		if err != nil {
			return "", err
		}
		updatedSql = sql
	default:
		return "", fmt.Errorf("unsupported driver %q when qualifying where condition", qb.driver)
	}
	index := strings.Index(strings.ToLower(updatedSql), "where")
	if index == -1 {
		// "where" not found
		return "", fmt.Errorf("unable to qualify where column names")
	}
	startIndex := index + len("where")
	return strings.TrimSpace(updatedSql[startIndex:]), nil
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

func (qb *QueryBuilder) getQualifiedTableName(table *TableInfo) exp.IdentifierExpression {
	schema := table.Schema
	if schema == "" {
		schema = qb.defaultSchema
	}
	return goqu.T(table.Name).Schema(schema)
}

func (qb *QueryBuilder) getTableKey(schema, tableName string) string {
	if schema == "" {
		schema = qb.defaultSchema
	}
	return fmt.Sprintf("%s.%s", schema, tableName)
}

const (
	hierarchyTableName = "hierarchy"
)

func (qb *QueryBuilder) buildRecursiveCTE(
	table *TableInfo,
	whereConditions []WhereCondition,
) (*goqu.SelectDataset, error) {
	dialect := qb.getDialect()

	baseTable := qb.getQualifiedTableName(table).As("b")
	baseColumns := make([]exp.Expression, len(table.Columns))
	tableKey := qb.getTableKey(table.Schema, table.Name)
	for i, col := range table.Columns {
		if qb.isJsonColumn(tableKey, col) {
			// json the type has no comparison operator and thus can't union but jsonb does
			baseColumns[i] = goqu.L("to_jsonb(?)", baseTable.Col(col)).As(col)
		} else {
			baseColumns[i] = baseTable.Col(col)
		}
	}

	// Base case
	baseQuery := dialect.
		From(baseTable).
		Select(toAnySlice(baseColumns)...)
	qualifiedWheres := []goqu.Expression{}
	for _, whereCond := range whereConditions {
		qualifiedCondition, err := qb.qualifyWhereCondition(newTableInfoAlias("b"), whereCond.Condition)
		if err != nil {
			return nil, err
		}
		qualifiedWheres = append(qualifiedWheres, goqu.L(qualifiedCondition, whereCond.Args...))
	}
	if len(qualifiedWheres) > 0 {
		baseQuery = baseQuery.Where(goqu.And(qualifiedWheres...))
	}

	recursiveTable := qb.getQualifiedTableName(table).As("r")
	hierarchicalTable := goqu.T(hierarchyTableName).As("h")

	recursiveColumns := make([]exp.IdentifierExpression, len(table.Columns))
	for i, col := range table.Columns {
		recursiveColumns[i] = recursiveTable.Col(col)
	}

	joinConditions := []exp.Expression{}
	for _, fk := range table.ForeignKeys {
		if fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
			for i, col := range fk.Columns {
				joinConditions = append(joinConditions,
					recursiveTable.Col(fk.ReferenceColumns[i]).Eq(hierarchicalTable.Col(col)),
				)
			}
		}
	}

	recursiveQuery := dialect.
		From(recursiveTable).
		Select(toAnySlice(recursiveColumns)...).
		Join(
			hierarchicalTable,
			goqu.On(
				goqu.Or(joinConditions...),
			),
		)

	cte := dialect.
		From(hierarchyTableName).
		Select(goqu.Star()).
		Distinct().
		WithRecursive(hierarchyTableName, baseQuery.Union(recursiveQuery))
	return cte, nil
}

func toAnySlice[T any](input []T) []any {
	anys := make([]any, len(input))
	for i := range input {
		anys[i] = input[i]
	}
	return anys
}
