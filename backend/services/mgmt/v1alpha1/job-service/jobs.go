package v1alpha1_jobservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
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
		sourceConnName := job.Spec.Source.Sql.ConnectionRef.Name
		sourceConnId := connsNameToIdMap[sourceConnName]
		destConnIds := []string{}
		for _, dest := range job.Spec.Destinations {
			destConnIds = append(destConnIds, connsNameToIdMap[dest.Sql.ConnectionRef.Name])
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
	jobs := &neosyncdevv1alpha1.JobConfigList{}
	err := s.k8sclient.CustomResourceClient.List(ctx, jobs, runtimeclient.InNamespace(s.cfg.JobConfigNamespace), &runtimeclient.MatchingLabels{
		k8s_utils.NeosyncUuidLabel: req.Msg.Id,
	})
	if err != nil {
		logger.Error("unable to retrieve job")
		return nil, err
	}
	if len(jobs.Items) == 0 {
		return nil, nucleuserrors.NewNotFound(fmt.Sprintf("connection not found. id: %s", req.Msg.Id))
	}
	if len(jobs.Items) > 1 {
		return nil, nucleuserrors.NewInternalError(fmt.Sprintf("more than 1 connection found. id: %s", req.Msg.Id))
	}

	job := jobs.Items[0]
	destConnNames := []string{}
	for _, dest := range job.Spec.Destinations {
		destConnNames = append(destConnNames, dest.Sql.ConnectionRef.Name)
	}
	connNames := []string{job.Spec.Source.Sql.ConnectionRef.Name}
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
		if conn.Name == job.Spec.Source.Sql.ConnectionRef.Name {
			sourceConnId = conn.Id
		} else if slices.Contains(destConnNames, conn.Name) {
			destConnIds = append(destConnIds, conn.Id)
		}
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
		Job: dtomaps.ToJobDto(&job, sourceConnId, destConnIds),
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
