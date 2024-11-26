package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type generateAIBuilder struct {
	transformerclient  mgmtv1alpha1connect.TransformersServiceClient
	sqlmanagerclient   sqlmanager.SqlManagerClient
	connectionclient   mgmtv1alpha1connect.ConnectionServiceClient
	driver             string
	aiGroupedTableCols map[string][]string
}

func NewGenerateAIBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	connectionclient mgmtv1alpha1connect.ConnectionServiceClient,
	driver string,
) bb_internal.BenthosBuilder {
	return &generateAIBuilder{
		transformerclient:  transformerclient,
		sqlmanagerclient:   sqlmanagerclient,
		connectionclient:   connectionclient,
		driver:             driver,
		aiGroupedTableCols: map[string][]string{},
	}
}

type aiGenerateMappings struct {
	Schema  string
	Table   string
	Columns []*aiGenerateColumn
	Count   int
}
type aiGenerateColumn struct {
	Column   string
	DataType string
}

func (b *generateAIBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	jobSource := params.Job.GetSource()
	sourceOptions := jobSource.GetOptions().GetAiGenerate()
	if sourceOptions == nil {
		return nil, fmt.Errorf("job does not have AiGenerate source options, has: %T", jobSource.GetOptions().Config)
	}
	sourceConnection := params.SourceConnection

	openaiConfig := sourceConnection.GetConnectionConfig().GetOpenaiConfig()
	if openaiConfig == nil {
		return nil, errors.New("configured source connection is not an openai configuration")
	}
	constraintConnection, err := getConstraintConnection(ctx, jobSource, b.connectionclient, shared.GetConnectionById)
	if err != nil {
		return nil, err
	}
	db, err := b.sqlmanagerclient.NewSqlConnection(ctx, constraintConnection, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db().Close()

	groupedSchemas, err := db.Db().GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}

	mappings := []*aiGenerateMappings{}
	for _, schema := range sourceOptions.GetSchemas() {
		for _, table := range schema.GetTables() {
			columns := []*aiGenerateColumn{}

			tableColsMap, ok := groupedSchemas[sqlmanager_shared.BuildTable(schema.GetSchema(), table.GetTable())]
			if !ok {
				return nil, fmt.Errorf("did not find schema data when building AI Generate config: %s", schema.GetSchema())
			}
			for col, info := range tableColsMap {
				columns = append(columns, &aiGenerateColumn{
					Column:   col,
					DataType: info.DataType,
				})
			}

			mappings = append(mappings, &aiGenerateMappings{
				Schema:  schema.GetSchema(),
				Table:   table.GetTable(),
				Count:   int(table.GetRowCount()),
				Columns: columns,
			})
		}
	}
	if len(mappings) == 0 {
		return nil, fmt.Errorf("did not generate any mapping configs during AI Generate build for connection: %s", constraintConnection.GetId())
	}

	var userPrompt *string
	if sourceOptions.GetUserPrompt() != "" {
		up := sourceOptions.GetUserPrompt()
		userPrompt = &up
	}
	var userBatchSize *int
	if sourceOptions.GenerateBatchSize != nil && *sourceOptions.GenerateBatchSize > 0 {
		ubs := int(*sourceOptions.GenerateBatchSize)
		userBatchSize = &ubs
	}
	sourceResponses := buildBenthosAiGenerateSourceConfigResponses(
		openaiConfig,
		mappings,
		sourceOptions.GetModelName(),
		userPrompt,
		userBatchSize,
	)

	// builds a map of table key to columns for AI Generated schemas as they are calculated lazily instead of via job mappings
	aiGroupedTableCols := map[string][]string{}
	for _, agm := range mappings {
		key := neosync_benthos.BuildBenthosTable(agm.Schema, agm.Table)
		for _, col := range agm.Columns {
			aiGroupedTableCols[key] = append(aiGroupedTableCols[key], col.Column)
		}
	}
	b.aiGroupedTableCols = aiGroupedTableCols

	return sourceResponses, nil
}

func buildBenthosAiGenerateSourceConfigResponses(
	openaiconfig *mgmtv1alpha1.OpenAiConnectionConfig,
	mappings []*aiGenerateMappings,
	model string,
	userPrompt *string,
	userBatchSize *int,
) []*bb_internal.BenthosSourceConfig {
	responses := []*bb_internal.BenthosSourceConfig{}

	for _, tableMapping := range mappings {
		columns := []string{}
		dataTypes := []string{}
		for _, col := range tableMapping.Columns {
			columns = append(columns, col.Column)
			dataTypes = append(dataTypes, col.DataType)
		}
		batchSize := tableMapping.Count
		if tableMapping.Count > 100 {
			batchSize = 10
		}
		if userBatchSize != nil && *userBatchSize > 0 {
			batchSize = *userBatchSize
		}
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						OpenAiGenerate: &neosync_benthos.OpenAiGenerate{
							ApiUrl:     openaiconfig.ApiUrl,
							ApiKey:     openaiconfig.ApiKey,
							UserPrompt: userPrompt,
							Columns:    columns,
							DataTypes:  dataTypes,
							Model:      model,
							Count:      tableMapping.Count,
							BatchSize:  batchSize,
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads:    -1,
					Processors: []neosync_benthos.ProcessorConfig{},
				},
				Output: &neosync_benthos.OutputConfig{
					Outputs: neosync_benthos.Outputs{
						Broker: &neosync_benthos.OutputBrokerConfig{
							Pattern: "fan_out",
							Outputs: []neosync_benthos.Outputs{},
						},
					},
				},
			},
		}

		responses = append(responses, &bb_internal.BenthosSourceConfig{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "ai-generate"),
			},
		})
	}

	return responses
}

func (b *generateAIBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	benthosConfig := params.SourceConfig
	destOpts, err := getDestinationOptions(params.DestinationOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to parse destination options: %w", err)
	}
	tableKey := neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName)

	cols, ok := b.aiGroupedTableCols[tableKey]
	if !ok {
		return nil, fmt.Errorf("unable to find table columns for key (%s) when building destination connection", tableKey)
	}

	processorConfigs := []neosync_benthos.ProcessorConfig{}
	for _, pc := range benthosConfig.Processors {
		processorConfigs = append(processorConfigs, *pc)
	}

	config.BenthosDsns = append(config.BenthosDsns, &bb_shared.BenthosDsn{EnvVarKey: params.DestEnvVarKey, ConnectionId: params.DestConnection.Id})
	config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
		Fallback: []neosync_benthos.Outputs{
			{
				// retry processor and output several times
				Retry: &neosync_benthos.RetryConfig{
					InlineRetryConfig: neosync_benthos.InlineRetryConfig{
						MaxRetries: 10,
					},
					Output: neosync_benthos.OutputConfig{
						Outputs: neosync_benthos.Outputs{
							PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
								Driver: b.driver,
								Dsn:    params.DSN,

								Schema:              benthosConfig.TableSchema,
								Table:               benthosConfig.TableName,
								Columns:             cols,
								OnConflictDoNothing: destOpts.OnConflictDoNothing,
								TruncateOnRetry:     destOpts.Truncate,

								ArgsMapping: buildPlainInsertArgs(cols),

								Batching: &neosync_benthos.Batching{
									Period: destOpts.BatchPeriod,
									Count:  destOpts.BatchCount,
								},
								MaxInFlight: int(destOpts.MaxInFlight),
							},
						},
						Processors: processorConfigs,
					},
				},
			},
			// kills activity depending on error
			{Error: &neosync_benthos.ErrorOutputConfig{
				ErrorMsg: `${! meta("fallback_error")}`,
				Batching: &neosync_benthos.Batching{
					Period: destOpts.BatchPeriod,
					Count:  destOpts.BatchCount,
				},
			}},
		},
	})

	return config, nil
}

func getConstraintConnection(
	ctx context.Context,
	jobSource *mgmtv1alpha1.JobSource,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	getConnectionById func(context.Context, mgmtv1alpha1connect.ConnectionServiceClient, string) (*mgmtv1alpha1.Connection, error),
) (*mgmtv1alpha1.Connection, error) {
	var connectionId string
	switch jsConfig := jobSource.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		connectionId = jsConfig.Generate.GetFkSourceConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		connectionId = jsConfig.AiGenerate.GetFkSourceConnectionId()
	default:
		return nil, fmt.Errorf("unsupported job source options type for constraint connection: %T", jsConfig)
	}
	connection, err := getConnectionById(ctx, connclient, connectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to get constraint connection by id (%s): %w", connectionId, err)
	}
	return connection, nil
}
