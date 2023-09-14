package v1alpha1_jobservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
	"github.com/nucleuscloud/neosync/worker/internal/workflos/datasync"
	temporalclient "go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetJobs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	accountUuid, err := nucleusdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	jobs, err := s.db.Q.GetJobsByAccount(ctx, accountUuid)
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

	accountUuid, err := nucleusdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
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

	var createdJob *db_queries.NeosyncApiJob
	if err := s.db.WithTx(ctx, nil, func(q *db_queries.Queries) error {
		job, err := q.CreateJob(ctx, db_queries.CreateJobParams{
			Name:               req.Msg.JobName,
			AccountID:          accountUuid,
			Status:             int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
			CronSchedule:       cron,
			ConnectionSourceID: connectionSourceUuid,
			ConnectionOptions:  connectionOptions,
			Mappings:           mappings,
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
		return nil
	}); err != nil {
		return nil, err
	}

	schedule := nucleusdb.ToNullableString(createdJob.CronSchedule)
	if schedule != nil && *schedule != "" {
		// create scheduled

	} else {
		// create one off

		workflowUuid := uuid.New().String()
		jobUuid := nucleusdb.UUIDString(createdJob.ID)
		workflowOptions := temporalclient.StartWorkflowOptions{
			ID:        workflowUuid,
			TaskQueue: s.cfg.TemporalTaskQueue,
			SearchAttributes: map[string]interface{}{ // optional search attributes when start workflow
				"JobId": jobUuid,
			},
		}

		we, err := s.temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, datasync.Workflow, &datasync.WorkflowRequest{JobId: jobUuid})
		if err != nil {
			logger.Error(fmt.Errorf("unable to execute workflow: %w", err).Error())
			return nil, err
		}
		logger.Info("started workflow", "workflowId", we.GetID(), "runId", we.GetRunID())

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

	err = s.db.Q.RemoveJobById(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
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
		// UpdatedByID:      "", TODO @alisha
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
	logger.Info("updating job source connection")
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

	connectionUuid, err := nucleusdb.ToUuid(req.Msg.Source.ConnectionId)
	if err != nil {
		return nil, err
	}

	connectionOptions := &jsonmodels.JobSourceOptions{}
	err = connectionOptions.FromDto(req.Msg.Source.Options)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Q.UpdateJobSource(ctx, db_queries.UpdateJobSourceParams{
		ID:                 job.ID,
		ConnectionSourceID: connectionUuid,
		ConnectionOptions:  connectionOptions,

		// UpdatedByID:      "", TODO @alisha
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

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobSourceConnectionResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) UpdateJobDestinationConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobDestinationConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobDestinationConnectionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job destination connections")

	jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	job, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}

	jobDestOptionsMap := map[string]*mgmtv1alpha1.JobDestinationOptions{}
	for _, dest := range job.Msg.Job.Destinations {
		jobDestOptionsMap[dest.ConnectionId] = dest.Options
	}

	connectionUuid, err := nucleusdb.ToUuid(req.Msg.Destination.ConnectionId)
	if err != nil {
		return nil, err
	}
	options := &jsonmodels.JobDestinationOptions{}
	err = options.FromDto(req.Msg.Destination.Options)
	if err != nil {
		return nil, err
	}
	_, ok := jobDestOptionsMap[req.Msg.Destination.ConnectionId]
	if ok {
		_, err = s.db.Q.UpdateJobDestination(ctx, db_queries.UpdateJobDestinationParams{
			JobID:        jobUuid,
			ConnectionID: connectionUuid,
			Options:      options,
		})
		if err != nil {
			return nil, err
		}

	} else {
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
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.UpdateJobDestinationConnectionResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) UpdateJobMappings(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobMappingsRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobMappingsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job mappings")

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

	mappings := []*jsonmodels.JobMapping{}
	for _, mapping := range req.Msg.Mappings {
		jm := &jsonmodels.JobMapping{}
		err = jm.FromDto(mapping)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, jm)
	}

	_, err = s.db.Q.UpdateJobMappings(ctx, db_queries.UpdateJobMappingsParams{
		ID:       job.ID,
		Mappings: mappings,
		// UpdatedByID:      "", TODO @alisha
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

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobMappingsResponse{
		Job: updatedJob.Msg.Job,
	}), nil
}

func (s *Service) IsJobNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsJobNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsJobNameAvailableResponse], error) {
	accountUuid, err := nucleusdb.ToUuid(req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsJobNameAvailable(ctx, db_queries.IsJobNameAvailableParams{
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
