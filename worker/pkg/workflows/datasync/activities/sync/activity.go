package sync_activity

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/gcp"
	_ "github.com/benthosdev/benthos/v4/public/components/io"

	_ "github.com/benthosdev/benthos/v4/public/components/mongodb"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/redis"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/javascript"

	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers/sqlprovider"
	benthos_environment "github.com/nucleuscloud/neosync/worker/pkg/benthos/environment"
	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/redis"
	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/metric"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"

	"github.com/google/uuid"
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
	meter metric.Meter,
	benthosStreamManager BenthosStreamManagerClient,
	disableReaper bool,
) *Activity {
	return &Activity{connclient: connclient, tunnelmanagermap: tunnelmanagermap, temporalclient: temporalclient, meter: meter, benthosStreamManager: benthosStreamManager, disableReaper: disableReaper}
}

type Activity struct {
	connclient           mgmtv1alpha1connect.ConnectionServiceClient
	tunnelmanagermap     *sync.Map
	temporalclient       client.Client
	meter                metric.Meter // optional
	benthosStreamManager BenthosStreamManagerClient
	disableReaper        bool
}

func (a *Activity) getTunnelManagerByRunId(wfId, runId string) (connectiontunnelmanager.Interface[any], error) {
	connectionProvider := providers.NewProvider(
		mongoprovider.NewProvider(),
		sqlprovider.NewProvider(),
	)
	val, loaded := a.tunnelmanagermap.LoadOrStore(runId, connectiontunnelmanager.NewConnectionTunnelManager[any, any](connectionProvider))
	manager, ok := val.(connectiontunnelmanager.Interface[any])
	if !ok {
		return nil, fmt.Errorf("unable to retrieve connection tunnel manager from tunnel manager map. Expected *ConnectionTunnelManager, received: %T", manager)
	}
	if a.disableReaper {
		return manager, nil
	}
	if !loaded {
		go manager.Reaper()
		go func() {
			// periodically waits for the workflow run to complete so that it can shut down the tunnel manager for that run
			for {
				time.Sleep(1 * time.Minute)
				exec, err := a.temporalclient.DescribeWorkflowExecution(context.Background(), wfId, runId)
				if (err != nil && errors.Is(err, &serviceerror.NotFound{})) || (err == nil && exec.GetWorkflowExecutionInfo().GetCloseTime() != nil) {
					a.tunnelmanagermap.Delete(runId)
					go manager.Shutdown()
					return
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
	isRetry := info.Attempt > 1
	loggerKeyVals := []any{
		"metadata", metadata,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"activitySession", session,
	}
	logger := log.With(activity.GetLogger(ctx), loggerKeyVals...)
	slogger := neosynclogger.NewJsonSLogger().With(loggerKeyVals...)
	stopActivityChan := make(chan error, 3)
	resultChan := make(chan error, 1)
	benthosStreamMutex := sync.Mutex{}
	var benthosStream BenthosStreamClient
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slogger.Error("recovered from panic in sync activity heartbeat loop: %v", r)
			}
		}()
		for {
			select {
			case activityErr := <-stopActivityChan:
				logger.Info("received stop activity from benthos channel, cleaning up...")
				resultChan <- activityErr
				benthosStreamMutex.Lock()
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.StopWithin(1 * time.Millisecond)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				benthosStreamMutex.Unlock()
				return
			case <-activity.GetWorkerStopChannel(ctx):
				logger.Info("received worker stop, cleaning up...")
				resultChan <- fmt.Errorf("received worker stop signal")
				benthosStreamMutex.Lock()
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.StopWithin(1 * time.Millisecond)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				benthosStreamMutex.Unlock()
				return
			case <-ctx.Done():
				logger.Info("received context done, cleaning up...")
				benthosStreamMutex.Lock()
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.StopWithin(1 * time.Millisecond)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				benthosStreamMutex.Unlock()
				return
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			}
		}
	}()

	tunnelmanager, err := a.getTunnelManagerByRunId(info.WorkflowExecution.ID, info.WorkflowExecution.RunID)
	if err != nil {
		return nil, err
	}
	defer func() {
		tunnelmanager.ReleaseSession(session)
	}()

	connections, err := getConnectionsFromBenthosDsns(ctx, a.connclient, req.BenthosDsns)
	if err != nil {
		return nil, err
	}
	connectionMap := map[string]*mgmtv1alpha1.Connection{}
	for _, connection := range connections {
		connectionMap[connection.Id] = connection
	}

	// todo: add support for gcp cloud storage authentication

	envKeyDsnSyncMap := sync.Map{}
	dsnToConnectionIdMap := sync.Map{}
	errgrp := errgroup.Group{}
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
			dsnToConnectionIdMap.Store(localConnStr, connection.Id)
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, fmt.Errorf("was unable to build connection details for some or all connections: %w", err)
	}

	benthosenv, err := benthos_environment.New(&benthos_environment.RegisterConfig{
		Meter: a.meter,
		SqlConfig: &benthos_environment.SqlConfig{
			Provider: newSqlPoolProvider(getSqlPoolProviderGetter(tunnelmanager, &dsnToConnectionIdMap, connectionMap, session, slogger)),
			IsRetry:  isRetry,
		},
		MongoConfig: &benthos_environment.MongoConfig{
			Provider: newMongoPoolProvider(getMongoPoolProviderGetter(tunnelmanager, &dsnToConnectionIdMap, connectionMap, session, slogger)),
		},
		StopChannel: stopActivityChan,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate benthos environment: %w", err)
	}

	envKeyMap := syncMapToStringMap(&envKeyDsnSyncMap)
	envKeyMap["TEMPORAL_WORKFLOW_ID"] = info.WorkflowExecution.ID
	envKeyMap["TEMPORAL_RUN_ID"] = info.WorkflowExecution.RunID

	streambldr := benthosenv.NewStreamBuilder()
	// would ideally use the activity logger here but can't convert it into a slog.
	streambldr.SetLogger(slogger.With(
		"benthos", "true",
	))

	// This must come before SetYaml as otherwise it will not be invoked
	streambldr.SetEnvVarLookupFunc(getEnvVarLookupFn(envKeyMap))

	err = streambldr.SetYAML(req.BenthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := a.benthosStreamManager.NewBenthosStreamFromBuilder(streambldr)
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos config: %w", err)
	}

	benthosStreamMutex.Lock()
	benthosStream = stream
	benthosStreamMutex.Unlock()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slogger.Error("recovered from panic in benthos stream run: %v", r)
			}
		}()
		err := stream.Run(ctx)
		if err != nil {
			resultChan <- fmt.Errorf("unable to run benthos stream: %w", err)
			return
		}
		resultChan <- nil
	}()

	err = <-resultChan
	if err != nil {
		return nil, fmt.Errorf("could not successfully complete sync activity: %w", err)
	}
	benthosStreamMutex.Lock()
	benthosStream = nil
	benthosStreamMutex.Unlock()

	logger.Info("sync complete")
	return &SyncResponse{}, nil
}

func getConnectionsFromBenthosDsns(
	ctx context.Context,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	dsns []*shared.BenthosDsn,
) ([]*mgmtv1alpha1.Connection, error) {
	connections := make([]*mgmtv1alpha1.Connection, len(dsns))

	errgrp, errctx := errgroup.WithContext(ctx)
	for idx, bdns := range dsns {
		idx := idx
		bdns := bdns
		errgrp.Go(func() error {
			resp, err := connclient.GetConnection(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: bdns.ConnectionId}))
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
	return connections, nil
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
