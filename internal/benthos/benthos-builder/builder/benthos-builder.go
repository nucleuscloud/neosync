package benthos_builder

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	bb_conns "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/connections"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

// Defines Provider methods for creating Benthos builders
type BuilderProvider interface {
	NewBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error)
}

type BenthosConfigManager struct {
	sourceProvider      BuilderProvider
	destinationProvider BuilderProvider
	metricsEnabled      bool
}

type SourceBuilderProvider struct {
	sqlmanagerclient  sqlmanager.SqlManagerClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
	redisConfig       *shared.RedisConfig
}

func NewSourceBuilderProvider(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) *SourceBuilderProvider {
	return &SourceBuilderProvider{
		sqlmanagerclient:  sqlmanagerclient,
		transformerclient: transformerclient,
		redisConfig:       redisConfig,
	}
}

func getDefaultBuilder(
	connType bb_shared.ConnectionType,
	jobType bb_shared.JobType,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) (bb_shared.ConnectionBenthosBuilder, error) {
	switch connType {
	case bb_shared.ConnectionTypePostgres:
		return newPostgresBuilder(jobType, sqlmanagerclient, transformerclient, redisConfig)
	case bb_shared.ConnectionTypeMysql:
		return newPostgresBuilder(jobType, sqlmanagerclient, transformerclient, redisConfig)
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", connType)
	}
}

func (s *SourceBuilderProvider) NewBuilder(
	connType bb_shared.ConnectionType,
	jobType bb_shared.JobType,
) (bb_shared.ConnectionBenthosBuilder, error) {
	return getDefaultBuilder(connType, jobType, s.sqlmanagerclient, s.transformerclient, s.redisConfig)
}

type DestinationBuilderProvider struct {
	sqlmanagerclient  sqlmanager.SqlManagerClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
	redisConfig       *shared.RedisConfig
}

func NewDestinationBuilderProvider(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) *DestinationBuilderProvider {
	return &DestinationBuilderProvider{
		sqlmanagerclient:  sqlmanagerclient,
		transformerclient: transformerclient,
		redisConfig:       redisConfig,
	}
}

func (d *DestinationBuilderProvider) NewBuilder(
	connType bb_shared.ConnectionType,
	jobType bb_shared.JobType,
) (bb_shared.ConnectionBenthosBuilder, error) {
	return getDefaultBuilder(connType, jobType, d.sqlmanagerclient, d.transformerclient, d.redisConfig)
}

// Shared builder creation functions
func newPostgresBuilder(
	jobType bb_shared.JobType,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) (bb_shared.ConnectionBenthosBuilder, error) {
	switch jobType {
	case bb_shared.JobTypeSync:
		return bb_conns.NewPostgresSyncBuilder(transformerclient, sqlmanagerclient, redisConfig), nil
	case bb_shared.JobTypeGenerate:
		return bb_conns.NewPostgresGenerateBuilder(), nil
	case bb_shared.JobTypeAIGenerate:
		return bb_conns.NewPostgresAIGenerateBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported postgres job type: %s", jobType)
	}
}

func NewWorkerBenthosConfigManager(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
	metricsEnabled bool,
) *BenthosConfigManager {
	return &BenthosConfigManager{
		sourceProvider: NewSourceBuilderProvider(
			sqlmanagerclient,
			transformerclient,
			redisConfig,
		),
		destinationProvider: NewDestinationBuilderProvider(
			sqlmanagerclient,
			transformerclient,
			redisConfig,
		),
		metricsEnabled: metricsEnabled,
	}
}

func NewCliBenthosConfigManager(
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
	sourceJobRunId *string,
	syncConfigs []*tabledependency.RunConfig,
	destinationConnection *mgmtv1alpha1.Connection,
) *BenthosConfigManager {
	return &BenthosConfigManager{
		sourceProvider: NewCliSourceBuilderProvider(connectiondataclient, sqlmanagerclient, sourceJobRunId, syncConfigs, destinationConnection),
		destinationProvider: NewDestinationBuilderProvider(
			sqlmanagerclient,
			transformerclient,
			redisConfig,
		),
	}
}

type CliSourceBuilderProvider struct {
	connectiondataclient  mgmtv1alpha1connect.ConnectionDataServiceClient
	sqlmanagerclient      sqlmanager.SqlManagerClient
	sourceJobRunId        *string // when AWS S3 is the source
	syncConfigs           []*tabledependency.RunConfig
	destinationConnection *mgmtv1alpha1.Connection
}

func NewCliSourceBuilderProvider(
	connectionclientdata mgmtv1alpha1connect.ConnectionDataServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceJobRunId *string,
	syncConfigs []*tabledependency.RunConfig,
	destinationConnection *mgmtv1alpha1.Connection,
) *CliSourceBuilderProvider {
	return &CliSourceBuilderProvider{
		connectiondataclient:  connectionclientdata,
		sqlmanagerclient:      sqlmanagerclient,
		sourceJobRunId:        sourceJobRunId,
		syncConfigs:           syncConfigs,
		destinationConnection: destinationConnection,
	}
}

func (f *CliSourceBuilderProvider) NewBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
	return bb_conns.NewNeosyncConnectionDataSyncBuilder(f.connectiondataclient, f.sqlmanagerclient, f.sourceJobRunId, f.syncConfigs, f.destinationConnection), nil
}

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
	BenthosDsns []*benthosbuilder.BenthosDsn
	RedisConfig []*benthosbuilder.BenthosRedisConfig

	// primaryKeys []string

	// metriclabels metrics.MetricLabels
}

// type BenthosConfigManager interface {
// 	NewSourceBenthosBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error)
// 	NewDestinationBenthosBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error)
// }

// // Base implementation with shared config generation logic
// type baseBenthosConfigManager struct {
// 	manager BenthosConfigManager // Holds concrete manager implementation
// }
// // Handles the overall process of building Benthos configs
// type WorkerBenthosConfigManager struct {
// 	sqlmanager        *sqlmanager.SqlManagerClient
// 	transformerclient mgmtv1alpha1connect.TransformersServiceClient
// 	redisConfig       *shared.RedisConfig
// 	metricsEnabled    bool
// }

// // Creates a new Benthos Manager. Used for creating benthos configs
// func NewWorkerBenthosConfigManager(
// 	sqlmanagerclient *sqlmanager.SqlManagerClient,
// 	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
// 	redisConfig *shared.RedisConfig,
// 	metricsEnabled bool,
// ) *WorkerBenthosConfigManager {
// 	return &WorkerBenthosConfigManager{
// 		sqlmanager:        sqlmanagerclient,
// 		transformerclient: transformerclient,
// 		redisConfig:       redisConfig,
// 		metricsEnabled:    metricsEnabled,
// 	}
// }

// // Handles the overall process of building Benthos configs
// type BenthosCliConfigManager struct {
// 	sqlmanager           *sqlmanager.SqlManagerClient
// 	transformerclient    mgmtv1alpha1connect.TransformersServiceClient
// 	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient
// 	metricsEnabled       bool
// }

// // Creates a new CLI Benthos Manger. Used for creating benthos configs running in CLI
// func NewCliBenthosConfigManager(
// 	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
// ) *BenthosCliConfigManager {
// 	return &BenthosCliConfigManager{
// 		sqlmanager:           nil,
// 		transformerclient:    nil,
// 		metricsEnabled:       false,
// 		connectiondataclient: connectiondataclient,
// 	}
// }

// type BenthosConfigResponse struct {
// 	Name                    string
// 	DependsOn               []*tabledependency.DependsOn
// 	RunType                 tabledependency.RunType
// 	Config                  *neosync_benthos.BenthosConfig
// 	TableSchema             string
// 	TableName               string
// 	Columns                 []string
// 	RedisDependsOn          map[string][]string
// 	ColumnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties
// 	SourceConnectionType    string // used for logging

// 	Processors  []*neosync_benthos.ProcessorConfig
// 	BenthosDsns []*bb_shared.BenthosDsn
// 	RedisConfig []*bb_shared.BenthosRedisConfig

// 	primaryKeys []string

// 	metriclabels metrics.MetricLabels
// }

// // Creates a new ConnectionBenthosBuilder based on connection and job type
// func (b *WorkerBenthosConfigManager) NewBenthosBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
// 	switch connType {
// 	case bb_shared.ConnectionTypePostgres:
// 		return b.NewPostgresBenthosBuilder(jobType)
// 	case bb_shared.ConnectionTypeMysql:
// 		return bb_conns.NewMysqlBenthosBuilder(jobType)
// 	default:
// 		return nil, fmt.Errorf("unsupported connection type and job type: [%s, %s]", connType, jobType)
// 	}
// }

// func (b *WorkerBenthosConfigManager) NewPostgresBenthosBuilder(jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
// 	switch jobType {
// 	case bb_shared.JobTypeSync:
// 		return bb_conns.NewPostgresSyncBuilder(), nil
// 	case bb_shared.JobTypeGenerate:
// 		return bb_conns.NewPostgresGenerateBuilder(), nil
// 	case bb_shared.JobTypeAIGenerate:
// 		return bb_conns.NewPostgresAIGenerateBuilder(), nil
// 	default:
// 		return nil, fmt.Errorf("unsupported postgres job type: %s", jobType)
// 	}
// }

// func (b *WorkerBenthosConfigManager) NewPostgresBenthosBuilder(jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
// 	switch jobType {
// 	case bb_shared.JobTypeSync:
// 		return bb_conns.NewPostgresSyncBuilder(), nil
// 	case bb_shared.JobTypeGenerate:
// 		return bb_conns.NewPostgresGenerateBuilder(), nil
// 	case bb_shared.JobTypeAIGenerate:
// 		return bb_conns.NewPostgresAIGenerateBuilder(), nil
// 	default:
// 		return nil, fmt.Errorf("unsupported postgres job type: %s", jobType)
// 	}
// }

// // registers input and outputs
// // Creates a new ConnectionBenthosBuilder based on connection and job type
// func (b *BenthosCliConfigManager) NewSourceBenthosBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
// 	return bb_conns.NewNeosyncConnectionDataStreamBuilder()
// }

// func (b *BenthosCliConfigManager) NewDestinationBenthosBuilder(connType bb_shared.ConnectionType, jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
// 	switch connType {
// 	case bb_shared.ConnectionTypePostgres:
// 		return b.NewPostgresBenthosBuilder(jobType)
// 	case bb_shared.ConnectionTypeMysql:
// 		return bb_conns.NewMysqlBenthosBuilder(jobType)
// 	default:
// 		return nil, fmt.Errorf("unsupported connection type and job type: [%s, %s]", connType, jobType)
// 	}
// }

// func (b *BenthosCliConfigManager) NewPostgresBenthosBuilder(jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
// 	switch jobType {
// 	case bb_shared.JobTypeSync:
// 		return bb_conns.NewPostgresSyncBuilder(), nil
// 	case bb_shared.JobTypeGenerate:
// 		return bb_conns.NewPostgresGenerateBuilder(), nil
// 	case bb_shared.JobTypeAIGenerate:
// 		return bb_conns.NewPostgresAIGenerateBuilder(), nil
// 	default:
// 		return nil, fmt.Errorf("unsupported postgres job type: %s", jobType)
// 	}
// }
