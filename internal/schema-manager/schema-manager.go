package schemamanager

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	schema_mssql "github.com/nucleuscloud/neosync/internal/schema-manager/mssql"
	schema_mysql "github.com/nucleuscloud/neosync/internal/schema-manager/mysql"
	schema_notsupported "github.com/nucleuscloud/neosync/internal/schema-manager/not-supported"
	schema_postgres "github.com/nucleuscloud/neosync/internal/schema-manager/postgres"
	schema_shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
)

type SchemaManagerService interface {
	InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*schema_shared.InitSchemaError, error)
	TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error
	CloseConnections()
}

type SchemaManager interface {
	New(
		ctx context.Context,
		sourceConnection *mgmtv1alpha1.Connection,
		destinationConnection *mgmtv1alpha1.Connection,
		destination *mgmtv1alpha1.JobDestination,
	) (SchemaManagerService, error)
}

type DefaultSchemaManager struct {
	sqlmanagerclient sqlmanager.SqlManagerClient
	session          connectionmanager.SessionInterface
	logger           *slog.Logger
	eelicense        license.EEInterface
}

func NewSchemaManager(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	session connectionmanager.SessionInterface,
	logger *slog.Logger,
	eelicense license.EEInterface,
) SchemaManager {
	return &DefaultSchemaManager{sqlmanagerclient: sqlmanagerclient, session: session, logger: logger, eelicense: eelicense}
}

func (d *DefaultSchemaManager) New(
	ctx context.Context,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnection *mgmtv1alpha1.Connection,
	destination *mgmtv1alpha1.JobDestination,
) (SchemaManagerService, error) {
	switch cfg := destination.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		opts := cfg.PostgresOptions
		return schema_postgres.NewPostgresSchemaManager(ctx, d.logger, d.session, d.sqlmanagerclient, sourceConnection, destinationConnection, opts)
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		opts := cfg.MysqlOptions
		return schema_mysql.NewMysqlSchemaManager(ctx, d.logger, d.session, d.sqlmanagerclient, sourceConnection, destinationConnection, opts)
	case *mgmtv1alpha1.JobDestinationOptions_MssqlOptions:
		opts := cfg.MssqlOptions
		return schema_mssql.NewMssqlSchemaManager(ctx, d.logger, d.eelicense, d.session, d.sqlmanagerclient, sourceConnection, destinationConnection, opts)
	case *mgmtv1alpha1.JobDestinationOptions_DynamodbOptions, *mgmtv1alpha1.JobDestinationOptions_MongodbOptions, *mgmtv1alpha1.JobDestinationOptions_AwsS3Options, *mgmtv1alpha1.JobDestinationOptions_GcpCloudstorageOptions:
		// For destinations like DynamoDB, MongoDB, S3, and GCP Cloud Storage, we use a no-op implementation
		// since schema initialization and data truncation don't apply to these data stores
		return schema_notsupported.NewNotSupportedSchemaManager()
	default:
		return nil, fmt.Errorf("unsupported connection type: %T", destination.GetOptions().GetConfig())
	}
}
