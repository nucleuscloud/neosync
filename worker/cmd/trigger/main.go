package main

import (
	"context"

	accounthook_events "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/events"
	accounthook_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow"
	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()

	clientOptions := client.Options{
		HostPort: "localhost:7233",
	}

	c, err := client.DialContext(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "sync-job",
	}

	req := &accounthook_workflow.AccountHookWorkflowRequest{
		Event: accounthook_events.NewEvent_JobRunCreated("c79eca6c-a9f8-40f4-9ca8-b0b8eb0d1a21", "456", "789"),
	}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, accounthook_workflow.ProcessAccountHook, req)
	if err != nil {
		panic(err)
	}

	var resp accounthook_workflow.AccountHookWorkflowResponse
	if err := we.Get(ctx, &resp); err != nil {
		panic(err)
	}
}
