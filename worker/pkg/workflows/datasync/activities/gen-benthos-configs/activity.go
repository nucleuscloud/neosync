package genbenthosconfigs_activity

import (
	"context"
	"time"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type GenerateBenthosConfigsRequest struct {
	JobId    string
	JobRunId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*benthosbuilder.BenthosConfigResponse
	AccountId      string
}

type Activity struct {
	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	sqlmanager sql_manager.SqlManagerClient

	redisConfig *neosync_redis.RedisConfig

	metricsEnabled bool

	pageLimit int
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanager sql_manager.SqlManagerClient,
	redisConfig *neosync_redis.RedisConfig,
	metricsEnabled bool,
	pageLimit int,
) *Activity {
	return &Activity{
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		sqlmanager:        sqlmanager,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
		pageLimit:         pageLimit,
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
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
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
		req.JobRunId,
		info.WorkflowExecution.RunID,
		a.redisConfig,
		a.metricsEnabled,
		a.pageLimit,
	)
	slogger := temporallogger.NewSlogger(logger)
	return bbuilder.GenerateBenthosConfigsNew(ctx, req, &workflowMetadata{WorkflowId: info.WorkflowExecution.ID, RunId: info.WorkflowExecution.RunID}, slogger)
}
