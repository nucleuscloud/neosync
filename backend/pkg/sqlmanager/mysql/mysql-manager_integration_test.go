package sqlmanager_mysql

import (
	"context"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetDatabaseSchema() {
	manager := MysqlManager{querier: s.querier, pool: s.pool}

	expected := &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{
			"sqlmanagermysql.container": {
				{Columns: []string{"container_status_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermysql.container_status",
					Columns: []string{"id"},
				}},
			},
			"sqlmanagermysql2.container": {
				{Columns: []string{"container_status_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "sqlmanagermysql2.container_status",
					Columns: []string{"id"},
				}},
			},
		},
		PrimaryKeyConstraints: map[string][]string{
			"sqlmanagermysql.container":         {"id"},
			"sqlmanagermysql.container_status":  {"id"},
			"sqlmanagermysql2.container":        {"id"},
			"sqlmanagermysql2.container_status": {"id"},
		},
	}

	actual, err := manager.GetTableConstraintsBySchema(context.Background(), []string{"sqlmanagermysql", "sqlmanagermysql2"})
	require.NoError(s.T(), err)
	require.Equal(s.T(), expected.ForeignKeyConstraints, actual.ForeignKeyConstraints)
	require.Equal(s.T(), expected.PrimaryKeyConstraints, actual.PrimaryKeyConstraints)
}
