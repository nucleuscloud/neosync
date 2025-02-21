package querybuilder2

import (
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	runconfigs "github.com/nucleuscloud/neosync/internal/runconfigs"
	querybuilder2 "github.com/nucleuscloud/neosync/worker/pkg/query-builder2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pageLimit = 100
)

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleReference() {
	whereId := "id = 1"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.department": {
			{Columns: []string{"company_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.company", ReferenceColumns: []string{"id"}},
		},
		"genbenthosconfigs_querybuilder.transaction": {
			{Columns: []string{"department_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.department", ReferenceColumns: []string{"id"}},
		},
		"genbenthosconfigs_querybuilder.expense_report": {
			{Columns: []string{"department_source_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.department", ReferenceColumns: []string{"id"}},
			{Columns: []string{"department_destination_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.department", ReferenceColumns: []string{"id"}},
			{Columns: []string{"transaction_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.transaction", ReferenceColumns: []string{"id"}},
		},
		"genbenthosconfigs_querybuilder.expense": {
			{Columns: []string{"report_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.expense_report", ReferenceColumns: []string{"id"}},
		},
		"genbenthosconfigs_querybuilder.item": {
			{Columns: []string{"expense_id"}, NotNullable: []bool{true}, ReferenceTable: "genbenthosconfigs_querybuilder.expense", ReferenceColumns: []string{"id"}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig("genbenthosconfigs_querybuilder.company", runconfigs.RunTypeInsert, []string{"id"}, &whereId, []string{"id"}, []string{"id"}, []*runconfigs.DependsOn{}, nil),
		buildRunConfig("genbenthosconfigs_querybuilder.department", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "company_id"}, []string{"id", "company_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.company", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.department"]),
		buildRunConfig("genbenthosconfigs_querybuilder.transaction", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "department_id"}, []string{"id", "department_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.transaction"]),
		buildRunConfig("genbenthosconfigs_querybuilder.expense_report", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, []string{"id", "department_source_id", "department_destination_id", "transaction_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.department", Columns: []string{"id"}}, {Table: "genbenthosconfigs_querybuilder.transaction", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.expense_report"]),
		buildRunConfig("genbenthosconfigs_querybuilder.expense", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "report_id"}, []string{"id", "report_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.expense_report", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.expense"]),
		buildRunConfig("genbenthosconfigs_querybuilder.item", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "expense_id"}, []string{"id", "expense_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.expense", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.item"]),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.company": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)
		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[table]
		assert.True(s.T(), ok)

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleRootSubset() {
	whereCreated := "created > '2023-06-03'"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.test_2_c": {
			{Columns: []string{"a_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_2_a", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
			{Columns: []string{"b_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_2_b", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_2_d": {
			{Columns: []string{"c_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_2_c", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_2_e": {
			{Columns: []string{"c_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_2_c", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_2_a": {
			{Columns: []string{"x_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_2_x", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig("genbenthosconfigs_querybuilder.test_2_x", runconfigs.RunTypeInsert, []string{"id"}, &whereCreated, []string{"id"}, []string{"id"}, []*runconfigs.DependsOn{}, nil),
		buildRunConfig("genbenthosconfigs_querybuilder.test_2_b", runconfigs.RunTypeInsert, []string{"id"}, &whereCreated, []string{"id"}, []string{"id"}, []*runconfigs.DependsOn{}, nil),
		buildRunConfig("genbenthosconfigs_querybuilder.test_2_a", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "x_id"}, []string{"id", "x_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_x", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.test_2_a"]),
		buildRunConfig("genbenthosconfigs_querybuilder.test_2_c", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "a_id", "b_id"}, []string{"id", "a_id", "b_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_a", Columns: []string{"id"}}, {Table: "genbenthosconfigs_querybuilder.test_2_b", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.test_2_c"]),
		buildRunConfig("genbenthosconfigs_querybuilder.test_2_d", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "c_id"}, []string{"id", "c_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_c", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.test_2_d"]),
		buildRunConfig("genbenthosconfigs_querybuilder.test_2_e", runconfigs.RunTypeInsert, []string{"id"}, nil, []string{"id", "c_id"}, []string{"id", "c_id"}, []*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_2_c", Columns: []string{"id"}}}, tableDependencies["genbenthosconfigs_querybuilder.test_2_e"]),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.test_2_x": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_b": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleRoots() {
	whereId := "id = 1"
	whereId4 := "id in (4,5)"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.test_3_b": {
			{Columns: []string{"a_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_a", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_c": {
			{Columns: []string{"b_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_b", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_d": {
			{Columns: []string{"c_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_c", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_e": {
			{Columns: []string{"d_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_d", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_g": {
			{Columns: []string{"f_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_f", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_h": {
			{Columns: []string{"g_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_g", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_i": {
			{Columns: []string{"h_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_h", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_a",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_b",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id", "a_id"},
			[]string{"id", "a_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_b"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_c",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "b_id"},
			[]string{"id", "b_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_c"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_d",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "c_id"},
			[]string{"id", "c_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_d"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_e",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "d_id"},
			[]string{"id", "d_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_e"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_f",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId4,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_g",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "f_id"},
			[]string{"id", "f_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_f", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_g"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_h",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "g_id"},
			[]string{"id", "g_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_g", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_h"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_i",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "h_id"},
			[]string{"id", "h_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_h", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_i"],
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.test_3_a": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()
		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name

				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleSubsets() {
	whereId := "id in (3,4,5)"
	whereId4 := "id = 4"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.test_3_b": {
			{Columns: []string{"a_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_a", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_c": {
			{Columns: []string{"b_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_b", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_d": {
			{Columns: []string{"c_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_c", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_e": {
			{Columns: []string{"d_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_d", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_a",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_b",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId4,
			[]string{"id", "a_id"},
			[]string{"id", "a_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_b"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_c",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "b_id"},
			[]string{"id", "b_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_c"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_d",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "c_id"},
			[]string{"id", "c_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_d"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_e",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "d_id"},
			[]string{"id", "d_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_e"],
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.test_3_a": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()
		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleSubsets_SubsetsByForeignKeysOff() {
	whereId := "id in (4,5)"
	whereId4 := "id = 4"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.test_3_b": {
			{Columns: []string{"a_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_a", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_c": {
			{Columns: []string{"b_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_b", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_d": {
			{Columns: []string{"c_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_c", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.test_3_e": {
			{Columns: []string{"d_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.test_3_d", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_a",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_b",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId4,
			[]string{"id", "a_id"},
			[]string{"id", "a_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_a", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_b"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_c",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "b_id"},
			[]string{"id", "b_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_b", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_c"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_d",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "c_id"},
			[]string{"id", "c_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_c", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_d"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_e",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "d_id"},
			[]string{"id", "d_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.test_3_d", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.test_3_e"],
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.test_3_a": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, false, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()
		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_CircularDependency() {
	whereId := "id in (1,5)"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.addresses": {
			{Columns: []string{"order_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.orders", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.customers": {
			{Columns: []string{"address_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.addresses", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.orders": {
			{Columns: []string{"customer_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.customers", ReferenceColumns: []string{"id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.payments": {
			{Columns: []string{"customer_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.customers", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.orders",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "customer_id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			tableDependencies["genbenthosconfigs_querybuilder.orders"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.addresses",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id", "order_id"},
			[]string{"id", "order_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.orders", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.addresses"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.customers",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "address_id"},
			[]string{"id", "address_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.addresses", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.customers"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.payments",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "customer_id"},
			[]string{"id", "customer_id"},
			[]*runconfigs.DependsOn{{Table: "genbenthosconfigs_querybuilder.customers", Columns: []string{"id"}}},
			tableDependencies["genbenthosconfigs_querybuilder.payments"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.orders",
			runconfigs.RunTypeUpdate,
			[]string{"id"},
			nil,
			[]string{"id", "customer_id"},
			[]string{"customer_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.orders", Columns: []string{"id"}},
				{Table: "genbenthosconfigs_querybuilder.customers", Columns: []string{"id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.orders"],
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{}

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.orders": {
			"customer_id": {1, 2, 3, 4, 5},
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
		"genbenthosconfigs_querybuilder.orders":    5,
		"genbenthosconfigs_querybuilder.addresses": 2,
		"genbenthosconfigs_querybuilder.customers": 2,
		"genbenthosconfigs_querybuilder.payments":  1,
	}

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		if table != "genbenthosconfigs_querybuilder.customers" {
			assert.Truef(s.T(), sql.IsNotForeignKeySafeSubset, fmt.Sprintf("table: %s IsNotForeignKeySafeSubset should be true", table))
		}

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()
		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_NoForeignKeys() {
	whereId := "id in (1,5)"
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.company",
			runconfigs.RunTypeInsert,
			[]string{},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_2_x",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_2_b",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_a",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.company":  {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_x": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_b": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_a": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))

	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()
		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_NoForeignKeys_NoSubsets() {
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.company",
			runconfigs.RunTypeInsert,
			[]string{},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_2_x",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_2_b",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.test_3_a",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.company":  {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_x": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_b": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"genbenthosconfigs_querybuilder.test_2_a": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.company":  3,
		"genbenthosconfigs_querybuilder.test_2_x": 5,
		"genbenthosconfigs_querybuilder.test_2_b": 5,
		"genbenthosconfigs_querybuilder.test_3_a": 5,
	}

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedCount), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		rowCount := 0
		for rows.Next() {
			rowCount++
		}
		rows.Close()

		tableExpectedCount, ok := expectedCount[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected row counts", table))
		assert.Equalf(s.T(), rowCount, tableExpectedCount, fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_SubsetCompositeKeys() {
	whereId := "id in (3,5)"
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.employees": {
			{Columns: []string{"division_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.division", ReferenceColumns: []string{"id"}, NotNullable: []bool{true}},
		},
		"genbenthosconfigs_querybuilder.projects": {
			{Columns: []string{"responsible_employee_id", "responsible_division_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.employees", ReferenceColumns: []string{"id", "division_id"}, NotNullable: []bool{true, true}},
		},
	}
	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.division",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			&whereId,
			[]string{"id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.employees",
			runconfigs.RunTypeInsert,
			[]string{"id", "division_id"},
			nil,
			[]string{"id", "division_id"},
			[]string{"id", "division_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.division", Columns: []string{"id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.employees"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.projects",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "responsible_employee_id", "responsible_division_id"},
			[]string{"id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.employees", Columns: []string{"id", "division_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.projects"],
		),
	}

	columnInfo := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.division": {"id": &sqlmanager_shared.DatabaseSchemaRow{}},
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

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfo, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()
		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_ComplexSubset_Postgres() {
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.attachments": {
			{Columns: []string{"uploaded_by"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{true}},
			{Columns: []string{"task_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.tasks", ReferenceColumns: []string{"task_id"}, NotNullable: []bool{true}},
			{Columns: []string{"initiative_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.initiatives", ReferenceColumns: []string{"initiative_id"}, NotNullable: []bool{false}},
			{Columns: []string{"comment_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.comments", ReferenceColumns: []string{"comment_id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.comments": {
			{Columns: []string{"user_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{true}},
			{Columns: []string{"task_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.tasks", ReferenceColumns: []string{"task_id"}, NotNullable: []bool{false}},
			{Columns: []string{"initiative_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.initiatives", ReferenceColumns: []string{"initiative_id"}, NotNullable: []bool{false}},
			{Columns: []string{"parent_comment_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.comments", ReferenceColumns: []string{"comment_id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.initiatives": {
			{Columns: []string{"lead_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
			{Columns: []string{"client_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.tasks": {
			{Columns: []string{"initiative_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.initiatives", ReferenceColumns: []string{"initiative_id"}, NotNullable: []bool{false}},
			{Columns: []string{"assignee_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
			{Columns: []string{"reviewer_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.user_skills": {
			{Columns: []string{"user_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{true}},
			{Columns: []string{"skill_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.skills", ReferenceColumns: []string{"skill_id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.users": {
			{Columns: []string{"manager_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
			{Columns: []string{"mentor_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
		},
	}

	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.comments",
			runconfigs.RunTypeInsert,
			[]string{"comment_id"},
			nil,
			[]string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id", "parent_comment_id"},
			[]string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}},
				{Table: "genbenthosconfigs_querybuilder.tasks", Columns: []string{"task_id"}},
				{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.comments"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.comments",
			runconfigs.RunTypeUpdate,
			[]string{"comment_id"},
			nil,
			[]string{"comment_id", "parent_comment_id"},
			[]string{"parent_comment_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.comments", Columns: []string{"comment_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.comments"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.users",
			runconfigs.RunTypeInsert,
			[]string{"user_id"},
			ptr("user_id in (1,2,5,6,7,8)"),
			[]string{"user_id", "name", "email", "manager_id", "mentor_id"},
			[]string{"user_id", "name", "email"},
			[]*runconfigs.DependsOn{},
			tableDependencies["genbenthosconfigs_querybuilder.users"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.users",
			runconfigs.RunTypeUpdate,
			[]string{"user_id"},
			ptr("user_id = 1"),
			[]string{"user_id", "manager_id", "mentor_id"},
			[]string{"manager_id", "mentor_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.users"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.initiatives",
			runconfigs.RunTypeInsert,
			[]string{"initiative_id"},
			nil,
			[]string{"initiative_id", "name", "description", "lead_id", "client_id"},
			[]string{"initiative_id", "name", "description", "lead_id", "client_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id", "user_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.initiatives"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.skills",
			runconfigs.RunTypeInsert,
			[]string{"skill_id"},
			nil,
			[]string{"skill_id", "name", "category"},
			[]string{"skill_id", "name", "category"},
			[]*runconfigs.DependsOn{},
			tableDependencies["genbenthosconfigs_querybuilder.skills"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.tasks",
			runconfigs.RunTypeInsert,
			[]string{"task_id"},
			nil,
			[]string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"},
			[]string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}},
				{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id", "user_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.tasks"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.user_skills",
			runconfigs.RunTypeInsert,
			[]string{"user_skill_id"},
			nil,
			[]string{"user_skill_id", "user_id", "skill_id", "proficiency_level"},
			[]string{"user_skill_id", "user_id", "skill_id", "proficiency_level"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}},
				{Table: "genbenthosconfigs_querybuilder.skills", Columns: []string{"skill_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.user_skills"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.attachments",
			runconfigs.RunTypeInsert,
			[]string{"attachment_id"},
			nil,
			[]string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"},
			[]string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.users", Columns: []string{"user_id"}},
				{Table: "genbenthosconfigs_querybuilder.tasks", Columns: []string{"task_id"}},
				{Table: "genbenthosconfigs_querybuilder.initiatives", Columns: []string{"initiative_id"}},
				{Table: "genbenthosconfigs_querybuilder.comments", Columns: []string{"comment_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.attachments"],
		),
	}

	columnInfoMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.attachments": {
			"attachment_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('attachments_attachment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"comment_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"file_name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: 255, NumericPrecision: -1, NumericScale: -1},
			"file_path":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: 255, NumericPrecision: -1, NumericScale: -1},
			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"task_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"uploaded_by":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"genbenthosconfigs_querybuilder.comments": {
			"comment_id":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('comments_comment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"content":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"created_at":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "CURRENT_TIMESTAMP", IsNullable: true, DataType: "timestamp without time zone", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"initiative_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"parent_comment_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"task_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"user_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"genbenthosconfigs_querybuilder.initiatives": {
			"client_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"description":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('initiatives_initiative_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"lead_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"name":          &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
		},
		"genbenthosconfigs_querybuilder.skills": {
			"category": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"skill_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('skills_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"genbenthosconfigs_querybuilder.tasks": {
			"assignee_id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"description":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"reviewer_id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"status":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "character varying(50)", CharacterMaximumLength: 50, NumericPrecision: -1, NumericScale: -1},
			"task_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('tasks_task_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"title":         &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(200)", CharacterMaximumLength: 200, NumericPrecision: -1, NumericScale: -1},
		},
		"genbenthosconfigs_querybuilder.user_skills": {
			"proficiency_level": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"skill_id":          &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"user_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"user_skill_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('user_skills_user_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"genbenthosconfigs_querybuilder.users": {
			"email":      &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"manager_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"mentor_id":  &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"name":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"user_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('users_user_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
	}

	expectedValues := map[string]map[string][]int32{
		"genbenthosconfigs_querybuilder.users": {
			"user_id": {1, 2, 5, 6, 7, 8},
		},
		"genbenthosconfigs_querybuilder.user_skills": {
			"user_skill_id": {1, 2, 5, 6, 7, 8},
			"skill_id":      {1, 2, 5, 6, 7, 8},
			"user_id":       {1, 2, 5, 6, 7, 8},
		},
		"genbenthosconfigs_querybuilder.tasks": {
			"task_id": {5, 6},
		},
		"genbenthosconfigs_querybuilder.skills": {
			"skill_id": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		"genbenthosconfigs_querybuilder.initiatives": {
			"initiative_id": {1, 5, 6, 7},
		},
		"genbenthosconfigs_querybuilder.comments": {
			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
		},
		"genbenthosconfigs_querybuilder.attachments": {
			"attachment_id": {5, 6},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.users":       6,
		"genbenthosconfigs_querybuilder.user_skills": 6,
		"genbenthosconfigs_querybuilder.tasks":       2,
		"genbenthosconfigs_querybuilder.skills":      10,
		"genbenthosconfigs_querybuilder.initiatives": 4,
		"genbenthosconfigs_querybuilder.comments":    13,
		"genbenthosconfigs_querybuilder.attachments": 2,
	}

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfoMap, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))

	allrows := []pgx.Rows{}
	defer func() {
		for _, r := range allrows {
			r.Close()
		}
	}()
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		if slices.Contains([]string{"genbenthosconfigs_querybuilder.skills", "genbenthosconfigs_querybuilder.user_skills", "genbenthosconfigs_querybuilder.users"}, table) {
			assert.Falsef(s.T(), sql.IsNotForeignKeySafeSubset, "table: %s IsNotForeginKeySafeSubset should be false", table)
		} else {
			assert.Truef(s.T(), sql.IsNotForeignKeySafeSubset, "table: %s IsNotForeginKeySafeSubset should be true", table)
		}

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		if rows != nil {
			allrows = append(allrows, rows)
		}
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := col.(int32)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_Pruned_Joins() {
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"genbenthosconfigs_querybuilder.network_users": {
			{Columns: []string{"network_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.networks", ReferenceColumns: []string{"id"}, NotNullable: []bool{false}},
		},
		"genbenthosconfigs_querybuilder.networks": {
			{Columns: []string{"network_type_id"}, ReferenceTable: "genbenthosconfigs_querybuilder.network_types", ReferenceColumns: []string{"id"}, NotNullable: []bool{false}},
		},
	}

	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"genbenthosconfigs_querybuilder.network_types",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "name"},
			[]string{},
			[]*runconfigs.DependsOn{},
			nil,
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.networks",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			nil,
			[]string{"id", "name", "address", "network_type_id"},
			[]string{},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.network_types", Columns: []string{"id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.networks"],
		),
		buildRunConfig(
			"genbenthosconfigs_querybuilder.network_users",
			runconfigs.RunTypeInsert,
			[]string{"id"},
			ptr("username = 'sophia_wilson'"),
			[]string{"id", "username", "email", "password_hash", "first_name", "last_name", "network_id", "created_at"},
			[]string{},
			[]*runconfigs.DependsOn{
				{Table: "genbenthosconfigs_querybuilder.networks", Columns: []string{"network_id"}},
			},
			tableDependencies["genbenthosconfigs_querybuilder.network_users"],
		),
	}

	columnInfoMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"genbenthosconfigs_querybuilder.network_types": {
			"id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"name": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"genbenthosconfigs_querybuilder.networks": {
			"id":              &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"name":            &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"address":         &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "timestamp without time zone", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"network_type_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"genbenthosconfigs_querybuilder.network_users": {
			"id":            &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"username":      &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"email":         &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"password_hash": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"first_name":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"last_name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"network_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"created_at":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 8, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
		},
	}

	expectedValues := map[string]map[string][]int32{
		"genbenthosconfigs_querybuilder.network_types": {
			"id": {1, 2},
		},
		"genbenthosconfigs_querybuilder.networks": {
			"id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.network_users": {
			"id": {8},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.network_types": 2,
		"genbenthosconfigs_querybuilder.networks":      5,
		"genbenthosconfigs_querybuilder.network_users": 1,
	}

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, columnInfoMap, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	allrows := []pgx.Rows{}
	defer func() {
		for _, r := range allrows {
			r.Close()
		}
	}()
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql, "table %s", table)

		rows, err := s.pgcontainer.DB.Query(s.ctx, sql.Query)
		if rows != nil {
			allrows = append(allrows, rows)
		}
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value, ok := col.(int32)
					assert.True(s.T(), ok, "col was not convertable to int32: %s", col)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[table], rowCount, fmt.Sprintf("table: %s ", table))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_ComplexSubset_Mssql() {
	tableDependencies := map[string][]*runconfigs.ForeignKey{
		"mssqltest.attachments": {
			{Columns: []string{"uploaded_by"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{true}},
			{Columns: []string{"task_id"}, ReferenceTable: "mssqltest.tasks", ReferenceColumns: []string{"task_id"}, NotNullable: []bool{true}},
			{Columns: []string{"initiative_id"}, ReferenceTable: "mssqltest.initiatives", ReferenceColumns: []string{"initiative_id"}, NotNullable: []bool{false}},
			{Columns: []string{"comment_id"}, ReferenceTable: "mssqltest.comments", ReferenceColumns: []string{"comment_id"}, NotNullable: []bool{false}},
		},
		"mssqltest.comments": {
			{Columns: []string{"user_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{true}},
			{Columns: []string{"task_id"}, ReferenceTable: "mssqltest.tasks", ReferenceColumns: []string{"task_id"}, NotNullable: []bool{false}},
			{Columns: []string{"initiative_id"}, ReferenceTable: "mssqltest.initiatives", ReferenceColumns: []string{"initiative_id"}, NotNullable: []bool{false}},
			{Columns: []string{"parent_comment_id"}, ReferenceTable: "mssqltest.comments", ReferenceColumns: []string{"comment_id"}, NotNullable: []bool{false}},
		},
		"mssqltest.initiatives": {
			{Columns: []string{"lead_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
			{Columns: []string{"client_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
		},
		"mssqltest.tasks": {
			{Columns: []string{"initiative_id"}, ReferenceTable: "mssqltest.initiatives", ReferenceColumns: []string{"initiative_id"}, NotNullable: []bool{false}},
			{Columns: []string{"assignee_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
			{Columns: []string{"reviewer_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
		},
		"mssqltest.user_skills": {
			{Columns: []string{"user_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{true}},
			{Columns: []string{"skill_id"}, ReferenceTable: "mssqltest.skills", ReferenceColumns: []string{"skill_id"}, NotNullable: []bool{false}},
		},
		"mssqltest.users": {
			{Columns: []string{"manager_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
			{Columns: []string{"mentor_id"}, ReferenceTable: "mssqltest.users", ReferenceColumns: []string{"user_id"}, NotNullable: []bool{false}},
		},
	}

	dependencyConfigs := []*runconfigs.RunConfig{
		buildRunConfig(
			"mssqltest.comments",
			runconfigs.RunTypeInsert,
			[]string{"comment_id"},
			nil,
			[]string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id", "parent_comment_id"},
			[]string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.users", Columns: []string{"user_id"}},
				{Table: "mssqltest.tasks", Columns: []string{"task_id"}},
				{Table: "mssqltest.initiatives", Columns: []string{"initiative_id"}},
			},
			tableDependencies["mssqltest.comments"],
		),
		buildRunConfig(
			"mssqltest.comments",
			runconfigs.RunTypeUpdate,
			[]string{"comment_id"},
			nil,
			[]string{"comment_id", "parent_comment_id"},
			[]string{"parent_comment_id"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.comments", Columns: []string{"comment_id"}},
			},
			tableDependencies["mssqltest.comments"],
		),
		buildRunConfig(
			"mssqltest.users",
			runconfigs.RunTypeInsert,
			[]string{"user_id"},
			ptr("user_id in (1,2,5,6,7,8)"),
			[]string{"user_id", "name", "email", "manager_id", "mentor_id"},
			[]string{"user_id", "name", "email"},
			[]*runconfigs.DependsOn{},
			tableDependencies["mssqltest.users"],
		),
		buildRunConfig(
			"mssqltest.users",
			runconfigs.RunTypeUpdate,
			[]string{"user_id"},
			ptr("user_id = 1"),
			[]string{"user_id", "manager_id", "mentor_id"},
			[]string{"manager_id", "mentor_id"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.users", Columns: []string{"user_id"}},
			},
			tableDependencies["mssqltest.users"],
		),
		buildRunConfig(
			"mssqltest.initiatives",
			runconfigs.RunTypeInsert,
			[]string{"initiative_id"},
			nil,
			[]string{"initiative_id", "name", "description", "lead_id", "client_id"},
			[]string{"initiative_id", "name", "description", "lead_id", "client_id"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.users", Columns: []string{"user_id", "user_id"}},
			},
			tableDependencies["mssqltest.initiatives"],
		),
		buildRunConfig(
			"mssqltest.skills",
			runconfigs.RunTypeInsert,
			[]string{"skill_id"},
			nil,
			[]string{"skill_id", "name", "category"},
			[]string{"skill_id", "name", "category"},
			[]*runconfigs.DependsOn{},
			tableDependencies["mssqltest.skills"],
		),
		buildRunConfig(
			"mssqltest.tasks",
			runconfigs.RunTypeInsert,
			[]string{"task_id"},
			nil,
			[]string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"},
			[]string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.initiatives", Columns: []string{"initiative_id"}},
				{Table: "mssqltest.users", Columns: []string{"user_id", "user_id"}},
			},
			tableDependencies["mssqltest.tasks"],
		),
		buildRunConfig(
			"mssqltest.user_skills",
			runconfigs.RunTypeInsert,
			[]string{"user_skill_id"},
			nil,
			[]string{"user_skill_id", "user_id", "skill_id", "proficiency_level"},
			[]string{"user_skill_id", "user_id", "skill_id", "proficiency_level"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.users", Columns: []string{"user_id"}},
				{Table: "mssqltest.skills", Columns: []string{"skill_id"}},
			},
			tableDependencies["mssqltest.user_skills"],
		),
		buildRunConfig(
			"mssqltest.attachments",
			runconfigs.RunTypeInsert,
			[]string{"attachment_id"},
			nil,
			[]string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"},
			[]string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"},
			[]*runconfigs.DependsOn{
				{Table: "mssqltest.users", Columns: []string{"user_id"}},
				{Table: "mssqltest.tasks", Columns: []string{"task_id"}},
				{Table: "mssqltest.initiatives", Columns: []string{"initiative_id"}},
				{Table: "mssqltest.comments", Columns: []string{"comment_id"}},
			},
			tableDependencies["mssqltest.attachments"],
		),
	}

	columnInfoMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"mssqltest.attachments": {
			"attachment_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('attachments_attachment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"comment_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"file_name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: 255, NumericPrecision: -1, NumericScale: -1},
			"file_path":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: 255, NumericPrecision: -1, NumericScale: -1},
			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"task_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"uploaded_by":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"mssqltest.comments": {
			"comment_id":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('comments_comment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"content":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"created_at":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "CURRENT_TIMESTAMP", IsNullable: true, DataType: "timestamp without time zone", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"initiative_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"parent_comment_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"task_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"user_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"mssqltest.initiatives": {
			"client_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"description":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('initiatives_initiative_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"lead_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"name":          &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
		},
		"mssqltest.skills": {
			"category": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"skill_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('skills_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"mssqltest.tasks": {
			"assignee_id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"description":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"reviewer_id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"status":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "character varying(50)", CharacterMaximumLength: 50, NumericPrecision: -1, NumericScale: -1},
			"task_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('tasks_task_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"title":         &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(200)", CharacterMaximumLength: 200, NumericPrecision: -1, NumericScale: -1},
		},
		"mssqltest.user_skills": {
			"proficiency_level": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"skill_id":          &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"user_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"user_skill_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('user_skills_user_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
		"mssqltest.users": {
			"email":      &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"manager_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"mentor_id":  &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
			"name":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
			"user_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('users_user_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
		},
	}

	expectedValues := map[string]map[string][]int64{
		"mssqltest.users": {
			"user_id": {1, 2, 5, 6, 7, 8},
		},
		"mssqltest.user_skills": {
			"user_skill_id": {1, 2, 5, 6, 7, 8},
			"skill_id":      {1, 2, 5, 6, 7, 8},
			"user_id":       {1, 2, 5, 6, 7, 8},
		},
		"mssqltest.tasks": {
			"task_id": {5, 6},
		},
		"mssqltest.skills": {
			"skill_id": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		"mssqltest.initiatives": {
			"initiative_id": {1, 5, 6, 7},
		},
		"mssqltest.comments": {
			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
		},
		"mssqltest.attachments": {
			"attachment_id": {5, 6},
		},
	}

	expectedCount := map[string]int{
		"mssqltest.users":       6,
		"mssqltest.user_skills": 6,
		"mssqltest.tasks":       2,
		"mssqltest.skills":      10,
		"mssqltest.initiatives": 4,
		"mssqltest.comments":    13,
		"mssqltest.attachments": 2,
	}

	sqlMap, err := querybuilder2.BuildSelectQueryMap(sqlmanager_shared.MssqlDriver, dependencyConfigs, true, columnInfoMap, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))
	for table, selectQueryRunType := range sqlMap {
		sql := selectQueryRunType[runconfigs.RunTypeInsert]
		assert.NotEmpty(s.T(), sql)

		if slices.Contains([]string{"mssqltest.skills", "mssqltest.user_skills", "mssqltest.users"}, table) {
			assert.Falsef(s.T(), sql.IsNotForeignKeySafeSubset, "table: %s IsNotForeginKeySafeSubset should be false", table)
		} else {
			assert.Truef(s.T(), sql.IsNotForeignKeySafeSubset, "table: %s IsNotForeginKeySafeSubset should be true", table)
		}

		rows, err := s.mssql.pool.QueryContext(s.ctx, sql.Query)
		assert.NoError(s.T(), err)

		columns, err := rows.Columns()
		assert.NoError(s.T(), err)

		tableExpectedValues, ok := expectedValues[table]
		assert.Truef(s.T(), ok, fmt.Sprintf("table: %s missing expected values", table))

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		rowCount := 0
		for rows.Next() {
			rowCount++
			err = rows.Scan(valuePtrs...)
			assert.NoError(s.T(), err)

			for i, colName := range columns {
				val := values[i]
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					value := val.(int64)
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("table: %s column: %s ", table, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), rowCount, expectedCount[table], fmt.Sprintf("table: %s ", table))
	}
}

func ptr[T any](input T) *T {
	return &input
}

func buildRunConfig(
	table string,
	runtype runconfigs.RunType,
	pks []string,
	where *string,
	selectCols, insertCols []string,
	dependsOn []*runconfigs.DependsOn,
	foreignKeys []*runconfigs.ForeignKey,
) *runconfigs.RunConfig {
	return runconfigs.NewRunConfig(table, runtype, pks, where, selectCols, insertCols, dependsOn, foreignKeys, false)
}
