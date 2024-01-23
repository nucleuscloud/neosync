package auth_interceptor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/assert"
)

func Test_Interceptor_WrapUnary_Disallow_All(t *testing.T) {
	interceptor := NewInterceptor(func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
		return nil, errors.New("no dice")
	})

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("oh no"))
		},
		connect.WithInterceptors(interceptor),
	))
	srv := startHTTPServer(t, mux)

	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	resp, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no dice")
	assert.Nil(t, resp)
}

func Test_Interceptor_WrapUnary_Allow_All(t *testing.T) {
	interceptor := NewInterceptor(func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
		return ctx, nil
	})

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("oh no"))
		},
		connect.WithInterceptors(interceptor),
	))
	srv := startHTTPServer(t, mux)

	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	resp, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "oh no")
	assert.Nil(t, resp)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}

func Test_Interceptor_WrapUnary_WithExclude(t *testing.T) {
	interceptor := NewInterceptorWithExclude(
		func(ctx context.Context, header http.Header, spec connect.Spec) (context.Context, error) {
			return nil, errors.New("no dice")
		},
		[]string{mgmtv1alpha1connect.UserAccountServiceSetUserProcedure},
	)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{}), nil
		},
		connect.WithInterceptors(interceptor),
	))
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceSetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceSetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.SetUserRequest]) (*connect.Response[mgmtv1alpha1.SetUserResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.SetUserResponse{}), nil
		},
		connect.WithInterceptors(interceptor),
	))
	srv := startHTTPServer(t, mux)

	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	resp, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no dice")
	assert.Nil(t, resp)

	resp2, err := client.SetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	assert.NoError(t, err)
	assert.NotNil(t, resp2)
}
