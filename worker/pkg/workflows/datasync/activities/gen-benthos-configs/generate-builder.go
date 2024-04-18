package genbenthosconfigs_activity

import (
	"context"
	"fmt"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type generateSourceTableOptions struct {
	Count int
}

func buildBenthosGenerateSourceConfigResponses(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	mappings []*tableMapping,
	sourceTableOpts map[string]*generateSourceTableOptions,
	columnInfo map[string]*dbschemas_utils.ColumnInfo,
	dependencyMap map[string][]*tabledependency.RunConfig,
	driver, dsnConnectionId string,
	tableConstraintsMap map[string]*dbschemas_utils.TableConstraints,
	primaryKeyMap map[string][]string,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for _, tableMapping := range mappings {
		if shared.AreAllColsNull(tableMapping.Mappings) {
			// skiping table as no columns are mapped
			continue
		}

		tableName := neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)
		runConfigs := dependencyMap[tableName]

		var count = 0
		tableOpt := sourceTableOpts[tableName]
		if tableOpt != nil {
			count = tableOpt.Count
		}

		jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, tableMapping.Mappings)
		if err != nil {
			return nil, err
		}

		mutations, err := buildMutationConfigs(ctx, transformerclient, tableMapping.Mappings, columnInfo)
		if err != nil {
			return nil, err
		}
		var processors []neosync_benthos.ProcessorConfig
		// for the generate input, benthos requires a mapping, so falling back to a
		// generic empty object if the mutations are empty
		if mutations == "" {
			mutations = "root = {}"
		}
		processors = append(processors, neosync_benthos.ProcessorConfig{Mutation: &mutations})

		if jsCode != "" {
			processors = append(processors, neosync_benthos.ProcessorConfig{Javascript: &neosync_benthos.JavascriptConfig{Code: jsCode}})
		}
		if len(processors) > 0 {
			// add catch and error processor
			processors = append(processors, neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
				{Error: &neosync_benthos.ErrorProcessorConfig{
					ErrorMsg: `${! meta("fallback_error")}`,
				}},
			}})
		}

		var bc *neosync_benthos.BenthosConfig
		if len(runConfigs) > 0 && len(runConfigs[0].DependsOn) > 0 {
			columnNameMap := map[string]string{}
			tableColsMaps := map[string][]string{}

			constraints := tableConstraintsMap[tableName]
			for _, tc := range constraints.Constraints {
				columnNameMap[fmt.Sprintf("%s.%s", tc.ForeignKey.Table, tc.ForeignKey.Column)] = tc.Column
				tableColsMaps[tc.ForeignKey.Table] = append(tableColsMaps[tc.ForeignKey.Table], tc.ForeignKey.Table)
			}

			bc = &neosync_benthos.BenthosConfig{
				StreamConfig: neosync_benthos.StreamConfig{
					Input: &neosync_benthos.InputConfig{
						Inputs: neosync_benthos.Inputs{
							GenerateSqlSelect: &neosync_benthos.GenerateSqlSelect{
								Count:           count,
								Mapping:         mutations,
								Driver:          driver,
								Dsn:             "${SOURCE_CONNECTION_DSN}",
								TableColumnsMap: tableColsMaps,
								ColumnNameMap:   columnNameMap,
							},
						},
					},
					Pipeline: &neosync_benthos.PipelineConfig{
						Threads:    -1,
						Processors: processors,
					},
					Output: &neosync_benthos.OutputConfig{
						Outputs: neosync_benthos.Outputs{
							Broker: &neosync_benthos.OutputBrokerConfig{
								Pattern: "fan_out_sequential_fail_fast",
								Outputs: []neosync_benthos.Outputs{},
							},
						},
					},
				},
			}
		} else {
			bc = &neosync_benthos.BenthosConfig{
				StreamConfig: neosync_benthos.StreamConfig{
					Input: &neosync_benthos.InputConfig{
						Inputs: neosync_benthos.Inputs{
							Generate: &neosync_benthos.Generate{
								Interval: "",
								Count:    count,
								Mapping:  "root = {}",
							},
						},
					},
					Pipeline: &neosync_benthos.PipelineConfig{
						Threads:    -1,
						Processors: processors,
					},
					Output: &neosync_benthos.OutputConfig{
						Outputs: neosync_benthos.Outputs{
							Broker: &neosync_benthos.OutputBrokerConfig{
								Pattern: "fan_out_sequential_fail_fast",
								Outputs: []neosync_benthos.Outputs{},
							},
						},
					},
				},
			}
		}

		resp := &BenthosConfigResponse{
			Name:        neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:      bc,
			DependsOn:   []*tabledependency.DependsOn{},
			BenthosDsns: []*shared.BenthosDsn{{ConnectionId: dsnConnectionId, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,

			// Processors: processors,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "generate"),
			},
		}
		if len(runConfigs) > 1 {
			// circular dependency
			for _, c := range runConfigs {
				if c.Columns != nil && c.Columns.Exclude != nil && len(c.Columns.Exclude) > 0 {
					resp.excludeColumns = c.Columns.Exclude
					resp.DependsOn = c.DependsOn
				} else if c.Columns != nil && c.Columns.Include != nil && len(c.Columns.Include) > 0 {
					pks := primaryKeyMap[tableName]
					if len(pks) == 0 {
						return nil, fmt.Errorf("no primary keys found for table (%s). Unable to build update query", tableName)
					}

					// config for sql update
					resp.updateConfig = c
					resp.primaryKeys = pks
				}
			}
		} else if len(runConfigs) == 1 {
			resp.DependsOn = runConfigs[0].DependsOn
		}

		responses = append(responses, resp)
	}

	return responses, nil
}
