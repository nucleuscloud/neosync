package benthos_builder

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeMysql    DatabaseType = "mysql"
	DatabaseTypeMssql    DatabaseType = "mssql"
	DatabaseTypeS3       DatabaseType = "s3"
	DatabaseTypeMongo    DatabaseType = "mongodb"
	DatabaseTypeDynamodb DatabaseType = "dynamodb"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeSync       JobType = "sync"
	JobTypeGenerate   JobType = "generate"
	JobTypeAIGenerate JobType = "ai-generate"
)

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
	JobId             string
	WorkflowId        string
	RunId             string
	RedisConfig       *shared.RedisConfig
	MetricsEnabled    bool
}

// DestinationParams contains all parameters needed to build a destination configuration
type DestinationParams struct {
	SourceConfig      *BenthosSourceConfig
	DestinationIdx    int
	Destination       *mgmtv1alpha1.JobDestination
	DestConnection    *mgmtv1alpha1.Connection
	Logger            *slog.Logger
	TransformerClient mgmtv1alpha1connect.TransformersServiceClient
	SqlManager        sqlmanager.SqlManagerClient
	JobId             string
	WorkflowId        string
	RunId             string
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
	DatabaseType      DatabaseType
	FlowType          JobType
	MetricLabels      []string
}

// BenthosDestinationConfig represents a Benthos destination configuration
type BenthosDestinationConfig struct {
	Outputs     []neosync_benthos.Outputs
	BenthosDsns []*shared.BenthosDsn
}

// New creates a new BenthosBuilder based on database type
func New(dbType DatabaseType) (DatabaseBenthosBuilder, error) {
	switch dbType {
	case DatabaseTypePostgres:
		return newPostgresBuilder(), nil
	case DatabaseTypeMysql:
		return newMysqlBuilder(), nil
	case DatabaseTypeMssql:
		return newMssqlBuilder(), nil
	case DatabaseTypeS3:
		return newS3Builder(), nil
	case DatabaseTypeMongo:
		return newMongoBuilder(), nil
	case DatabaseTypeDynamodb:
		return newDynamoBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
