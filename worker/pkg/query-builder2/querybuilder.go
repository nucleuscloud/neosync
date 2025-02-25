package querybuilder2

import (
	"crypto/md5" //nolint:gosec // This is not being used for a purpose that requires security
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/doug-martin/goqu/v9/exp"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	tsql_parser "github.com/nucleuscloud/neosync/worker/pkg/query-builder2/tsql"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	"github.com/xwb1989/sqlparser"
)

// TableInfo represents a table’s metadata.
type TableInfo struct {
	Schema         string
	Name           string
	Columns        []string
	PrimaryKeys    []string
	SubsetPaths    []runconfigs.SubsetPath
	OrderByColumns []string
}

// GetIdentifierExpression returns a goqu expression for the table.
func (t *TableInfo) GetIdentifierExpression() exp.IdentifierExpression {
	table := goqu.T(t.Name)
	if t.Schema == "" {
		return table
	}
	return table.Schema(t.Schema)
}

// // ForeignKey represents a foreign key relationship.
// type ForeignKey struct {
// 	Columns          []string
// 	NotNullable      []bool
// 	ReferenceSchema  string
// 	ReferenceTable   string
// 	ReferenceColumns []string
// }

// type SubsetPath struct {
// 	Subset    string
// 	Root      string
// 	Path      []string
// 	JoinSteps []JoinStep
// }

// // joinStep represents one “edge” (join) in a join chain from one table to a related table.
// type JoinStep struct {
// 	fromKey string
// 	toKey   string
// 	fk      *ForeignKey
// }

// WhereCondition represents a condition to be applied in the WHERE clause.
type WhereCondition struct {
	Condition string
	Args      []any
}

// set is a simple string set.
type set map[string]struct{}

func (s set) add(key string) {
	s[key] = struct{}{}
}

// func (s set) contains(key string) bool {
// 	_, exists := s[key]
// 	return exists
// }

// QueryBuilder holds state for building the query.
type QueryBuilder struct {
	tables          map[string]*TableInfo
	whereConditions map[string][]WhereCondition
	orderBy         map[string][]string
	// schema.table -> column -> { column info }
	columnInfo                    map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow
	defaultSchema                 string
	driver                        string
	subsetByForeignKeyConstraints bool
	tablesWithWhereConditions     set
	// pathCache                     set
	aliasCounter int
	pageLimit    uint
}

// NewQueryBuilder constructs a new QueryBuilder.
func NewQueryBuilder(defaultSchema, driver string, subsetByForeignKeyConstraints bool, columnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow, pageLimit int) *QueryBuilder {
	limit := uint(0)
	if pageLimit > 0 {
		limit = uint(pageLimit)
	}
	return &QueryBuilder{
		tables:                        make(map[string]*TableInfo),
		whereConditions:               make(map[string][]WhereCondition),
		orderBy:                       make(map[string][]string),
		defaultSchema:                 defaultSchema,
		driver:                        driver,
		subsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
		columnInfo:                    columnInfo,
		tablesWithWhereConditions:     make(set),
		// pathCache:                     make(set),
		aliasCounter: 0,
		pageLimit:    limit,
	}
}

func (qb *QueryBuilder) AddTable(table *TableInfo) {
	key := qb.getTableKey(table.Schema, table.Name)
	qb.tables[key] = table
}

func (qb *QueryBuilder) AddOrderBy(schema, tableName string, orderBy []string) {
	key := qb.getTableKey(schema, tableName)
	qb.orderBy[key] = orderBy
}

func (qb *QueryBuilder) getDialect() goqu.DialectWrapper {
	switch qb.driver {
	case sqlmanager_shared.PostgresDriver:
		return goqu.Dialect(sqlmanager_shared.GoquPostgresDriver)
	default:
		return goqu.Dialect(qb.driver)
	}
}

func (qb *QueryBuilder) AddWhereCondition(schema, tableName, condition string, args ...any) {
	key := qb.getTableKey(schema, tableName)
	qb.whereConditions[key] = append(qb.whereConditions[key], WhereCondition{Condition: condition, Args: args})
	qb.tablesWithWhereConditions.add(key)
	// qb.clearPathCache()
}

// func (qb *QueryBuilder) clearPathCache() {
// 	qb.pathCache = make(set)
// }

// BuildQuery creates the final SQL query.
func (qb *QueryBuilder) BuildQuery(schema, tableName string) (sqlstatement string, args []any, pagesql string, isNotForeignKeySafeSubset bool, err error) {
	fmt.Println("---------------------------------")
	fmt.Println("Building query for", schema, tableName)
	key := qb.getTableKey(schema, tableName)
	table, ok := qb.tables[key]
	if !ok {
		return "", nil, "", false, fmt.Errorf("table not found: %s", key)
	}
	query, pageQuery, notFkSafe, err := qb.buildFlattenedQuery(table)
	if query == nil {
		return "", nil, "", false, fmt.Errorf("received no error, but query was nil for %s.%s", schema, tableName)
	}
	if err != nil {
		return "", nil, "", false, err
	}

	sql, args, err := query.Limit(qb.pageLimit).ToSQL()
	if err != nil {
		return "", nil, "", false, fmt.Errorf("unable to convert structured query to string for %s.%s: %w", schema, tableName, err)
	}

	pageSql, _, err := pageQuery.Limit(qb.pageLimit).ToSQL()
	if err != nil {
		return "", nil, "", false, fmt.Errorf("unable to convert structured page query to string for %s.%s: %w", schema, tableName, err)
	}
	if qb.driver == sqlmanager_shared.MssqlDriver {
		// MSSQL TOP needs to be cast to int
		pageSql = strings.Replace(pageSql, "TOP (@p1)", "TOP (CAST(@p1 AS INT))", 1)
	}
	fmt.Println("sql", sql)
	fmt.Println()
	fmt.Println()
	return sql, args, pageSql, notFkSafe, nil
}

// buildFlattenedQuery builds the query for the root table, adding joins if needed.
func (qb *QueryBuilder) buildFlattenedQuery(rootTable *TableInfo) (sql, pageSql *goqu.SelectDataset, isNotForeignKeySafeSubset bool, err error) {
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

	// Add order by if specified
	if orderBy, ok := qb.orderBy[qb.getTableKey(rootTable.Schema, rootTable.Name)]; ok {
		orderByExpressions := make([]exp.OrderedExpression, len(orderBy))
		for i, col := range orderBy {
			orderByExpressions[i] = rootAliasExpression.Col(col).Asc()
		}
		query = query.Order(orderByExpressions...)
	}

	// If using subset-by-foreign-key constraints, add joins using the shortest path logic.
	var notFkSafeSubset bool
	if qb.subsetByForeignKeyConstraints && len(qb.whereConditions) > 0 {
		fmt.Println("HERE ADD SHORTEST PATH JOINS")
		query, err = qb.addShortestPathJoins(query, rootTable, rootAlias)
		if err != nil {
			return nil, nil, false, err
		}
		notFkSafeSubset = true
	}

	// Build page query (for pagination)
	pageQuery := qb.buildPageQuery(rootTable.Schema, rootTable.Name, query, rootAlias)

	// In this new approach we assume all joins are safe so we return false.
	return query, pageQuery, notFkSafeSubset, nil
}

// buildPageQuery builds a pagination version of the query.
func (qb *QueryBuilder) buildPageQuery(schema, tableName string, query *goqu.SelectDataset, rootAlias string) *goqu.SelectDataset {
	key := qb.getTableKey(schema, tableName)
	orderBy := qb.orderBy[key]
	if len(orderBy) > 0 {
		// Build lexicographical ordering conditions
		var conditions []exp.Expression
		for i := 0; i < len(orderBy); i++ {
			var subConditions []exp.Expression
			// Add equality conditions for all columns before current
			for j := 0; j < i; j++ {
				subConditions = append(subConditions, goqu.T(rootAlias).Col(orderBy[j]).Eq(goqu.L("?", 0)))
			}
			// Add greater than condition for current column
			subConditions = append(subConditions, goqu.T(rootAlias).Col(orderBy[i]).Gt(goqu.L("?", 0)))
			conditions = append(conditions, goqu.And(subConditions...))
		}
		query = query.Where(goqu.Or(conditions...))
	}
	return query.Prepared(true)
}

// addShortestPathJoins uses BFS to find the shortest join chains (per where condition) and adds them to the query.
func (qb *QueryBuilder) addShortestPathJoins(query *goqu.SelectDataset, rootTable *TableInfo, rootAlias string) (*goqu.SelectDataset, error) {
	// joinPaths := qb.findShortestJoinPaths(rootTable)
	subsets := rootTable.SubsetPaths
	fmt.Println("rootTable", rootTable.Name)

	jsonF, _ := json.MarshalIndent(subsets, "", " ")
	fmt.Printf("\n\n %s \n\n", string(jsonF))

	// We'll union all join steps from all shortest join chains.
	// tableAliasMap will hold the alias assigned for each table key.
	tableAliasMap := map[string]string{
		qb.getTableKey(rootTable.Schema, rootTable.Name): rootAlias,
	}
	// usedAliases is used by generateUniqueAlias to ensure uniqueness.
	usedAliases := make(set)

	// Create a slice of join steps with their depth (position in the join chain).
	type joinChainEntry struct {
		depth int
		step  runconfigs.JoinStep
	}
	var allSteps []joinChainEntry
	for _, subset := range subsets {
		for i, step := range subset.JoinSteps {
			allSteps = append(allSteps, joinChainEntry{depth: i + 1, step: step})
		}
	}
	// Sort by depth (shorter chains first)
	sort.Slice(allSteps, func(i, j int) bool {
		return allSteps[i].depth < allSteps[j].depth
	})

	fmt.Println("allSteps", len(allSteps))

	// To avoid adding duplicate joins, track them using a key "fromKey->toKey"
	addedJoins := make(map[string]bool)
	for _, entry := range allSteps {
		step := entry.step
		edgeKey := step.FromKey + "->" + step.ToKey
		if addedJoins[edgeKey] {
			continue
		}
		// Ensure the parent (fromKey) already has an alias.
		parentAlias, ok := tableAliasMap[step.FromKey]
		if !ok {
			continue
		}
		// Retrieve the child table.
		childTable, ok := qb.tables[step.ToKey]
		if !ok {
			continue
		}
		// If we haven’t already assigned an alias for the child, do so now.
		childAlias, exists := tableAliasMap[step.ToKey]
		if !exists {
			// Use the parent’s alias as a prefix.
			prefix := parentAlias + "_"
			childAlias = qb.generateUniqueAlias(prefix, childTable.Name, usedAliases)
			tableAliasMap[step.ToKey] = childAlias
		}
		// Build join conditions based on the foreign key.
		joinConditions := make([]exp.Expression, len(step.Fk.Columns))
		for i, col := range step.Fk.Columns {
			joinConditions[i] = goqu.T(childAlias).Col(step.Fk.ReferenceColumns[i]).Eq(goqu.T(parentAlias).Col(col))
		}
		query = query.InnerJoin(
			childTable.GetIdentifierExpression().As(childAlias),
			goqu.On(joinConditions...),
		)
		addedJoins[edgeKey] = true

		refKey := qb.getTableKey(step.Fk.ReferenceSchema, step.Fk.ReferenceTable)
		// Also, if the joined table has WHERE conditions, add them.
		if conditions, ok := qb.whereConditions[refKey]; ok {
			for _, cond := range conditions {
				qualifiedCondition, err := qb.qualifyWhereCondition(nil, childAlias, cond.Condition)
				if err != nil {
					return nil, err
				}
				query = query.Where(goqu.L(qualifiedCondition, cond.Args...))
			}
		}
	}
	return query, nil
}

// // findShortestJoinPaths performs a BFS over foreign key relationships (from the root table)
// // and returns, for each table that has a where condition (other than the root), the shortest join chain (as a slice of joinSteps).
// func (qb *QueryBuilder) findShortestJoinPaths(rootTable *TableInfo) map[string][]joinStep {
// 	type queueEntry struct {
// 		tableKey string
// 		path     []joinStep
// 	}
// 	queue := []queueEntry{
// 		{tableKey: qb.getTableKey(rootTable.Schema, rootTable.Name), path: []joinStep{}},
// 	}
// 	visited := map[string]int{
// 		qb.getTableKey(rootTable.Schema, rootTable.Name): 0,
// 	}
// 	joinPaths := make(map[string][]joinStep) // keys for tables with where conditions (other than root)
// 	for len(queue) > 0 {
// 		current := queue[0]
// 		queue = queue[1:]
// 		currentTable, exists := qb.tables[current.tableKey]
// 		if !exists {
// 			continue
// 		}
// 		for _, fk := range currentTable.ForeignKeys {
// 			neighborKey := qb.getTableKey(fk.ReferenceSchema, fk.ReferenceTable)
// 			// Skip if the neighbor table does not exist.
// 			if _, exists := qb.tables[neighborKey]; !exists {
// 				continue
// 			}
// 			// Build the new path by appending this join step.
// 			newPath := make([]joinStep, len(current.path))
// 			copy(newPath, current.path)
// 			newPath = append(newPath, joinStep{
// 				fromKey: current.tableKey,
// 				toKey:   neighborKey,
// 				fk:      fk,
// 			})
// 			distance := len(newPath)
// 			if prev, ok := visited[neighborKey]; !ok || distance < prev {
// 				visited[neighborKey] = distance
// 				queue = append(queue, queueEntry{tableKey: neighborKey, path: newPath})
// 				// If this neighbor has a WHERE condition, record its join chain.
// 				if qb.tablesWithWhereConditions.contains(neighborKey) {
// 					joinPaths[neighborKey] = newPath
// 				}
// 			}
// 		}
// 	}
// 	return joinPaths
// }

// generateUniqueAlias produces a short alias from a prefix and table name.
func (qb *QueryBuilder) generateUniqueAlias(prefix, tableName string, joinedTables map[string]struct{}) string {
	qb.aliasCounter++
	baseString := fmt.Sprintf("%s%s%d", prefix, tableName, qb.aliasCounter)
	alias := "t_" + getClippedHash(baseString)

	// Ensure uniqueness
	for {
		if _, exists := joinedTables[alias]; !exists {
			joinedTables[alias] = struct{}{}
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
