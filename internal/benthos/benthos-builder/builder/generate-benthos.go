package benthos_builder

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

func (b *BenthosConfigManager) GenerateBenthosConfigs(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnections []*mgmtv1alpha1.Connection,
	runId string,
	metricLabelKeyVals map[string]string,
	logger *slog.Logger,
) ([]*BenthosConfigResponse, error) {
	sourceConnectionType := bb_shared.GetConnectionType(sourceConnection)
	jobType := bb_shared.GetJobType(job)
	logger = logger.With(
		"sourceConnectionType", sourceConnectionType,
		"jobType", jobType,
	)
	dbBuilder, err := b.sourceProvider.NewBuilder(sourceConnectionType, jobType)
	if err != nil {
		return nil, fmt.Errorf("unable to create benthos builder: %w", err)
	}
	logger.Debug(fmt.Sprintf("created source benthos builder for %s", sourceConnectionType))

	sourceParams := &bb_shared.SourceParams{
		Job:              job,
		RunId:            runId,
		SourceConnection: sourceConnection,
		Logger:           logger,
	}

	// also builds processors
	sourceConfigs, err := dbBuilder.BuildSourceConfigs(ctx, sourceParams)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("built %d source configs", len(sourceConfigs)))

	destinationOpts := buildDestinationOptionsMap(job.GetDestinations())

	logger.Debug(fmt.Sprintf("building %d destination configs", len(destinationConnections)))
	responses := []*BenthosConfigResponse{}
	for destIdx, destConnection := range destinationConnections {
		// Create destination builder
		destConnectionType := bb_shared.GetConnectionType(destConnection)
		destBuilder, err := b.destinationProvider.NewBuilder(destConnectionType, jobType)
		if err != nil {
			return nil, fmt.Errorf("unable to create destination builder: %w", err)
		}
		logger.Debug(fmt.Sprintf("created destination benthos builder for %s", destConnectionType))

		destOpts, ok := destinationOpts[destConnection.GetId()]
		if !ok {
			return nil, fmt.Errorf("unable to find destination options for connection: %s", destConnection.GetId())
		}

		for _, sourceConfig := range sourceConfigs {
			destParams := &bb_shared.DestinationParams{
				SourceConfig:    sourceConfig,
				Job:             job,
				RunId:           runId,
				DestinationIdx:  destIdx,
				DestinationOpts: destOpts,
				DestConnection:  destConnection,
				Logger:          logger,
			}

			destConfig, err := destBuilder.BuildDestinationConfig(ctx, destParams)
			if err != nil {
				return nil, err
			}
			sourceConfig.Config.Output.Broker.Outputs = append(sourceConfig.Config.Output.Broker.Outputs, destConfig.Outputs...)
			sourceConfig.BenthosDsns = append(sourceConfig.BenthosDsns, destConfig.BenthosDsns...)
		}
		logger.Debug(fmt.Sprintf("applied destination (%s) to %d source configs", destConnectionType, len(sourceConfigs)))
	}

	for _, sourceConfig := range sourceConfigs {
		response := convertToResponse(sourceConfig, sourceConnectionType)
		responses = append(responses, response)
	}

	// pass in all the labels??
	if b.metricsEnabled {
		logger.Debug("metrics enabled. applying metric labels")
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, job.Id),
			// need to pass these in
			// metrics.NewEqLabel(metrics.TemporalWorkflowId, withEnvInterpolation(metrics.TemporalWorkflowIdEnvKey)),
			// metrics.NewEqLabel(metrics.TemporalRunId, withEnvInterpolation(metrics.TemporalRunIdEnvKey)),
			metrics.NewEqLabel(metrics.NeosyncDateLabel, withEnvInterpolation(metrics.NeosyncDateEnvKey)),
		}
		for key, val := range metricLabelKeyVals {
			labels = append(labels, metrics.NewEqLabel(key, val))
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

	logger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
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

func convertToResponse(sourceConfig *bb_shared.BenthosSourceConfig, sourceConnectionType bb_shared.ConnectionType) *BenthosConfigResponse {
	return &BenthosConfigResponse{
		Name:                    sourceConfig.Name,
		Config:                  sourceConfig.Config,
		DependsOn:               sourceConfig.DependsOn,
		RunType:                 sourceConfig.RunType,
		TableSchema:             sourceConfig.TableSchema,
		TableName:               sourceConfig.TableName,
		Columns:                 sourceConfig.Columns,
		RedisDependsOn:          sourceConfig.RedisDependsOn,
		ColumnDefaultProperties: sourceConfig.ColumnDefaultProperties,
		Processors:              sourceConfig.Processors,
		BenthosDsns:             sourceConfig.BenthosDsns,
		RedisConfig:             sourceConfig.RedisConfig,
		SourceConnectionType:    string(sourceConnectionType),
		metriclabels:            sourceConfig.Metriclabels,
	}
}

/*

benthosbuilder.registersource(job sourceConnection)
benthosbuilder.registerdestionation(job, destinationConnection)
benthosbuilder.generateconfigs

registersource
	newpostgressyncbuilder

registerdestination
	newmysqlsyncbuilder



*/
