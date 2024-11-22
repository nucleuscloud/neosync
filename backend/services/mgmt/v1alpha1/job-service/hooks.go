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
	resp, err := s.hookService.GetJobHooks(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) GetJobHook(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetJobHookRequest],
) (*connect.Response[mgmtv1alpha1.GetJobHookResponse], error) {
	resp, err := s.hookService.GetJobHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) CreateJobHook(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateJobHookRequest],
) (*connect.Response[mgmtv1alpha1.CreateJobHookResponse], error) {
	resp, err := s.hookService.CreateJobHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) DeleteJobHook(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteJobHookRequest],
) (*connect.Response[mgmtv1alpha1.DeleteJobHookResponse], error) {
	resp, err := s.hookService.DeleteJobHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) IsJobHookNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsJobHookNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsJobHookNameAvailableResponse], error) {
	resp, err := s.hookService.IsJobHookNameAvailable(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}
