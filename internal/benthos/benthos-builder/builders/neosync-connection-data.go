package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

type neosyncConnectionDataBuilder struct {
	connectiondataclient  mgmtv1alpha1connect.ConnectionDataServiceClient
	sqlmanagerclient      sqlmanager.SqlManagerClient
	sourceJobRunId        *string
	syncConfigs           []*runconfigs.RunConfig
	destinationConnection *mgmtv1alpha1.Connection
	sourceConnectionType  bb_shared.ConnectionType
}

func NewNeosyncConnectionDataSyncBuilder(
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceJobRunId *string,
	syncConfigs []*runconfigs.RunConfig,
	destinationConnection *mgmtv1alpha1.Connection,
	sourceConnectionType bb_shared.ConnectionType,
) bb_internal.BenthosBuilder {
	return &neosyncConnectionDataBuilder{
		connectiondataclient:  connectiondataclient,
		sqlmanagerclient:      sqlmanagerclient,
		sourceJobRunId:        sourceJobRunId,
		syncConfigs:           syncConfigs,
		destinationConnection: destinationConnection,
		sourceConnectionType:  sourceConnectionType,
	}
}

func (b *neosyncConnectionDataBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job
	configs := []*bb_internal.BenthosSourceConfig{}

	for _, config := range b.syncConfigs {
		schema, table := sqlmanager_shared.SplitTableKey(config.Table())

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Logger: &neosync_benthos.LoggerConfig{
					Level:        "ERROR",
					AddTimestamp: true,
				},
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						NeosyncConnectionData: &neosync_benthos.NeosyncConnectionData{
							ConnectionId:   sourceConnection.GetId(),
							ConnectionType: string(b.sourceConnectionType),
							JobId:          &job.Id,
							JobRunId:       b.sourceJobRunId,
							Schema:         schema,
							Table:          table,
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{},
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
		configs = append(configs, &bb_internal.BenthosSourceConfig{
			Name:      fmt.Sprintf("%s.%s", config.Table(), config.RunType()),
			Config:    bc,
			DependsOn: config.DependsOn(),
			RunType:   config.RunType(),

			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: sourceConnection.Id}},

			TableSchema: schema,
			TableName:   table,
			Columns:     config.InsertColumns(),
			PrimaryKeys: config.PrimaryKeys(),
		})
	}

	return configs, nil
}

func (b *neosyncConnectionDataBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	return nil, errors.ErrUnsupported
}
