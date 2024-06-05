package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
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
) ([]*BenthosConfigResponse, map[string][]string, error) {
	jobSource := job.GetSource()
	sourceOptions := job.GetSource().GetOptions().GetAiGenerate()
	if sourceOptions == nil {
		return nil, nil, fmt.Errorf("job does not have AiGenerate source options, has: %T", jobSource.GetOptions().Config)
	}
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, nil, err
	}
	openaiConfig := sourceConnection.GetConnectionConfig().GetOpenaiConfig()
	if openaiConfig == nil {
		return nil, nil, errors.New("configured source connection is not an openai configuration")
	}
	constraintConnection, err := getConstraintConnection(ctx, jobSource, b.connclient, shared.GetConnectionById)
	if err != nil {
		return nil, nil, err
	}
	db, err := b.sqlmanagerclient.NewPooledSqlDb(ctx, slogger, constraintConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db.Close()

	groupedSchemas, err := db.Db.GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}

	mappings := []*aiGenerateMappings{}
	for _, schema := range sourceOptions.GetSchemas() {
		for _, table := range schema.GetTables() {
			columns := []*aiGenerateColumn{}

			tableColsMap, ok := groupedSchemas[sqlmanager_shared.BuildTable(schema.GetSchema(), table.GetTable())]
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

	// builds a map of table key to columns for AI Generated schemas as they are calculated lazily instead of via job mappings
	aiGroupedTableCols := map[string][]string{}
	for _, agm := range mappings {
		key := neosync_benthos.BuildBenthosTable(agm.Schema, agm.Table)
		for _, col := range agm.Columns {
			aiGroupedTableCols[key] = append(aiGroupedTableCols[key], col.Column)
		}
	}

	return sourceResponses, aiGroupedTableCols, nil
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
