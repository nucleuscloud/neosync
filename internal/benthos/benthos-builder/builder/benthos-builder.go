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

// Creates a ConnectionBenthosBuilder
type BenthosBuilders func(
	jobType bb_shared.JobType,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) (bb_shared.ConnectionBenthosBuilder, error)

type BuilderKey struct {
	ConnType bb_shared.ConnectionType
	JobType  bb_shared.JobType
}

func (b *BuilderKey) String() string {
	return fmt.Sprintf("%s.%s", b.JobType, b.ConnType)
}

// BuilderRegistry maintains a mapping of connection types to benthos builders
type BuilderProvider struct {
	builders map[string]bb_shared.ConnectionBenthosBuilder
}

// Creates a new BuilderRegistry with default builders registered
func NewBuilderProvider() *BuilderProvider {
	r := &BuilderProvider{
		builders: make(map[string]bb_shared.ConnectionBenthosBuilder),
	}
	return r
}

func (r *BuilderProvider) Register(jobType bb_shared.JobType, connType bb_shared.ConnectionType, builder bb_shared.ConnectionBenthosBuilder) {
	key := BuilderKey{ConnType: connType, JobType: jobType}
	r.builders[key.String()] = builder
}

// Creates a new builder for the given connection and job type
func (r *BuilderProvider) GetBuilder(
	connType bb_shared.ConnectionType,
	jobType bb_shared.JobType,
) (bb_shared.ConnectionBenthosBuilder, error) {
	key := BuilderKey{ConnType: connType, JobType: jobType}
	builder, exists := r.builders[key.String()]
	if !exists {
		return nil, fmt.Errorf("unsupported connection type: %s", connType)
	}
	return builder, nil
}

func (b *BuilderProvider) registerStandardBuilders(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
) {
	// be smarter about registering these based on job
	sqlbuilder := bb_conns.NewSqlSyncBuilder(transformerclient, sqlmanagerclient, redisConfig)
	b.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypePostgres, sqlbuilder)
	b.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeMysql, sqlbuilder)
	b.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeMssql, sqlbuilder)
}

type BenthosConfigManager struct {
	sourceProvider      *BuilderProvider
	destinationProvider *BuilderProvider
	metricsEnabled      bool
}

func NewWorkerBenthosConfigManager(
	sqlmanagerclient sqlmanager.SqlManagerClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	redisConfig *shared.RedisConfig,
	metricsEnabled bool,
) *BenthosConfigManager {
	fmt.Println()
	fmt.Println("NEW WORKER BENTHOS CONFIG MANAGER")
	fmt.Println()
	provider := NewBuilderProvider()
	provider.registerStandardBuilders(sqlmanagerclient, transformerclient, redisConfig)
	return &BenthosConfigManager{
		sourceProvider:      provider,
		destinationProvider: provider,
		metricsEnabled:      metricsEnabled,
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
	destinationProvider := NewBuilderProvider()
	return &BenthosConfigManager{
		sourceProvider: NewCliSourceBuilderProvider(
			connectiondataclient,
			sqlmanagerclient,
			sourceJobRunId,
			syncConfigs,
			destinationConnection,
		),
		destinationProvider: destinationProvider,
		metricsEnabled:      false,
	}
}

// NewCliSourceBuilderProvider creates a specialized provider for CLI source operations
func NewCliSourceBuilderProvider(
	connectionclientdata mgmtv1alpha1connect.ConnectionDataServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceJobRunId *string,
	syncConfigs []*tabledependency.RunConfig,
	destinationConnection *mgmtv1alpha1.Connection,
) *BuilderProvider {
	provider := NewBuilderProvider()

	// Register CLI-specific builder constructor
	builder := bb_conns.NewNeosyncConnectionDataSyncBuilder(
		connectionclientdata,
		sqlmanagerclient,
		sourceJobRunId,
		syncConfigs,
		destinationConnection,
	)

	// be smarter about registering these based on job
	provider.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypePostgres, builder)
	provider.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeMysql, builder)
	provider.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeMssql, builder)
	provider.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeGCP, builder)
	provider.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeS3, builder)
	provider.Register(bb_shared.JobTypeSync, bb_shared.ConnectionTypeDynamodb, builder)

	return provider
}
