package genbenthosconfigs_activity

import (
	"context"
	"time"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type GenerateBenthosConfigsRequest struct {
	JobId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*BenthosConfigResponse
	AccountId      string
}

type BenthosRedisConfig struct {
	Key    string
	Table  string // schema.table
	Column string
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

	sqlmanager sql_manager.SqlManagerClient

	redisConfig *shared.RedisConfig

	metricsEnabled bool
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanager sql_manager.SqlManagerClient,
	redisConfig *shared.RedisConfig,
	metricsEnabled bool,
) *Activity {
	return &Activity{
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		sqlmanager:        sqlmanager,
		redisConfig:       redisConfig,
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

	bbuilder := newBenthosBuilder(
		a.sqlmanager,
		a.jobclient,
		a.connclient,
		a.transformerclient,
		req.JobId,
		info.WorkflowExecution.ID,
		info.WorkflowExecution.RunID,
		a.redisConfig,
		a.metricsEnabled,
	)
	slogger := neosynclogger.NewJsonSLogger().With(loggerKeyVals...)
	return bbuilder.GenerateBenthosConfigs(ctx, req, &workflowMetadata{WorkflowId: info.WorkflowExecution.ID}, slogger)
}
