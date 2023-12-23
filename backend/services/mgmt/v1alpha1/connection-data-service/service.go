package v1alpha1_connectiondataservice

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

type SqlConnector interface {
	MysqlOpen(dataSourceName string) (*sql.DB, error)
	PgPoolOpen(ctx context.Context, dataSourceName string) (*pgxpool.Pool, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) MysqlOpen(dataSourceName string) (*sql.DB, error) {
	return sql.Open("mysql", dataSourceName)
}

func (rc *SqlOpenConnector) PgPoolOpen(ctx context.Context, dataSourceName string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, dataSourceName)
}

type Service struct {
	cfg                *Config
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	connectionService  mgmtv1alpha1connect.ConnectionServiceClient
	jobService         mgmtv1alpha1connect.JobServiceClient
	sqlConnector       SqlConnector

	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier
}

type Config struct {
}

func New(
	cfg *Config,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	jobService mgmtv1alpha1connect.JobServiceClient,

	sqlConnector SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
) *Service {
	if sqlConnector != nil {
		return &Service{
			cfg:                cfg,
			useraccountService: useraccountService,
			connectionService:  connectionService,
			jobService:         jobService,
			sqlConnector:       sqlConnector,
			pgquerier:          pgquerier,
			mysqlquerier:       mysqlquerier,
		}
	}
	return &Service{
		cfg:                cfg,
		useraccountService: useraccountService,
		connectionService:  connectionService,
		jobService:         jobService,
		sqlConnector:       &SqlOpenConnector{},
		pgquerier:          pgquerier,
		mysqlquerier:       mysqlquerier,
	}
}
