package sqlmanager_mysql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetDatabaseSchema() {
	manager := MysqlManager{querier: s.querier, pool: s.pool}

	// expectedSubset := []*sqlmanager_shared.DatabaseSchemaRow{
	// 	{
	// 		TableSchema:            s.schema,
	// 		TableName:              "users",
	// 		ColumnName:             "id",
	// 		DataType:               "text",
	// 		ColumnDefault:          "",
	// 		IsNullable:             "NO",
	// 		CharacterMaximumLength: -1,
	// 		NumericPrecision:       -1,
	// 		NumericScale:           -1,
	// 		OrdinalPosition:        1,
	// 	},
	// }

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{"sqlmanangermysql", "sqlmanagermysql2"})
	require.NoError(s.T(), err)
	// containsSubset(s.T(), actual, expectedSubset)
	jsonF, _ := json.MarshalIndent(actual, "", " ")
	fmt.Printf("%s \n", string(jsonF))
	require.Error(s.T(), err)

}
