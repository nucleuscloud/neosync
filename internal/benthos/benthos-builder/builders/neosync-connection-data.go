package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

type ConnectionType string

const (
	awsS3Connection           ConnectionType = "awsS3"
	gcpCloudStorageConnection ConnectionType = "gcpCloudStorage"
	postgresConnection        ConnectionType = "postgres"
	mysqlConnection           ConnectionType = "mysql"
	awsDynamoDBConnection     ConnectionType = "awsDynamoDB"
)

type neosyncConnectionDataBuilder struct {
	connectiondataclient  mgmtv1alpha1connect.ConnectionDataServiceClient
	sqlmanagerclient      sqlmanager.SqlManagerClient
	sourceJobRunId        *string
	syncConfigs           []*tabledependency.RunConfig
	destinationConnection *mgmtv1alpha1.Connection
}

func NewNeosyncConnectionDataSyncBuilder(
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceJobRunId *string,
	syncConfigs []*tabledependency.RunConfig,
	destinationConnection *mgmtv1alpha1.Connection,
) bb_internal.ConnectionBenthosBuilder {
	return &neosyncConnectionDataBuilder{
		connectiondataclient:  connectiondataclient,
		sqlmanagerclient:      sqlmanagerclient,
		sourceJobRunId:        sourceJobRunId,
		syncConfigs:           syncConfigs,
		destinationConnection: destinationConnection,
	}
}

func (b *neosyncConnectionDataBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job
	logger := params.Logger
	configs := []*bb_internal.BenthosSourceConfig{}
	connectionType, err := getConnectionType(sourceConnection)
	if err != nil {
		return nil, err
	}

	tableColumnDefaults, err := b.getSqlDestinationColumnDefaultsByTable(ctx, logger, job)
	if err != nil {
		return nil, err
	}

	for _, config := range b.syncConfigs {
		schema, table := sqlmanager_shared.SplitTableKey(config.Table())
		columnDefaultProperties := tableColumnDefaults[config.Table()]

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Logger: &neosync_benthos.LoggerConfig{
					Level:        "ERROR",
					AddTimestamp: true,
				},
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						NeosyncConnectionData: &neosync_benthos.NeosyncConnectionData{
							ConnectionId:   sourceConnection.GetId(),
							ConnectionType: string(connectionType),
							JobId:          &job.Id,
							JobRunId:       b.sourceJobRunId,
							Schema:         schema,
							Table:          table,
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{},
				// Output:   &neosync_benthos.OutputConfig{},
				Output: &neosync_benthos.OutputConfig{
					Outputs: neosync_benthos.Outputs{
						Broker: &neosync_benthos.OutputBrokerConfig{
							Pattern: "fan_out",
							Outputs: []neosync_benthos.Outputs{},
						},
					},
				},
			},
		}
		configs = append(configs, &bb_internal.BenthosSourceConfig{
			Name:      fmt.Sprintf("%s.%s", config.Table(), config.RunType()),
			Config:    bc,
			DependsOn: config.DependsOn(),
			RunType:   config.RunType(),

			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: sourceConnection.Id, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema:             schema,
			TableName:               table,
			Columns:                 config.InsertColumns(),
			ColumnDefaultProperties: columnDefaultProperties,
			PrimaryKeys:             config.PrimaryKeys(),
		})

	}

	return configs, nil
}

func (b *neosyncConnectionDataBuilder) getSqlDestinationColumnDefaultsByTable(
	ctx context.Context,
	logger *slog.Logger,
	job *mgmtv1alpha1.Job,
) (map[string]map[string]*neosync_benthos.ColumnDefaultProperties, error) {
	tableColDefaults := map[string]map[string]*neosync_benthos.ColumnDefaultProperties{}
	switch b.destinationConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		groupedMappings := groupMappingsByTable(job.Mappings)
		groupedTableMapping := getTableMappingsMap(groupedMappings)
		colTransformerMap := getColumnTransformerMap(groupedTableMapping) // schema.table ->  column -> transformer

		db, err := b.sqlmanagerclient.NewPooledSqlDb(ctx, logger, b.destinationConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to create new sql db: %w", err)
		}
		defer db.Db.Close()

		groupedColumnInfo, err := db.Db.GetSchemaColumnMap(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
		}
		for _, config := range b.syncConfigs {
			tableColTransformers := colTransformerMap[config.Table()]
			colInfoMap := groupedColumnInfo[config.Table()]

			columnDefaultProperties, err := getColumnDefaultProperties(logger, db.Driver, config.InsertColumns(), colInfoMap, tableColTransformers)
			if err != nil {
				return nil, err
			}
			tableColDefaults[config.Table()] = columnDefaultProperties
		}
	}

	return tableColDefaults, nil
}

func (b *neosyncConnectionDataBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	return nil, errors.ErrUnsupported
}

func getConnectionType(connection *mgmtv1alpha1.Connection) (ConnectionType, error) {
	if connection.ConnectionConfig.GetAwsS3Config() != nil {
		return awsS3Connection, nil
	}
	if connection.GetConnectionConfig().GetGcpCloudstorageConfig() != nil {
		return gcpCloudStorageConnection, nil
	}
	if connection.ConnectionConfig.GetMysqlConfig() != nil {
		return mysqlConnection, nil
	}
	if connection.ConnectionConfig.GetPgConfig() != nil {
		return postgresConnection, nil
	}
	if connection.ConnectionConfig.GetDynamodbConfig() != nil {
		return awsDynamoDBConnection, nil
	}
	return "", errors.New("unsupported connection type")
}
