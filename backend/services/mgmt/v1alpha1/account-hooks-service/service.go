package v1alpha1_accounthookservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	accounthooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/accounts"
)

type Service struct {
	hookservice accounthooks.Interface
}

var _ mgmtv1alpha1connect.AccountHookServiceHandler = (*Service)(nil)

func New(
	hookservice accounthooks.Interface,
) *Service {
	return &Service{
		hookservice: hookservice,
	}
}

func (s *Service) GetAccountHooks(ctx context.Context, req *connect.Request[mgmtv1alpha1.GetAccountHooksRequest]) (*connect.Response[mgmtv1alpha1.GetAccountHooksResponse], error) {
	resp, err := s.hookservice.GetAccountHooks(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) GetAccountHook(ctx context.Context, req *connect.Request[mgmtv1alpha1.GetAccountHookRequest]) (*connect.Response[mgmtv1alpha1.GetAccountHookResponse], error) {
	resp, err := s.hookservice.GetAccountHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) CreateAccountHook(ctx context.Context, req *connect.Request[mgmtv1alpha1.CreateAccountHookRequest]) (*connect.Response[mgmtv1alpha1.CreateAccountHookResponse], error) {
	resp, err := s.hookservice.CreateAccountHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) DeleteAccountHook(ctx context.Context, req *connect.Request[mgmtv1alpha1.DeleteAccountHookRequest]) (*connect.Response[mgmtv1alpha1.DeleteAccountHookResponse], error) {
	resp, err := s.hookservice.DeleteAccountHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) IsAccountHookNameAvailable(ctx context.Context, req *connect.Request[mgmtv1alpha1.IsAccountHookNameAvailableRequest]) (*connect.Response[mgmtv1alpha1.IsAccountHookNameAvailableResponse], error) {
	resp, err := s.hookservice.IsAccountHookNameAvailable(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) UpdateAccountHook(ctx context.Context, req *connect.Request[mgmtv1alpha1.UpdateAccountHookRequest]) (*connect.Response[mgmtv1alpha1.UpdateAccountHookResponse], error) {
	resp, err := s.hookservice.UpdateAccountHook(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) SetAccountHookEnabled(ctx context.Context, req *connect.Request[mgmtv1alpha1.SetAccountHookEnabledRequest]) (*connect.Response[mgmtv1alpha1.SetAccountHookEnabledResponse], error) {
	resp, err := s.hookservice.SetAccountHookEnabled(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) GetActiveAccountHooksByEvent(ctx context.Context, req *connect.Request[mgmtv1alpha1.GetActiveAccountHooksByEventRequest]) (*connect.Response[mgmtv1alpha1.GetActiveAccountHooksByEventResponse], error) {
	resp, err := s.hookservice.GetActiveAccountHooksByEvent(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) GetSlackConnectionUrl(ctx context.Context, req *connect.Request[mgmtv1alpha1.GetSlackConnectionUrlRequest]) (*connect.Response[mgmtv1alpha1.GetSlackConnectionUrlResponse], error) {
	resp, err := s.hookservice.GetSlackConnectionUrl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) HandleSlackOAuthCallback(ctx context.Context, req *connect.Request[mgmtv1alpha1.HandleSlackOAuthCallbackRequest]) (*connect.Response[mgmtv1alpha1.HandleSlackOAuthCallbackResponse], error) {
	resp, err := s.hookservice.HandleSlackOAuthCallback(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) TestSlackConnection(ctx context.Context, req *connect.Request[mgmtv1alpha1.TestSlackConnectionRequest]) (*connect.Response[mgmtv1alpha1.TestSlackConnectionResponse], error) {
	resp, err := s.hookservice.TestSlackConnection(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}
