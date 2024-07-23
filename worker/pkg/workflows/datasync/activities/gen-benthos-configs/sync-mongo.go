package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type mongoSyncResp struct {
	BenthosConfigs []*BenthosConfigResponse
}

func (b *benthosBuilder) getMongoDbSyncBenthosConfigResponses(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	slogger *slog.Logger,
) (*mongoSyncResp, error) {
	_ = slogger
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get source connection by id: %w", err)
	}

	groupedMappings := groupMappingsByTable(job.GetMappings())

	benthosConfigs := []*BenthosConfigResponse{}
	for _, tableMapping := range groupedMappings {
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						PooledMongoDB: &neosync_benthos.InputMongoDb{
							Url:        "${SOURCE_CONNECTION_DSN}",
							Database:   tableMapping.Schema,
							Collection: tableMapping.Table,
							Query:      "root = this",
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads:    -1,
					Processors: []neosync_benthos.ProcessorConfig{},
				},
				Output: &neosync_benthos.OutputConfig{
					Outputs: neosync_benthos.Outputs{},
				},
			},
		}

		columns := []string{}
		for _, jm := range tableMapping.Mappings {
			columns = append(columns, jm.Column)
		}

		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			b.transformerclient,
			&tabledependency.RunConfig{RunType: tabledependency.RunTypeInsert},
			map[string][]*referenceKey{},
			map[string][]*referenceKey{},
			b.jobId,
			b.runId,
			&shared.RedisConfig{},
			tableMapping.Mappings,
			map[string]*sqlmanager_shared.ColumnInfo{},
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		benthosConfigs = append(benthosConfigs, &BenthosConfigResponse{
			Config:      bc,
			Name:        fmt.Sprintf("%s.%s", tableMapping.Schema, tableMapping.Table), // todo
			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,
			RunType:     tabledependency.RunTypeInsert,
			DependsOn:   []*tabledependency.DependsOn{},
			Columns:     columns,
			BenthosDsns: []*shared.BenthosDsn{{ConnectionId: sourceConnection.GetId(), EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, tableMapping.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, tableMapping.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return &mongoSyncResp{
		BenthosConfigs: benthosConfigs,
	}, nil
}
