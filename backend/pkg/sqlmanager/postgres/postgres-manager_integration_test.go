package sqlmanager_postgres

import (
	"context"
	"fmt"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetDatabaseSchema() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	expectedSubset := []*sqlmanager_shared.DatabaseSchemaRow{
		{
			TableSchema:            s.schema,
			TableName:              "users",
			ColumnName:             "id",
			DataType:               "text",
			ColumnDefault:          "",
			IsNullable:             "NO",
			CharacterMaximumLength: -1,
			NumericPrecision:       -1,
			NumericScale:           -1,
			OrdinalPosition:        1,
		},
	}

	actual, err := manager.GetDatabaseSchema(context.Background())
	require.NoError(s.T(), err)
	containsSubset(s.T(), actual, expectedSubset)
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{s.schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	constraints, ok := actual.ForeignKeyConstraints[fmt.Sprintf("%s.child1", s.schema)]
	require.True(s.T(), ok)
	require.NotEmpty(s.T(), constraints)

	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"parent_id"},
			NotNullable: []bool{false},
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Table:   fmt.Sprintf("%s.parent1", s.schema),
				Columns: []string{"id"},
			},
		},
	})
}

func containsSubset[T any](t testing.TB, array, subset []T) {
	t.Helper()
	for _, elem := range subset {
		require.Contains(t, array, elem)
	}
}
