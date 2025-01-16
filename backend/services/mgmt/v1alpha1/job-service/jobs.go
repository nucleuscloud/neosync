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
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	job_mappings "github.com/nucleuscloud/neosync/internal/job-mappings"
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

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(req.Msg.GetAccountId()), rbac.JobAction_View)
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, accountUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to get jobs by account: %w", err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetJobsResponse{Jobs: []*mgmtv1alpha1.Job{}}), nil
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
		if err != nil && !neosyncdb.IsNoRows(err) {
			return nil, fmt.Errorf("unable to get job connection destinations by job ids: %w", err)
		} else if err != nil && neosyncdb.IsNoRows(err) {
			logger.Debug("found no job connection destinations by job ids")
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
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	jobUuid, err := neosyncdb.ToUuid(req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	errgrp, errctx := errgroup.WithContext(ctx)

	var job db_queries.NeosyncApiJob
	errgrp.Go(func() error {
		j, err := s.db.Q.GetJobById(errctx, s.db.Db, jobUuid)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return fmt.Errorf("unable to get job by id: %w", err)
		} else if err != nil && neosyncdb.IsNoRows(err) {
			return nucleuserrors.NewNotFound("job with that id does not exist")
		}
		job = j
		return nil
	})
	var destConnections []db_queries.NeosyncApiJobDestinationConnectionAssociation
	errgrp.Go(func() error {
		dcs, err := s.db.Q.GetJobConnectionDestinations(ctx, s.db.Db, jobUuid)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return fmt.Errorf("unable to get job connection destinations by job id: %w", err)
		} else if err != nil && neosyncdb.IsNoRows(err) {
			return nil
		}
		destConnections = dcs
		return nil
	})
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	if err := user.EnforceJob(ctx, userdata.NewDbDomainEntity(job.AccountID, job.ID), rbac.JobAction_View); err != nil {
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
	logger = logger.With("jobId", req.Msg.GetJobId())
	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.Msg.GetJobId()}))
	if err != nil {
		return nil, err
	}

	schedule, err := s.temporalmgr.DescribeSchedule(ctx, jobResp.Msg.GetJob().GetAccountId(), jobResp.Msg.GetJob().GetId(), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to describe temporal schedule when retrieving job status: %w", err)
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

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(req.Msg.GetAccountId()), rbac.JobAction_View)
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	jobs, err := s.db.Q.GetJobsByAccount(ctx, s.db.Db, accountUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to get jobs by account: %w", err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetJobStatusesResponse{Statuses: []*mgmtv1alpha1.JobStatusRecord{}}), nil
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
		return nil, fmt.Errorf("unable to describe temporal schedules when retrieving job statuses: %w", err)
	}

	dtos := make([]*mgmtv1alpha1.JobStatusRecord, len(jobs))
	for idx, resp := range responses {
		if resp.Error != nil {
			logger.Warn(fmt.Errorf("unable to describe temporal schedule when retrieving job statuses: %w", resp.Error).Error())
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

	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.Msg.GetJobId()}))
	if err != nil {
		return nil, err
	}

	schedule, err := s.temporalmgr.DescribeSchedule(ctx, jobResp.Msg.GetJob().GetAccountId(), jobResp.Msg.GetJob().GetId(), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to describe temporal schedule when retrieving job recent runs: %w", err)
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
	logger = logger.With("jobId", req.Msg.GetJobId())

	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.Msg.GetJobId()}))
	if err != nil {
		return nil, err
	}

	schedule, err := s.temporalmgr.DescribeSchedule(ctx, jobResp.Msg.GetJob().GetAccountId(), jobResp.Msg.GetJob().GetId(), logger)
	if err != nil {
		return nil, fmt.Errorf("unable to describe temporal schedule when retrieving job next runs: %w", err)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobNextRunsResponse{
		NextRuns: dtomaps.ToJobNextRunsDto(schedule),
	}), nil
}

type destination struct {
	ConnectionId pgtype.UUID
	Options      *pg_models.JobDestinationOptions
}

func (s *Service) CreateJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobName", req.Msg.JobName, "accountId", req.Msg.AccountId)

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(req.Msg.GetAccountId()), rbac.JobAction_Create)
	if err != nil {
		return nil, err
	}
	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	connectionUuids := []pgtype.UUID{}
	connectionIds := []string{}
	destinations := []*destination{}
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
		destinations = append(destinations, &destination{ConnectionId: destUuid, Options: options})
		connectionIds = append(connectionIds, dest.ConnectionId)
		connectionUuids = append(connectionUuids, destUuid)
	}

	logger.Debug("verifying connections")
	count, err := s.db.Q.AreConnectionsInAccount(ctx, s.db.Db, db_queries.AreConnectionsInAccountParams{
		AccountId:     accountUuid,
		ConnectionIds: connectionUuids,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to check if connections are in provided account: %w", err)
	}
	if count != int64(len(connectionUuids)) {
		return nil, nucleuserrors.NewForbidden("provided connection id(s) are not all in account")
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
			return nil, fmt.Errorf("unable to verify if all connections are compatible: %w", err)
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

	logger.Debug("verifying temporal workspace")
	hasNs, err := s.temporalmgr.DoesAccountHaveNamespace(ctx, req.Msg.AccountId, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to verify account's temporal workspace. error: %w", err)
	}
	if !hasNs {
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
		AccountID:          accountUuid,
		Status:             int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
		CronSchedule:       cronText,
		ConnectionOptions:  connectionOptions,
		Mappings:           mappings,
		VirtualForeignKeys: virtualForeignKeys,
		CreatedByID:        user.PgId(),
		UpdatedByID:        user.PgId(),
		WorkflowOptions:    workflowOptions,
		SyncOptions:        activitySyncOptions,
	}, connDestParams)
	if err != nil {
		return nil, fmt.Errorf("unable to create job: %w", err)
	}
	jobUuid := neosyncdb.UUIDString(cj.ID)
	logger = logger.With("jobId", jobUuid)
	logger.Debug("created job record")

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
		logger.Debug("deleting newly created job")
		removeJobErr := s.db.Q.RemoveJobById(ctx, s.db.Db, cj.ID)
		if removeJobErr != nil {
			return nil, fmt.Errorf("unable to create scheduled job and was unable to fully cleanup partially created resources: %w: %w", removeJobErr, err)
		}
		return nil, fmt.Errorf("unable to create scheduled job: %w", err)
	}
	logger = logger.With("scheduleId", scheduleId)
	logger.Debug("created new temporal schedule")

	if req.Msg.InitiateJobRun {
		logger.Debug("triggering initial job run")
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

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, userdata.NewDbDomainEntity(job.AccountID, job.ID), rbac.JobAction_Delete)
	if err != nil {
		return nil, err
	}

	logger.Debug("deleting temporal schedule")
	err = s.temporalmgr.DeleteSchedule(
		ctx,
		neosyncdb.UUIDString(job.AccountID),
		neosyncdb.UUIDString(job.ID),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to remove schedule when deleting job")
	}

	logger.Debug("deleting job")
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
		return nil, err
	}
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job.Msg.GetJob(), rbac.JobAction_Create)
	if err != nil {
		return nil, err
	}
	accountUuid, err := neosyncdb.ToUuid(job.Msg.GetJob().GetAccountId())
	if err != nil {
		return nil, err
	}

	connectionIds := []string{}
	connectionUuids := []pgtype.UUID{}
	destinations := []*destination{}
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
		destinations = append(destinations, &destination{ConnectionId: destUuid, Options: options})
		connectionIds = append(connectionIds, dest.ConnectionId)
		connectionUuids = append(connectionUuids, destUuid)
	}

	if !verifyConnectionIdsUnique(connectionIds) {
		return nil, nucleuserrors.NewBadRequest("connections ids are not unique")
	}

	isInSameAccount, err := verifyConnectionsInAccount(ctx, s.db, connectionUuids, accountUuid)
	if err != nil {
		return nil, err
	}
	if !isInSameAccount {
		return nil, nucleuserrors.NewBadRequest("connections are not all within the provided account")
	}

	logger.Debug("creating job destination connections", "connectionIds", connectionIds)
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
	logger.Debug("updating job schedule")

	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job, rbac.JobAction_Edit)
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

	jobUuid, err := neosyncdb.ToUuid(job.GetId())
	if err != nil {
		return nil, err
	}

	if err := s.db.WithTx(ctx, nil, func(dbtx neosyncdb.BaseDBTX) error {
		_, err = s.db.Q.UpdateJobSchedule(ctx, dbtx, db_queries.UpdateJobScheduleParams{
			ID:           jobUuid,
			CronSchedule: cronText,
			UpdatedByID:  user.PgId(),
		})
		if err != nil {
			return err
		}

		spec := &temporalclient.ScheduleSpec{}
		spec.CronExpressions = []string{cronStr}

		// update temporal scheduled job
		err = s.temporalmgr.UpdateSchedule(
			ctx,
			job.GetAccountId(),
			job.GetId(),
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
			return fmt.Errorf("unable to update temporal schedule: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
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
	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job, rbac.JobAction_Edit)
	if err != nil {
		return nil, err
	}

	if req.Msg.Pause {
		logger.Debug("pausing job")
		err = s.temporalmgr.PauseSchedule(
			ctx,
			job.GetAccountId(),
			job.GetId(),
			&temporalclient.SchedulePauseOptions{Note: req.Msg.GetNote()},
			logger,
		)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Debug("unpausing job")
		err = s.temporalmgr.UnpauseSchedule(
			ctx,
			job.GetAccountId(),
			job.GetId(),
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
	logger.Debug("updating job source connection and mappings")

	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job, rbac.JobAction_Edit)
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
	if err := s.verifyConnectionInAccount(ctx, connectionIdToVerify, job.GetAccountId()); err != nil {
		return nil, err
	}

	// retrieves the connection details
	conn, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionIdToVerify,
	}))
	if err != nil {
		return nil, err
	}

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

	jobUuid, err := neosyncdb.ToUuid(job.GetId())
	if err != nil {
		return nil, err
	}

	if err := s.db.WithTx(ctx, nil, func(dbtx neosyncdb.BaseDBTX) error {
		_, err = s.db.Q.UpdateJobSource(ctx, dbtx, db_queries.UpdateJobSourceParams{
			ID:                jobUuid,
			ConnectionOptions: connectionOptions,

			UpdatedByID: user.PgId(),
		})
		if err != nil {
			return fmt.Errorf("unable to update job source: %w", err)
		}

		_, err = s.db.Q.UpdateJobMappings(ctx, dbtx, db_queries.UpdateJobMappingsParams{
			ID:          jobUuid,
			Mappings:    mappings,
			UpdatedByID: user.PgId(),
		})
		if err != nil {
			return fmt.Errorf("unable to update job mappings: %w", err)
		}

		args := db_queries.UpdateJobVirtualForeignKeysParams{
			VirtualForeignKeys: virtualForeignKeys,
			UpdatedByID:        user.PgId(),
			ID:                 jobUuid,
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
	logger.Debug("updating job source sql connection subsets")

	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()
	jobUuid, err := neosyncdb.ToUuid(job.GetId())
	if err != nil {
		return nil, err
	}
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job, rbac.JobAction_Edit)
	if err != nil {
		return nil, err
	}

	var connectionId *string
	if job.GetSource().GetOptions() != nil {
		if job.GetSource().GetOptions().GetMysql() != nil {
			connectionId = &job.GetSource().GetOptions().GetMysql().ConnectionId
		} else if job.GetSource().GetOptions().GetPostgres() != nil {
			connectionId = &job.GetSource().GetOptions().GetPostgres().ConnectionId
		} else if job.GetSource().GetOptions().GetDynamodb() != nil {
			connectionId = &job.GetSource().GetOptions().GetDynamodb().ConnectionId
		} else if job.GetSource().GetOptions().GetMssql() != nil {
			connectionId = &job.GetSource().GetOptions().GetMssql().ConnectionId
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
		user.PgId(),
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
	logger = logger.With("jobId", req.Msg.GetJobId(), "connectionId", req.Msg.GetConnectionId())

	jobUuid, err := neosyncdb.ToUuid(req.Msg.GetJobId())
	if err != nil {
		return nil, err
	}
	destinationUuid, err := neosyncdb.ToUuid(req.Msg.GetDestinationId())
	if err != nil {
		return nil, err
	}
	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.JobId,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job, rbac.JobAction_Edit)
	if err != nil {
		return nil, err
	}

	connectionUuid, err := neosyncdb.ToUuid(req.Msg.GetConnectionId())
	if err != nil {
		return nil, err
	}
	if err := s.verifyConnectionInAccount(ctx, req.Msg.GetConnectionId(), job.GetAccountId()); err != nil {
		return nil, err
	}
	options := &pg_models.JobDestinationOptions{}
	err = options.FromDto(req.Msg.Options)
	if err != nil {
		return nil, err
	}

	// todo(NEOS-1281):  need a lot more validation here for changing connection uuid, matching options, as well as creating a new destination
	// if that destination is not supported with the source type
	logger.Debug("updating job destination connection")
	_, err = s.db.Q.UpdateJobConnectionDestination(ctx, s.db.Db, db_queries.UpdateJobConnectionDestinationParams{
		ID:           destinationUuid,
		ConnectionID: connectionUuid,
		Options:      options,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		logger.Debug("destination not found. creating job destination connection")
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
		Id: job.GetId(),
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
	logger = logger.With("destinationId", req.Msg.GetDestinationId())

	destinationUuid, err := neosyncdb.ToUuid(req.Msg.GetDestinationId())
	if err != nil {
		return nil, err
	}

	destination, err := s.db.Q.GetJobConnectionDestination(ctx, s.db.Db, destinationUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobDestinationConnectionResponse{}), nil
	}

	jobId := neosyncdb.UUIDString(destination.JobID)

	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.GetJob()

	logger = logger.With("jobId", job.GetId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job, rbac.JobAction_Edit)
	if err != nil {
		return nil, err
	}

	logger.Debug("deleting job destination connection")
	err = s.db.Q.RemoveJobConnectionDestination(ctx, s.db.Db, destinationUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		logger.Debug("destination not found, during delete")
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteJobDestinationConnectionResponse{}), nil
}

func (s *Service) IsJobNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsJobNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsJobNameAvailableResponse], error) {
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(req.Msg.GetAccountId()), rbac.JobAction_View); err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsJobNameAvailable(ctx, s.db.Db, db_queries.IsJobNameAvailableParams{
		AccountId: accountUuid,
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

func verifyConnectionsAreCompatible(ctx context.Context, db *neosyncdb.NeosyncDb, sourceConnId pgtype.UUID, destinations []*destination) (bool, error) {
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
		return nil, err
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, job.Msg.GetJob(), rbac.JobAction_Edit)
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

	// update temporal scheduled job
	if err := s.db.WithTx(ctx, nil, func(dbtx neosyncdb.BaseDBTX) error {
		_, err = s.db.Q.SetJobWorkflowOptions(ctx, dbtx, db_queries.SetJobWorkflowOptionsParams{
			ID:              jobUuid,
			WorkflowOptions: wfOptions,
			UpdatedByID:     user.PgId(),
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
	jobResp, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, jobResp.Msg.GetJob(), rbac.JobAction_Edit)
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

	_, err = s.db.Q.SetJobSyncOptions(ctx, s.db.Db, db_queries.SetJobSyncOptionsParams{
		ID:          jobUuid,
		SyncOptions: syncOptions,
		UpdatedByID: user.PgId(),
	})
	if err != nil {
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.GetId(),
	}))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.SetJobSyncOptionsResponse{Job: updatedJob.Msg.Job}), nil
}

func (s *Service) ValidateJobMappings(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.ValidateJobMappingsRequest],
) (*connect.Response[mgmtv1alpha1.ValidateJobMappingsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.Msg.GetAccountId())

	connection, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.GetConnectionId(),
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

	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), connection.Msg.GetConnection(), logger)
	if err != nil {
		return nil, err
	}
	defer db.Db().Close()

	colInfoMap, err := db.Db().GetSchemaColumnMap(ctx)
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

	tableConstraints, err := db.Db().GetTableConstraintsBySchema(ctx, schemas)
	if err != nil {
		return nil, err
	}

	validator := job_mappings.NewJobMappingsValidator(req.Msg.Mappings)
	result, err := validator.Validate(colInfoMap, req.Msg.VirtualForeignKeys, tableConstraints)
	if err != nil {
		return nil, err
	}

	dbErrors := &mgmtv1alpha1.DatabaseError{
		Errors: []string{},
	}
	dbErrors.Errors = append(dbErrors.Errors, result.DatabaseErrors...)

	colErrors := []*mgmtv1alpha1.ColumnError{}
	for tableName, colMap := range result.ColumnErrors {
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
