package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	benthos_builder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/builder"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
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
	benthosManager := benthos_builder.NewBenthosConfigManager(b.sqlmanagerclient, b.transformerclient, b.redisConfig, false)
	responses, err := benthosManager.GenerateBenthosConfigs(ctx, job, sourceConnection, destConnections, wfmetadata.WorkflowId, nil, slogger)
	if err != nil {
		return nil, err
	}
	return &GenerateBenthosConfigsResponse{
		AccountId:      job.AccountId,
		BenthosConfigs: responses,
	}, nil
}
