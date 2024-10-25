package benthos_builder

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

// ConnectionType represents supported connection types
type ConnectionType string

const (
	ConnectionTypePostgres ConnectionType = "postgres"
	ConnectionTypeMysql    ConnectionType = "mysql"
	ConnectionTypeMssql    ConnectionType = "mssql"
	ConnectionTypeS3       ConnectionType = "aws-s3"
	ConnectionTypeGCP      ConnectionType = "gcp-cloud-storage"
	ConnectionTypeMongo    ConnectionType = "mongodb"
	ConnectionTypeDynamodb ConnectionType = "aws-dynamodb"
	ConnectionTypeLocalDir ConnectionType = "local-directory"
	ConnectionTypeOpenAI   ConnectionType = "openai"
)

func getConnectionType(connection *mgmtv1alpha1.Connection) ConnectionType {
	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return ConnectionTypePostgres
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return ConnectionTypeMysql
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return ConnectionTypeMssql
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return ConnectionTypeS3
	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		return ConnectionTypeGCP
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return ConnectionTypeMongo
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return ConnectionTypeDynamodb
	case *mgmtv1alpha1.ConnectionConfig_LocalDirConfig:
		return ConnectionTypeLocalDir
	case *mgmtv1alpha1.ConnectionConfig_OpenaiConfig:
		return ConnectionTypeOpenAI
	default:
		return "unknown"
	}
}

// JobType represents the type of job
type JobType string

const (
	JobTypeSync       JobType = "sync"
	JobTypeGenerate   JobType = "generate"
	JobTypeAIGenerate JobType = "ai-generate"
)

type BenthosRedisConfig struct {
	Key    string
	Table  string // schema.table
	Column string
}

// BenthosManager handles the overall process of building Benthos configs
type BenthosConfigManager struct {
	sqlmanager sqlmanager.SqlManagerClient
	// jobclient         mgmtv1alpha1connect.JobServiceClient
	// connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
	redisConfig       *shared.RedisConfig
	metricsEnabled    bool
}

// Creates a new Benthos Manager. Used for creating benthos configs
func NewBenthosConfigManager(
	sqlmanager sqlmanager.SqlManagerClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	jobId string,
	workflowId string,
	runId string,
	redisConfig *shared.RedisConfig,
	metricsEnabled bool,
) *BenthosConfigManager {
	return &BenthosConfigManager{
		sqlmanager: sqlmanager,
		// jobclient:         jobclient,
		// connclient:        connclient,
		transformerclient: transformerclient,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
	}
}

// BenthosBuilder is the main interface for building Benthos configs
type BenthosBuilder interface {
	// BuildSourceConfig builds the source configuration for a given database and flow type
	BuildSourceConfig(ctx context.Context, params *SourceParams) (*BenthosSourceConfig, error)

	// BuildDestinationConfig builds the destination configuration
	BuildDestinationConfig(ctx context.Context, params *DestinationParams) (*BenthosDestinationConfig, error)
}

// DatabaseBenthosBuilder is the interface that each database type must implement
type DatabaseBenthosBuilder interface {
	// BuildSyncSourceConfig builds a sync source configuration
	BuildSyncSourceConfig(ctx context.Context, params *SourceParams) (*BenthosSourceConfig, error)

	// BuildGenerateSourceConfig builds a generate source configuration
	BuildGenerateSourceConfig(ctx context.Context, params *SourceParams) (*BenthosSourceConfig, error)

	// BuildAIGenerateSourceConfig builds an AI generate source configuration
	BuildAIGenerateSourceConfig(ctx context.Context, params *SourceParams) (*BenthosSourceConfig, error)

	// BuildDestinationConfig builds a destination configuration
	BuildDestinationConfig(ctx context.Context, params *DestinationParams) (*BenthosDestinationConfig, error)
}

// SourceParams contains all parameters needed to build a source configuration
type SourceParams struct {
	Job               *mgmtv1alpha1.Job
	SourceConnection  *mgmtv1alpha1.Connection
	Logger            *slog.Logger
	TransformerClient mgmtv1alpha1connect.TransformersServiceClient
	SqlManager        sqlmanager.SqlManagerClient
	RedisConfig       *shared.RedisConfig
	MetricsEnabled    bool
}

type referenceKey struct {
	Table  string
	Column string
}

// DestinationParams contains all parameters needed to build a destination configuration
type DestinationParams struct {
	SourceConfig      *BenthosSourceConfig
	DestinationIdx    int
	DestinationOpts   *mgmtv1alpha1.JobDestinationOptions
	DestConnection    *mgmtv1alpha1.Connection
	Logger            *slog.Logger
	TransformerClient mgmtv1alpha1connect.TransformersServiceClient
	SqlManager        sqlmanager.SqlManagerClient
	RedisConfig       *shared.RedisConfig
	MetricsEnabled    bool
	// Additional fields specific to source type
	PrimaryKeyToFKMap   map[string]map[string][]*referenceKey
	ColTransformerMap   map[string]map[string]*mgmtv1alpha1.JobMappingTransformer
	SchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo
}

// BenthosSourceConfig represents a Benthos source configuration
type BenthosSourceConfig struct {
	Config            *neosync_benthos.BenthosConfig
	Name              string
	DependsOn         []*tabledependency.DependsOn
	RunType           tabledependency.RunType
	TableSchema       string
	TableName         string
	Columns           []string
	RedisDependsOn    map[string][]string
	DefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	Processors        []*neosync_benthos.ProcessorConfig
	BenthosDsns       []*shared.BenthosDsn
	RedisConfig       []*BenthosRedisConfig
	PrimaryKeys       []string
	ConnectionType    ConnectionType
	JobType           JobType
	MetricLabels      []string
}

// BenthosDestinationConfig represents a Benthos destination configuration
type BenthosDestinationConfig struct {
	Outputs     []neosync_benthos.Outputs
	BenthosDsns []*shared.BenthosDsn
}

type BenthosConfigResponse struct {
	Name                    string
	DependsOn               []*tabledependency.DependsOn
	RunType                 tabledependency.RunType
	Config                  *neosync_benthos.BenthosConfig
	TableSchema             string
	TableName               string
	Columns                 []string
	RedisDependsOn          map[string][]string
	ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	SourceConnectionType    string // used for logging

	Processors  []*neosync_benthos.ProcessorConfig
	BenthosDsns []*shared.BenthosDsn
	RedisConfig []*BenthosRedisConfig

	primaryKeys []string

	metriclabels metrics.MetricLabels
}

// New creates a new BenthosBuilder based on database type
func NewBenthosBuilder(connType ConnectionType) (DatabaseBenthosBuilder, error) {
	switch connType {
	case ConnectionTypePostgres:
		return NewPostgresBuilder(), nil
	// case ConnectionTypeMysql:
	// 	return newMysqlBuilder(), nil
	// case ConnectionTypeMssql:
	// 	return newMssqlBuilder(), nil
	// case ConnectionTypeS3:
	// 	return newS3Builder(), nil
	// case ConnectionTypeMongo:
	// 	return newMongoBuilder(), nil
	// case ConnectionTypeDynamodb:
	// 	return newDynamoBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connType)
	}
}
