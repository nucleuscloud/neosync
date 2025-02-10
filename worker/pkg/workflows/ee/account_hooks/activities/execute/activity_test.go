package execute_hook_activity

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/testutil"
	accounthook_events "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/events"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

func Test_New(t *testing.T) {
	a := New(
		mgmtv1alpha1connect.NewMockAccountHookServiceClient(t),
	)
	require.NotNil(t, a)
}

func Test_Activity_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	hookId := uuid.NewString()
	accountId := uuid.NewString()

	mux := http.NewServeMux()
	var srv *httptest.Server
	secret := "test-secret"
	mux.Handle(mgmtv1alpha1connect.AccountHookServiceGetAccountHookProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.AccountHookServiceGetAccountHookProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetAccountHookRequest]) (*connect.Response[mgmtv1alpha1.GetAccountHookResponse], error) {
			if r.Msg.GetId() == hookId {
				return connect.NewResponse(&mgmtv1alpha1.GetAccountHookResponse{
					Hook: &mgmtv1alpha1.AccountHook{
						Id:          hookId,
						AccountId:   accountId,
						Name:        "test-hook",
						Description: "test-description",
						Events:      []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED},
						Config: &mgmtv1alpha1.AccountHookConfig{
							Config: &mgmtv1alpha1.AccountHookConfig_Webhook{
								Webhook: &mgmtv1alpha1.AccountHookConfig_WebHook{
									Url:                    fmt.Sprintf("%s/webhook", srv.URL),
									Secret:                 secret,
									DisableSslVerification: false,
								},
							},
						},
					},
				}), nil
			}
			return nil, connect.NewError(connect.CodeNotFound, errors.New("invalid test input"))
		},
	))
	var (
		receivedSignature     string
		receivedSignatureType string
		receivedPayload       []byte
	)
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get(WEBHOOK_SIG_HEADER)
		receivedSignatureType = r.Header.Get(WEBHOOK_SIG_TYPE)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		receivedPayload = payload
		w.WriteHeader(http.StatusOK)
	})
	srv = startHTTPServer(t, mux)
	accounthookclient := mgmtv1alpha1connect.NewAccountHookServiceClient(srv.Client(), srv.URL)
	activity := New(accounthookclient)

	env.RegisterActivity(activity)

	val, err := env.ExecuteActivity(activity.ExecuteHook, &ExecuteHookRequest{
		HookId: hookId,
		Event:  accounthook_events.NewEvent_JobRunSucceeded("test-account-id", "test-job-id", "test-run-id"),
	})
	require.NoError(t, err)
	res := &ExecuteHookResponse{}
	err = val.Get(res)
	require.NoError(t, err)
	require.NotNil(t, res)

	require.Equal(t, receivedSignatureType, "sha256")
	verified, err := verifyHmac(secret, receivedPayload, receivedSignature)
	require.NoError(t, err)
	require.True(t, verified)

	var webhookPayload webhookPayload
	err = json.Unmarshal(receivedPayload, &webhookPayload)
	require.NoError(t, err)
	require.Equal(t, webhookPayload.EventName, mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED.String())
	eventData, ok := webhookPayload.EventData.(map[string]any)
	require.True(t, ok)
	jobRunSucceededEvent, ok := eventData["jobRunSucceeded"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, jobRunSucceededEvent["accountId"], "test-account-id")
	require.Equal(t, jobRunSucceededEvent["jobId"], "test-job-id")
	require.Equal(t, jobRunSucceededEvent["jobRunId"], "test-run-id")
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}

func verifyHmac(secret string, payload []byte, signature string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write(payload)
	if err != nil {
		return false, fmt.Errorf("unable to write payload to hmac: %w", err)
	}
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}
