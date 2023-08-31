package v1alpha1_jobservice

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/types"
)

func (s *Service) GetJobs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobsResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	jobs, err := s.db.Q.GetJobsByAccount(ctx, *accountUuid)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	jobIds := []pgtype.UUID{}
	for _, job := range jobs {
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
	for _, job := range jobs {
		job := job // This is necessary otherwise the same job gets set in the map as the memory pointer is re-used
		jobMap[job.ID] = &job
	}

	associationMap := map[pgtype.UUID][]pgtype.UUID{}
	for _, assoc := range destinationAssociations {
		if _, ok := associationMap[assoc.JobID]; ok {
			associationMap[assoc.JobID] = append(associationMap[assoc.JobID], assoc.ConnectionID)
		} else {
			associationMap[assoc.JobID] = append([]pgtype.UUID{}, assoc.ConnectionID)
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
	var destConnections []pgtype.UUID
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

// func (s *Service) CreateJob(
// 	ctx context.Context,
// 	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
// ) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
// 	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	userUuid, err := s.getUserUuid(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

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
// 		err = cron.Scan(req.Msg.GetCronSchedule())
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
// 	if err = s.db.WithTx(ctx, nil, func(q *db_queries.Queries) error {
// 		job, err := q.CreateJob(ctx, db_queries.CreateJobParams{
// 			Name:                    req.Msg.JobName,
// 			AccountID:               *accountUuid,
// 			Status:                  int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
// 			CronSchedule:            cron,
// 			HaltOnNewColumnAddition: nucleusdb.BoolToInt16(req.Msg.HaltOnNewColumnAddition),
// 			CreatedByID:             *userUuid,
// 			UpdatedByID:             *userUuid,
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
// 		Job: dtomaps.ToJobDto(createdJob, destUuids),
// 	}), nil
// }

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

func (s *Service) CreateJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobResponse], error) {
	// TODO use go routines

	// check connections exist

	schemas := createSqlSchemas(req.Msg.Mappings)

	// create job config
	trueBool := true
	job := &neosyncdevv1alpha1.JobConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.k8sclient.Namespace,
			Name:      req.Msg.JobName,
		},
		Spec: neosyncdevv1alpha1.JobConfigSpec{
			Source: &neosyncdevv1alpha1.JobConfigSource{
				Sql: &neosyncdevv1alpha1.SourceSql{
					ConnectionRef: neosyncdevv1alpha1.LocalResourceRef{
						Name: "",
					},
					HaltOnSchemaChange: &req.Msg.HaltOnNewColumnAddition,
					Schemas:            schemas,
				},
			},
			Destinations: []*neosyncdevv1alpha1.JobConfigDestination{
				{
					Sql: &neosyncdevv1alpha1.DestinationSql{
						ConnectionRef: &neosyncdevv1alpha1.LocalResourceRef{
							Name: "destinationConnection.Name",
						},
						TruncateBeforeInsert: &trueBool, //TODO
						InitDbSchema:         &trueBool, //TODO
					},
				},
			},
		},
	}

	err := s.k8sclient.CustomResourceClient.Create(ctx, job)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateJobResponse{
		// Job: dtomaps.ToJobDto(createdJob, destUuids),
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

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(job.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveJobById(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
}

func (s *Service) verifyUserInAccount(
	ctx context.Context,
	accountId string,
) (*pgtype.UUID, error) {
	resp, err := s.userAccountService.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return nil, err
	}
	if !resp.Msg.Ok {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}

	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}
	return &accountUuid, nil
}

func (s *Service) getUserUuid(
	ctx context.Context,
) (*pgtype.UUID, error) {
	user, err := s.userAccountService.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userUuid, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}
	return &userUuid, nil
}
