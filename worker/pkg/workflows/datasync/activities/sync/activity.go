package sync_activity

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redpanda-data/benthos/v4/public/bloblang"

	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"

	_ "github.com/nucleuscloud/neosync/internal/benthos/imports"

	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	pool_mongo_provider "github.com/nucleuscloud/neosync/internal/connection-manager/pool/providers/mongo"
	pool_sql_provider "github.com/nucleuscloud/neosync/internal/connection-manager/pool/providers/sql"
	benthos_environment "github.com/nucleuscloud/neosync/worker/pkg/benthos/environment"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/metric"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"

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
	sqlconnmanager connectionmanager.Interface[neosync_benthos_sql.SqlDbtx],
	mongoconnmanager connectionmanager.Interface[neosync_benthos_mongodb.MongoClient],
	meter metric.Meter,
	benthosStreamManager BenthosStreamManagerClient,
) *Activity {
	return &Activity{
		connclient:           connclient,
		jobclient:            jobclient,
		sqlconnmanager:       sqlconnmanager,
		mongoconnmanager:     mongoconnmanager,
		meter:                meter,
		benthosStreamManager: benthosStreamManager,
	}
}

type Activity struct {
	connclient           mgmtv1alpha1connect.ConnectionServiceClient
	jobclient            mgmtv1alpha1connect.JobServiceClient
	sqlconnmanager       connectionmanager.Interface[neosync_benthos_sql.SqlDbtx]
	mongoconnmanager     connectionmanager.Interface[neosync_benthos_mongodb.MongoClient]
	meter                metric.Meter // optional
	benthosStreamManager BenthosStreamManagerClient
}

var (
	// Hack that locks the instanced bento stream builder build step that causes data races if done in parallel
	streamBuilderMu sync.Mutex
)

// Temporal activity that runs benthos and syncs a source connection to one or more destination connections
func (a *Activity) Sync(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
	info := activity.GetInfo(ctx)
	session := connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(info.WorkflowExecution.ID))
	isRetry := info.Attempt > 1
	loggerKeyVals := []any{
		"metadata", metadata,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"activitySession", session.String(),
	}
	if req.AccountId != "" {
		loggerKeyVals = append(loggerKeyVals, "accountId", req.AccountId)
	}
	logger := log.With(activity.GetLogger(ctx), loggerKeyVals...)
	slogger := temporallogger.NewSlogger(logger)

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
			case <-ctx.Done():
				logger.Info("received context done, cleaning up...")
				resultChan <- fmt.Errorf("received context done signal")

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

	defer func() {
		logger.Debug("releasing session", "session", session.String())
		a.sqlconnmanager.ReleaseSession(session, slogger)
		a.mongoconnmanager.ReleaseSession(session, slogger)
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

	getConnectionById := getConnectionByIdFn(connectionMap)

	benenv, err := benthos_environment.NewEnvironment(
		slogger,
		benthos_environment.WithMeter(a.meter),
		benthos_environment.WithSqlConfig(&benthos_environment.SqlConfig{
			Provider: pool_sql_provider.NewConnectionProvider(a.sqlconnmanager, getConnectionById, session, slogger),
			IsRetry:  isRetry,
		}),
		benthos_environment.WithMongoConfig(&benthos_environment.MongoConfig{
			Provider: pool_mongo_provider.NewProvider(a.mongoconnmanager, getConnectionById, session, slogger),
		}),
		benthos_environment.WithStopChannel(stopActivityChan),
		benthos_environment.WithBlobEnv(bloblang.NewEnvironment()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate benthos environment: %w", err)
	}

	envKeyMap := map[string]string{}
	envKeyMap[metrics.TemporalWorkflowIdEnvKey] = info.WorkflowExecution.ID
	envKeyMap[metrics.TemporalRunIdEnvKey] = info.WorkflowExecution.RunID
	envKeyMap[metrics.NeosyncDateEnvKey] = time.Now().UTC().Format(metrics.NeosyncDateFormat)

	streamBuilderMu.Lock()
	streambldr := benenv.NewStreamBuilder()
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

func getConnectionByIdFn(connectionCache map[string]*mgmtv1alpha1.Connection) func(connectionId string) (connectionmanager.ConnectionInput, error) {
	return func(connectionId string) (connectionmanager.ConnectionInput, error) {
		connection, ok := connectionCache[connectionId]
		if !ok {
			return nil, fmt.Errorf("unable to find connection by id: %q", connectionId)
		}
		return connection, nil
	}
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
