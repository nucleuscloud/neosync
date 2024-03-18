package genbenthosconfigs_activity

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func qualifyMysqlWhereColumnNames(where, schema, table string) (string, error) {
	sqlSelect := fmt.Sprintf("select * from %s where ", buildSqlIdentifier(schema, table))
	sql := fmt.Sprintf("%s%s", sqlSelect, where)
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", err
	}
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) {
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
	}

	updatedSql := sqlparser.String(stmt)
	updatedWhere := strings.Replace(updatedSql, sqlSelect, "", 1)
	return updatedWhere, nil
}
