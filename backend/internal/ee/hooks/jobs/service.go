package jobhooks

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

type Service struct {
	cfg                *config
	db                 *neosyncdb.NeosyncDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
}

var _ Interface = (*Service)(nil)

type Interface interface {
	GetJobHooks(ctx context.Context, req *mgmtv1alpha1.GetJobHooksRequest) (*mgmtv1alpha1.GetJobHooksResponse, error)
	GetJobHook(ctx context.Context, req *mgmtv1alpha1.GetJobHookRequest) (*mgmtv1alpha1.GetJobHookResponse, error)
	CreateJobHook(ctx context.Context, req *mgmtv1alpha1.CreateJobHookRequest) (*mgmtv1alpha1.CreateJobHookResponse, error)
	DeleteJobHook(ctx context.Context, req *mgmtv1alpha1.DeleteJobHookRequest) (*mgmtv1alpha1.DeleteJobHookResponse, error)
	IsJobHookNameAvailable(ctx context.Context, req *mgmtv1alpha1.IsJobHookNameAvailableRequest) (*mgmtv1alpha1.IsJobHookNameAvailableResponse, error)
}

type config struct {
	isEnabled bool
}

func WithEnabled() Option {
	return func(c *config) {
		c.isEnabled = true
	}
}

type Option func(*config)

func New(
	db *neosyncdb.NeosyncDb,
	useraccountservice mgmtv1alpha1connect.UserAccountServiceClient,
	opts ...Option,
) *Service {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Service{cfg: cfg, db: db, useraccountService: useraccountservice}
}

func (s *Service) GetJobHooks(
	ctx context.Context,
	req *mgmtv1alpha1.GetJobHooksRequest,
) (*mgmtv1alpha1.GetJobHooksResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(mgmtv1alpha1connect.JobServiceGetJobHooksProcedure)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId())
	if err != nil {
		return nil, err
	}
	logger = logger.With("accountId", neosyncdb.UUIDString(verifyResp.AccountUuid))

	hooks, err := s.db.Q.GetJobHooksByJob(ctx, s.db.Db, verifyResp.JobUuid)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("found %d hooks", len(hooks)))

	dtos := make([]*mgmtv1alpha1.JobHook, len(hooks))
	for idx := range hooks {
		hook := hooks[idx]
		dto, err := dtomaps.ToJobHookDto(&hook)
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
	}

	return &mgmtv1alpha1.GetJobHooksResponse{Hooks: dtos}, nil
}

func (s *Service) GetJobHook(
	ctx context.Context,
	req *mgmtv1alpha1.GetJobHookRequest,
) (*mgmtv1alpha1.GetJobHookResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(mgmtv1alpha1connect.JobServiceGetJobHookProcedure)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	hookuuid, err := neosyncdb.ToUuid(req.GetId())
	if err != nil {
		return nil, err
	}

	hook, err := s.db.Q.GetJobHookById(ctx, s.db.Db, hookuuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job hook by id")
	}

	verifyResp, err := s.verifyUserHasJob(ctx, neosyncdb.UUIDString(hook.JobID))
	if err != nil {
		return nil, err
	}
	logger = logger.With(
		"accountId", neosyncdb.UUIDString(verifyResp.AccountUuid),
		"jobId", neosyncdb.UUIDString(verifyResp.JobUuid),
	)

	logger.Debug("hook successfully found")

	dto, err := dtomaps.ToJobHookDto(&hook)
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.GetJobHookResponse{Hook: dto}, nil
}

func (s *Service) DeleteJobHook(
	ctx context.Context,
	req *mgmtv1alpha1.DeleteJobHookRequest,
) (*mgmtv1alpha1.DeleteJobHookResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(mgmtv1alpha1connect.JobServiceGetJobHooksProcedure)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	hookuuid, err := neosyncdb.ToUuid(req.GetId())
	if err != nil {
		return nil, err
	}

	hook, err := s.db.Q.GetJobHookById(ctx, s.db.Db, hookuuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		logger.Debug("unable to find hook during deletion")
		return &mgmtv1alpha1.DeleteJobHookResponse{}, nil
	}

	verifyResp, err := s.verifyUserHasJob(ctx, neosyncdb.UUIDString(hook.JobID))
	if err != nil {
		return nil, err
	}
	logger = logger.With(
		"accountId", neosyncdb.UUIDString(verifyResp.AccountUuid),
		"jobId", neosyncdb.UUIDString(verifyResp.JobUuid),
	)
	logger.Debug("attempting to remove hook")
	err = s.db.Q.RemoveJobHookById(ctx, s.db.Db, hookuuid)
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.DeleteJobHookResponse{}, nil
}

func (s *Service) IsJobHookNameAvailable(
	ctx context.Context,
	req *mgmtv1alpha1.IsJobHookNameAvailableRequest,
) (*mgmtv1alpha1.IsJobHookNameAvailableResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(mgmtv1alpha1connect.JobServiceIsJobHookNameAvailableProcedure)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	jobuuid, err := neosyncdb.ToUuid(req.GetJobId())
	if err != nil {
		return nil, err
	}
	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId())
	if err != nil {
		return nil, err
	}
	logger = logger.With(
		"accountId", neosyncdb.UUIDString(verifyResp.AccountUuid),
		"jobId", neosyncdb.UUIDString(verifyResp.JobUuid),
	)
	logger.Debug("checking if job hook name is available")
	ok, err := s.db.Q.IsJobHookNameAvailable(ctx, s.db.Db, db_queries.IsJobHookNameAvailableParams{
		JobID: jobuuid,
		Name:  req.GetName(),
	})
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.IsJobHookNameAvailableResponse{IsAvailable: ok}, nil
}

func (s *Service) CreateJobHook(
	ctx context.Context,
	req *mgmtv1alpha1.CreateJobHookRequest,
) (*mgmtv1alpha1.CreateJobHookResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(mgmtv1alpha1connect.JobServiceCreateJobHookProcedure)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	jobuuid, err := neosyncdb.ToUuid(req.GetJobId())
	if err != nil {
		return nil, err
	}
	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId())
	if err != nil {
		return nil, err
	}
	logger = logger.With(
		"accountId", neosyncdb.UUIDString(verifyResp.AccountUuid),
		"jobId", neosyncdb.UUIDString(verifyResp.JobUuid),
	)

	useruuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	// todo: verify all connections are within the account as well as the job
	hookReq := req.GetHook()
	logger.Debug(fmt.Sprintf("attempting to create new job hook %q", hookReq.GetName()))

	config, err := hookReq.GetConfig().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("unable to map config to valid json for db storage: %w", err)
	}

	priority, err := safeInt32(hookReq.GetPriority())
	if err != nil {
		return nil, err
	}

	hook, err := s.db.Q.CreateJobHook(ctx, s.db.Db, db_queries.CreateJobHookParams{
		Name:            hookReq.GetName(),
		Description:     hookReq.GetDescription(),
		JobID:           jobuuid,
		Enabled:         hookReq.GetEnabled(),
		Priority:        priority,
		CreatedByUserID: *useruuid,
		UpdatedByUserID: *useruuid,
		Config:          config,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create job hook: %w", err)
	}

	dto, err := dtomaps.ToJobHookDto(&hook)
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.CreateJobHookResponse{Hook: dto}, nil
}

type verifyUserJobResponse struct {
	JobUuid     pgtype.UUID
	AccountUuid pgtype.UUID
}

func (s *Service) verifyUserHasJob(ctx context.Context, jobId string) (*verifyUserJobResponse, error) {
	jobuuid, err := neosyncdb.ToUuid(jobId)
	if err != nil {
		return nil, err
	}

	accountUuid, err := s.db.Q.GetAccountIdFromJobId(ctx, s.db.Db, jobuuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find job id")
	}
	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(accountUuid))
	if err != nil {
		return nil, err
	}
	return &verifyUserJobResponse{
		JobUuid:     jobuuid,
		AccountUuid: accountUuid,
	}, nil
}

func safeInt32(v uint32) (int32, error) {
	if v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d exceeds max int32", v)
	}
	return int32(v), nil //nolint:gosec // safe due to check above
}