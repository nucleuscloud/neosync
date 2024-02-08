package v1alpha1_jobservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"

	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetJobs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	jobIds := []pgtype.UUID{}
	for idx := range jobs {
		job := jobs[idx]
		jobIds = append(jobIds, job.ID)
	}

	var destinationAssociations []db_queries.NeosyncApiJobDestinationConnectionAssociation
	if len(jobIds) > 0 {
		destinationAssociations, err = s.db.Q.GetJobConnectionDestinationsByJobIds(ctx, s.db.Db, jobIds)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	}

	jobMap := map[pgtype.UUID]*db_queries.NeosyncApiJob{}
	for idx := range jobs {
		job := jobs[idx]
		jobMap[job.ID] = &job
	}

	associationMap := map[pgtype.UUID][]db_queries.NeosyncApiJobDestinationConnectionAssociation{}
	for i := range destinationAssociations {
		assoc := destinationAssociations[i]
		if _, ok := associationMap[assoc.JobID]; ok {
			associationMap[assoc.JobID] = append(associationMap[assoc.JobID], assoc)
		} else {
			associationMap[assoc.JobID] = append([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, assoc)
		}
	}

	dtos := []*mgmtv1alpha1.Job{}
	// Use jobIds to retain original query order
	for _, jobId := range jobIds {
		job := jobMap[jobId]
		dtos = append(dtos, dtomaps.ToJobDto(job, associationMap[job.ID]))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobsResponse{
		Jobs: dtos,
	}), nil
}

func (s *Service) GetJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRequest],
) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	errgrp, errctx := errgroup.WithContext(ctx)

	var job db_queries.NeosyncApiJob
	errgrp.Go(func() error {
		j, err := s.db.Q.GetJobById(errctx, s.db.Db, jobUuid)
		if err != nil {
			return err
		}
		job = j
		return nil
	})
	var destConnections []db_queries.NeosyncApiJobDestinationConnectionAssociation
	errgrp.Go(func() error {
		dcs, err := s.db.Q.GetJobConnectionDestinations(ctx, s.db.Db, jobUuid)
		if err != nil {
			return err
		}
		destConnections = dcs
		return nil
	})
	if err = errgrp.Wait(); err != nil {
		if nucleusdb.IsNoRows(err) {
			return nil, nucleuserrors.NewNotFound("unable to find job by id")
		}
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
		Job: dtomaps.ToJobDto(&job, destConnections),
	}), nil
}

func (s *Service) GetJobStatus(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobStatusRequest],
) (*connect.Response[mgmtv1alpha1.GetJobStatusResponse], error) {
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

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	scheduleHandle, err := s.temporalWfManager.GetScheduleHandleClientByAccount(ctx, nucleusdb.UUIDString(job.AccountID), nucleusdb.UUIDString(job.ID), logger)
	if err != nil {
		return nil, err
	}

	schedule, err := scheduleHandle.Describe(ctx)
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve schedule: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobStatusResponse{
		Status: dtomaps.ToJobStatus(schedule),
	}), nil
}

func (s *Service) GetJobStatuses(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobStatusesRequest],
) (*connect.Response[mgmtv1alpha1.GetJobStatusesResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	var scheduleclient temporalclient.ScheduleClient
	if len(jobs) > 0 {
		tclient, err := s.temporalWfManager.GetScheduleClientByAccount(ctx, req.Msg.AccountId, logger)
		if err != nil {
			return nil, err
		}
		scheduleclient = tclient
	}

	dtos := make([]*mgmtv1alpha1.JobStatusRecord, len(jobs))
	group := new(errgroup.Group)
	for i := range jobs {
		i := i
		j := jobs[i]
		group.Go(func() error {
			jobId := nucleusdb.UUIDString(j.ID)
			scheduleHandle := scheduleclient.GetHandle(ctx, jobId)
			schedule, err := scheduleHandle.Describe(ctx)
			if err != nil {
				logger.Error(fmt.Errorf("unable to retrieve schedule: %w", err).Error())
			} else {
				dtos[i] = &mgmtv1alpha1.JobStatusRecord{JobId: jobId, Status: dtomaps.ToJobStatus(schedule)}
			}
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job statuses: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobStatusesResponse{
		Statuses: dtos,
	}), nil
}

func (s *Service) GetJobRecentRuns(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRecentRunsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRecentRunsResponse], error) {
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

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	scheduleHandle, err := s.temporalWfManager.GetScheduleHandleClientByAccount(ctx, nucleusdb.UUIDString(job.AccountID), nucleusdb.UUIDString(job.ID), logger)
	if err != nil {
		return nil, err
	}

	schedule, err := scheduleHandle.Describe(ctx)
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve schedule: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRecentRunsResponse{
		RecentRuns: dtomaps.ToJobRecentRunsDto(schedule),
	}), nil
}

func (s *Service) GetJobNextRuns(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobNextRunsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobNextRunsResponse], error) {
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

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, nucleusdb.UUIDString(job.AccountID), logger)
	if err != nil {
		return nil, err
	}

	scheduleHandle := tclient.ScheduleClient().GetHandle(ctx, nucleusdb.UUIDString(job.ID))
	schedule, err := scheduleHandle.Describe(ctx)
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve schedule: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobNextRunsResponse{
		NextRuns: dtomaps.ToJobNextRunsDto(schedule),
	}), nil
}

type Destination struct {
	ConnectionId pgtype.UUID
	Options      *pg_models.JobDestinationOptions
}

func (s *Service) CreateJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobName", req.Msg.JobName, "accountId", req.Msg.AccountId)

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	connectionUuids := []pgtype.UUID{}
	connectionIds := []string{}
	destinations := []*Destination{}
	for _, dest := range req.Msg.Destinations {
		destUuid, err := nucleusdb.ToUuid(dest.ConnectionId)
		if err != nil {
			return nil, err
		}
		options := &pg_models.JobDestinationOptions{}
		err = options.FromDto(dest.Options)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, &Destination{ConnectionId: destUuid, Options: options})
		connectionIds = append(connectionIds, dest.ConnectionId)
		connectionUuids = append(connectionUuids, destUuid)
	}

	logger.Info("verifying connections")
	count, err := s.db.Q.AreConnectionsInAccount(ctx, s.db.Db, db_queries.AreConnectionsInAccountParams{
		AccountId:     *accountUuid,
		ConnectionIds: connectionUuids,
	})
	if err != nil {
		return nil, err
	}
	if count != int64(len(connectionUuids)) {
		logger.Error("connection is not in account")
		return nil, nucleuserrors.NewForbidden("provided connection id is not in account")
	}

	// we leave out generation fk source connection id as it might be set to a destination id
	switch config := req.Msg.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		connectionIds = append(connectionIds, config.Mysql.ConnectionId)
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		connectionIds = append(connectionIds, config.Postgres.ConnectionId)
	case *mgmtv1alpha1.JobSourceOptions_AwsS3:
		connectionIds = append(connectionIds, config.AwsS3.ConnectionId)
	default:
	}

	if !verifyConnectionIdsUnique(connectionIds) {
		logger.Error("connection ids are not unqiue")
		return nil, nucleuserrors.NewBadRequest("connections ids are not unique")
	}

	var connectionIdToVerify *string
	switch config := req.Msg.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		connectionIdToVerify = &config.Mysql.ConnectionId
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		connectionIdToVerify = &config.Postgres.ConnectionId
	case *mgmtv1alpha1.JobSourceOptions_AwsS3:
		connectionIdToVerify = &config.AwsS3.ConnectionId
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		fkConnId := config.Generate.GetFkSourceConnectionId()
		if fkConnId != "" {
			connectionIdToVerify = &fkConnId
		}
	default:
		return nil, errors.New("unsupported source option config type")
	}
	if connectionIdToVerify != nil {
		if err := s.verifyConnectionInAccount(ctx, *connectionIdToVerify, req.Msg.AccountId); err != nil {
			return nil, err
		}

		sourceUuid, err := nucleusdb.ToUuid(*connectionIdToVerify)
		if err != nil {
			return nil, err
		}
		areConnectionsCompatible, err := verifyConnectionsAreCompatible(ctx, s.db, sourceUuid, destinations)
		if err != nil {
			logger.Error(fmt.Errorf("unable to verify if connections are compatible: %w", err).Error())
			return nil, err
		}
		if !areConnectionsCompatible {
			return nil, nucleuserrors.NewBadRequest("connection types are incompatible")
		}
	}

	cron := pgtype.Text{}
	if req.Msg.CronSchedule != nil {
		err := cron.Scan(req.Msg.GetCronSchedule())
		if err != nil {
			return nil, err
		}
	}

	mappings := []*pg_models.JobMapping{}
	for _, mapping := range req.Msg.Mappings {
		jm := &pg_models.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	connectionOptions := &pg_models.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.Source.Options)
	if err != nil {
		return nil, err
	}

	connDestParams := []*nucleusdb.CreateJobConnectionDestination{}
	for _, dest := range destinations {
		connDestParams = append(connDestParams, &nucleusdb.CreateJobConnectionDestination{
			ConnectionId: dest.ConnectionId,
			Options:      dest.Options,
		})
	}

	logger.Info("verifying temporal workspace")
	hasNs, err := s.temporalWfManager.DoesAccountHaveTemporalWorkspace(ctx, req.Msg.AccountId, logger)
	if err != nil {
		wrappedErr := fmt.Errorf("unable to verify account's temporal workspace. error: %w", err)
		logger.Error(wrappedErr.Error())
		return nil, wrappedErr
	}
	if !hasNs {
		logger.Error("temporal namespace not configured")
		return nil, nucleuserrors.NewBadRequest("must first configure temporal namespace in account settings")
	}

	workflowOptions := &pg_models.WorkflowOptions{}
	if req.Msg.WorkflowOptions != nil {
		workflowOptions.FromDto(req.Msg.WorkflowOptions)
	}

	activitySyncOptions := &pg_models.ActivityOptions{}
	if req.Msg.SyncOptions != nil {
		activitySyncOptions.FromDto(req.Msg.SyncOptions)
	}

	tScheduleClient, err := s.temporalWfManager.GetScheduleClientByAccount(ctx, req.Msg.AccountId, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to build temporal schedule client by account: %w", err)
	}
	logger.Info("successfully created temporal schedule client")
	tconfig, err := s.temporalWfManager.GetTemporalConfigByAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve temporal config by account: %w", err)
	}
	logger.Info("successfully retrieved temporal account config")

	cj, err := s.db.CreateJob(ctx, &db_queries.CreateJobParams{
		Name:              req.Msg.JobName,
		AccountID:         *accountUuid,
		Status:            int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
		CronSchedule:      cron,
		ConnectionOptions: connectionOptions,
		Mappings:          mappings,
		CreatedByID:       *userUuid,
		UpdatedByID:       *userUuid,
		WorkflowOptions:   workflowOptions,
		SyncOptions:       activitySyncOptions,
	}, connDestParams)
	if err != nil {
		return nil, fmt.Errorf("unable to create job: %w", err)
	}
	jobUuid := nucleusdb.UUIDString(cj.ID)
	logger = logger.With("jobId", jobUuid)
	logger.Info("created job")

	logger = logger.With("jobId", jobUuid)
	schedule := nucleusdb.ToNullableString(cj.CronSchedule)
	paused := true
	spec := temporalclient.ScheduleSpec{}
	if schedule != nil && *schedule != "" {
		spec.CronExpressions = []string{*schedule}
		paused = false
	}
	action := &temporalclient.ScheduleWorkflowAction{
		Workflow:  datasync_workflow.Workflow,
		TaskQueue: tconfig.SyncJobQueueName,
		Args:      []any{&datasync_workflow.WorkflowRequest{JobId: jobUuid}},
	}
	if cj.WorkflowOptions != nil && cj.WorkflowOptions.RunTimeout != nil {
		action.WorkflowRunTimeout = time.Duration(*cj.WorkflowOptions.RunTimeout)
	}

	scheduleHandle, err := tScheduleClient.Create(ctx, temporalclient.ScheduleOptions{
		ID:     jobUuid,
		Spec:   spec,
		Paused: paused,
		Action: action,
	})
	if err != nil {
		logger.Error(fmt.Errorf("unable to create schedule workflow in temporal: %w", err).Error())
		logger.Info("deleting newly created job")
		removeJobErr := s.db.Q.RemoveJobById(ctx, s.db.Db, cj.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to create scheduled job and was unable to fully cleanup partially created resources: %w: %w", removeJobErr, err)
		}
		return nil, fmt.Errorf("unable to create schedule job: %w", err)
	}
	logger.Info("scheduled workflow", "workflowId", scheduleHandle.GetID())

	if req.Msg.InitiateJobRun {
		// manually trigger job run
		err := scheduleHandle.Trigger(ctx, temporalclient.ScheduleTriggerOptions{})
		if err != nil {
			// don't return error here
			logger.Error(fmt.Errorf("unable to trigger job: %w", err).Error())
		}
	}

	destinationConnections, err := s.db.Q.GetJobConnectionDestinations(ctx, s.db.Db, cj.ID)
	if err != nil {
		logger.Error("unable to retrieve job destination connections")
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateJobResponse{
		Job: dtomaps.ToJobDto(cj, destinationConnections),
	}), nil
}

func (s *Service) DeleteJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	idUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	job, err := s.db.Q.GetJobById(ctx, s.db.Db, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	tclient, err := s.temporalWfManager.GetWorkflowClientByAccount(ctx, nucleusdb.UUIDString(job.AccountID), logger)
	if err != nil {
		return nil, err
	}
	tconfig, err := s.temporalWfManager.GetTemporalConfigByAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	logger.Info("deleting schedule's workflow executions")
	workflows, err := getWorkflowExecutionsByJobIds(ctx, tclient, tconfig.Namespace, []string{req.Msg.Id})
	if err != nil {
		return nil, err
	}

	group := new(errgroup.Group)
	for _, w := range workflows {
		w := w
		group.Go(func() error {
			_, err := tclient.WorkflowService().DeleteWorkflowExecution(ctx, &workflowservice.DeleteWorkflowExecutionRequest{
				Namespace:         tconfig.Namespace,
				WorkflowExecution: w.Execution,
			})
			return err
		})
	}

	err = group.Wait()
	if err != nil {
		logger.Error(fmt.Errorf("unable to delete schedule's workflow executions: %w", err).Error())
		return nil, err
	}

	logger.Info("deleting schedule")
	scheduleHandle := tclient.ScheduleClient().GetHandle(ctx, nucleusdb.UUIDString(job.ID))
	description, err := scheduleHandle.Describe(ctx)
	if err != nil && !strings.Contains(err.Error(), "schedule not found") && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, err
	}

	if description != nil {
		err = scheduleHandle.Delete(ctx)
		if err != nil {
			logger.Error(fmt.Errorf("unable to delete schedule: %w", err).Error())
			return nil, err
		}
	}

	logger.Info("deleting job")
	err = s.db.Q.RemoveJobById(ctx, s.db.Db, job.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
}

func (s *Service) CreateJobDestinationConnections(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobDestinationConnectionsRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobDestinationConnectionsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.JobId)

	jobUuid, err := nucleusdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	job, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.JobId,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	accountUuid, err := s.verifyUserInAccount(ctx, job.Msg.Job.AccountId)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	logger = logger.With("userId", userUuid)

	connectionIds := []string{}
	connectionUuids := []pgtype.UUID{}
	destinations := []*Destination{}
	for _, dest := range req.Msg.Destinations {
		destUuid, err := nucleusdb.ToUuid(dest.ConnectionId)
		if err != nil {
			return nil, err
		}
		options := &pg_models.JobDestinationOptions{}
		err = options.FromDto(dest.Options)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, &Destination{ConnectionId: destUuid, Options: options})
		connectionIds = append(connectionIds, dest.ConnectionId)
		connectionUuids = append(connectionUuids, destUuid)
	}

	if !verifyConnectionIdsUnique(connectionIds) {
		return nil, nucleuserrors.NewBadRequest("connections ids are not unique")
	}

	isInSameAccount, err := verifyConnectionsInAccount(ctx, s.db, connectionUuids, *accountUuid)
	if err != nil {
		return nil, err
	}
	if !isInSameAccount {
		return nil, nucleuserrors.NewBadRequest("connections ids are not unique")
	}

	logger.Info("creating job destination connections", "connectionIds", connectionIds)
	connDestParams := []db_queries.CreateJobConnectionDestinationsParams{}
	for _, dest := range destinations {
		connDestParams = append(connDestParams, db_queries.CreateJobConnectionDestinationsParams{
			JobID:        jobUuid,
			ConnectionID: dest.ConnectionId,
			Options:      dest.Options,
		})
	}
	if len(connDestParams) > 0 {
		_, err = s.db.Q.CreateJobConnectionDestinations(ctx, s.db.Db, connDestParams)
		if err != nil {
			return nil, err
		}
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.JobId,
	}))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateJobDestinationConnectionsResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) UpdateJobSchedule(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobScheduleRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobScheduleResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job schedule")
	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	cron := pgtype.Text{}
	if req.Msg.CronSchedule != nil {
		err := cron.Scan(req.Msg.GetCronSchedule())
		if err != nil {
			return nil, err
		}
	}

	if err := s.db.WithTx(ctx, nil, func(dbtx nucleusdb.BaseDBTX) error {
		_, err = s.db.Q.UpdateJobSchedule(ctx, dbtx, db_queries.UpdateJobScheduleParams{
			ID:           job.ID,
			CronSchedule: cron,
			UpdatedByID:  *userUuid,
		})
		if err != nil {
			return err
		}

		spec := &temporalclient.ScheduleSpec{}
		if req.Msg.CronSchedule != nil && *req.Msg.CronSchedule != "" {
			spec.CronExpressions = []string{*req.Msg.CronSchedule}
		}

		// update temporal scheduled job
		scheduleHandle, err := s.temporalWfManager.GetScheduleHandleClientByAccount(ctx, nucleusdb.UUIDString(job.AccountID), nucleusdb.UUIDString(job.ID), logger)
		if err != nil {
			return err
		}
		err = scheduleHandle.Update(ctx, temporalclient.ScheduleUpdateOptions{
			DoUpdate: func(schedule temporalclient.ScheduleUpdateInput) (*temporalclient.ScheduleUpdate, error) {
				schedule.Description.Schedule.Spec = spec
				return &temporalclient.ScheduleUpdate{
					Schedule: &schedule.Description.Schedule,
				}, nil
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("unable to update schedule: %w", err).Error())
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobScheduleResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) PauseJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.PauseJobRequest],
) (*connect.Response[mgmtv1alpha1.PauseJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	scheduleHandle, err := s.temporalWfManager.GetScheduleHandleClientByAccount(ctx, nucleusdb.UUIDString(job.AccountID), nucleusdb.UUIDString(job.ID), logger)
	if err != nil {
		return nil, err
	}
	if req.Msg.Pause {
		logger.Info("pausing job")
		err = scheduleHandle.Pause(ctx, temporalclient.SchedulePauseOptions{Note: req.Msg.GetNote()})
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("unpausing job")
		err = scheduleHandle.Unpause(ctx, temporalclient.ScheduleUnpauseOptions{Note: req.Msg.GetNote()})
		if err != nil {
			return nil, err
		}
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.PauseJobResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) UpdateJobSourceConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobSourceConnectionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job source connection and mappings")
	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	var connectionIdToVerify *string
	switch config := req.Msg.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		connectionIdToVerify = &config.Mysql.ConnectionId
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		connectionIdToVerify = &config.Postgres.ConnectionId
	case *mgmtv1alpha1.JobSourceOptions_AwsS3:
		connectionIdToVerify = &config.AwsS3.ConnectionId
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		fkConnId := config.Generate.GetFkSourceConnectionId()
		if fkConnId != "" {
			connectionIdToVerify = &fkConnId
		}
	}

	if connectionIdToVerify != nil {
		if err := s.verifyConnectionInAccount(ctx, *connectionIdToVerify, nucleusdb.UUIDString(job.AccountID)); err != nil {
			return nil, err
		}
	}

	conn, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: *connectionIdToVerify,
	}))

	switch conn.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if _, ok := req.Msg.Source.Options.Config.(*mgmtv1alpha1.JobSourceOptions_Mysql); !ok {
			return nil, fmt.Errorf("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if _, ok := req.Msg.Source.Options.Config.(*mgmtv1alpha1.JobSourceOptions_Postgres); !ok {
			return nil, fmt.Errorf("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		if _, ok := req.Msg.Source.Options.Config.(*mgmtv1alpha1.JobSourceOptions_AwsS3); !ok {
			return nil, fmt.Errorf("job source option config type and connection type mismatch")
		}
	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}

	connectionOptions := &pg_models.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.Source.Options)
	if err != nil {
		return nil, err
	}

	mappings := []*pg_models.JobMapping{}
	for _, mapping := range req.Msg.Mappings {
		jm := &pg_models.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	if err := s.db.WithTx(ctx, nil, func(dbtx nucleusdb.BaseDBTX) error {
		_, err = s.db.Q.UpdateJobSource(ctx, dbtx, db_queries.UpdateJobSourceParams{
			ID:                job.ID,
			ConnectionOptions: connectionOptions,

			UpdatedByID: *userUuid,
		})
		if err != nil {
			logger.Error(fmt.Errorf("unable to update job source: %w", err).Error())
			return err
		}

		_, err = s.db.Q.UpdateJobMappings(ctx, dbtx, db_queries.UpdateJobMappingsParams{
			ID:          job.ID,
			Mappings:    mappings,
			UpdatedByID: *userUuid,
		})
		if err != nil {
			logger.Error(fmt.Errorf("unable to update job mappings: %w", err).Error())
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobSourceConnectionResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) SetJobSourceSqlConnectionSubsets(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetJobSourceSqlConnectionSubsetsRequest],
) (*connect.Response[mgmtv1alpha1.SetJobSourceSqlConnectionSubsetsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job source sql connection subsets")
	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	var connectionId *string
	if job.ConnectionOptions != nil {
		if job.ConnectionOptions.MysqlOptions != nil {
			connectionId = &job.ConnectionOptions.MysqlOptions.ConnectionId
		} else if job.ConnectionOptions.PostgresOptions != nil {
			connectionId = &job.ConnectionOptions.PostgresOptions.ConnectionId
		} else {
			return nil, nucleuserrors.NewBadRequest("only jobs with a valid source connection id may be subset")
		}
	}
	if connectionId == nil || *connectionId == "" {
		return nil, nucleuserrors.NewInternalError("unable to find connection id")
	}

	connectionResp, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: *connectionId,
	}))
	if err != nil {
		return nil, err
	}
	connection := connectionResp.Msg.Connection

	if connection.ConnectionConfig == nil ||
		(connection.ConnectionConfig.GetPgConfig() == nil && connection.ConnectionConfig.GetMysqlConfig() == nil) {
		return nil, nucleuserrors.NewBadRequest("may only update subsets if the source connection is a SQL-based connection")
	}

	if err := s.db.SetSqlSourceSubsets(
		ctx,
		jobUuid,
		req.Msg.Schemas,
		*userUuid,
	); err != nil {
		return nil, err
	}

	updatedJobRes, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.SetJobSourceSqlConnectionSubsetsResponse{
		Job: updatedJobRes.Msg.Job,
	}), nil
}

func (s *Service) UpdateJobDestinationConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobDestinationConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobDestinationConnectionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.JobId, "connectionId", req.Msg.ConnectionId)

	jobUuid, err := nucleusdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	destinationUuid, err := nucleusdb.ToUuid(req.Msg.DestinationId)
	if err != nil {
		return nil, err
	}
	job, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.JobId,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	_, err = s.verifyUserInAccount(ctx, job.Msg.Job.AccountId)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	logger = logger.With("userId", userUuid)

	connectionUuid, err := nucleusdb.ToUuid(req.Msg.ConnectionId)
	if err != nil {
		return nil, err
	}
	if err := s.verifyConnectionInAccount(ctx, req.Msg.ConnectionId, job.Msg.Job.AccountId); err != nil {
		return nil, err
	}
	options := &pg_models.JobDestinationOptions{}
	err = options.FromDto(req.Msg.Options)
	if err != nil {
		return nil, err
	}

	logger.Info("updating job destination connection")
	_, err = s.db.Q.UpdateJobConnectionDestination(ctx, s.db.Db, db_queries.UpdateJobConnectionDestinationParams{
		ID:           destinationUuid,
		ConnectionID: connectionUuid,
		Options:      options,
	})
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		logger.Info("destination not found. creating job destination connection")
		_, err = s.db.Q.CreateJobConnectionDestination(ctx, s.db.Db, db_queries.CreateJobConnectionDestinationParams{
			JobID:        jobUuid,
			ConnectionID: connectionUuid,
			Options:      options,
		})
		if err != nil {
			return nil, err
		}
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: job.Msg.Job.Id,
	}))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobDestinationConnectionResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) DeleteJobDestinationConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobDestinationConnectionRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobDestinationConnectionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("destinationId", req.Msg.DestinationId)

	destinationUuid, err := nucleusdb.ToUuid(req.Msg.DestinationId)
	if err != nil {
		return nil, err
	}

	destination, err := s.db.Q.GetJobConnectionDestination(ctx, s.db.Db, destinationUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobDestinationConnectionResponse{}), nil
	}

	job, err := s.db.Q.GetJobById(ctx, s.db.Db, destination.JobID)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	logger = logger.With("userId", userUuid, "jobId", job.ID)

	if err := s.verifyConnectionInAccount(
		ctx,
		nucleusdb.UUIDString(destination.ConnectionID),
		nucleusdb.UUIDString(job.AccountID)); err != nil {
		return nil, err
	}

	logger.Info("deleting job destination connection")
	err = s.db.Q.RemoveJobConnectionDestination(ctx, s.db.Db, destinationUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		logger.Info("destination not found")
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteJobDestinationConnectionResponse{}), nil
}

func (s *Service) IsJobNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsJobNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsJobNameAvailableResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsJobNameAvailable(ctx, s.db.Db, db_queries.IsJobNameAvailableParams{
		AccountId: *accountUuid,
		JobName:   req.Msg.Name,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.IsJobNameAvailableResponse{
		IsAvailable: count == 0,
	}), nil
}

func (s *Service) verifyConnectionInAccount(
	ctx context.Context,
	connectionId string,
	accountId string,
) error {
	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return err
	}
	connectionUuid, err := nucleusdb.ToUuid(connectionId)
	if err != nil {
		return err
	}

	count, err := s.db.Q.IsConnectionInAccount(ctx, s.db.Db, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: connectionUuid,
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return nucleuserrors.NewForbidden("provided connection id is not in account")
	}
	return nil
}

func getWorkflowExecutionsByJobIds(
	ctx context.Context,
	tc temporalclient.Client,
	namespace string,
	jobIds []string,
) ([]*workflowpb.WorkflowExecutionInfo, error) {
	jobIdStr := ""
	for _, id := range jobIds {
		jobIdStr += fmt.Sprintf(`%q,`, id)
	}
	query := fmt.Sprintf("TemporalScheduledById IN (%s)", strings.TrimSuffix(jobIdStr, ","))
	executions := []*workflowpb.WorkflowExecutionInfo{}
	if len(jobIds) == 0 {
		return executions, nil
	}
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := tc.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      20,
			NextPageToken: nextPageToken,
			Query:         query,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve workflow executions: %w", err)
		}

		executions = append(executions, resp.Executions...)
		nextPageToken = resp.NextPageToken
	}

	return executions, nil
}

func verifyConnectionsInAccount(ctx context.Context, db *nucleusdb.NucleusDb, connectionUuids []pgtype.UUID, accountUuid pgtype.UUID) (bool, error) {
	conns, err := db.Q.GetConnectionsByIds(ctx, db.Db, connectionUuids)
	if err != nil {
		return false, err
	}

	for i := range conns {
		c := conns[i]
		if c.AccountID != accountUuid {
			return false, nil
		}
	}
	return true, nil
}

func verifyConnectionIdsUnique(connectionIds []string) bool {
	occurrenceMap := make(map[string]bool)

	for _, id := range connectionIds {
		if occurrenceMap[id] {
			return false
		}
		occurrenceMap[id] = true
	}
	return true
}

func verifyConnectionsAreCompatible(ctx context.Context, db *nucleusdb.NucleusDb, sourceConnId pgtype.UUID, destinations []*Destination) (bool, error) {
	var sourceConnection db_queries.NeosyncApiConnection
	dests := make([]db_queries.NeosyncApiConnection, len(destinations))
	group := new(errgroup.Group)
	group.Go(func() error {
		source, err := db.Q.GetConnectionById(ctx, db.Db, sourceConnId)
		if err != nil {
			return err
		}
		sourceConnection = source
		return nil
	})
	for i := range destinations {
		i := i
		d := destinations[i]
		group.Go(func() error {
			connection, err := db.Q.GetConnectionById(ctx, db.Db, d.ConnectionId)
			if err != nil {
				return err
			}
			dests[i] = connection
			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return false, err
	}

	for i := range dests {
		d := dests[i]
		// AWS S3 is always a valid destination regardless of source connection type
		if d.ConnectionConfig.AwsS3Config != nil {
			continue
		}
		if sourceConnection.ConnectionConfig.PgConfig != nil && d.ConnectionConfig.MysqlConfig != nil {
			// invalid Postgres source cannot have Mysql destination
			return false, nil
		}
		if sourceConnection.ConnectionConfig.MysqlConfig != nil && d.ConnectionConfig.PgConfig != nil {
			// invalid Mysql source cannot habe Postgres destination
			return false, nil
		}
	}

	return true, nil
}

func (s *Service) SetJobWorkflowOptions(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetJobWorkflowOptionsRequest],
) (*connect.Response[mgmtv1alpha1.SetJobWorkflowOptionsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)

	job, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	_, err = s.verifyUserInAccount(ctx, job.Msg.Job.AccountId)
	if err != nil {
		return nil, err
	}

	wfOptions := &pg_models.WorkflowOptions{}
	if req.Msg.WorfklowOptions != nil {
		wfOptions.FromDto(req.Msg.WorfklowOptions)
	}

	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Q.SetJobWorkflowOptions(ctx, s.db.Db, db_queries.SetJobWorkflowOptionsParams{
		ID:              jobUuid,
		WorkflowOptions: wfOptions,
		UpdatedByID:     *userUuid,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to set job workflow options: %w", err)
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.SetJobWorkflowOptionsResponse{Job: updatedJob.Msg.Job}), nil
}

func (s *Service) SetJobSyncOptions(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.SetJobSyncOptionsRequest],
) (*connect.Response[mgmtv1alpha1.SetJobSyncOptionsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)

	job, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	_, err = s.verifyUserInAccount(ctx, job.Msg.Job.AccountId)
	if err != nil {
		return nil, err
	}

	syncOptions := &pg_models.ActivityOptions{}
	if req.Msg.SyncOptions != nil {
		syncOptions.FromDto(req.Msg.SyncOptions)
	}

	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Q.SetJobSyncOptions(ctx, s.db.Db, db_queries.SetJobSyncOptionsParams{
		ID:          jobUuid,
		SyncOptions: syncOptions,
		UpdatedByID: *userUuid,
	})
	if err != nil {
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.SetJobSyncOptionsResponse{Job: updatedJob.Msg.Job}), nil
}
