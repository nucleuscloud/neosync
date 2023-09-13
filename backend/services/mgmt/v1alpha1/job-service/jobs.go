package v1alpha1_jobservice

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
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
	for _, assoc := range destinationAssociations {
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
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
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
			Name: req.Msg.JobName,
			// AccountID:          *accountId,
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

type cronScheduleSpec struct {
	CronSchedule *string `json:"cronSchedule"`
}

type updateJobScheduleSpec struct {
	Spec *cronScheduleSpec `json:"spec,omitempty"`
}

func (s *Service) UpdateJobSchedule(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobScheduleRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobScheduleResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job schedule")
	job, err := getJobById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}

	var schedule *string
	if req.Msg.CronSchedule != nil && *req.Msg.CronSchedule != "" {
		schedule = req.Msg.CronSchedule
	}
	patch := &updateJobScheduleSpec{
		Spec: &cronScheduleSpec{
			CronSchedule: schedule,
		},
	}
	patchBits, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	err = s.k8sclient.CustomResourceClient.Patch(ctx, job, runtimeclient.RawPatch(types.MergePatchType, patchBits))
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

type patchUpdateJobConfigSpec struct {
	Spec *jobConfigSpec `json:"spec"`
}

type jobConfigSpec struct {
	Source       *jobSourceConnection        `json:"source,omitempty"`
	Destinations []*jobDestinationConnection `json:"destinations,omitempty"`
	CronSchedule *string                     `json:"cronSchedule,omitempty"`
}

type jobDestinationConnection struct {
	Sql *sqlDestinationConnection `json:"sql,omitempty"`
}

type jobSourceConnection struct {
	Sql *sqlSourceConnection `json:"sql,omitempty"`
}

type sqlSourceConnection struct {
	ConnectionRef      *connectionRef                        `json:"connectionRef,omitempty"`
	HaltOnSchemaChange *bool                                 `json:"haltOnSchemaChange,omitempty"`
	Schemas            []*neosyncdevv1alpha1.SourceSqlSchema `json:"schemas,omitempty"`
}

type sqlDestinationConnection struct {
	ConnectionRef        *connectionRef `json:"connectionRef,omitempty"`
	TruncateBeforeInsert *bool          `json:"truncateBeforeInsert,omitempty"`
	InitDbSchema         *bool          `json:"initDbSchema,omitempty"`
}

type connectionRef struct {
	Name string `json:"name"`
}

func (s *Service) UpdateJobSourceConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobSourceConnectionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job source connection")
	job, err := getJobById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}

	conn, err := s.connectionService.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.Source.ConnectionId,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve source connection: %w", err).Error())
		return nil, err
	}

	var source *jobSourceConnection
	switch conn.Msg.Connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		source = &jobSourceConnection{
			Sql: &sqlSourceConnection{
				ConnectionRef: &connectionRef{
					Name: conn.Msg.Connection.Name,
				},
				HaltOnSchemaChange: req.Msg.Source.Options.GetSqlOptions().HaltOnNewColumnAddition,
			},
		}
	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}

	patch := &patchUpdateJobConfigSpec{
		Spec: &jobConfigSpec{
			Source: source,
		},
	}
	patchBits, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	err = s.k8sclient.CustomResourceClient.Patch(ctx, job, runtimeclient.RawPatch(types.MergePatchType, patchBits))
	if err != nil {
		logger.Error(fmt.Errorf("unable to update job source connection: %w", err).Error())
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

func (s *Service) UpdateJobDestinationConnections(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobDestinationConnectionsRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobDestinationConnectionsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("updating job destination connections")

	var job *neosyncdevv1alpha1.JobConfig
	destConns := make([]*mgmtv1alpha1.Connection, len(req.Msg.Destinations))
	errs, errCtx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		jobConfig, err := getJobById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
		if err != nil {
			return err
		}
		job = jobConfig
		return nil
	})

	destMap := map[string]*mgmtv1alpha1.JobDestination{}
	for index, dest := range req.Msg.Destinations {
		dest := dest
		index := index
		destMap[dest.ConnectionId] = dest
		errs.Go(func() error {
			conn, err := s.connectionService.GetConnection(errCtx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
				Id: dest.ConnectionId,
			}))
			if err != nil {
				return err
			}
			destConns[index] = conn.Msg.Connection
			return nil
		})
	}
	err := errs.Wait()
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job connections: %w", err).Error())
		return nil, err
	}

	jobDestinations := []*jobDestinationConnection{}
	for _, conn := range destConns {
		destOptions := destMap[conn.Id]
		switch conn.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
			jobDestinations = append(jobDestinations, &jobDestinationConnection{
				Sql: &sqlDestinationConnection{
					ConnectionRef: &connectionRef{
						Name: conn.Name,
					},
					TruncateBeforeInsert: destOptions.Options.GetSqlOptions().TruncateBeforeInsert,
					InitDbSchema:         destOptions.Options.GetSqlOptions().InitDbSchema,
				},
			})
		default:
			return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
		}
	}

	patch := &patchUpdateJobConfigSpec{
		Spec: &jobConfigSpec{
			Destinations: jobDestinations,
		},
	}
	patchBits, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	err = s.k8sclient.CustomResourceClient.Patch(ctx, job, runtimeclient.RawPatch(types.MergePatchType, patchBits))
	if err != nil {
		logger.Error(fmt.Errorf("unable to update job destination connections: %w", err).Error())
		return nil, err
	}

	updatedJob, err := s.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job: %w", err).Error())
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.UpdateJobDestinationConnectionsResponse{
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
	job, err := getJobById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}

	var patch *patchUpdateJobConfigSpec
	if job.Spec.Source.Sql != nil {
		schemas, err := createSqlSchemas(req.Msg.Mappings)
		if err != nil {
			return nil, fmt.Errorf("unable to generate SQL job mapping: %w", err)
		}
		patch = &patchUpdateJobConfigSpec{
			Spec: &jobConfigSpec{
				Source: &jobSourceConnection{
					Sql: &sqlSourceConnection{
						Schemas: schemas,
					},
				},
			},
		}
	} else {
		return nil, nucleuserrors.NewBadRequest("this job config is not currently supported")
	}
	patchBits, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	err = s.k8sclient.CustomResourceClient.Patch(ctx, job, runtimeclient.RawPatch(types.MergePatchType, patchBits))
	if err != nil {
		logger.Error(fmt.Errorf("unable to update job destination connections: %w", err).Error())
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
	job := &neosyncdevv1alpha1.JobConfig{}
	err := s.k8sclient.CustomResourceClient.Get(ctx, types.NamespacedName{Name: req.Msg.Name, Namespace: s.cfg.JobConfigNamespace}, job)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.IsJobNameAvailableResponse{
			IsAvailable: true,
		}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.IsJobNameAvailableResponse{
		IsAvailable: false,
	}), nil
}

func createSqlSchemas(mappings []*mgmtv1alpha1.JobMapping) ([]*neosyncdevv1alpha1.SourceSqlSchema, error) {
	schema := []*neosyncdevv1alpha1.SourceSqlSchema{}
	schemaMap := map[string]map[string][]*neosyncdevv1alpha1.SourceSqlColumn{}
	for _, row := range mappings {
		row := row
		transformer, err := getColumnTransformer(row.Transformer)
		if err != nil {
			return nil, err
		}
		_, ok := schemaMap[row.Schema][row.Table]
		if !ok {
			schemaMap[row.Schema] = map[string][]*neosyncdevv1alpha1.SourceSqlColumn{
				row.Table: {
					{
						Name:        row.Column,
						Exclude:     &row.Exclude,
						Transformer: transformer,
					},
				},
			}
			continue
		}

		schemaMap[row.Schema][row.Table] = append(schemaMap[row.Schema][row.Table], &neosyncdevv1alpha1.SourceSqlColumn{
			Name:        row.Column,
			Exclude:     &row.Exclude,
			Transformer: transformer,
		})

	}

	for s, table := range schemaMap {
		for t, columns := range table {
			schema = append(schema, &neosyncdevv1alpha1.SourceSqlSchema{
				Schema:  s,
				Table:   t,
				Columns: columns,
			})
		}
	}

	return schema, nil
}
