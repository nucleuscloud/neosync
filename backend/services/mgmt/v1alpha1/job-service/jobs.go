package v1alpha1_jobservice

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"

	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"

	wf_datasync "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync"
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
	jobs, err := s.db.Q.GetJobsByAccount(ctx, *accountUuid)
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
		destinationAssociations, err = s.db.Q.GetJobConnectionDestinationsByJobIds(ctx, jobIds)
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
		j, err := s.db.Q.GetJobById(errctx, jobUuid)
		if err != nil {
			return err
		}
		job = j
		return nil
	})
	var destConnections []db_queries.NeosyncApiJobDestinationConnectionAssociation
	errgrp.Go(func() error {
		dcs, err := s.db.Q.GetJobConnectionDestinations(ctx, jobUuid)
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

type Destination struct {
	ConnectionId pgtype.UUID
	Options      *jsonmodels.JobDestinationOptions
}

func (s *Service) CreateJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobName", req.Msg.JobName)

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	connectionIds := []string{req.Msg.Source.ConnectionId}
	destinations := []*Destination{}
	for _, dest := range req.Msg.Destinations {
		destUuid, err := nucleusdb.ToUuid(dest.ConnectionId)
		if err != nil {
			return nil, err
		}
		options := &jsonmodels.JobDestinationOptions{}
		err = options.FromDto(dest.Options)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, &Destination{ConnectionId: destUuid, Options: options})
		connectionIds = append(connectionIds, dest.ConnectionId)
	}

	if !verifyConnectionIdsUnique(connectionIds) {
		return nil, nucleuserrors.NewBadRequest("connections ids are not unique")
	}

	cron := pgtype.Text{}
	if req.Msg.CronSchedule != nil {
		err := cron.Scan(req.Msg.GetCronSchedule())
		if err != nil {
			return nil, err
		}
	}
	connectionSourceUuid, err := nucleusdb.ToUuid(req.Msg.Source.ConnectionId)
	if err != nil {
		return nil, err
	}

	mappings := []*jsonmodels.JobMapping{}
	for _, mapping := range req.Msg.Mappings {
		jm := &jsonmodels.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	connectionOptions := &jsonmodels.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.Source.Options)
	if err != nil {
		return nil, err
	}

	// todo: verify connection ids are all in this account

	var createdJob *db_queries.NeosyncApiJob
	if err := s.db.WithTx(ctx, nil, func(q *db_queries.Queries) error {
		job, err := q.CreateJob(ctx, db_queries.CreateJobParams{
			Name:               req.Msg.JobName,
			AccountID:          *accountUuid,
			Status:             int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
			CronSchedule:       cron,
			ConnectionSourceID: connectionSourceUuid,
			ConnectionOptions:  connectionOptions,
			Mappings:           mappings,
			CreatedByID:        *userUuid,
			UpdatedByID:        *userUuid,
		})
		if err != nil {
			return err
		}

		connDestParams := []db_queries.CreateJobConnectionDestinationsParams{}
		for _, dest := range destinations {
			connDestParams = append(connDestParams, db_queries.CreateJobConnectionDestinationsParams{
				JobID:        job.ID,
				ConnectionID: dest.ConnectionId,
				Options:      dest.Options,
			})
		}
		if len(connDestParams) > 0 {
			_, err = q.CreateJobConnectionDestinations(ctx, connDestParams)
			if err != nil {
				return err
			}
		}
		createdJob = &job

		jobUuid := nucleusdb.UUIDString(createdJob.ID)
		logger = logger.With("jobId", jobUuid)
		schedule := nucleusdb.ToNullableString(createdJob.CronSchedule)
		paused := true
		spec := temporalclient.ScheduleSpec{}
		if schedule != nil && *schedule != "" {
			spec.CronExpressions = []string{*schedule}
			paused = false
		}

		// schedule will not run if no spec is defined
		scheduleHandle, err := s.temporalClient.ScheduleClient().Create(ctx, temporalclient.ScheduleOptions{
			ID:     jobUuid,
			Spec:   spec,
			Paused: paused,
			Action: &temporalclient.ScheduleWorkflowAction{
				Workflow:  wf_datasync.Workflow,
				TaskQueue: s.cfg.TemporalTaskQueue,
				Args:      []any{&wf_datasync.WorkflowRequest{JobId: jobUuid}},
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("unable to create schedule workflow: %w", err).Error())
			return err
		}
		logger.Info("scheduled workflow", "workflowId", scheduleHandle.GetID())

		if paused {
			// one off job, manually trigger run
			err := scheduleHandle.Trigger(ctx, temporalclient.ScheduleTriggerOptions{})
			if err != nil {
				logger.Error(fmt.Errorf("unable to trigger job: %w", err).Error())
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	destinationConnections, err := s.db.Q.GetJobConnectionDestinations(ctx, createdJob.ID)
	if err != nil {
		logger.Error("unable to retrieve job destination connections")
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateJobResponse{
		Job: dtomaps.ToJobDto(createdJob, destinationConnections),
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

	job, err := s.db.Q.GetJobById(ctx, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	logger.Info("deleting schedule's workflow executions")
	workflows, err := getWorkflowExecutionsByJobIds(ctx, s.temporalClient, logger, s.cfg.TemporalNamespace, []string{req.Msg.Id})
	if err != nil {
		return nil, err
	}

	group := new(errgroup.Group)
	for _, w := range workflows {
		w := w
		group.Go(func() error {
			_, err := s.temporalClient.WorkflowService().DeleteWorkflowExecution(ctx, &workflowservice.DeleteWorkflowExecutionRequest{
				Namespace:         s.cfg.TemporalNamespace,
				WorkflowExecution: w.Execution,
			})
			if err != nil {
				return err
			}
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		logger.Error(fmt.Errorf("unable to delete schedule's workflow executions: %w", err).Error())
		return nil, err
	}

	logger.Info("deleting schedule")
	scheduleHandle := s.temporalClient.ScheduleClient().GetHandle(ctx, nucleusdb.UUIDString(job.ID))
	err = scheduleHandle.Delete(ctx)
	if err != nil {
		logger.Error(fmt.Errorf("unable to delete schedule: %w", err).Error())
		return nil, err
	}

	logger.Info("deleting job")
	err = s.db.Q.RemoveJobById(ctx, job.ID)
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
	_, err = s.verifyUserInAccount(ctx, job.Msg.Job.AccountId)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	logger = logger.With("userId", userUuid)

	connectionIds := []string{}
	destinations := []*Destination{}
	for _, dest := range req.Msg.Destinations {
		destUuid, err := nucleusdb.ToUuid(dest.ConnectionId)
		if err != nil {
			return nil, err
		}
		options := &jsonmodels.JobDestinationOptions{}
		err = options.FromDto(dest.Options)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, &Destination{ConnectionId: destUuid, Options: options})
		connectionIds = append(connectionIds, dest.ConnectionId)
	}

	if !verifyConnectionIdsUnique(connectionIds) {
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
		_, err = s.db.Q.CreateJobConnectionDestinations(ctx, connDestParams)
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
	job, err := s.db.Q.GetJobById(ctx, jobUuid)
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

	_, err = s.db.Q.UpdateJobSchedule(ctx, db_queries.UpdateJobScheduleParams{
		ID:           job.ID,
		CronSchedule: cron,
		UpdatedByID:  *userUuid,
	})
	if err != nil {
		return nil, err
	}

	spec := &temporalclient.ScheduleSpec{}
	paused := true
	if req.Msg.CronSchedule != nil && *req.Msg.CronSchedule != "" {
		paused = false
		spec.CronExpressions = []string{*req.Msg.CronSchedule}
	}

	// update temporal scheduled job
	scheduleHandle := s.temporalClient.ScheduleClient().GetHandle(ctx, nucleusdb.UUIDString(job.ID))
	err = scheduleHandle.Update(ctx, temporalclient.ScheduleUpdateOptions{
		DoUpdate: func(schedule temporalclient.ScheduleUpdateInput) (*temporalclient.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = spec
			schedule.Description.Schedule.State.Paused = paused
			return &temporalclient.ScheduleUpdate{
				Schedule: &schedule.Description.Schedule,
			}, nil
		},
	})
	// todo handle roll back if this fails
	if err != nil {
		logger.Error(fmt.Errorf("unable to update schedule: %w", err).Error())
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
	job, err := s.db.Q.GetJobById(ctx, jobUuid)
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

	connectionUuid, err := nucleusdb.ToUuid(req.Msg.Source.ConnectionId)
	if err != nil {
		return nil, err
	}

	if err := s.verifyConnectionInAccount(ctx, req.Msg.Source.ConnectionId, nucleusdb.UUIDString(job.AccountID)); err != nil {
		return nil, err
	}

	connectionOptions := &jsonmodels.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.Source.Options)
	if err != nil {
		return nil, err
	}

	mappings := []*jsonmodels.JobMapping{}
	for _, mapping := range req.Msg.Mappings {
		jm := &jsonmodels.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	if err := s.db.WithTx(ctx, nil, func(q *db_queries.Queries) error {
		_, err = q.UpdateJobSource(ctx, db_queries.UpdateJobSourceParams{
			ID:                 job.ID,
			ConnectionSourceID: connectionUuid,
			ConnectionOptions:  connectionOptions,

			UpdatedByID: *userUuid,
		})
		if err != nil {
			logger.Error(fmt.Errorf("unable to update job source: %w", err).Error())
			return err
		}

		_, err = q.UpdateJobMappings(ctx, db_queries.UpdateJobMappingsParams{
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
	options := &jsonmodels.JobDestinationOptions{}
	err = options.FromDto(req.Msg.Options)
	if err != nil {
		return nil, err
	}

	logger.Info("updating job destination connection")
	_, err = s.db.Q.UpdateJobConnectionDestination(ctx, db_queries.UpdateJobConnectionDestinationParams{
		ID:           destinationUuid,
		ConnectionID: connectionUuid,
		Options:      options,
	})
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		logger.Info("destination not found. creating job destination connection")
		_, err = s.db.Q.CreateJobConnectionDestination(ctx, db_queries.CreateJobConnectionDestinationParams{
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

	destination, err := s.db.Q.GetJobConnectionDestination(ctx, destinationUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job destination by id")
	}

	job, err := s.db.Q.GetJobById(ctx, destination.JobID)
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
	err = s.db.Q.RemoveJobConnectionDestination(ctx, destinationUuid)
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

	count, err := s.db.Q.IsJobNameAvailable(ctx, db_queries.IsJobNameAvailableParams{
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

	count, err := s.db.Q.IsConnectionInAccount(ctx, db_queries.IsConnectionInAccountParams{
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
	logger *slog.Logger,
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
			logger.Error(fmt.Errorf("unable to retrieve workflow executions: %w", err).Error())
			return nil, err
		}

		executions = append(executions, resp.Executions...)
		nextPageToken = resp.NextPageToken
	}

	return executions, nil
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
