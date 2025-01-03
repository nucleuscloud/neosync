package v1alpha1_connectiondataservice

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/connectiondata"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
)

type Service struct {
	cfg               *Config
	connectionService mgmtv1alpha1connect.ConnectionServiceClient
	jobService        mgmtv1alpha1connect.JobServiceHandler

	sqlConnector sqlconnect.SqlConnector
	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier
	sqlmanager   sql_manager.SqlManagerClient

	connectiondatabuilder connectiondata.ConnectionDataBuilder
}

type Config struct {
}

func New(
	cfg *Config,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	jobService mgmtv1alpha1connect.JobServiceHandler,

	sqlConnector sqlconnect.SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
	sqlmanager sql_manager.SqlManagerClient,

	connectiondatabuilder connectiondata.ConnectionDataBuilder,
) *Service {
	return &Service{
		cfg:                   cfg,
		connectionService:     connectionService,
		jobService:            jobService,
		sqlConnector:          sqlConnector,
		pgquerier:             pgquerier,
		mysqlquerier:          mysqlquerier,
		sqlmanager:            sqlmanager,
		connectiondatabuilder: connectiondatabuilder,
	}
}
