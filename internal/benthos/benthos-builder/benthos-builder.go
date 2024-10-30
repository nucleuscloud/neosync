package benthosbuilder

import (
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_conns "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/builders"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

/* this should really be

type BenthosConfigResponse struct {

	Name: "schema.table_b.insert"
	DependsOn: []string{"schema.table_a.insert"}

}
*/

type BenthosConfigResponse struct {
	Name      string
	DependsOn []*tabledependency.DependsOn
	// RunType                 tabledependency.RunType
	Config         *neosync_benthos.BenthosConfig
	TableSchema    string
	TableName      string
	Columns        []string
	RedisDependsOn map[string][]string
	// ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	// SourceConnectionType    string // used for logging

	// Processors  []*neosync_benthos.ProcessorConfig
	BenthosDsns []*bb_shared.BenthosDsn
	RedisConfig []*bb_shared.BenthosRedisConfig

	// primaryKeys []string

	// metriclabels metrics.MetricLabels
}

// Creates a ConnectionBenthosBuilder
type BenthosBuilders func(
	jobType bb_internal.JobType,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) (bb_internal.ConnectionBenthosBuilder, error)

type BuilderKey struct {
	ConnType bb_internal.ConnectionType
	JobType  bb_internal.JobType
}

func (b *BuilderKey) String() string {
	return fmt.Sprintf("%s.%s", b.JobType, b.ConnType)
}

// BuilderRegistry maintains a mapping of connection types to benthos builders
type BuilderProvider struct {
	builders map[string]bb_internal.ConnectionBenthosBuilder
	logger   *slog.Logger
}

// Creates a new BuilderRegistry with default builders registered
func NewBuilderProvider(logger *slog.Logger) *BuilderProvider {
	r := &BuilderProvider{
		builders: make(map[string]bb_internal.ConnectionBenthosBuilder),
		logger:   logger,
	}
	return r
}

func (r *BuilderProvider) Register(jobType bb_internal.JobType, connType bb_internal.ConnectionType, builder bb_internal.ConnectionBenthosBuilder) {
	key := BuilderKey{ConnType: connType, JobType: jobType}
	_, exists := r.builders[key.String()]
	if !exists {
		r.logger.Debug(fmt.Sprintf("registering benthos builder for job type %s and connection type %s", jobType, connType))
		r.builders[key.String()] = builder
	}
}

// Creates a new builder for the given connection and job type
func (r *BuilderProvider) GetBuilder(
	job *mgmtv1alpha1.Job,
	connection *mgmtv1alpha1.Connection,
) (bb_internal.ConnectionBenthosBuilder, error) {
	connectionType := bb_internal.GetConnectionType(connection)
	jobType := bb_internal.GetJobType(job)
	key := BuilderKey{ConnType: connectionType, JobType: jobType}
	builder, exists := r.builders[key.String()]
	if !exists {
		return nil, fmt.Errorf("unsupported connection type: %s", connectionType)
	}
	return builder, nil
}

func (b *BuilderProvider) registerStandardBuilders(
	job *mgmtv1alpha1.Job,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnections []*mgmtv1alpha1.Connection,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
	postgresDriverOverride *string,
) {
	sourceConnectionType := bb_internal.GetConnectionType(sourceConnection)
	jobType := bb_internal.GetJobType(job)
	connectionTypes := []bb_internal.ConnectionType{sourceConnectionType}
	for _, dest := range destinationConnections {
		connectionTypes = append(connectionTypes, bb_internal.GetConnectionType(dest))
	}

	if jobType == bb_internal.JobTypeSync {
		for _, connectionType := range connectionTypes {
			switch connectionType {
			case bb_internal.ConnectionTypePostgres:
				driver := sqlmanager_shared.PostgresDriver
				if postgresDriverOverride != nil && *postgresDriverOverride != "" {
					driver = *postgresDriverOverride
				}
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, driver)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_internal.ConnectionTypeMysql:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.MysqlDriver)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_internal.ConnectionTypeMssql:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.MssqlDriver)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_internal.ConnectionTypeAwsS3:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeAwsS3, bb_conns.NewAwsS3SyncBuilder())
			case bb_internal.ConnectionTypeDynamodb:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeDynamodb, bb_conns.NewDynamoDbSyncBuilder(transformerclient))
			case bb_internal.ConnectionTypeMongo:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeMongo, bb_conns.NewMongoDbSyncBuilder(transformerclient))
			case bb_internal.ConnectionTypeGCP:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeGCP, bb_conns.NewGcpCloudStorageSyncBuilder())
			}
		}
	}
}

func withBenthosConfigLoggerTags(
	job *mgmtv1alpha1.Job,
	sourceConnection *mgmtv1alpha1.Connection,
) []any {
	keyvals := []any{}

	sourceConnectionType := bb_internal.GetConnectionType(sourceConnection)
	jobType := bb_internal.GetJobType(job)

	if sourceConnectionType != "" {
		keyvals = append(keyvals, "sourceConnectionType", sourceConnectionType)
	}
	if jobType != "" {
		keyvals = append(keyvals, "jobType", jobType)
	}

	return keyvals
}

type BenthosConfigManager struct {
	sourceProvider         *BuilderProvider
	destinationProvider    *BuilderProvider
	metricsEnabled         bool
	metricLabelKeyVals     map[string]string
	logger                 *slog.Logger
	job                    *mgmtv1alpha1.Job
	sourceConnection       *mgmtv1alpha1.Connection
	destinationConnections []*mgmtv1alpha1.Connection
	runId                  string
}

type WorkerBenthosConfig struct {
	Job                    *mgmtv1alpha1.Job
	SourceConnection       *mgmtv1alpha1.Connection
	DestinationConnections []*mgmtv1alpha1.Connection
	RunId                  string
	MetricLabelKeyVals     map[string]string
	Logger                 *slog.Logger
	Sqlmanagerclient       sqlmanager.SqlManagerClient
	Transformerclient      mgmtv1alpha1connect.TransformersServiceClient
	RedisConfig            *shared.RedisConfig
	MetricsEnabled         bool
}

func NewWorkerBenthosConfigManager(
	config *WorkerBenthosConfig,
) *BenthosConfigManager {
	provider := NewBuilderProvider(config.Logger)
	provider.registerStandardBuilders(
		config.Job,
		config.SourceConnection,
		config.DestinationConnections,
		config.Sqlmanagerclient,
		config.Transformerclient,
		config.RedisConfig,
		nil,
	)
	logger := config.Logger.With(withBenthosConfigLoggerTags(config.Job, config.SourceConnection)...)
	return &BenthosConfigManager{
		sourceProvider:         provider,
		destinationProvider:    provider,
		metricsEnabled:         config.MetricsEnabled,
		metricLabelKeyVals:     config.MetricLabelKeyVals,
		logger:                 logger,
		job:                    config.Job,
		sourceConnection:       config.SourceConnection,
		destinationConnections: config.DestinationConnections,
		runId:                  config.RunId,
	}
}

type CliBenthosConfig struct {
	Job                    *mgmtv1alpha1.Job
	SourceConnection       *mgmtv1alpha1.Connection
	DestinationConnection  *mgmtv1alpha1.Connection
	SourceJobRunId         *string // for use when AWS S3 is the source
	PostgresDriverOverride *string // optional driver override. used for postgres
	SyncConfigs            []*tabledependency.RunConfig
	RunId                  string
	MetricLabelKeyVals     map[string]string
	Logger                 *slog.Logger
	Sqlmanagerclient       sqlmanager.SqlManagerClient
	Transformerclient      mgmtv1alpha1connect.TransformersServiceClient
	Connectiondataclient   mgmtv1alpha1connect.ConnectionDataServiceClient
	RedisConfig            *shared.RedisConfig
	MetricsEnabled         bool
}

func NewCliBenthosConfigManager(
	config *CliBenthosConfig,
) *BenthosConfigManager {
	destinationProvider := NewBuilderProvider(config.Logger)
	destinationProvider.registerStandardBuilders(
		config.Job,
		config.SourceConnection,
		[]*mgmtv1alpha1.Connection{config.DestinationConnection},
		config.Sqlmanagerclient,
		config.Transformerclient,
		config.RedisConfig,
		config.PostgresDriverOverride,
	)

	sourceProvider := NewCliSourceBuilderProvider(config)

	logger := config.Logger.With(withBenthosConfigLoggerTags(config.Job, config.SourceConnection)...)
	return &BenthosConfigManager{
		sourceProvider:         sourceProvider,
		destinationProvider:    destinationProvider,
		metricsEnabled:         config.MetricsEnabled,
		logger:                 logger,
		job:                    config.Job,
		sourceConnection:       config.SourceConnection,
		destinationConnections: []*mgmtv1alpha1.Connection{config.DestinationConnection},
		runId:                  config.RunId,
	}
}

// NewCliSourceBuilderProvider creates a specialized provider for CLI source operations
func NewCliSourceBuilderProvider(
	config *CliBenthosConfig,
) *BuilderProvider {
	provider := NewBuilderProvider(config.Logger)

	builder := bb_conns.NewNeosyncConnectionDataSyncBuilder(
		config.Connectiondataclient,
		config.Sqlmanagerclient,
		config.SourceJobRunId,
		config.SyncConfigs,
		config.SourceConnection,
	)

	sourceConnectionType := bb_internal.GetConnectionType(config.SourceConnection)
	jobType := bb_internal.GetJobType(config.Job)

	if jobType == bb_internal.JobTypeSync {
		switch sourceConnectionType {
		case bb_internal.ConnectionTypePostgres, bb_internal.ConnectionTypeMysql,
			bb_internal.ConnectionTypeMssql, bb_internal.ConnectionTypeAwsS3, bb_internal.ConnectionTypeDynamodb:
			provider.Register(bb_internal.JobTypeSync, sourceConnectionType, builder)
		}
	}

	return provider
}
