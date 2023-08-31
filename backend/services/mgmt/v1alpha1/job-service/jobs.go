package v1alpha1_jobservice

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Service) GetJobs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobsResponse], error) {
	// logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	// accountUuid, err := nucleusdb.ToUuid(req.Msg.AccountId)
	// if err != nil {
	// 	return nil, err
	// }
	// jobs, err := s.db.Q.GetJobsByAccount(ctx, accountUuid)
	// if err != nil {
	// 	logger.Error(err.Error())
	// 	return nil, err
	// }

	// jobIds := []pgtype.UUID{}
	// for idx := range jobs {
	// 	job := jobs[idx]
	// 	jobIds = append(jobIds, job.ID)
	// }

	// var destinationAssociations []db_queries.NeosyncApiJobDestinationConnectionAssociation
	// if len(jobIds) > 0 {
	// 	destinationAssociations, err = s.db.Q.GetJobConnectionDestinationsByJobIds(ctx, jobIds)
	// 	if err != nil {
	// 		logger.Error(err.Error())
	// 		return nil, err
	// 	}
	// }

	// jobMap := map[pgtype.UUID]*db_queries.NeosyncApiJob{}
	// for idx := range jobs {
	// 	job := jobs[idx]
	// 	jobMap[job.ID] = &job
	// }

	// associationMap := map[pgtype.UUID][]pgtype.UUID{}
	// for _, assoc := range destinationAssociations {
	// 	if _, ok := associationMap[assoc.JobID]; ok {
	// 		associationMap[assoc.JobID] = append(associationMap[assoc.JobID], assoc.ConnectionID)
	// 	} else {
	// 		associationMap[assoc.JobID] = append([]pgtype.UUID{}, assoc.ConnectionID)
	// 	}
	// }

	// dtos := []*mgmtv1alpha1.Job{}

	// Use jobIds to retain original query order
	// for _, jobId := range jobIds {
	// 	job := jobMap[jobId]
	// 	dtos = append(dtos, dtomaps.ToJobDto(job, associationMap[job.ID]))
	// }

	return connect.NewResponse(&mgmtv1alpha1.GetJobsResponse{
		// Jobs: dtos,
	}), nil
}

func (s *Service) GetJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRequest],
) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
	// jobUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	// if err != nil {
	// 	return nil, err
	// }

	// errgrp, errctx := errgroup.WithContext(ctx)

	// var job db_queries.NeosyncApiJob
	// errgrp.Go(func() error {
	// 	j, err := s.db.Q.GetJobById(errctx, jobUuid)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	job = j
	// 	return nil
	// })
	// var destConnections []pgtype.UUID
	// errgrp.Go(func() error {
	// 	dcs, err := s.db.Q.GetJobConnectionDestinations(ctx, jobUuid)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	destConnections = dcs
	// 	return nil
	// })
	// if err = errgrp.Wait(); err != nil {
	// 	if nucleusdb.IsNoRows(err) {
	// 		return nil, nucleuserrors.NewNotFound("unable to find job by id")
	// 	}
	// 	return nil, err
	// }

	return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
		// Job: dtomaps.ToJobDto(&job, destConnections),
	}), nil
}

// func (s *Service) CreateJobOld(
// 	ctx context.Context,
// 	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
// ) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
// 	destUuids := []pgtype.UUID{}
// 	for _, destId := range req.Msg.ConnectionDestinationIds {
// 		destUuid, err := nucleusdb.ToUuid(destId)
// 		if err != nil {
// 			return nil, err
// 		}
// 		destUuids = append(destUuids, destUuid)
// 	}

// 	cron := pgtype.Text{}
// 	if req.Msg.CronSchedule != nil {
// 		err := cron.Scan(req.Msg.GetCronSchedule())
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	connectionSourceUuid, err := nucleusdb.ToUuid(req.Msg.ConnectionSourceId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	mappings := []*jsonmodels.JobMapping{}
// 	for _, mapping := range req.Msg.Mappings {
// 		jm := &jsonmodels.JobMapping{}
// 		err = jm.FromDto(mapping)
// 		if err != nil {
// 			return nil, err
// 		}
// 		mappings = append(mappings, jm)
// 	}

// 	var createdJob *db_queries.NeosyncApiJob
// 	if err := s.db.WithTx(ctx, nil, func(q *db_queries.Queries) error {
// 		job, err := q.CreateJob(ctx, db_queries.CreateJobParams{
// 			Name: req.Msg.JobName,
// 			// AccountID:               *accountUuid,
// 			Status:                  int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
// 			CronSchedule:            cron,
// 			HaltOnNewColumnAddition: nucleusdb.BoolToInt16(req.Msg.HaltOnNewColumnAddition),
// 			ConnectionSourceID:      connectionSourceUuid,
// 			Mappings:                mappings,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		connDestParams := []db_queries.CreateJobConnectionDestinationsParams{}
// 		for _, destUuid := range destUuids {
// 			connDestParams = append(connDestParams, db_queries.CreateJobConnectionDestinationsParams{
// 				JobID:        job.ID,
// 				ConnectionID: destUuid,
// 			})
// 		}
// 		if len(connDestParams) > 0 {
// 			_, err = q.CreateJobConnectionDestinations(ctx, connDestParams)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		createdJob = &job
// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}

// 	return connect.NewResponse(&mgmtv1alpha1.CreateJobResponse{
// 		// Job: dtomaps.ToJobDto(createdJob, destUuids),
// 	}), nil
// }

func (s *Service) CreateJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobName", req.Msg.JobName)
	logger.Info("creating job")
	jobUuid := uuid.NewString()

	var sourceConnName *string
	destConnNames := make([]string, len(req.Msg.ConnectionDestinationIds))

	errs, errCtx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		conn, err := s.connectionService.GetConnection(errCtx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: req.Msg.ConnectionSourceId,
		}))
		if err != nil {
			return nil
		}
		sourceConnName = &conn.Msg.Connection.Name
		return nil
	})

	for index, id := range req.Msg.ConnectionDestinationIds {
		connId := id
		errs.Go(func() error {
			conn, err := s.connectionService.GetConnection(errCtx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
				Id: connId,
			}))
			if err != nil {
				return err
			}
			destConnNames[index] = conn.Msg.Connection.Name
			return nil
		})
	}
	err := errs.Wait()
	if err != nil {
		logger.Error("unable to retrieve job connections")
		return nil, err
	}

	trueBool := true // TODO
	jobDestinations := []*neosyncdevv1alpha1.JobConfigDestination{}
	for _, name := range destConnNames {
		jobDestinations = append(jobDestinations, &neosyncdevv1alpha1.JobConfigDestination{
			Sql: &neosyncdevv1alpha1.DestinationSql{
				ConnectionRef: &neosyncdevv1alpha1.LocalResourceRef{
					Name: name,
				},
				TruncateBeforeInsert: &trueBool, // TODO
				InitDbSchema:         &trueBool, // TODO
			},
		})
	}

	schemas := createSqlSchemas(req.Msg.Mappings)
	job := &neosyncdevv1alpha1.JobConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.k8sclient.Namespace,
			Name:      req.Msg.JobName,
			Labels: map[string]string{
				k8s_utils.NeosyncUuidLabel: jobUuid,
			},
		},
		Spec: neosyncdevv1alpha1.JobConfigSpec{
			Source: &neosyncdevv1alpha1.JobConfigSource{
				Sql: &neosyncdevv1alpha1.SourceSql{
					ConnectionRef: neosyncdevv1alpha1.LocalResourceRef{
						Name: *sourceConnName,
					},
					HaltOnSchemaChange: &req.Msg.HaltOnNewColumnAddition,
					Schemas:            schemas,
				},
			},
			Destinations: jobDestinations,
		},
	}

	err = s.k8sclient.CustomResourceClient.Create(ctx, job)
	if err != nil {
		logger.Error("unable to create job")
		return nil, err
	}
	logger.Info("created job", "jobId", jobUuid)

	return connect.NewResponse(&mgmtv1alpha1.CreateJobResponse{
		Job: dtomaps.ToJobDto(job, req.Msg.ConnectionSourceId, req.Msg.ConnectionDestinationIds),
	}), nil
}

func (s *Service) UpdateJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobResponse{}), nil
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

func createSqlSchemas(mappings []*mgmtv1alpha1.JobMapping) []*neosyncdevv1alpha1.SourceSqlSchema {
	schema := []*neosyncdevv1alpha1.SourceSqlSchema{}

	schemaMap := map[string]map[string][]*neosyncdevv1alpha1.SourceSqlColumn{}
	for _, row := range mappings {

		_, ok := schemaMap[row.Schema][row.Table]
		if !ok {
			schemaMap[row.Schema] = map[string][]*neosyncdevv1alpha1.SourceSqlColumn{
				row.Table: {
					{
						Name:    row.Column,
						Exclude: &row.Exclude,
						Transformer: &neosyncdevv1alpha1.ColumnTransformer{
							Name: row.Transformer.String(),
						},
					},
				},
			}
			break
		}

		schemaMap[row.Schema][row.Table] = append(schemaMap[row.Schema][row.Table], &neosyncdevv1alpha1.SourceSqlColumn{
			Name:    row.Column,
			Exclude: &row.Exclude,
			Transformer: &neosyncdevv1alpha1.ColumnTransformer{
				Name: row.Transformer.String(),
			},
		})

	}

	return schema
}
