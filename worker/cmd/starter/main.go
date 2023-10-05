package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	jobId := uuid.New()
	taskqueue := "sync-job"

	wfOpts := client.StartWorkflowOptions{
		ID:                       jobId.String(),
		TaskQueue:                taskqueue,
		WorkflowExecutionTimeout: 1 * time.Minute,
		RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 1},
	}

	req := &datasync.WorkflowRequest{
		JobId: "3f7186b1-ab31-4646-a8ba-968b3c2e2769",
	}
	we, err := c.ExecuteWorkflow(context.Background(), wfOpts, datasync.Workflow, req)
	if err != nil {
		panic(err)
	}

	var result *datasync.WorkflowResponse
	err = we.Get(context.Background(), &result)
	if err != nil {
		panic(err)
	}
	log.Println("Workflow Result", result)
}
