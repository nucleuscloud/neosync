package benthosbuilder_builders

import (
	"context"
	"fmt"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

type mongodbSyncBuilder struct {
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
}

func NewMongoDbSyncBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
) bb_internal.BenthosBuilder {
	return &mongodbSyncBuilder{
		transformerclient: transformerclient,
	}
}

func (b *mongodbSyncBuilder) BuildSourceConfigs(
	ctx context.Context,
	params *bb_internal.SourceParams,
) ([]*bb_internal.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job
	groupedMappings := groupMappingsByTable(jobMappingsFromLegacyMappings(job.GetMappings()))

	benthosConfigs := []*bb_internal.BenthosSourceConfig{}
	for _, tableMapping := range groupedMappings {
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						PooledMongoDB: &neosync_benthos.InputMongoDb{
							ConnectionId: params.SourceConnection.GetId(),
							Database:     tableMapping.Schema,
							Collection:   tableMapping.Table,
							Query:        "root = this",
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

		columns := []string{}
		for _, jm := range tableMapping.Mappings {
			columns = append(columns, jm.Column)
		}

		schemaTable := sqlmanager_shared.SchemaTable{
			Schema: tableMapping.Schema,
			Table:  tableMapping.Table,
		}
		runconfigType := runconfigs.RunTypeInsert
		runconfigId := fmt.Sprintf("%s.%s", schemaTable.String(), runconfigType)
		splitColumnPaths := true
		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			b.transformerclient,
			runconfigs.NewRunConfig(
				runconfigId,
				schemaTable,
				runconfigType,
				[]string{},
				nil,
				columns,
				columns,
				nil,
				splitColumnPaths,
			),
			map[string][]*bb_internal.ReferenceKey{},
			map[string][]*bb_internal.ReferenceKey{},
			params.Job.Id,
			params.JobRunId,
			&neosync_redis.RedisConfig{},
			tableMapping.Mappings,
			map[string]*sqlmanager_shared.DatabaseSchemaRow{},
			job.GetSource().GetOptions(),
			columns,
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.Pipeline.Processors = append(bc.Pipeline.Processors, *pc)
		}

		benthosConfigs = append(benthosConfigs, &bb_internal.BenthosSourceConfig{
			Config:      bc,
			Name:        fmt.Sprintf("%s.%s", tableMapping.Schema, tableMapping.Table), // todo
			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,
			RunType:     runconfigs.RunTypeInsert,
			DependsOn:   []*runconfigs.DependsOn{},
			Columns:     columns,
			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: sourceConnection.GetId()}},

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return benthosConfigs, nil
}

func (b *mongodbSyncBuilder) BuildDestinationConfig(
	ctx context.Context,
	params *bb_internal.DestinationParams,
) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	benthosConfig := params.SourceConfig
	config.BenthosDsns = append(
		config.BenthosDsns,
		&bb_shared.BenthosDsn{ConnectionId: params.DestConnection.GetId()},
	)
	config.Outputs = append(
		config.Outputs,
		neosync_benthos.Outputs{PooledMongoDB: &neosync_benthos.OutputMongoDb{
			ConnectionId: params.DestConnection.GetId(),

			Database:   benthosConfig.TableSchema,
			Collection: benthosConfig.TableName,
			Operation:  "update-one",
			Upsert:     true,
			DocumentMap: `
			root = {
				"$set": this
			}
		`,
			FilterMap: `
			root._id = this._id
		`,
			WriteConcern: &neosync_benthos.MongoWriteConcern{
				W: "1",
			},
		},
		},
	)
	return config, nil
}
