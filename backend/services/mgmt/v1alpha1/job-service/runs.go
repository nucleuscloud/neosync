package v1alpha1_jobservice

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/internal/loki"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	piidetect_job_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/job/activities"
	piidetect_table_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table"
	piidetect_table_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table/activities"
	tablesync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/workflow"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (s *Service) GetJobRuns(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	var accountId string
	jobIds := []string{}
	switch id := req.Msg.Id.(type) {
	case *mgmtv1alpha1.GetJobRunsRequest_JobId:
		jobUuid, err := neosyncdb.ToUuid(id.JobId)
		if err != nil {
			return nil, err
		}
		job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
		if err != nil {
			return nil, err
		}

		accountId = neosyncdb.UUIDString(job.AccountID)
		jobIds = append(jobIds, id.JobId)
	case *mgmtv1alpha1.GetJobRunsRequest_AccountId:
		accountId = id.AccountId
		accountPgUuid, err := neosyncdb.ToUuid(accountId)
		if err != nil {
			return nil, err
		}
		jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, accountPgUuid)
		if err != nil {
			return nil, err
		}
		for i := range jobs {
			job := jobs[i]
			jobIds = append(jobIds, neosyncdb.UUIDString(job.ID))
		}
	default:
		return nil, fmt.Errorf("must provide jobId or accountId")
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(accountId), rbac.JobAction_View); err != nil {
		return nil, err
	}

	workflows, err := s.temporalmgr.GetWorkflowExecutionsByScheduleIds(
		ctx,
		accountId,
		jobIds,
		logger,
	)
	if err != nil {
		return nil, err
	}

	runs := make([]*mgmtv1alpha1.JobRun, len(workflows))
	for idx, workflow := range workflows {
		runs[idx] = dtomaps.ToJobRunDtoFromWorkflowExecutionInfo(workflow, logger)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
		JobRuns: runs,
	}), nil
}

func (s *Service) GetJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.JobRunId)

	res, err := s.temporalmgr.DescribeWorklowExecution(
		ctx,
		req.Msg.GetAccountId(),
		req.Msg.GetJobRunId(),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to describe workflow execution: %w", err)
	}

	dto := dtomaps.ToJobRunDto(logger, res)

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewDomainEntity(req.Msg.GetAccountId(), dto.GetJobId()), rbac.JobAction_View); err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{
		JobRun: dto,
	}), nil
}

func (s *Service) GetJobRunEvents(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunEventsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunEventsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.Msg.GetAccountId(), "jobRunId", req.Msg.GetJobRunId())

	jrResp, err := s.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{
		AccountId: req.Msg.GetAccountId(),
		JobRunId:  req.Msg.GetJobRunId(),
	}))
	if err != nil {
		return nil, err
	}

	jobRun := jrResp.Msg.GetJobRun()

	resp, err := s.getEventsByWorkflowId(ctx, req.Msg.GetAccountId(), jobRun.GetId(), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to get events by workflow id: %w", err)
	}

	return connect.NewResponse(resp), nil
}

func (s *Service) getEventsByWorkflowId(
	ctx context.Context,
	accountId, workflowId string,
	logger *slog.Logger,
) (*mgmtv1alpha1.GetJobRunEventsResponse, error) {
	isRunComplete := false
	activityOrder := []int64{}
	activityMap := map[int64]*mgmtv1alpha1.JobRunEvent{}
	iter, err := s.temporalmgr.GetWorkflowHistory(
		ctx,
		accountId,
		workflowId,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow history: %w", err)
	}
	childEvents := map[int64]string{} // eventId -> workflowId
	highestEventId := int64(0)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if event.GetEventId() > highestEventId {
			highestEventId = event.GetEventId()
		}

		switch event.EventType {
		case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			activityOrder = append(activityOrder, event.EventId)
			attributes := event.GetActivityTaskScheduledEventAttributes()
			jobRunEvent := &mgmtv1alpha1.JobRunEvent{
				Id:        event.EventId,
				Type:      attributes.ActivityType.Name,
				StartTime: event.EventTime,
				Tasks: []*mgmtv1alpha1.JobRunEventTask{
					dtomaps.ToJobRunEventTaskDto(event, nil),
				},
			}
			if len(attributes.Input.Payloads) > 1 {
				var rawMap map[string]string
				err := converter.GetDefaultDataConverter().
					FromPayload(attributes.Input.Payloads[1], &rawMap)
				if err != nil {
					logger.Error(
						fmt.Errorf("unable to convert to event input payload: %w", err).Error(),
					)
				}

				schema, schemaExists := rawMap["Schema"]
				table, tableExists := rawMap["Table"]

				metadata := &mgmtv1alpha1.JobRunEventMetadata{}

				if schemaExists && tableExists {
					metadata.Metadata = &mgmtv1alpha1.JobRunEventMetadata_SyncMetadata{
						SyncMetadata: &mgmtv1alpha1.JobRunSyncMetadata{
							Schema: schema,
							Table:  table,
						},
					}
				}

				jobRunEvent.Metadata = metadata
			}
			activityMap[event.EventId] = jobRunEvent
		case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
			attributes := event.GetActivityTaskStartedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.StartTime = event.EventTime
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
			attributes := event.GetActivityTaskCompletedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.CloseTime = event.EventTime
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))
		case enums.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED:
			attributes := event.GetActivityTaskCancelRequestedEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
			attributes := event.GetActivityTaskCanceledEventAttributes()
			activity := activityMap[attributes.ScheduledEventId]
			activity.CloseTime = event.EventTime
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
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
			isRunComplete = true
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
			isRunComplete = true

		case enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_FAILED:
			isRunComplete = true
		case enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED:
			activityOrder = append(activityOrder, event.GetEventId())
			attributes := event.GetStartChildWorkflowExecutionInitiatedEventAttributes()

			childEvents[event.GetEventId()] = attributes.GetWorkflowId()

			jobRunEvent := &mgmtv1alpha1.JobRunEvent{
				Id:        event.EventId,
				Type:      attributes.GetWorkflowType().GetName(),
				StartTime: event.EventTime,
				Tasks: []*mgmtv1alpha1.JobRunEventTask{
					dtomaps.ToJobRunEventTaskDto(event, nil),
				},
			}
			if len(attributes.Input.Payloads) > 0 {
				switch attributes.GetWorkflowType().GetName() {
				case "TableSync":
					var tableSyncRequest tablesync_workflow.TableSyncRequest
					err := converter.GetDefaultDataConverter().
						FromPayload(attributes.Input.Payloads[0], &tableSyncRequest)
					if err != nil {
						logger.Error(
							fmt.Errorf("unable to convert to event input payload: %w", err).Error(),
						)
					}

					metadata := &mgmtv1alpha1.JobRunEventMetadata{}
					metadata.Metadata = &mgmtv1alpha1.JobRunEventMetadata_SyncMetadata{
						SyncMetadata: &mgmtv1alpha1.JobRunSyncMetadata{
							Schema: tableSyncRequest.TableSchema,
							Table:  tableSyncRequest.TableName,
						},
					}
					jobRunEvent.Metadata = metadata
				case "TablePiiDetect":
					var piiDetectTableRequest piidetect_table_workflow.TablePiiDetectRequest
					err := converter.GetDefaultDataConverter().
						FromPayload(attributes.Input.Payloads[0], &piiDetectTableRequest)
					if err != nil {
						logger.Error(
							fmt.Errorf("unable to convert to event input payload: %w", err).Error(),
						)
					}
					metadata := &mgmtv1alpha1.JobRunEventMetadata{}
					metadata.Metadata = &mgmtv1alpha1.JobRunEventMetadata_SyncMetadata{
						SyncMetadata: &mgmtv1alpha1.JobRunSyncMetadata{
							Schema: piiDetectTableRequest.TableSchema,
							Table:  piiDetectTableRequest.TableName,
						},
					}
					jobRunEvent.Metadata = metadata
				}
			}
			activityMap[event.EventId] = jobRunEvent

		case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
			attributes := event.GetChildWorkflowExecutionStartedEventAttributes()
			activity := activityMap[attributes.InitiatedEventId]
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
			attributes := event.GetChildWorkflowExecutionCompletedEventAttributes()
			activity := activityMap[attributes.InitiatedEventId]
			activity.CloseTime = event.EventTime
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))

		case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
			attributes := event.GetChildWorkflowExecutionFailedEventAttributes()
			activity := activityMap[attributes.InitiatedEventId]
			activity.CloseTime = event.EventTime
			errorDto := dtomaps.ToJobRunEventTaskErrorDto(attributes.Failure, attributes.RetryState)
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, errorDto))

			isRunComplete = true

		case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
			attributes := event.GetChildWorkflowExecutionTimedOutEventAttributes()
			activity := activityMap[attributes.InitiatedEventId]
			activity.CloseTime = event.EventTime
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))
			isRunComplete = true

		case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
			attributes := event.GetChildWorkflowExecutionCanceledEventAttributes()
			activity := activityMap[attributes.InitiatedEventId]
			activity.CloseTime = event.EventTime
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))
			isRunComplete = true

		case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED:
			attributes := event.GetChildWorkflowExecutionTerminatedEventAttributes()
			activity := activityMap[attributes.InitiatedEventId]
			activity.CloseTime = event.EventTime
			activity.Tasks = append(activity.Tasks, dtomaps.ToJobRunEventTaskDto(event, nil))
			isRunComplete = true
		default:
		}
	}

	if isRunComplete && len(childEvents) > 0 {
		logger.Debug("checking if child workflows are complete")
		for _, eventId := range activityOrder {
			childWorkflowId, ok := childEvents[eventId]
			if !ok {
				continue
			}

			childEvent, ok := activityMap[eventId]
			if !ok {
				continue
			}

			if childEvent.GetCloseTime() != nil {
				continue
			}
			logger.Debug("child workflow is not complete, checking if it is closed")
			info, err := s.temporalmgr.GetWorkflowExecutionById(
				ctx,
				accountId,
				childWorkflowId,
				logger,
			)
			if err != nil {
				logger.Warn(
					fmt.Sprintf(
						"unable to get workflow execution info for %s: %s",
						childWorkflowId,
						err,
					),
				)
				continue
			}

			if info.GetCloseTime() != nil {
				logger.Debug("child workflow is closed, updating event")
				childEvent.CloseTime = info.GetCloseTime()
			}

			switch info.GetStatus() {
			case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
				highestEventId++
				childEvent.Tasks = append(
					childEvent.Tasks,
					dtomaps.ToJobRunEventTaskDto(&history.HistoryEvent{
						EventId:    highestEventId,
						EventTime:  info.GetCloseTime(),
						EventType:  enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED,
						Attributes: nil,
					}, nil),
				)
			case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
				resp, err := s.getEventsByWorkflowId(ctx, accountId, childWorkflowId, logger)
				if err != nil {
					logger.Warn(
						fmt.Sprintf(
							"unable to get events by workflow id for %s: %s",
							childWorkflowId,
							err,
						),
					)
					continue
				}
				var eventErr *mgmtv1alpha1.JobRunEventTaskError
				for _, event := range resp.GetEvents() {
					for _, task := range event.GetTasks() {
						if task.GetError() != nil {
							eventErr = task.GetError()
							break
						}
					}
				}
				highestEventId++
				childEvent.Tasks = append(
					childEvent.Tasks,
					dtomaps.ToJobRunEventTaskDto(&history.HistoryEvent{
						EventId:    highestEventId,
						EventTime:  info.GetCloseTime(),
						EventType:  enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED,
						Attributes: nil,
					}, eventErr),
				)
			case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
				highestEventId++
				childEvent.Tasks = append(
					childEvent.Tasks,
					dtomaps.ToJobRunEventTaskDto(&history.HistoryEvent{
						EventId:    highestEventId,
						EventTime:  info.GetCloseTime(),
						EventType:  enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT,
						Attributes: nil,
					}, nil),
				)
			case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
				highestEventId++
				childEvent.Tasks = append(
					childEvent.Tasks,
					dtomaps.ToJobRunEventTaskDto(&history.HistoryEvent{
						EventId:    highestEventId,
						EventTime:  info.GetCloseTime(),
						EventType:  enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED,
						Attributes: nil,
					}, nil),
				)
			case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
				highestEventId++
				childEvent.Tasks = append(
					childEvent.Tasks,
					dtomaps.ToJobRunEventTaskDto(&history.HistoryEvent{
						EventId:    highestEventId,
						EventTime:  info.GetCloseTime(),
						EventType:  enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED,
						Attributes: nil,
					}, nil),
				)
			}
		}
	}

	events := []*mgmtv1alpha1.JobRunEvent{}
	for _, index := range activityOrder {
		value := activityMap[index]
		events = append(events, value)
	}

	return &mgmtv1alpha1.GetJobRunEventsResponse{Events: events, IsRunComplete: isRunComplete}, nil
}

func (s *Service) CreateJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.GetJobId())
	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.GetJobId(),
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewDomainEntity(job.GetAccountId(), job.GetId()), rbac.JobAction_Execute); err != nil {
		return nil, err
	}

	logger.Debug("creating job run by triggering temporal schedule")
	err = s.temporalmgr.TriggerSchedule(
		ctx,
		job.GetAccountId(),
		job.GetId(),
		&temporalclient.ScheduleTriggerOptions{},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create job run by triggering temporal schedule: %w", err)
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateJobRunResponse{}), nil
}

func (s *Service) CancelJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CancelJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CancelJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With(
		"accountId", req.Msg.GetAccountId(),
		"jobRunId", req.Msg.GetJobRunId(),
	)

	jobRunResp, err := s.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{
		AccountId: req.Msg.GetAccountId(),
		JobRunId:  req.Msg.GetJobRunId(),
	}))
	if err != nil {
		return nil, err
	}
	jobRun := jobRunResp.Msg.GetJobRun()
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewDomainEntity(req.Msg.GetAccountId(), jobRun.GetJobId()), rbac.JobAction_Execute); err != nil {
		return nil, err
	}

	err = s.temporalmgr.CancelWorkflow(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to cancel job run: %w", err)
	}
	return connect.NewResponse(&mgmtv1alpha1.CancelJobRunResponse{}), nil
}

func (s *Service) TerminateJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.TerminateJobRunRequest],
) (*connect.Response[mgmtv1alpha1.TerminateJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With(
		"accountId", req.Msg.GetAccountId(),
		"jobRunId", req.Msg.GetJobRunId(),
	)
	jobRunResp, err := s.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{
		AccountId: req.Msg.GetAccountId(),
		JobRunId:  req.Msg.GetJobRunId(),
	}))
	if err != nil {
		return nil, err
	}
	jobRun := jobRunResp.Msg.GetJobRun()
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewDomainEntity(req.Msg.GetAccountId(), jobRun.GetJobId()), rbac.JobAction_Execute); err != nil {
		return nil, err
	}

	err = s.temporalmgr.TerminateWorkflow(
		ctx,
		req.Msg.GetAccountId(),
		req.Msg.GetJobRunId(),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to terminate job run: %w", err)
	}
	return connect.NewResponse(&mgmtv1alpha1.TerminateJobRunResponse{}), nil
}

func (s *Service) DeleteJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRunRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With(
		"accountId", req.Msg.GetAccountId(),
		"jobRunId", req.Msg.GetJobRunId(),
	)
	jobRunResp, err := s.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{
		AccountId: req.Msg.GetAccountId(),
		JobRunId:  req.Msg.GetJobRunId(),
	}))
	if err != nil {
		return nil, err
	}
	jobRun := jobRunResp.Msg.GetJobRun()
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewDomainEntity(req.Msg.GetAccountId(), jobRun.GetJobId()), rbac.JobAction_Delete); err != nil {
		return nil, err
	}

	err = s.temporalmgr.DeleteWorkflowExecution(
		ctx,
		req.Msg.GetAccountId(),
		req.Msg.GetJobRunId(),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to delete job run: %w", err)
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobRunResponse{}), nil
}

type LogLine struct {
	WorkflowID string     `json:"WorkflowID"`
	Time       *time.Time `json:"time,omitempty"`
	Level      string     `json:"level"`
}

func (s *Service) GetJobRunLogsStream(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunLogsStreamRequest],
	stream *connect.ServerStream[mgmtv1alpha1.GetJobRunLogsStreamResponse],
) error {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.GetJobRunId())

	onLogLine := func(logline *mgmtv1alpha1.GetJobRunLogsResponse_LogLine) error {
		return stream.Send(
			&mgmtv1alpha1.GetJobRunLogsStreamResponse{
				LogLine:   logline.LogLine,
				Timestamp: logline.Timestamp,
			},
		)
	}
	return s.streamLogs(ctx, req.Msg, &logLineStreamer{onLogLine: onLogLine}, logger)
}

type logLineStreamer struct {
	onLogLine func(logline *mgmtv1alpha1.GetJobRunLogsResponse_LogLine) error
}

func (s *logLineStreamer) Send(logline *mgmtv1alpha1.GetJobRunLogsResponse_LogLine) error {
	return s.onLogLine(logline)
}

func (s *Service) GetJobRunLogs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunLogsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunLogsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.GetJobRunId())

	loglines := []*mgmtv1alpha1.GetJobRunLogsResponse_LogLine{}
	onLogLine := func(logline *mgmtv1alpha1.GetJobRunLogsResponse_LogLine) error {
		loglines = append(loglines, logline)
		return nil
	}

	err := s.streamLogs(
		ctx,
		&unaryLogStreamRequest{GetJobRunLogsRequest: req.Msg},
		&logLineStreamer{onLogLine: onLogLine},
		logger,
	)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.GetJobRunLogsResponse{LogLines: loglines}), nil
}

func (s *Service) streamLogs(
	ctx context.Context,
	req logStreamRequest,
	stream logStreamer,
	logger *slog.Logger,
) error {
	if s.cfg.RunLogConfig == nil || !s.cfg.RunLogConfig.IsEnabled ||
		s.cfg.RunLogConfig.RunLogType == nil {
		return nucleuserrors.NewNotImplemented(
			"job run logs is not enabled. please configure or contact system administrator to enable logs.",
		)
	}

	jobRunResp, err := s.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{
		AccountId: req.GetAccountId(),
		JobRunId:  req.GetJobRunId(),
	}))
	if err != nil {
		return err
	}
	jobRun := jobRunResp.Msg.GetJobRun()
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return err
	}
	if err := user.EnforceJob(ctx, userdata.NewDomainEntity(req.GetAccountId(), jobRun.GetJobId()), rbac.JobAction_View); err != nil {
		return err
	}

	switch *s.cfg.RunLogConfig.RunLogType {
	case KubePodRunLogType:
		err := s.streamK8sWorkerPodLogs(ctx, req, stream, logger)
		if err != nil {
			return err
		}
		return nil
	case LokiRunLogType:
		err := s.streamLokiWorkerLogs(ctx, req, stream, logger)
		if err != nil {
			return err
		}
		return nil
	default:
		return nucleuserrors.NewNotImplemented(
			"streaming log pods not implemented for this container type",
		)
	}
}

type logStreamer interface {
	Send(logline *mgmtv1alpha1.GetJobRunLogsResponse_LogLine) error
}

type logStreamRequest interface {
	GetAccountId() string
	GetJobRunId() string
	GetLogLevels() []mgmtv1alpha1.LogLevel
	GetWindow() mgmtv1alpha1.LogWindow
	GetMaxLogLines() int64
	GetShouldTail() bool
}

type unaryLogStreamRequest struct {
	*mgmtv1alpha1.GetJobRunLogsRequest
}

func (r *unaryLogStreamRequest) GetShouldTail() bool {
	return false
}

func (s *Service) streamK8sWorkerPodLogs(
	ctx context.Context,
	req logStreamRequest,
	stream logStreamer,
	logger *slog.Logger,
) error {
	if s.cfg.RunLogConfig.RunLogPodConfig == nil {
		return nucleuserrors.NewInternalError("run logs configured but no config provided")
	}
	workflowExecution, err := s.temporalmgr.GetWorkflowExecutionById(
		ctx,
		req.GetAccountId(),
		req.GetJobRunId(),
		logger,
	)
	if err != nil {
		return err
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to retrieve k8s in cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to create kubernetes clientset: %w", err)
	}

	appNameSelector, err := labels.NewRequirement(
		"app",
		selection.Equals,
		[]string{s.cfg.RunLogConfig.RunLogPodConfig.WorkerAppName},
	)
	if err != nil {
		return fmt.Errorf("unable to build label selector when finding k8s logs: %w", err)
	}
	podclient := clientset.CoreV1().Pods(s.cfg.RunLogConfig.RunLogPodConfig.Namespace)
	pods, err := podclient.List(ctx, metav1.ListOptions{
		LabelSelector: appNameSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("unable to retrieve list of pods from k8s: %w", err)
	}

	loglevels := getLogLevelFilters(req.GetLogLevels())
	uniqueloglevels := map[string]any{}
	for _, ll := range loglevels {
		uniqueloglevels[ll] = struct{}{}
	}
	var maxLogLints *int64
	if req.GetMaxLogLines() > 0 {
		maxLines := req.GetMaxLogLines()
		maxLogLints = &maxLines
	}
	for idx := range pods.Items {
		pod := pods.Items[idx]
		logsReq := podclient.GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: "user-container",
			Follow:    req.GetShouldTail(),
			TailLines: maxLogLints,
			SinceTime: &metav1.Time{Time: getLogFilterTime(req.GetWindow(), time.Now())},
		})
		logstream, err := logsReq.Stream(ctx)
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		} else if err != nil && k8serrors.IsNotFound(err) {
			return nucleuserrors.NewNotFound("pod no longer exists")
		}

		scanner := bufio.NewScanner(logstream)

		for scanner.Scan() {
			txt := scanner.Text()
			var logLine LogLine
			err := json.Unmarshal([]byte(txt), &logLine)
			if err != nil {
				logger.Error(fmt.Sprintf("error unmarshaling log line: %v\n", err))
				continue // Skip lines that can't be unmarshaled
			}

			if logLine.WorkflowID == workflowExecution.GetExecution().GetWorkflowId() {
				if _, ok := uniqueloglevels[logLine.Level]; !ok && len(uniqueloglevels) > 0 {
					continue
				}
				var timestamp *timestamppb.Timestamp
				if logLine.Time != nil {
					timestamp = timestamppb.New(*logLine.Time)
				}
				if err := stream.Send(&mgmtv1alpha1.GetJobRunLogsResponse_LogLine{LogLine: txt, Labels: map[string]string{}, Timestamp: timestamp}); err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
			}
		}
		logstream.Close()
	}
	return nil
}

func (s *Service) streamLokiWorkerLogs(
	ctx context.Context,
	req logStreamRequest,
	stream logStreamer,
	logger *slog.Logger,
) error {
	if s.cfg.RunLogConfig == nil || !s.cfg.RunLogConfig.IsEnabled ||
		s.cfg.RunLogConfig.LokiRunLogConfig == nil {
		return nucleuserrors.NewInternalError("run logs configured but no config provided")
	}
	if s.cfg.RunLogConfig.LokiRunLogConfig.LabelsQuery == "" {
		return nucleuserrors.NewInternalError("must provide a labels query for loki to filter by")
	}
	workflowExecution, err := s.temporalmgr.GetWorkflowExecutionById(
		ctx,
		req.GetAccountId(),
		req.GetJobRunId(),
		logger,
	)
	if err != nil {
		return fmt.Errorf("unable to retrieve workflow execution: %w", err)
	}

	lokiclient := loki.New(s.cfg.RunLogConfig.LokiRunLogConfig.BaseUrl, http.DefaultClient)
	direction := loki.BACKWARD
	end := time.Now()
	if workflowExecution.CloseTime != nil {
		end = workflowExecution.CloseTime.AsTime()
	}
	start := getLogFilterTime(req.GetWindow(), end)
	query := buildLokiQuery(
		s.cfg.RunLogConfig.LokiRunLogConfig.LabelsQuery,
		s.cfg.RunLogConfig.LokiRunLogConfig.KeepLabels,
		workflowExecution.GetExecution().GetWorkflowId(),
		getLogLevelFilters(req.GetLogLevels()),
	)

	var maxLogLints *int64
	if req.GetMaxLogLines() > 0 {
		maxLines := req.GetMaxLogLines()
		maxLogLints = &maxLines
	}
	resp, err := lokiclient.QueryRange(ctx, &loki.QueryRangeRequest{
		Query: query,
		Limit: maxLogLints,

		Direction: &direction,
		Start:     &start,
		End:       &end,
	}, logger)
	if err != nil {
		logger.Error("failed to query loki", "query", query)
		return fmt.Errorf("unable to query loki for logs: %w", err)
	}
	if resp.Status != "success" {
		return fmt.Errorf("received non-success status response from loki: %s", resp.Status)
	}
	streams, err := loki.GetStreamsFromResponseData(&resp.Data)
	if err != nil {
		return err
	}
	entries := loki.GetEntriesFromStreams(streams)
	// Loki logs have issues with ordering, so we need to sort them manually
	// Issue: https://github.com/grafana/loki/issues/13295
	if direction == loki.BACKWARD {
		sort.Slice(entries, func(i, j int) bool {
			// sorts in descending order
			return entries[i].Timestamp.After(entries[j].Timestamp)
		})
	} else {
		sort.Slice(entries, func(i, j int) bool {
			// sorts in ascending order
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})
	}

	for _, entry := range entries {
		err := stream.Send(
			&mgmtv1alpha1.GetJobRunLogsResponse_LogLine{
				LogLine:   entry.Line,
				Labels:    entry.Labels.Map(),
				Timestamp: timestamppb.New(entry.Timestamp),
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLogLevelFilters(loglevels []mgmtv1alpha1.LogLevel) []string {
	levels := []string{}

	for _, ll := range loglevels {
		if ll == mgmtv1alpha1.LogLevel_LOG_LEVEL_UNSPECIFIED {
			return []string{}
		}
		llstr := logLevelToString(ll)
		if llstr == "" {
			continue
		}
		levels = append(levels, llstr)
	}
	return levels
}

func logLevelToString(loglevel mgmtv1alpha1.LogLevel) string {
	switch loglevel {
	case mgmtv1alpha1.LogLevel_LOG_LEVEL_DEBUG:
		return "DEBUG"
	case mgmtv1alpha1.LogLevel_LOG_LEVEL_ERROR:
		return "ERROR"
	case mgmtv1alpha1.LogLevel_LOG_LEVEL_INFO:
		return "INFO"
	case mgmtv1alpha1.LogLevel_LOG_LEVEL_WARN:
		return "WARN"
	default:
		return ""
	}
}

func buildLokiQuery(
	lokiLables string,
	keep []string,
	workflowId string,
	loglevels []string,
) string {
	query := fmt.Sprintf("{%s} | json", lokiLables)
	query = fmt.Sprintf("%s | JobRunId=%q", query, workflowId)

	if len(loglevels) > 0 {
		query = fmt.Sprintf("%s | level=~%q", query, strings.Join(loglevels, "|"))
	}

	query = fmt.Sprintf("%s | line_format %q", query, "[{{.level}}] - {{.msg}}")

	if len(keep) > 0 {
		query = fmt.Sprintf("%s | keep %s", query, strings.Join(keep, ", "))
	}
	return query
}

func getLogFilterTime(window mgmtv1alpha1.LogWindow, endTime time.Time) time.Time {
	switch window {
	case mgmtv1alpha1.LogWindow_LOG_WINDOW_FIFTEEN_MIN:
		return endTime.Add(-15 * time.Minute)
	case mgmtv1alpha1.LogWindow_LOG_WINDOW_ONE_HOUR:
		return endTime.Add(-1 * time.Hour)
	case mgmtv1alpha1.LogWindow_LOG_WINDOW_ONE_DAY:
		return endTime.Add(-24 * time.Hour)
	default:
		return endTime.Add(-15 * time.Minute)
	}
}

func (s *Service) GetRunContext(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetRunContextRequest],
) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
	id := req.Msg.GetId()

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(id.GetAccountId()), rbac.JobAction_View); err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(id.GetAccountId())
	if err != nil {
		return nil, err
	}

	runContext, err := s.db.Q.GetRunContextByKey(ctx, s.db.Db, db_queries.GetRunContextByKeyParams{
		WorkflowId: id.GetJobRunId(),
		ExternalId: id.GetExternalId(),
		AccountId:  accountUuid,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to retrieve run context by key: %w", err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("no run context exists with the provided key")
	}

	return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
		Value: runContext.Value,
	}), nil
}

func (s *Service) SetRunContext(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetRunContextRequest],
) (*connect.Response[mgmtv1alpha1.SetRunContextResponse], error) {
	id := req.Msg.GetId()

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(id.GetAccountId()), rbac.JobAction_Edit); err != nil {
		return nil, err
	}

	if s.cfg.IsNeosyncCloud && !user.IsWorkerApiKey() {
		return nil, nucleuserrors.NewUnauthenticated(
			"must provide valid authentication credentials for this endpoint",
		)
	}

	accountUuid, err := neosyncdb.ToUuid(id.GetAccountId())
	if err != nil {
		return nil, err
	}

	err = s.db.Q.SetRunContext(ctx, s.db.Db, db_queries.SetRunContextParams{
		WorkflowID:  id.GetJobRunId(),
		ExternalID:  id.GetExternalId(),
		AccountID:   accountUuid,
		Value:       req.Msg.GetValue(),
		CreatedByID: user.PgId(),
		UpdatedByID: user.PgId(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to set run context: %w", err)
	}

	return connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil
}

func (s *Service) SetRunContexts(
	ctx context.Context,
	stream *connect.ClientStream[mgmtv1alpha1.SetRunContextsRequest],
) (*connect.Response[mgmtv1alpha1.SetRunContextsResponse], error) {
	for stream.Receive() {
		req := stream.Msg()
		id := req.GetId()
		user, err := s.userdataclient.GetUser(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get user: %w", err)
		}
		if err := user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(id.GetAccountId()), rbac.JobAction_Edit); err != nil {
			return nil, err
		}

		if s.cfg.IsNeosyncCloud && !user.IsWorkerApiKey() {
			return nil, nucleuserrors.NewUnauthenticated(
				"must provide valid authentication credentials for this endpoint",
			)
		}

		accountUuid, err := neosyncdb.ToUuid(id.GetAccountId())
		if err != nil {
			return nil, err
		}

		err = s.db.Q.SetRunContext(ctx, s.db.Db, db_queries.SetRunContextParams{
			WorkflowID:  id.GetJobRunId(),
			ExternalID:  id.GetExternalId(),
			AccountID:   accountUuid,
			Value:       req.GetValue(),
			CreatedByID: user.PgId(),
			UpdatedByID: user.PgId(),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to set run context: %w", err)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	return connect.NewResponse(&mgmtv1alpha1.SetRunContextsResponse{}), nil
}

func (s *Service) GetPiiDetectionReport(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetPiiDetectionReportRequest],
) (*connect.Response[mgmtv1alpha1.GetPiiDetectionReportResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.Msg.GetAccountId(), "jobRunId", req.Msg.GetJobRunId())

	jrResp, err := s.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{
		AccountId: req.Msg.GetAccountId(),
		JobRunId:  req.Msg.GetJobRunId(),
	}))
	if err != nil {
		return nil, err
	}

	jobRun := jrResp.Msg.GetJobRun()

	logger.Debug("building pii detection report")

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	tableRunContexts, err := s.getTableRunContextsFromJobReport(ctx, jobRun, accountUuid)
	if err != nil {
		return nil, fmt.Errorf("unable to get table run contexts from job report: %w", err)
	}

	// this can happen if the run is currently running and the job report hasn't been updated yet
	// this allows us to effectively stream in the latest reports while the job is running
	if len(tableRunContexts) == 0 {
		logger.Debug("no table run contexts found in job report, fetching table level reports")
		runContexts, err := s.db.Q.GetRunContextsByExternalIdSuffix(
			ctx,
			s.db.Db,
			db_queries.GetRunContextsByExternalIdSuffixParams{
				WorkflowId:       jobRun.GetId(),
				ExternalIdSuffix: piidetect_table_activities.PiiTableReportSuffix,
				AccountId:        accountUuid,
			},
		)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return nil, fmt.Errorf("unable to retrieve run contexts: %w", err)
		}

		for i := range runContexts {
			tableRunContexts = append(tableRunContexts, &runContexts[i])
		}
	}

	// Sort tableRunContexts by UpdatedAt in descending order (newest first)
	sort.Slice(tableRunContexts, func(i, j int) bool {
		// Compare timestamps, return true if i should come before j
		return tableRunContexts[i].UpdatedAt.Time.After(tableRunContexts[j].UpdatedAt.Time)
	})

	reports, err := getReportsFromTableContexts(tableRunContexts)
	if err != nil {
		return nil, fmt.Errorf("unable to get reports from table contexts: %w", err)
	}

	reportDtos := getTableReportDtos(reports)

	logger.Debug("got report dtos", "count", len(reportDtos))

	return connect.NewResponse(&mgmtv1alpha1.GetPiiDetectionReportResponse{
		Report: &mgmtv1alpha1.PiiDetectionReport{
			Tables: reportDtos,
		},
	}), nil
}

func (s *Service) getTableRunContextsFromJobReport(
	ctx context.Context,
	jobRun *mgmtv1alpha1.JobRun,
	accountUuid pgtype.UUID,
) ([]*db_queries.NeosyncApiRuncontext, error) {
	runContext, err := s.db.Q.GetRunContextByKey(ctx, s.db.Db, db_queries.GetRunContextByKeyParams{
		WorkflowId: jobRun.GetId(),
		ExternalId: piidetect_job_activities.BuildJobReportExternalId(jobRun.GetJobId()),
		AccountId:  accountUuid,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to retrieve run context: %w", err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nil
	}
	var jobReport piidetect_job_activities.JobPiiDetectReport
	err = json.Unmarshal(runContext.Value, &jobReport)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal run context for job pii detect report: %w", err)
	}
	tableRunContextKeys := make(
		[]*mgmtv1alpha1.RunContextKey,
		0,
		len(jobReport.SuccessfulTableReports),
	)
	for _, tableReport := range jobReport.SuccessfulTableReports {
		tableRunContextKeys = append(tableRunContextKeys, tableReport.ReportKey)
	}
	tableRunContexts, err := s.getDbRunContextsFromKeys(ctx, tableRunContextKeys)
	if err != nil {
		return nil, fmt.Errorf("unable to get table run contexts from job report: %w", err)
	}
	return tableRunContexts, nil
}

func (s *Service) getDbRunContextsFromKeys(
	ctx context.Context,
	keys []*mgmtv1alpha1.RunContextKey,
) ([]*db_queries.NeosyncApiRuncontext, error) {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(10)
	runContexts := []*db_queries.NeosyncApiRuncontext{}
	mu := sync.Mutex{}
	// this could be further optimized by fetching all the run contexts in a single query
	// where the account id and workflow id are the same
	// there are cases where they may not be the same because of incremental syncs.
	// run contexts for incremental syncs will contain workflow ids that are outside of the current run (due to being saved from a previous run)
	for _, key := range keys {
		key := key
		errgrp.Go(func() error {
			accountUuid, err := neosyncdb.ToUuid(key.GetAccountId())
			if err != nil {
				return fmt.Errorf("unable to convert account id to uuid: %w", err)
			}
			runContext, err := s.db.Q.GetRunContextByKey(
				errctx,
				s.db.Db,
				db_queries.GetRunContextByKeyParams{
					WorkflowId: key.GetJobRunId(),
					ExternalId: key.GetExternalId(),
					AccountId:  accountUuid,
				},
			)
			if err != nil {
				return fmt.Errorf("unable to get run context: %w", err)
			}
			mu.Lock()
			defer mu.Unlock()
			runContexts = append(runContexts, &runContext)
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}
	return runContexts, nil
}

func getReportsFromTableContexts(
	tableContexts []*db_queries.NeosyncApiRuncontext,
) ([]*piidetect_table_activities.TableReport, error) {
	reports := make([]*piidetect_table_activities.TableReport, len(tableContexts))
	for i := range tableContexts {
		runContext := tableContexts[i]
		var report *piidetect_table_activities.TableReport
		err := json.Unmarshal(runContext.Value, &report)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal run context: %w", err)
		}
		reports[i] = report
	}
	return reports, nil
}

func getTableReportDtos(
	reports []*piidetect_table_activities.TableReport,
) []*mgmtv1alpha1.PiiDetectionReport_TableReport {
	reportDtos := make([]*mgmtv1alpha1.PiiDetectionReport_TableReport, len(reports))
	for i, report := range reports {
		reportDtos[i] = &mgmtv1alpha1.PiiDetectionReport_TableReport{
			Schema: report.TableSchema,
			Table:  report.TableName,
			Columns: make(
				[]*mgmtv1alpha1.PiiDetectionReport_TableReport_ColumnReport,
				0,
				len(report.ColumnReports),
			),
		}
		for _, columnReport := range report.ColumnReports {
			columnReportDto := &mgmtv1alpha1.PiiDetectionReport_TableReport_ColumnReport{
				Column: columnReport.ColumnName,
			}
			if columnReport.Report.Regex != nil {
				columnReportDto.RegexReport = &mgmtv1alpha1.PiiDetectionReport_TableReport_ColumnReport_Regex{
					Category: columnReport.Report.Regex.Category.String(),
				}
			}
			if columnReport.Report.LLM != nil {
				columnReportDto.LlmReport = &mgmtv1alpha1.PiiDetectionReport_TableReport_ColumnReport_LLM{
					Category:   string(columnReport.Report.LLM.Category),
					Confidence: columnReport.Report.LLM.Confidence,
				}
			}
			reportDtos[i].Columns = append(reportDtos[i].Columns, columnReportDto)
		}
	}
	return reportDtos
}
