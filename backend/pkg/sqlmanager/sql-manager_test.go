package sqlmanager

import (
	"testing"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	"github.com/stretchr/testify/require"
)

func Test_NewSqlManager(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		mgr := NewSqlManager(connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{})))
		require.NotNil(t, mgr)
	})
	t.Run("new_with_opts", func(t *testing.T) {
		mgr := NewSqlManager(
			connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{})),
			WithPostgresQuerier(pg_queries.New()),
			WithMssqlQuerier(mssql_queries.New()),
			WithMysqlQuerier(mysql_queries.New()),
		)
		require.NotNil(t, mgr)
	})
}
