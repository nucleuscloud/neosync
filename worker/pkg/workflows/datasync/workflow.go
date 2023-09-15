package datasync

import (
	"time"

	"github.com/spf13/viper"
	"go.temporal.io/sdk/workflow"
)

type WorkflowRequest struct {
	JobId string
}

type WorkflowResponse struct{}

func Workflow(ctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		neosyncUrl = "localhost:8080"
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var bcResp *GenerateBenthosConfigsResponse

	var wfActivites *Activities
	err := workflow.ExecuteActivity(ctx, wfActivites.GenerateBenthosConfigs, &GenerateBenthosConfigsRequest{
		JobId:      req.JobId,
		BackendUrl: neosyncUrl,
	}).Get(ctx, bcResp)
	if err != nil {
		return nil, err
	}

	futures := make([]workflow.Future, len(bcResp.BenthosConfigs))
	for idx := range bcResp.BenthosConfigs {
		bc := bcResp.BenthosConfigs[idx]
		future := workflow.ExecuteActivity(ctx, nil, bc)
		futures = append(futures, future)
	}

	// todo: this should be heavily optimized
	for _, future := range futures {
		var resp any
		err := future.Get(ctx, resp)
		if err != nil {
			return nil, err
		}
	}

	return &WorkflowResponse{}, nil
}
