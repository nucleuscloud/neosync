package v1alpha1_jobservice

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/loki"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"go.temporal.io/api/enums/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
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

	_, err := s.verifyUserInAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}

	workflows, err := s.temporalmgr.GetWorkflowExecutionsByScheduleIds(ctx, accountId, jobIds, logger)
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
	res, err := s.temporalmgr.DescribeWorklowExecution(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
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
	logger = logger.With("accountId", req.Msg.GetAccountId(), "jobRunId", req.Msg.GetJobRunId())
	isRunComplete := false
	activityOrder := []int64{}
	activityMap := map[int64]*mgmtv1alpha1.JobRunEvent{}
	iter, err := s.temporalmgr.GetWorkflowHistory(
		ctx,
		req.Msg.GetAccountId(),
		req.Msg.GetJobRunId(),
		logger,
	)
	if err != nil {
		return nil, err
	}
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
				StartTime: event.EventTime,
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil {
		return nil, err
	}
	accountId := neosyncdb.UUIDString(job.AccountID)
	_, err = s.verifyUserInAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}

	logger.Info("creating job run")
	err = s.temporalmgr.TriggerSchedule(
		ctx,
		neosyncdb.UUIDString(job.AccountID),
		neosyncdb.UUIDString(job.ID),
		&temporalclient.ScheduleTriggerOptions{},
		logger,
	)
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
	logger = logger.With(
		"accountId", req.Msg.GetAccountId(),
		"jobRunId", req.Msg.GetJobRunId(),
	)
	err := s.temporalmgr.CancelWorkflow(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
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
	err := s.temporalmgr.TerminateWorkflow(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
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
	err := s.temporalmgr.DeleteWorkflowExecution(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
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
	logger = logger.With("jobRunId", req.Msg.JobRunId)

	if s.cfg.RunLogConfig == nil || !s.cfg.RunLogConfig.IsEnabled || s.cfg.RunLogConfig.RunLogType == nil {
		return nucleuserrors.NewNotImplemented("job run logs streaming is not enabled. please configure or contact system administrator to enable logs.")
	}

	switch *s.cfg.RunLogConfig.RunLogType {
	case KubePodRunLogType:
		return s.streamK8sWorkerPodLogs(ctx, req, stream, logger)
	case LokiRunLogType:
		return s.streamLokiWorkerLogs(ctx, req, stream, logger)
	default:
		return nucleuserrors.NewNotImplemented("streaming log pods not implemented for this container type")
	}
}

func (s *Service) streamK8sWorkerPodLogs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunLogsStreamRequest],
	stream *connect.ServerStream[mgmtv1alpha1.GetJobRunLogsStreamResponse],
	logger *slog.Logger,
) error {
	if s.cfg.RunLogConfig.RunLogPodConfig == nil {
		return nucleuserrors.NewInternalError("run logs configured but no config provided")
	}
	workflowExecution, err := s.temporalmgr.GetWorkflowExecutionById(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
	if err != nil {
		return err
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Error(fmt.Errorf("error getting kubernetes config: %w", err).Error())
		return err
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Error(fmt.Errorf("error getting kubernetes clientset: %w", err).Error())
		return err
	}

	appNameSelector, err := labels.NewRequirement("app", selection.Equals, []string{s.cfg.RunLogConfig.RunLogPodConfig.WorkerAppName})
	if err != nil {
		logger.Error(fmt.Errorf("unable to build label selector to find logs: %w", err).Error())
		return err
	}
	podclient := clientset.CoreV1().Pods(s.cfg.RunLogConfig.RunLogPodConfig.Namespace)
	pods, err := podclient.List(ctx, metav1.ListOptions{
		LabelSelector: appNameSelector.String(),
	})
	if err != nil {
		logger.Error(fmt.Errorf("error getting pods: %w", err).Error())
		return err
	}

	loglevels := getLogLevelFilters(req.Msg.GetLogLevels())
	uniqueloglevels := map[string]any{}
	for _, ll := range loglevels {
		uniqueloglevels[ll] = struct{}{}
	}

	for idx := range pods.Items {
		pod := pods.Items[idx]
		logsReq := podclient.GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: "user-container",
			Follow:    req.Msg.ShouldTail,
			TailLines: req.Msg.MaxLogLines,
			SinceTime: &metav1.Time{Time: getLogFilterTime(req.Msg.GetWindow(), time.Now())},
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
				if err := stream.Send(&mgmtv1alpha1.GetJobRunLogsStreamResponse{LogLine: txt, Timestamp: timestamp}); err != nil {
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
	req *connect.Request[mgmtv1alpha1.GetJobRunLogsStreamRequest],
	stream *connect.ServerStream[mgmtv1alpha1.GetJobRunLogsStreamResponse],
	logger *slog.Logger,
) error {
	if s.cfg.RunLogConfig == nil || !s.cfg.RunLogConfig.IsEnabled || s.cfg.RunLogConfig.LokiRunLogConfig == nil {
		return nucleuserrors.NewInternalError("run logs configured but no config provided")
	}
	if s.cfg.RunLogConfig.LokiRunLogConfig.LabelsQuery == "" {
		return nucleuserrors.NewInternalError("must provide a labels query for loki to filter by")
	}
	workflowExecution, err := s.temporalmgr.GetWorkflowExecutionById(ctx, req.Msg.GetAccountId(), req.Msg.GetJobRunId(), logger)
	if err != nil {
		return err
	}

	lokiclient := loki.New(s.cfg.RunLogConfig.LokiRunLogConfig.BaseUrl, http.DefaultClient)
	direction := loki.BACKWARD
	end := time.Now()
	start := getLogFilterTime(req.Msg.GetWindow(), end)
	query := buildLokiQuery(
		s.cfg.RunLogConfig.LokiRunLogConfig.LabelsQuery,
		s.cfg.RunLogConfig.LokiRunLogConfig.KeepLabels,
		workflowExecution.GetExecution().GetWorkflowId(),
		getLogLevelFilters(req.Msg.GetLogLevels()),
	)
	resp, err := lokiclient.QueryRange(ctx, &loki.QueryRangeRequest{
		Query: query,
		Limit: req.Msg.MaxLogLines,

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
	for _, entry := range loki.GetEntriesFromStreams(streams) {
		err := stream.Send(&mgmtv1alpha1.GetJobRunLogsStreamResponse{LogLine: entry.Line, Timestamp: timestamppb.New(entry.Timestamp)})
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

func buildLokiQuery(lokiLables string, keep []string, workflowId string, loglevels []string) string {
	query := fmt.Sprintf("{%s} | json", lokiLables)
	query = fmt.Sprintf("%s | WorkflowID=%q", query, workflowId)

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
	accountUuid, err := s.verifyUserInAccount(ctx, id.GetAccountId())
	if err != nil {
		return nil, err
	}

	runContext, err := s.db.Q.GetRunContextByKey(ctx, s.db.Db, db_queries.GetRunContextByKeyParams{
		WorkflowId: id.GetJobRunId(),
		ExternalId: id.GetExternalId(),
		AccountId:  *accountUuid,
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
	accountUuid, err := s.verifyUserInAccount(ctx, id.GetAccountId())
	if err != nil {
		return nil, err
	}
	if s.cfg.IsNeosyncCloud && !isWorkerApiKey(ctx) {
		return nil, nucleuserrors.NewUnauthenticated("must provide valid authentication credentials for this endpoint")
	}

	var userId *pgtype.UUID
	if isWorkerApiKey(ctx) {
		uid, err := neosyncdb.ToUuid("00000000-0000-0000-0000-000000000000")
		if err != nil {
			return nil, err
		}
		userId = &uid
	} else {
		userUuid, err := s.getUserUuid(ctx)
		if err != nil {
			return nil, err
		}
		userId = userUuid
	}

	err = s.db.Q.SetRunContext(ctx, s.db.Db, db_queries.SetRunContextParams{
		WorkflowID:  id.GetJobRunId(),
		ExternalID:  id.GetExternalId(),
		AccountID:   *accountUuid,
		Value:       req.Msg.GetValue(),
		CreatedByID: *userId,
		UpdatedByID: *userId,
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
		accountUuid, err := s.verifyUserInAccount(ctx, id.GetAccountId())
		if err != nil {
			return nil, err
		}
		if s.cfg.IsNeosyncCloud && !isWorkerApiKey(ctx) {
			return nil, nucleuserrors.NewUnauthenticated("must provide valid authentication credentials for this endpoint")
		}

		var userId *pgtype.UUID
		if isWorkerApiKey(ctx) {
			uid, err := neosyncdb.ToUuid("00000000-0000-0000-0000-000000000000")
			if err != nil {
				return nil, err
			}
			userId = &uid
		} else {
			userUuid, err := s.getUserUuid(ctx)
			if err != nil {
				return nil, err
			}
			userId = userUuid
		}

		err = s.db.Q.SetRunContext(ctx, s.db.Db, db_queries.SetRunContextParams{
			WorkflowID:  id.GetJobRunId(),
			ExternalID:  id.GetExternalId(),
			AccountID:   *accountUuid,
			Value:       req.GetValue(),
			CreatedByID: *userId,
			UpdatedByID: *userId,
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
