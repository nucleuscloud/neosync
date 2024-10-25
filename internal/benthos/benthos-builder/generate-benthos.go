package benthos_builder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	wfmetadata *workflowMetadata,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}

	// Get source connection and determine database type
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get source connection: %w", err)
	}

	dbType := getDatabaseType(sourceConnection)
	builder, err := benthosbuilder.New(dbType)
	if err != nil {
		return nil, fmt.Errorf("unable to create builder: %w", err)
	}

	// Build source config based on flow type
	sourceParams := &benthosbuilder.SourceParams{
		Job:              job,
		SourceConnection: sourceConnection,
		Logger:           slogger,
		// ... other params
	}

	var sourceConfig *benthosbuilder.BenthosSourceConfig
	switch getFlowType(job) {
	case benthosbuilder.FlowTypeSync:
		sourceConfig, err = builder.BuildSyncSourceConfig(ctx, sourceParams)
	case benthosbuilder.FlowTypeGenerate:
		sourceConfig, err = builder.BuildGenerateSourceConfig(ctx, sourceParams)
	case benthosbuilder.FlowTypeAIGenerate:
		sourceConfig, err = builder.BuildAIGenerateSourceConfig(ctx, sourceParams)
	default:
		return nil, fmt.Errorf("unsupported flow type")
	}
	if err != nil {
		return nil, err
	}

	// Build destination configs
	var responses []*BenthosConfigResponse
	for destIdx, destination := range job.Destinations {
		destConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, err
		}

		destDbType := getDatabaseType(destConnection)
		destBuilder, err := benthosbuilder.New(destDbType)
		if err != nil {
			return nil, err
		}

		destParams := &benthosbuilder.DestinationParams{
			SourceConfig:   sourceConfig,
			DestinationIdx: destIdx,
			Destination:    destination,
			DestConnection: destConnection,
			// ... other params
		}

		destConfig, err := destBuilder.BuildDestinationConfig(ctx, destParams)
		if err != nil {
			return nil, err
		}

		// Convert configs to response format
		response := convertToResponse(sourceConfig, destConfig)
		responses = append(responses, response)
	}

	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
		AccountId:      job.GetAccountId(),
	}, nil
}
