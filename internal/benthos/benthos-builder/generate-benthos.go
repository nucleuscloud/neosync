package benthosbuilder

import (
	"context"
	"encoding/json"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

func (b *BenthosConfigManager) GenerateBenthosConfigs(
	ctx context.Context,
) ([]*BenthosConfigResponse, error) {
	dbBuilder, err := b.sourceProvider.GetBuilder(b.job, b.sourceConnection)
	if err != nil {
		return nil, fmt.Errorf("unable to create benthos builder: %w", err)
	}
	b.logger.Debug("created source benthos builder")

	sourceParams := &bb_internal.SourceParams{
		Job:              b.job,
		RunId:            b.runId,
		SourceConnection: b.sourceConnection,
		Logger:           b.logger,
	}

	sourceConfigs, err := dbBuilder.BuildSourceConfigs(ctx, sourceParams)
	if err != nil {
		return nil, err
	}
	b.logger.Debug(fmt.Sprintf("built %d source configs", len(sourceConfigs)))

	destinationOpts := buildDestinationOptionsMap(b.job.GetDestinations())

	b.logger.Debug(fmt.Sprintf("building %d destination configs", len(b.destinationConnections)))
	responses := []*BenthosConfigResponse{}
	for destIdx, destConnection := range b.destinationConnections {
		destBuilder, err := b.destinationProvider.GetBuilder(b.job, destConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to create destination builder: %w", err)
		}
		b.logger.Debug("created destination benthos builder for destination")

		destOpts, ok := destinationOpts[destConnection.GetId()]
		if !ok {
			return nil, fmt.Errorf("unable to find destination options for connection: %s", destConnection.GetId())
		}

		for _, sourceConfig := range sourceConfigs {
			dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", destIdx)
			dsn := fmt.Sprintf("${%s}", dstEnvVarKey)
			destParams := &bb_internal.DestinationParams{
				SourceConfig:    sourceConfig,
				Job:             b.job,
				RunId:           b.runId,
				DestinationOpts: destOpts,
				DestConnection:  destConnection,
				DestEnvVarKey:   dstEnvVarKey,
				DSN:             dsn,
				Logger:          b.logger,
			}

			destConfig, err := destBuilder.BuildDestinationConfig(ctx, destParams)
			if err != nil {
				return nil, err
			}
			sourceConfig.Config.Output.Broker.Outputs = append(sourceConfig.Config.Output.Broker.Outputs, destConfig.Outputs...)
			sourceConfig.BenthosDsns = append(sourceConfig.BenthosDsns, destConfig.BenthosDsns...)
		}
		b.logger.Debug(fmt.Sprintf("applied destination to %d source configs", len(sourceConfigs)))
	}

	if b.metricsEnabled {
		b.logger.Debug("metrics enabled. applying metric labels")
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, b.job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, b.job.Id),
			metrics.NewEqLabel(metrics.NeosyncDateLabel, withEnvInterpolation(metrics.NeosyncDateEnvKey)),
		}
		for key, val := range b.metricLabelKeyVals {
			labels = append(labels, metrics.NewEqLabel(key, val))
		}
		for _, resp := range sourceConfigs {
			joinedLabels := append(labels, resp.Metriclabels...) //nolint:gocritic
			resp.Config.Metrics = &neosync_benthos.Metrics{
				OtelCollector: &neosync_benthos.MetricsOtelCollector{},
				Mapping:       joinedLabels.ToBenthosMeta(),
			}
		}
	}

	for _, sourceConfig := range sourceConfigs {
		response := convertToResponse(sourceConfig)
		responses = append(responses, response)
	}

	jsonF, _ := json.MarshalIndent(responses, "", " ")
	fmt.Printf("%s \n", string(jsonF))

	b.logger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
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

func convertToResponse(sourceConfig *bb_internal.BenthosSourceConfig) *BenthosConfigResponse {
	return &BenthosConfigResponse{
		Name:           sourceConfig.Name,
		Config:         sourceConfig.Config,
		DependsOn:      sourceConfig.DependsOn,
		TableSchema:    sourceConfig.TableSchema,
		TableName:      sourceConfig.TableName,
		Columns:        sourceConfig.Columns,
		RedisDependsOn: sourceConfig.RedisDependsOn,
		BenthosDsns:    sourceConfig.BenthosDsns,
		RedisConfig:    sourceConfig.RedisConfig,
	}
}
