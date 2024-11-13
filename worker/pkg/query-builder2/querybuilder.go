package querybuilder2

import (
	"crypto/md5" //nolint:gosec // This is not being used for a purpose that requires security
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/doug-martin/goqu/v9/sqlgen"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tsql_parser "github.com/nucleuscloud/neosync/worker/pkg/query-builder2/tsql"
	pg_query "github.com/pganalyze/pg_query_go/v5"
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

type set map[string]struct{}

func (s set) add(key string) {
	s[key] = struct{}{}
}

func (s set) contains(key string) bool {
	_, exists := s[key]
	return exists
}

type QueryBuilder struct {
	tables          map[string]*TableInfo
	whereConditions map[string][]WhereCondition
	// schema.table -> column -> { column info }
	columnInfo                    map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow
	defaultSchema                 string
	driver                        string
	subsetByForeignKeyConstraints bool
	tablesWithWhereConditions     set
	pathCache                     set
	aliasCounter                  int
}

func NewQueryBuilder(defaultSchema, driver string, subsetByForeignKeyConstraints bool, columnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow) *QueryBuilder {
	return &QueryBuilder{
		tables:                        make(map[string]*TableInfo),
		whereConditions:               make(map[string][]WhereCondition),
		defaultSchema:                 defaultSchema,
		driver:                        driver,
		subsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
		columnInfo:                    columnInfo,
		tablesWithWhereConditions:     make(set),
		pathCache:                     make(set),
		aliasCounter:                  0,
	}
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

func (qb *QueryBuilder) AddWhereCondition(schema, tableName, condition string, args ...any) {
	key := qb.getTableKey(schema, tableName)
	qb.whereConditions[key] = append(qb.whereConditions[key], WhereCondition{Condition: condition, Args: args})
	qb.tablesWithWhereConditions.add(key)
	qb.clearPathCache()
}
func (qb *QueryBuilder) clearPathCache() {
	qb.pathCache = make(set)
}

func (qb *QueryBuilder) BuildQuery(schema, tableName string) (sqlstatement string, args []any, err error) {
	key := qb.getTableKey(schema, tableName)
	table, ok := qb.tables[key]
	if !ok {
		return "", nil, fmt.Errorf("table not found: %s", key)
	}
	query, err := qb.buildFlattenedQuery(table)
	if query == nil {
		return "", nil, fmt.Errorf("received no error, but query was nil for %s.%s", schema, tableName)
	}
	if err != nil {
		return "", nil, err
	}
	sql, args, err := query.ToSQL()
	if err != nil {
		return "", nil, fmt.Errorf("unable to convery structured query to string for %s.%s: %w", schema, tableName, err)
	}
	return sql, args, nil
}

func (qb *QueryBuilder) buildFlattenedQuery(rootTable *TableInfo) (*goqu.SelectDataset, error) {
	dialect := qb.getDialect()
	rootAlias := rootTable.Name
	rootAliasExpression := rootTable.GetIdentifierExpression().As(rootAlias)
	query := dialect.From(rootAliasExpression)

	// Select columns for the root table
	cols := make([]exp.Expression, len(rootTable.Columns))
	for i, col := range rootTable.Columns {
		cols[i] = rootAliasExpression.Col(col)
	}
	query = query.Select(toAnySlice(cols)...)

	// Add WHERE conditions for the root table
	if conditions, ok := qb.whereConditions[qb.getTableKey(rootTable.Schema, rootTable.Name)]; ok {
		for _, cond := range conditions {
			query = query.Where(goqu.L(cond.Condition, cond.Args...))
		}
	}

	// Flatten and add necessary joins
	if qb.subsetByForeignKeyConstraints && len(qb.whereConditions) > 0 {
		joinedTables := make(map[string]bool)
		var err error
		query, err = qb.addFlattenedJoins(query, rootTable, rootTable, rootAlias, joinedTables, "")
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

func (qb *QueryBuilder) addFlattenedJoins(
	query *goqu.SelectDataset,
	rootTable *TableInfo,
	table *TableInfo,
	tableAlias string,
	joinedTables map[string]bool,
	prefix string,
) (*goqu.SelectDataset, error) {
	tableKey := qb.getTableKey(table.Schema, table.Name)
	rootTableKey := qb.getTableKey(rootTable.Schema, rootTable.Name)

	if joinedTables[tableKey] {
		return query, nil // Avoid circular dependencies
	}
	joinedTables[tableKey] = true

	for _, fk := range table.ForeignKeys {
		// Skip self-referencing foreign keys
		if fk.ReferenceSchema == table.Schema && fk.ReferenceTable == table.Name {
			continue
		}
		refKey := qb.getTableKey(fk.ReferenceSchema, fk.ReferenceTable)
		refTable := qb.tables[refKey]
		if refTable == nil {
			continue
		}

		if refKey == rootTableKey {
			continue
		}

		// Only add join if the referenced table has WHERE conditions
		if !qb.hasPathToWhereCondition(refTable, make(set)) {
			continue
		}

		aliasName := qb.generateUniqueAlias(prefix, refTable.Name, joinedTables)
		joinConditions := make([]exp.Expression, len(fk.Columns))
		for i, col := range fk.Columns {
			joinConditions[i] = goqu.T(aliasName).Col(fk.ReferenceColumns[i]).Eq(goqu.T(tableAlias).Col(col))
		}

		query = query.InnerJoin(
			refTable.GetIdentifierExpression().As(aliasName),
			goqu.On(joinConditions...),
		)

		// Add WHERE conditions for the joined table
		if conditions, ok := qb.whereConditions[refKey]; ok {
			for _, cond := range conditions {
				qualifiedCondition, err := qb.qualifyWhereCondition(nil, aliasName, cond.Condition)
				if err != nil {
					return nil, err
				}
				query = query.Where(goqu.L(qualifiedCondition, cond.Args...))
			}
		}

		// Recursively add joins for the referenced table
		var err error
		query, err = qb.addFlattenedJoins(query, rootTable, refTable, aliasName, joinedTables, aliasName+"_")
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

func (qb *QueryBuilder) generateUniqueAlias(prefix, tableName string, joinedTables map[string]bool) string {
	qb.aliasCounter++
	baseString := fmt.Sprintf("%s%s%d", prefix, tableName, qb.aliasCounter)
	alias := "t_" + getClippedHash(baseString)

	// Ensure uniqueness
	for {
		if _, exists := joinedTables[alias]; !exists {
			joinedTables[alias] = true
			return alias
		}
		qb.aliasCounter++
		baseString = fmt.Sprintf("%s%s%d", prefix, tableName, qb.aliasCounter)
		alias = "t_" + getClippedHash(baseString)
	}
}

func getClippedHash(input string) string {
	hash := md5.Sum([]byte(input)) //nolint:gosec // Not used for anything that is security related
	return hex.EncodeToString(hash[:][:8])
}

func (qb *QueryBuilder) hasPathToWhereCondition(
	table *TableInfo,
	visited set,
) bool {
	tableKey := qb.getTableKey(table.Schema, table.Name)
	if visited.contains(tableKey) {
		return false
	}
	visited.add(tableKey)

	if qb.pathCache.contains(tableKey) {
		return true
	}

	if qb.tablesWithWhereConditions.contains(tableKey) {
		qb.pathCache.add(tableKey)
		return true
	}

	for _, fk := range table.ForeignKeys {
		refKey := qb.getTableKey(fk.ReferenceSchema, fk.ReferenceTable)
		refTable := qb.tables[refKey]
		if refTable == nil {
			continue
		}

		if qb.hasPathToWhereCondition(refTable, visited) {
			qb.pathCache.add(tableKey)
			return true
		}
	}

	return false
}

func (qb *QueryBuilder) qualifyWhereCondition(schema *string, table, condition string) (string, error) {
	query := qb.getDialect().From(goqu.T(table)).Select(goqu.Star()).Where(goqu.L(condition))
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", fmt.Errorf("unable to build where condition: %w", err)
	}

	var updatedSql string
	switch qb.driver {
	case sqlmanager_shared.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, schema, table)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sqlmanager_shared.PostgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, schema, table)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sqlmanager_shared.MssqlDriver:
		sql, err := tsql_parser.QualifyWhereCondition(sql)
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
	tree, err := pg_query.Parse(sql)
	if err != nil {
		return "", err
	}

	for _, stmt := range tree.GetStmts() {
		selectStmt := stmt.GetStmt().GetSelectStmt()

		if selectStmt.WhereClause != nil {
			updatePostgresExpr(schema, table, selectStmt.WhereClause)
		}
	}
	updatedSql, err := pg_query.Deparse(tree)
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

func (qb *QueryBuilder) getTableKey(schema, tableName string) string {
	if schema == "" {
		schema = qb.defaultSchema
	}
	return fmt.Sprintf("%s.%s", schema, tableName)
}

func toAnySlice[T any](input []T) []any {
	anys := make([]any, len(input))
	for i := range input {
		anys[i] = input[i]
	}
	return anys
}
