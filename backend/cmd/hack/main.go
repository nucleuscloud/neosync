package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ForeignKey struct {
	Table  string `json:"Table"`
	Column string `json:"Column"`
}

type ColumnDefinition struct {
	Column     string     `json:"Column"`
	ForeignKey ForeignKey `json:"ForeignKey"`
}

type QueryData map[string][]ColumnDefinition

func buildQuery(table string, data QueryData, whereClauses map[string]string, visited map[string]bool, shouldSubsetMap map[string]bool) (string, []string) {
	fmt.Println("------- buildquery")
	joins := []string{}
	wheres := []string{}
	fmt.Printf("table: %s \n", table)

	if condition, exists := whereClauses[table]; exists {
		fmt.Println("where exists. returning")
		wheres = append(wheres, condition)
		return strings.Join(joins, " "), wheres
	}

	jsonF, _ := json.MarshalIndent(data[table], "", " ")
	fmt.Printf("data[table]: %s \n \n", string(jsonF))

	if columns, exists := data[table]; exists {
		for _, col := range columns {
			if col.ForeignKey.Table != "" && col.ForeignKey.Column != "" {
				shouldSubset := shouldSubsetMap[col.ForeignKey.Table]
				if shouldSubset {
					join := fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
						col.ForeignKey.Table, table, col.Column, col.ForeignKey.Table, col.ForeignKey.Column)
					joins = append(joins, join)
				}

				if !visited[col.ForeignKey.Table] {
					visited[col.ForeignKey.Table] = true
					subQuery, subWheres := buildQuery(col.ForeignKey.Table, data, whereClauses, visited, shouldSubsetMap)
					fmt.Printf("add subquery: %s \n", col.ForeignKey.Table)

					if shouldSubset {
						if subQuery != "" {
							joins = append(joins, subQuery)
						}
					}
					wheres = append(wheres, subWheres...)
					jsonF, _ := json.MarshalIndent(joins, "", " ")
					fmt.Printf("\n joins: %s \n", string(jsonF))
					jsonF, _ = json.MarshalIndent(wheres, "", " ")
					fmt.Printf("\n wheres: %s \n", string(jsonF))
				}
			}
		}
	}

	return strings.Join(joins, " "), wheres
}

func SubsetQueries(data QueryData, whereClauses map[string]string, shouldSubsetMap map[string]bool) map[string]string {
	queries := make(map[string]string)

	for table := range data {
		fmt.Println()
		fmt.Printf("---- query table: %s \n", table)
		visited := make(map[string]bool)
		joins, wheres := buildQuery(table, data, whereClauses, visited, shouldSubsetMap)
		query := fmt.Sprintf("SELECT %s.* FROM %s %s", table, table, joins)
		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}
		queries[table] = query
	}

	return queries
}

func checkParentWhereClause(table string, data QueryData, whereClauses map[string]string, visited map[string]bool) bool {
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

func ParentsWithWhereClauses(data QueryData, whereClauses map[string]string) map[string]bool {
	result := make(map[string]bool)
	for table := range data {
		visited := make(map[string]bool)
		result[table] = checkParentWhereClause(table, data, whereClauses, visited)
	}
	return result
}

func main() {
	queryData := QueryData{
		"public.a": []ColumnDefinition{
			{
				Column: "x_id",
				ForeignKey: ForeignKey{
					Table:  "public.x",
					Column: "id",
				},
			},
		},
		"public.c": []ColumnDefinition{
			{
				Column: "a_id",
				ForeignKey: ForeignKey{
					Table:  "public.a",
					Column: "id",
				},
			},
			{
				Column: "b_id",
				ForeignKey: ForeignKey{
					Table:  "public.b",
					Column: "id",
				},
			},
		},
		"public.d": []ColumnDefinition{
			{
				Column: "c_id",
				ForeignKey: ForeignKey{
					Table:  "public.c",
					Column: "id",
				},
			},
		},
		"public.e": []ColumnDefinition{
			{
				Column: "c_id",
				ForeignKey: ForeignKey{
					Table:  "public.c",
					Column: "id",
				},
			},
		},
		"public.x": []ColumnDefinition{},
		"public.b": []ColumnDefinition{},
		"public.z": []ColumnDefinition{},
	}

	whereClauses := map[string]string{
		"public.b": "public.b.id = '1'",
		"public.x": "public.x.id = '2'",
	}

	parentsMap := ParentsWithWhereClauses(queryData, whereClauses)

	queries := SubsetQueries(queryData, whereClauses, parentsMap)
	// for table, query := range queries {
	// 	fmt.Println(table + ":")
	// 	fmt.Println(query + "\n")
	// }

	// fmt.Println("------------")
	queryData = QueryData{
		"public.a": []ColumnDefinition{},
		"public.b": []ColumnDefinition{
			{
				Column: "a_id",
				ForeignKey: ForeignKey{
					Table:  "public.a",
					Column: "id",
				},
			},
		},
		"public.c": []ColumnDefinition{
			{
				Column: "b_id",
				ForeignKey: ForeignKey{
					Table:  "public.b",
					Column: "id",
				},
			},
			{
				Column: "d_id",
				ForeignKey: ForeignKey{
					Table:  "public.d",
					Column: "id",
				},
			},
		},
		"public.d": []ColumnDefinition{},
	}

	whereClauses = map[string]string{
		"public.a": "public.a.id = '1'",
	}

	parentsMap = ParentsWithWhereClauses(queryData, whereClauses)

	// jsonF, _ := json.MarshalIndent(parentsMap, "", " ")
	// fmt.Printf("\n parentsMap: %s \n", string(jsonF))

	// jsonF, _ = json.MarshalIndent(queryData, "", " ")
	// fmt.Printf("\n queryData: %s \n", string(jsonF))

	queries = SubsetQueries(queryData, whereClauses, parentsMap)
	// for table, query := range queries {
	// 	fmt.Println(table + ":")
	// 	fmt.Println(query + "\n")
	// }

	fmt.Println("------------")
	queryData = QueryData{
		"public.a": []ColumnDefinition{},
		"public.b": []ColumnDefinition{
			{
				Column: "a_id",
				ForeignKey: ForeignKey{
					Table:  "public.a",
					Column: "id",
				},
			},
		},
		"public.c": []ColumnDefinition{
			{
				Column: "b_id",
				ForeignKey: ForeignKey{
					Table:  "public.b",
					Column: "id",
				},
			},
			{
				Column: "bb_id",
				ForeignKey: ForeignKey{
					Table:  "public.b",
					Column: "id",
				},
			},
		},
	}

	whereClauses = map[string]string{
		"public.a": "public.a.id = '1'",
	}

	parentsMap = ParentsWithWhereClauses(queryData, whereClauses)
	for table, hasParentWithWhere := range parentsMap {
		fmt.Printf("Table %s has parent with WHERE: %t\n", table, hasParentWithWhere)
	}

	jsonF, _ := json.MarshalIndent(parentsMap, "", " ")
	fmt.Printf("\n parentsMap: %s \n", string(jsonF))

	jsonF, _ = json.MarshalIndent(queryData, "", " ")
	fmt.Printf("\n queryData: %s \n", string(jsonF))

	queries = SubsetQueries(queryData, whereClauses, parentsMap)
	for table, query := range queries {
		fmt.Println(table + ":")
		fmt.Println(query + "\n")
	}
}
