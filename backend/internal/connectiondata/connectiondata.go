package connectiondata

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosync_gcp "github.com/nucleuscloud/neosync/backend/internal/gcp"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	aws_manager "github.com/nucleuscloud/neosync/internal/aws"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
)

type ConnectionDataService interface {
	StreamData(
		ctx context.Context,
		stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
		StreamConfig *mgmtv1alpha1.ConnectionStreamConfig,
		schema, table string,
	) error
	GetSchema(ctx context.Context, config *mgmtv1alpha1.ConnectionSchemaConfig) ([]*mgmtv1alpha1.DatabaseColumn, error)
}

type ConnectionDataBuilder interface {
	NewDataConnection(logger *slog.Logger, connection *mgmtv1alpha1.Connection) (ConnectionDataService, error)
}

type DefaultConnectionDataBuilder struct {
	sqlconnector        sqlconnect.SqlConnector
	sqlmanager          sql_manager.SqlManagerClient
	pgquerier           pg_queries.Querier
	mysqlquerier        mysql_queries.Querier
	awsmanager          aws_manager.NeosyncAwsManagerClient
	gcpmanager          neosync_gcp.ManagerInterface
	mongoconnector      mongoconnect.Interface
	neosynctyperegistry neosynctypes.NeosyncTypeRegistry
	jobservice          mgmtv1alpha1connect.JobServiceHandler // TODO: remove this dependency
}

func NewConnectionDataBuilder(
	sqlconnector sqlconnect.SqlConnector,
	sqlmanager sql_manager.SqlManagerClient,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
	awsmanager aws_manager.NeosyncAwsManagerClient,
	gcpmanager neosync_gcp.ManagerInterface,
	mongoconnector mongoconnect.Interface,
	neosynctyperegistry neosynctypes.NeosyncTypeRegistry,
	jobservice mgmtv1alpha1connect.JobServiceHandler,
) ConnectionDataBuilder {
	return &DefaultConnectionDataBuilder{
		sqlconnector:        sqlconnector,
		sqlmanager:          sqlmanager,
		pgquerier:           pgquerier,
		mysqlquerier:        mysqlquerier,
		awsmanager:          awsmanager,
		gcpmanager:          gcpmanager,
		mongoconnector:      mongoconnector,
		neosynctyperegistry: neosynctyperegistry,
		jobservice:          jobservice,
	}
}

func (b *DefaultConnectionDataBuilder) NewDataConnection(
	logger *slog.Logger,
	connection *mgmtv1alpha1.Connection,
) (ConnectionDataService, error) {
	switch config := connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig,
		*mgmtv1alpha1.ConnectionConfig_PgConfig:
		return NewSQLConnectionDataService(logger, b.sqlconnector, b.sqlmanager, connection), nil
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return NewAwsS3ConnectionDataService(logger, b.awsmanager, b.neosynctyperegistry, connection), nil
	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		return NewGcpConnectionDataService(logger, b.gcpmanager, connection, b.jobservice), nil
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return NewAwsDynamodbConnectionDataService(logger, b.awsmanager, connection), nil
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return NewMongoDbConnectionDataService(logger, connection, b.mongoconnector), nil
	default:
		return nil, fmt.Errorf("connection config not supported for connection data service: %T", config)
	}
}

type SchemaOpts struct {
	JobId    *string
	JobRunId *string
}
