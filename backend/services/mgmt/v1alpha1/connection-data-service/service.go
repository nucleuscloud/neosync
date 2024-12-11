package v1alpha1_connectiondataservice

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
)

type Service struct {
	cfg               *Config
	connectionService mgmtv1alpha1connect.ConnectionServiceClient
	jobService        mgmtv1alpha1connect.JobServiceHandler

	awsManager awsmanager.NeosyncAwsManagerClient

	sqlConnector sqlconnect.SqlConnector
	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier
	sqlmanager   sql_manager.SqlManagerClient

	mongoconnector mongoconnect.Interface
	gcpmanager     neosync_gcp.ManagerInterface
}

type Config struct {
}

func New(
	cfg *Config,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	jobService mgmtv1alpha1connect.JobServiceHandler,

	awsManager awsmanager.NeosyncAwsManagerClient,

	sqlConnector sqlconnect.SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
	mongoconnector mongoconnect.Interface,
	sqlmanager sql_manager.SqlManagerClient,
	gcpmanager neosync_gcp.ManagerInterface,
) *Service {
	return &Service{
		cfg:               cfg,
		connectionService: connectionService,
		jobService:        jobService,
		awsManager:        awsManager,
		sqlConnector:      sqlConnector,
		pgquerier:         pgquerier,
		mysqlquerier:      mysqlquerier,
		sqlmanager:        sqlmanager,
		mongoconnector:    mongoconnector,
		gcpmanager:        gcpmanager,
	}
}
