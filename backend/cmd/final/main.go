package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"

	"github.com/xwb1989/sqlparser"
)

type ForeignKey struct {
	Table         string
	Columns       []string
	OriginalTable *string
}

type FkColumnDefinition struct {
	Columns    []string
	ForeignKey ForeignKey
}
type QueryColumnDefinition struct {
	Columns    []string
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
			if col.ForeignKey.Table != "" && col.ForeignKey.Columns != nil {
				var alias *string
				joinTable := col.ForeignKey.Table
				if col.ForeignKey.OriginalTable != nil && *col.ForeignKey.OriginalTable != "" {
					alias = &col.ForeignKey.Table
					joinTable = *col.ForeignKey.OriginalTable
				}
				// jsonF, _ := json.MarshalIndent(col, "", " ")
				// fmt.Printf("\n col: %s \n", string(jsonF))
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
			sql, _ := buildSelectQuery("postgres", schema, table, []string{"*"}, &where)
			fmt.Println(sql)
		} else {
			sql, _ := buildSelectJoinQuery("postgres", schema, table, []string{"*"}, joins, wheres)
			fmt.Println(sql)
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

	seenTables := map[string]struct{}{}
	for table, colDefs := range data {
		if len(colDefs) == 0 {
			newData[table] = []QueryColumnDefinition{}
			continue
		}

		for _, colDef := range colDefs {
			if colDef.ForeignKey.Table == table {
				// self reference skip
				newData[table] = []QueryColumnDefinition{}
				continue
			}
			// seenTables := map[string]struct{}{}

			// this check is no longer true
			// check if this already exists in newdata if so create an alias
			// multiRefCols := getMultiReferenceColumns(colDefs, colDef.ForeignKey.Table, colDef.Columns)
			if _, exists := seenTables[colDef.ForeignKey.Table]; exists {
				newTable := fmt.Sprintf("%s_%s", strings.Replace(colDef.ForeignKey.Table, ".", "_", -1), strings.Join(colDef.Columns, "_"))
				alias := ToSha256(newTable)
				newData[table] = append(newData[table], QueryColumnDefinition{
					Columns: colDef.Columns, // how to tell difference between double reference and composite keys
					ForeignKey: ForeignKey{
						Table:         alias,
						OriginalTable: &colDef.ForeignKey.Table,
						Columns:       colDef.ForeignKey.Columns,
					},
				})
				reverseReferences[alias] = colDef.ForeignKey.Table
			} else {
				newData[table] = append(newData[table], QueryColumnDefinition{
					Columns:    colDef.Columns,
					ForeignKey: colDef.ForeignKey,
				})
				seenTables[colDef.ForeignKey.Table] = struct{}{}
			}
		}
	}
	// jsonF, _ := json.MarshalIndent(whereClauses, "", " ")
	// fmt.Printf("\n whereClauses: %s \n", string(jsonF))
	// jsonF, _ := json.MarshalIndent(newData, "", " ")
	// fmt.Printf("\n newData: %s \n", string(jsonF))

	for len(reverseReferences) > 0 {
		for alias, table := range reverseReferences {
			newColDefs := []QueryColumnDefinition{}
			colDefs := newData[table]
			where := newWhereClauses[table]
			// fmt.Printf("old where: %s \n", where)

			if where != "" {
				// fmt.Printf("alias: %s \n", alias)

				newWhere, err := qualifyWhereWithTableAlias("postgres", where, alias)
				if err != nil {
					fmt.Println(err)
				}
				// fmt.Printf("new where: %s \n", newWhere)
				newWhereClauses[alias] = newWhere
			}
			for _, c := range colDefs {
				newAlias := ToSha256(fmt.Sprintf("%s_%s", alias, strings.Replace(c.ForeignKey.Table, ".", "_", -1)))
				reverseReferences[newAlias] = c.ForeignKey.Table
				newColDefs = append(newColDefs, QueryColumnDefinition{
					Columns: c.Columns,
					ForeignKey: ForeignKey{
						Columns:       c.ForeignKey.Columns,
						OriginalTable: &c.ForeignKey.Table,
						Table:         newAlias,
					},
				})
			}
			newData[alias] = newColDefs
			delete(reverseReferences, alias)
		}
	}
	// jsonF, _ = json.MarshalIndent(newWhereClauses, "", " ")
	// fmt.Printf("\n newWhereClauses: %s \n", string(jsonF))
	jsonF, _ := json.MarshalIndent(newData, "", " ")
	fmt.Printf("\n newData after: %s \n", string(jsonF))

	return newData, newWhereClauses
}

func getMultiReferenceColumns(foreignKeys []FkColumnDefinition, table string, cols []string) [][]string {
	// fmt.Println("------------")
	// fmt.Printf("table: %s \n", table)
	// fmt.Printf("cols: %v \n", cols)
	refCols := [][]string{}
	for _, fk := range foreignKeys {
		if fk.ForeignKey.Table != table {
			continue
		}
		if len(fk.ForeignKey.Columns) != len(cols) {
			continue
		}
		for idx, c := range fk.ForeignKey.Columns {
			if c != cols[idx] {
				continue
			}
		}
		refCols = append(refCols, fk.Columns)
	}
	// jsonF, _ := json.MarshalIndent(refCols, "", " ")
	// fmt.Printf("\n refCols: %s \n", string(jsonF))
	return refCols
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
	// 				Table:   "public.x",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.c": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.a",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"b_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.b",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.d": {
	// 		{
	// 			Columns: []string{"c_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.c",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.e": {
	// 		{
	// 			Columns: []string{"c_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.c",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.x": {},
	// 	"public.b": {},
	// 	"public.z": {},
	// }

	// whereClauses := map[string]string{
	// 	"public.b": "public.b.id = '1'",
	// 	"public.x": "public.x.id = '2'",
	// }

	// ###################################################

	// circular dependency
	// queryData := map[string][]FkColumnDefinition{
	// 	"public.a": {
	// 		{
	// 			Columns: []string{"c_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.c",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.c": {
	// 		{
	// 			Columns: []string{"b_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.b",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.b": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.a",
	// 				Columns: []string{"id"},
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
	// 			Columns: []string{"aa_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.a",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.a",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"public.b": {
	// 		{
	// 			Columns: []string{"a_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "public.a",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"public.a": "public.a.id = '1'",
	// }

	// ###################################################
	// double reference more complex
	// queryData := map[string][]FkColumnDefinition{
	// 	"company": {},
	// 	"department": {
	// 		{
	// 			Columns: []string{"company_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "company",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"transaction": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "department",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"expense_report": {
	// 		{
	// 			Columns: []string{"destination_transaction_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "transaction",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"source_transaction_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "transaction",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"expense": {
	// 		{
	// 			Columns: []string{"expense_report_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "expense_report",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"company": "company.id = '1'",
	// }

	// ###################################################
	// double reference more complex even more complex
	// queryData := map[string][]FkColumnDefinition{
	// 	"company": {},
	// 	"department": {
	// 		{
	// 			Columns: []string{"company_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "company",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"transaction": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "department",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// 	"expense_report": {
	// 		{
	// 			Columns: []string{"department_destination_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "department",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"department_source_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "department",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"transaction_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "transaction",
	// 				Columns: []string{"id"},
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"company": "company.id = '1'",
	// }

	// ##################################################
	// nested composite keys

	// queryData := map[string][]FkColumnDefinition{
	// 	"composite_keys.department": {},
	// 	"composite_keys.employees": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.department",
	// 				Columns: []string{"department_id"},
	// 			},
	// 		},
	// 	},
	// 	"composite_keys.projects": {
	// 		{
	// 			Columns: []string{"responsible_department_id", "responsible_employee_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.employees",
	// 				Columns: []string{"department_id", "employee_id"},
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"composite_keys.department": "composite_keys.department.department_id = '1'",
	// }

	// #############################################
	// separate multiple foreigkeys same table

	// queryData := map[string][]FkColumnDefinition{
	// 	"composite_keys.department": {},
	// 	"composite_keys.employees": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.department",
	// 				Columns: []string{"department_id"},
	// 			},
	// 		},
	// 	},
	// 	"composite_keys.projects": {
	// 		{
	// 			Columns: []string{"responsible_department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.employees",
	// 				Columns: []string{"department_id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"responsible_employee_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.employees",
	// 				Columns: []string{"employee_id"},
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"composite_keys.department": "composite_keys.department.department_id = '1'",
	// }

	// #############################################
	// separate multiple foreigkeys different tables

	// queryData := map[string][]FkColumnDefinition{
	// 	"composite_keys.department": {},
	// 	"composite_keys.employees": {
	// 		{
	// 			Columns: []string{"department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.department",
	// 				Columns: []string{"department_id"},
	// 			},
	// 		},
	// 	},
	// 	"composite_keys.projects": {
	// 		{
	// 			Columns: []string{"responsible_department_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.department",
	// 				Columns: []string{"department_id"},
	// 			},
	// 		},
	// 		{
	// 			Columns: []string{"responsible_employee_id"},
	// 			ForeignKey: ForeignKey{
	// 				Table:   "composite_keys.employees",
	// 				Columns: []string{"employee_id"},
	// 			},
	// 		},
	// 	},
	// }

	// whereClauses := map[string]string{
	// 	"composite_keys.department": "composite_keys.department.department_id = '1'",
	// }

	filteredData := ForeignKeysWithSubset(queryData, whereClauses)
	// jsonF, _ := json.MarshalIndent(filteredData, "", " ")
	// fmt.Printf("\n filteredData: %s \n", string(jsonF))

	newData, newWhereClauses := handleDoubleReferences(filteredData, whereClauses)
	// jsonF, _ = json.MarshalIndent(newData, "", " ")
	// fmt.Printf("\n newData: %s \n", string(jsonF))

	SubsetQueries(newData, newWhereClauses)
}

/// NOTES
/*

double reference will show up twice
each should be its own join with alias

map[string][]FkColumnDefinition{
	"projects": {
		{
			Columns: []string{"other_id"},
			ForeignKey: ForeignKey{
				Table:   "employees",
				Columns: []string{"id"},
			},
		},
		{
			Columns: []string{"employee_id"},
			ForeignKey: ForeignKey{
				Table:   "employees",
				Columns: []string{"id"},
			},
		},
	},
}


composite key will have multiple cols in col slice
one join but all keys are included in the ON clause of join
map[string][]FkColumnDefinition{
	"projects": {
		{
			Columns: []string{"empolyee_id", "department_id"},
			ForeignKey: ForeignKey{
				Table:   "employees",
				Columns: []string{"id", "department_id"},
			},
		},
	},
}

multiple but separate foreign keys. each should have a join
map[string][]FkColumnDefinition{
	"projects": {
		{
			Columns: []string{"department_id"},
			ForeignKey: ForeignKey{
				Table:   "employees",
				Columns: []string{"department_id"},
			},
		},
		{
			Columns: []string{"employee_id"},
			ForeignKey: ForeignKey{
				Table:   "employees",
				Columns: []string{"id"},
			},
		},
	},
}


*/

func qualifyWhereWithTableAlias(driver, where, alias string) (string, error) {
	query := goqu.Dialect(driver).From(goqu.T(alias)).Select("*").Where(goqu.L(where))
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}
	var updatedSql string
	switch driver {
	case "mysql":
		sql, err := qualifyMysqlWhereColumnNames(sql, nil, alias)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case "postgres":
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

type selfReferencingCircularDependency struct {
	PrimaryKeyColumn  string
	ForeignKeyColumns []string
}

func formatSqlQuery(sql string) string {
	return fmt.Sprintf("%s;", sql)
}

func buildSqlIdentifier(identifiers ...string) string {
	return strings.Join(identifiers, ".")
}

func reverseSlice[T any](slice []T) []T {
	newSlice := make([]T, len(slice))
	copy(newSlice, slice)

	for i, j := 0, len(newSlice)-1; i < j; i, j = i+1, j-1 {
		newSlice[i], newSlice[j] = newSlice[j], newSlice[i]
	}

	return newSlice
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
