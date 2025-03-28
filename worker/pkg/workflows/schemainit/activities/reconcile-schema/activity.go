package reconcileschema_activity

import (
	"context"
	"time"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient

	sqlmanager sql_manager.SqlManagerClient

	eelicense license.EEInterface
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	sqlmanager sql_manager.SqlManagerClient,
	eelicense license.EEInterface,
) *Activity {
	return &Activity{
		jobclient:  jobclient,
		connclient: connclient,
		sqlmanager: sqlmanager,
		eelicense:  eelicense,
	}
}

type RunReconcileSchemaRequest struct {
	JobId         string
	JobRunId      string
	DestinationId string
}

type RunReconcileSchemaResponse struct {
}

func (a *Activity) RunReconcileSchema(
	ctx context.Context,
	req *RunReconcileSchemaRequest,
) (*RunReconcileSchemaResponse, error) {
	info := activity.GetInfo(ctx)
	logger := log.With(
		activity.GetLogger(ctx),
		"jobId", req.JobId,
		"jobRunId", req.JobRunId,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"destinationId", req.DestinationId,
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

	builder := newReconcileSchemaBuilder(
		a.sqlmanager,
		a.jobclient,
		a.connclient,
		a.eelicense,
		req.JobRunId,
	)
	slogger := temporallogger.NewSlogger(logger)
	return builder.RunReconcileSchema(
		ctx,
		req,
		connectionmanager.NewUniqueSession(
			connectionmanager.WithSessionGroup(info.WorkflowExecution.ID),
		),
		slogger,
	)
}
