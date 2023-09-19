package datasync

import (
	"time"

	"github.com/spf13/viper"
	"go.temporal.io/sdk/workflow"
	"gopkg.in/yaml.v3"
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

	var wfActivites *Activities
	var bcResp *GenerateBenthosConfigsResponse
	err := workflow.ExecuteActivity(ctx, wfActivites.GenerateBenthosConfigs, &GenerateBenthosConfigsRequest{
		JobId:      req.JobId,
		BackendUrl: neosyncUrl,
	}).Get(ctx, &bcResp)
	if err != nil {
		return nil, err
	}

	// todo: figure this out as we want to parallelize this

	//nolint:gocritic
	// futures := make([]workflow.Future, len(bcResp.BenthosConfigs))
	for idx := range bcResp.BenthosConfigs {
		bc := bcResp.BenthosConfigs[idx]
		bits, err := yaml.Marshal(bc.Config)
		if err != nil {
			return nil, err
		}
		var resp *SyncResponse
		err = workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(bits)}).Get(ctx, &resp)
		if err != nil {
			return nil, err
		}
		//nolint:gocritic
		//future := workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(bits)})
		// futures = append(futures, future)
	}

	// todo: for some reason the future in the list is nil. We must be doing something here we shouldn't be

	// todo: this should be heavily optimized
	// for _, future := range futures {
	// 	fmt.Println("future", future)
	// 	var resp *SyncResponse
	// 	err := future.Get(ctx, &resp)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return &WorkflowResponse{}, nil
}
