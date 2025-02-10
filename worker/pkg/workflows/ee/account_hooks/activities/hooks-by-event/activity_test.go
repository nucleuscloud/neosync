package hooks_by_event_activity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

func Test_New(t *testing.T) {
	accounthookclient := mgmtv1alpha1connect.NewMockAccountHookServiceClient(t)
	activity := New(accounthookclient)
	require.NotNil(t, activity)
}

func Test_Activity_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	hookId := uuid.NewString()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.AccountHookServiceGetActiveAccountHooksByEventProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.AccountHookServiceGetActiveAccountHooksByEventProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetActiveAccountHooksByEventRequest]) (*connect.Response[mgmtv1alpha1.GetActiveAccountHooksByEventResponse], error) {
			if r.Msg.GetAccountId() == accountId && r.Msg.GetEvent() == mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED {
				return connect.NewResponse(&mgmtv1alpha1.GetActiveAccountHooksByEventResponse{
					Hooks: []*mgmtv1alpha1.AccountHook{
						{Id: hookId},
					},
				}), nil
			}
			return nil, nil
		},
	))
	srv := startHTTPServer(t, mux)
	accounthookclient := mgmtv1alpha1connect.NewAccountHookServiceClient(srv.Client(), srv.URL)
	activity := New(accounthookclient)

	env.RegisterActivity(activity)

	val, err := env.ExecuteActivity(activity.GetHooksByEvent, &RunHooksByEventRequest{
		AccountId: accountId,
		EventName: mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED,
	})
	require.NoError(t, err)
	res := &RunHooksByEventResponse{}
	err = val.Get(res)
	require.NoError(t, err)
	require.Equal(t, len(res.HookIds), 1)
	require.Equal(t, res.HookIds[0], hookId)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
