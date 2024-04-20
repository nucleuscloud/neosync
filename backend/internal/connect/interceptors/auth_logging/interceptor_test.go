package authlogging_interceptor

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Interceptor_WrapUnary_JwtContextData_ValidUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	genuuid, _ := nucleusdb.ToUuid(uuid.NewString())
	mockQuerier.On("GetUserByProviderSub", mock.Anything, mock.Anything, "auth-user-id").
		Return(db_queries.NeosyncApiUser{ID: genuuid}, nil)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{UserId: "123"}), nil
		},
		connect.WithInterceptors(
			logger_interceptor.NewInterceptor(logger),
			&mockAuthInterceptor{data: &auth_jwt.TokenContextData{AuthUserId: "auth-user-id"}},
			NewInterceptor(nucleusdb.New(mockDbtx, mockQuerier)),
		),
	))

	srv := startHTTPServer(t, mux)
	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	_, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	require.NoError(t, err)
}

func Test_Interceptor_WrapUnary_JwtContextData_NoUser_NoFail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.UserAccountServiceGetUserProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.UserAccountServiceGetUserProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetUserResponse{UserId: "123"}), nil
		},
		connect.WithInterceptors(
			logger_interceptor.NewInterceptor(logger),
			NewInterceptor(nucleusdb.New(mockDbtx, mockQuerier)),
		),
	))

	srv := startHTTPServer(t, mux)
	client := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	_, err := client.GetUser(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	require.NoError(t, err)
}

type mockAuthInterceptor struct {
	data *auth_jwt.TokenContextData
}

func (i *mockAuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		return next(context.WithValue(ctx, auth_jwt.TokenContextKey{}, i.data), request)
	}
}

func (i *mockAuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *mockAuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	}
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}

func Test_getAuthValues_NoTokenCtx(t *testing.T) {
	vals := getAuthValues(context.Background(), &nucleusdb.NucleusDb{})
	require.Empty(t, vals)
}

func Test_getAuthValues_Valid_Jwt(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	uuidstr := uuid.NewString()
	genuuid, _ := nucleusdb.ToUuid(uuidstr)
	mockQuerier.On("GetUserByProviderSub", mock.Anything, mock.Anything, "auth-user-id").
		Return(db_queries.NeosyncApiUser{ID: genuuid}, nil)

	ctx := context.WithValue(context.Background(), auth_jwt.TokenContextKey{}, &auth_jwt.TokenContextData{
		AuthUserId: "auth-user-id",
	})

	vals := getAuthValues(ctx, nucleusdb.New(mockDbtx, mockQuerier))
	require.Equal(
		t,
		[]any{"authUserId", "auth-user-id", "userId", uuidstr},
		vals,
	)
}

func Test_getAuthValues_Valid_Jwt_No_User(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	mockQuerier.On("GetUserByProviderSub", mock.Anything, mock.Anything, "auth-user-id").
		Return(db_queries.NeosyncApiUser{}, errors.New("test err"))

	ctx := context.WithValue(context.Background(), auth_jwt.TokenContextKey{}, &auth_jwt.TokenContextData{
		AuthUserId: "auth-user-id",
	})

	vals := getAuthValues(ctx, nucleusdb.New(mockDbtx, mockQuerier))
	require.Equal(
		t,
		[]any{"authUserId", "auth-user-id"},
		vals,
	)
}

func Test_getAuthValues_Valid_ApiKey(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	apikeyid := uuid.NewString()
	accountid := uuid.NewString()
	userid := uuid.NewString()

	apikeyuuid, _ := nucleusdb.ToUuid(apikeyid)
	accountiduuid, _ := nucleusdb.ToUuid(accountid)
	useriduuid, _ := nucleusdb.ToUuid(userid)

	ctx := context.WithValue(context.Background(), auth_apikey.TokenContextKey{}, &auth_apikey.TokenContextData{
		ApiKeyType: apikey.AccountApiKey,
		ApiKey: &db_queries.NeosyncApiAccountApiKey{
			ID:        apikeyuuid,
			AccountID: accountiduuid,
			UserID:    useriduuid,
		},
	})

	vals := getAuthValues(ctx, nucleusdb.New(mockDbtx, mockQuerier))
	require.Equal(
		t,
		[]any{"apiKeyType", apikey.AccountApiKey, "apiKeyId", apikeyid, "accountId", accountid, "userId", userid},
		vals,
	)
}

func Test_getAuthValues_Valid_ApiKey_No_Apikey(t *testing.T) {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)

	ctx := context.WithValue(context.Background(), auth_apikey.TokenContextKey{}, &auth_apikey.TokenContextData{
		ApiKeyType: apikey.AccountApiKey,
	})

	vals := getAuthValues(ctx, nucleusdb.New(mockDbtx, mockQuerier))
	require.Equal(
		t,
		[]any{"apiKeyType", apikey.AccountApiKey},
		vals,
	)
}
