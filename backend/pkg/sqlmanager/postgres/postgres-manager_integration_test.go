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

	constraints, ok := actual.ForeignKeyConstraints[s.buildTable("child1")]
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

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable("t1")]
	require.True(s.T(), ok, "t1 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t1 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t1"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable("t2")]
	require.True(s.T(), ok, "t2 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t2 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t3"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable("t3")]
	require.True(s.T(), ok, "t3 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t3 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t2"), Columns: []string{"a"}},
		},
	})
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap_BasicCircular() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{s.schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	constraints, ok := actual.ForeignKeyConstraints[s.buildTable("t1")]
	require.True(s.T(), ok, "t1 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t1 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t1"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable("t2")]
	require.True(s.T(), ok, "t2 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t2 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t3"), Columns: []string{"a"}},
		},
	})

	constraints, ok = actual.ForeignKeyConstraints[s.buildTable("t3")]
	require.True(s.T(), ok, "t3 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t3 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"b"},
			NotNullable: []bool{false},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t2"), Columns: []string{"a"}},
		},
	})
}

func (s *IntegrationTestSuite) Test_GetForeignKeyConstraintsMap_Composite() {
	manager := PostgresManager{querier: s.querier, pool: s.pgpool}

	actual, err := manager.GetTableConstraintsBySchema(s.ctx, []string{s.schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)

	constraints, ok := actual.ForeignKeyConstraints[s.buildTable("t5")]
	require.True(s.T(), ok, "t5 should be in map")
	require.NotEmpty(s.T(), constraints, "should contain t5 constraints")
	containsSubset(s.T(), constraints, []*sqlmanager_shared.ForeignConstraint{
		{
			Columns:     []string{"x", "y"},
			NotNullable: []bool{true, true},
			ForeignKey:  &sqlmanager_shared.ForeignKey{Table: s.buildTable("t4"), Columns: []string{"a", "b"}},
		},
	})
}

func (s *IntegrationTestSuite) Test_GetPrimaryKeyConstraintsMap() {
	manager := NewManager(s.querier, s.pgpool, func() {})

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{s.schema})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), actual)
	require.NotEmpty(s.T(), actual.PrimaryKeyConstraints)

	pkeys, ok := actual.PrimaryKeyConstraints[s.buildTable("users_with_identity")]
	require.True(s.T(), ok, "users_with_identity had no entries")
	require.ElementsMatch(s.T(), []string{"id"}, pkeys)

	pkeys, ok = actual.PrimaryKeyConstraints[s.buildTable("t4")]
	require.True(s.T(), ok, "t4 had no entries")
	require.ElementsMatch(s.T(), []string{"a", "b"}, pkeys)
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
