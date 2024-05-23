package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	generateDefault            = "generate_default"
	passthrough                = "passthrough"
	dbDefault                  = "DEFAULT"
	jobmappingSubsetErrMsg     = "job mappings are not equal to or a subset of the database schema found in the source connection"
	haltOnSchemaAdditionErrMsg = "job mappings does not contain a column mapping for all " +
		"columns found in the source connection for the selected schemas and tables"
)

type benthosBuilder struct {
	sqlmanager sql_manager.SqlManagerClient

	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	jobId string
	runId string

	redisConfig *shared.RedisConfig

	metricsEnabled bool
}

func newBenthosBuilder(
	sqlmanager sql_manager.SqlManagerClient,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,

	jobId, runId string,

	redisConfig *shared.RedisConfig,

	metricsEnabled bool,
) *benthosBuilder {
	return &benthosBuilder{
		sqlmanager:        sqlmanager,
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		jobId:             jobId,
		runId:             runId,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
	}
}

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	responses := []*BenthosConfigResponse{}

	// reverse of table dependency
	// map of foreign key to source table + column
	var primaryKeyToForeignKeysMap map[string]map[string][]*referenceKey            // schema.table -> column -> ForeignKey
	var colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer // schema.table -> column -> transformer
	var aiGroupedTableCols map[string][]string                                      // map of table key to columns for AI Generated schemas

	switch job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		sourceResponses, aimappings, err := b.getAiGenerateBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos AI Generate source config responses: %w", err)
		}
		aiGroupedTableCols = aimappings
		responses = append(responses, sourceResponses...)
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceResponses, err := b.getGenerateBenthosConfigResponses(ctx, job)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos Generate source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)
	case *mgmtv1alpha1.JobSourceOptions_Postgres, *mgmtv1alpha1.JobSourceOptions_Mysql:
		resp, err := b.getSqlSyncBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos sql sync source config responses: %w", err)
		}
		primaryKeyToForeignKeysMap = resp.primaryKeyToForeignKeysMap
		colTransformerMap = resp.ColumnTransformerMap
		responses = append(responses, resp.BenthosConfigs...)
	default:
		return nil, errors.New("unsupported job source")
	}

	for destIdx, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}
		for _, resp := range responses {
			dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", destIdx)
			dsn := fmt.Sprintf("${%s}", dstEnvVarKey)

			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				driver, err := getSqlDriverFromConnection(destinationConnection)
				if err != nil {
					return nil, err
				}
				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.Id})
				if resp.Config.Input.SqlSelect != nil || resp.Config.Input.PooledSqlRaw != nil {
					// SQL sync output
					outputs, err := b.getSqlSyncBenthosOutput(driver, destination, resp, dsn, primaryKeyToForeignKeysMap, colTransformerMap)
					if err != nil {
						return nil, err
					}
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
				} else if resp.Config.Input.Generate != nil {
					// SQL generate output
					outputs := b.getSqlGenerateOutput(driver, resp, destination, dsn)
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
				} else if resp.Config.Input.OpenAiGenerate != nil {
					// SQL AI generate output
					outputs, err := b.getSqlAiGenerateOutput(driver, resp, destination, dsn, aiGroupedTableCols)
					if err != nil {
						return nil, err
					}
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
				} else {
					return nil, errors.New("unable to build destination connection due to unsupported source connection")
				}
			case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
				if resp.RunType == tabledependency.RunTypeUpdate {
					continue
				}
				outputs := b.getAwsS3SyncBenthosOutput(connection, resp, req.WorkflowId)
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
			default:
				return nil, fmt.Errorf("unsupported destination connection config")
			}
		}
	}

	if b.metricsEnabled {
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, job.Id),
			metrics.NewEqLabel(metrics.TemporalWorkflowId, "${TEMPORAL_WORKFLOW_ID}"),
			metrics.NewEqLabel(metrics.TemporalRunId, "${TEMPORAL_RUN_ID}"),
		}
		for _, resp := range responses {
			joinedLabels := append(labels, resp.metriclabels...) //nolint:gocritic
			resp.Config.Metrics = &neosync_benthos.Metrics{
				OtelCollector: &neosync_benthos.MetricsOtelCollector{},
				Mapping:       joinedLabels.ToBenthosMeta(),
			}
		}
	}

	slogger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
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

func hasTransformer(t mgmtv1alpha1.TransformerSource) bool {
	return t != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED && t != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH
}

func groupGenerateSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.GenerateSourceSchemaOption,
) map[string]*generateSourceTableOptions {
	groupedMappings := map[string]*generateSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &generateSourceTableOptions{
				Count: int(tableOpt.RowCount), // todo: probably need to update rowcount int64 to int32
			}
		}
	}

	return groupedMappings
}

func getSqlDriverFromConnection(conn *mgmtv1alpha1.Connection) (string, error) {
	switch conn.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return sql_manager.PostgresDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return sql_manager.MysqlDriver, nil
	default:
		return "", fmt.Errorf("unsupported sql connection config")
	}
}

func groupJobSourceOptionsByTable(
	sqlSourceOpts *sqlJobSourceOpts,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}
	for _, schemaOpt := range sqlSourceOpts.SchemaOpt {
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sqlSourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}
	return groupedMappings
}

type tableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
}

func groupMappingsByTable(
	mappings []*mgmtv1alpha1.JobMapping,
) []*tableMapping {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}

	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		groupedMappings[key] = append(groupedMappings[key], mapping)
	}

	output := make([]*tableMapping, 0, len(groupedMappings))
	for key, mappings := range groupedMappings {
		schema, table := shared.SplitTableKey(key)
		output = append(output, &tableMapping{
			Schema:   schema,
			Table:    table,
			Mappings: mappings,
		})
	}
	return output
}

func getTableMappingsMap(groupedMappings []*tableMapping) map[string]*tableMapping {
	groupedTableMapping := map[string]*tableMapping{}
	for _, tm := range groupedMappings {
		groupedTableMapping[neosync_benthos.BuildBenthosTable(tm.Schema, tm.Table)] = tm
	}
	return groupedTableMapping
}

func getColumnTransformerMap(tableMappingMap map[string]*tableMapping) map[string]map[string]*mgmtv1alpha1.JobMappingTransformer {
	colTransformerMap := map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{} // schema.table ->  column -> transformer
	for table, mapping := range tableMappingMap {
		colTransformerMap[table] = map[string]*mgmtv1alpha1.JobMappingTransformer{}
		for _, m := range mapping.Mappings {
			colTransformerMap[table][m.Column] = m.Transformer
		}
	}
	return colTransformerMap
}
