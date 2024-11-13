package runsqlinittablestmts_activity

import (
	"context"
	"time"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient

	sqlmanager sql_manager.SqlManagerClient

	eelicense      *license.EELicense
	isNeosyncCloud bool
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	sqlmanager sql_manager.SqlManagerClient,
	eelicense *license.EELicense,
	isNeosyncCloud bool,
) *Activity {
	return &Activity{
		jobclient:      jobclient,
		connclient:     connclient,
		sqlmanager:     sqlmanager,
		eelicense:      eelicense,
		isNeosyncCloud: isNeosyncCloud,
	}
}

type RunSqlInitTableStatementsRequest struct {
	JobId string
}

type RunSqlInitTableStatementsResponse struct {
}

func (a *Activity) RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
) (*RunSqlInitTableStatementsResponse, error) {
	info := activity.GetInfo(ctx)
	logger := log.With(
		activity.GetLogger(ctx),
		"jobId", req.JobId,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
	)
	_ = logger

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

	builder := newInitStatementBuilder(
		a.sqlmanager,
		a.jobclient,
		a.connclient,
		a.eelicense,
		a.isNeosyncCloud,
		info.WorkflowExecution.ID,
	)
	slogger := neosynclogger.NewJsonSLogger().With(
		"jobId", req.JobId,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
	)
	return builder.RunSqlInitTableStatements(ctx, req, slogger)
}
