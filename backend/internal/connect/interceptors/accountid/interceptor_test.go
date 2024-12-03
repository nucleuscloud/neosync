package accountid_interceptor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_Interceptor(t *testing.T) {
	interceptor := NewInterceptor()
	mux := http.NewServeMux()
	logger := testutil.GetTestLogger(t)
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceIsUserInAccountProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceIsUserInAccountProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.IsUserInAccountRequest]) (*connect.Response[mgmtv1alpha1.IsUserInAccountResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{Ok: true}), nil
		},
		connect.WithInterceptors(logger_interceptor.NewInterceptor(logger), interceptor),
	))
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceConvertPersonalToTeamAccountProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceConvertPersonalToTeamAccountProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.ConvertPersonalToTeamAccountRequest]) (*connect.Response[mgmtv1alpha1.ConvertPersonalToTeamAccountResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.ConvertPersonalToTeamAccountResponse{}), nil
		},
	))
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceSetPersonalAccountProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceSetPersonalAccountProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.SetPersonalAccountRequest]) (*connect.Response[mgmtv1alpha1.SetPersonalAccountResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.SetPersonalAccountResponse{}), nil
		},
	))
	srv := startHTTPServer(t, mux)

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)

	t.Run("WrapUnary", func(t *testing.T) {
		t.Run("account id string", func(t *testing.T) {
			_, err := userclient.IsUserInAccount(context.Background(), connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{
				AccountId: "123",
			}))
			assert.NoError(t, err)
		})
		t.Run("account id *string", func(t *testing.T) {
			accId := "123"
			_, err := userclient.ConvertPersonalToTeamAccount(context.Background(), connect.NewRequest(&mgmtv1alpha1.ConvertPersonalToTeamAccountRequest{
				AccountId: &accId,
				Name:      "foo",
			}))
			assert.NoError(t, err)
		})
		t.Run("no account id", func(t *testing.T) {
			_, err := userclient.SetPersonalAccount(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
			assert.NoError(t, err)
		})
	})
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
