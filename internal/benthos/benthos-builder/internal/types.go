package benthosbuilder_internal

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
)

// Determines SQL driver from connection type
func GetSqlDriverByConnectionType(connectionType bb_shared.ConnectionType) (string, error) {
	switch connectionType {
	case bb_shared.ConnectionTypePostgres:
		return sqlmanager_shared.PostgresDriver, nil
	case bb_shared.ConnectionTypeMysql:
		return sqlmanager_shared.MysqlDriver, nil
	case bb_shared.ConnectionTypeMssql:
		return sqlmanager_shared.MssqlDriver, nil
	default:
		return "", fmt.Errorf("unsupported SQL connection type: %s", connectionType)
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

// Handles both source (input) and destination (output) configurations for different
// connection types (postgres, mysql...) and job types (e.g., sync, generate...).
type BenthosBuilder interface {
	// BuildSourceConfigs generates Benthos source configurations for reading and processing data.
	// Returns a config for each schema.table in job mappings
	BuildSourceConfigs(ctx context.Context, params *SourceParams) ([]*BenthosSourceConfig, error)
	// BuildDestinationConfig creates a Benthos destination configuration for writing processed data.
	// Returns single config for a schema.table configuration
	BuildDestinationConfig(ctx context.Context, params *DestinationParams) (*BenthosDestinationConfig, error)
}

// SourceParams contains all parameters needed to build a source benthos configuration
type SourceParams struct {
	Job              *mgmtv1alpha1.Job
	JobRunId         string
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
	JobRunId        string
	DestinationOpts *mgmtv1alpha1.JobDestinationOptions
	DestConnection  *mgmtv1alpha1.Connection
	Logger          *slog.Logger
}

// BenthosSourceConfig represents a Benthos source configuration
type BenthosSourceConfig struct {
	Config                  *neosync_benthos.BenthosConfig
	Name                    string
	DependsOn               []*runconfigs.DependsOn
	RunType                 runconfigs.RunType
	TableSchema             string
	TableName               string
	Columns                 []string
	RedisDependsOn          map[string][]string
	ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
	Processors              []*neosync_benthos.ProcessorConfig
	BenthosDsns             []*bb_shared.BenthosDsn
	RedisConfig             []*bb_shared.BenthosRedisConfig
	PrimaryKeys             []string
	Metriclabels            metrics.MetricLabels
	ColumnIdentityCursors   map[string]*tablesync_shared.IdentityCursor
}

// BenthosDestinationConfig represents a Benthos destination configuration
type BenthosDestinationConfig struct {
	Outputs     []neosync_benthos.Outputs
	BenthosDsns []*bb_shared.BenthosDsn
}
