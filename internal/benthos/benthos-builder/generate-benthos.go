package benthosbuilder

import (
	"context"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	"github.com/nucleuscloud/neosync/internal/runconfigs"

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
		JobRunId:         b.jobRunId,
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
	for _, destConnection := range b.destinationConnections {
		destBuilder, err := b.destinationProvider.GetBuilder(b.job, destConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to create destination builder: %w", err)
		}
		b.logger.Debug("created destination benthos builder for destination")

		destOpts, ok := destinationOpts[destConnection.GetId()]
		if !ok {
			return nil, fmt.Errorf(
				"unable to find destination options for connection: %s",
				destConnection.GetId(),
			)
		}

		for _, sourceConfig := range sourceConfigs {
			destParams := &bb_internal.DestinationParams{
				SourceConfig:    sourceConfig,
				Job:             b.job,
				JobRunId:        b.jobRunId,
				DestinationOpts: destOpts,
				DestConnection:  destConnection,
				Logger:          b.logger,
			}

			destConfig, err := destBuilder.BuildDestinationConfig(ctx, destParams)
			if err != nil {
				return nil, err
			}
			sourceConfig.Config.Output.Broker.Outputs = append(
				sourceConfig.Config.Output.Broker.Outputs,
				destConfig.Outputs...)
			sourceConfig.BenthosDsns = append(sourceConfig.BenthosDsns, destConfig.BenthosDsns...)
		}
		b.logger.Debug(fmt.Sprintf("applied destination to %d source configs", len(sourceConfigs)))
	}

	if b.metricsEnabled {
		b.logger.Debug("metrics enabled. applying metric labels")
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, b.job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, b.job.Id),
			metrics.NewEqLabel(
				metrics.NeosyncDateLabel,
				bb_shared.WithEnvInterpolation(metrics.NeosyncDateEnvKey),
			),
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

	var outputConfigs []*bb_internal.BenthosSourceConfig
	if isOnlyBucketDestinations(b.job.Destinations) {
		for _, sc := range sourceConfigs {
			if sc.RunType == runconfigs.RunTypeInsert {
				sc.DependsOn = []*runconfigs.DependsOn{}
				outputConfigs = append(outputConfigs, sc)
			}
		}
	} else {
		outputConfigs = sourceConfigs
	}

	for _, config := range outputConfigs {
		response := convertToResponse(config)
		responses = append(responses, response)
	}

	b.logger.Info(fmt.Sprintf("successfully built %d benthos configs", len(responses)))
	return responses, nil
}

// builds map of destination id -> destination options
func buildDestinationOptionsMap(
	jobDests []*mgmtv1alpha1.JobDestination,
) map[string]*mgmtv1alpha1.JobDestinationOptions {
	destOpts := map[string]*mgmtv1alpha1.JobDestinationOptions{}
	for _, dest := range jobDests {
		destOpts[dest.GetConnectionId()] = dest.GetOptions()
	}
	return destOpts
}

func convertToResponse(sourceConfig *bb_internal.BenthosSourceConfig) *BenthosConfigResponse {
	return &BenthosConfigResponse{
		Name:                    sourceConfig.Name,
		Config:                  sourceConfig.Config,
		DependsOn:               sourceConfig.DependsOn,
		TableSchema:             sourceConfig.TableSchema,
		TableName:               sourceConfig.TableName,
		Columns:                 sourceConfig.Columns,
		RunType:                 sourceConfig.RunType,
		ColumnDefaultProperties: sourceConfig.ColumnDefaultProperties,
		RedisDependsOn:          sourceConfig.RedisDependsOn,
		BenthosDsns:             sourceConfig.BenthosDsns,
		RedisConfig:             sourceConfig.RedisConfig,
		ColumnIdentityCursors:   sourceConfig.ColumnIdentityCursors,
	}
}

func isOnlyBucketDestinations(destinations []*mgmtv1alpha1.JobDestination) bool {
	for _, dest := range destinations {
		if dest.GetOptions().GetAwsS3Options() == nil &&
			dest.GetOptions().GetGcpCloudstorageOptions() == nil {
			return false
		}
	}
	return true
}
