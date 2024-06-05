package genbenthosconfigs_querybuilder

import (
	"fmt"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleReference() {
	whereId := "id = 1"
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"genbenthosconfigs_querybuilder.department": {
			{Columns: []string{"company_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.company", Columns: []string{"id"}}},
		},
		"genbenthosconfigs_querybuilder.transaction": {
			{Columns: []string{"department_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}},
		},
		"genbenthosconfigs_querybuilder.expense_report": {
			{Columns: []string{"department_source_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}},
			{Columns: []string{"department_destination_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}},
			{Columns: []string{"transaction_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.transaction", Columns: []string{"id"}}},
		},
	}
	dependencyConfigs := []*tabledependency.RunConfig{
		{Table: "genbenthosconfigs_querybuilder.company", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
		{Table: "genbenthosconfigs_querybuilder.department", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "company_id"}, InsertColumns: []string{"id", "company_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.company", Columns: []string{"id"}}}},
		{Table: "genbenthosconfigs_querybuilder.transaction", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "department_id"}, InsertColumns: []string{"id", "department_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}}},
		{Table: "genbenthosconfigs_querybuilder.expense_report", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, InsertColumns: []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}, {Table: "public.transaction", Columns: []string{"id"}}}},
	}

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.company": {
			"id": {1},
		},
		"genbenthosconfigs_querybuilder.department": {
			"company_id": {1},
		},
		"genbenthosconfigs_querybuilder.expense_report": {
			"department_source_id":      {1},
			"department_destination_id": {2},
		},
		"genbenthosconfigs_querybuilder.transaction": {
			"department_id": {1, 2},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.company":        1,
		"genbenthosconfigs_querybuilder.department":     2,
		"genbenthosconfigs_querybuilder.expense_report": 1,
		"genbenthosconfigs_querybuilder.transaction":    2,
	}

	sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, map[string]map[string]*sqlmanager_shared.ColumnInfo{})
	require.NoError(s.T(), err)
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[tabledependency.RunTypeInsert]
		require.NotEmpty(s.T(), sql)
		rows, err := s.pgpool.Query(s.ctx, sql)
		require.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[table]
		require.True(s.T(), ok)

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			require.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					require.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		require.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}
