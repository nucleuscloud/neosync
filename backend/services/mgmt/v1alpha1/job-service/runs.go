package v1alpha1_jobservice

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
)

func (s *Service) GetJobRuns(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	var accountId string
	var workflows []*workflowpb.WorkflowExecutionInfo
	switch id := req.Msg.Id.(type) {
	case *mgmtv1alpha1.GetJobRunsRequest_JobId:
		jobUuid, err := nucleusdb.ToUuid(id.JobId)
		if err != nil {
			return nil, err
		}
		job, err := s.db.Q.GetJobById(ctx, jobUuid)
		if err != nil {
			return nil, err
		}
		accountId = nucleusdb.UUIDString(job.AccountID)
		workflows, err = getWorkflowExecutionsByJobIds(ctx, s.temporalClient, logger, s.cfg.TemporalNamespace, []string{id.JobId})
		if err != nil {
			return nil, err
		}
	case *mgmtv1alpha1.GetJobRunsRequest_AccountId:
		accountId = id.AccountId
		accountUuid, err := nucleusdb.ToUuid(accountId)
		if err != nil {
			return nil, err
		}
		jobs, err := s.db.Q.GetJobsByAccount(ctx, accountUuid)
		if err != nil {
			return nil, err
		}
		jobIds := []string{}
		for i := range jobs {
			job := jobs[i]
			jobIds = append(jobIds, nucleusdb.UUIDString(job.ID))
		}
		workflows, err = getWorkflowExecutionsByJobIds(ctx, s.temporalClient, logger, s.cfg.TemporalNamespace, jobIds)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("must provide jobId or accountId")
	}

	_, err := s.verifyUserInAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}

	var workflowStartedEvent *history.WorkflowExecutionStartedEventAttributes
	runs := []*mgmtv1alpha1.JobRun{}
	for _, w := range workflows {
		jsonF, _ := json.MarshalIndent(w, "", " ")
		fmt.Printf("\n\n w: %s \n\n", string(jsonF))
		iter := s.temporalClient.GetWorkflowHistory(ctx, w.Execution.WorkflowId, w.Execution.RunId, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, _ := iter.Next()
			if event.GetWorkflowExecutionStartedEventAttributes() != nil {
				workflowStartedEvent = event.GetWorkflowExecutionStartedEventAttributes()
			}

			jsonF, _ := json.MarshalIndent(event, "", " ")
			fmt.Printf("\n\n event: %s \n\n", string(jsonF))
		}
		res, err := s.temporalClient.DescribeWorkflowExecution(ctx, w.Execution.WorkflowId, w.Execution.RunId)
		if err != nil {
			return nil, err
		}
		d, _ := json.MarshalIndent(res, "", " ")
		fmt.Printf("\n\n res: %s \n\n", string(d))
		runs = append(runs, dtomaps.ToJobRunDto(w))
	}
	fmt.Println(workflowStartedEvent)

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
		JobRuns: runs,
	}), nil
}

func (s *Service) GetJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{}), nil
}

func (s *Service) CreateJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobRunResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.CreateJobRunResponse{}), nil
}

func (s *Service) CancelJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CancelJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CancelJobRunResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.CancelJobRunResponse{}), nil
}

func (s *Service) DeleteJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRunRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobRunResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobRunResponse{}), nil
}
