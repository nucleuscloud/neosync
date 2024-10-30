package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"gopkg.in/yaml.v3"
)

func (b *benthosBuilder) GenerateBenthosConfigsNew(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	wfmetadata *workflowMetadata,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	destConnections := []*mgmtv1alpha1.Connection{}
	for _, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}
		destConnections = append(destConnections, destinationConnection)
	}

	benthosManagerConfig := &benthosbuilder.WorkerBenthosConfig{
		Job:                    job,
		SourceConnection:       sourceConnection,
		DestinationConnections: destConnections,
		RunId:                  wfmetadata.WorkflowId,
		Logger:                 slogger,
		Sqlmanagerclient:       b.sqlmanagerclient,
		Transformerclient:      b.transformerclient,
		RedisConfig:            b.redisConfig,
		MetricsEnabled:         false,
		MetricLabelKeyVals: map[string]string{
			metrics.TemporalWorkflowId: withEnvInterpolation(metrics.TemporalWorkflowIdEnvKey),
			metrics.TemporalRunId:      withEnvInterpolation(metrics.TemporalRunIdEnvKey),
		},
	}
	benthosManager, err := benthosbuilder.NewWorkerBenthosConfigManager(benthosManagerConfig)
	if err != nil {
		return nil, err
	}
	responses, err := benthosManager.GenerateBenthosConfigs(ctx)
	if err != nil {
		return nil, err
	}

	outputConfigs, err := b.setRunContextsNew(ctx, responses, job.GetAccountId())
	if err != nil {
		return nil, fmt.Errorf("unable to set all run contexts for benthos configs: %w", err)
	}
	return &GenerateBenthosConfigsResponse{
		AccountId:      job.AccountId,
		BenthosConfigs: outputConfigs,
	}, nil
}

// this method modifies the input responses by nilling out the benthos config. it returns the same slice for convenience
func (b *benthosBuilder) setRunContextsNew(
	ctx context.Context,
	responses []*benthosbuilder.BenthosConfigResponse,
	accountId string,
) ([]*benthosbuilder.BenthosConfigResponse, error) {
	rcstream := b.jobclient.SetRunContexts(ctx)

	for _, config := range responses {
		bits, err := yaml.Marshal(config.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal benthos config: %w", err)
		}
		err = rcstream.Send(&mgmtv1alpha1.SetRunContextsRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				JobRunId:   b.workflowId,
				ExternalId: shared.GetBenthosConfigExternalId(config.Name),
				AccountId:  accountId,
			},
			Value: bits,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send run context: %w", err)
		}
		config.Config = nil // nilling this out so that it does not persist in temporal
	}

	_, err := rcstream.CloseAndReceive()
	if err != nil {
		return nil, fmt.Errorf("unable to receive response from benthos runcontext request: %w", err)
	}
	return responses, nil
}
