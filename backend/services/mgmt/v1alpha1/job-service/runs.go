package v1alpha1_jobservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		dtoJobRuns = append(dtoJobRuns, &mgmtv1alpha1.JobRun{
			Id:        run.Labels[k8s_utils.NeosyncUuidLabel],
			JobId:     jobNameIdMap[run.Spec.Job.JobRef.Name],
			Name:      run.Name,
			Status:    mgmtv1alpha1.JobRunStatus(0), // TODO @alisha implement
			CreatedAt: timestamppb.New(run.CreationTimestamp.Time),
		})
	}

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunsResponse{
		JobRuns: dtoJobRuns,
	}), nil
}

func (s *Service) GetJobRun(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobRunRequest],
) (*connect.Response[mgmtv1alpha1.GetJobRunResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{}), nil
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
