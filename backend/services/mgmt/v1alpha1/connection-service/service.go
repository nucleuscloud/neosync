package v1alpha1_connectionservice

import (
	"sync"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
)

type Service struct {
	cfg                *Config
	db                 *neosyncdb.NeosyncDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlConnector       sqlconnect.SqlConnector
	sqlmanager         sql_manager.SqlManagerClient

	mongoconnector mongoconnect.Interface

	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier

	awsManager awsmanager.NeosyncAwsManagerClient
}

type Config struct {
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	sqlConnector sqlconnect.SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
	mssqlquerier mssql_queries.Querier,
	mongoconnector mongoconnect.Interface,
	awsManager awsmanager.NeosyncAwsManagerClient,
) *Service {
	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}
	mssqlpoolmap := &sync.Map{}
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlConnector)
	return &Service{
		cfg:                cfg,
		db:                 db,
		useraccountService: useraccountService,
		sqlConnector:       sqlConnector,
		pgquerier:          pgquerier,
		mysqlquerier:       mysqlquerier,
		sqlmanager:         sqlmanager,
		mongoconnector:     mongoconnector,
		awsManager:         awsManager,
	}
}
