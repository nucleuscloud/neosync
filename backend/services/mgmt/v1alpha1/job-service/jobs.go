package v1alpha1_jobservice

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"

	temporalclient "go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
)

const (
	defaultCronStr = "0 0 1 1 *"
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
	logger = logger.With("accountId", req.Msg.AccountId)
	jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Debug(fmt.Sprintf("found %d jobs", len(jobs)))

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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)
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
		if neosyncdb.IsNoRows(err) {
			return nil, nucleuserrors.NewNotFound("unable to find job by id")
		}
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	schedule, err := s.temporalmgr.DescribeSchedule(ctx, neosyncdb.UUIDString(job.AccountID), neosyncdb.UUIDString(job.ID), logger)
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

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}
	jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	scheduleIds := make([]string, 0, len(jobs))
	for idx := range jobs {
		scheduleIds = append(scheduleIds, neosyncdb.UUIDString(jobs[idx].ID))
	}

	responses, err := s.temporalmgr.DescribeSchedules(
		ctx,
		req.Msg.GetAccountId(),
		scheduleIds,
		logger,
	)
	if err != nil {
		return nil, err
	}

	dtos := make([]*mgmtv1alpha1.JobStatusRecord, len(jobs))
	for idx, resp := range responses {
		if resp.Error != nil {
			logger.Error(fmt.Errorf("unable to retrieve schedule: %w", err).Error())
		} else if resp.Schedule != nil {
			dtos[idx] = &mgmtv1alpha1.JobStatusRecord{
				JobId:  scheduleIds[idx],
				Status: dtomaps.ToJobStatus(resp.Schedule),
			}
		}
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	schedule, err := s.temporalmgr.DescribeSchedule(ctx, neosyncdb.UUIDString(job.AccountID), neosyncdb.UUIDString(job.ID), logger)
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil {
		return nil, err
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	schedule, err := s.temporalmgr.DescribeSchedule(ctx, neosyncdb.UUIDString(job.AccountID), neosyncdb.UUIDString(job.ID), logger)
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
		destUuid, err := neosyncdb.ToUuid(dest.ConnectionId)
		if err != nil {
			return nil, err
		}
		options := &pg_models.JobDestinationOptions{}
		err = options.FromDto(dest.GetOptions())
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
	case *mgmtv1alpha1.JobSourceOptions_Mongodb:
		connectionIds = append(connectionIds, config.Mongodb.GetConnectionId())
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		connectionIds = append(connectionIds, config.Dynamodb.GetConnectionId())
	case *mgmtv1alpha1.JobSourceOptions_Mssql:
		connectionIds = append(connectionIds, config.Mssql.GetConnectionId())
	default:
	}

	if !verifyConnectionIdsUnique(connectionIds) {
		logger.Error("connection ids are not unqiue")
		return nil, nucleuserrors.NewBadRequest("connections ids are not unique")
	}

	connectionIdToVerify, err := getJobSourceConnectionId(req.Msg.GetSource())
	if err != nil {
		return nil, err
	}
	if connectionIdToVerify != nil {
		if err := s.verifyConnectionInAccount(ctx, *connectionIdToVerify, req.Msg.AccountId); err != nil {
			return nil, err
		}

		sourceUuid, err := neosyncdb.ToUuid(*connectionIdToVerify)
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

	cronStr := req.Msg.GetCronSchedule()
	if cronStr == "" {
		cronStr = defaultCronStr
	}
	cronText := pgtype.Text{}
	err = cronText.Scan(cronStr)
	if err != nil {
		return nil, err
	}

	mappings := []*pg_models.JobMapping{}
	for _, mapping := range req.Msg.GetMappings() {
		jm := &pg_models.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	virtualForeignKeys := []*pg_models.VirtualForeignConstraint{}
	for _, fk := range req.Msg.GetVirtualForeignKeys() {
		vfk := &pg_models.VirtualForeignConstraint{}
		err = vfk.FromDto(fk)
		if err != nil {
			return nil, err
		}
		virtualForeignKeys = append(virtualForeignKeys, vfk)
	}

	connectionOptions := &pg_models.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.Source.Options)
	if err != nil {
		return nil, err
	}

	connDestParams := []*neosyncdb.CreateJobConnectionDestination{}
	for _, dest := range destinations {
		connDestParams = append(connDestParams, &neosyncdb.CreateJobConnectionDestination{
			ConnectionId: dest.ConnectionId,
			Options:      dest.Options,
		})
	}

	logger.Info("verifying temporal workspace")
	hasNs, err := s.temporalmgr.DoesAccountHaveNamespace(ctx, req.Msg.AccountId, logger)
	if err != nil {
		wrappedErr := fmt.Errorf("unable to verify account's temporal workspace. error: %w", err)
		logger.Error(wrappedErr.Error())
		return nil, wrappedErr
	}
	if !hasNs {
		logger.Error("temporal namespace not configured")
		return nil, nucleuserrors.NewBadRequest("must first configure temporal namespace in account settings")
	}

	taskQueue, err := s.temporalmgr.GetSyncJobTaskQueue(ctx, req.Msg.GetAccountId(), logger)
	if err != nil {
		return nil, err
	}

	workflowOptions := &pg_models.WorkflowOptions{}
	if req.Msg.WorkflowOptions != nil {
		workflowOptions.FromDto(req.Msg.WorkflowOptions)
	}

	activitySyncOptions := &pg_models.ActivityOptions{}
	if req.Msg.SyncOptions != nil {
		activitySyncOptions.FromDto(req.Msg.SyncOptions)
	}

	cj, err := s.db.CreateJob(ctx, &db_queries.CreateJobParams{
		Name:               req.Msg.JobName,
		AccountID:          *accountUuid,
		Status:             int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
		CronSchedule:       cronText,
		ConnectionOptions:  connectionOptions,
		Mappings:           mappings,
		VirtualForeignKeys: virtualForeignKeys,
		CreatedByID:        *userUuid,
		UpdatedByID:        *userUuid,
		WorkflowOptions:    workflowOptions,
		SyncOptions:        activitySyncOptions,
	}, connDestParams)
	if err != nil {
		return nil, fmt.Errorf("unable to create job: %w", err)
	}
	jobUuid := neosyncdb.UUIDString(cj.ID)
	logger = logger.With("jobId", jobUuid)
	logger.Info("created job")

	logger = logger.With("jobId", jobUuid)
	schedule := neosyncdb.ToNullableString(cj.CronSchedule)
	paused := true
	spec := temporalclient.ScheduleSpec{}
	if schedule != nil {
		spec.CronExpressions = []string{*schedule}
		// we only want to unpause the temporal schedule if the user provided the cronstring directly
		if req.Msg.GetCronSchedule() != "" {
			paused = false
		}
	}
	action := &temporalclient.ScheduleWorkflowAction{
		Workflow:  datasync_workflow.Workflow,
		TaskQueue: taskQueue,
		Args:      []any{&datasync_workflow.WorkflowRequest{JobId: jobUuid}},
		ID:        neosyncdb.UUIDString(cj.ID),
	}
	if cj.WorkflowOptions != nil && cj.WorkflowOptions.RunTimeout != nil {
		action.WorkflowRunTimeout = time.Duration(*cj.WorkflowOptions.RunTimeout)
	}

	scheduleId, err := s.temporalmgr.CreateSchedule(
		ctx,
		req.Msg.GetAccountId(),
		&temporalclient.ScheduleOptions{
			ID:     jobUuid,
			Spec:   spec,
			Paused: paused,
			Action: action,
		},
		logger,
	)
	if err != nil {
		logger.Error(fmt.Errorf("unable to create schedule workflow in temporal: %w", err).Error())
		logger.Info("deleting newly created job")
		removeJobErr := s.db.Q.RemoveJobById(ctx, s.db.Db, cj.ID)
		if removeJobErr != nil {
			return nil, fmt.Errorf("unable to create scheduled job and was unable to fully cleanup partially created resources: %w: %w", removeJobErr, err)
		}
		return nil, fmt.Errorf("unable to create scheduled job: %w", err)
	}
	logger.Info("scheduled workflow", "workflowId", scheduleId)

	if req.Msg.InitiateJobRun {
		// manually trigger job run
		err := s.temporalmgr.TriggerSchedule(ctx, req.Msg.GetAccountId(), scheduleId, &temporalclient.ScheduleTriggerOptions{}, logger)
		if err != nil {
			// don't return error here
			logger.Error(fmt.Errorf("unable to trigger job: %w", err).Error())
		}
	}

	destinationConnections, err := s.db.Q.GetJobConnectionDestinations(ctx, s.db.Db, cj.ID)
	// not returning an error here because the job has already been successfully created and we just want to return the created job
	if err != nil {
		logger.Error(fmt.Sprintf("unable to retrieve job destination connections: %s", err.Error()))
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
	idUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	job, err := s.db.Q.GetJobById(ctx, s.db.Db, idUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.temporalmgr.DeleteSchedule(
		ctx,
		neosyncdb.UUIDString(job.AccountID),
		neosyncdb.UUIDString(job.ID),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to remove schedule when deleting job")
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

	jobUuid, err := neosyncdb.ToUuid(req.Msg.JobId)
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
		destUuid, err := neosyncdb.ToUuid(dest.ConnectionId)
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	cronStr := req.Msg.GetCronSchedule()
	if cronStr == "" {
		cronStr = defaultCronStr
	}
	cronText := pgtype.Text{}
	err = cronText.Scan(cronStr)
	if err != nil {
		return nil, err
	}

	if err := s.db.WithTx(ctx, nil, func(dbtx neosyncdb.BaseDBTX) error {
		_, err = s.db.Q.UpdateJobSchedule(ctx, dbtx, db_queries.UpdateJobScheduleParams{
			ID:           job.ID,
			CronSchedule: cronText,
			UpdatedByID:  *userUuid,
		})
		if err != nil {
			return err
		}

		spec := &temporalclient.ScheduleSpec{}
		spec.CronExpressions = []string{cronStr}

		// update temporal scheduled job
		err = s.temporalmgr.UpdateSchedule(
			ctx,
			neosyncdb.UUIDString(job.AccountID),
			neosyncdb.UUIDString(job.ID),
			&temporalclient.ScheduleUpdateOptions{
				DoUpdate: func(schedule temporalclient.ScheduleUpdateInput) (*temporalclient.ScheduleUpdate, error) {
					schedule.Description.Schedule.Spec = spec
					return &temporalclient.ScheduleUpdate{
						Schedule: &schedule.Description.Schedule,
					}, nil
				},
			},
			logger,
		)
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	if req.Msg.Pause {
		logger.Info("pausing job")
		err = s.temporalmgr.PauseSchedule(
			ctx,
			neosyncdb.UUIDString(job.AccountID),
			neosyncdb.UUIDString(job.ID),
			&temporalclient.SchedulePauseOptions{Note: req.Msg.GetNote()},
			logger,
		)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("unpausing job")
		err = s.temporalmgr.UnpauseSchedule(
			ctx,
			neosyncdb.UUIDString(job.AccountID),
			neosyncdb.UUIDString(job.ID),
			&temporalclient.ScheduleUnpauseOptions{Note: req.Msg.GetNote()},
			logger,
		)
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)

	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	var connectionIdToVerify string
	switch config := req.Msg.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		connectionIdToVerify = config.Mysql.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		connectionIdToVerify = config.Postgres.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_AwsS3:
		connectionIdToVerify = config.AwsS3.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		connectionIdToVerify = config.Generate.GetFkSourceConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		connectionIdToVerify = config.AiGenerate.GetFkSourceConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Mongodb:
		connectionIdToVerify = config.Mongodb.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		connectionIdToVerify = config.Dynamodb.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Mssql:
		connectionIdToVerify = config.Mssql.GetConnectionId()
	default:
		return nil, fmt.Errorf("unable to find connection id to verify for config: %T", config)
	}

	if connectionIdToVerify == "" {
		return nil, nucleuserrors.NewBadRequest("must provide valid non empty connection id")
	}

	// verifies that the account has access to that connection id
	if err := s.verifyConnectionInAccount(ctx, connectionIdToVerify, neosyncdb.UUIDString(job.AccountID)); err != nil {
		return nil, err
	}

	// retrieves the connection details
	conn, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionIdToVerify,
	}))

	// Type checking that the connection config that we want to use for the job is the same as the incoming job source config type
	switch cconfig := conn.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		dbConf := req.Msg.GetSource().GetOptions().GetMysql()
		generateConf := req.Msg.GetSource().GetOptions().GetGenerate()
		aigenerateConf := req.Msg.GetSource().GetOptions().GetAiGenerate()
		if dbConf == nil && generateConf == nil && aigenerateConf == nil {
			return nil, nucleuserrors.NewBadRequest("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		dbConf := req.Msg.GetSource().GetOptions().GetPostgres()
		generateConf := req.Msg.GetSource().GetOptions().GetGenerate()
		aigenerateConf := req.Msg.GetSource().GetOptions().GetAiGenerate()
		if dbConf == nil && generateConf == nil && aigenerateConf == nil {
			return nil, nucleuserrors.NewBadRequest("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		if _, ok := req.Msg.Source.Options.Config.(*mgmtv1alpha1.JobSourceOptions_AwsS3); !ok {
			return nil, fmt.Errorf("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		dbConf := req.Msg.GetSource().GetOptions().GetMongodb()
		generateConf := req.Msg.GetSource().GetOptions().GetGenerate()
		aigenerateConf := req.Msg.GetSource().GetOptions().GetAiGenerate()
		if dbConf == nil && generateConf == nil && aigenerateConf == nil {
			return nil, nucleuserrors.NewBadRequest("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		dbConf := req.Msg.GetSource().GetOptions().GetDynamodb()
		generateConf := req.Msg.GetSource().GetOptions().GetGenerate()
		aigenerateConf := req.Msg.GetSource().GetOptions().GetAiGenerate()
		if dbConf == nil && generateConf == nil && aigenerateConf == nil {
			return nil, nucleuserrors.NewBadRequest("job source option config type and connection type mismatch")
		}
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		dbConf := req.Msg.GetSource().GetOptions().GetMssql()
		generateConf := req.Msg.GetSource().GetOptions().GetGenerate()
		aigenerateConf := req.Msg.GetSource().GetOptions().GetAiGenerate()
		if dbConf == nil && generateConf == nil && aigenerateConf == nil {
			return nil, nucleuserrors.NewBadRequest("job source option config type and connection type mismatch")
		}
	default:
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("connection config is not currently supported: %T", cconfig))
	}

	connectionOptions := &pg_models.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.GetSource().GetOptions())
	if err != nil {
		return nil, err
	}

	mappings := []*pg_models.JobMapping{}
	for _, mapping := range req.Msg.GetMappings() {
		jm := &pg_models.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	vfkKeys := map[string]struct{}{}
	virtualForeignKeys := []*pg_models.VirtualForeignConstraint{}
	for _, fk := range req.Msg.GetVirtualForeignKeys() {
		key := fmt.Sprintf("%s.%s.%s.%s.%s.%s", fk.GetSchema(), fk.GetTable(), strings.Join(fk.GetColumns(), "."), fk.GetForeignKey().GetSchema(), fk.GetForeignKey().GetTable(), strings.Join(fk.GetForeignKey().GetColumns(), "."))
		if _, exists := vfkKeys[key]; exists {
			// skip duplicates
			continue
		}
		vfk := &pg_models.VirtualForeignConstraint{}
		err = vfk.FromDto(fk)
		if err != nil {
			return nil, err
		}
		virtualForeignKeys = append(virtualForeignKeys, vfk)
		vfkKeys[key] = struct{}{}
	}

	if err := s.db.WithTx(ctx, nil, func(dbtx neosyncdb.BaseDBTX) error {
		_, err = s.db.Q.UpdateJobSource(ctx, dbtx, db_queries.UpdateJobSourceParams{
			ID:                job.ID,
			ConnectionOptions: connectionOptions,

			UpdatedByID: *userUuid,
		})
		if err != nil {
			return fmt.Errorf("unable to update job source: %w", err)
		}

		_, err = s.db.Q.UpdateJobMappings(ctx, dbtx, db_queries.UpdateJobMappingsParams{
			ID:          job.ID,
			Mappings:    mappings,
			UpdatedByID: *userUuid,
		})
		if err != nil {
			return fmt.Errorf("unable to update job mappings: %w", err)
		}

		args := db_queries.UpdateJobVirtualForeignKeysParams{
			VirtualForeignKeys: virtualForeignKeys,
			UpdatedByID:        *userUuid,
			ID:                 job.ID,
		}
		_, err = s.db.Q.UpdateJobVirtualForeignKeys(ctx, dbtx, args)
		if err != nil {
			return fmt.Errorf("unable to update virtual foreign key: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve job after source update: %w", err)
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
	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.db.Q.GetJobById(ctx, s.db.Db, jobUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
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
		} else if job.ConnectionOptions.DynamoDBOptions != nil {
			connectionId = &job.ConnectionOptions.DynamoDBOptions.ConnectionId
		} else if job.ConnectionOptions.MssqlOptions != nil {
			connectionId = &job.ConnectionOptions.MssqlOptions.ConnectionId
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
		(connection.ConnectionConfig.GetPgConfig() == nil && connection.ConnectionConfig.GetMysqlConfig() == nil && connection.ConnectionConfig.GetDynamodbConfig() == nil && connection.ConnectionConfig.GetMssqlConfig() == nil) {
		return nil, nucleuserrors.NewBadRequest("may only update subsets for select source connections")
	}

	if err := s.db.SetSourceSubsets(
		ctx,
		jobUuid,
		req.Msg.Schemas,
		req.Msg.SubsetByForeignKeyConstraints,
		*userUuid,
	); err != nil {
		return nil, err
	}

	updatedJobRes, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve job: %w", err)
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

	jobUuid, err := neosyncdb.ToUuid(req.Msg.JobId)
	if err != nil {
		return nil, err
	}
	destinationUuid, err := neosyncdb.ToUuid(req.Msg.DestinationId)
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

	connectionUuid, err := neosyncdb.ToUuid(req.Msg.ConnectionId)
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

	// todo(NEOS-1281):  need a lot more validation here for changing connection uuid, matching options, as well as creating a new destination
	// if that destination is not supported with the source type
	logger.Info("updating job destination connection")
	_, err = s.db.Q.UpdateJobConnectionDestination(ctx, s.db.Db, db_queries.UpdateJobConnectionDestinationParams{
		ID:           destinationUuid,
		ConnectionID: connectionUuid,
		Options:      options,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
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

	destinationUuid, err := neosyncdb.ToUuid(req.Msg.DestinationId)
	if err != nil {
		return nil, err
	}

	destination, err := s.db.Q.GetJobConnectionDestination(ctx, s.db.Db, destinationUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobDestinationConnectionResponse{}), nil
	}

	job, err := s.db.Q.GetJobById(ctx, s.db.Db, destination.JobID)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(job.AccountID))
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
		neosyncdb.UUIDString(destination.ConnectionID),
		neosyncdb.UUIDString(job.AccountID)); err != nil {
		return nil, err
	}

	logger.Info("deleting job destination connection")
	err = s.db.Q.RemoveJobConnectionDestination(ctx, s.db.Db, destinationUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
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
	accountUuid, err := neosyncdb.ToUuid(accountId)
	if err != nil {
		return err
	}
	connectionUuid, err := neosyncdb.ToUuid(connectionId)
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

func verifyConnectionsInAccount(ctx context.Context, db *neosyncdb.NeosyncDb, connectionUuids []pgtype.UUID, accountUuid pgtype.UUID) (bool, error) {
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

func verifyConnectionsAreCompatible(ctx context.Context, db *neosyncdb.NeosyncDb, sourceConnId pgtype.UUID, destinations []*Destination) (bool, error) {
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
		// AWS S3 and GCP CloudStorage are always a valid destination regardless of source connection type
		if d.ConnectionConfig.AwsS3Config != nil || d.ConnectionConfig.GcpCloudStorageConfig != nil {
			continue
		}
		if sourceConnection.ConnectionConfig.PgConfig != nil && d.ConnectionConfig.PgConfig == nil {
			// invalid Postgres source cannot have Mysql destination
			return false, nil
		}
		if sourceConnection.ConnectionConfig.MysqlConfig != nil && d.ConnectionConfig.MysqlConfig == nil {
			// invalid Mysql source cannot have non-Mysql or non-AWS connection
			return false, nil
		}
		if sourceConnection.ConnectionConfig.MongoConfig != nil && d.ConnectionConfig.MongoConfig == nil {
			// invalid Mongo source cannot have anything other than mongo to start
			return false, nil
		}
		if sourceConnection.ConnectionConfig.DynamoDBConfig != nil && d.ConnectionConfig.DynamoDBConfig == nil {
			// invalid DynamoDB source cannot have anything other than dynamodb to start
			return false, nil
		}
		if sourceConnection.ConnectionConfig.MssqlConfig != nil && d.ConnectionConfig.MssqlConfig == nil {
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

	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	// update temporal scheduled job
	if err := s.db.WithTx(ctx, nil, func(dbtx neosyncdb.BaseDBTX) error {
		_, err = s.db.Q.SetJobWorkflowOptions(ctx, dbtx, db_queries.SetJobWorkflowOptionsParams{
			ID:              jobUuid,
			WorkflowOptions: wfOptions,
			UpdatedByID:     *userUuid,
		})
		if err != nil {
			return fmt.Errorf("unable to set job workflow options: %w", err)
		}

		err = s.temporalmgr.UpdateSchedule(
			ctx,
			job.Msg.GetJob().GetAccountId(),
			job.Msg.GetJob().GetId(),
			&temporalclient.ScheduleUpdateOptions{
				DoUpdate: func(schedule temporalclient.ScheduleUpdateInput) (*temporalclient.ScheduleUpdate, error) {
					action, ok := schedule.Description.Schedule.Action.(*temporalclient.ScheduleWorkflowAction)
					if !ok {
						return nil, fmt.Errorf("unable to cast temporal action to *temporalclient.ScheduleWorkflowAction. Type was: %T", schedule.Description.Schedule.Action)
					}
					action.WorkflowRunTimeout = getDurationFromInt(wfOptions.RunTimeout)
					schedule.Description.Schedule.Action = action
					return &temporalclient.ScheduleUpdate{
						Schedule: &schedule.Description.Schedule,
					}, nil
				},
			},
			logger,
		)
		if err != nil {
			logger.Error(fmt.Errorf("unable to update schedule: %w", err).Error())
			return fmt.Errorf("unable to update workflow run timeout on temporal schedule: %w", err)
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

	return connect.NewResponse(&mgmtv1alpha1.SetJobWorkflowOptionsResponse{Job: updatedJob.Msg.Job}), nil
}

func getDurationFromInt(input *int64) time.Duration {
	if input == nil {
		return time.Duration(0)
	}
	return time.Duration(*input)
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

	jobUuid, err := neosyncdb.ToUuid(req.Msg.Id)
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

func (s *Service) ValidateJobMappings(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.ValidateJobMappingsRequest],
) (*connect.Response[mgmtv1alpha1.ValidateJobMappingsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.Msg.AccountId)

	_, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}

	connConfig := connection.Msg.GetConnection().GetConnectionConfig()
	if connConfig == nil {
		return nil, errors.New("connection config for connection was nil")
	}

	if connConfig.GetAwsS3Config() != nil || connConfig.GetMongoConfig() != nil || connConfig.GetDynamodbConfig() != nil {
		return connect.NewResponse(&mgmtv1alpha1.ValidateJobMappingsResponse{}), nil
	}

	connectionTimeout := 5
	db, err := s.sqlmanager.NewSqlDb(ctx, logger, connection.Msg.GetConnection(), &connectionTimeout)
	if err != nil {
		return nil, err
	}
	defer db.Db.Close()

	colInfoMap, err := db.Db.GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, err
	}

	schemasMap := map[string]struct{}{}
	for tableName := range colInfoMap {
		schema, _ := sqlmanager_shared.SplitTableKey(tableName)
		schemasMap[schema] = struct{}{}
	}

	schemas := []string{}
	for s := range schemasMap {
		schemas = append(schemas, s)
	}

	tableConstraints, err := db.Db.GetTableConstraintsBySchema(ctx, schemas)
	if err != nil {
		return nil, err
	}

	tableColMappings := map[string]map[string]*mgmtv1alpha1.JobMapping{}
	for _, m := range req.Msg.Mappings {
		tn := sqlmanager_shared.BuildTable(m.Schema, m.Table)
		if _, ok := tableColMappings[tn]; !ok {
			tableColMappings[tn] = map[string]*mgmtv1alpha1.JobMapping{}
		}
		tableColMappings[tn][m.Column] = m
	}

	colErrorsMap := map[string]map[string][]string{}
	dbErrors := &mgmtv1alpha1.DatabaseError{
		Errors: []string{},
	}

	// verify job mapping tables
	for table := range tableColMappings {
		if _, ok := colInfoMap[table]; !ok {
			dbErrors.Errors = append(dbErrors.Errors, fmt.Sprintf("Table does not exist [%s]", table))
		}
	}

	vfkErrs := validateVirtualForeignKeys(req.Msg.GetVirtualForeignKeys(), tableColMappings, tableConstraints, colInfoMap)
	dbErrors.Errors = append(dbErrors.Errors, vfkErrs.DbErrors...)

	for t, colErrMap := range vfkErrs.ColumnErrorMap {
		if _, ok := colErrorsMap[t]; !ok {
			colErrorsMap[t] = map[string][]string{}
		}
		for c, e := range colErrMap {
			colErrorsMap[t][c] = append(colErrorsMap[t][c], e...)
		}
	}

	// verify that all circular dependencies have a nullable entrypoint
	filteredDepsMap := map[string][]string{} // only include tables that are in tables arg list
	for table, fks := range tableConstraints.ForeignKeyConstraints {
		colMappings, ok := tableColMappings[table]
		if !ok {
			// skip. table not in mapping
			continue
		}
		for _, fk := range fks {
			for idx, col := range fk.Columns {
				_, ok := colMappings[col]
				if !ok {
					continue
				}
				fkColMappings, ok := tableColMappings[fk.ForeignKey.Table]
				if !ok {
					continue
				}
				fkCol := fk.ForeignKey.Columns[idx]
				_, ok = fkColMappings[fkCol]
				if !ok {
					continue
				}
				filteredDepsMap[table] = append(filteredDepsMap[table], fk.ForeignKey.Table)
			}
		}
	}

	allForeignKeys := tableConstraints.ForeignKeyConstraints
	for _, vfk := range req.Msg.GetVirtualForeignKeys() {
		tableName := sqlmanager_shared.BuildTable(vfk.Schema, vfk.Table)
		fkTable := sqlmanager_shared.BuildTable(vfk.ForeignKey.Schema, vfk.ForeignKey.Table)
		filteredDepsMap[tableName] = append(filteredDepsMap[tableName], fkTable)

		// merge virtual foreign keys with table foreign keys
		tableCols, ok := colInfoMap[tableName]
		if !ok {
			continue
		}
		notNullable := []bool{}
		for _, col := range vfk.GetColumns() {
			colInfo, ok := tableCols[col]
			if !ok {
				return nil, fmt.Errorf("Column does not exist in schema: %s.%s", tableName, col)
			}
			notNullable = append(notNullable, !colInfo.IsNullable)
		}

		allForeignKeys[tableName] = append(allForeignKeys[tableName], &sqlmanager_shared.ForeignConstraint{
			Columns:     vfk.GetColumns(),
			NotNullable: notNullable,
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Columns: vfk.GetColumns(),
				Table:   fkTable,
			},
		})
	}

	for table, deps := range filteredDepsMap {
		filteredDepsMap[table] = utils.DedupeSlice(deps)
	}

	cycles := tabledependency.FindCircularDependencies(filteredDepsMap)
	startTables, err := tabledependency.DetermineCycleStarts(cycles, map[string]string{}, allForeignKeys)
	if err != nil {
		return nil, err
	}

	containsStart := func(t string) bool {
		return slices.Contains(startTables, t)
	}

	for _, cycle := range cycles {
		if !slices.ContainsFunc(cycle, containsStart) {
			dbErrors.Errors = append(dbErrors.Errors, fmt.Sprintf("Unsupported circular dependency. At least one foreign key in circular dependency must be nullable. Tables: %+v", cycle))
		}
	}

	// verify that all non nullable foreign key constraints are not missing from mapping
	for table, fks := range tableConstraints.ForeignKeyConstraints {
		_, ok := tableColMappings[table]
		if !ok {
			// skip. table not in mapping
			continue
		}
		for _, fk := range fks {
			for idx, notNull := range fk.NotNullable {
				if !notNull {
					// skip. foreign key is nullable
					continue
				}
				fkColMappings, ok := tableColMappings[fk.ForeignKey.Table]
				fkCol := fk.ForeignKey.Columns[idx]
				if !ok {
					if _, ok := colErrorsMap[fk.ForeignKey.Table]; !ok {
						colErrorsMap[fk.ForeignKey.Table] = map[string][]string{}
					}
					colErrorsMap[fk.ForeignKey.Table][fkCol] = append(colErrorsMap[fk.ForeignKey.Table][fkCol], fmt.Sprintf("Missing required foreign key. Table: %s  Column: %s", fk.ForeignKey.Table, fkCol))
					continue
				}
				_, ok = fkColMappings[fkCol]
				if !ok {
					if _, ok := colErrorsMap[fk.ForeignKey.Table]; !ok {
						colErrorsMap[fk.ForeignKey.Table] = map[string][]string{}
					}
					colErrorsMap[fk.ForeignKey.Table][fkCol] = append(colErrorsMap[fk.ForeignKey.Table][fkCol], fmt.Sprintf("Missing required foreign key. Table: %s  Column: %s", fk.ForeignKey.Table, fkCol))
				}
			}
		}
	}

	// verify that no non nullable columns are missing for tables in mapping
	for table, colMap := range colInfoMap {
		cm, ok := tableColMappings[table]
		if !ok {
			// skip. table not in mapping
			continue
		}
		for col, info := range colMap {
			if info.IsNullable {
				// skip. column is nullable
				continue
			}
			if _, ok := cm[col]; !ok {
				if _, ok := colErrorsMap[table]; !ok {
					colErrorsMap[table] = map[string][]string{}
				}
				// add error
				colErrorsMap[table][col] = append(colErrorsMap[table][col], fmt.Sprintf("Violates not-null constraint. Missing required column. Table: %s  Column: %s", table, col))
			}
		}
	}

	colErrors := []*mgmtv1alpha1.ColumnError{}
	for tableName, colMap := range colErrorsMap {
		for col, errors := range colMap {
			schema, table := sqlmanager_shared.SplitTableKey(tableName)
			colErrors = append(colErrors, &mgmtv1alpha1.ColumnError{
				Schema: schema,
				Table:  table,
				Column: col,
				Errors: errors,
			})
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.ValidateJobMappingsResponse{
		DatabaseErrors: dbErrors,
		ColumnErrors:   colErrors,
	}), nil
}

type validateVirtualForeignKeysResponse struct {
	IsValid        bool
	ColumnErrorMap map[string]map[string][]string
	DbErrors       []string
}

func validateVirtualForeignKeys(
	virtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint,
	jobColMappings map[string]map[string]*mgmtv1alpha1.JobMapping,
	tc *sqlmanager_shared.TableConstraints,
	colMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) *validateVirtualForeignKeysResponse {
	dbErrors := []string{}
	colErrorsMap := map[string]map[string][]string{}
	isValid := true

	for _, vfk := range virtualForeignKeys {
		sourceTable := fmt.Sprintf("%s.%s", vfk.ForeignKey.Schema, vfk.ForeignKey.Table)
		targetTable := fmt.Sprintf("%s.%s", vfk.Schema, vfk.Table)

		// check that source table exist in job mappings
		sourceColMappings, ok := jobColMappings[sourceTable]
		if !ok {
			isValid = false
			dbErrors = append(dbErrors, fmt.Sprintf("Virtual foreign key source table missing in job mappings. Table: %s", sourceTable))
			continue
		}

		// check that source table and columns exist in source
		sourceCols, ok := colMap[sourceTable]
		if !ok {
			isValid = false
			dbErrors = append(dbErrors, fmt.Sprintf("Virtual foreign key source table missing in source database. Table: %s", sourceTable))
			continue
		}
		for _, c := range vfk.GetForeignKey().GetColumns() {
			_, ok = sourceColMappings[c]
			if !ok {
				isValid = false
				if _, ok := colErrorsMap[sourceTable]; !ok {
					colErrorsMap[sourceTable] = map[string][]string{}
				}
				colErrorsMap[sourceTable][c] = append(colErrorsMap[sourceTable][c], fmt.Sprintf("Virtual foreign key source column missing in job mappings. Table: %s Column: %s", sourceTable, c))
			}
			_, ok := sourceCols[c]
			if !ok {
				isValid = false
				if _, ok := colErrorsMap[sourceTable]; !ok {
					colErrorsMap[sourceTable] = map[string][]string{}
				}
				colErrorsMap[sourceTable][c] = append(colErrorsMap[sourceTable][c], fmt.Sprintf("Virtual foreign key source column missing in source database. Table: %s Column: %s", sourceTable, c))
			}
		}
		// check that all sources of virtual foreign keys are either a primary key or have a unique constraint
		pks := tc.PrimaryKeyConstraints[sourceTable]
		uniqueConstraints := tc.UniqueConstraints[sourceTable]
		isVfkValid := isVirtualForeignKeySourceUnique(vfk, pks, uniqueConstraints)
		if !isVfkValid {
			isValid = false
			if _, ok := colErrorsMap[sourceTable]; !ok {
				colErrorsMap[sourceTable] = map[string][]string{}
			}
			for _, c := range vfk.GetForeignKey().GetColumns() {
				colErrorsMap[sourceTable][c] = append(colErrorsMap[sourceTable][c], fmt.Sprintf("Virtual foreign key source must be either a primary key or have a unique constraint. Table: %s  Columns: %+v", sourceTable, vfk.GetForeignKey().GetColumns()))
			}
		}

		// check that target table exist in job mappings
		targetColMappings, ok := jobColMappings[targetTable]
		if !ok {
			isValid = false
			dbErrors = append(dbErrors, fmt.Sprintf("Virtual foreign key target table missing in job mappings. Table: %s", targetTable))
			continue
		}
		targetCols, ok := colMap[targetTable]
		if !ok {
			isValid = false
			dbErrors = append(dbErrors, fmt.Sprintf("Virtual foreign key target table missing in source database. Table: %s", targetTable))
			continue
		}
		// check that all self referencing virtual foreign keys are on nullable columns
		if sourceTable == targetTable {
			for _, c := range vfk.GetColumns() {
				_, ok = targetColMappings[c]
				if !ok {
					isValid = false
					if _, ok := colErrorsMap[targetTable]; !ok {
						colErrorsMap[targetTable] = map[string][]string{}
					}
					colErrorsMap[targetTable][c] = append(colErrorsMap[targetTable][c], fmt.Sprintf("Virtual foreign key target column missing in job mappings. Table: %s Column: %s", targetTable, c))
				}
				colInfo, ok := targetCols[c]
				if !ok {
					isValid = false
					if _, ok := colErrorsMap[targetTable]; !ok {
						colErrorsMap[targetTable] = map[string][]string{}
					}
					colErrorsMap[targetTable][c] = append(colErrorsMap[targetTable][c], fmt.Sprintf("Virtual foreign key target column missing in source database. Table: %s Column: %s", targetTable, c))
					continue
				}
				if !colInfo.IsNullable {
					isValid = false
					if _, ok := colErrorsMap[targetTable]; !ok {
						colErrorsMap[targetTable] = map[string][]string{}
					}
					colErrorsMap[targetTable][c] = append(colErrorsMap[targetTable][c], fmt.Sprintf("Self referencing virtual foreign key target column must be nullable. Table: %s  Column: %s", targetTable, c))
				}
			}
		}

		if len(vfk.GetColumns()) != len(vfk.GetForeignKey().GetColumns()) {
			isValid = false
			dbErrors = append(dbErrors, fmt.Sprintf("length of source columns was not equal to length of foreign key cols: %d %d. SourceTable: %s SourceColumn: %+v TargetTable: %s  TargetColumn: %+v", len(vfk.GetColumns()), len(vfk.GetForeignKey().GetColumns()), sourceTable, vfk.GetColumns(), targetTable, vfk.GetForeignKey().GetColumns()))
			continue
		}
		// check that source and target column datatypes are the same
		for idx, srcCol := range vfk.GetForeignKey().GetColumns() {
			tarCol := vfk.GetColumns()[idx]
			srcColInfo := sourceCols[srcCol]
			tarColInfo := targetCols[tarCol]
			if srcColInfo.DataType != tarColInfo.DataType {
				isValid = false
				if _, ok := colErrorsMap[targetTable]; !ok {
					colErrorsMap[targetTable] = map[string][]string{}
				}
				colErrorsMap[targetTable][tarCol] = append(colErrorsMap[targetTable][tarCol], fmt.Sprintf("Column datatype mismatch. Source: %s.%s %s Target: %s.%s %s", sourceTable, srcCol, srcColInfo.DataType, targetTable, tarCol, tarColInfo.DataType))
			}
		}
	}
	return &validateVirtualForeignKeysResponse{
		IsValid:        isValid,
		ColumnErrorMap: colErrorsMap,
		DbErrors:       dbErrors,
	}
}

func isVirtualForeignKeySourceUnique(
	virtualForeignKey *mgmtv1alpha1.VirtualForeignConstraint,
	primaryKeys []string,
	uniqueConstraints [][]string,
) bool {
	if utils.CompareSlices(virtualForeignKey.GetForeignKey().GetColumns(), primaryKeys) {
		return true
	}
	for _, uc := range uniqueConstraints {
		if utils.CompareSlices(virtualForeignKey.GetForeignKey().GetColumns(), uc) {
			return true
		}
	}
	return false
}

func getJobSourceConnectionId(jobSource *mgmtv1alpha1.JobSource) (*string, error) {
	var connectionIdToVerify *string
	switch config := jobSource.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Mongodb:
		connectionIdToVerify = &config.Mongodb.ConnectionId
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
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		fkConnId := config.AiGenerate.GetFkSourceConnectionId()
		if fkConnId != "" {
			connectionIdToVerify = &fkConnId
		}
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		connId := config.Dynamodb.GetConnectionId()
		connectionIdToVerify = &connId
	case *mgmtv1alpha1.JobSourceOptions_Mssql:
		connId := config.Mssql.GetConnectionId()
		connectionIdToVerify = &connId
	default:
		return nil, fmt.Errorf("unsupported source option config type: %T", config)
	}
	return connectionIdToVerify, nil
}
