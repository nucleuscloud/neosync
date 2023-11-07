package v1alpha1_jobservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	jsonmodels "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
		if err != nil {
			return nil, err
		}
		accountId = nucleusdb.UUIDString(job.AccountID)
		tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, accountId, logger)
		if err != nil {
			return nil, err
		}
		tconfig, err := s.db.Q.GetTemporalConfigByAccount(ctx, s.db.Db, job.AccountID)
		if err != nil {
			return nil, err
		}
		workflows, err = getWorkflowExecutionsByJobIds(ctx, tclient, logger, tconfig.Namespace, []string{id.JobId})
		if err != nil {
			return nil, err
		}
	case *mgmtv1alpha1.GetJobRunsRequest_AccountId:
		accountId = id.AccountId
		accountUuid, err := nucleusdb.ToUuid(accountId)
		if err != nil {
			return nil, err
		}
		jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, accountUuid)
		if err != nil {
			return nil, err
		}
		jobIds := []string{}
		for i := range jobs {
			job := jobs[i]
			jobIds = append(jobIds, nucleusdb.UUIDString(job.ID))
		}
		if len(jobIds) > 0 {
			tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, accountId, logger)
			if err != nil {
				return nil, err
			}
			tconfig, err := s.db.Q.GetTemporalConfigByAccount(ctx, s.db.Db, accountUuid)
			if err != nil {
				return nil, err
			}

			workflows, err = getWorkflowExecutionsByJobIds(ctx, tclient, logger, tconfig.Namespace, jobIds)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("must provide jobId or accountId")
	}

	_, err := s.verifyUserInAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}
	var tclient temporalclient.Client
	if len(workflows) > 0 {
		tclient, err = s.temporalWfManager.GetWorkflowClientByAccount(ctx, accountId, logger)
		if err != nil {
			return nil, err
		}
	}

	runs := make([]*mgmtv1alpha1.JobRun, len(workflows))
	errGrp, errCtx := errgroup.WithContext(ctx)
	for index, workflow := range workflows {
		index := index
		workflow := workflow
		errGrp.Go(func() error {
			res, err := tclient.DescribeWorkflowExecution(errCtx, workflow.Execution.WorkflowId, workflow.Execution.RunId)
			if err != nil && !strings.Contains(err.Error(), "Workflow executionsRow not found") {
				return err
			} else if err != nil && strings.Contains(err.Error(), "Workflow executionsRow not found") {
				return nil
			}
			runs[index] = dtomaps.ToJobRunDto(logger, res)
			return nil
		})
	}

	err = errGrp.Wait()
	if err != nil {
		return nil, err
	}

	filteredRuns := utils.FilterSlice[*mgmtv1alpha1.JobRun](runs, func(run *mgmtv1alpha1.JobRun) bool {
		return run != nil && run.Id != ""
	})

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
		JobRuns: filteredRuns,
	}), nil
}

func (s *Service) GetJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.JobRunId)
	verifResp, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}
	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, verifResp.NeosyncAccountId, logger)
	if err != nil {
		return nil, err
	}
	res, err := tclient.DescribeWorkflowExecution(
		ctx,
		verifResp.WorkflowExecution.Execution.WorkflowId,
		verifResp.WorkflowExecution.Execution.RunId,
	)
	if err != nil {
		return nil, err
	}

	dto := dtomaps.ToJobRunDto(logger, res)
	return connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{
		JobRun: dto,
	}), nil
}

func (s *Service) GetJobRunEvents(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunEventsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunEventsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.JobRunId)
	verifResp, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}

	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, verifResp.NeosyncAccountId, logger)
	if err != nil {
		return nil, err
	}

	isRunComplete := false
	activityOrder := []int64{}
	activityMap := map[int64]*mgmtv1alpha1.JobRunEvent{}
	iter := tclient.GetWorkflowHistory(
		ctx,
		verifResp.WorkflowExecution.Execution.WorkflowId,
		verifResp.WorkflowExecution.Execution.RunId,
		false,
		enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT,
	)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}

		switch event.EventType {
		case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			activityOrder = append(activityOrder, event.EventId)
			attributes := event.GetActivityTaskScheduledEventAttributes()
			jobRunEvent := &mgmtv1alpha1.JobRunEvent{
				Id:        event.EventId,
				Type:      attributes.ActivityType.Name,
				StartTime: timestamppb.New(*event.EventTime),
				Tasks: []*mgmtv1alpha1.JobRunEventTask{
					dtomaps.ToJobRunEventTaskDto(event, nil),
				},
			}
			if len(attributes.Input.Payloads) > 1 {
				var input mgmtv1alpha1.JobRunSyncMetadata
				err := converter.GetDefaultDataConverter().FromPayload(attributes.Input.Payloads[1], &input)
				if err != nil {
					logger.Error(fmt.Errorf("unable to convert event input payload: %w", err).Error())
				}
				jobRunEvent.Metadata = &mgmtv1alpha1.JobRunEventMetadata{
					Metadata: &mgmtv1alpha1.JobRunEventMetadata_SyncMetadata{
						SyncMetadata: &mgmtv1alpha1.JobRunSyncMetadata{
							Schema: input.Schema,
							Table:  input.Table,
						},
					},
				}
			}
			activityMap[event.EventId] = jobRunEvent
		case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
			attributes := event.GetActivityTaskStartedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.StartTime = timestamppb.New(*event.EventTime)
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
			attributes := event.GetActivityTaskCompletedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.CloseTime = timestamppb.New(*event.EventTime)
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED:
			attributes := event.GetActivityTaskCancelRequestedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
			attributes := event.GetActivityTaskCanceledEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.CloseTime = timestamppb.New(*event.EventTime)
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_ACTIVITY_TASK_FAILED:
			attributes := event.GetActivityTaskFailedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			errorDto := dtomaps.ToJobRunEventTaskErrorDto(attributes.Failure, attributes.RetryState)
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, errorDto))

		case enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
			attributes := event.GetActivityTaskTimedOutEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			errorDto := dtomaps.ToJobRunEventTaskErrorDto(attributes.Failure, attributes.RetryState)
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, errorDto))
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
			isRunComplete = true
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
			isRunComplete = true
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
			isRunComplete = true
		default:

		}
	}

	events := []*mgmtv1alpha1.JobRunEvent{}
	for _, index := range activityOrder {
		value := activityMap[index]
		events = append(events, value)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunEventsResponse{
		Events:        events,
		IsRunComplete: isRunComplete,
	}), nil
}

func (s *Service) CreateJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.JobId)
	jobUuid, err := nucleusdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil {
		return nil, err
	}
	accountId := nucleusdb.UUIDString(job.AccountID)
	_, err = s.verifyUserInAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}

	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}

	scheduleHandle := tclient.ScheduleClient().GetHandle(ctx, nucleusdb.UUIDString(job.ID))
	logger.Info("creating job run")
	err = scheduleHandle.Trigger(ctx, temporalclient.ScheduleTriggerOptions{})
	if err != nil {
		logger.Error(fmt.Errorf("unable to create job run: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateJobRunResponse{}), nil
}

func (s *Service) CancelJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CancelJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CancelJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.JobRunId)
	verifResp, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}

	logger.Info("canceling job run")
	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, verifResp.NeosyncAccountId, logger)
	if err != nil {
		return nil, err
	}
	err = tclient.CancelWorkflow(
		ctx,
		verifResp.WorkflowExecution.Execution.WorkflowId,
		verifResp.WorkflowExecution.Execution.RunId,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to cancel job run: %w", err)
	}
	return connect.NewResponse(&mgmtv1alpha1.CancelJobRunResponse{}), nil
}

func (s *Service) DeleteJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRunRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.JobRunId)
	verifResp, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}
	logger.Info("deleting job run")
	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, verifResp.NeosyncAccountId, logger)
	if err != nil {
		return nil, err
	}
	_, err = tclient.WorkflowService().DeleteWorkflowExecution(ctx, &workflowservice.DeleteWorkflowExecutionRequest{
		Namespace: verifResp.TemporalConfig.Namespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: verifResp.WorkflowExecution.Execution.WorkflowId,
			RunId:      verifResp.WorkflowExecution.Execution.RunId,
		},
	})
	if err != nil {
		logger.Error(fmt.Errorf("unable to delete job run: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteJobRunResponse{}), nil
}

func getWorkflowExecutionsByRunId(
	ctx context.Context,
	tc temporalclient.Client,
	namespace string,
	runId string,
) (*workflowpb.WorkflowExecutionInfo, error) {
	query := fmt.Sprintf("WorkflowId = %q", runId)
	request := &workflowservice.ListWorkflowExecutionsRequest{Query: query, Namespace: namespace}
	resp, err := tc.ListWorkflow(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(resp.Executions) == 0 {
		return nil, nucleuserrors.NewNotFound("job run not found")
	}
	if len(resp.Executions) > 1 {
		return nil, nucleuserrors.NewInternalError("found more than 1 job run")
	}
	return resp.Executions[0], nil
}

type getVerifiedJobRunResponse struct {
	WorkflowExecution *workflowpb.WorkflowExecutionInfo
	NeosyncAccountId  string
	TemporalConfig    *jsonmodels.TemporalConfig
}

func (s *Service) getVerifiedJobRun(
	ctx context.Context,
	logger *slog.Logger,
	runId string,

) (*getVerifiedJobRunResponse, error) {
	userAccountsResp, err := s.useraccountService.GetUserAccounts(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}))
	if err != nil {
		return nil, err
	}

	for _, userAccount := range userAccountsResp.Msg.Accounts {
		accountUuid, err := nucleusdb.ToUuid(userAccount.Id)
		if err != nil {
			return nil, err
		}
		hasNs, err := s.doesAccountHaveTemporalNamespace(ctx, accountUuid, logger)
		if err != nil {
			return nil, err
		}
		if !hasNs {
			continue
		}
		tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, userAccount.Id, logger)
		if err != nil {
			return nil, err
		}
		tconfig, err := s.db.Q.GetTemporalConfigByAccount(ctx, s.db.Db, accountUuid)
		if err != nil {
			return nil, err
		}
		run, err := getWorkflowExecutionsByRunId(ctx, tclient, tconfig.Namespace, runId)
		if err != nil {
			if errors.Is(err, nucleuserrors.NewNotFound("job run not found")) {
				continue
			}
			return nil, fmt.Errorf("unable to retrieve job run: %w", err)
		}
		return &getVerifiedJobRunResponse{
			WorkflowExecution: run,
			NeosyncAccountId:  userAccount.Id,
			TemporalConfig:    tconfig,
		}, nil
	}
	return nil, nucleuserrors.NewNotFound("unable to find job run by run id")
}
