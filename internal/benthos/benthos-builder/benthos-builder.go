package benthosbuilder

import (
	"fmt"
	"log/slog"
	"sync"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_conns "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/builders"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

// BenthosConfigResponse represents a complete Benthos data pipeline configuration for a specific table,
type BenthosConfigResponse struct {
	Name      string
	DependsOn []*runconfigs.DependsOn

	// TODO refactor these out
	Config                  *neosync_benthos.BenthosConfig
	TableSchema             string
	TableName               string
	Columns                 []string
	RunType                 runconfigs.RunType
	ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	RedisDependsOn          map[string][]string
	BenthosDsns             []*bb_shared.BenthosDsn
	RedisConfig             []*bb_shared.BenthosRedisConfig
}

// Combines a connection type and job type to uniquely identify a builder configuration
type BuilderKey struct {
	ConnType bb_shared.ConnectionType
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
func (r *BuilderProvider) Register(jobType bb_internal.JobType, connType bb_shared.ConnectionType, builder bb_internal.BenthosBuilder) {
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
	connectionType, err := bb_shared.GetConnectionType(connection)
	if err != nil {
		return nil, err
	}
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
	redisConfig *neosync_redis.RedisConfig,
	selectQueryBuilder bb_shared.SelectQueryMapBuilder,
) error {
	sourceConnectionType, err := bb_shared.GetConnectionType(sourceConnection)
	if err != nil {
		return err
	}
	jobType := bb_internal.GetJobType(job)
	connectionTypes := []bb_shared.ConnectionType{sourceConnectionType}
	for _, dest := range destinationConnections {
		destConnType, err := bb_shared.GetConnectionType(dest)
		if err != nil {
			return err
		}
		connectionTypes = append(connectionTypes, destConnType)
	}

	if jobType == bb_internal.JobTypeSync {
		for _, connectionType := range connectionTypes {
			switch connectionType {
			case bb_shared.ConnectionTypePostgres:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.PostgresDriver, selectQueryBuilder)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_shared.ConnectionTypeMysql:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.MysqlDriver, selectQueryBuilder)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_shared.ConnectionTypeMssql:
				sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig, sqlmanager_shared.MssqlDriver, selectQueryBuilder)
				b.Register(bb_internal.JobTypeSync, connectionType, sqlbuilder)
			case bb_shared.ConnectionTypeAwsS3:
				b.Register(bb_internal.JobTypeSync, bb_shared.ConnectionTypeAwsS3, bb_conns.NewAwsS3SyncBuilder())
			case bb_shared.ConnectionTypeDynamodb:
				b.Register(bb_internal.JobTypeSync, bb_shared.ConnectionTypeDynamodb, bb_conns.NewDynamoDbSyncBuilder(transformerclient))
			case bb_shared.ConnectionTypeMongo:
				b.Register(bb_internal.JobTypeSync, bb_shared.ConnectionTypeMongo, bb_conns.NewMongoDbSyncBuilder(transformerclient))
			case bb_shared.ConnectionTypeGCP:
				b.Register(bb_internal.JobTypeSync, bb_shared.ConnectionTypeGCP, bb_conns.NewGcpCloudStorageSyncBuilder())
			default:
				return fmt.Errorf("unsupport connection type for sync job: %s", connectionType)
			}
		}
	}

	if jobType == bb_internal.JobTypeAIGenerate {
		if len(destinationConnections) != 1 {
			return fmt.Errorf("unsupported destination count for AI generate job: %d", len(destinationConnections))
		}
		destConnType, err := bb_shared.GetConnectionType(destinationConnections[0])
		if err != nil {
			return err
		}
		driver, err := bb_internal.GetSqlDriverByConnectionType(destConnType)
		if err != nil {
			return err
		}
		builder := bb_conns.NewGenerateAIBuilder(transformerclient, sqlmanagerclient, connectionclient, driver)
		b.Register(bb_internal.JobTypeAIGenerate, bb_shared.ConnectionTypeOpenAI, builder)
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

	sourceConnectionType, err := bb_shared.GetConnectionType(sourceConnection)
	if err != nil {
		sourceConnectionType = "unknown"
	}
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
	jobRunId               string
}

// Manages all necessary configuration parameters for creating
// a worker-based Benthos configuration manager
type WorkerBenthosConfig struct {
	Job                    *mgmtv1alpha1.Job
	SourceConnection       *mgmtv1alpha1.Connection
	DestinationConnections []*mgmtv1alpha1.Connection
	JobRunId               string
	MetricLabelKeyVals     map[string]string
	Logger                 *slog.Logger
	Sqlmanagerclient       sqlmanager.SqlManagerClient
	Transformerclient      mgmtv1alpha1connect.TransformersServiceClient
	Connectionclient       mgmtv1alpha1connect.ConnectionServiceClient
	RedisConfig            *neosync_redis.RedisConfig
	MetricsEnabled         bool
	SelectQueryBuilder     bb_shared.SelectQueryMapBuilder
}

// Creates a new BenthosConfigManager configured for worker
func NewWorkerBenthosConfigManager(
	config *WorkerBenthosConfig,
) (*BenthosConfigManager, error) {
	provider := NewBuilderProvider(config.Logger)
	err := validateBenthosConfig(
		config.Job,
		config.SourceConnection,
		config.DestinationConnections,
	)
	if err != nil {
		return nil, err
	}
	err = provider.registerStandardBuilders(
		config.Job,
		config.SourceConnection,
		config.DestinationConnections,
		config.Sqlmanagerclient,
		config.Transformerclient,
		config.Connectionclient,
		config.RedisConfig,
		config.SelectQueryBuilder,
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
		jobRunId:               config.JobRunId,
	}, nil
}

func validateBenthosConfig(
	job *mgmtv1alpha1.Job,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnections []*mgmtv1alpha1.Connection,
) error {
	if sourceConnection == nil {
		return fmt.Errorf("data sync job must have a source connection")
	}
	if len(destinationConnections) == 0 {
		return fmt.Errorf("data sync job must have at least one destination connection")
	}
	if len(job.GetDestinations()) == 0 {
		return fmt.Errorf("data sync job must have at least one destination")
	}
	if job.GetSource() == nil {
		return fmt.Errorf("data sync job must have a source")
	}
	return nil
}

// Manages all necessary configuration parameters for creating
// a CLI-based Benthos configuration manager
type CliBenthosConfig struct {
	Job                   *mgmtv1alpha1.Job
	SourceConnection      *mgmtv1alpha1.Connection
	DestinationConnection *mgmtv1alpha1.Connection
	SourceJobRunId        *string // for use when AWS S3 is the source
	SyncConfigs           []*runconfigs.RunConfig
	JobRunId              string
	MetricLabelKeyVals    map[string]string
	Logger                *slog.Logger
	Sqlmanagerclient      sqlmanager.SqlManagerClient
	Transformerclient     mgmtv1alpha1connect.TransformersServiceClient
	Connectiondataclient  mgmtv1alpha1connect.ConnectionDataServiceClient
	RedisConfig           *neosync_redis.RedisConfig
	MetricsEnabled        bool
}

// Creates a new BenthosConfigManager configured for CLI
func NewCliBenthosConfigManager(
	config *CliBenthosConfig,
) (*BenthosConfigManager, error) {
	err := validateBenthosConfig(
		config.Job,
		config.SourceConnection,
		[]*mgmtv1alpha1.Connection{config.DestinationConnection},
	)
	if err != nil {
		return nil, err
	}
	destinationProvider := NewBuilderProvider(config.Logger)
	err = destinationProvider.registerStandardBuilders(
		config.Job,
		config.SourceConnection,
		[]*mgmtv1alpha1.Connection{config.DestinationConnection},
		config.Sqlmanagerclient,
		config.Transformerclient,
		nil,
		config.RedisConfig,
		nil,
	)
	if err != nil {
		return nil, err
	}

	sourceProvider, err := NewCliSourceBuilderProvider(config)
	if err != nil {
		return nil, err
	}

	logger := config.Logger.With(withBenthosConfigLoggerTags(config.Job, config.SourceConnection)...)
	return &BenthosConfigManager{
		sourceProvider:         sourceProvider,
		destinationProvider:    destinationProvider,
		metricsEnabled:         config.MetricsEnabled,
		logger:                 logger,
		job:                    config.Job,
		sourceConnection:       config.SourceConnection,
		destinationConnections: []*mgmtv1alpha1.Connection{config.DestinationConnection},
		jobRunId:               config.JobRunId,
	}, nil
}

// NewCliSourceBuilderProvider creates a specialized provider for CLI source operations
func NewCliSourceBuilderProvider(
	config *CliBenthosConfig,
) (*BuilderProvider, error) {
	provider := NewBuilderProvider(config.Logger)

	sourceConnectionType, err := bb_shared.GetConnectionType(config.SourceConnection)
	if err != nil {
		return nil, err
	}
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
		case bb_shared.ConnectionTypePostgres, bb_shared.ConnectionTypeMysql,
			bb_shared.ConnectionTypeMssql, bb_shared.ConnectionTypeAwsS3, bb_shared.ConnectionTypeDynamodb:
			provider.Register(bb_internal.JobTypeSync, sourceConnectionType, builder)
		}
	}

	return provider, nil
}
