package v1alpha1_connectiondataservice

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	awsmanager "github.com/nucleuscloud/neosync/backend/internal/aws"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
)

type Service struct {
	cfg                *Config
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	connectionService  mgmtv1alpha1connect.ConnectionServiceClient
	jobService         mgmtv1alpha1connect.JobServiceHandler

	awsManager awsmanager.NeosyncAwsManagerClient

	sqlConnector sqlconnect.SqlConnector
	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier
	sqlmanager   sql_manager.SqlManagerClient

	mongoconnector mongoconnect.Interface
}

type Config struct {
}

func New(
	cfg *Config,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	jobService mgmtv1alpha1connect.JobServiceHandler,

	awsManager awsmanager.NeosyncAwsManagerClient,

	sqlConnector sqlconnect.SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
	mongoconnector mongoconnect.Interface,
	sqlmanager sql_manager.SqlManagerClient,
) *Service {
	return &Service{
		cfg:                cfg,
		useraccountService: useraccountService,
		connectionService:  connectionService,
		jobService:         jobService,
		awsManager:         awsManager,
		sqlConnector:       sqlConnector,
		pgquerier:          pgquerier,
		mysqlquerier:       mysqlquerier,
		sqlmanager:         sqlmanager,
		mongoconnector:     mongoconnector,
	}
}
