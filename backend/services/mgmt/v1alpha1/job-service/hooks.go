package v1alpha1_jobservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func (s *Service) GetJobHooks(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobHooksRequest],
) (*connect.Response[mgmtv1alpha1.GetJobHooksResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetJobHooksResponse{}), nil
}

func (s *Service) GetJobHook(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobHookRequest],
) (*connect.Response[mgmtv1alpha1.GetJobHookResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetJobHookResponse{}), nil
}

func (s *Service) CreateJobHook(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobHookRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobHookResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.CreateJobHookResponse{}), nil
}

func (s *Service) DeleteJobHook(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobHookRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobHookResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.DeleteJobHookResponse{}), nil
}

func (s *Service) IsJobHookNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsJobHookNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsJobHookNameAvailableResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.IsJobHookNameAvailableResponse{}), nil
}
