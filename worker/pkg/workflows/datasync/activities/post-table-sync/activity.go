package posttablesync_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
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
	Errors []*PostTableSyncError
}

type PostTableSyncError struct {
	ConnectionId string
	Errors       []*StatementError
}

type StatementError struct {
	Statement string
	Error     string
}

func (a *Activity) RunPostTableSync(
	ctx context.Context,
	req *RunPostTableSyncRequest,
) (*RunPostTableSyncResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	session := connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(activityInfo.WorkflowExecution.ID))
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
	slogger := temporallogger.NewSlogger(logger)

	rcResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
		Id: &mgmtv1alpha1.RunContextKey{
			JobRunId:   activityInfo.WorkflowExecution.ID,
			ExternalId: externalId,
			AccountId:  req.AccountId,
		},
	}))
	if err != nil && runContextNotFound(err) {
		slogger.Debug("no runcontext found. continuing")
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

	destconns := []*sqlmanager.SqlConnection{}
	defer func() {
		for _, conn := range destconns {
			conn.Db().Close()
		}
	}()

	errors := []*PostTableSyncError{}
	for destConnectionId, destCfg := range config.DestinationConfigs {
		slogger.Debug(fmt.Sprintf("found %d post table sync statements", len(destCfg.Statements)), "destinationConnectionId", destConnectionId)
		if len(destCfg.Statements) == 0 {
			continue
		}
		destinationConnection, err := shared.GetConnectionById(ctx, a.connclient, destConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destConnectionId, err)
		}
		execErrors := &PostTableSyncError{
			ConnectionId: destConnectionId,
		}
		switch destinationConnection.GetConnectionConfig().GetConfig().(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
			destDb, err := a.sqlmanagerclient.NewSqlConnection(ctx, session, destinationConnection, slogger)
			if err != nil {
				slogger.Error("unable to connection to destination", "connectionId", destConnectionId)
				continue
			}
			destconns = append(destconns, destDb)
			for _, stmt := range destCfg.Statements {
				err = destDb.Db().Exec(ctx, stmt)
				if err != nil {
					slogger.Error("unable to exec destination statement", "connectionId", destConnectionId, "error", err.Error())
					execErrors.Errors = append(execErrors.Errors, &StatementError{
						Statement: stmt,
						Error:     err.Error(),
					})
				}
			}
		default:
			slogger.Warn("unsupported destination type", "connectionId", destConnectionId)
		}
		if len(execErrors.Errors) > 0 {
			errors = append(errors, execErrors)
		}
	}

	return &RunPostTableSyncResponse{
		Errors: errors,
	}, nil
}

func runContextNotFound(err error) bool {
	connectErr, ok := err.(*connect.Error)
	if ok && connectErr.Code() == connect.CodeNotFound {
		return true
	}
	return strings.Contains(err.Error(), "unable to find key") || strings.Contains(err.Error(), "no run context exists with the provided key")
}
