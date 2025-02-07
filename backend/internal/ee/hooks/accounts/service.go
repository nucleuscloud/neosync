package accounthooks

import (
	"context"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
)

type Service struct {
	cfg            *config
	db             *neosyncdb.NeosyncDb
	userdataclient userdata.Interface
}

var _ Interface = (*Service)(nil)

type Interface interface {
	GetAccountHooks(ctx context.Context, req *mgmtv1alpha1.GetAccountHooksRequest) (*mgmtv1alpha1.GetAccountHooksResponse, error)
	// GetAccountHook(ctx context.Context, req *mgmtv1alpha1.GetAccountHookRequest) (*mgmtv1alpha1.GetAccountHookResponse, error)
	// CreateAccountHook(ctx context.Context, req *mgmtv1alpha1.CreateAccountHookRequest) (*mgmtv1alpha1.CreateAccountHookResponse, error)
	// DeleteAccountHook(ctx context.Context, req *mgmtv1alpha1.DeleteAccountHookRequest) (*mgmtv1alpha1.DeleteAccountHookResponse, error)
	// IsAccountHookNameAvailable(ctx context.Context, req *mgmtv1alpha1.IsAccountHookNameAvailableRequest) (*mgmtv1alpha1.IsAccountHookNameAvailableResponse, error)
	// SetAccountHookEnabled(ctx context.Context, req *mgmtv1alpha1.SetAccountHookEnabledRequest) (*mgmtv1alpha1.SetAccountHookEnabledResponse, error)
	// GetActiveAccountHooksByEvent(ctx context.Context, req *mgmtv1alpha1.GetActiveAccountHooksByEventRequest) (*mgmtv1alpha1.GetActiveAccountHooksByEventResponse, error)
}

type config struct {
	isEnabled bool
}

type Option func(*config)

func New(
	db *neosyncdb.NeosyncDb,
	userdataclient userdata.Interface,
	opts ...Option,
) *Service {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Service{cfg: cfg, db: db, userdataclient: userdataclient}
}

func (s *Service) GetAccountHooks(ctx context.Context, req *mgmtv1alpha1.GetAccountHooksRequest) (*mgmtv1alpha1.GetAccountHooksResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(mgmtv1alpha1connect.AccountHookServiceGetAccountHooksProcedure)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_View); err != nil {
		return nil, err
	}

	// hooks, err := s.db.Q.GetAccountHooksByAccount(ctx, s.db.Db, req.GetAccountId())

	return &mgmtv1alpha1.GetAccountHooksResponse{
		Hooks: []*mgmtv1alpha1.AccountHook{},
	}, nil
}
