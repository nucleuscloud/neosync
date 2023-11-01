package v1alpha1_connectionservice

import (
	"database/sql"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
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
	db                 *nucleusdb.NucleusDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlConnector       sqlConnector
}

type Config struct {
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	sqlConnector sqlConnector,
) *Service {
	if sqlConnector != nil {
		return &Service{
			cfg:                cfg,
			db:                 db,
			useraccountService: useraccountService,
			sqlConnector:       sqlConnector,
		}
	}

	return &Service{
		cfg:                cfg,
		db:                 db,
		useraccountService: useraccountService,
		sqlConnector:       &sqlOpenConnector{},
	}
}
