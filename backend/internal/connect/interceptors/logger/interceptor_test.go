package logger_interceptor

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_Interceptor_WrapUnary_InjectLogger(t *testing.T) {
	logger := testutil.GetTestLogger(t)
	interceptor := NewInterceptor(logger)

	var ctxlogger *slog.Logger

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			ctxlogger = GetLoggerFromContextOrDefault(ctx)
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{UserId: "123"}), nil
		},
		connect.WithInterceptors(interceptor),
	))
	srv := startHTTPServer(t, mux)

	assert.Nil(t, ctxlogger, "ctxlogger has not been set yet")
	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	_, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Nil(t, err)
	assert.NotNil(t, ctxlogger)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
