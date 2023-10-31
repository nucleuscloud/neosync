package v1alpha1_jobservice

import (
	"context"
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
		jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, accountUuid)
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

	runs := make([]*mgmtv1alpha1.JobRun, len(workflows))
	errGrp, errCtx := errgroup.WithContext(ctx)
	for index, workflow := range workflows {
		index := index
		workflow := workflow
		errGrp.Go(func() error {
			res, err := s.temporalClient.DescribeWorkflowExecution(errCtx, workflow.Execution.WorkflowId, workflow.Execution.RunId)
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
	run, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}
	res, err := s.temporalClient.DescribeWorkflowExecution(ctx, run.Execution.WorkflowId, run.Execution.RunId)
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
	run, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}

	activityOrder := []int64{}
	activityMap := map[int64]*mgmtv1alpha1.JobRunEvent{}
	iter := s.temporalClient.GetWorkflowHistory(
		ctx,
		run.Execution.WorkflowId,
		run.Execution.RunId,
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

		default:

		}
	}

	events := []*mgmtv1alpha1.JobRunEvent{}
	for _, index := range activityOrder {
		value := activityMap[index]
		events = append(events, value)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunEventsResponse{
		Events: events,
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

	scheduleHandle := s.temporalClient.ScheduleClient().GetHandle(ctx, nucleusdb.UUIDString(job.ID))
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
	run, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}

	logger.Info("canceling job run")
	err = s.temporalClient.CancelWorkflow(ctx, run.Execution.WorkflowId, run.Execution.RunId)
	if err != nil {
		logger.Error(fmt.Errorf("unable to cancel job run: %w", err).Error())
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.CancelJobRunResponse{}), nil
}

func (s *Service) DeleteJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRunRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.JobRunId)
	run, err := s.getVerifiedJobRun(ctx, logger, req.Msg.JobRunId)
	if err != nil {
		return nil, err
	}
	logger.Info("deleting job run")
	_, err = s.temporalClient.WorkflowService().DeleteWorkflowExecution(ctx, &workflowservice.DeleteWorkflowExecutionRequest{
		Namespace: s.cfg.TemporalNamespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: run.Execution.WorkflowId,
			RunId:      run.Execution.RunId,
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

func (s *Service) getVerifiedJobRun(
	ctx context.Context,
	logger *slog.Logger,
	runId string,
) (*workflowpb.WorkflowExecutionInfo, error) {
	run, err := getWorkflowExecutionsByRunId(ctx, s.temporalClient, s.cfg.TemporalNamespace, runId)
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job run: %w", err).Error())
		return nil, err
	}
	jobId := dtomaps.GetJobIdFromWorkflow(logger, run.GetSearchAttributes())
	jobUuid, err := nucleusdb.ToUuid(jobId)
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
	return run, nil
}
