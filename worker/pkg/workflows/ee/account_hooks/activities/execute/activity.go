package execute_hook_activity

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	accounthook_events "github.com/nucleuscloud/neosync/internal/ee/events"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	accounthookclient mgmtv1alpha1connect.AccountHookServiceClient
}

func New(
	accounthookclient mgmtv1alpha1connect.AccountHookServiceClient,
) *Activity {
	return &Activity{accounthookclient: accounthookclient}
}

type ExecuteHookRequest struct {
	HookId string
	Event  *accounthook_events.Event
}

type ExecuteHookResponse struct{}

func (a *Activity) ExecuteAccountHook(
	ctx context.Context,
	req *ExecuteHookRequest,
) (*ExecuteHookResponse, error) {
	if req.Event == nil {
		return nil, errors.New("event is required")
	}
	activityInfo := activity.GetInfo(ctx)
	loggerKeyVals := []any{
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
		"HookId", req.HookId,
		"Event", req.Event.Name.String(),
	}
	slogger := temporallogger.NewSlogger(log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
	))

	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	slogger.Debug("retrieving hook")

	resp, err := a.accounthookclient.GetAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHookRequest{
		Id: req.HookId,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve hook: %w", err)
	}

	hook := resp.Msg.GetHook()

	switch cfg := hook.GetConfig().GetConfig().(type) {
	case *mgmtv1alpha1.AccountHookConfig_Webhook:
		slogger.Debug("executing webhook")
		if cfg.Webhook == nil {
			return nil, errors.New("webhook config was nil for account hook configuration")
		}
		if err := executeWebhook(ctx, cfg.Webhook, req.Event, slogger); err != nil {
			return nil, fmt.Errorf("unable to execute webhook: %w", err)
		}

	case *mgmtv1alpha1.AccountHookConfig_Slack:
		slogger.Debug("executing slack message")
		if cfg.Slack == nil {
			return nil, errors.New("slack config was nil for account hook configuration")
		}
		if err := executeSlackMessage(ctx, a.accounthookclient, req.HookId, req.Event); err != nil {
			return nil, fmt.Errorf("unable to execute slack message: %w", err)
		}
	default:
		slogger.Warn(fmt.Sprintf("hook config type %T not supported", cfg))
	}

	return &ExecuteHookResponse{}, nil
}

func executeSlackMessage(
	ctx context.Context,
	client mgmtv1alpha1connect.AccountHookServiceClient,
	hookId string,
	event *accounthook_events.Event,
) error {
	bits, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("unable to marshal event: %w", err)
	}
	_, err = client.SendSlackMessage(ctx, connect.NewRequest(&mgmtv1alpha1.SendSlackMessageRequest{
		AccountHookId: hookId,
		Event:         bits,
	}))
	if err != nil {
		return fmt.Errorf("unable to send slack message: %w", err)
	}
	return nil
}

func executeWebhook(
	ctx context.Context,
	webhook *mgmtv1alpha1.AccountHookConfig_WebHook,
	event *accounthook_events.Event,
	logger *slog.Logger,
) error {
	logger.Debug(fmt.Sprintf("webhook url: %s, skipVerify: %t", webhook.GetUrl(), webhook.GetDisableSslVerification()))
	jsonPayload, err := getPayload(event)
	if err != nil {
		return fmt.Errorf("unable to get payload: %w", err)
	}

	logger.Debug("generating hmac signature")
	signature, err := generateHmac(webhook.GetSecret(), jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to generate hmac signature: %w", err)
	}

	logger.Debug("executing webhook request")
	if err := executeWebhookRequest(ctx, webhook.GetUrl(), jsonPayload, signature, webhook.GetDisableSslVerification()); err != nil {
		return fmt.Errorf("unable to execute webhook request: %w", err)
	}

	return nil
}

const (
	WEBHOOK_USERAGENT       = "neosync"
	WEBHOOK_TIMEOUT         = 10 * time.Second
	WEBHOOK_SIG_HEADER      = "X-Neosync-Signature"
	WEBHOOK_SIG_TYPE        = "X-Neosync-Signature-Type"
	WEBHOOK_SIG_TYPE_SHA256 = "sha256"
	WEBHOOK_CONTENT_TYPE    = "application/json"
)

func executeWebhookRequest(
	ctx context.Context,
	url string,
	jsonPayload []byte,
	signature string,
	skipSslVerification bool,
) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("unable to create webhook request: %w", err)
	}

	httpReq.Header.Set("Content-Type", WEBHOOK_CONTENT_TYPE)
	httpReq.Header.Set("User-Agent", WEBHOOK_USERAGENT)

	httpReq.Header.Set(WEBHOOK_SIG_HEADER, signature)
	httpReq.Header.Set(WEBHOOK_SIG_TYPE, WEBHOOK_SIG_TYPE_SHA256)

	client := &http.Client{
		Timeout: WEBHOOK_TIMEOUT,
	}
	if skipSslVerification {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // we want to enable this if it's user specified
		}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error executing webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func generateHmac(secret string, payload []byte) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write(payload)
	if err != nil {
		return "", fmt.Errorf("unable to write payload to hmac: %w", err)
	}
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature, nil
}

type webhookPayload struct {
	EventName string `json:"event_name"`
	EventData any    `json:"event_data"`
}

func getPayload(event *accounthook_events.Event) ([]byte, error) {
	payload := webhookPayload{
		EventName: event.Name.String(),
		EventData: event,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal webhook payload: %w", err)
	}

	return jsonPayload, nil
}
