package genbenthosconfigs_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	querybuilder2 "github.com/nucleuscloud/neosync/worker/pkg/query-builder2"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	"gopkg.in/yaml.v3"
)

type benthosBuilder struct {
	sqlmanagerclient sqlmanager.SqlManagerClient

	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	jobId      string
	workflowId string
	runId      string

	redisConfig *shared.RedisConfig

	metricsEnabled bool
}

func newBenthosBuilder(
	sqlmanagerclient sqlmanager.SqlManagerClient,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,

	jobId, workflowId string, runId string,

	redisConfig *shared.RedisConfig,

	metricsEnabled bool,
) *benthosBuilder {
	return &benthosBuilder{
		sqlmanagerclient:  sqlmanagerclient,
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		jobId:             jobId,
		workflowId:        workflowId,
		runId:             runId,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
	}
}

type workflowMetadata struct {
	WorkflowId string
	RunId      string
}

func (b *benthosBuilder) GenerateBenthosConfigsNew(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	wfmetadata *workflowMetadata,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	destConnections := []*mgmtv1alpha1.Connection{}
	for _, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}
		destConnections = append(destConnections, destinationConnection)
	}

	benthosManagerConfig := &benthosbuilder.WorkerBenthosConfig{
		Job:                    job,
		SourceConnection:       sourceConnection,
		DestinationConnections: destConnections,
		WorkflowId:             wfmetadata.WorkflowId,
		Logger:                 slogger,
		Sqlmanagerclient:       b.sqlmanagerclient,
		Transformerclient:      b.transformerclient,
		Connectionclient:       b.connclient,
		RedisConfig:            b.redisConfig,
		SelectQueryBuilder:     &querybuilder2.QueryMapBuilderWrapper{},
		MetricsEnabled:         b.metricsEnabled,
		MetricLabelKeyVals: map[string]string{
			metrics.TemporalWorkflowId: bb_shared.WithEnvInterpolation(metrics.TemporalWorkflowIdEnvKey),
			metrics.TemporalRunId:      bb_shared.WithEnvInterpolation(metrics.TemporalRunIdEnvKey),
		},
	}
	benthosManager, err := benthosbuilder.NewWorkerBenthosConfigManager(benthosManagerConfig)
	if err != nil {
		return nil, err
	}
	responses, err := benthosManager.GenerateBenthosConfigs(ctx)
	if err != nil {
		return nil, err
	}

	// TODO move run context logic into benthos builder
	postTableSyncRunCtx := buildPostTableSyncRunCtx(responses, job.Destinations)
	err = b.setPostTableSyncRunCtx(ctx, postTableSyncRunCtx, job.GetAccountId())
	if err != nil {
		return nil, fmt.Errorf("unable to set all run contexts for post table sync configs: %w", err)
	}

	outputConfigs, err := b.setRunContexts(ctx, responses, job.GetAccountId())
	if err != nil {
		return nil, fmt.Errorf("unable to set all run contexts for benthos configs: %w", err)
	}
	return &GenerateBenthosConfigsResponse{
		AccountId:      job.AccountId,
		BenthosConfigs: outputConfigs,
	}, nil
}

// this method modifies the input responses by nilling out the benthos config. it returns the same slice for convenience
func (b *benthosBuilder) setRunContexts(
	ctx context.Context,
	responses []*benthosbuilder.BenthosConfigResponse,
	accountId string,
) ([]*benthosbuilder.BenthosConfigResponse, error) {
	rcstream := b.jobclient.SetRunContexts(ctx)

	for _, config := range responses {
		bits, err := yaml.Marshal(config.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal benthos config: %w", err)
		}
		err = rcstream.Send(&mgmtv1alpha1.SetRunContextsRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				JobRunId:   b.workflowId,
				ExternalId: shared.GetBenthosConfigExternalId(config.Name),
				AccountId:  accountId,
			},
			Value: bits,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send run context: %w", err)
		}
		config.Config = nil // nilling this out so that it does not persist in temporal
	}

	_, err := rcstream.CloseAndReceive()
	if err != nil {
		return nil, fmt.Errorf("unable to receive response from benthos runcontext request: %w", err)
	}
	return responses, nil
}

func (b *benthosBuilder) setPostTableSyncRunCtx(
	ctx context.Context,
	postSyncConfigs map[string]*shared.PostTableSyncConfig,
	accountId string,
) error {
	rcstream := b.jobclient.SetRunContexts(ctx)

	for name, config := range postSyncConfigs {
		bits, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal post table sync config: %w", err)
		}
		err = rcstream.Send(&mgmtv1alpha1.SetRunContextsRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				JobRunId:   b.workflowId,
				ExternalId: shared.GetPostTableSyncConfigExternalId(name),
				AccountId:  accountId,
			},
			Value: bits,
		})
		if err != nil {
			return fmt.Errorf("failed to send post table sync run context: %w", err)
		}
	}

	_, err := rcstream.CloseAndReceive()
	if err != nil {
		return fmt.Errorf("unable to receive response from post table sync runcontext request: %w", err)
	}
	return nil
}

func (b *benthosBuilder) getJobById(
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

func buildPostTableSyncRunCtx(benthosConfigs []*benthosbuilder.BenthosConfigResponse, destinations []*mgmtv1alpha1.JobDestination) map[string]*shared.PostTableSyncConfig {
	postTableSyncRunCtx := map[string]*shared.PostTableSyncConfig{} // benthos_config_name -> config
	for _, bc := range benthosConfigs {
		destConfigs := map[string]*shared.PostTableSyncDestConfig{}
		for _, destination := range destinations {
			var stmts []string
			switch destination.GetOptions().GetConfig().(type) {
			case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
				stmts = buildPgPostTableSyncStatement(bc)
			case *mgmtv1alpha1.JobDestinationOptions_MssqlOptions:
				stmts = buildMssqlPostTableSyncStatement(bc)
			}
			if len(stmts) != 0 {
				destConfigs[destination.GetConnectionId()] = &shared.PostTableSyncDestConfig{
					Statements: stmts,
				}
			}
		}
		if len(destConfigs) != 0 {
			postTableSyncRunCtx[bc.Name] = &shared.PostTableSyncConfig{
				DestinationConfigs: destConfigs,
			}
		}
	}
	return postTableSyncRunCtx
}

func buildPgPostTableSyncStatement(bc *benthosbuilder.BenthosConfigResponse) []string {
	statements := []string{}
	if bc.RunType == tabledependency.RunTypeUpdate {
		return statements
	}
	colDefaultProps := bc.ColumnDefaultProperties
	for colName, p := range colDefaultProps {
		if p.NeedsReset && !p.HasDefaultTransformer {
			// resets sequences and identities
			resetSql := sqlmanager_postgres.BuildPgIdentityColumnResetCurrentSql(bc.TableSchema, bc.TableName, colName)
			statements = append(statements, resetSql)
		}
	}
	return statements
}

func buildMssqlPostTableSyncStatement(bc *benthosbuilder.BenthosConfigResponse) []string {
	statements := []string{}
	if bc.RunType == tabledependency.RunTypeUpdate {
		return statements
	}
	colDefaultProps := bc.ColumnDefaultProperties
	for _, p := range colDefaultProps {
		if p.NeedsOverride {
			// reset identity
			resetSql := sqlmanager_mssql.BuildMssqlIdentityColumnResetCurrent(bc.TableSchema, bc.TableName)
			statements = append(statements, resetSql)
		}
	}
	return statements
}
