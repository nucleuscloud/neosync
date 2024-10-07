package posttablesync_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
	externalId := shared.GetPostTableSyncConfigExternalId(req.Name)
	loggerKeyVals := []any{
		"accountId", req.AccountId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
		"RunContextExternalId", externalId,
	}
	logger := log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
	)
	logger.Debug("running post table sync activity")
	slogger := neosynclogger.NewJsonSLogger().With(loggerKeyVals...)

	rcResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
		Id: &mgmtv1alpha1.RunContextKey{
			JobRunId:   activityInfo.WorkflowExecution.ID,
			ExternalId: externalId,
			AccountId:  req.AccountId,
		},
	}))
	if err != nil && runContextNotFound(err) {
		slogger.Info("no runcontext found. continuing")
		return nil, nil
	} else if err != nil && !runContextNotFound(err) {
		return nil, fmt.Errorf("unable to retrieve posttablesync runcontext for %s: %w", req.Name, err)
	}

	configBits := rcResp.Msg.GetValue()
	if len(configBits) == 0 {
		slogger.Warn("post table sync value is empty")
		return &RunPostTableSyncResponse{}, nil
	}

	var config *shared.PostTableSyncConfig
	err = json.Unmarshal(configBits, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal posttablesync runcontext for %s: %w", req.Name, err)
	}

	if len(config.DestinationConfigs) == 0 {
		slogger.Debug("post table sync destination configs empty")
		return &RunPostTableSyncResponse{}, nil
	}

	for destConnectionId, destCfg := range config.DestinationConfigs {
		slogger.Debug(fmt.Sprintf("found %d post table sync statements", len(destCfg.Statements)), "destinationConnectionId", destConnectionId)
		if len(destCfg.Statements) == 0 {
			continue
		}
		destinationConnection, err := shared.GetConnectionById(ctx, a.connclient, destConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destConnectionId, err)
		}
		switch destinationConnection.GetConnectionConfig().GetConfig().(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
			destDb, err := a.sqlmanagerclient.NewPooledSqlDb(ctx, slogger, destinationConnection)
			if err != nil {
				destDb.Db.Close()
				slogger.Error("unable to connection to destination", "connectionId", destConnectionId)
				continue
			}
			err = destDb.Db.BatchExec(ctx, 5, destCfg.Statements, &sqlmanager_shared.BatchExecOpts{})
			if err != nil {
				slogger.Error("unable to exec destination statement", "connectionId", destConnectionId, "error", err.Error())
				continue
			}
		default:
			slogger.Warn("unsupported destination type", "connectionId", destConnectionId)
		}
	}

	return &RunPostTableSyncResponse{}, nil
}

func runContextNotFound(err error) bool {
	return strings.Contains(err.Error(), "no run context exists")
}
