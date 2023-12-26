package v1alpha1_connectionservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/internal/sqlconnect"
)

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlConnector       sqlconnect.SqlConnector
}

type Config struct {
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	sqlConnector sqlconnect.SqlConnector,
) *Service {
	return &Service{
		cfg:                cfg,
		db:                 db,
		useraccountService: useraccountService,
		sqlConnector:       sqlConnector,
	}
}
