package benthos_builder

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

func (b *BenthosConfigManager) GenerateBenthosConfigs(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnections []*mgmtv1alpha1.Connection,
	destinationOptions []*mgmtv1alpha1.JobDestination,
	slogger *slog.Logger,
) ([]*BenthosConfigResponse, error) {
	// Create appropriate database builder based on source type
	dbType := bb_shared.GetConnectionType(sourceConnection)
	jobType := bb_shared.GetJobType(job)
	dbBuilder, err := NewBenthosBuilder(dbType, jobType)
	if err != nil {
		return nil, fmt.Errorf("unable to create database builder: %w", err)
	}

	// Build source config based on flow type
	sourceParams := &bb_shared.SourceParams{
		Job:               job,
		SourceConnection:  sourceConnection,
		Logger:            slogger,
		TransformerClient: b.transformerclient,
		SqlManager:        b.sqlmanager,
		RedisConfig:       b.redisConfig,
		MetricsEnabled:    b.metricsEnabled,
	}

	sourceConfig, err := dbBuilder.BuildSourceConfig(ctx, sourceParams)
	if err != nil {
		return nil, err
	}

	destinationOpts := buildDestinationOptionsMap(destinationOptions)
	// Process each destination
	responses := []*BenthosConfigResponse{}
	for destIdx, destConnection := range destinationConnections {

		// Create destination builder
		destDbType := bb_shared.GetConnectionType(destConnection)
		destBuilder, err := NewBenthosBuilder(destDbType, jobType)
		if err != nil {
			return nil, fmt.Errorf("unable to create destination builder: %w", err)
		}

		destOpts, ok := destinationOpts[destConnection.GetId()]
		if !ok {
			return nil, fmt.Errorf("unable to find destination options for connection: %s", destConnection.GetId())
		}

		destParams := &bb_shared.DestinationParams{
			SourceConfig:      sourceConfig,
			DestinationIdx:    destIdx,
			DestinationOpts:   destOpts,
			DestConnection:    destConnection,
			Logger:            slogger,
			TransformerClient: b.transformerclient,
			SqlManager:        b.sqlmanager,
			RedisConfig:       b.redisConfig,
			MetricsEnabled:    b.metricsEnabled,
		}

		destConfig, err := destBuilder.BuildDestinationConfig(ctx, destParams)
		if err != nil {
			return nil, err
		}

		// more work to do here
		// needs to handle multiple destinations
		response := convertToResponse(sourceConfig, destConfig)
		responses = append(responses, response)
	}

	// pass in all the labels??
	if b.metricsEnabled {
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, job.Id),
			// need to pass these in??
			// metrics.NewEqLabel(metrics.TemporalWorkflowId, withEnvInterpolation(metrics.TemporalWorkflowIdEnvKey)),
			// metrics.NewEqLabel(metrics.TemporalRunId, withEnvInterpolation(metrics.TemporalRunIdEnvKey)),
			metrics.NewEqLabel(metrics.NeosyncDateLabel, withEnvInterpolation(metrics.NeosyncDateEnvKey)),
		}
		for _, resp := range responses {
			joinedLabels := append(labels, resp.metriclabels...) //nolint:gocritic
			resp.Config.Metrics = &neosync_benthos.Metrics{
				OtelCollector: &neosync_benthos.MetricsOtelCollector{},
				Mapping:       joinedLabels.ToBenthosMeta(),
			}
		}
	}

	// TODO should this be in benthos builder? how to handle this
	// // Set post table sync run context
	// postTableSyncRunCtx := buildPostTableSyncRunCtx(responses, job.Destinations)
	// err = b.setPostTableSyncRunCtx(ctx, postTableSyncRunCtx, job.GetAccountId())
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to set post table sync run contexts: %w", err)
	// }

	// // Set run contexts
	// responses, err = b.setRunContexts(ctx, responses, job.GetAccountId())
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to set run contexts: %w", err)
	// }

	slogger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
	return responses, nil
}

func withEnvInterpolation(input string) string {
	return fmt.Sprintf("${%s}", input)
}

// builds map of destination id -> destination options
func buildDestinationOptionsMap(jobDests []*mgmtv1alpha1.JobDestination) map[string]*mgmtv1alpha1.JobDestinationOptions {
	destOpts := map[string]*mgmtv1alpha1.JobDestinationOptions{}
	for _, dest := range jobDests {
		destOpts[dest.GetConnectionId()] = dest.GetOptions()
	}
	return destOpts
}

func convertToResponse(sourceConfig *bb_shared.BenthosSourceConfig, destConfig *bb_shared.BenthosDestinationConfig) *BenthosConfigResponse {
	return &BenthosConfigResponse{
		Name:                    sourceConfig.Name,
		Config:                  sourceConfig.Config,
		DependsOn:               sourceConfig.DependsOn,
		RunType:                 sourceConfig.RunType,
		TableSchema:             sourceConfig.TableSchema,
		TableName:               sourceConfig.TableName,
		Columns:                 sourceConfig.Columns,
		RedisDependsOn:          sourceConfig.RedisDependsOn,
		ColumnDefaultProperties: sourceConfig.DefaultProperties,
		Processors:              sourceConfig.Processors,
		BenthosDsns:             append(sourceConfig.BenthosDsns, destConfig.BenthosDsns...),
		RedisConfig:             sourceConfig.RedisConfig,
		SourceConnectionType:    string(sourceConfig.ConnectionType),
		// metriclabels:            convertMetricLabels(sourceConfig.MetricLabels),
	}
}
