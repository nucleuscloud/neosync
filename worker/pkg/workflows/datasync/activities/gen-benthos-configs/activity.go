package genbenthosconfigs_activity

import (
	"context"
	"sync"
	"time"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	logger_utils "github.com/nucleuscloud/neosync/worker/internal/logger"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type GenerateBenthosConfigsRequest struct {
	JobId      string
	WorkflowId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*BenthosConfigResponse
}

type BenthosRedisConfig struct {
	Key    string
	Table  string // schema.table
	Column string
}

type BenthosConfigResponse struct {
	Name        string
	DependsOn   []*tabledependency.DependsOn
	RunType     tabledependency.RunType
	Config      *neosync_benthos.BenthosConfig
	TableSchema string
	TableName   string
	Columns     []string

	Processors  []*neosync_benthos.ProcessorConfig
	BenthosDsns []*shared.BenthosDsn
	RedisConfig []*BenthosRedisConfig

	primaryKeys []string

	metriclabels metrics.MetricLabels
}

type Activity struct {
	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	sqlconnector sqlconnect.SqlConnector

	redisConfig *shared.RedisConfig

	pgquerier    pg_queries.Querier
	mysqlquerier mysql_queries.Querier

	metricsEnabled bool
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlconnector sqlconnect.SqlConnector,
	redisConfig *shared.RedisConfig,
	metricsEnabled bool,
) *Activity {
	return &Activity{
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		sqlconnector:      sqlconnector,
		redisConfig:       redisConfig,
		pgquerier:         pg_queries.New(),
		mysqlquerier:      mysql_queries.New(),
		metricsEnabled:    metricsEnabled,
	}
}

func (a *Activity) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
) (*GenerateBenthosConfigsResponse, error) {
	info := activity.GetInfo(ctx)
	loggerKeyVals := []any{
		"jobId", req.JobId,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
	}
	logger := log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
	)
	_ = logger
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-activity.GetWorkerStopChannel(ctx):
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}

	sqladapter := sql_manager.NewSqlManager(pgpoolmap, a.pgquerier, mysqlpoolmap, a.mysqlquerier, a.sqlconnector)

	bbuilder := newBenthosBuilder(
		*sqladapter,
		a.jobclient,
		a.connclient,
		a.transformerclient,
		req.JobId,
		info.WorkflowExecution.RunID,
		a.redisConfig,
		a.metricsEnabled,
	)
	slogger := logger_utils.NewJsonSLogger().With(loggerKeyVals...)
	return bbuilder.GenerateBenthosConfigs(ctx, req, slogger)
}
