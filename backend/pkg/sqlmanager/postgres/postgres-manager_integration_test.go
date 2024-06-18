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

func (s *IntegrationTestSuite) Test_GetDatabaseSchema_With_Identity() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	expectedSubset := []*sqlmanager_shared.DatabaseSchemaRow{
		{
			TableSchema:            s.schema,
			TableName:              "users_with_identity",
			ColumnName:             "id",
			DataType:               "integer",
			ColumnDefault:          "",
			IsNullable:             "NO",
			CharacterMaximumLength: -1,
			NumericPrecision:       32,
			NumericScale:           0,
			OrdinalPosition:        1,
			IdentityGeneration:     sqlmanager_shared.Ptr("a"),
		},
	}

	actual, err := manager.GetDatabaseSchema(context.Background())
	require.NoError(s.T(), err)
	containsSubset(s.T(), actual, expectedSubset)
}

func (s *IntegrationTestSuite) Test_GetSchemaColumnMap() {
	manager := NewManager(s.querier, s.pgpool, func() {})

	actual, err := manager.GetSchemaColumnMap(context.Background())
	require.NoError(s.T(), err)

	usersKey := s.buildTable("users")

	usersMap, ok := actual[usersKey]
	require.True(s.T(), ok, fmt.Sprintf("%s map should exist", usersKey))
	require.NotEmpty(s.T(), usersMap)
	_, ok = usersMap["id"]
	require.True(s.T(), ok, "users map should have id column")
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

func (s *IntegrationTestSuite) Test_GetRolePermissionsMap() {
	manager := NewManager(s.querier, s.pgpool, func() {})

	actual, err := manager.GetRolePermissionsMap(context.Background())
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	usersKey := s.buildTable("users")

	usersRecord, ok := actual[usersKey]
	require.True(s.T(), ok, "map should have users perms")
	require.ElementsMatch(
		s.T(),
		[]string{"INSERT", "SELECT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
		usersRecord,
	)
}

func containsSubset[T any](t testing.TB, array, subset []T) {
	t.Helper()
	for _, elem := range subset {
		require.Contains(t, array, elem)
	}
}
