package datasync_workflow2

// https://chatgpt.com/share/67b90562-2884-8010-889e-404d4586eba9

import (
	"github.com/nucleuscloud/neosync/internal/ee/license"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type WorkflowRequest struct {
	JobId string
}

type WorkflowResponse struct{}

type Workflow struct {
	eelicense license.EEInterface
}

func New(eelicense license.EEInterface) *Workflow {
	return &Workflow{
		eelicense: eelicense,
	}
}

func (w *Workflow) Workflow(ctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	logger := workflow.GetLogger(ctx)

	_, err := w.preSync(ctx, req, logger)
	if err != nil {
		return nil, err
	}

	if _, err := w.spawnAccountHook(ctx, req, logger); err != nil {
		return nil, err
	}

	if _, err := w.executeCoreSync(ctx, req, logger); err != nil {
		return nil, err
	}

	if _, err := w.postSync(ctx, req, logger); err != nil {
		return nil, err
	}

	if _, err := w.spawnAccountHook(ctx, req, logger); err != nil {
		return nil, err
	}

	return nil, nil
}

func (w *Workflow) preSync(ctx workflow.Context, req *WorkflowRequest, logger log.Logger) (any, error) {
	return nil, nil
}

func (w *Workflow) spawnAccountHook(ctx workflow.Context, req *WorkflowRequest, logger log.Logger) (any, error) {
	return nil, nil
}

func (w *Workflow) executeCoreSync(ctx workflow.Context, req *WorkflowRequest, logger log.Logger) (any, error) {
	return nil, nil
}

func (w *Workflow) postSync(ctx workflow.Context, req *WorkflowRequest, logger log.Logger) (any, error) {
	return nil, nil
}
