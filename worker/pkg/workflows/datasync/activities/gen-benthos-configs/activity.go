package genbenthosconfigs_activity

import (
	"context"
	"time"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type GenerateBenthosConfigsRequest struct {
	JobId string
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
	return bbuilder.GenerateBenthosConfigsNew(ctx, req, &workflowMetadata{WorkflowId: info.WorkflowExecution.ID}, slogger)
}
