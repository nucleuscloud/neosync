package genbenthosconfigs_activity

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
)

type SqlAdapter struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool    map[string]mysql_queries.DBTX
	mysqlquerier mysql_queries.Querier

	sqlconnector sqlconnect.SqlConnector
}

func NewSqlAdapter(
	pgpool map[string]pg_queries.DBTX,
	pgquerier pg_queries.Querier,

	mysqlpool map[string]mysql_queries.DBTX,
	mysqlquerier mysql_queries.Querier,

	sqlconnector sqlconnect.SqlConnector,
) *SqlAdapter {
	return &SqlAdapter{
		pgpool:       pgpool,
		pgquerier:    pgquerier,
		mysqlpool:    mysqlpool,
		mysqlquerier: mysqlquerier,
		sqlconnector: sqlconnector,
	}
}
