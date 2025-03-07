package piidetect_job_workflow

import (
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct{}

func New() *Workflow {
	return &Workflow{}
}

type PiiDetectRequest struct {
	JobId string
}

type PiiDetectResponse struct{}

func (w *Workflow) JobPiiDetect(ctx workflow.Context, req *PiiDetectRequest) (*PiiDetectResponse, error) {
	logger := log.With(
		workflow.GetLogger(ctx),
		"jobId", req.JobId,
	)

	logger.Info("starting PII detection")

	// get job from activity
	// spawn activity for each table to detect pii for that specific table
	// wait for all activities to complete
	// results are stored after each activity completes

	return &PiiDetectResponse{}, nil
}
