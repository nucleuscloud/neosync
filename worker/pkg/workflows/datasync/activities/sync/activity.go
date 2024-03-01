package sync_activity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/javascript"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/redis"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	"github.com/google/uuid"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/redis"
	custom_sql "github.com/nucleuscloud/neosync/worker/internal/benthos/sql"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
	logger_utils "github.com/nucleuscloud/neosync/worker/internal/logger"

	"connectrpc.com/connect"
	"github.com/benthosdev/benthos/v4/public/service"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"golang.org/x/sync/errgroup"
)

type SyncMetadata struct {
	Schema string
	Table  string
}

type SyncRequest struct {
	BenthosConfig string
	BenthosDsns   []*shared.BenthosDsn
}
type SyncResponse struct{}

func New(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	tunnelmanagermap *sync.Map,
	temporalclient client.Client,
) *Activity {
	return &Activity{connclient: connclient, tunnelmanagermap: tunnelmanagermap, temporalclient: temporalclient}
}

type Activity struct {
	connclient       mgmtv1alpha1connect.ConnectionServiceClient
	tunnelmanagermap *sync.Map
	temporalclient   client.Client
}

func (a *Activity) getTunnelManagerByRunId(wfId, runId string) (*ConnectionTunnelManager, error) {
	val, loaded := a.tunnelmanagermap.LoadOrStore(runId, NewConnectionTunnelManager())
	manager, ok := val.(*ConnectionTunnelManager)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve connection tunnel manager from tunnel manager map. Expected *ConnectionTunnelManager, received: %T", manager)
	}
	if loaded {
		go manager.Reaper()
		go func() {
			// periodically waits for the workflow to complete so that it can shut down the tunnel manager for that run
			for {
				time.Sleep(1 * time.Minute)
				exec, err := a.temporalclient.DescribeWorkflowExecution(context.Background(), wfId, runId)
				if err != nil && (errors.Is(err, &serviceerror.Internal{}) || errors.Is(err, &serviceerror.Unavailable{})) {
					continue
				} else if (err != nil && errors.Is(err, &serviceerror.NotFound{})) {
					a.tunnelmanagermap.Delete(runId)
					go manager.Shutdown()
					return
				} else if err == nil && exec.GetWorkflowExecutionInfo().GetCloseTime() != nil {
					a.tunnelmanagermap.Delete(runId)
					go manager.Shutdown()
					return
				} else {
					continue
				}
			}
		}()
	}
	return manager, nil
}

// Temporal activity that runs benthos and syncs a source connection to one or more destination connections
func (a *Activity) Sync(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
	session := uuid.NewString()
	info := activity.GetInfo(ctx)
	loggerKeyVals := []any{
		"metadata", metadata,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"activitySession", session,
	}
	logger := log.With(activity.GetLogger(ctx), loggerKeyVals...)
	slogger := logger_utils.NewJsonSLogger().With(loggerKeyVals...)
	var benthosStream *service.Stream
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.Stop(ctx)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				return
			}
		}
	}()

	tunnelmanager, err := a.getTunnelManagerByRunId(info.WorkflowExecution.ID, info.WorkflowExecution.RunID)
	if err != nil {
		return nil, err
	}

	connections := make([]*mgmtv1alpha1.Connection, len(req.BenthosDsns))

	errgrp, errctx := errgroup.WithContext(ctx)
	for idx, bdns := range req.BenthosDsns {
		idx := idx
		bdns := bdns
		errgrp.Go(func() error {
			resp, err := a.connclient.GetConnection(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: bdns.ConnectionId}))
			if err != nil {
				return fmt.Errorf("failed to retrieve connection: %w", err)
			}
			connections[idx] = resp.Msg.Connection
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, fmt.Errorf("unable to retrieve all or some connections: %w", err)
	}

	connectionMap := map[string]*mgmtv1alpha1.Connection{}
	for _, connection := range connections {
		connectionMap[connection.Id] = connection
	}

	benthosenv := service.NewEnvironment()

	envKeyDsnSyncMap := sync.Map{}

	defer func() {
		tunnelmanager.ReleaseSession(session)
		tunnelmanager.LogStats(slogger, slog.LevelDebug)
	}()

	dsnToConnectionIdMap := map[string]string{}
	errgrp, errctx = errgroup.WithContext(ctx)
	for idx, bdns := range req.BenthosDsns {
		idx := idx
		bdns := bdns
		errgrp.Go(func() error {
			connection := connections[idx]
			// benthos raws will need to have a map of connetions due to there possibly being more than one connection per benthos run associated to the configs
			// so the raws need to have connections that will be good for every connection string it will encounter in a single run
			localConnStr, err := tunnelmanager.GetConnectionString(session, connection, slogger)
			if err != nil {
				return err
			}
			envKeyDsnSyncMap.Store(bdns.EnvVarKey, localConnStr)
			dsnToConnectionIdMap[localConnStr] = connection.Id
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, fmt.Errorf("was unable to build connection details for some or all connections: %w", err)
	}

	poolprovider := newPoolProvider(func(dsn string) (*sql.DB, error) {
		connid, ok := dsnToConnectionIdMap[dsn]
		if !ok {
			return nil, errors.New("unable to find connection id by dsn when getting db pool")
		}
		connection, ok := connectionMap[connid]
		if !ok {
			return nil, errors.New("unable to find connection by connection id when getting db pool")
		}
		return tunnelmanager.GetConnection(session, connection, slogger)
	})
	err = custom_sql.RegisterPooledSqlRawOutput(benthosenv, poolprovider)
	if err != nil {
		return nil, err
	}

	envKeyDnsMap := syncMapToStringMap(&envKeyDsnSyncMap)

	streambldr := benthosenv.NewStreamBuilder()
	// would ideally use the activity logger here but can't convert it into a slog.
	streambldr.SetLogger(slogger.With(
		"benthos", "true",
	))

	// This must come before SetYaml as otherwise it will not be invoked
	streambldr.SetEnvVarLookupFunc(getEnvVarLookupFn(envKeyDnsMap))

	err = streambldr.SetYAML(req.BenthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := streambldr.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos config: %w", err)
	}
	benthosStream = stream
	err = stream.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run benthos stream: %w", err)
	}
	benthosStream = nil
	logger.Info("sync complete")
	return &SyncResponse{}, nil
}

func getEnvVarLookupFn(input map[string]string) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		if input == nil {
			return "", false
		}
		output, ok := input[key]
		return output, ok
	}
}

func syncMapToStringMap(incoming *sync.Map) map[string]string {
	out := map[string]string{}
	if incoming == nil {
		return out
	}

	incoming.Range(func(key, value any) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		valStr, ok := value.(string)
		if !ok {
			return true
		}
		out[keyStr] = valStr
		return true
	})
	return out
}
