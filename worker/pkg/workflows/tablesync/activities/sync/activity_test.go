package sync_activity

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	continuation_token "github.com/nucleuscloud/neosync/internal/continuation-token"
	"github.com/nucleuscloud/neosync/internal/testutil"

	benthosstream "github.com/nucleuscloud/neosync/internal/benthos-stream"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.temporal.io/sdk/log"
	tmprl_mocks "go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

func Test_Sync_RunContext_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := benthosstream.NewBenthosStreamManager()

	mux := http.NewServeMux()
	benthosConfig := strings.TrimSpace(`
input:
  generate:
    count: 1
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  stdout:
    codec: lines
`)
	accountId := uuid.NewString()

	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetBenthosConfigExternalId("test") {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(benthosConfig),
				}), nil
			} else if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetConnectionIdsExternalId() {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(`["conn-id-1"]`),
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{
					Id: "conn-id-1",
				},
			}), nil
		},
	))

	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithCloseOnRelease())
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider(), connectionmanager.WithCloseOnRelease())
	var meter metric.Meter
	temporalclient := tmprl_mocks.NewClient(t)

	activity := New(connclient, jobclient, sqlconnmanager, mongoconnmanager, meter, benthosStreamManager, temporalclient, nil, nil)

	env.RegisterActivity(activity.SyncTable)

	val, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
		Id:                "test",
		AccountId:         accountId,
		JobRunId:          "job-run-id",
		ContinuationToken: nil,
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_RunContext_WithContinuationToken(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := benthosstream.NewBenthosStreamManager()

	mux := http.NewServeMux()
	benthosConfig := strings.TrimSpace(`
input:
  generate:
    count: 1000
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  stdout:
    codec: lines
`)
	accountId := uuid.NewString()

	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetBenthosConfigExternalId("test") {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(benthosConfig),
				}), nil
			} else if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetConnectionIdsExternalId() {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(`["conn-id-1"]`),
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{
					Id: "conn-id-1",
				},
			}), nil
		},
	))

	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlconnmanager := connectionmanager.NewConnectionManager(
		sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}),
		connectionmanager.WithCloseOnRelease(),
	)
	mongoconnmanager := connectionmanager.NewConnectionManager(
		mongoprovider.NewProvider(),
		connectionmanager.WithCloseOnRelease(),
	)
	var meter metric.Meter
	temporalclient := tmprl_mocks.NewClient(t)

	activity := New(connclient, jobclient, sqlconnmanager, mongoconnmanager, meter, benthosStreamManager, temporalclient, nil, nil)
	env.RegisterActivity(activity.SyncTable)

	t.Run("valid continuation token", func(t *testing.T) {
		validToken := continuation_token.
			NewFromContents(continuation_token.NewContents([]any{"dummy"})).
			String()

		val, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
			Id:                "test",
			AccountId:         accountId,
			JobRunId:          "job-run-id",
			ContinuationToken: &validToken,
		}, &SyncMetadata{Schema: "public", Table: "test"})
		require.NoError(t, err)

		var resp SyncTableResponse
		err = val.Get(&resp)
		require.NoError(t, err)

		require.Nil(t, resp.ContinuationToken)
	})

	t.Run("invalid continuation token", func(t *testing.T) {
		invalidToken := "not-a-valid-token"
		_, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
			Id:                "test",
			AccountId:         accountId,
			JobRunId:          "job-run-id",
			ContinuationToken: &invalidToken,
		}, &SyncMetadata{Schema: "public", Table: "test"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to load continuation token")
	})
}

func Test_Sync_Run_No_BenthosConfig(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := benthosstream.NewBenthosStreamManager()
	temporalclient := tmprl_mocks.NewClient(t)
	activity := New(nil, nil, nil, nil, nil, benthosStreamManager, temporalclient, nil, nil)

	env.RegisterActivity(activity.SyncTable)

	val, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err)
	require.Nil(t, val)
}

func Test_Sync_Run_Metrics_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()

	benthosConfig := strings.TrimSpace(`
input:
  generate:
    count: 1
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  stdout:
    codec: lines
metrics:
  otel_collector: {}
  `)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetBenthosConfigExternalId("test") {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(benthosConfig),
				}), nil
			} else if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetConnectionIdsExternalId() {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(`["conn-id-1"]`),
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{
					Id: "conn-id-1",
				},
			}), nil
		},
	))

	srv := startHTTPServer(t, mux)

	meterProvider := metricsdk.NewMeterProvider()
	meter := meterProvider.Meter("test")
	benthosStreamManager := benthosstream.NewBenthosStreamManager()
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithCloseOnRelease())
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider(), connectionmanager.WithCloseOnRelease())
	temporalclient := tmprl_mocks.NewClient(t)
	activity := New(connclient, jobclient, sqlconnmanager, mongoconnmanager, meter, benthosStreamManager, temporalclient, nil, nil)

	env.RegisterActivity(activity.SyncTable)

	val, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
		Id:                "test",
		AccountId:         accountId,
		JobRunId:          "job-run-id",
		ContinuationToken: nil,
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Run_Processor_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	benthosConfig := strings.TrimSpace(`
  input:
    generate:
      count: 1000
      interval: ""
      mapping: 'root = { "name": "nick" }'
  pipeline:
    threads: 1
    processors:
      - error:
          error_msg: ${! error()}
  output:
    label: ""
    stdout:
      codec: lines
  `)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetBenthosConfigExternalId("test") {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(benthosConfig),
				}), nil
			} else if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetConnectionIdsExternalId() {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(`["conn-id-1"]`),
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{
					Id: "conn-id-1",
				},
			}), nil
		},
	))

	srv := startHTTPServer(t, mux)

	benthosStreamManager := benthosstream.NewBenthosStreamManager()
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithCloseOnRelease())
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider(), connectionmanager.WithCloseOnRelease())
	var meter metric.Meter
	temporalclient := tmprl_mocks.NewClient(t)
	activity := New(connclient, jobclient, sqlconnmanager, mongoconnmanager, meter, benthosStreamManager, temporalclient, nil, nil)

	env.RegisterActivity(activity.SyncTable)

	_, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
		Id:                "test",
		AccountId:         accountId,
		JobRunId:          "job-run-id",
		ContinuationToken: nil,
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err, "error was nil when it should be present")
}

func Test_Sync_Run_Output_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	benthosConfig := strings.TrimSpace(`
input:
  generate:
    count: 1000
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  error:
    error_msg: ${! meta("fallback_error")}
`)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetBenthosConfigExternalId("test") {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(benthosConfig),
				}), nil
			} else if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetConnectionIdsExternalId() {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(`["conn-id-1"]`),
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{
					Id: "conn-id-1",
				},
			}), nil
		},
	))

	srv := startHTTPServer(t, mux)

	mockBenthosStreamManager := benthosstream.NewMockBenthosStreamManagerClient(t)
	mockBenthosStream := benthosstream.NewMockBenthosStreamClient(t)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithCloseOnRelease())
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider(), connectionmanager.WithCloseOnRelease())
	var meter metric.Meter
	temporalclient := tmprl_mocks.NewClient(t)
	activity := New(connclient, jobclient, sqlconnmanager, mongoconnmanager, meter, mockBenthosStreamManager, temporalclient, nil, nil)

	env.RegisterActivity(activity.SyncTable)

	errmsg := "duplicate key value violates unique constraint"
	mockBenthosStreamManager.
		On("NewBenthosStreamFromBuilder", mock.Anything).
		Return(mockBenthosStream, nil).
		Once()

	mockBenthosStream.
		On("Run", mock.Anything).
		Return(errors.New(errmsg)).
		Once()

	mockBenthosStream.
		On("StopWithin", mock.Anything).
		Return(nil).
		Maybe()

	_, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
		Id:                "test",
		AccountId:         accountId,
		JobRunId:          "job-run-id",
		ContinuationToken: nil,
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err, "error was nil when it should be present")
	require.Contains(t, err.Error(), "activity error")
}

func Test_Sync_Run_BenthosError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	mockBenthosStreamManager := benthosstream.NewMockBenthosStreamManagerClient(t)
	mockBenthosStream := benthosstream.NewMockBenthosStreamClient(t)
	benthosConfig := strings.TrimSpace(`
input:
  generate:
    count: 1000
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  stdout:
    codec: lines
`)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetBenthosConfigExternalId("test") {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(benthosConfig),
				}), nil
			} else if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetConnectionIdsExternalId() {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: []byte(`["conn-id-1"]`),
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{
					Id: "conn-id-1",
				},
			}), nil
		},
	))

	srv := startHTTPServer(t, mux)

	mockBenthosStreamManager.On("NewBenthosStreamFromBuilder", mock.Anything).Return(mockBenthosStream, nil)
	errmsg := "benthos error"
	mockBenthosStream.On("Run", mock.Anything).Return(errors.New(errmsg))
	mockBenthosStream.On("StopWithin", mock.Anything).Return(nil).Maybe()

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithCloseOnRelease())
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider(), connectionmanager.WithCloseOnRelease())
	var meter metric.Meter
	temporalclient := tmprl_mocks.NewClient(t)
	activity := New(connclient, jobclient, sqlconnmanager, mongoconnmanager, meter, mockBenthosStreamManager, temporalclient, nil, nil)

	env.RegisterActivity(activity.SyncTable)
	_, err := env.ExecuteActivity(activity.SyncTable, &SyncTableRequest{
		Id:                "test",
		AccountId:         accountId,
		JobRunId:          "job-run-id",
		ContinuationToken: nil,
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err)
	require.Contains(t, err.Error(), errmsg)
}

func Test_getEnvVarLookupFn(t *testing.T) {
	fn := getEnvVarLookupFn(nil)
	assert.NotNil(t, fn)
	val, ok := fn("foo")
	assert.False(t, ok)
	assert.Empty(t, val)

	fn = getEnvVarLookupFn(map[string]string{"foo": "bar"})
	assert.NotNil(t, fn)
	val, ok = fn("foo")
	assert.True(t, ok)
	assert.Equal(t, val, "bar")

	val, ok = fn("bar")
	assert.False(t, ok)
	assert.Empty(t, val)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
