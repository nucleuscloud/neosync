package main

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type Join struct {
	JoinType       string
	JoinTable      string
	BaseTable      string
	Alias          *string
	JoinColumnsMap map[string]string
}

type QueryInput struct {
	Joins                             []Join
	WhereClauses                      []string
	SelfReferencingCircularDependency []string
}

func buildSQLQuery(input QueryInput) (string, error) {
	dialect := goqu.Dialect("postgres")

	// Start with the base table
	baseTable := "genbenthosconfigs_querybuilder.item"
	ds := dialect.From(baseTable)

	tableAliases := make(map[string][]string)
	aliasCounter := 1

	// Assign alias to the base table
	baseAlias := fmt.Sprintf("t%d", aliasCounter)
	tableAliases[baseTable] = []string{baseAlias}
	ds = ds.As(baseAlias)
	aliasCounter++

	// Process joins
	for _, join := range input.Joins {
		joinType := strings.ToUpper(join.JoinType)
		baseTable := join.BaseTable
		joinTable := join.JoinTable

		if _, exists := tableAliases[baseTable]; !exists {
			tableAliases[baseTable] = []string{fmt.Sprintf("t%d", aliasCounter)}
			aliasCounter++
		}

		// Create a new alias for the join table
		joinAlias := fmt.Sprintf("t%d", aliasCounter)
		tableAliases[joinTable] = append(tableAliases[joinTable], joinAlias)
		aliasCounter++

		baseAlias := tableAliases[baseTable][len(tableAliases[baseTable])-1]

		// Build join condition
		// var joinCondition goqu.Expression
		joinCondition := goqu.Ex{}
		for joinCol, baseCol := range join.JoinColumnsMap {
			joinCondition[fmt.Sprintf("%s.%s", joinAlias, joinCol)] = goqu.I(fmt.Sprintf("%s.%s", baseAlias, baseCol))
		}

		// Apply join
		switch joinType {
		case "INNER":
			ds = ds.InnerJoin(goqu.T(joinTable).As(joinAlias), goqu.On(joinCondition))
		case "LEFT":
			ds = ds.LeftJoin(goqu.T(joinTable).As(joinAlias), goqu.On(joinCondition))
		case "RIGHT":
			ds = ds.RightJoin(goqu.T(joinTable).As(joinAlias), goqu.On(joinCondition))
		default:
			return "", fmt.Errorf("unsupported join type: %s", joinType)
		}
	}

	// Process WHERE clauses
	whereClauses := []string{}
	for _, clause := range input.WhereClauses {
		parts := strings.Split(clause, ".")
		if len(parts) == 3 { // Assuming format: schema.table.column
			tableName := strings.Join(parts[:2], ".")
			if aliases, exists := tableAliases[tableName]; exists {
				for _, alias := range aliases {
					newClause := strings.Replace(clause, tableName+".", alias+".", 1)
					if !contains(whereClauses, newClause) {
						whereClauses = append(whereClauses, newClause)
						ds = ds.Where(goqu.L(newClause))
					}
				}
			}
		}
	}

	// Generate SQL
	sql, _, err := ds.ToSQL()
	if err != nil {
		return "", err
	}

	return sql, nil
}
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func main() {
	input := QueryInput{
		Joins: []Join{
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.expense",
				BaseTable: "genbenthosconfigs_querybuilder.item",
				JoinColumnsMap: map[string]string{
					"id": "expense_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.expense_report",
				BaseTable: "genbenthosconfigs_querybuilder.expense",
				JoinColumnsMap: map[string]string{
					"id": "report_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.department",
				BaseTable: "genbenthosconfigs_querybuilder.expense_report",
				JoinColumnsMap: map[string]string{
					"id": "department_source_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.company",
				BaseTable: "genbenthosconfigs_querybuilder.department",
				JoinColumnsMap: map[string]string{
					"id": "company_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.department",
				BaseTable: "genbenthosconfigs_querybuilder.expense_report",
				JoinColumnsMap: map[string]string{
					"id": "department_destination_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.company",
				BaseTable: "genbenthosconfigs_querybuilder.department",
				JoinColumnsMap: map[string]string{
					"id": "company_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.transaction",
				BaseTable: "genbenthosconfigs_querybuilder.expense_report",
				JoinColumnsMap: map[string]string{
					"id": "transaction_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.department",
				BaseTable: "genbenthosconfigs_querybuilder.transaction",
				JoinColumnsMap: map[string]string{
					"id": "department_id",
				},
			},
			{
				JoinType:  "INNER",
				JoinTable: "genbenthosconfigs_querybuilder.company",
				BaseTable: "genbenthosconfigs_querybuilder.department",
				JoinColumnsMap: map[string]string{
					"id": "company_id",
				},
			},
		},
		WhereClauses: []string{
			"genbenthosconfigs_querybuilder.company.id = 1",
			"genbenthosconfigs_querybuilder.company.id = 1",
			"genbenthosconfigs_querybuilder.company.id = 1",
		},
	}

	query, err := buildSQLQuery(input)
	if err != nil {
		fmt.Printf("Error building query: %v\n", err)
		return
	}
	fmt.Println(query)
}
