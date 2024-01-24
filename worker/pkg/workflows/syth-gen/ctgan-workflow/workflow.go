package ctganworkflow

import (
	"time"

	ctganactivities "github.com/nucleuscloud/neosync/worker/pkg/workflows/syth-gen/ctgan-activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type WorkflowRequest struct {
	JobId string
}

type WorkflowResponse struct{}

type CtganSingleTableTrainInput struct {
	Epochs          uint     `json:"epochs"`
	DiscreteColumns []string `json:"discrete_columns"`
	ModelPath       string   `json:"modelpath"`
	Dsn             string   `json:"dsn"`
	Schema          string   `json:"schema"`
	Table           string   `json:"table"`
	Columns         []string `json:"columns"`
}

const (
	ctganSingleTableActivityName = "ctgan_single_table_train"
)

func CtganWorkflow(wfctx workflow.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	ctx := workflow.WithActivityOptions(wfctx, ao)
	logger := workflow.GetLogger(ctx)
	_ = logger
	logger.Info("running training workflow", "jobId", req.JobId)

	var inputResponse *ctganactivities.GetTrainModelInputResponse
	err := workflow.ExecuteActivity(ctx, ctganactivities.GetTrainModelInput, &ctganactivities.GetTrainModelInputRequest{
		JobId: req.JobId,
	}).Get(ctx, &inputResponse)
	if err != nil {
		return nil, err
	}

	ctx = workflow.WithTaskQueue(ctx, "ml")
	err = workflow.ExecuteActivity(ctx, ctganSingleTableActivityName, &CtganSingleTableTrainInput{
		Epochs:          uint(inputResponse.Epochs),
		DiscreteColumns: inputResponse.DiscreteColumns,
		Columns:         inputResponse.Columns,
		ModelPath:       inputResponse.ModelPath,
		Dsn:             *inputResponse.SourceDsn,
		Schema:          *inputResponse.Schema,
		Table:           *inputResponse.Table,
	}).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &WorkflowResponse{}, nil
}
