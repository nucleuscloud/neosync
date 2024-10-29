package benthosbuilder_shared

import (
	"context"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

// ConnectionType represents supported connection types
type ConnectionType string

const (
	ConnectionTypePostgres    ConnectionType = "postgres"
	ConnectionTypeMysql       ConnectionType = "mysql"
	ConnectionTypeMssql       ConnectionType = "mssql"
	ConnectionTypeS3          ConnectionType = "aws-s3"
	ConnectionTypeGCP         ConnectionType = "gcp-cloud-storage"
	ConnectionTypeMongo       ConnectionType = "mongodb"
	ConnectionTypeDynamodb    ConnectionType = "aws-dynamodb"
	ConnectionTypeLocalDir    ConnectionType = "local-directory"
	ConnectionTypeOpenAI      ConnectionType = "openai"
	ConnectionTypeNeosyncData ConnectionType = "neosync-data-stream"
)

// Determines type fo connection from Connection
func GetConnectionType(connection *mgmtv1alpha1.Connection) ConnectionType {
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

// Determines type of job from Job
func GetJobType(job *mgmtv1alpha1.Job) JobType {
	switch job.GetSource().GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres,
		*mgmtv1alpha1.JobSourceOptions_Mysql,
		*mgmtv1alpha1.JobSourceOptions_Mssql,
		*mgmtv1alpha1.JobSourceOptions_Mongodb,
		*mgmtv1alpha1.JobSourceOptions_Dynamodb,
		*mgmtv1alpha1.JobSourceOptions_AwsS3:
		return JobTypeSync
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		return JobTypeGenerate
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		return JobTypeAIGenerate
	default:
		return ""
	}
}

// source: pg - PgConnectionBenthosBUIlder (I know how to calcualte pg source data, and I support destionaation PG, Mysql, S3, etc.)
// source: mysql
// source: mssql

type ConnectionBenthosBuilder interface {
	// Builds benthos source configs
	BuildSourceConfigs(ctx context.Context, params *SourceParams) ([]*BenthosSourceConfig, error) // benthos input

	// BuildProcessors?

	// Builds a benthos destination config
	BuildDestinationConfig(ctx context.Context, params *DestinationParams) (*BenthosDestinationConfig, error) // benthos output

}

// SourceParams contains all parameters needed to build a source benthos configuration
type SourceParams struct {
	Job              *mgmtv1alpha1.Job
	RunId            string
	SourceConnection *mgmtv1alpha1.Connection
	Logger           *slog.Logger
}

type ReferenceKey struct {
	Table  string
	Column string
}

// DestinationParams contains all parameters needed to build a destination benthos configuration
type DestinationParams struct {
	SourceConfig    *BenthosSourceConfig
	Job             *mgmtv1alpha1.Job
	RunId           string
	DestinationIdx  int
	DestinationOpts *mgmtv1alpha1.JobDestinationOptions
	DestConnection  *mgmtv1alpha1.Connection
	Logger          *slog.Logger
}

// BenthosSourceConfig represents a Benthos source configuration
type BenthosSourceConfig struct {
	Config                  *neosync_benthos.BenthosConfig
	Name                    string
	DependsOn               []*tabledependency.DependsOn
	RunType                 tabledependency.RunType
	TableSchema             string
	TableName               string
	Columns                 []string
	RedisDependsOn          map[string][]string
	ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	Processors              []*neosync_benthos.ProcessorConfig
	BenthosDsns             []*benthosbuilder.BenthosDsn
	RedisConfig             []*benthosbuilder.BenthosRedisConfig
	PrimaryKeys             []string
	Metriclabels            metrics.MetricLabels
}

// BenthosDestinationConfig represents a Benthos destination configuration
type BenthosDestinationConfig struct {
	Outputs     []neosync_benthos.Outputs
	BenthosDsns []*benthosbuilder.BenthosDsn
}
