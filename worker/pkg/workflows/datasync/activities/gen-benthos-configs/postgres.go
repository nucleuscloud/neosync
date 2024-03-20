package genbenthosconfigs_activity

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
)

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
