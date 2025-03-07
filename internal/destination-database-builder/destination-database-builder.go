package destinationdatabasebuilder

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	ddbuilder_mssql "github.com/nucleuscloud/neosync/internal/destination-database-builder/mssql"
	ddbuilder_mysql "github.com/nucleuscloud/neosync/internal/destination-database-builder/mysql"
	ddbuilder_notsupported "github.com/nucleuscloud/neosync/internal/destination-database-builder/not-supported"
	ddbuilder_postgres "github.com/nucleuscloud/neosync/internal/destination-database-builder/postgres"
	destdb_shared "github.com/nucleuscloud/neosync/internal/destination-database-builder/shared"
	"github.com/nucleuscloud/neosync/internal/ee/license"
)

type DestinationDatabaseBuilderService interface {
	InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*destdb_shared.InitSchemaError, error)
	TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error
	CloseConnections()
}

type DestinationDatabaseBuilder interface {
	NewDestinationDatabaseBuilderService(
		ctx context.Context,
		sourceConnection *mgmtv1alpha1.Connection,
		destinationConnection *mgmtv1alpha1.Connection,
		destination *mgmtv1alpha1.JobDestination,
	) (DestinationDatabaseBuilderService, error)
}

type DefaultDestinationDatabaseBuilder struct {
	sqlmanagerclient sqlmanager.SqlManagerClient
	session          connectionmanager.SessionInterface
	logger           *slog.Logger
	eelicense        license.EEInterface
}

func NewDestinationDatabaseBuilder(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	session connectionmanager.SessionInterface,
	logger *slog.Logger,
	eelicense license.EEInterface,
) DestinationDatabaseBuilder {
	return &DefaultDestinationDatabaseBuilder{sqlmanagerclient: sqlmanagerclient, session: session, logger: logger, eelicense: eelicense}
}

func (d *DefaultDestinationDatabaseBuilder) NewDestinationDatabaseBuilderService(
	ctx context.Context,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnection *mgmtv1alpha1.Connection,
	destination *mgmtv1alpha1.JobDestination,
) (DestinationDatabaseBuilderService, error) {
	switch cfg := destination.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		opts := cfg.PostgresOptions
		return ddbuilder_postgres.NewPostgresDestinationDatabaseBuilderService(ctx, d.logger, d.session, d.sqlmanagerclient, sourceConnection, destinationConnection, opts)
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		opts := cfg.MysqlOptions
		return ddbuilder_mysql.NewMysqlDestinationDatabaseBuilderService(ctx, d.logger, d.session, d.sqlmanagerclient, sourceConnection, destinationConnection, opts)
	case *mgmtv1alpha1.JobDestinationOptions_MssqlOptions:
		opts := cfg.MssqlOptions
		return ddbuilder_mssql.NewMssqlDestinationDatabaseBuilderService(ctx, d.logger, d.eelicense, d.session, d.sqlmanagerclient, sourceConnection, destinationConnection, opts)
	case *mgmtv1alpha1.JobDestinationOptions_DynamodbOptions, *mgmtv1alpha1.JobDestinationOptions_MongodbOptions, *mgmtv1alpha1.JobDestinationOptions_AwsS3Options, *mgmtv1alpha1.JobDestinationOptions_GcpCloudstorageOptions:
		// For destinations like DynamoDB, MongoDB, S3, and GCP Cloud Storage, we use a no-op implementation
		// since schema initialization and data truncation don't apply to these data stores
		return ddbuilder_notsupported.NewNotSupportedDestinationDatabaseBuilderService()
	default:
		return nil, fmt.Errorf("unsupported connection type: %T", destination.GetOptions().GetConfig())
	}
}
