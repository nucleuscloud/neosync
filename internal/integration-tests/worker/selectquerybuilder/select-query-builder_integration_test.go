package selectquerybuilder

import (
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

	runconfigs "github.com/nucleuscloud/neosync/internal/runconfigs"

	selectbuilder "github.com/nucleuscloud/neosync/worker/pkg/select-query-builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pageLimit = 100
)

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleReference() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.company": "id = 1",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.company":        {},
		"genbenthosconfigs_querybuilder.department":     {},
		"genbenthosconfigs_querybuilder.transaction":    {},
		"genbenthosconfigs_querybuilder.expense_report": {},
		"genbenthosconfigs_querybuilder.expense":        {},
		"genbenthosconfigs_querybuilder.item":           {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.company.insert": {
			"id": {1},
		},
		"genbenthosconfigs_querybuilder.department.insert": {
			"company_id": {1},
		},
		"genbenthosconfigs_querybuilder.expense_report.insert": {
			"department_source_id":      {1, 2, 3},
			"department_destination_id": {1, 2},
		},
		"genbenthosconfigs_querybuilder.expense_report.update.1": {
			"department_source_id":      {1, 2, 3},
			"department_destination_id": {1, 2},
		},
		"genbenthosconfigs_querybuilder.expense_report.update.2": {
			"department_source_id":      {1, 2, 3},
			"department_destination_id": {1, 2},
		},
		"genbenthosconfigs_querybuilder.transaction.insert": {
			"department_id": {1, 2},
		},
		"genbenthosconfigs_querybuilder.expense.insert": {
			"report_id": {3, 1, 2},
		},
		"genbenthosconfigs_querybuilder.item.insert": {
			"expense_id": {3, 2, 1},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.company.insert":          1,
		"genbenthosconfigs_querybuilder.department.insert":       2,
		"genbenthosconfigs_querybuilder.expense_report.insert":   3,
		"genbenthosconfigs_querybuilder.expense_report.update.1": 3,
		"genbenthosconfigs_querybuilder.expense_report.update.2": 3,
		"genbenthosconfigs_querybuilder.transaction.insert":      2,
		"genbenthosconfigs_querybuilder.expense.insert":          3,
		"genbenthosconfigs_querybuilder.item.insert":             3,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err, "failed to build select query map")
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_DoubleRootSubset() {
	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.test_2_x": {},
		"genbenthosconfigs_querybuilder.test_2_b": {},
		"genbenthosconfigs_querybuilder.test_2_a": {},
		"genbenthosconfigs_querybuilder.test_2_c": {},
		"genbenthosconfigs_querybuilder.test_2_d": {},
		"genbenthosconfigs_querybuilder.test_2_e": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.test_2_x": "created > '2023-06-03'",
		"genbenthosconfigs_querybuilder.test_2_b": "created > '2023-06-03'",
	}
	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.test_2_x.insert": {
			"id": {3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_2_b.insert": {
			"id": {3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_2_a.insert": {
			"x_id": {3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_2_c.insert": {
			"a_id": {3, 4},
			"x_id": {3, 4},
		},
		"genbenthosconfigs_querybuilder.test_2_d.insert": {
			"c_id": {3, 4},
		},
		"genbenthosconfigs_querybuilder.test_2_e.insert": {
			"c_id": {3, 4},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.test_2_x.insert": 3,
		"genbenthosconfigs_querybuilder.test_2_b.insert": 3,
		"genbenthosconfigs_querybuilder.test_2_a.insert": 4,
		"genbenthosconfigs_querybuilder.test_2_c.insert": 2,
		"genbenthosconfigs_querybuilder.test_2_d.insert": 2,
		"genbenthosconfigs_querybuilder.test_2_e.insert": 2,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleRoots() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.test_3_b": "id = 1",
		"genbenthosconfigs_querybuilder.test_3_f": "id in (4,5)",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.test_3_a": {},
		"genbenthosconfigs_querybuilder.test_3_b": {},
		"genbenthosconfigs_querybuilder.test_3_c": {},
		"genbenthosconfigs_querybuilder.test_3_d": {},
		"genbenthosconfigs_querybuilder.test_3_e": {},
		"genbenthosconfigs_querybuilder.test_3_f": {},
		"genbenthosconfigs_querybuilder.test_3_g": {},
		"genbenthosconfigs_querybuilder.test_3_h": {},
		"genbenthosconfigs_querybuilder.test_3_i": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.test_3_a.insert": {
			"id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_b.insert": {
			"a_id": {3},
		},
		"genbenthosconfigs_querybuilder.test_3_c.insert": {
			"b_id": {1},
		},
		"genbenthosconfigs_querybuilder.test_3_d.insert": {
			"c_id": {3},
		},
		"genbenthosconfigs_querybuilder.test_3_e.insert": {
			"d_id": {5},
		},
		"genbenthosconfigs_querybuilder.test_3_f.insert": {
			"id": {4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_g.insert": {
			"f_id": {4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_h.insert": {
			"g_id": {1, 3},
		},
		"genbenthosconfigs_querybuilder.test_3_i.insert": {
			"h_id": {4},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.test_3_a.insert": 5,
		"genbenthosconfigs_querybuilder.test_3_b.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_c.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_d.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_e.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_f.insert": 2,
		"genbenthosconfigs_querybuilder.test_3_g.insert": 2,
		"genbenthosconfigs_querybuilder.test_3_h.insert": 2,
		"genbenthosconfigs_querybuilder.test_3_i.insert": 1,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleSubsets() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.test_3_a": "id in (3,4,5)",
		"genbenthosconfigs_querybuilder.test_3_b": "id = 4",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.test_3_a": {},
		"genbenthosconfigs_querybuilder.test_3_b": {},
		"genbenthosconfigs_querybuilder.test_3_c": {},
		"genbenthosconfigs_querybuilder.test_3_d": {},
		"genbenthosconfigs_querybuilder.test_3_e": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.test_3_a.insert": {
			"id": {3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_b.insert": {
			"a_id": {4},
		},
		"genbenthosconfigs_querybuilder.test_3_c.insert": {
			"b_id": {4},
		},
		"genbenthosconfigs_querybuilder.test_3_d.insert": {
			"c_id": {2},
		},
		"genbenthosconfigs_querybuilder.test_3_e.insert": {
			"d_id": {4},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.test_3_a.insert": 3,
		"genbenthosconfigs_querybuilder.test_3_b.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_c.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_d.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_e.insert": 1,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)

}

func (s *IntegrationTestSuite) Test_BuildQueryMap_MultipleSubsets_SubsetsByForeignKeysOff() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.test_3_a": "id in (4,5)",
		"genbenthosconfigs_querybuilder.test_3_b": "id = 4",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.test_3_a": {},
		"genbenthosconfigs_querybuilder.test_3_b": {},
		"genbenthosconfigs_querybuilder.test_3_c": {},
		"genbenthosconfigs_querybuilder.test_3_d": {},
		"genbenthosconfigs_querybuilder.test_3_e": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.test_3_a.insert": {
			"id": {4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_b.insert": {
			"a_id": {4},
		},
		"genbenthosconfigs_querybuilder.test_3_c.insert": {
			"b_id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_d.insert": {
			"c_id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_e.insert": {
			"d_id": {1, 2, 3, 4, 5},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.test_3_a.insert": 2,
		"genbenthosconfigs_querybuilder.test_3_b.insert": 1,
		"genbenthosconfigs_querybuilder.test_3_c.insert": 5,
		"genbenthosconfigs_querybuilder.test_3_d.insert": 5,
		"genbenthosconfigs_querybuilder.test_3_e.insert": 5,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, false, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_CircularDependency() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.addresses": "id in (1,5)",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.addresses": {},
		"genbenthosconfigs_querybuilder.customers": {},
		"genbenthosconfigs_querybuilder.orders":    {},
		"genbenthosconfigs_querybuilder.payments":  {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.orders.insert": {
			"customer_id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.addresses.insert": {
			"order_id": {1, 5},
		},
		"genbenthosconfigs_querybuilder.customers.insert": {
			"address_id": {1, 5},
		},
		"genbenthosconfigs_querybuilder.payments.insert": {
			"customer_id": {2},
		},
		"genbenthosconfigs_querybuilder.orders.update.1": {
			"customer_id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.addresses.update.1": {
			"order_id": {1, 5},
		},
		"genbenthosconfigs_querybuilder.customers.update.1": {
			"address_id": {1, 5},
		},
		"genbenthosconfigs_querybuilder.payments.update.1": {
			"customer_id": {2},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.orders.insert":      2,
		"genbenthosconfigs_querybuilder.addresses.insert":   2,
		"genbenthosconfigs_querybuilder.customers.insert":   2,
		"genbenthosconfigs_querybuilder.payments.insert":    1,
		"genbenthosconfigs_querybuilder.orders.update.1":    2,
		"genbenthosconfigs_querybuilder.addresses.update.1": 2,
		"genbenthosconfigs_querybuilder.customers.update.1": 2,
		"genbenthosconfigs_querybuilder.payments.update.1":  1,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_NoForeignKeys() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.test_2_x": "id in (1,5)",
		"genbenthosconfigs_querybuilder.test_2_b": "id in (1,5)",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.company":  {},
		"genbenthosconfigs_querybuilder.test_2_x": {},
		"genbenthosconfigs_querybuilder.test_2_b": {},
		"genbenthosconfigs_querybuilder.test_3_a": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.company.insert": {
			"id": {1, 2, 3},
		},
		"genbenthosconfigs_querybuilder.test_2_x.insert": {
			"id": {1, 5},
		},
		"genbenthosconfigs_querybuilder.test_2_b.insert": {
			"id": {1, 5},
		},
		"genbenthosconfigs_querybuilder.test_3_a.insert": {
			"customer_id": {1, 2, 3, 4, 5},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.company.insert":  3,
		"genbenthosconfigs_querybuilder.test_2_x.insert": 2,
		"genbenthosconfigs_querybuilder.test_2_b.insert": 2,
		"genbenthosconfigs_querybuilder.test_3_a.insert": 5,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_NoForeignKeys_NoSubsets() {
	subsets := map[string]string{}
	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.company":  {},
		"genbenthosconfigs_querybuilder.test_2_x": {},
		"genbenthosconfigs_querybuilder.test_2_b": {},
		"genbenthosconfigs_querybuilder.test_3_a": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.company.insert":  3,
		"genbenthosconfigs_querybuilder.test_2_x.insert": 5,
		"genbenthosconfigs_querybuilder.test_2_b.insert": 5,
		"genbenthosconfigs_querybuilder.test_3_a.insert": 5,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedCount), len(sqlMap))
	for table, query := range sqlMap {
		rows, err := s.postgres.pgcontainer.DB.Query(s.ctx, query.Query)
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
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.division": "id in (3,5)",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.division":  {},
		"genbenthosconfigs_querybuilder.employees": {},
		"genbenthosconfigs_querybuilder.projects":  {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.division.insert": {
			"id": {3, 5},
		},
		"genbenthosconfigs_querybuilder.employees.insert": {
			"division_id": {3, 5},
			"id":          {8, 10},
		},
		"genbenthosconfigs_querybuilder.projects.insert": {
			"responsible_division_id": {3, 5},
			"responsible_employee_id": {8, 10},
		},
		"genbenthosconfigs_querybuilder.projects.update.1": {
			"responsible_division_id": {3, 5},
			"responsible_employee_id": {8, 10},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.division.insert":   2,
		"genbenthosconfigs_querybuilder.employees.insert":  2,
		"genbenthosconfigs_querybuilder.projects.insert":   2,
		"genbenthosconfigs_querybuilder.projects.update.1": 2,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_ComplexSubset_Postgres() {
	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.users": "user_id in (1,2,5,6,7,8)",
	}

	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.users":       {},
		"genbenthosconfigs_querybuilder.user_skills": {},
		"genbenthosconfigs_querybuilder.tasks":       {},
		"genbenthosconfigs_querybuilder.skills":      {},
		"genbenthosconfigs_querybuilder.initiatives": {},
		"genbenthosconfigs_querybuilder.comments":    {},
		"genbenthosconfigs_querybuilder.attachments": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int32{
		"genbenthosconfigs_querybuilder.users.insert": {
			"user_id": {1, 2, 5, 6, 7, 8},
		},
		"genbenthosconfigs_querybuilder.users.update.1": {
			"user_id": {1, 2, 5, 6, 7, 8},
		},
		"genbenthosconfigs_querybuilder.users.update.2": {
			"user_id": {1, 2, 5, 6, 7, 8},
		},
		"genbenthosconfigs_querybuilder.user_skills.insert": {
			"user_skill_id": {1, 2, 5, 6, 7, 8},
			"skill_id":      {1, 2, 5, 6, 7, 8},
			"user_id":       {1, 2, 5, 6, 7, 8},
		},
		"genbenthosconfigs_querybuilder.tasks.insert": {
			"task_id": {3, 4, 5, 6, 9, 10},
		},
		"genbenthosconfigs_querybuilder.skills.insert": {
			"skill_id": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		"genbenthosconfigs_querybuilder.initiatives.insert": {
			"initiative_id": {1, 4, 5, 6, 7, 10},
		},
		"genbenthosconfigs_querybuilder.comments.insert": {
			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
		},
		"genbenthosconfigs_querybuilder.comments.update.1": {
			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
		},
		"genbenthosconfigs_querybuilder.comments.update.2": {
			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
		},
		"genbenthosconfigs_querybuilder.comments.update.3": {
			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
		},
		"genbenthosconfigs_querybuilder.attachments.insert": {
			"attachment_id": {3, 4, 5, 6, 9, 10},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.users.insert":       6,
		"genbenthosconfigs_querybuilder.users.update.1":     6,
		"genbenthosconfigs_querybuilder.users.update.2":     6,
		"genbenthosconfigs_querybuilder.user_skills.insert": 6,
		"genbenthosconfigs_querybuilder.tasks.insert":       6,
		"genbenthosconfigs_querybuilder.skills.insert":      10,
		"genbenthosconfigs_querybuilder.initiatives.insert": 6,
		"genbenthosconfigs_querybuilder.comments.insert":    13,
		"genbenthosconfigs_querybuilder.comments.update.1":  13,
		"genbenthosconfigs_querybuilder.comments.update.2":  13,
		"genbenthosconfigs_querybuilder.comments.update.3":  13,
		"genbenthosconfigs_querybuilder.attachments.insert": 6,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	require.Equal(s.T(), len(expectedValues), len(sqlMap))

	allrows := []pgx.Rows{}
	defer func() {
		for _, r := range allrows {
			r.Close()
		}
	}()
	for id, query := range sqlMap {
		assert.NotEmpty(s.T(), query.Query)

		if slices.Contains([]string{"genbenthosconfigs_querybuilder.users.insert", "genbenthosconfigs_querybuilder.skills.insert"}, id) {
			assert.Falsef(s.T(), query.IsNotForeignKeySafeSubset, "id: %s IsNotForeginKeySafeSubset should be false", id)
		} else {
			assert.Truef(s.T(), query.IsNotForeignKeySafeSubset, "id: %s IsNotForeginKeySafeSubset should be true", id)
		}

		rows, err := s.postgres.pgcontainer.DB.Query(s.ctx, query.Query)
		if rows != nil {
			allrows = append(allrows, rows)
		}
		assert.NoError(s.T(), err)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[id]
		assert.Truef(s.T(), ok, fmt.Sprintf("id: %s missing expected values", id))

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
					assert.Containsf(s.T(), allowedValues, value, fmt.Sprintf("id: %s column: %s ", id, colName))
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[id], rowCount, fmt.Sprintf("id: %s ", id))
	}
}

func (s *IntegrationTestSuite) Test_BuildQueryMap_Pruned_Joins() {
	tables := map[string]struct{}{
		"genbenthosconfigs_querybuilder.network_types": {},
		"genbenthosconfigs_querybuilder.networks":      {},
		"genbenthosconfigs_querybuilder.network_users": {},
	}
	tableColumnsMap := map[string][]string{}
	for table, columns := range s.postgres.groupedColumnInfo {
		if _, ok := tables[table]; !ok {
			continue
		}
		for col := range columns {
			tableColumnsMap[table] = append(tableColumnsMap[table], col)
		}
	}

	subsets := map[string]string{
		"genbenthosconfigs_querybuilder.network_users": "username = 'sophia_wilson'",
	}
	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
	require.NoError(s.T(), err)

	expectedValues := map[string]map[string][]int64{
		"genbenthosconfigs_querybuilder.network_types.insert": {
			"id": {1, 2},
		},
		"genbenthosconfigs_querybuilder.networks.insert": {
			"id": {1, 2, 3, 4, 5},
		},
		"genbenthosconfigs_querybuilder.network_users.insert": {
			"id": {8},
		},
		"genbenthosconfigs_querybuilder.network_users.update.1": {
			"id": {8},
		},
	}

	expectedCount := map[string]int{
		"genbenthosconfigs_querybuilder.network_types.insert":   2,
		"genbenthosconfigs_querybuilder.networks.insert":        5,
		"genbenthosconfigs_querybuilder.network_users.insert":   1,
		"genbenthosconfigs_querybuilder.network_users.update.1": 1,
	}

	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.PostgresDriver, dependencyConfigs, true, pageLimit)
	require.NoError(s.T(), err)
	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
}

// func (s *IntegrationTestSuite) Test_BuildQueryMap_ComplexSubset_Mssql() {
//   tables := map[string]struct{}{
// 		"genbenthosconfigs_querybuilder.network_types": {},
// 		"genbenthosconfigs_querybuilder.networks":      {},
// 		"genbenthosconfigs_querybuilder.network_users": {},
// 	}
// 	tableColumnsMap := map[string][]string{}
// 	for table, columns := range s.postgres.groupedColumnInfo {
// 		if _, ok := tables[table]; !ok {
// 			continue
// 		}
// 		for col := range columns {
// 			tableColumnsMap[table] = append(tableColumnsMap[table], col)
// 		}
// 	}

// 	subsets := map[string]string{
// 		"genbenthosconfigs_querybuilder.network_users": "username = 'sophia_wilson'",
// 	}
// 	dependencyConfigs, err := runconfigs.BuildRunConfigs(s.postgres.tableConstraints.ForeignKeyConstraints, subsets, s.postgres.tableConstraints.PrimaryKeyConstraints, tableColumnsMap, nil, nil)
// 	require.NoError(s.T(), err)

// 	dependencyConfigs := []*runconfigs.RunConfig{
// 		buildRunConfig(
// 			"mssqltest.comments",
// 			runconfigs.RunTypeInsert,
// 			[]string{"comment_id"},
// 			nil,
// 			[]string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id", "parent_comment_id"},
// 			[]string{"comment_id", "content", "created_at", "user_id", "task_id", "initiative_id"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.users", Columns: []string{"user_id"}},
// 				{Table: "mssqltest.tasks", Columns: []string{"task_id"}},
// 				{Table: "mssqltest.initiatives", Columns: []string{"initiative_id"}},
// 			},
// 			tableDependencies["mssqltest.comments"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.comments",
// 			runconfigs.RunTypeUpdate,
// 			[]string{"comment_id"},
// 			nil,
// 			[]string{"comment_id", "parent_comment_id"},
// 			[]string{"parent_comment_id"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.comments", Columns: []string{"comment_id"}},
// 			},
// 			tableDependencies["mssqltest.comments"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.users",
// 			runconfigs.RunTypeInsert,
// 			[]string{"user_id"},
// 			ptr("user_id in (1,2,5,6,7,8)"),
// 			[]string{"user_id", "name", "email", "manager_id", "mentor_id"},
// 			[]string{"user_id", "name", "email"},
// 			[]*runconfigs.DependsOn{},
// 			tableDependencies["mssqltest.users"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.users",
// 			runconfigs.RunTypeUpdate,
// 			[]string{"user_id"},
// 			ptr("user_id = 1"),
// 			[]string{"user_id", "manager_id", "mentor_id"},
// 			[]string{"manager_id", "mentor_id"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.users", Columns: []string{"user_id"}},
// 			},
// 			tableDependencies["mssqltest.users"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.initiatives",
// 			runconfigs.RunTypeInsert,
// 			[]string{"initiative_id"},
// 			nil,
// 			[]string{"initiative_id", "name", "description", "lead_id", "client_id"},
// 			[]string{"initiative_id", "name", "description", "lead_id", "client_id"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.users", Columns: []string{"user_id", "user_id"}},
// 			},
// 			tableDependencies["mssqltest.initiatives"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.skills",
// 			runconfigs.RunTypeInsert,
// 			[]string{"skill_id"},
// 			nil,
// 			[]string{"skill_id", "name", "category"},
// 			[]string{"skill_id", "name", "category"},
// 			[]*runconfigs.DependsOn{},
// 			tableDependencies["mssqltest.skills"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.tasks",
// 			runconfigs.RunTypeInsert,
// 			[]string{"task_id"},
// 			nil,
// 			[]string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"},
// 			[]string{"task_id", "title", "description", "status", "initiative_id", "assignee_id", "reviewer_id"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.initiatives", Columns: []string{"initiative_id"}},
// 				{Table: "mssqltest.users", Columns: []string{"user_id", "user_id"}},
// 			},
// 			tableDependencies["mssqltest.tasks"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.user_skills",
// 			runconfigs.RunTypeInsert,
// 			[]string{"user_skill_id"},
// 			nil,
// 			[]string{"user_skill_id", "user_id", "skill_id", "proficiency_level"},
// 			[]string{"user_skill_id", "user_id", "skill_id", "proficiency_level"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.users", Columns: []string{"user_id"}},
// 				{Table: "mssqltest.skills", Columns: []string{"skill_id"}},
// 			},
// 			tableDependencies["mssqltest.user_skills"],
// 		),
// 		buildRunConfig(
// 			"mssqltest.attachments",
// 			runconfigs.RunTypeInsert,
// 			[]string{"attachment_id"},
// 			nil,
// 			[]string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"},
// 			[]string{"attachment_id", "file_name", "file_path", "uploaded_by", "task_id", "initiative_id", "comment_id"},
// 			[]*runconfigs.DependsOn{
// 				{Table: "mssqltest.users", Columns: []string{"user_id"}},
// 				{Table: "mssqltest.tasks", Columns: []string{"task_id"}},
// 				{Table: "mssqltest.initiatives", Columns: []string{"initiative_id"}},
// 				{Table: "mssqltest.comments", Columns: []string{"comment_id"}},
// 			},
// 			tableDependencies["mssqltest.attachments"],
// 		),
// 	}

// 	columnInfoMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
// 		"mssqltest.attachments": {
// 			"attachment_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('attachments_attachment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"comment_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"file_name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: 255, NumericPrecision: -1, NumericScale: -1},
// 			"file_path":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(255)", CharacterMaximumLength: 255, NumericPrecision: -1, NumericScale: -1},
// 			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"task_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"uploaded_by":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 		},
// 		"mssqltest.comments": {
// 			"comment_id":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('comments_comment_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"content":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
// 			"created_at":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "CURRENT_TIMESTAMP", IsNullable: true, DataType: "timestamp without time zone", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
// 			"initiative_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"parent_comment_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"task_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"user_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 		},
// 		"mssqltest.initiatives": {
// 			"client_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"description":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
// 			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('initiatives_initiative_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"lead_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"name":          &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
// 		},
// 		"mssqltest.skills": {
// 			"category": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
// 			"name":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
// 			"skill_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('skills_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 		},
// 		"mssqltest.tasks": {
// 			"assignee_id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 6, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"description":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "text", CharacterMaximumLength: -1, NumericPrecision: -1, NumericScale: -1},
// 			"initiative_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"reviewer_id":   &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 7, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"status":        &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "character varying(50)", CharacterMaximumLength: 50, NumericPrecision: -1, NumericScale: -1},
// 			"task_id":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('tasks_task_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"title":         &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(200)", CharacterMaximumLength: 200, NumericPrecision: -1, NumericScale: -1},
// 		},
// 		"mssqltest.user_skills": {
// 			"proficiency_level": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"skill_id":          &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"user_id":           &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"user_skill_id":     &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('user_skills_user_skill_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 		},
// 		"mssqltest.users": {
// 			"email":      &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 3, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
// 			"manager_id": &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 4, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"mentor_id":  &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 5, ColumnDefault: "", IsNullable: true, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 			"name":       &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 2, ColumnDefault: "", IsNullable: false, DataType: "character varying(100)", CharacterMaximumLength: 100, NumericPrecision: -1, NumericScale: -1},
// 			"user_id":    &sqlmanager_shared.DatabaseSchemaRow{OrdinalPosition: 1, ColumnDefault: "nextval('users_user_id_seq'::regclass)", IsNullable: false, DataType: "integer", CharacterMaximumLength: -1, NumericPrecision: 32, NumericScale: 0},
// 		},
// 	}

// 	expectedValues := map[string]map[string][]int64{
// 		"mssqltest.users": {
// 			"user_id": {1, 2, 5, 6, 7, 8},
// 		},
// 		"mssqltest.user_skills": {
// 			"user_skill_id": {1, 2, 5, 6, 7, 8},
// 			"skill_id":      {1, 2, 5, 6, 7, 8},
// 			"user_id":       {1, 2, 5, 6, 7, 8},
// 		},
// 		"mssqltest.tasks": {
// 			"task_id": {5, 6},
// 		},
// 		"mssqltest.skills": {
// 			"skill_id": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
// 		},
// 		"mssqltest.initiatives": {
// 			"initiative_id": {1, 5, 6, 7},
// 		},
// 		"mssqltest.comments": {
// 			"comment_id": {1, 3, 6, 8, 9, 10, 11, 12, 13, 15, 18, 20, 21},
// 		},
// 		"mssqltest.attachments": {
// 			"attachment_id": {5, 6},
// 		},
// 	}

// 	expectedCount := map[string]int{
// 		"mssqltest.users":       6,
// 		"mssqltest.user_skills": 6,
// 		"mssqltest.tasks":       2,
// 		"mssqltest.skills":      10,
// 		"mssqltest.initiatives": 4,
// 		"mssqltest.comments":    13,
// 		"mssqltest.attachments": 2,
// 	}

// 	sqlMap, err := selectbuilder.BuildSelectQueryMap(sqlmanager_shared.MssqlDriver, dependencyConfigs, true, columnInfoMap, pageLimit)
// 	require.NoError(s.T(), err)
// 	s.assertQueryMap(sqlMap, expectedValues, expectedCount)
// }

func (s *IntegrationTestSuite) assertQueryMap(sqlMap map[string]*sqlmanager_shared.SelectQuery, expectedValues map[string]map[string][]int64, expectedCount map[string]int) {
	require.Equal(s.T(), len(expectedValues), len(sqlMap), "number of queries in sqlMap doesn't match expected values")
	for configId, query := range sqlMap {
		rows, err := s.postgres.pgcontainer.DB.Query(s.ctx, query.Query)
		assert.NoError(s.T(), err, "failed to execute query for config %s: %s", configId, query.Query)

		columnDescriptions := rows.FieldDescriptions()

		tableExpectedValues, ok := expectedValues[configId]
		assert.True(s.T(), ok, "missing expected values for config %s", configId)

		rowCount := 0
		for rows.Next() {
			rowCount++
			values, err := rows.Values()
			assert.NoError(s.T(), err, "failed to get row values for config %s", configId)

			for i, col := range values {
				colName := columnDescriptions[i].Name
				allowedValues, ok := tableExpectedValues[colName]
				if ok {
					var value int64
					switch v := col.(type) {
					case int32:
						value = int64(v)
					case int64:
						value = v
					default:
						assert.Failf(s.T(), "unexpected type for column %s", "expected int32 or int64, got %T for column %s", col, colName)
					}
					assert.Containsf(s.T(), allowedValues, value, "config %s: column %s value %d not in expected values %v", configId, colName, value, allowedValues)
				}
			}
		}
		rows.Close()
		assert.Equalf(s.T(), expectedCount[configId], rowCount, "config %s: row count %d doesn't match expected count %d", configId, rowCount, expectedCount[configId])
	}
}
