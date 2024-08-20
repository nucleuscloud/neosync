package querybuilder2

import (
	"fmt"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleReference() {
	s.runParallel("Test_BuildQueryMap_DoubleReference", func() {
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
			"genbenthosconfigs_querybuilder.expense": {
				{Columns: []string{"report_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.expense_report", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.item": {
				{Columns: []string{"expense_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.expense", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.company", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.department", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "company_id"}, InsertColumns: []string{"id", "company_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.company", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.transaction", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "department_id"}, InsertColumns: []string{"id", "department_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.expense_report", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, InsertColumns: []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}, {Table: "genbenthosconfigs_querybuilder.transaction", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.company": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.company": {
				"id": {1},
			},
			"genbenthosconfigs_querybuilder.department": {
				"company_id": {1},
			},
			"genbenthosconfigs_querybuilder.expense_report": {
				"department_source_id":      {1, 2},
				"department_destination_id": {1, 2},
			},
			"genbenthosconfigs_querybuilder.transaction": {
				"department_id": {1, 2},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.company":        1,
			"genbenthosconfigs_querybuilder.department":     2,
			"genbenthosconfigs_querybuilder.expense_report": 2,
			"genbenthosconfigs_querybuilder.transaction":    2,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)
			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err, "error running query: %s", sql)

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleReference_Complex() {
	s.runParallel("Test_BuildQueryMap_DoubleReference_Complex", func() {
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
			"genbenthosconfigs_querybuilder.expense": {
				{Columns: []string{"report_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.expense_report", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.item": {
				{Columns: []string{"expense_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.expense", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.company", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.department", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "company_id"}, InsertColumns: []string{"id", "company_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.company", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.transaction", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "department_id"}, InsertColumns: []string{"id", "department_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.expense_report", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, InsertColumns: []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}, {Table: "genbenthosconfigs_querybuilder.transaction", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.expense", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "report_id"}, InsertColumns: []string{"id", "report_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.expense_report", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.item", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id", "expense_id"}, InsertColumns: []string{"id", "expense_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.expense", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.company": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.company": {
				"id": {1},
			},
			"genbenthosconfigs_querybuilder.department": {
				"company_id": {1},
			},
			"genbenthosconfigs_querybuilder.expense_report": {
				"department_source_id":      {1, 2},
				"department_destination_id": {1, 2},
			},
			"genbenthosconfigs_querybuilder.transaction": {
				"department_id": {1, 2},
			},
			"genbenthosconfigs_querybuilder.expense": {
				"report_id": {3, 1},
			},
			"genbenthosconfigs_querybuilder.item": {
				"expense_id": {3, 2},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.company":        1,
			"genbenthosconfigs_querybuilder.department":     2,
			"genbenthosconfigs_querybuilder.expense_report": 2,
			"genbenthosconfigs_querybuilder.transaction":    2,
			"genbenthosconfigs_querybuilder.expense":        2,
			"genbenthosconfigs_querybuilder.item":           2,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleRootSubset() {
	s.runParallel("Test_BuildQueryMap_DoubleRootSubset", func() {
		whereCreated := "created > '2023-06-03'"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.test_2_c": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_2_a", Columns: []string{"id"}}},
				{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_2_b", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_2_d": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_2_c", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_2_e": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_2_c", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_2_a": {
				{Columns: []string{"x_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_2_x", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.test_2_x", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereCreated},
			{Table: "genbenthosconfigs_querybuilder.test_2_b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereCreated},
			{Table: "genbenthosconfigs_querybuilder.test_2_a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "x_id"}, InsertColumns: []string{"id", "x_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_x", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_2_c", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id", "b_id"}, InsertColumns: []string{"id", "a_id", "b_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_a", Columns: []string{"id"}}, {Table: "genbenthosconfigs_querybuilder.test_2_b", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_2_d", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_c", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_2_e", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_c", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.test_2_x": {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_b": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.test_2_x": {
				"id": {3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_2_b": {
				"id": {3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_2_a": {
				"x_id": {3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_2_c": {
				"a_id": {3, 4},
				"x_id": {3, 4},
			},
			"genbenthosconfigs_querybuilder.test_2_d": {
				"c_id": {3, 4},
			},
			"genbenthosconfigs_querybuilder.test_2_e": {
				"c_id": {3, 4},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.test_2_x": 3,
			"genbenthosconfigs_querybuilder.test_2_b": 3,
			"genbenthosconfigs_querybuilder.test_2_a": 4,
			"genbenthosconfigs_querybuilder.test_2_c": 2,
			"genbenthosconfigs_querybuilder.test_2_d": 2,
			"genbenthosconfigs_querybuilder.test_2_e": 2,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()

			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleRoots() {
	s.runParallel("Test_BuildQueryMap_MultipleRoots", func() {
		whereId := "id = 1"
		whereId4 := "id in (4,5)"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.test_3_b": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_c": {
				{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_d": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_e": {
				{Columns: []string{"d_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_g": {
				{Columns: []string{"f_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_f", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_h": {
				{Columns: []string{"g_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_g", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_i": {
				{Columns: []string{"h_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_h", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.test_3_a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
			{Table: "genbenthosconfigs_querybuilder.test_3_b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"id", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.test_3_c", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "b_id"}, InsertColumns: []string{"id", "b_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_d", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_e", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "d_id"}, InsertColumns: []string{"id", "d_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_f", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId4},
			{Table: "genbenthosconfigs_querybuilder.test_3_g", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "f_id"}, InsertColumns: []string{"id", "f_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_f", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_h", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "g_id"}, InsertColumns: []string{"id", "g_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_g", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_i", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "h_id"}, InsertColumns: []string{"id", "h_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_h", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.test_3_a": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.test_3_a": {
				"id": {1, 2, 3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_b": {
				"a_id": {3},
			},
			"genbenthosconfigs_querybuilder.test_3_c": {
				"b_id": {1},
			},
			"genbenthosconfigs_querybuilder.test_3_d": {
				"c_id": {3},
			},
			"genbenthosconfigs_querybuilder.test_3_e": {
				"d_id": {5},
			},
			"genbenthosconfigs_querybuilder.test_3_f": {
				"id": {4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_g": {
				"f_id": {4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_h": {
				"g_id": {1, 3},
			},
			"genbenthosconfigs_querybuilder.test_3_i": {
				"h_id": {4},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.test_3_a": 5,
			"genbenthosconfigs_querybuilder.test_3_b": 1,
			"genbenthosconfigs_querybuilder.test_3_c": 1,
			"genbenthosconfigs_querybuilder.test_3_d": 1,
			"genbenthosconfigs_querybuilder.test_3_e": 1,
			"genbenthosconfigs_querybuilder.test_3_f": 2,
			"genbenthosconfigs_querybuilder.test_3_g": 2,
			"genbenthosconfigs_querybuilder.test_3_h": 2,
			"genbenthosconfigs_querybuilder.test_3_i": 1,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleSubsets() {
	s.runParallel("Test_BuildQueryMap_MultipleSubsets", func() {
		whereId := "id in (3,4,5)"
		whereId4 := "id = 4"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.test_3_b": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_c": {
				{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_d": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_e": {
				{Columns: []string{"d_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.test_3_a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.test_3_b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"id", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}}, WhereClause: &whereId4},
			{Table: "genbenthosconfigs_querybuilder.test_3_c", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "b_id"}, InsertColumns: []string{"id", "b_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_d", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_e", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "d_id"}, InsertColumns: []string{"id", "d_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.test_3_a": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.test_3_a": {
				"id": {3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_b": {
				"a_id": {4},
			},
			"genbenthosconfigs_querybuilder.test_3_c": {
				"b_id": {4},
			},
			"genbenthosconfigs_querybuilder.test_3_d": {
				"c_id": {2},
			},
			"genbenthosconfigs_querybuilder.test_3_e": {
				"d_id": {4},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.test_3_a": 3,
			"genbenthosconfigs_querybuilder.test_3_b": 1,
			"genbenthosconfigs_querybuilder.test_3_c": 1,
			"genbenthosconfigs_querybuilder.test_3_d": 1,
			"genbenthosconfigs_querybuilder.test_3_e": 1,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleSubsets_SubsetsByForeignKeysOff() {
	s.runParallel("Test_BuildQueryMap_MultipleSubsets_SubsetsByForeignKeysOff", func() {
		whereId := "id in (4,5)"
		whereId4 := "id = 4"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.test_3_b": {
				{Columns: []string{"a_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_c": {
				{Columns: []string{"b_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_d": {
				{Columns: []string{"c_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.test_3_e": {
				{Columns: []string{"d_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.test_3_a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.test_3_b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "a_id"}, InsertColumns: []string{"id", "a_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}}, WhereClause: &whereId4},
			{Table: "genbenthosconfigs_querybuilder.test_3_c", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "b_id"}, InsertColumns: []string{"id", "b_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_d", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "c_id"}, InsertColumns: []string{"id", "c_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.test_3_e", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "d_id"}, InsertColumns: []string{"id", "d_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.test_3_a": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.test_3_a": {
				"id": {4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_b": {
				"a_id": {4},
			},
			"genbenthosconfigs_querybuilder.test_3_c": {
				"b_id": {1, 2, 3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_d": {
				"c_id": {1, 2, 3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_e": {
				"d_id": {1, 2, 3, 4, 5},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.test_3_a": 2,
			"genbenthosconfigs_querybuilder.test_3_b": 1,
			"genbenthosconfigs_querybuilder.test_3_c": 5,
			"genbenthosconfigs_querybuilder.test_3_d": 5,
			"genbenthosconfigs_querybuilder.test_3_e": 5,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, false, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_CircularDependency() {
	s.runParallel("Test_BuildQueryMap_CircularDependency", func() {
		whereId := "id in (1,5)"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.addresses": {
				{Columns: []string{"order_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.orders", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.customers": {
				{Columns: []string{"address_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.addresses", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.orders": {
				{Columns: []string{"customer_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.customers", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.payments": {
				{Columns: []string{"customer_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.customers", Columns: []string{"id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.orders", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "customer_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
			{Table: "genbenthosconfigs_querybuilder.addresses", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "order_id"}, InsertColumns: []string{"id", "order_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.orders", Columns: []string{"id"}}}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.customers", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "address_id"}, InsertColumns: []string{"id", "address_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.addresses", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.payments", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "customer_id"}, InsertColumns: []string{"id", "customer_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.customers", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.orders", RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "customer_id"}, InsertColumns: []string{"customer_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.orders", Columns: []string{"id"}}, {Table: "genbenthosconfigs_querybuilder.customers", Columns: []string{"id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.orders": {
				"customer_id": {2, 3},
			},
			"genbenthosconfigs_querybuilder.addresses": {
				"order_id": {1, 5},
			},
			"genbenthosconfigs_querybuilder.customers": {
				"address_id": {1, 5},
			},
			"genbenthosconfigs_querybuilder.payments": {
				"customer_id": {2},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.orders":    2,
			"genbenthosconfigs_querybuilder.addresses": 2,
			"genbenthosconfigs_querybuilder.customers": 2,
			"genbenthosconfigs_querybuilder.payments":  1,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_NoForeignKeys() {
	s.runParallel("Test_BuildQueryMap_NoForeignKeys", func() {
		whereId := "id in (1,5)"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.company", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
			{Table: "genbenthosconfigs_querybuilder.test_2_x", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.test_2_b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.test_3_a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.company":  {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_x": {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_b": {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_a": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.company": {
				"id": {1, 2, 3},
			},
			"genbenthosconfigs_querybuilder.test_2_x": {
				"id": {1, 5},
			},
			"genbenthosconfigs_querybuilder.test_2_b": {
				"id": {1, 5},
			},
			"genbenthosconfigs_querybuilder.test_3_a": {
				"customer_id": {1, 2, 3, 4, 5},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.company":  3,
			"genbenthosconfigs_querybuilder.test_2_x": 2,
			"genbenthosconfigs_querybuilder.test_2_b": 2,
			"genbenthosconfigs_querybuilder.test_3_a": 5,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))

		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_NoForeignKeys_NoSubsets() {
	s.runParallel("Test_BuildQueryMap_NoForeignKeys_NoSubsets", func() {
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.company", RunType: tabledependency.RunTypeInsert, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
			{Table: "genbenthosconfigs_querybuilder.test_2_x", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
			{Table: "genbenthosconfigs_querybuilder.test_2_b", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
			{Table: "genbenthosconfigs_querybuilder.test_3_a", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.company":  {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_x": {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_b": {"id": &sqlmanager_shared.ColumnInfo{}},
			"genbenthosconfigs_querybuilder.test_2_a": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.company":  3,
			"genbenthosconfigs_querybuilder.test_2_x": 5,
			"genbenthosconfigs_querybuilder.test_2_b": 5,
			"genbenthosconfigs_querybuilder.test_3_a": 5,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedCount), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			rowCount := 0
			for rows.Next() {
				rowCount++
			}
			rows.Close()

			tableExpectedCount, ok := expectedCount[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected row counts", table))
			require.Equalf(s.T(), rowCount, tableExpectedCount, fmt.Sprintf("table: %s ", table))
		}
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_SubsetCompositeKeys() {
	s.runParallel("Test_BuildQueryMap_SubsetCompositeKeys", func() {
		whereId := "id in (3,5)"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.employees": {
				{Columns: []string{"division_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.division", Columns: []string{"id"}}},
			},
			"genbenthosconfigs_querybuilder.projects": {
				{Columns: []string{"responsible_employee_id", "responsible_division_id"}, NotNullable: []bool{true, true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.employees", Columns: []string{"id", "division_id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.division", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.employees", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id", "division_id"}, SelectColumns: []string{"id", "division_id"}, InsertColumns: []string{"id", "division_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.division", Columns: []string{"id"}}}},
			{Table: "genbenthosconfigs_querybuilder.projects", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "responsible_employee_id", "responsible_division_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.employees", Columns: []string{"id", "division_id"}}}},
		}

		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.division": {"id": &sqlmanager_shared.ColumnInfo{}},
		}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.division": {
				"id": {3, 5},
			},
			"genbenthosconfigs_querybuilder.employees": {
				"division_id": {3, 5},
				"id":          {8, 10},
			},
			"genbenthosconfigs_querybuilder.projects": {
				"responsible_division_id": {3, 5},
				"responsible_employee_id": {8, 10},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.division":  2,
			"genbenthosconfigs_querybuilder.employees": 2,
			"genbenthosconfigs_querybuilder.projects":  2,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

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
			require.Equalf(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
		}
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_SubsetSelfReferencing() {
	s.runParallel("Test_BuildQueryMap_SubsetSelfReferencing", func() {
		whereId := "id in (3,5)"
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.bosses": {
				{Columns: []string{"manager_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.bosses", Columns: []string{"id"}}},
				{Columns: []string{"big_boss_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.bosses", Columns: []string{"manager_id"}}}, // todo: should the FK here be id instead of manager_id?
			},
			"genbenthosconfigs_querybuilder.minions": {
				{Columns: []string{"boss_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.bosses", Columns: []string{"big_boss_id"}}},
			},
		}
		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.bosses", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "manager_id", "big_boss_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{}, WhereClause: &whereId},
			{Table: "genbenthosconfigs_querybuilder.minions", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, SelectColumns: []string{"id", "boss_id"}, InsertColumns: []string{"id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.bosses", Columns: []string{"manager_id"}}}},
		}
		columnInfo := map[string]map[string]*sqlmanager_shared.ColumnInfo{}

		expectedValues := map[string]map[string][]int64{
			"genbenthosconfigs_querybuilder.bosses": {
				"id":         {1, 2, 3, 4, 5},
				"manager_id": {1, 2, 3, 4},
				"boss_id":    {1, 2, 3},
			},
			"genbenthosconfigs_querybuilder.minions": {
				"boss_id": {1, 3},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.bosses":  5,
			"genbenthosconfigs_querybuilder.minions": 2,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfo)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))

		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()
			tableExpectedValues, ok := expectedValues[table]
			require.True(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

			rowCount := 0
			for rows.Next() {
				rowCount++
				values, err := rows.Values()
				require.NoError(s.T(), err)

				for i, col := range values {
					colName := columnDescriptions[i].Name
					if (colName == "manager_id" || colName == "big_boss_id") && col == nil {
						continue
					}
					allowedValues, ok := tableExpectedValues[colName]
					if ok {
						value := col.(int64)
						require.Contains(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
					}
				}
			}
			rows.Close()
			require.Equal(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
		}
	})
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_ComplexSubset() {
	s.runParallel("Test_BuildQueryMap_ComplexSubset", func() {
		tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
			"genbenthosconfigs_querybuilder.attachments": {
				{Columns: []string{"uploaded_by"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
				{Columns: []string{"task_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.tasks", Columns: []string{"task_id"}}},
				{Columns: []string{"initiative_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}}},
				{Columns: []string{"comment_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.comments", Columns: []string{"comment_id"}}},
			},
			"genbenthosconfigs_querybuilder.comments": {
				{Columns: []string{"user_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
				{Columns: []string{"task_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.tasks", Columns: []string{"task_id"}}},
				{Columns: []string{"initiative_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}}},
				{Columns: []string{"parent_comment_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.comments", Columns: []string{"comment_id"}}},
			},
			"genbenthosconfigs_querybuilder.initiatives": {
				{Columns: []string{"lead_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
				{Columns: []string{"client_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
			},
			"genbenthosconfigs_querybuilder.tasks": {
				{Columns: []string{"initiative_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}}},
				{Columns: []string{"assignee_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
				{Columns: []string{"reviewer_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
			},
			"genbenthosconfigs_querybuilder.user_skills": {
				{Columns: []string{"user_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
				{Columns: []string{"skill_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.skills", Columns: []string{"skill_id"}}},
			},
			"genbenthosconfigs_querybuilder.users": {
				{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
				{Columns: []string{"mentor_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}},
			},
		}

		dependencyConfigs := []*tabledependency.RunConfig{
			{Table: "genbenthosconfigs_querybuilder.comments", SelectColumns: []string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id", "parent_comment_id"}, InsertColumns: []string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}, {Table: "genbenthosconfigs_querybuilder.tasks", Columns: []string{"task_id"}}, {Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}}}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"comment_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.comments", SelectColumns: []string{"comment_id", "parent_comment_id"}, InsertColumns: []string{"parent_comment_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.comments", Columns: []string{"comment_id"}}}, RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"comment_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.users", SelectColumns: []string{"user_id", "name", "email", "manager_id", "mentor_id"}, InsertColumns: []string{"user_id", "name", "email"}, DependsOn: []*tabledependency.DependsOn{}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"user_id"}, WhereClause: ptrString("user_id in (1,2,3,4,5)"), SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.users", SelectColumns: []string{"user_id", "manager_id", "mentor_id"}, InsertColumns: []string{"manager_id", "mentor_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}}, RunType: tabledependency.RunTypeUpdate, PrimaryKeys: []string{"user_id"}, WhereClause: ptrString("user_id = 1"), SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.initiatives", SelectColumns: []string{"initiative_id", "name", "description", "lead_id", "client_id"}, InsertColumns: []string{"initiative_id", "name", "description", "lead_id", "client_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id", "user_id"}}}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"initiative_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.skills", SelectColumns: []string{"skill_id", "name", "category"}, InsertColumns: []string{"skill_id", "name", "category"}, DependsOn: []*tabledependency.DependsOn{}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"skill_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.tasks", SelectColumns: []string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"}, InsertColumns: []string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}}, {Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id", "user_id"}}}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"task_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.user_skills", SelectColumns: []string{"user_skill_id", "user_id", "skill_id", "proficiency_level"}, InsertColumns: []string{"user_skill_id", "user_id", "skill_id", "proficiency_level"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}, {Table: "genbenthosconfigs_querybuilder.skills", Columns: []string{"skill_id"}}}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"user_skill_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
			{Table: "genbenthosconfigs_querybuilder.attachments", SelectColumns: []string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"}, InsertColumns: []string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"}, DependsOn: []*tabledependency.DependsOn{{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}}, {Table: "genbenthosconfigs_querybuilder.tasks", Columns: []string{"task_id"}}, {Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}}, {Table: "genbenthosconfigs_querybuilder.comments", Columns: []string{"comment_id"}}}, RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"attachment_id"}, WhereClause: nil, SelectQuery: nil, SplitColumnPaths: false},
		}

		columnInfoMap := map[string]map[string]*sqlmanager_shared.ColumnInfo{
			"genbenthosconfigs_querybuilder.attachments": {
				"attachment_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('attachments_attachment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"comment_id":    &sqlmanager_shared.ColumnInfo{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"file_name":     &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: ptrInt32(255), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"file_path":     &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: ptrInt32(255), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"initiative_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"task_id":       &sqlmanager_shared.ColumnInfo{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"uploaded_by":   &sqlmanager_shared.ColumnInfo{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
			},
			"genbenthosconfigs_querybuilder.comments": {
				"comment_id":        &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('comments_comment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"content":           &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "text", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"created_at":        &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "CURRENT_TIMESTAMP", IsNullable: true, DataType: "timestamp without time zone", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"initiative_id":     &sqlmanager_shared.ColumnInfo{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"parent_comment_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"task_id":           &sqlmanager_shared.ColumnInfo{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"user_id":           &sqlmanager_shared.ColumnInfo{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
			},
			"genbenthosconfigs_querybuilder.initiatives": {
				"client_id":     &sqlmanager_shared.ColumnInfo{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"description":   &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"initiative_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('initiatives_initiative_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"lead_id":       &sqlmanager_shared.ColumnInfo{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"name":          &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: ptrInt32(100), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
			},
			"genbenthosconfigs_querybuilder.skills": {
				"category": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "character varying(100)", CharacterMaximumLength: ptrInt32(100), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"name":     &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: ptrInt32(100), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"skill_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('skills_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
			},
			"genbenthosconfigs_querybuilder.tasks": {
				"assignee_id":   &sqlmanager_shared.ColumnInfo{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"description":   &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"initiative_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"reviewer_id":   &sqlmanager_shared.ColumnInfo{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"status":        &sqlmanager_shared.ColumnInfo{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "character varying(50)", CharacterMaximumLength: ptrInt32(50), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"task_id":       &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('tasks_task_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"title":         &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(200)", CharacterMaximumLength: ptrInt32(200), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
			},
			"genbenthosconfigs_querybuilder.user_skills": {
				"proficiency_level": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"skill_id":          &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"user_id":           &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"user_skill_id":     &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('user_skills_user_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
			},
			"genbenthosconfigs_querybuilder.users": {
				"email":      &sqlmanager_shared.ColumnInfo{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: ptrInt32(100), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"manager_id": &sqlmanager_shared.ColumnInfo{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"mentor_id":  &sqlmanager_shared.ColumnInfo{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
				"name":       &sqlmanager_shared.ColumnInfo{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: ptrInt32(100), NumericPrecision: ptrInt32(-1), NumericScale: ptrInt32(-1)},
				"user_id":    &sqlmanager_shared.ColumnInfo{OrdinalPosition: 1, ColumnDefault: "nextval('users_user_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: ptrInt32(-1), NumericPrecision: ptrInt32(32), NumericScale: ptrInt32(0)},
			},
		}

		expectedValues := map[string]map[string][]int32{
			"genbenthosconfigs_querybuilder.users": {
				"user_id": {1, 2, 3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.user_skills": {
				"user_skill_id": {1, 2, 3, 4, 5},
				"skill_id":      {1, 2, 3, 4, 5},
				"user_id":       {1, 2, 3, 4, 5},
			},
			"genbenthosconfigs_querybuilder.tasks": {
				"task_id": {1, 2, 3},
			},
			"genbenthosconfigs_querybuilder.skills": {
				"skill_id": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			"genbenthosconfigs_querybuilder.initiatives": {
				"initiative_id": {1, 2, 3, 4},
			},
			"genbenthosconfigs_querybuilder.comments": {
				"comment_id": {1, 2, 3, 4, 5, 6},
			},
			"genbenthosconfigs_querybuilder.attachments": {
				"attachment_id": {1, 2, 3},
			},
		}

		expectedCount := map[string]int{
			"genbenthosconfigs_querybuilder.users":       5,
			"genbenthosconfigs_querybuilder.user_skills": 5,
			"genbenthosconfigs_querybuilder.tasks":       3,
			"genbenthosconfigs_querybuilder.skills":      10,
			"genbenthosconfigs_querybuilder.initiatives": 4,
			"genbenthosconfigs_querybuilder.comments":    6,
			"genbenthosconfigs_querybuilder.attachments": 3,
		}

		sqlMap, err := BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, tableDependencies, dependencyConfigs, true, columnInfoMap)
		require.NoError(s.T(), err)
		require.Equal(s.T(), len(expectedValues), len(sqlMap))
		for table, selectQueryRunType := range sqlMap {
			sql := selectQueryRunType[tabledependency.RunTypeInsert]
			require.NotEmpty(s.T(), sql)

			rows, err := s.pgpool.Query(s.ctx, sql)
			require.NoError(s.T(), err)

			columnDescriptions := rows.FieldDescriptions()

			tableExpectedValues, ok := expectedValues[table]
			require.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

			rowCount := 0
			for rows.Next() {
				rowCount++
				values, err := rows.Values()
				require.NoError(s.T(), err)

				for i, col := range values {
					colName := columnDescriptions[i].Name
					allowedValues, ok := tableExpectedValues[colName]
					if ok {
						value := col.(int32)
						require.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
					}
				}
			}
			rows.Close()
			require.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
		}
	})
}
func ptrString(s string) *string {
	return &s
}

func ptrInt32(i int32) *int32 {
	return &i
}

func (s *IntegrationTestSuite) runParallel(name string, testFunc func()) {
	s.T().Run(name, func(t *testing.T) {
		s.wg.Add(1)
		defer s.wg.Done()
		t.Parallel()
		testFunc()
	})
}
