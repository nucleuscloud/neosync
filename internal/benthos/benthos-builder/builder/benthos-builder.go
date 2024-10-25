package benthos_builder

import (
	"context"
	"fmt"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_connections "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/connections"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

// BenthosManager handles the overall process of building Benthos configs
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

// BenthosBuilder is the main interface for building Benthos configs
type BenthosBuilder interface {
	// BuildSourceConfig builds the source configuration for a given database and flow type
	BuildSourceConfig(ctx context.Context, params *bb_shared.SourceParams) (*bb_shared.BenthosSourceConfig, error)

	// BuildDestinationConfig builds the destination configuration
	BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error)
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
	RedisConfig []*bb_shared.BenthosRedisConfig

	primaryKeys []string

	metriclabels metrics.MetricLabels
}

// New creates a new BenthosBuilder based on database type
func NewBenthosBuilder(connType bb_shared.ConnectionType) (bb_shared.DatabaseBenthosBuilder, error) {
	switch connType {
	case bb_shared.ConnectionTypePostgres:
		return bb_connections.NewPostgresBuilder(), nil
	// case ConnectionTypeMysql:
	// 	return newMysqlBuilder(), nil
	// case ConnectionTypeMssql:
	// 	return newMssqlBuilder(), nil
	// case ConnectionTypeS3:
	// 	return newS3Builder(), nil
	case bb_shared.ConnectionTypeMongo:
		return bb_connections.NewMongoDbBuilder(), nil
	// case ConnectionTypeDynamodb:
	// 	return newDynamoBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connType)
	}
}
