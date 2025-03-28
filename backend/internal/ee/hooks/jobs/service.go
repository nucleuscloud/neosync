package jobhooks

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
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
	GetJobHooks(
		ctx context.Context,
		req *mgmtv1alpha1.GetJobHooksRequest,
	) (*mgmtv1alpha1.GetJobHooksResponse, error)
	GetJobHook(
		ctx context.Context,
		req *mgmtv1alpha1.GetJobHookRequest,
	) (*mgmtv1alpha1.GetJobHookResponse, error)
	CreateJobHook(
		ctx context.Context,
		req *mgmtv1alpha1.CreateJobHookRequest,
	) (*mgmtv1alpha1.CreateJobHookResponse, error)
	DeleteJobHook(
		ctx context.Context,
		req *mgmtv1alpha1.DeleteJobHookRequest,
	) (*mgmtv1alpha1.DeleteJobHookResponse, error)
	IsJobHookNameAvailable(
		ctx context.Context,
		req *mgmtv1alpha1.IsJobHookNameAvailableRequest,
	) (*mgmtv1alpha1.IsJobHookNameAvailableResponse, error)
	UpdateJobHook(
		ctx context.Context,
		req *mgmtv1alpha1.UpdateJobHookRequest,
	) (*mgmtv1alpha1.UpdateJobHookResponse, error)
	SetJobHookEnabled(
		ctx context.Context,
		req *mgmtv1alpha1.SetJobHookEnabledRequest,
	) (*mgmtv1alpha1.SetJobHookEnabledResponse, error)
	GetActiveJobHooksByTiming(
		ctx context.Context,
		req *mgmtv1alpha1.GetActiveJobHooksByTimingRequest,
	) (*mgmtv1alpha1.GetActiveJobHooksByTimingResponse, error)
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
	userdataclient userdata.Interface,
	opts ...Option,
) *Service {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Service{cfg: cfg, db: db, userdataclient: userdataclient}
}

func (s *Service) GetJobHooks(
	ctx context.Context,
	req *mgmtv1alpha1.GetJobHooksRequest,
) (*mgmtv1alpha1.GetJobHooksResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(
			mgmtv1alpha1connect.JobServiceGetJobHooksProcedure,
		)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId(), rbac.JobAction_View)
	if err != nil {
		return nil, err
	}
	logger = logger.With("accountId", neosyncdb.UUIDString(verifyResp.AccountUuid))

	hooks, err := s.db.Q.GetJobHooksByJob(ctx, s.db.Db, verifyResp.JobUuid)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("found %d hooks", len(hooks)))

	dtos, err := dtomaps.ToJobHooksDto(hooks)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.GetJobHooksResponse{Hooks: dtos}, nil
}

func (s *Service) GetJobHook(
	ctx context.Context,
	req *mgmtv1alpha1.GetJobHookRequest,
) (*mgmtv1alpha1.GetJobHookResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(
			mgmtv1alpha1connect.JobServiceGetJobHookProcedure,
		)
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

	verifyResp, err := s.verifyUserHasJob(
		ctx,
		neosyncdb.UUIDString(hook.JobID),
		rbac.JobAction_View,
	)
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
		return nil, nucleuserrors.NewNotImplementedProcedure(
			mgmtv1alpha1connect.JobServiceGetJobHooksProcedure,
		)
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

	verifyResp, err := s.verifyUserHasJob(
		ctx,
		neosyncdb.UUIDString(hook.JobID),
		rbac.JobAction_Delete,
	)
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
		return nil, nucleuserrors.NewNotImplementedProcedure(
			mgmtv1alpha1connect.JobServiceIsJobHookNameAvailableProcedure,
		)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	jobuuid, err := neosyncdb.ToUuid(req.GetJobId())
	if err != nil {
		return nil, err
	}
	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId(), rbac.JobAction_View)
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
		return nil, nucleuserrors.NewNotImplementedProcedure(
			mgmtv1alpha1connect.JobServiceCreateJobHookProcedure,
		)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	jobuuid, err := neosyncdb.ToUuid(req.GetJobId())
	if err != nil {
		return nil, err
	}
	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId(), rbac.JobAction_Create)
	if err != nil {
		return nil, err
	}
	logger = logger.With(
		"accountId", neosyncdb.UUIDString(verifyResp.AccountUuid),
		"jobId", neosyncdb.UUIDString(verifyResp.JobUuid),
	)

	hookReq := req.GetHook()
	logger.Debug(fmt.Sprintf("attempting to create new job hook %q", hookReq.GetName()))

	isValid, err := s.verifyHookHasValidConnections(
		ctx,
		hookReq.GetConfig(),
		jobuuid,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to validate if job hook has valid connections: %w", err)
	}
	if !isValid {
		logger.Debug("job hook creation did not pass connection id verification")
		return nil, nucleuserrors.NewBadRequest(
			"connection id specified in hook is not a part of job",
		)
	}

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
		CreatedByUserID: verifyResp.UserUuid,
		UpdatedByUserID: verifyResp.UserUuid,
		Config:          config,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create job hook: %w", err)
	}
	logger.Debug("created job hook")

	dto, err := dtomaps.ToJobHookDto(&hook)
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.CreateJobHookResponse{Hook: dto}, nil
}

func (s *Service) UpdateJobHook(
	ctx context.Context,
	req *mgmtv1alpha1.UpdateJobHookRequest,
) (*mgmtv1alpha1.UpdateJobHookResponse, error) {
	getResp, err := s.GetJobHook(ctx, &mgmtv1alpha1.GetJobHookRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	jobuuid, err := neosyncdb.ToUuid(getResp.GetHook().GetJobId())
	if err != nil {
		return nil, err
	}

	isValid, err := s.verifyHookHasValidConnections(
		ctx,
		req.GetConfig(),
		jobuuid,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to validate if job hook has valid connections: %w", err)
	}
	if !isValid {
		logger.Debug("job hook creation did not pass connection id verification")
		return nil, nucleuserrors.NewBadRequest(
			"connection id specified in hook is not a part of job",
		)
	}

	config, err := req.GetConfig().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("unable to map config to valid json for db storage: %w", err)
	}

	priority, err := safeInt32(req.GetPriority())
	if err != nil {
		return nil, err
	}

	hookuuid, err := neosyncdb.ToUuid(getResp.GetHook().GetId())
	if err != nil {
		return nil, err
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	_, err = s.verifyUserHasJob(ctx, neosyncdb.UUIDString(jobuuid), rbac.JobAction_Edit)
	if err != nil {
		return nil, err
	}

	updatedhook, err := s.db.Q.UpdateJobHook(ctx, s.db.Db, db_queries.UpdateJobHookParams{
		Name:            req.GetName(),
		Description:     req.GetDescription(),
		Config:          config,
		Enabled:         req.GetEnabled(),
		Priority:        priority,
		UpdatedByUserID: user.PgId(),
		ID:              hookuuid,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update job hook: %w", err)
	}

	dto, err := dtomaps.ToJobHookDto(&updatedhook)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.UpdateJobHookResponse{Hook: dto}, nil
}

func (s *Service) SetJobHookEnabled(
	ctx context.Context,
	req *mgmtv1alpha1.SetJobHookEnabledRequest,
) (*mgmtv1alpha1.SetJobHookEnabledResponse, error) {
	getResp, err := s.GetJobHook(ctx, &mgmtv1alpha1.GetJobHookRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	if req.GetEnabled() == getResp.GetHook().GetEnabled() {
		logger.Debug("hook enabled flag unchanged, no need to update")
		return &mgmtv1alpha1.SetJobHookEnabledResponse{Hook: getResp.GetHook()}, nil
	}

	verifyResp, err := s.verifyUserHasJob(ctx, getResp.GetHook().GetJobId(), rbac.JobAction_Edit)
	if err != nil {
		return nil, err
	}

	hookuuid, err := neosyncdb.ToUuid(getResp.GetHook().GetId())
	if err != nil {
		return nil, err
	}

	logger.Debug(
		fmt.Sprintf(
			"attempting to update job hook enabled status from %v to %v",
			getResp.GetHook().GetEnabled(),
			req.GetEnabled(),
		),
	)
	updatedHook, err := s.db.Q.SetJobHookEnabled(ctx, s.db.Db, db_queries.SetJobHookEnabledParams{
		Enabled:         req.GetEnabled(),
		UpdatedByUserID: verifyResp.UserUuid,
		ID:              hookuuid,
	})
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToJobHookDto(&updatedHook)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.SetJobHookEnabledResponse{Hook: dto}, nil
}

func (s *Service) GetActiveJobHooksByTiming(
	ctx context.Context,
	req *mgmtv1alpha1.GetActiveJobHooksByTimingRequest,
) (*mgmtv1alpha1.GetActiveJobHooksByTimingResponse, error) {
	if !s.cfg.isEnabled {
		return nil, nucleuserrors.NewNotImplementedProcedure(
			mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure,
		)
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("jobId", req.GetJobId())

	verifyResp, err := s.verifyUserHasJob(ctx, req.GetJobId(), rbac.JobAction_View)
	if err != nil {
		return nil, err
	}
	logger = logger.With("accountId", neosyncdb.UUIDString(verifyResp.AccountUuid))

	jobuuid, err := neosyncdb.ToUuid(req.GetJobId())
	if err != nil {
		return nil, err
	}

	var hooks []db_queries.NeosyncApiJobHook
	switch req.GetTiming() {
	case mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_UNSPECIFIED:
		logger.Debug("searching for active job hooks")
		hooks, err = s.db.Q.GetActiveJobHooks(ctx, s.db.Db, jobuuid)
		if err != nil {
			return nil, err
		}
	case mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC:
		logger.Debug("searching for active job presync hooks")
		hooks, err = s.db.Q.GetActivePreSyncJobHooks(ctx, s.db.Db, jobuuid)
		if err != nil {
			return nil, err
		}
	case mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_POSTSYNC:
		logger.Debug("searching for active job postsync hooks")
		hooks, err = s.db.Q.GetActivePostSyncJobHooks(ctx, s.db.Db, jobuuid)
		if err != nil {
			return nil, err
		}
	default:
		return nil, nucleuserrors.NewBadRequest(
			fmt.Sprintf("invalid hook timing: %T", req.GetTiming()),
		)
	}

	logger.Debug(fmt.Sprintf("found %d hooks", len(hooks)))

	dtos, err := dtomaps.ToJobHooksDto(hooks)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.GetActiveJobHooksByTimingResponse{Hooks: dtos}, nil
}

type verifyUserJobResponse struct {
	JobUuid     pgtype.UUID
	AccountUuid pgtype.UUID
	UserUuid    pgtype.UUID
}

func (s *Service) verifyUserHasJob(
	ctx context.Context,
	jobId string,
	permission rbac.JobAction,
) (*verifyUserJobResponse, error) {
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

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := user.EnforceJob(ctx, userdata.NewDbDomainEntity(accountUuid, jobuuid), permission); err != nil {
		return nil, err
	}

	return &verifyUserJobResponse{
		JobUuid:     jobuuid,
		AccountUuid: accountUuid,
		UserUuid:    user.PgId(),
	}, nil
}

func safeInt32(v uint32) (int32, error) {
	if v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d exceeds max int32", v)
	}
	return int32(v), nil
}

func (s *Service) verifyHookHasValidConnections(
	ctx context.Context,
	config *mgmtv1alpha1.JobHookConfig,
	jobuuid pgtype.UUID,
) (bool, error) {
	switch cfg := config.GetConfig().(type) {
	case *mgmtv1alpha1.JobHookConfig_Sql:
		if cfg.Sql == nil {
			return false, errors.New("job hook config was type Sql, but the config was nil")
		}
		connuuid, err := neosyncdb.ToUuid(cfg.Sql.GetConnectionId())
		if err != nil {
			return false, err
		}
		return s.db.Q.DoesJobHaveConnectionId(ctx, s.db.Db, db_queries.DoesJobHaveConnectionIdParams{
			JobId:        jobuuid,
			ConnectionId: connuuid,
		})
	default:
		return false, fmt.Errorf("job hook config %T is not currently supported when checking valid connections", cfg)
	}
}
