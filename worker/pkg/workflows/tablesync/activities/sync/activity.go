package sync_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	benthosstream "github.com/nucleuscloud/neosync/internal/benthos-stream"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	pool_mongo_provider "github.com/nucleuscloud/neosync/internal/connection-manager/pool/providers/mongo"
	pool_sql_provider "github.com/nucleuscloud/neosync/internal/connection-manager/pool/providers/sql"
	continuation_token "github.com/nucleuscloud/neosync/internal/continuation-token"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	benthos_environment "github.com/nucleuscloud/neosync/worker/pkg/benthos/environment"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/redpanda-data/benthos/v4/public/service"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/activity"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"golang.org/x/sync/errgroup"

	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	_ "github.com/nucleuscloud/neosync/internal/benthos/imports"
)

type Activity struct {
	connclient           mgmtv1alpha1connect.ConnectionServiceClient
	jobclient            mgmtv1alpha1connect.JobServiceClient
	sqlconnmanager       connectionmanager.Interface[neosync_benthos_sql.SqlDbtx]
	mongoconnmanager     connectionmanager.Interface[neosync_benthos_mongodb.MongoClient]
	meter                metric.Meter // optional
	benthosStreamManager benthosstream.BenthosStreamManagerClient
	temporalclient       temporalclient.Client
}

func New(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	sqlconnmanager connectionmanager.Interface[neosync_benthos_sql.SqlDbtx],
	mongoconnmanager connectionmanager.Interface[neosync_benthos_mongodb.MongoClient],
	meter metric.Meter,
	benthosStreamManager benthosstream.BenthosStreamManagerClient,
	temporalclient temporalclient.Client,
) *Activity {
	return &Activity{
		connclient:           connclient,
		jobclient:            jobclient,
		sqlconnmanager:       sqlconnmanager,
		mongoconnmanager:     mongoconnmanager,
		meter:                meter,
		benthosStreamManager: benthosStreamManager,
		temporalclient:       temporalclient,
	}
}

type SyncTableRequest struct {
	Id        string
	AccountId string
	JobRunId  string

	ContinuationToken *string
}

type SyncTableResponse struct {
	ContinuationToken *string
}

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

// Deprecated
func (a *Activity) Sync(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
	info := activity.GetInfo(ctx)

	_, err := a.SyncTable(ctx, &SyncTableRequest{
		Id:        req.Name,
		AccountId: req.AccountId,
		JobRunId:  info.WorkflowExecution.ID,
	}, metadata)
	if err != nil {
		return nil, err
	}

	return &SyncResponse{
		Schema: metadata.Schema,
		Table:  metadata.Table,
	}, nil
}

func (a *Activity) SyncTable(ctx context.Context, req *SyncTableRequest, metadata *SyncMetadata) (*SyncTableResponse, error) {
	info := activity.GetInfo(ctx)

	session := connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(req.JobRunId))
	loggerKeyVals := []any{
		"metadata", metadata,
		"JobRunId", req.JobRunId,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"activitySession", session.String(),
		"accountId", req.AccountId,
		"hasContinuationToken", req.ContinuationToken != nil,
		"id", req.Id,
	}
	logger := temporallogger.NewSlogger(
		log.With(
			activity.GetLogger(ctx),
			loggerKeyVals...,
		),
	)

	stopActivityChan := make(chan error, 3)
	syncResultChan := make(chan error, 1)

	var benthosStream benthosstream.BenthosStreamClient

	go monitorActivityHeartbeat(ctx, stopActivityChan, func(logMessage string, err error) {
		handleStreamStop(benthosStream, syncResultChan, err, logMessage, logger)
	}, logger)

	benthosConfig, err := a.getBenthosConfig(ctx, &mgmtv1alpha1.RunContextKey{
		JobRunId:   req.JobRunId,
		ExternalId: shared.GetBenthosConfigExternalId(req.Id),
		AccountId:  req.AccountId,
	}, metadata)
	if err != nil {
		return nil, err
	}

	defer func() {
		logger.Debug("releasing session", "session", session.String())
		a.sqlconnmanager.ReleaseSession(session, logger)
		a.mongoconnmanager.ReleaseSession(session, logger)
	}()

	getConnectionById, err := a.getConnectionByIdFn(ctx, &mgmtv1alpha1.RunContextKey{
		JobRunId:   req.JobRunId,
		ExternalId: shared.GetConnectionIdsExternalId(),
		AccountId:  req.AccountId,
	}, metadata)
	if err != nil {
		return nil, err
	}

	var continuationTokenToReturn *string
	hasMorePages := func(lastReadOrderValues []any) {
		token := continuation_token.NewFromContents(continuation_token.NewContents(lastReadOrderValues))
		tokenStr := token.String()
		continuationTokenToReturn = &tokenStr
	}

	var continuationToken *continuation_token.ContinuationToken
	if req.ContinuationToken != nil {
		continuationToken, err = continuation_token.FromTokenString(*req.ContinuationToken)
		if err != nil {
			return nil, fmt.Errorf("unable to load continuation token: %w", err)
		}
	}

	identityAllocator := a.getIdentityAllocator(a.temporalclient, &info)

	bstream, err := a.getBenthosStream(
		&info,
		benthosConfig,
		session,
		stopActivityChan,
		getConnectionById,
		hasMorePages,
		continuationToken,
		identityAllocator,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get benthos stream: %w", err)
	}

	benthosStream = bstream

	go runStream(benthosStream, ctx, syncResultChan, logger)

	err = <-syncResultChan
	if err != nil {
		return nil, fmt.Errorf("could not successfully complete sync activity: %w", err)
	}

	logger.Info("sync complete")
	return &SyncTableResponse{
		ContinuationToken: continuationTokenToReturn,
	}, nil
}

func (a *Activity) getConnectionByIdFn(
	ctx context.Context,
	rcKey *mgmtv1alpha1.RunContextKey,
	metadata *SyncMetadata,
) (func(connectionId string) (connectionmanager.ConnectionInput, error), error) {
	connectionIds, err := a.getConnectionIds(ctx, rcKey, metadata)
	if err != nil {
		return nil, err
	}

	connections, err := a.getConnectionsFromConnectionIds(ctx, connectionIds)
	if err != nil {
		return nil, err
	}

	return getConnectionByIdFn(maps.Collect(getDtoSeq(connections))), nil
}

const (
	allocatorBlockSize = 1_000 // todo: should be the page limit
)

func (a *Activity) getIdentityAllocator(tclient temporalclient.Client, info *activity.Info) tablesync_shared.IdentityAllocator {
	blockAllocator := tablesync_shared.NewTemporalBlockAllocator(
		tclient,
		info.WorkflowExecution.ID,
		info.WorkflowExecution.RunID,
	)
	seed := info.StartedTime.UnixNano()
	return tablesync_shared.NewSingleIdentityAllocator(blockAllocator, allocatorBlockSize, rng.New(seed))
}

func monitorActivityHeartbeat(
	ctx context.Context,
	stopActivityChan <-chan error,
	handleStreamStop func(logMessage string, err error),
	logger *slog.Logger,
) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("recovered from panic in sync activity heartbeat loop: %v", r)
		}
	}()

	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case activityErr := <-stopActivityChan:
			handleStreamStop("received stop activity from benthos channel", activityErr)
			return

		case <-ctx.Done():
			handleStreamStop("received context done signal", ctx.Err())
			return

		case <-heartbeatTicker.C:
			activity.RecordHeartbeat(ctx)
		}
	}
}

func handleStreamStop(
	benthosStream benthosstream.BenthosStreamClient,
	syncResultChan chan<- error,
	err error,
	logMessage string,
	logger log.Logger,
) {
	logger.Info(logMessage + ", cleaning up...")
	syncResultChan <- err

	if benthosStream != nil {
		// Stop stream explicitly since stream.Run(ctx) doesn't fully obey canceled context when sink is in error state
		if stopErr := benthosStream.StopWithin(1 * time.Millisecond); stopErr != nil {
			logger.Error(stopErr.Error())
		}
	}
}

func runStream(
	stream benthosstream.BenthosStreamClient,
	ctx context.Context,
	syncResultChan chan<- error,
	logger *slog.Logger,
) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in benthos stream: %v", r)
			logger.Error("recovered from panic", "error", err)
			syncResultChan <- err // Send panic as error to channel
		}
	}()
	if err := stream.Run(ctx); err != nil {
		err = fmt.Errorf("unable to run benthos stream: %w", err)
		logger.Error("stream run failed", "error", err)
		syncResultChan <- err
		return
	}

	logger.Debug("stream completed successfully")
	syncResultChan <- nil
}

func (a *Activity) getBenthosStream(
	info *activity.Info,
	benthosConfig string,
	session connectionmanager.SessionInterface,
	stopActivityChan chan error,
	getConnectionById func(connectionId string) (connectionmanager.ConnectionInput, error),
	hasMorePages neosync_benthos_sql.OnHasMorePagesFn,
	continuationToken *continuation_token.ContinuationToken,
	identityAllocator tablesync_shared.IdentityAllocator,
	logger *slog.Logger,
) (benthosstream.BenthosStreamClient, error) {
	benenv, err := a.getBenthosEnvironment(
		logger,
		info.Attempt > 1,
		getConnectionById,
		session,
		stopActivityChan,
		hasMorePages,
		continuationToken,
		identityAllocator,
	)
	if err != nil {
		return nil, err
	}

	envKeyMap := map[string]string{}
	envKeyMap[metrics.TemporalWorkflowIdEnvKey] = info.WorkflowExecution.ID
	envKeyMap[metrics.TemporalRunIdEnvKey] = info.WorkflowExecution.RunID
	envKeyMap[metrics.NeosyncDateEnvKey] = time.Now().UTC().Format(metrics.NeosyncDateFormat)

	streambldr := benenv.NewStreamBuilder()
	streambldr.SetLogger(logger.With(
		"benthos", "true",
	))

	// This must come before SetYaml as otherwise it will not be invoked
	streambldr.SetEnvVarLookupFunc(getEnvVarLookupFn(envKeyMap))

	err = streambldr.SetYAML(benthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := a.benthosStreamManager.NewBenthosStreamFromBuilder(streambldr)
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos config: %w", err)
	}

	return stream, nil
}

func (a *Activity) getBenthosEnvironment(
	logger *slog.Logger,
	isRetry bool,
	getConnectionById func(connectionId string) (connectionmanager.ConnectionInput, error),
	session connectionmanager.SessionInterface,
	stopActivityChan chan error,
	hasMorePages neosync_benthos_sql.OnHasMorePagesFn,
	continuationToken *continuation_token.ContinuationToken,
	identityAllocator tablesync_shared.IdentityAllocator,
) (*service.Environment, error) {
	blobEnv := bloblang.NewEnvironment()
	err := transformers.RegisterTransformIdentityScramble(blobEnv, identityAllocator)
	if err != nil {
		return nil, fmt.Errorf("unable to register identity scramble transformer: %w", err)
	}
	benenv, err := benthos_environment.NewEnvironment(
		logger,
		benthos_environment.WithMeter(a.meter),
		benthos_environment.WithSqlConfig(&benthos_environment.SqlConfig{
			Provider:               pool_sql_provider.NewConnectionProvider(a.sqlconnmanager, getConnectionById, session, logger),
			IsRetry:                isRetry,
			InputHasMorePages:      hasMorePages,
			InputContinuationToken: continuationToken,
		}),
		benthos_environment.WithMongoConfig(&benthos_environment.MongoConfig{
			Provider: pool_mongo_provider.NewProvider(a.mongoconnmanager, getConnectionById, session, logger),
		}),
		benthos_environment.WithStopChannel(stopActivityChan),
		benthos_environment.WithBlobEnv(blobEnv),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate benthos environment: %w", err)
	}
	return benenv, nil
}

type DtoSeq interface {
	GetId() string
}

func getDtoSeq[T DtoSeq](dtos []T) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for _, dto := range dtos {
			if !yield(dto.GetId(), dto) {
				return
			}
		}
	}
}

func (a *Activity) getBenthosConfig(
	ctx context.Context,
	req *mgmtv1alpha1.RunContextKey,
	metadata *SyncMetadata,
) (string, error) {
	rcResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
		Id: req,
	}))
	if err != nil {
		return "", fmt.Errorf("unable to retrieve benthosconfig runcontext for %s.%s: %w", metadata.Schema, metadata.Table, err)
	}
	return string(rcResp.Msg.GetValue()), nil
}

func (a *Activity) getConnectionIds(
	ctx context.Context,
	req *mgmtv1alpha1.RunContextKey,
	metadata *SyncMetadata,
) ([]string, error) {
	rcResp, err := a.jobclient.GetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
		Id: req,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve connection ids runcontext for %s.%s: %w", metadata.Schema, metadata.Table, err)
	}
	var connectionIds []string
	err = json.Unmarshal(rcResp.Msg.GetValue(), &connectionIds)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal connection ids: %w", err)
	}
	return connectionIds, nil
}

func (a *Activity) getConnectionsFromConnectionIds(
	ctx context.Context,
	connectionIds []string,
) ([]*mgmtv1alpha1.Connection, error) {
	connections := make([]*mgmtv1alpha1.Connection, len(connectionIds))

	errgrp, errctx := errgroup.WithContext(ctx)
	for idx, connectionId := range connectionIds {
		idx := idx
		connectionId := connectionId
		errgrp.Go(func() error {
			resp, err := a.connclient.GetConnection(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: connectionId}))
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

func getConnectionByIdFn(connectionCache map[string]*mgmtv1alpha1.Connection) func(connectionId string) (connectionmanager.ConnectionInput, error) {
	return func(connectionId string) (connectionmanager.ConnectionInput, error) {
		connection, ok := connectionCache[connectionId]
		if !ok {
			return nil, fmt.Errorf("unable to find connection by id: %q", connectionId)
		}
		return connection, nil
	}
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
