package v1alpha1_connectionservice

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
)

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlConnector       sqlconnect.SqlConnector

	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier
}

type Config struct {
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	sqlConnector sqlconnect.SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
) *Service {
	return &Service{
		cfg:                cfg,
		db:                 db,
		useraccountService: useraccountService,
		sqlConnector:       sqlConnector,
		pgquerier:          pgquerier,
		mysqlquerier:       mysqlquerier,
	}
}
