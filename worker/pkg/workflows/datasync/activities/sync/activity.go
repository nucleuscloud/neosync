package sync_activity

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/warpstreamlabs/bento/public/bloblang"
	_ "github.com/warpstreamlabs/bento/public/components/gcp"
	_ "github.com/warpstreamlabs/bento/public/components/io"

	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/javascript"
	_ "github.com/warpstreamlabs/bento/public/components/aws"
	_ "github.com/warpstreamlabs/bento/public/components/mongodb"
	_ "github.com/warpstreamlabs/bento/public/components/pure"
	_ "github.com/warpstreamlabs/bento/public/components/pure/extended"
	_ "github.com/warpstreamlabs/bento/public/components/redis"

	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	pool_mongo_provider "github.com/nucleuscloud/neosync/internal/connection-manager/pool/providers/mongo"
	pool_sql_provider "github.com/nucleuscloud/neosync/internal/connection-manager/pool/providers/sql"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
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
	// Deprecated
	BenthosConfig string
	BenthosDsns   []*benthosbuilder_shared.BenthosDsn
	// Identifier that is used in combination with the AccountId to retrieve the benthos config
	Name      string
	AccountId string
}
type SyncResponse struct {
	Schema string
	Table  string
}

func New(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	sqlconnector sqlconnect.SqlConnector,
	tunnelmanagermap *sync.Map,
	temporalclient client.Client,
	meter metric.Meter,
	benthosStreamManager BenthosStreamManagerClient,
	disableReaper bool,
) *Activity {
	return &Activity{
		connclient:           connclient,
		jobclient:            jobclient,
		sqlconnector:         sqlconnector,
		tunnelmanagermap:     tunnelmanagermap,
		temporalclient:       temporalclient,
		meter:                meter,
		benthosStreamManager: benthosStreamManager,
		disableReaper:        disableReaper,
	}
}

type Activity struct {
	sqlconnector         sqlconnect.SqlConnector
	connclient           mgmtv1alpha1connect.ConnectionServiceClient
	jobclient            mgmtv1alpha1connect.JobServiceClient
	tunnelmanagermap     *sync.Map
	temporalclient       client.Client
	meter                metric.Meter // optional
	benthosStreamManager BenthosStreamManagerClient
	disableReaper        bool
}

func (a *Activity) getTunnelManagerByRunId(wfId, runId string) (connectionmanager.Interface[any], error) {
	connectionProvider := providers.NewProvider(
		mongoprovider.NewProvider(),
		sqlprovider.NewProvider(a.sqlconnector),
	)
	val, loaded := a.tunnelmanagermap.LoadOrStore(runId, connectionmanager.NewConnectionManager[any](connectionProvider))
	manager, ok := val.(connectionmanager.Interface[any])
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

var (
	// Hack that locks the instanced bento stream builder build step that causes data races if done in parallel
	streamBuilderMu sync.Mutex
)

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
	if req.AccountId != "" {
		loggerKeyVals = append(loggerKeyVals, "accountId", req.AccountId)
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

	var benthosConfig string
	if req.AccountId != "" && req.Name != "" {
		rcResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				JobRunId:   info.WorkflowExecution.ID,
				ExternalId: shared.GetBenthosConfigExternalId(req.Name),
				AccountId:  req.AccountId,
			},
		}))
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve benthosconfig runcontext for %s.%s: %w", metadata.Schema, metadata.Table, err)
		}
		benthosConfig = string(rcResp.Msg.GetValue())
	} else if req.BenthosConfig != "" {
		benthosConfig = req.BenthosConfig
	} else {
		return nil, fmt.Errorf("must provide means to retrieve benthos config either directly or via runcontext")
	}

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
			envKeyDsnSyncMap.Store(bdns.EnvVarKey, connection.Id)
			dsnToConnectionIdMap.Store(connection.Id, connection.Id)
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, fmt.Errorf("was unable to build connection details for some or all connections: %w", err)
	}

	benenv, err := benthos_environment.NewEnvironment(
		slogger,
		benthos_environment.WithMeter(a.meter),
		benthos_environment.WithSqlConfig(&benthos_environment.SqlConfig{
			Provider: pool_sql_provider.NewProvider(pool_sql_provider.GetSqlPoolProviderGetter(tunnelmanager, &dsnToConnectionIdMap, connectionMap, session, slogger)),
			IsRetry:  isRetry,
		}),
		benthos_environment.WithMongoConfig(&benthos_environment.MongoConfig{
			Provider: pool_mongo_provider.NewProvider(pool_mongo_provider.GetMongoPoolProviderGetter(tunnelmanager, &dsnToConnectionIdMap, connectionMap, session, slogger)),
		}),
		benthos_environment.WithStopChannel(stopActivityChan),
		benthos_environment.WithBlobEnv(bloblang.NewEnvironment()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate benthos environment: %w", err)
	}

	envKeyMap := syncMapToStringMap(&envKeyDsnSyncMap)
	envKeyMap[metrics.TemporalWorkflowIdEnvKey] = info.WorkflowExecution.ID
	envKeyMap[metrics.TemporalRunIdEnvKey] = info.WorkflowExecution.RunID
	envKeyMap[metrics.NeosyncDateEnvKey] = time.Now().UTC().Format(metrics.NeosyncDateFormat)

	streamBuilderMu.Lock()
	streambldr := benenv.NewStreamBuilder()
	// would ideally use the activity logger here but can't convert it into a slog.
	streambldr.SetLogger(slogger.With(
		"benthos", "true",
	))

	// This must come before SetYaml as otherwise it will not be invoked
	streambldr.SetEnvVarLookupFunc(getEnvVarLookupFn(envKeyMap))

	err = streambldr.SetYAML(benthosConfig)
	if err != nil {
		streamBuilderMu.Unlock()
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := a.benthosStreamManager.NewBenthosStreamFromBuilder(streambldr)
	streamBuilderMu.Unlock()
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
	return &SyncResponse{Schema: metadata.Schema, Table: metadata.Table}, nil
}

func getConnectionsFromBenthosDsns(
	ctx context.Context,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	dsns []*benthosbuilder_shared.BenthosDsn,
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
