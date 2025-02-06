package jobhooks_by_timing_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type License interface {
	IsValid() bool
}

type Activity struct {
	jobclient        mgmtv1alpha1connect.JobServiceClient
	connclient       mgmtv1alpha1connect.ConnectionServiceClient
	sqlmanagerclient sqlmanager.SqlManagerClient
	license          License
}

func New(
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	license License,
) *Activity {
	return &Activity{jobclient: jobclient, connclient: connclient, sqlmanagerclient: sqlmanagerclient, license: license}
}

type RunJobHooksByTimingRequest struct {
	JobId  string
	Timing mgmtv1alpha1.GetActiveJobHooksByTimingRequest_Timing
}

type RunJobHooksByTimingResponse struct {
	ExecCount uint
}

// Runs active job hooks by the provided timing value
func (a *Activity) RunJobHooksByTiming(
	ctx context.Context,
	req *RunJobHooksByTimingRequest,
) (*RunJobHooksByTimingResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	timingName, ok := mgmtv1alpha1.GetActiveJobHooksByTimingRequest_Timing_name[int32(req.Timing)]
	if !ok {
		return nil, fmt.Errorf("timing was invalid and not resolvable: %d", req.Timing)
	}
	loggerKeyVals := []any{
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
		"jobId", req.JobId,
		"timing", timingName,
	}
	logger := log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
	)
	slogger := temporallogger.NewSlogger(logger)

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
	if !a.license.IsValid() {
		logger.Debug("skipping job hooks due to EE license not being active")
		return &RunJobHooksByTimingResponse{ExecCount: 0}, nil
	}

	logger.Debug(fmt.Sprintf("retrieving job hooks by timing %q", req.Timing))

	resp, err := a.jobclient.GetActiveJobHooksByTiming(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveJobHooksByTimingRequest{
		JobId:  req.JobId,
		Timing: req.Timing,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve active hooks by timing: %w", err)
	}
	hooks := resp.Msg.GetHooks()
	logger.Debug(fmt.Sprintf("found %d active hooks", len(hooks)))

	connections := make(map[string]*sqlmanager.SqlConnection)
	defer func() {
		for _, conn := range connections {
			conn.Db().Close()
		}
	}()

	session := connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(activityInfo.WorkflowExecution.ID))
	execCount := uint(0)

	for _, hook := range hooks {
		logger.Debug(fmt.Sprintf("running hook %q", hook.GetName()))
		logger := log.With(logger, "hookName", hook.GetName())

		switch hookConfig := hook.GetConfig().GetConfig().(type) {
		case *mgmtv1alpha1.JobHookConfig_Sql:
			logger.Debug("running SQL hook")
			if hookConfig.Sql == nil {
				return nil, errors.New("SQL hook config has undefined SQL configuration")
			}
			if err := a.executeSqlHook(
				ctx,
				hookConfig.Sql,
				a.getCachedConnectionFn(connections, session, logger, slogger),
			); err != nil {
				return nil, fmt.Errorf("unable to execute sql hook: %w", err)
			}
			execCount++
		default:
			logger.Warn(fmt.Sprintf("hook config with type %T is not currently supported!", hookConfig))
		}
	}

	return &RunJobHooksByTimingResponse{ExecCount: execCount}, nil
}

// Given a connection id, returns an initialized sql database connection
type getSqlDbFromConnectionId = func(ctx context.Context, connectionId string) (sqlmanager.SqlDatabase, error)

func (a *Activity) getCachedConnectionFn(
	connections map[string]*sqlmanager.SqlConnection,
	session connectionmanager.SessionInterface,
	logger log.Logger,
	slogger *slog.Logger,
) getSqlDbFromConnectionId {
	return func(ctx context.Context, connectionId string) (sqlmanager.SqlDatabase, error) {
		conn, ok := connections[connectionId]
		if ok {
			logger.Debug("found cached connection when running hook")
			return conn.Db(), nil
		}
		logger.Debug("initializing connection for hook")
		connectionResp, err := a.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: connectionId,
		}))
		if err != nil {
			return nil, err
		}
		connection := connectionResp.Msg.GetConnection()
		sqlconnection, err := a.sqlmanagerclient.NewSqlConnection(
			ctx,
			session,
			connection,
			slogger.With(
				"connectionId", connection.GetId(),
				"accountId", connection.GetAccountId(),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize pooled sql connection: %W", err)
		}
		connections[connectionId] = sqlconnection
		return sqlconnection.Db(), nil
	}
}

func (a *Activity) executeSqlHook(
	ctx context.Context,
	hook *mgmtv1alpha1.JobHookConfig_JobSqlHook,
	getSqlConnection getSqlDbFromConnectionId,
) error {
	db, err := getSqlConnection(ctx, hook.GetConnectionId())
	if err != nil {
		return err
	}
	if err := db.Exec(ctx, hook.GetQuery()); err != nil {
		return fmt.Errorf("unable to execute SQL hook statement: %w", err)
	}
	return nil
}
