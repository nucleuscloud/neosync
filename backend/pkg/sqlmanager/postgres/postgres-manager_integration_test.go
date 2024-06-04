package sqlmanager_postgres

import (
	"context"

	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetDatabaseSchema() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	// expected := []*DatabaseSchemaRow{
	// 	{
	// 		TableSchema:            "public",
	// 		TableName:              "users",
	// 		ColumnName:             "id",
	// 		DataType:               "varchar",
	// 		ColumnDefault:          "",
	// 		IsNullable:             "NO",
	// 		CharacterMaximumLength: 220,
	// 		NumericPrecision:       -1,
	// 		NumericScale:           -1,
	// 		OrdinalPosition:        4,
	// 	},
	// 	{
	// 		TableSchema:            "public",
	// 		TableName:              "orders",
	// 		ColumnName:             "buyer_id",
	// 		DataType:               "integer",
	// 		ColumnDefault:          "",
	// 		IsNullable:             "NO",
	// 		CharacterMaximumLength: -1,
	// 		NumericPrecision:       32,
	// 		NumericScale:           0,
	// 		OrdinalPosition:        5,
	// 	},
	// }

	actual, err := manager.GetDatabaseSchema(context.Background())
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	// require.ElementsMatch(s.T(), expected, actual)
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	actual, err := manager.GetForeignKeyConstraintsMap(s.ctx, []string{"public"})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
}
