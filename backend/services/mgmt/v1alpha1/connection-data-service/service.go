package v1alpha1_connectiondataservice

import (
	"database/sql"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

type sqlConnector interface {
	Open(driverName, dataSourceName string) (*sql.DB, error)
}

type sqlOpenConnector struct{}

func (rc *sqlOpenConnector) Open(driverName, dataSourceName string) (*sql.DB, error) {
	return sql.Open(driverName, dataSourceName)
}

type Service struct {
	cfg                *Config
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	connectionService  mgmtv1alpha1connect.ConnectionServiceClient
	jobService         mgmtv1alpha1connect.JobServiceClient
	sqlConnector       sqlConnector
}

type Config struct {
}

func New(
	cfg *Config,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	jobService mgmtv1alpha1connect.JobServiceClient,
	sqlConnector sqlConnector,
) *Service {
	if sqlConnector != nil {
		return &Service{
			cfg:                cfg,
			useraccountService: useraccountService,
			connectionService:  connectionService,
			jobService:         jobService,
			sqlConnector:       sqlConnector,
		}
	}
	return &Service{
		cfg:                cfg,
		useraccountService: useraccountService,
		connectionService:  connectionService,
		jobService:         jobService,
		sqlConnector:       &sqlOpenConnector{},
	}
}
