package v1alpha1_jobservice

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	neosync_k8sclient "github.com/nucleuscloud/neosync/backend/internal/k8s/client"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Service) GetJobRuns(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunsRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	jobRuns := &neosyncdevv1alpha1.JobRunList{}
	err := s.k8sclient.CustomResourceClient.List(ctx, jobRuns, runtimeclient.InNamespace(s.cfg.JobConfigNamespace))
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(fmt.Errorf("unable to retrieve job runs: %w", err).Error())
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
			JobRuns: []*mgmtv1alpha1.JobRun{},
		}), nil
	}
	if len(jobRuns.Items) == 0 {
		return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
			JobRuns: []*mgmtv1alpha1.JobRun{},
		}), nil
	}

	jobs, err := s.GetJobs(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobsRequest{}))
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve jobs: %w", err).Error())
		return nil, err
	}

	jobNameIdMap := map[string]string{}
	for _, job := range jobs.Msg.GetJobs() {
		jobNameIdMap[job.Name] = job.Id
	}

	dtoJobRuns := []*mgmtv1alpha1.JobRun{}
	for i := range jobRuns.Items {
		run := jobRuns.Items[i]
		jobId := jobNameIdMap[run.Spec.Job.JobRef.Name]
		dtoJobRuns = append(dtoJobRuns, dtomaps.ToJobRunDto(&run, &jobId))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
		JobRuns: dtoJobRuns,
	}), nil
}

func (s *Service) GetJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.Id)
	jobRun, err := getJobRunById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}

	job, err := getJobByName(ctx, logger, s.k8sclient, jobRun.Spec.Job.JobRef.Name, s.cfg.JobConfigNamespace)
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job config: %w", err).Error(), "jobName", jobRun.Spec.Job.JobRef.Name)
		return nil, err
	}

	jobId := job.Labels[k8s_utils.NeosyncUuidLabel]
	dto := dtomaps.ToJobRunDto(jobRun, &jobId)
	return connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{
		JobRun: dto,
	}), nil
}

func (s *Service) CreateJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobRunResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.CreateJobRunResponse{}), nil
}

func (s *Service) CancelJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CancelJobRunRequest],
) (*connect.Response[mgmtv1alpha1.CancelJobRunResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.CancelJobRunResponse{}), nil
}

func (s *Service) DeleteJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobRunRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobRunResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobRunId", req.Msg.Id)
	jobRun, err := getJobRunById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(fmt.Errorf("unable to retrieve job runs: %w", err).Error())
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		logger.Info("job run not found")
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobRunResponse{}), nil
	}

	logger.Info("deleting job run")
	err = s.k8sclient.CustomResourceClient.Delete(ctx, jobRun, &runtimeclient.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteJobRunResponse{}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobRunResponse{}), nil
}

func getJobRunById(
	ctx context.Context,
	logger *slog.Logger,
	k8sclient *neosync_k8sclient.Client,
	id string,
	namespace string,
) (*neosyncdevv1alpha1.JobRun, error) {
	jobRuns := &neosyncdevv1alpha1.JobRunList{}
	err := k8sclient.CustomResourceClient.List(ctx, jobRuns, runtimeclient.InNamespace(namespace), &runtimeclient.MatchingLabels{
		k8s_utils.NeosyncUuidLabel: id,
	})
	if err != nil {
		logger.Error(fmt.Errorf("unable to retrieve job run: %w", err).Error())
		return nil, err
	}
	if len(jobRuns.Items) == 0 {
		return nil, nucleuserrors.NewNotFound(fmt.Sprintf("job run not found. id: %s", id))
	}
	if len(jobRuns.Items) > 1 {
		return nil, nucleuserrors.NewInternalError(fmt.Sprintf("more than 1 job run found. id: %s", id))
	}
	return &jobRuns.Items[0], nil
}
