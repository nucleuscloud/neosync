package genbenthosconfigs_activity

import (
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
)

func qualifyPostgresWhereColumnNames(where, schema, table string) (string, error) {
	sqlSelect := fmt.Sprintf("SELECT * FROM %s WHERE ", buildSqlIdentifier(schema, table))
	sql := fmt.Sprintf("%s%s", sqlSelect, where)
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
	updatedWhere := strings.Replace(updatedSql, sqlSelect, "", 1)
	return updatedWhere, nil
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
		if len(col.Fields) > 0 {
			for idx, f := range col.Fields {
				colName := f.GetString_().GetSval()
				col.Fields[idx] = pg_query.MakeStrNode(fmt.Sprintf("%s.%s.%s", schema, table, colName))
			}
		}
	}
}
