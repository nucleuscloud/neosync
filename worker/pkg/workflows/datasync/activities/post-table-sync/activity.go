package posttablesync_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	jobclient        mgmtv1alpha1connect.JobServiceClient
	sqlmanagerclient sqlmanager.SqlManagerClient
	connclient       mgmtv1alpha1connect.ConnectionServiceClient
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
) *Activity {
	return &Activity{
		jobclient:        jobclient,
		sqlmanagerclient: sqlmanagerclient,
		connclient:       connclient,
	}
}

type RunPostTableSyncRequest struct {
	JobId string
	// Identifier that is used in combination with the AccountId to retrieve the post table run config
	Name      string
	AccountId string
}
type RunPostTableSyncResponse struct {
}

func (a *Activity) RunPostTableSync(
	ctx context.Context,
	req *RunPostTableSyncRequest,
) (*RunPostTableSyncResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	loggerKeyVals := []any{
		"jobId", req.JobId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
	}
	logger := log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
	)
	logger.Debug("running post table sync activity")
	slogger := neosynclogger.NewJsonSLogger().With(loggerKeyVals...)

	jobResp, err := a.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.JobId}))
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	job := jobResp.Msg.GetJob()

	switch job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres, *mgmtv1alpha1.JobSourceOptions_Mysql, *mgmtv1alpha1.JobSourceOptions_Mssql:
		err = a.runSqlPostTableSync(ctx, req, slogger, job, activityInfo.WorkflowExecution.ID)
		if err != nil {
			return nil, err
		}
	default:
	}

	return &RunPostTableSyncResponse{}, nil
}

func (a *Activity) runSqlPostTableSync(
	ctx context.Context,
	req *RunPostTableSyncRequest,
	slogger *slog.Logger,
	job *mgmtv1alpha1.Job,
	workflowId string,
) error {
	rcResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
		Id: &mgmtv1alpha1.RunContextKey{
			JobRunId:   workflowId,
			ExternalId: shared.GetPostTableSyncConfigExternalId(req.Name),
			AccountId:  req.AccountId,
		},
	}))
	if err != nil {
		return fmt.Errorf("unable to retrieve posttablesync runcontext for %s: %w", req.Name, err)
	}

	configBits := rcResp.Msg.GetValue()
	if len(configBits) == 0 {
		return nil
	}

	var config *shared.SqlPostTableSyncConfig
	err = json.Unmarshal(configBits, &config)
	if err != nil {
		return err
	}
	if len(config.DestinationStatements) == 0 {
		return nil
	}

	for _, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, a.connclient, destination.ConnectionId)
		if err != nil {
			return fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}
		destDb, err := a.sqlmanagerclient.NewPooledSqlDb(ctx, slogger, destinationConnection)
		if err != nil {
			destDb.Db.Close()
			slogger.Error("unable to connection to destination", "connectionId", destination.ConnectionId)
			continue
		}
		for _, stmt := range config.DestinationStatements {
			err := destDb.Db.Exec(ctx, stmt)
			if err != nil {
				slogger.Error("unable to exec destination statement", "connectionId", destination.ConnectionId, "error", err.Error())
				continue
			}
		}
	}
	return nil
}
