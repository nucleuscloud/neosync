package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
)

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

func (b *benthosBuilder) getAiGenerateBenthosConfigResponses(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	slogger *slog.Logger,
) ([]*BenthosConfigResponse, []*aiGenerateMappings, error) {
	jobSource := job.GetSource()
	sourceOptions := job.GetSource().GetOptions().GetAiGenerate()
	if sourceOptions == nil {
		return nil, nil, fmt.Errorf("job does not have AiGenerate source options, has: %T", jobSource.GetOptions().Config)
	}
	sourceConnection, err := getJobSourceConnection(ctx, jobSource, b.getConnectionById)
	if err != nil {
		return nil, nil, err
	}
	openaiConfig := sourceConnection.GetConnectionConfig().GetOpenaiConfig()
	if openaiConfig == nil {
		return nil, nil, errors.New("configured source connection is not an openai configuration")
	}
	constraintConnection, err := getConstraintConnection(ctx, jobSource, b.getConnectionById)
	if err != nil {
		return nil, nil, err
	}
	db, err := b.sqladapter.NewSqlDb(ctx, slogger, constraintConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.ClosePool()

	dbschemas, err := db.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}
	groupedSchemas := sql_manager.GetUniqueSchemaColMappings(dbschemas)

	mappings := []*aiGenerateMappings{}
	for _, schema := range sourceOptions.GetSchemas() {
		for _, table := range schema.GetTables() {
			columns := []*aiGenerateColumn{}

			tableColsMap, ok := groupedSchemas[dbschemas_utils.BuildTable(schema.GetSchema(), table.GetTable())]
			if !ok {
				return nil, nil, fmt.Errorf("did not find schema data when building AI Generate config: %s", schema.GetSchema())
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
		return nil, nil, fmt.Errorf("did not generate any mapping configs during AI Generate build for connection: %s", constraintConnection.GetId())
	}

	var userPrompt *string
	if sourceOptions.GetUserPrompt() != "" {
		up := sourceOptions.GetUserPrompt()
		userPrompt = &up
	}
	sourceResponses := buildBenthosAiGenerateSourceConfigResponses(
		openaiConfig,
		mappings,
		sourceOptions.GetModelName(),
		userPrompt,
	)

	return sourceResponses, mappings, nil
}

func buildBenthosAiGenerateSourceConfigResponses(
	openaiconfig *mgmtv1alpha1.OpenAiConnectionConfig,
	mappings []*aiGenerateMappings,
	model string,
	userPrompt *string,
) []*BenthosConfigResponse {
	responses := []*BenthosConfigResponse{}

	for _, tableMapping := range mappings {
		columns := []string{}
		dataTypes := []string{}
		for _, col := range tableMapping.Columns {
			columns = append(columns, col.Column)
			dataTypes = append(dataTypes, col.DataType)
		}
		batchSize := tableMapping.Count
		if tableMapping.Count > 100 {
			batchSize = tableMapping.Count / 10
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

		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,

			// Processors: processors,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "ai-generate"),
			},
		})
	}

	return responses
}

func getConstraintConnection(
	ctx context.Context,
	jobSource *mgmtv1alpha1.JobSource,
	getConnectionById func(context.Context, string) (*mgmtv1alpha1.Connection, error),
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
	connection, err := getConnectionById(ctx, connectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to get constraint connection by id (%s): %w", connectionId, err)
	}
	return connection, nil
}
