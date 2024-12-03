package bookend_logging_interceptor

import (
	"context"
	"errors"
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

func Test_Interceptor_WrapUnary_Without_Error(t *testing.T) {
	logger := testutil.GetTestLogger(t)
	interceptor := NewInterceptor()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{UserId: "123"}), nil
		},
		connect.WithInterceptors(logger_interceptor.NewInterceptor(logger), interceptor),
	))
	srv := startHTTPServer(t, mux)

	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	_, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Nil(t, err)
}

func Test_Interceptor_WrapUnary_With_Generic_Error(t *testing.T) {
	logger := testutil.GetTestLogger(t)
	interceptor := NewInterceptor()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return nil, errors.New("test")
		},
		connect.WithInterceptors(logger_interceptor.NewInterceptor(logger), interceptor),
	))
	srv := startHTTPServer(t, mux)

	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	_, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Error(t, err)
}

func Test_Interceptor_WrapUnary_With_Connect_Error(t *testing.T) {
	logger := testutil.GetTestLogger(t)
	interceptor := NewInterceptor()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("test"))
		},
		connect.WithInterceptors(logger_interceptor.NewInterceptor(logger), interceptor),
	))
	srv := startHTTPServer(t, mux)

	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	_, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Error(t, err)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
