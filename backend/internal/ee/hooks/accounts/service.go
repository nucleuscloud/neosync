package accounthooks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	accounthook_events "github.com/nucleuscloud/neosync/internal/ee/events"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
	ee_slack "github.com/nucleuscloud/neosync/internal/ee/slack"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"github.com/slack-go/slack"
)

type Service struct {
	cfg            *config
	db             *neosyncdb.NeosyncDb
	userdataclient userdata.Interface
	slackClient    ee_slack.Interface
}

var _ Interface = (*Service)(nil)

type Interface interface {
	GetAccountHooks(ctx context.Context, req *mgmtv1alpha1.GetAccountHooksRequest) (*mgmtv1alpha1.GetAccountHooksResponse, error)
	GetAccountHook(ctx context.Context, req *mgmtv1alpha1.GetAccountHookRequest) (*mgmtv1alpha1.GetAccountHookResponse, error)
	CreateAccountHook(ctx context.Context, req *mgmtv1alpha1.CreateAccountHookRequest) (*mgmtv1alpha1.CreateAccountHookResponse, error)
	UpdateAccountHook(ctx context.Context, req *mgmtv1alpha1.UpdateAccountHookRequest) (*mgmtv1alpha1.UpdateAccountHookResponse, error)
	DeleteAccountHook(ctx context.Context, req *mgmtv1alpha1.DeleteAccountHookRequest) (*mgmtv1alpha1.DeleteAccountHookResponse, error)
	IsAccountHookNameAvailable(ctx context.Context, req *mgmtv1alpha1.IsAccountHookNameAvailableRequest) (*mgmtv1alpha1.IsAccountHookNameAvailableResponse, error)
	SetAccountHookEnabled(ctx context.Context, req *mgmtv1alpha1.SetAccountHookEnabledRequest) (*mgmtv1alpha1.SetAccountHookEnabledResponse, error)
	GetActiveAccountHooksByEvent(ctx context.Context, req *mgmtv1alpha1.GetActiveAccountHooksByEventRequest) (*mgmtv1alpha1.GetActiveAccountHooksByEventResponse, error)
	GetSlackConnectionUrl(ctx context.Context, req *mgmtv1alpha1.GetSlackConnectionUrlRequest) (*mgmtv1alpha1.GetSlackConnectionUrlResponse, error)
	HandleSlackOAuthCallback(ctx context.Context, req *mgmtv1alpha1.HandleSlackOAuthCallbackRequest) (*mgmtv1alpha1.HandleSlackOAuthCallbackResponse, error)
	TestSlackConnection(ctx context.Context, req *mgmtv1alpha1.TestSlackConnectionRequest) (*mgmtv1alpha1.TestSlackConnectionResponse, error)
	SendSlackMessage(ctx context.Context, req *mgmtv1alpha1.SendSlackMessageRequest) (*mgmtv1alpha1.SendSlackMessageResponse, error)
}

type config struct {
}

type Option func(*config)

func New(
	db *neosyncdb.NeosyncDb,
	userdataclient userdata.Interface,
	slackClient ee_slack.Interface,
	opts ...Option,
) *Service {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Service{cfg: cfg, db: db, userdataclient: userdataclient, slackClient: slackClient}
}

func (s *Service) GetAccountHooks(ctx context.Context, req *mgmtv1alpha1.GetAccountHooksRequest) (*mgmtv1alpha1.GetAccountHooksResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_View); err != nil {
		return nil, err
	}

	accountId, err := neosyncdb.ToUuid(req.GetAccountId())
	if err != nil {
		return nil, err
	}

	hooks, err := s.db.Q.GetAccountHooksByAccount(ctx, s.db.Db, accountId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return &mgmtv1alpha1.GetAccountHooksResponse{
			Hooks: []*mgmtv1alpha1.AccountHook{},
		}, nil
	}
	logger.Debug(fmt.Sprintf("found %d hooks", len(hooks)))

	hooksDto, err := dtomaps.ToAccountHooksDto(hooks)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.GetAccountHooksResponse{
		Hooks: hooksDto,
	}, nil
}

func (s *Service) GetAccountHook(ctx context.Context, req *mgmtv1alpha1.GetAccountHookRequest) (*mgmtv1alpha1.GetAccountHookResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	hookuuid, err := neosyncdb.ToUuid(req.GetId())
	if err != nil {
		return nil, err
	}

	hook, err := s.db.Q.GetAccountHookById(ctx, s.db.Db, hookuuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find account hook by id")
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(neosyncdb.UUIDString(hook.AccountID)), rbac.AccountAction_View); err != nil {
		return nil, err
	}

	logger.Debug("hook successfully found")

	dto, err := dtomaps.ToAccountHookDto(&hook)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.GetAccountHookResponse{
		Hook: dto,
	}, nil
}

func (s *Service) DeleteAccountHook(ctx context.Context, req *mgmtv1alpha1.DeleteAccountHookRequest) (*mgmtv1alpha1.DeleteAccountHookResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	hookuuid, err := neosyncdb.ToUuid(req.GetId())
	if err != nil {
		return nil, err
	}

	hook, err := s.db.Q.GetAccountHookById(ctx, s.db.Db, hookuuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		logger.Debug("unable to find hook during deletion")
		return &mgmtv1alpha1.DeleteAccountHookResponse{}, nil
	}

	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(neosyncdb.UUIDString(hook.AccountID)), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}
	logger.Debug("attempting to remove hook")
	err = s.db.Q.RemoveAccountHookById(ctx, s.db.Db, hookuuid)
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.DeleteAccountHookResponse{}, nil
}

func (s *Service) IsAccountHookNameAvailable(ctx context.Context, req *mgmtv1alpha1.IsAccountHookNameAvailableRequest) (*mgmtv1alpha1.IsAccountHookNameAvailableResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookName", req.GetName(), "accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_View); err != nil {
		return nil, err
	}

	accountId, err := neosyncdb.ToUuid(req.GetAccountId())
	if err != nil {
		return nil, err
	}

	logger.Debug("checking if hook name is available")
	ok, err := s.db.Q.IsAccountHookNameAvailable(ctx, s.db.Db, db_queries.IsAccountHookNameAvailableParams{
		AccountID: accountId,
		Name:      req.GetName(),
	})
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.IsAccountHookNameAvailableResponse{
		IsAvailable: ok,
	}, nil
}

func (s *Service) SetAccountHookEnabled(ctx context.Context, req *mgmtv1alpha1.SetAccountHookEnabledRequest) (*mgmtv1alpha1.SetAccountHookEnabledResponse, error) {
	getResp, err := s.GetAccountHook(ctx, &mgmtv1alpha1.GetAccountHookRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(getResp.Hook.AccountId), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}

	if req.GetEnabled() == getResp.GetHook().GetEnabled() {
		logger.Debug("hook is already in the desired state")
		return &mgmtv1alpha1.SetAccountHookEnabledResponse{
			Hook: getResp.GetHook(),
		}, nil
	}

	hookuuid, err := neosyncdb.ToUuid(getResp.GetHook().GetId())
	if err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("attempting to update account hook enabled status from %v to %v", getResp.GetHook().GetEnabled(), req.GetEnabled()))
	updatedHook, err := s.db.Q.SetAccountHookEnabled(ctx, s.db.Db, db_queries.SetAccountHookEnabledParams{
		ID:              hookuuid,
		Enabled:         req.GetEnabled(),
		UpdatedByUserID: user.PgId(),
	})
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToAccountHookDto(&updatedHook)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.SetAccountHookEnabledResponse{
		Hook: dto,
	}, nil
}

func (s *Service) GetActiveAccountHooksByEvent(ctx context.Context, req *mgmtv1alpha1.GetActiveAccountHooksByEventRequest) (*mgmtv1alpha1.GetActiveAccountHooksByEventResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("event", req.GetEvent())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_View); err != nil {
		return nil, err
	}

	accountId, err := neosyncdb.ToUuid(req.GetAccountId())
	if err != nil {
		return nil, err
	}

	// We always want to include the unspecified (wildcard) event to return webhooks that are listening to all events
	validEvents := []int32{int32(mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED)}
	if req.GetEvent() != mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED {
		validEvents = append(validEvents, int32(req.GetEvent()))
	}

	eventStrings := make([]string, len(validEvents))
	for i, event := range validEvents {
		eventStrings[i] = mgmtv1alpha1.AccountHookEvent(event).String()
	}
	logger.Debug(fmt.Sprintf("searching for active account hooks by events %v", eventStrings))

	hooks, err := s.db.Q.GetActiveAccountHooksByEvent(ctx, s.db.Db, db_queries.GetActiveAccountHooksByEventParams{
		AccountID: accountId,
		Events:    validEvents,
	})
	if err != nil {
		return nil, err
	}

	hooksDto, err := dtomaps.ToAccountHooksDto(hooks)
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.GetActiveAccountHooksByEventResponse{
		Hooks: hooksDto,
	}, nil
}

func (s *Service) CreateAccountHook(ctx context.Context, req *mgmtv1alpha1.CreateAccountHookRequest) (*mgmtv1alpha1.CreateAccountHookResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}

	hookReq := req.GetHook()
	logger.Debug(fmt.Sprintf("attempting to create new account hook %q", hookReq.GetName()))

	config, err := json.Marshal(hookReq.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to map config to valid json for db storage: %w", err)
	}

	accountId, err := neosyncdb.ToUuid(req.GetAccountId())
	if err != nil {
		return nil, err
	}

	validEvents := []int32{}
	for _, event := range hookReq.GetEvents() {
		if _, ok := mgmtv1alpha1.AccountHookEvent_name[int32(event)]; !ok {
			return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("invalid event: %v", event))
		}
		validEvents = append(validEvents, int32(event))
	}

	hook, err := s.db.Q.CreateAccountHook(ctx, s.db.Db, db_queries.CreateAccountHookParams{
		Name:            hookReq.GetName(),
		Description:     hookReq.GetDescription(),
		AccountID:       accountId,
		Events:          validEvents,
		Config:          config,
		CreatedByUserID: user.PgId(),
		UpdatedByUserID: user.PgId(),
		Enabled:         hookReq.GetEnabled(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create account hook: %w", err)
	}
	logger.Debug("created account hook")

	dto, err := dtomaps.ToAccountHookDto(&hook)
	if err != nil {
		return nil, err
	}

	// todo: if slack, join channel

	return &mgmtv1alpha1.CreateAccountHookResponse{
		Hook: dto,
	}, nil
}

func (s *Service) UpdateAccountHook(ctx context.Context, req *mgmtv1alpha1.UpdateAccountHookRequest) (*mgmtv1alpha1.UpdateAccountHookResponse, error) {
	getResp, err := s.GetAccountHook(ctx, &mgmtv1alpha1.GetAccountHookRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("hookId", req.GetId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(getResp.GetHook().GetAccountId()), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("attempting to update account hook %q", getResp.GetHook().GetName()))

	validEvents := []int32{}
	for _, event := range req.GetEvents() {
		if _, ok := mgmtv1alpha1.AccountHookEvent_name[int32(event)]; !ok {
			return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("invalid event: %v", event))
		}
		validEvents = append(validEvents, int32(event))
	}

	config, err := json.Marshal(req.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to map config to valid json for db storage: %w", err)
	}

	hookuuid, err := neosyncdb.ToUuid(getResp.GetHook().GetId())
	if err != nil {
		return nil, err
	}

	updatedHook, err := s.db.Q.UpdateAccountHook(ctx, s.db.Db, db_queries.UpdateAccountHookParams{
		ID:              hookuuid,
		Enabled:         req.GetEnabled(),
		Name:            req.GetName(),
		Description:     req.GetDescription(),
		Events:          validEvents,
		Config:          config,
		UpdatedByUserID: user.PgId(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update account hook: %w", err)
	}

	dto, err := dtomaps.ToAccountHookDto(&updatedHook)
	if err != nil {
		return nil, err
	}

	// todo: if slack and channel has changed, join channel

	return &mgmtv1alpha1.UpdateAccountHookResponse{
		Hook: dto,
	}, nil
}

func (s *Service) GetSlackConnectionUrl(
	ctx context.Context,
	req *mgmtv1alpha1.GetSlackConnectionUrlRequest,
) (*mgmtv1alpha1.GetSlackConnectionUrlResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}

	slackUrl, err := s.slackClient.GetAuthorizeUrl(req.GetAccountId(), user.Id())
	if err != nil {
		return nil, fmt.Errorf("unable to get slack authorize url: %w", err)
	}
	logger.Debug("slack authorize url retrieved")

	return &mgmtv1alpha1.GetSlackConnectionUrlResponse{
		Url: slackUrl,
	}, nil
}

func (s *Service) HandleSlackOAuthCallback(
	ctx context.Context,
	req *mgmtv1alpha1.HandleSlackOAuthCallbackRequest,
) (*mgmtv1alpha1.HandleSlackOAuthCallbackResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get user: %w", err)
	}

	oauthState, err := s.slackClient.ValidateState(ctx, req.GetState(), user.Id(), func(ctx context.Context, userId, accountId string) (bool, error) {
		parsedAccountUuid, err := neosyncdb.ToUuid(accountId)
		if err != nil {
			return false, err
		}
		ok, err := s.db.Q.IsUserInAccount(ctx, s.db.Db, db_queries.IsUserInAccountParams{
			AccountId: parsedAccountUuid,
			UserId:    user.PgId(),
		})
		if err != nil {
			return false, fmt.Errorf("unable to check if user is in account: %w", err)
		}
		return ok != 0, nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to validate slack oauth state: %w", err)
	}

	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(oauthState.AccountId), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}
	logger.Debug("slack oauth state validated")

	slackCode := req.GetCode()
	oauthResp, err := s.slackClient.ExchangeCodeForAccessToken(ctx, slackCode)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange slack code for access token: %w", err)
	}

	oauthRespBytes, err := json.Marshal(oauthResp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal slack oauth response: %w", err)
	}

	accountUuid, err := neosyncdb.ToUuid(oauthState.AccountId)
	if err != nil {
		return nil, fmt.Errorf("unable to convert account id to uuid: %w", err)
	}

	_, err = s.db.Q.CreateSlackOAuthConnection(ctx, s.db.Db, db_queries.CreateSlackOAuthConnectionParams{
		AccountID:       accountUuid,
		OauthV2Response: oauthRespBytes,
		CreatedByUserID: user.PgId(),
		UpdatedByUserID: user.PgId(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to store slack access token: %w", err)
	}

	return &mgmtv1alpha1.HandleSlackOAuthCallbackResponse{}, nil
}

func (s *Service) TestSlackConnection(
	ctx context.Context,
	req *mgmtv1alpha1.TestSlackConnectionRequest,
) (*mgmtv1alpha1.TestSlackConnectionResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_View); err != nil {
		return nil, err
	}

	accountId, err := neosyncdb.ToUuid(req.GetAccountId())
	if err != nil {
		return nil, err
	}

	logger.Debug("retrieving slack access token")

	accessToken, err := s.db.Q.GetSlackAccessToken(ctx, s.db.Db, accountId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to get slack access token: %w", err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		msg := "slack oauth connection not found"
		return &mgmtv1alpha1.TestSlackConnectionResponse{
			HasConfiguration: false,
			Error:            &msg,
		}, nil
	}

	if accessToken == "" {
		msg := "slack access token not found"
		return &mgmtv1alpha1.TestSlackConnectionResponse{
			HasConfiguration: false,
			Error:            &msg,
		}, nil
	}

	logger.Debug("testing slack connection")
	testResp, err := s.slackClient.Test(ctx, accessToken)
	if err != nil {
		msg := err.Error()
		return &mgmtv1alpha1.TestSlackConnectionResponse{
			HasConfiguration: true,
			Error:            &msg,
		}, nil
	}
	logger.Debug("slack connection test successful")
	return &mgmtv1alpha1.TestSlackConnectionResponse{
		HasConfiguration: true,
		Error:            nil,
		TestResponse: &mgmtv1alpha1.TestSlackConnectionResponse_Response{
			Url:  testResp.URL,
			Team: testResp.Team,
		},
	}, nil
}

func (s *Service) SendSlackMessage(
	ctx context.Context,
	req *mgmtv1alpha1.SendSlackMessageRequest,
) (*mgmtv1alpha1.SendSlackMessageResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountHookId", req.GetAccountHookId())

	hook, err := s.GetAccountHook(ctx, &mgmtv1alpha1.GetAccountHookRequest{
		Id: req.GetAccountHookId(),
	})
	if err != nil {
		return nil, err
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(hook.GetHook().GetAccountId()), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}

	if !hook.GetHook().GetEnabled() {
		logger.Warn("received event for disabled account hook")
		return &mgmtv1alpha1.SendSlackMessageResponse{}, nil
	}

	slackHook := hook.GetHook().GetConfig().GetSlack()
	if slackHook == nil {
		return nil, nucleuserrors.NewNotFound("slack hook not found")
	}
	slackChannelId := slackHook.GetChannel()

	accountId, err := neosyncdb.ToUuid(hook.GetHook().GetAccountId())
	if err != nil {
		return nil, err
	}

	accessToken, err := s.db.Q.GetSlackAccessToken(ctx, s.db.Db, accountId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to get slack access token: %w", err)
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("slack oauth connection not found")
	}

	var event *accounthook_events.Event
	if err := json.Unmarshal(req.GetEvent(), &event); err != nil {
		return nil, fmt.Errorf("unable to unmarshal event: %w", err)
	}

	blocks := getSlackBlocksByEvent(event, logger)

	if len(blocks) == 0 {
		logger.Warn("received event that generated no slack blocks")
		return &mgmtv1alpha1.SendSlackMessageResponse{}, nil
	}

	logger.Debug("sending slack message")
	err = s.slackClient.SendMessage(ctx, accessToken, slackChannelId, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		return nil, fmt.Errorf("unable to send slack message: %w", err)
	}

	return &mgmtv1alpha1.SendSlackMessageResponse{}, nil
}

func getSlackBlocksByEvent(event *accounthook_events.Event, logger *slog.Logger) []slack.Block {
	switch event.Name {
	case mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_CREATED:
		if event.JobRunCreated == nil {
			logger.Warn("received job run created event with no job run created data")
			return nil
		}

		headerText := slack.NewTextBlockObject(slack.PlainTextType, "🚀 New Job Run Started", false, false)
		headerSection := slack.NewHeaderBlock(headerText)

		jobFields := []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Job ID:*\n"+event.JobRunCreated.JobId, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, "*Job Run ID:*\n"+event.JobRunCreated.JobRunId, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, "*Started At:*\n<!date^"+fmt.Sprint(event.Timestamp.Unix())+"^{date_short_pretty} at {time}|"+event.Timestamp.Format(time.RFC3339)+">", false, false),
		}
		fieldsSection := slack.NewSectionBlock(nil, jobFields, nil)

		divider := slack.NewDividerBlock()

		blocks := []slack.Block{
			headerSection,
			divider,
			fieldsSection,
		}

		return blocks

	case mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED:
		if event.JobRunFailed == nil {
			logger.Warn("received job run failed event with no job run failed data")
			return nil
		}

		headerText := slack.NewTextBlockObject(slack.PlainTextType, "� Job Run Failed", false, false)
		headerSection := slack.NewHeaderBlock(headerText)

		jobFields := []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Job ID:*\n"+event.JobRunFailed.JobId, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, "*Job Run ID:*\n"+event.JobRunFailed.JobRunId, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, "*Failed At:*\n<!date^"+fmt.Sprint(event.Timestamp.Unix())+"^{date_short_pretty} at {time}|"+event.Timestamp.Format(time.RFC3339)+">", false, false),
		}
		fieldsSection := slack.NewSectionBlock(nil, jobFields, nil)

		divider := slack.NewDividerBlock()

		blocks := []slack.Block{
			headerSection,
			divider,
			fieldsSection,
		}

		return blocks

	case mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED:
		if event.JobRunSucceeded == nil {
			logger.Warn("received job run succeeded event with no job run succeeded data")
			return nil
		}

		headerText := slack.NewTextBlockObject(slack.PlainTextType, "✅ Job Run Succeeded", false, false)
		headerSection := slack.NewHeaderBlock(headerText)

		jobFields := []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Job ID:*\n"+event.JobRunSucceeded.JobId, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, "*Job Run ID:*\n"+event.JobRunSucceeded.JobRunId, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, "*Succeeded At:*\n<!date^"+fmt.Sprint(event.Timestamp.Unix())+"^{date_short_pretty} at {time}|"+event.Timestamp.Format(time.RFC3339)+">", false, false),
		}
		fieldsSection := slack.NewSectionBlock(nil, jobFields, nil)

		divider := slack.NewDividerBlock()

		blocks := []slack.Block{
			headerSection,
			divider,
			fieldsSection,
		}

		return blocks

	default:
		logger.Warn("received unsupported slack event", "event", event.Name)
		return nil
	}
}
