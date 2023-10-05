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

func Workflow(wfctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		neosyncUrl = "localhost:8080"
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	wfinfo := workflow.GetInfo(wfctx)

	ctx := workflow.WithActivityOptions(wfctx, ao)

	var wfActivites *Activities
	var bcResp *GenerateBenthosConfigsResponse
	err := workflow.ExecuteActivity(ctx, wfActivites.GenerateBenthosConfigs, &GenerateBenthosConfigsRequest{
		JobId:      req.JobId,
		BackendUrl: neosyncUrl,
		WorkflowId: wfinfo.WorkflowExecution.ID,
	}).Get(ctx, &bcResp)
	if err != nil {
		return nil, err
	}

	completed := map[string]struct{}{}

	// todo: figure this out as we want to parallelize this
	bchan := workflow.NewChannel(wfctx)

	numwaits := 1 + len(bcResp.BenthosConfigs)*2

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(bchan, func(c workflow.ReceiveChannel, more bool) {
		var key string
		c.Receive(ctx, &key)
		completed[key] = struct{}{}

		for _, bc := range bcResp.BenthosConfigs {
			if _, ok := completed[bc.Name]; ok {
				continue
			}
			isReady := true
			for _, dep := range bc.DependsOn {
				if _, ok := completed[dep]; !ok {
					isReady = false
					break
				}
			}
			if isReady {
				// spawn activity
				bits, _ := yaml.Marshal(bc.Config) // todo: handle error
				future := workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(bits)})
				selector.AddFuture(future, func(f workflow.Future) {
					logger.Info(fmt.Sprintf("completed %s sync", bc.Name))
					bchan.Send(ctx, bc.Name)
				})
			}
		}
	})

	for _, bc := range bcResp.BenthosConfigs {
		if _, ok := completed[bc.Name]; ok {
			continue
		}
		if len(bc.DependsOn) != 0 {
			continue
		}
		bits, err := yaml.Marshal(bc.Config)
		if err != nil {
			return nil, err
		}
		future := workflow.ExecuteActivity(ctx, wfActivites.Sync, &SyncRequest{BenthosConfig: string(bits)})
		selector.AddFuture(future, func(f workflow.Future) {
			bchan.Send(ctx, bc.Name)
		})
	}

	for i := 0; i < numwaits; i++ {
		selector.Select(ctx)
	}

	return &WorkflowResponse{}, nil
}
