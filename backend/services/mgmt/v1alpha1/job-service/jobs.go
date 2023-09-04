package v1alpha1_jobservice

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	neosync_k8sclient "github.com/nucleuscloud/neosync/backend/internal/k8s/client"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Service) GetJobs(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	jobs := &neosyncdevv1alpha1.JobConfigList{}
	err := s.k8sclient.CustomResourceClient.List(ctx, jobs, runtimeclient.InNamespace(s.cfg.JobConfigNamespace))
	if err != nil && !errors.IsNotFound(err) {
		logger.Error("unable to retrieve jobs")
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetJobsResponse{
			Jobs: []*mgmtv1alpha1.Job{},
		}), nil
	}
	if len(jobs.Items) == 0 {
		return connect.NewResponse(&mgmtv1alpha1.GetJobsResponse{
			Jobs: []*mgmtv1alpha1.Job{},
		}), nil
	}

	connections, err := s.connectionService.GetConnections(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionsRequest{}))
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	connsNameToIdMap := map[string]string{}
	for _, conn := range connections.Msg.Connections {
		connsNameToIdMap[conn.Name] = conn.Id
	}

	dtoJobs := []*mgmtv1alpha1.Job{}
	for i := range jobs.Items {
		job := jobs.Items[i]
		sourceConnName, err := getSourceConnectionName(job.Spec.Source)
		if err != nil {
			return nil, err
		}
		sourceConnId := connsNameToIdMap[sourceConnName]
		destConnIds := []string{}
		for _, dest := range job.Spec.Destinations {
			destConnName, err := getDestinationConnectionName(dest)
			if err != nil {
				return nil, err
			}
			destConnIds = append(destConnIds, connsNameToIdMap[destConnName])
		}

		dto := dtomaps.ToJobDto(&job, sourceConnId, destConnIds)
		if err != nil {
			return nil, err
		}
		dtoJobs = append(dtoJobs, dto)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobsResponse{
		Jobs: dtoJobs,
	}), nil
}

func (s *Service) GetJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRequest],
) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	job, err := getJobById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}
	destConnNames := []string{}
	for _, dest := range job.Spec.Destinations {
		destConnName, err := getDestinationConnectionName(dest)
		if err != nil {
			return nil, err
		}
		destConnNames = append(destConnNames, destConnName)
	}
	sourceConnName, err := getSourceConnectionName(job.Spec.Source)
	if err != nil {
		return nil, err // TODO @alisha should return job even without connections
	}
	connNames := []string{sourceConnName}
	connNames = append(connNames, destConnNames...)

	connections, err := s.connectionService.GetConnectionsByNames(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionsByNamesRequest{
		Names: connNames,
	}))
	if err != nil {
		return nil, err
	}

	destConnIds := []string{}
	var sourceConnId string
	for _, conn := range connections.Msg.Connections {
		if conn.Name == sourceConnName {
			sourceConnId = conn.Id
		} else if slices.Contains(destConnNames, conn.Name) {
			destConnIds = append(destConnIds, conn.Id)
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
		Job: dtomaps.ToJobDto(job, sourceConnId, destConnIds),
	}), nil
}

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
		index := index
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

	trueBool := true // TODO @alisha
	jobDestinations := []*neosyncdevv1alpha1.JobConfigDestination{}
	for _, name := range destConnNames {
		jobDestinations = append(jobDestinations, &neosyncdevv1alpha1.JobConfigDestination{
			Sql: &neosyncdevv1alpha1.DestinationSql{
				ConnectionRef: &neosyncdevv1alpha1.LocalResourceRef{
					Name: name,
				},
				TruncateBeforeInsert: &trueBool, // TODO @alisha
				InitDbSchema:         &trueBool, // TODO @alisha
			},
		})
	}

	schemas := createSqlSchemas(req.Msg.Mappings)
	job := &neosyncdevv1alpha1.JobConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cfg.JobConfigNamespace,
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

func (s *Service) DeleteJob(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.Msg.Id)
	logger.Info("deleting job config", "jobId", req.Msg.Id)
	job, err := getJobById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
	}

	err = s.k8sclient.CustomResourceClient.Delete(ctx, job, &runtimeclient.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteJobResponse{}), nil
}

func (s *Service) UpdateJobSchedule(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobScheduleRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobScheduleResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobScheduleResponse{}), nil
}

func (s *Service) UpdateJobSourceConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobSourceConnectionResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobSourceConnectionResponse{}), nil
}

func (s *Service) UpdateJobDestinationConnections(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobDestinationConnectionsRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobDestinationConnectionsResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobDestinationConnectionsResponse{}), nil
}

func (s *Service) UpdateJobMappings(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobMappingsRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobMappingsResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobMappingsResponse{}), nil
}

func (s *Service) UpdateJobHaltOnNewColumnAddition(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateJobHaltOnNewColumnAdditionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateJobHaltOnNewColumnAdditionResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.UpdateJobHaltOnNewColumnAdditionResponse{}), nil
}

func createSqlSchemas(mappings []*mgmtv1alpha1.JobMapping) []*neosyncdevv1alpha1.SourceSqlSchema {
	schema := []*neosyncdevv1alpha1.SourceSqlSchema{}
	schemaMap := map[string]map[string][]*neosyncdevv1alpha1.SourceSqlColumn{}
	for _, row := range mappings {
		row := row
		_, ok := schemaMap[row.Schema][row.Table]
		if !ok {
			schemaMap[row.Schema] = map[string][]*neosyncdevv1alpha1.SourceSqlColumn{
				row.Table: {
					{
						Name:    row.Column,
						Exclude: &row.Exclude,
						Transformer: &neosyncdevv1alpha1.ColumnTransformer{
							Name: row.Transformer.Enum().String(),
						},
					},
				},
			}
			continue
		}

		schemaMap[row.Schema][row.Table] = append(schemaMap[row.Schema][row.Table], &neosyncdevv1alpha1.SourceSqlColumn{
			Name:    row.Column,
			Exclude: &row.Exclude,
			Transformer: &neosyncdevv1alpha1.ColumnTransformer{
				Name: row.Transformer.Enum().String(),
			},
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

	return schema
}

func getSourceConnectionName(jobConfig *neosyncdevv1alpha1.JobConfigSource) (string, error) {
	if jobConfig.Sql != nil {
		return jobConfig.Sql.ConnectionRef.Name, nil
	}
	return "", nucleuserrors.NewBadRequest("this job source connection config is not currently supported")
}

func getDestinationConnectionName(jobConfig *neosyncdevv1alpha1.JobConfigDestination) (string, error) {
	if jobConfig.Sql != nil {
		return jobConfig.Sql.ConnectionRef.Name, nil
	}
	return "", nucleuserrors.NewBadRequest("this job destination connection config is not currently supported")
}

func getJobById(
	ctx context.Context,
	logger *slog.Logger,
	k8sclient *neosync_k8sclient.Client,
	id string,
	namespace string,
) (*neosyncdevv1alpha1.JobConfig, error) {
	jobs := &neosyncdevv1alpha1.JobConfigList{}
	err := k8sclient.CustomResourceClient.List(ctx, jobs, runtimeclient.InNamespace(namespace), &runtimeclient.MatchingLabels{
		k8s_utils.NeosyncUuidLabel: id,
	})
	if err != nil {
		logger.Error("unable to retrieve job")
		return nil, err
	}
	if len(jobs.Items) == 0 {
		return nil, nucleuserrors.NewNotFound(fmt.Sprintf("job config not found. id: %s", id))
	}
	if len(jobs.Items) > 1 {
		return nil, nucleuserrors.NewInternalError(fmt.Sprintf("more than 1 job config found. id: %s", id))
	}
	return &jobs.Items[0], nil
}
