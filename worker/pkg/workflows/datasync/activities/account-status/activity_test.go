package accountstatus_activity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_New(t *testing.T) {
	a := New(mgmtv1alpha1connect.NewMockUserAccountServiceClient(t))
	require.NotNil(t, a)
}

func Test_Activity_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.IsAccountStatusValidRequest]) (*connect.Response[mgmtv1alpha1.IsAccountStatusValidResponse], error) {
			if r.Msg.GetAccountId() == accountId {
				return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
			}
			return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: false}), nil
		},
	))
	srv := startHTTPServer(t, mux)

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	activity := New(userclient)

	env.RegisterActivity(activity)

	val, err := env.ExecuteActivity(activity.CheckAccountStatus, &CheckAccountStatusRequest{AccountId: accountId})
	require.NoError(t, err)
	res := &CheckAccountStatusResponse{}
	err = val.Get(res)
	require.NoError(t, err)
	require.True(t, res.IsValid)
	require.Nil(t, res.Reason)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
