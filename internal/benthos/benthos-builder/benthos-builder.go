package benthosbuilder

import (
	"fmt"
	"log/slog"
	"sync"

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

// BenthosConfigResponse represents a complete Benthos data pipeline configuration for a specific table,
type BenthosConfigResponse struct {
	Name      string
	DependsOn []*tabledependency.DependsOn

	// TODO refactor these out
	Config                  *neosync_benthos.BenthosConfig
	TableSchema             string
	TableName               string
	Columns                 []string
	RunType                 tabledependency.RunType
	ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	RedisDependsOn          map[string][]string
	BenthosDsns             []*bb_shared.BenthosDsn
	RedisConfig             []*bb_shared.BenthosRedisConfig
}

// Combines a connection type and job type to uniquely identify a builder configuration
type BuilderKey struct {
	ConnType bb_internal.ConnectionType
	JobType  bb_internal.JobType
}

func (b *BuilderKey) String() string {
	return fmt.Sprintf("%s.%s", b.JobType, b.ConnType)
}

// Manages and provides access to different Benthos builders based on connection and job types
type BuilderProvider struct {
	mu       sync.RWMutex
	builders map[string]bb_internal.BenthosBuilder
	logger   *slog.Logger
}

// Creates a new BuilderProvider for managing builders
func NewBuilderProvider(logger *slog.Logger) *BuilderProvider {
	r := &BuilderProvider{
		builders: make(map[string]bb_internal.BenthosBuilder),
		logger:   logger,
	}
	return r
}

// Handles registering new builders
func (r *BuilderProvider) Register(jobType bb_internal.JobType, connType bb_internal.ConnectionType, builder bb_internal.BenthosBuilder) {
	key := BuilderKey{ConnType: connType, JobType: jobType}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.builders[key.String()]
	if !exists {
		r.logger.Debug(fmt.Sprintf("registering benthos builder for job type %s and connection type %s", jobType, connType))
		r.builders[key.String()] = builder
	}
}

// Handles getting builder based on job and connection type
func (r *BuilderProvider) GetBuilder(
	job *mgmtv1alpha1.Job,
	connection *mgmtv1alpha1.Connection,
) (bb_internal.BenthosBuilder, error) {
	connectionType := bb_internal.GetConnectionType(connection)
	jobType := bb_internal.GetJobType(job)
	key := BuilderKey{ConnType: connectionType, JobType: jobType}

	r.mu.RLock()
	builder, exists := r.builders[key.String()]
	r.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("builder not registered for connection type (%s) and job type (%s)", connectionType, jobType)
	}
	return builder, nil
}

// Handles registering what is considered standard builders
func (b *BuilderProvider) registerStandardBuilders(
	job *mgmtv1alpha1.Job,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnections []*mgmtv1alpha1.Connection,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	connectionclient mgmtv1alpha1connect.ConnectionServiceClient,
	redisConfig *shared.RedisConfig,
	postgresDriverOverride *string,
	selectQueryBuilder bb_shared.SelectQueryMapBuilder,
	rawSqlInsertMode bool,
) error {
	sourceConnectionType := bb_internal.GetConnectionType(sourceConnection)
	jobType := bb_internal.GetJobType(job)
	connectionTypes := []bb_internal.ConnectionType{sourceConnectionType}
	for _, dest := range destinationConnections {
		connectionTypes = append(connectionTypes, bb_internal.GetConnectionType(dest))
	}

	sqlSyncOptions := []bb_conns.SqlSyncOption{}
	if rawSqlInsertMode {
		sqlSyncOptions = append(sqlSyncOptions, bb_conns.WithRawInsertMode())
	}

	if jobType == bb_internal.JobTypeSync {
		for _, connectionType := range connectionTypes {
			switch connectionType {
			case bb_internal.ConnectionTypePostgres:
				driver := sqlmanager_shared.PostgresDriver
				if postgresDriverOverride != nil && *postgresDriverOverride != "" {
					driver = *postgresDriverOverride
				}
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, driver, selectQueryBuilder, sqlSyncOptions...)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_internal.ConnectionTypeMysql:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.MysqlDriver, selectQueryBuilder, sqlSyncOptions...)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_internal.ConnectionTypeMssql:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.MssqlDriver, selectQueryBuilder, sqlSyncOptions...)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_internal.ConnectionTypeAwsS3:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeAwsS3, bb_conns.NewAwsS3SyncBuilder())
			case bb_internal.ConnectionTypeDynamodb:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeDynamodb, bb_conns.NewDynamoDbSyncBuilder(transformerclient))
			case bb_internal.ConnectionTypeMongo:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeMongo, bb_conns.NewMongoDbSyncBuilder(transformerclient))
			case bb_internal.ConnectionTypeGCP:
				b.Register(bb_internal.JobTypeSync, bb_internal.ConnectionTypeGCP, bb_conns.NewGcpCloudStorageSyncBuilder())
			default:
				return fmt.Errorf("unsupport connection type for sync job: %s", connectionType)
			}
		}
	}

	if jobType == bb_internal.JobTypeAIGenerate {
		if len(destinationConnections) != 1 {
			return fmt.Errorf("unsupported destination count for AI generate job: %d", len(destinationConnections))
		}
		destConnType := bb_internal.GetConnectionType(destinationConnections[0])
		driver, err := bb_internal.GetSqlDriverByConnectionType(destConnType)
		if err != nil {
			return err
		}
		builder := bb_conns.NewGenerateAIBuilder(transformerclient, sqlmanagerclient, connectionclient, driver)
		b.Register(bb_internal.JobTypeAIGenerate, bb_internal.ConnectionTypeOpenAI, builder)
		b.Register(bb_internal.JobTypeAIGenerate, destConnType, builder)
	}
	if jobType == bb_internal.JobTypeGenerate {
		for _, connectionType := range connectionTypes {
			b.Register(bb_internal.JobTypeGenerate, connectionType, bb_conns.NewGenerateBuilder(transformerclient, sqlmanagerclient, connectionclient))
		}
	}
	return nil
}

// Adds builder logger tags
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

// Manages the creation and management of Benthos configurations
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

// Manages all necessary configuration parameters for creating
// a worker-based Benthos configuration manager
type WorkerBenthosConfig struct {
	Job                    *mgmtv1alpha1.Job
	SourceConnection       *mgmtv1alpha1.Connection
	DestinationConnections []*mgmtv1alpha1.Connection
	RunId                  string
	MetricLabelKeyVals     map[string]string
	Logger                 *slog.Logger
	Sqlmanagerclient       sqlmanager.SqlManagerClient
	Transformerclient      mgmtv1alpha1connect.TransformersServiceClient
	Connectionclient       mgmtv1alpha1connect.ConnectionServiceClient
	RedisConfig            *shared.RedisConfig
	MetricsEnabled         bool
	SelectQueryBuilder     bb_shared.SelectQueryMapBuilder
}

// Creates a new BenthosConfigManager configured for worker
func NewWorkerBenthosConfigManager(
	config *WorkerBenthosConfig,
) (*BenthosConfigManager, error) {
	rawInsertMode := false
	provider := NewBuilderProvider(config.Logger)
	err := provider.registerStandardBuilders(
		config.Job,
		config.SourceConnection,
		config.DestinationConnections,
		config.Sqlmanagerclient,
		config.Transformerclient,
		config.Connectionclient,
		config.RedisConfig,
		nil,
		config.SelectQueryBuilder,
		rawInsertMode,
	)
	if err != nil {
		return nil, err
	}
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
	}, nil
}

// Manages all necessary configuration parameters for creating
// a CLI-based Benthos configuration manager
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

// Creates a new BenthosConfigManager configured for CLI
func NewCliBenthosConfigManager(
	config *CliBenthosConfig,
) (*BenthosConfigManager, error) {
	rawInsertMode := true
	destinationProvider := NewBuilderProvider(config.Logger)
	err := destinationProvider.registerStandardBuilders(
		config.Job,
		config.SourceConnection,
		[]*mgmtv1alpha1.Connection{config.DestinationConnection},
		config.Sqlmanagerclient,
		config.Transformerclient,
		nil,
		config.RedisConfig,
		config.PostgresDriverOverride,
		nil,
		rawInsertMode,
	)
	if err != nil {
		return nil, err
	}

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
	}, nil
}

// NewCliSourceBuilderProvider creates a specialized provider for CLI source operations
func NewCliSourceBuilderProvider(
	config *CliBenthosConfig,
) *BuilderProvider {
	provider := NewBuilderProvider(config.Logger)

	sourceConnectionType := bb_internal.GetConnectionType(config.SourceConnection)
	jobType := bb_internal.GetJobType(config.Job)

	builder := bb_conns.NewNeosyncConnectionDataSyncBuilder(
		config.Connectiondataclient,
		config.Sqlmanagerclient,
		config.SourceJobRunId,
		config.SyncConfigs,
		config.DestinationConnection,
		sourceConnectionType,
	)

	if jobType == bb_internal.JobTypeSync {
		switch sourceConnectionType {
		case bb_internal.ConnectionTypePostgres, bb_internal.ConnectionTypeMysql,
			bb_internal.ConnectionTypeMssql, bb_internal.ConnectionTypeAwsS3, bb_internal.ConnectionTypeDynamodb:
			provider.Register(bb_internal.JobTypeSync, sourceConnectionType, builder)
		}
	}

	return provider
}
