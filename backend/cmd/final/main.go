package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type ForeignKey struct {
	Table         string
	Column        string
	OriginalTable *string
}

type FkColumnDefinition struct {
	Columns    []string
	ForeignKey ForeignKey
}
type QueryColumnDefinition struct {
	Column     string
	ForeignKey ForeignKey
}

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

func formatSqlQuery(sql string) string {
	return fmt.Sprintf("%s;", sql)
}

func buildSqlIdentifier(identifiers ...string) string {
	return strings.Join(identifiers, ".")
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
		if j.Alias != nil && *j.Alias != "" {
			joinCondition := goqu.Ex{}
			for joinCol, baseCol := range j.JoinColumnsMap {
				joinCondition[buildSqlIdentifier(*j.Alias, joinCol)] = goqu.I(buildSqlIdentifier(j.BaseTable, baseCol))
			}
			if j.JoinType == innerJoin {
				joinTable := goqu.I(j.JoinTable).As(*j.Alias)
				query = query.InnerJoin(
					joinTable,
					goqu.On(joinCondition),
				)
			}
		} else {
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

func buildQuery(table string, data map[string][]QueryColumnDefinition, whereClauses map[string]string, visited map[string]bool) ([]*sqlJoin, []string) {
	joins := []*sqlJoin{}
	wheres := []string{}

	if condition, exists := whereClauses[table]; exists {
		wheres = append(wheres, condition)
		return joins, wheres
	}

	if columns, exists := data[table]; exists {
		for _, col := range columns {
			if col.ForeignKey.Table != "" && col.ForeignKey.Column != "" {
				var alias *string
				joinTable := col.ForeignKey.Table
				if col.ForeignKey.OriginalTable != nil && *col.ForeignKey.OriginalTable != "" {
					alias = &col.ForeignKey.Table
					joinTable = *col.ForeignKey.OriginalTable
				}
				joins = append(joins, &sqlJoin{
					JoinType:  innerJoin,
					JoinTable: joinTable,
					BaseTable: table,
					Alias:     alias,
					JoinColumnsMap: map[string]string{
						col.ForeignKey.Column: col.Column,
					},
				})

				if !visited[col.ForeignKey.Table] {
					visited[col.ForeignKey.Table] = true
					subQuery, subWheres := buildQuery(col.ForeignKey.Table, data, whereClauses, visited)

					joins = append(joins, subQuery...)
					wheres = append(wheres, subWheres...)
				}
			}
		}
	}

	return joins, wheres
}

func SubsetQueries(data map[string][]QueryColumnDefinition, whereClauses map[string]string) {
	// fmt.Println()
	// fmt.Println()
	for table := range data {
		visited := make(map[string]bool)
		joins, wheres := buildQuery(table, data, whereClauses, visited)
		fmt.Println()
		fmt.Println()
		// fmt.Printf("table: %s  -------------\n", table)
		// jsonF, _ := json.MarshalIndent(joins, "", " ")
		// fmt.Printf("joins: %s \n", string(jsonF))

		// jsonF, _ = json.MarshalIndent(wheres, "", " ")
		// fmt.Printf("wheres: %s \n", string(jsonF))
		// fmt.Println()
		split := strings.Split(table, ".")
		schema := split[0]
		var table string
		if len(split) > 1 {
			table = split[1]
		}
		if len(joins) == 0 {
			where := strings.Join(wheres, " AND ")
			fmt.Println(buildSelectQuery("postgres", schema, table, []string{"*"}, &where))
		} else {
			fmt.Println(buildSelectJoinQuery("postgres", schema, table, []string{"*"}, joins, wheres))
		}

	}
}

func handleDoubleReferences(data map[string][]FkColumnDefinition, whereClauses map[string]string) (map[string][]QueryColumnDefinition, map[string]string) {
	newData := make(map[string][]QueryColumnDefinition)
	reverseReferences := make(map[string]string) // alias name to table name
	newWhereClauses := map[string]string{}

	for table, where := range whereClauses {
		newWhereClauses[table] = where
	}

	for table, colDefs := range data {
		if len(colDefs) == 0 {
			newData[table] = []QueryColumnDefinition{}
			continue
		}
		for _, colDef := range colDefs {
			if colDef.ForeignKey.Table == table {
				// self reference skipp
				newData[table] = []QueryColumnDefinition{}
				continue
			}
			if len(colDef.Columns) > 1 {
				for _, col := range colDef.Columns {
					newTable := fmt.Sprintf("%s_%s", strings.Replace(colDef.ForeignKey.Table, ".", "_", -1), col)
					alias := ToSha256(newTable)
					newData[table] = append(newData[table], QueryColumnDefinition{
						Column: col,
						ForeignKey: ForeignKey{
							Table:         alias,
							OriginalTable: &colDef.ForeignKey.Table,
							Column:        colDef.ForeignKey.Column,
						},
					})
					reverseReferences[alias] = colDef.ForeignKey.Table
				}
			} else {
				newData[table] = append(newData[table], QueryColumnDefinition{
					Column:     colDef.Columns[0],
					ForeignKey: colDef.ForeignKey,
				})
			}
		}
	}
	jsonF, _ := json.MarshalIndent(whereClauses, "", " ")
	fmt.Printf("\n whereClauses: %s \n", string(jsonF))
	jsonF, _ = json.MarshalIndent(reverseReferences, "", " ")
	fmt.Printf("\n reverseReferences: %s \n", string(jsonF))

	for len(reverseReferences) > 0 {
		for alias, table := range reverseReferences {
			newColDefs := []QueryColumnDefinition{}
			colDefs := newData[table]
			where := newWhereClauses[table]
			if where != "" {
				newWhereClauses[alias] = where
			}
			for _, c := range colDefs {
				newAlias := ToSha256(fmt.Sprintf("%s_%s", alias, strings.Replace(c.ForeignKey.Table, ".", "_", -1)))
				reverseReferences[newAlias] = c.ForeignKey.Table
				newColDefs = append(newColDefs, QueryColumnDefinition{
					Column: c.Column,
					ForeignKey: ForeignKey{
						Column:        c.ForeignKey.Column,
						OriginalTable: &c.ForeignKey.Table,
						Table:         newAlias,
					},
				})
			}
			newData[alias] = newColDefs
			delete(reverseReferences, alias)
		}
	}
	jsonF, _ = json.MarshalIndent(newWhereClauses, "", " ")
	fmt.Printf("\n newWhereClauses: %s \n", string(jsonF))
	jsonF, _ = json.MarshalIndent(newData, "", " ")
	fmt.Printf("\n newData: %s \n", string(jsonF))

	return newData, newWhereClauses
}

func checkParentWhereClause(table string, data map[string][]FkColumnDefinition, whereClauses map[string]string, visited map[string]bool) bool {
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
				if checkParentWhereClause(col.ForeignKey.Table, data, whereClauses, visited) {
					return true
				}
			}
		}
	}
	return false
}

func ForeignKeysWithSubset(data map[string][]FkColumnDefinition, whereClauses map[string]string) map[string][]FkColumnDefinition {
	tableSubsetMap := make(map[string]bool)
	for table := range data {
		visited := make(map[string]bool)
		tableSubsetMap[table] = checkParentWhereClause(table, data, whereClauses, visited)
	}
	newData := map[string][]FkColumnDefinition{}
	for table, colDefs := range data {
		newData[table] = []FkColumnDefinition{}
		for _, colDef := range colDefs {
			if exists := tableSubsetMap[colDef.ForeignKey.Table]; exists {
				newData[table] = append(newData[table], colDef)
			}
		}
	}
	return newData
}
func ToSha256(input string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
	if len(hash) > 14 {
		hash = hash[:14]
	}
	return hash
}

func main() {
	// ###################################################
	// queryData := map[string][]FkColumnDefinition{
	// 	"public.a": {
	// 		{
	// 			Columns: []string{"x_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.x",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.c": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.a",
	// 				Column: "id",
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"b_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.b",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.d": {
	// 		{
	// 			Columns: []string{"c_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.c",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.e": {
	// 		{
	// 			Columns: []string{"c_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.c",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.x": {},
	// 	"public.b": {},
	// 	"public.z": {},
	// }

	// whereClauses := map[string]string{
	// 	"public.b": "public.b.id = '1'",
	// 	// "public.x": "public.x.id = '2'",
	// }

	// ###################################################
	// double reference
	// queryData := map[string][]FkColumnDefinition{
	// 	"public.a": {},
	// 	"public.b": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.a",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.c": {
	// 		{
	// 			Columns: []string{"b_id", "bb_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.b",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// }
	// whereClauses := map[string]string{
	// 	"public.a": "public.a.id = '1'",
	// }
	// ###################################################

	// circular dependency
	// queryData := map[string][]FkColumnDefinition{
	// 	"public.a": {
	// 		{
	// 			Columns: []string{"c_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.c",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.c": {
	// 		{
	// 			Columns: []string{"b_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.b",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.b": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.a",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"public.a": "public.a.id = '1'",
	// }

	// ###################################################

	// double circular dependency
	// queryData := map[string][]FkColumnDefinition{
	// 	"public.a": {
	// 		{
	// 			Columns: []string{"a_id", "aa_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.a",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"public.b": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "public.a",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"public.a": "public.a.id = '1'",
	// }

	// ###################################################

	// queryData := map[string][]FkColumnDefinition{
	// 	"department": {
	// 		{
	// 			Columns: []string{"company_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "company",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"transaction": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "department",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// 	"expense": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "department",
	// 				Column: "id",
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"transaction_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:  "transaction",
	// 				Column: "id",
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"transaction": "transaction.id = '1'",
	// }

	// ###################################################

	queryData := map[string][]FkColumnDefinition{

		"company": {},
		"department": {
			{
				Columns: []string{"company_id"},
				ForeignKey: ForeignKey{
					Table:  "company",
					Column: "id",
				},
			},
		},
		"transaction": {
			{
				Columns: []string{"department_id"},
				ForeignKey: ForeignKey{
					Table:  "department",
					Column: "id",
				},
			},
		},
		"expense_report": {
			{
				Columns: []string{"destination_transaction_id", "source_transaction_id"},
				ForeignKey: ForeignKey{
					Table:  "transaction",
					Column: "id",
				},
			},
		},
		"expense": {
			{
				Columns: []string{"expense_report_id"},
				ForeignKey: ForeignKey{
					Table:  "expense_report",
					Column: "id",
				},
			},
		},
	}

	whereClauses := map[string]string{
		"company": "company.id = '1'",
	}

	filteredData := ForeignKeysWithSubset(queryData, whereClauses)
	// jsonF, _ := json.MarshalIndent(filteredData, "", " ")
	// fmt.Printf("\n filteredData: %s \n", string(jsonF))

	newData, newWhereClauses := handleDoubleReferences(filteredData, whereClauses)
	// jsonF, _ = json.MarshalIndent(newData, "", " ")
	// fmt.Printf("\n newData: %s \n", string(jsonF))

	SubsetQueries(newData, newWhereClauses)
}
