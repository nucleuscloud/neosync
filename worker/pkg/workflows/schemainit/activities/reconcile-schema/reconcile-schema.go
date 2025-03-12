package reconcileschema_activity


package initschema_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	schemamanager "github.com/nucleuscloud/neosync/internal/schema-manager"
	schemamanager_shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type reconcileSchemaBuilder struct {
	sqlmanager sql_manager.SqlManagerClient
	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient
	eelicense  license.EEInterface
	workflowId string
}

func newReconcileSchemaBuilder(
	sqlmanagerclient sql_manager.SqlManagerClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	eelicense license.EEInterface,
	workflowId string,
) *reconcileSchemaBuilder {
	return &reconcileSchemaBuilder{
		sqlmanager: sqlmanagerclient,
		jobclient:  jobclient,
		connclient: connclient,
		eelicense:  eelicense,
		workflowId: workflowId,
	}
}

func (b *reconcileSchemaBuilder) RunReconcileSchema(
	ctx context.Context,
	req *RunReconcileSchemaRequest,
	session connectionmanager.SessionInterface,
	slogger *slog.Logger,
) (*RunReconcileSchemaResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}

	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	sourceConnectionType := shared.GetConnectionType(sourceConnection)
	slogger = slogger.With(
		"sourceConnectionType", sourceConnectionType,
	)

	if job.GetSource().GetOptions().GetAiGenerate() != nil {
		sourceConnection, err = shared.GetConnectionById(ctx, b.connclient, *job.GetSource().GetOptions().GetAiGenerate().FkSourceConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by id: %w", err)
		}
	}

	if sourceConnection.GetConnectionConfig().GetMongoConfig() != nil || sourceConnection.GetConnectionConfig().GetDynamodbConfig() != nil {
		return &RunSqlInitTableStatementsResponse{}, nil
	}

	var destination *mgmtv1alpha1.JobDestination
	for _, d := range job.Destinations {
		if d.Id == req.DestinationId {
			destination = d
			break
		}
	}
	if destination == nil {
		return nil, fmt.Errorf("unable to find destination by id (%s)", req.DestinationId)
	}

	uniqueTables := shared.GetUniqueTablesMapFromJob(job)
	uniqueSchemas := shared.GetUniqueSchemasFromJob(job)

	reconcileSchemaRunContext := []*ReconcileSchemaRunContext{}

	destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to get destination connection by id (%s): %w", destination.ConnectionId, err)
	}
	destinationConnectionType := shared.GetConnectionType(destinationConnection)
	slogger = slogger.With(
		"destinationConnectionType", destinationConnectionType,
	)

	shouldInitSchema := true
	if job.GetSource().GetOptions().GetAiGenerate() != nil {
		fkSrcConnId := job.GetSource().GetOptions().GetAiGenerate().GetFkSourceConnectionId()
		if fkSrcConnId == destination.GetConnectionId() {
			slogger.Warn("cannot init schema when destination connection is the same as the foreign key source connection")
			shouldInitSchema = false
		}
	}

	if job.GetSource().GetOptions().GetGenerate() != nil {
		fkSrcConnId := job.GetSource().GetOptions().GetGenerate().GetFkSourceConnectionId()
		if fkSrcConnId == destination.GetConnectionId() {
			slogger.Warn("cannot init schema when destination connection is the same as the foreign key source connection")
			shouldInitSchema = false
		}
	}

	manager := schemamanager.NewSchemaManager(b.sqlmanager, session, slogger, b.eelicense)
	schemaManager, err := manager.New(ctx, sourceConnection, destinationConnection, destination)
	if err != nil {
		return nil, fmt.Errorf("unable to create new schema manager: %w", err)
	}

	if shouldInitSchema {
		reconcileSchemaErrors, err := schemaManager.ReconcileDestinationSchema(ctx, uniqueTables, schemaStatements)
		if err != nil {
			return nil, fmt.Errorf("unable to reconcile schema: %w", err)
		}

		reconcileSchemaRunContext = append(reconcileSchemaRunContext, &ReconcileSchemaRunContext{
			ConnectionId: destination.GetConnectionId(),
			Errors:       reconcileSchemaErrors,
		})

		err = b.setReconcileSchemaRunCtx(ctx, reconcileSchemaRunContext, job.AccountId)
		if err != nil {
			return nil, err
		}
	}

	err = schemaManager.TruncateData(ctx, uniqueTables, uniqueSchemas)
	if err != nil {
		return nil, fmt.Errorf("unable to truncate data: %w", err)
	}

	schemaManager.CloseConnections()

	return &RunSqlInitTableStatementsResponse{}, nil
}

type ReconcileSchemaRunContext struct {
	ConnectionId string
	Errors       []*schemamanager_shared.InitSchemaError
}

func (b *reconcileSchemaBuilder) setReconcileSchemaRunCtx(
	ctx context.Context,
	reconcileSchemaRunContexts []*ReconcileSchemaRunContext,
	accountId string,
) error {
	bits, err := json.Marshal(reconcileSchemaRunContexts)
	if err != nil {
		return fmt.Errorf("failed to marshal reconcile schema run context: %w", err)
	}
	_, err = b.jobclient.SetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.SetRunContextRequest{
		Id: &mgmtv1alpha1.RunContextKey{
			JobRunId:   b.workflowId,
			ExternalId: "init-schema-report",
			AccountId:  accountId,
		},
		Value: bits,
	}))
	if err != nil {
		return fmt.Errorf("failed to set reconcile schema run context: %w", err)
	}
	return nil
}

func (b *reconcileSchemaBuilder) getJobById(
	ctx context.Context,
	jobId string,
) (*mgmtv1alpha1.Job, error) {
	getjobResp, err := b.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}
	return getjobResp.Msg.Job, nil
}
