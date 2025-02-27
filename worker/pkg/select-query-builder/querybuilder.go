package selectquerybuilder

import (
	"crypto/md5" //nolint:gosec // This is not being used for a purpose that requires security
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/doug-martin/goqu/v9/exp"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	tsql_parser "github.com/nucleuscloud/neosync/worker/pkg/select-query-builder/tsql"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	"github.com/xwb1989/sqlparser"
)

// QueryBuilder holds global state for building the queries.
type QueryBuilder struct {
	defaultSchema                 string
	driver                        string
	subsetByForeignKeyConstraints bool
	aliasCounter                  int
	pageLimit                     uint
}

func NewSelectQueryBuilder(defaultSchema, driver string, subsetByForeignKeyConstraints bool, pageLimit int) *QueryBuilder {
	limit := uint(0)
	if pageLimit > 0 {
		limit = uint(pageLimit)
	}
	return &QueryBuilder{
		defaultSchema:                 defaultSchema,
		driver:                        driver,
		subsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
		aliasCounter:                  0,
		pageLimit:                     limit,
	}
}

// getDialect returns the dialect for the given driver.
func (qb *QueryBuilder) getDialect() goqu.DialectWrapper {
	switch qb.driver {
	case sqlmanager_shared.PostgresDriver:
		return goqu.Dialect(sqlmanager_shared.GoquPostgresDriver)
	default:
		return goqu.Dialect(qb.driver)
	}
}

// BuildQuery constructs a SQL Select query from a RunConfig, returning the query string,
// returns initial select and paged select queries, a flag indicating foreign key safety
func (qb *QueryBuilder) BuildQuery(runconfig *runconfigs.RunConfig) (sqlstatement string, args []any, pagesql string, isNotForeignKeySafeSubset bool, err error) {
	query, pageQuery, notFkSafe, err := qb.buildFlattenedQuery(runconfig)
	if query == nil {
		return "", nil, "", false, fmt.Errorf("received no error, but query was nil for %s", runconfig.Id())
	}
	if err != nil {
		return "", nil, "", false, err
	}

	sql, args, err := query.Limit(qb.pageLimit).ToSQL()
	if err != nil {
		return "", nil, "", false, fmt.Errorf("unable to convert structured query to string for %s: %w", runconfig.Id(), err)
	}

	pageSql, _, err := pageQuery.Limit(qb.pageLimit).ToSQL()
	if err != nil {
		return "", nil, "", false, fmt.Errorf("unable to convert structured page query to string for %s: %w", runconfig.Id(), err)
	}
	if qb.driver == sqlmanager_shared.MssqlDriver {
		// MSSQL TOP needs to be cast to int
		pageSql = strings.Replace(pageSql, "TOP (@p1)", "TOP (CAST(@p1 AS INT))", 1)
	}
	return sql, args, pageSql, notFkSafe, nil
}

// buildFlattenedQuery builds the query for the root table, adding joins if needed.
func (qb *QueryBuilder) buildFlattenedQuery(rootTable *runconfigs.RunConfig) (sql, pageSql *goqu.SelectDataset, isNotForeignKeySafeSubset bool, err error) {
	dialect := qb.getDialect()
	rootAlias := rootTable.SchemaTable().Table
	rootAliasExpression := goqu.S(rootTable.SchemaTable().Schema).Table(rootTable.SchemaTable().Table).As(rootAlias)
	query := dialect.From(rootAliasExpression)

	// Select columns for the root table
	cols := make([]exp.Expression, len(rootTable.SelectColumns()))
	for i, col := range rootTable.SelectColumns() {
		cols[i] = rootAliasExpression.Col(col)
	}
	query = query.Select(toAnySlice(cols)...)

	// Add order by if specified
	if len(rootTable.OrderByColumns()) > 0 {
		orderByExpressions := make([]exp.OrderedExpression, len(rootTable.OrderByColumns()))
		for i, col := range rootTable.OrderByColumns() {
			orderByExpressions[i] = rootAliasExpression.Col(col).Asc()
		}
		query = query.Order(orderByExpressions...)
	}

	// If using subset-by-foreign-key constraints, add joins using the subset path
	var notFkSafeSubset bool
	if qb.subsetByForeignKeyConstraints && len(rootTable.SubsetPaths()) > 0 {
		query, notFkSafeSubset, err = qb.addSubsetJoins(query, rootTable, rootAlias)
		if err != nil {
			return nil, nil, false, err
		}
	} else if !qb.subsetByForeignKeyConstraints && rootTable.WhereClause() != nil && *rootTable.WhereClause() != "" {
		// No subset-by-foreign-key constraints, but a where clause was provided
		qualifiedCondition, err := qb.qualifyWhereCondition(nil, rootAlias, *rootTable.WhereClause())
		if err != nil {
			return nil, nil, false, err
		}
		query = query.Where(goqu.L(qualifiedCondition))
	}

	// Build page query (for pagination)
	pageQuery := qb.buildPageQuery(query, rootAlias, rootTable.OrderByColumns())

	return query, pageQuery, notFkSafeSubset, nil
}

// buildPageQuery builds a pagination version of the query.
func (qb *QueryBuilder) buildPageQuery(query *goqu.SelectDataset, rootAlias string, orderByColumns []string) *goqu.SelectDataset {
	if len(orderByColumns) > 0 {
		// Build lexicographical ordering conditions
		var conditions []exp.Expression
		for i := 0; i < len(orderByColumns); i++ {
			var subConditions []exp.Expression
			// Add equality conditions for all columns before current
			for j := 0; j < i; j++ {
				subConditions = append(subConditions, goqu.T(rootAlias).Col(orderByColumns[j]).Eq(goqu.L("?", 0)))
			}
			// Add greater than condition for current column
			subConditions = append(subConditions, goqu.T(rootAlias).Col(orderByColumns[i]).Gt(goqu.L("?", 0)))
			conditions = append(conditions, goqu.And(subConditions...))
		}
		query = query.Where(goqu.Or(conditions...))
	}
	return query.Prepared(true)
}

// addSubsetJoins adds joins to the query based on foreign key relationships defined in the subset paths.
// returns the modified query, a boolean indicating if the subset is not foreign key safe.
func (qb *QueryBuilder) addSubsetJoins(query *goqu.SelectDataset, rootTable *runconfigs.RunConfig, rootAlias string) (*goqu.SelectDataset, bool, error) {
	subsets := rootTable.SubsetPaths()
	isSubset := false

	// We'll union all join steps from all shortest join chains.
	// tableAliasMap will hold the alias assigned for each table key.
	tableAliasMap := map[string]string{
		rootTable.Table(): rootAlias,
	}

	// To avoid adding duplicate joins, track them using a key "fromKey->toKey"
	addedJoins := make(map[string]bool)
	for _, subset := range subsets {
		// If there are no join steps, and there is a subset condition, apply it
		// This handles case where the root table has a where clause
		if len(subset.JoinSteps) == 0 && subset.Subset != "" {
			qualifiedCondition, err := qb.qualifyWhereCondition(nil, rootAlias, subset.Subset)
			if err != nil {
				return nil, false, err
			}
			query = query.Where(goqu.L(qualifiedCondition))
			isSubset = rootTable.RunType() == runconfigs.RunTypeUpdate
			continue
		}

		for idx, step := range subset.JoinSteps {

			edgeKey := step.FromKey + "->" + step.ToKey
			if addedJoins[edgeKey] {
				continue
			}
			// Ensure the parent (fromKey) already has an alias.
			parentAlias, ok := tableAliasMap[step.FromKey]
			if !ok {
				continue
			}

			childTable := step.ToKey
			// If we haven’t already assigned an alias for the child, do so now.
			childAlias, exists := tableAliasMap[step.ToKey]
			if !exists {
				// Use the parent’s alias as a prefix.
				prefix := parentAlias + "_"
				childAlias = qb.generateUniqueAlias(prefix, childTable)
				tableAliasMap[step.ToKey] = childAlias
			}
			// Build join conditions based on the foreign key.
			joinConditions := make([]exp.Expression, len(step.ForeignKey.Columns))
			for i, col := range step.ForeignKey.Columns {
				joinConditions[i] = goqu.T(childAlias).Col(step.ForeignKey.ReferenceColumns[i]).Eq(goqu.T(parentAlias).Col(col))
			}
			query = query.InnerJoin(
				goqu.I(childTable).As(childAlias),
				goqu.On(joinConditions...),
			)
			addedJoins[edgeKey] = true

			// If this is the last step in chain and there's a subset condition, apply it
			if idx == len(subset.JoinSteps)-1 && subset.Subset != "" {
				isSubset = true
				qualifiedCondition, err := qb.qualifyWhereCondition(nil, childAlias, subset.Subset)
				if err != nil {
					return nil, false, err
				}
				query = query.Where(goqu.L(qualifiedCondition))
			}
		}
	}
	return query, isSubset, nil
}

// generateUniqueAlias produces a short alias from a prefix and table name.
func (qb *QueryBuilder) generateUniqueAlias(prefix, tableName string) string {
	qb.aliasCounter++
	baseString := fmt.Sprintf("%s%s%d", prefix, tableName, qb.aliasCounter)
	return "t_" + getClippedHash(baseString)
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

func toAnySlice[T any](input []T) []any {
	anys := make([]any, len(input))
	for i := range input {
		anys[i] = input[i]
	}
	return anys
}
