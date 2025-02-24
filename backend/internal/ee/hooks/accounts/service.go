package accounthooks

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
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
}

type config struct {
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

	return &mgmtv1alpha1.UpdateAccountHookResponse{
		Hook: dto,
	}, nil
}

var (
	// We're using a 32 byte long secret key.
	// This is probably something you generate first
	// then put into and environment variable.
	secretKey string = "N1PCdw3M2B1TfJhoaY2mL736p2vCUc47"
)

func encrypt(plaintext string) (string, error) {
	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return "", err
	}

	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return string(ciphertext), nil
}

func decrypt(ciphertext string) (string, error) {
	aesCipher, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return "", err
	}

	// Since we know the ciphertext is actually nonce+ciphertext
	// And len(nonce) == NonceSize(). We can separate the two.
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

type slackOauthState struct {
	AccountId string `json:"accountId"`
	UserId    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
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

	slackClientId := "33336676.569200954261"
	slackRedirectUrl := "https://hooks.neosync.dev/slack/oauth/callback"

	state := &slackOauthState{
		AccountId: req.GetAccountId(),
		UserId:    user.Id(),
		Timestamp: time.Now().Unix(),
	}

	stateBits, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal slack oauth state: %w", err)
	}

	stateEncrypted, err := encrypt(string(stateBits))
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt slack oauth state: %w", err)
	}
	logger.Debug("slack state encrypted")

	slackUrl := &url.URL{
		Scheme: "https",
		Host:   "slack.com",
		Path:   "/oauth/v2/authorize",
		RawQuery: url.Values{
			"scope":        {"channels:join,channels:read,chat:write"},
			"client_id":    {slackClientId},
			"redirect_uri": {slackRedirectUrl},
			"state":        {stateEncrypted},
		}.Encode(),
	}

	return &mgmtv1alpha1.GetSlackConnectionUrlResponse{
		Url: slackUrl.String(),
	}, nil
}

func (s *Service) HandleSlackOAuthCallback(
	ctx context.Context,
	req *mgmtv1alpha1.HandleSlackOAuthCallbackRequest,
) (*mgmtv1alpha1.HandleSlackOAuthCallbackResponse, error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.GetAccountId())

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := user.EnforceAccount(ctx, userdata.NewIdentifier(req.GetAccountId()), rbac.AccountAction_Edit); err != nil {
		return nil, err
	}

	stateEncrypted := req.GetState()
	stateDecrypted, err := decrypt(stateEncrypted)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt slack oauth state: %w", err)
	}
	logger.Debug("slack state decrypted")
	var state slackOauthState
	if err := json.Unmarshal([]byte(stateDecrypted), &state); err != nil {
		return nil, fmt.Errorf("unable to unmarshal slack oauth state: %w", err)
	}

	if state.AccountId != req.GetAccountId() {
		return nil, nucleuserrors.NewBadRequest("invalid account id")
	}

	if state.UserId != user.Id() {
		return nil, nucleuserrors.NewBadRequest("invalid user id")
	}

	// Validate timestamp is within 15 minutes
	if time.Now().Unix()-state.Timestamp > 900 {
		return nil, nucleuserrors.NewBadRequest("oauth state expired")
	}

	slackCode := req.GetCode()
	_ = slackCode

	accountId, err := neosyncdb.ToUuid(req.GetAccountId())
	if err != nil {
		return nil, err
	}

	accessToken := "" // todo: exchange code for access token

	_, err = s.db.Q.CreateSlackAccessToken(ctx, s.db.Db, db_queries.CreateSlackAccessTokenParams{
		AccountID:       accountId,
		AccessToken:     accessToken,
		CreatedByUserID: user.PgId(),
		UpdatedByUserID: user.PgId(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to store slack access token: %w", err)
	}

	return &mgmtv1alpha1.HandleSlackOAuthCallbackResponse{}, nil
}
