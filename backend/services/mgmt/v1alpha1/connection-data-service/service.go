package v1alpha1_connectiondataservice

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	awsmanager "github.com/nucleuscloud/neosync/backend/internal/aws"
	"github.com/nucleuscloud/neosync/backend/internal/sqlconnect"
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
	}
}
