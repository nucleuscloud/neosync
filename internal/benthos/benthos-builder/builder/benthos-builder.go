package benthos_builder

import (
	"fmt"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_conns "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/connections"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

// Handles the overall process of building Benthos configs
type BenthosConfigManager struct {
	sqlmanager        sqlmanager.SqlManagerClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
	redisConfig       *shared.RedisConfig
	metricsEnabled    bool
}

// Creates a new Benthos Manager. Used for creating benthos configs
func NewBenthosConfigManager(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
	metricsEnabled bool,
) *BenthosConfigManager {
	return &BenthosConfigManager{
		sqlmanager:        sqlmanagerclient,
		transformerclient: transformerclient,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
	}
}

// type BenthosBuilder interface {
// 	BuildSourceConfigs(ctx context.Context, params *bb_shared.SourceParams) ([]*bb_shared.BenthosSourceConfig, error)

// 	BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error)
// }

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
	BenthosDsns []*bb_shared.BenthosDsn
	RedisConfig []*bb_shared.BenthosRedisConfig

	primaryKeys []string

	metriclabels metrics.MetricLabels
}

// Creates a new ConnectionBenthosBuilder based on connection and job type
func NewBenthosBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
	switch connType {
	case bb_shared.ConnectionTypePostgres:
		return bb_conns.NewPostgresBenthosBuilder(jobType)
	case bb_shared.ConnectionTypeMysql:
		return bb_conns.NewMysqlBenthosBuilder(jobType)
	default:
		return nil, fmt.Errorf("unsupported connection type and job type: [%s, %s]", connType, jobType)
	}
}
