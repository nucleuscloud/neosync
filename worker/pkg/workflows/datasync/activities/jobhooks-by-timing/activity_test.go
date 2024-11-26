package jobhooks_by_timing_activity

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

type fakeELicense struct {
	isValid bool
}

func (f *fakeELicense) IsValid() bool {
	return f.isValid
}

func Test_New(t *testing.T) {
	a := New(
		mgmtv1alpha1connect.NewMockJobServiceClient(t),
		mgmtv1alpha1connect.NewMockConnectionServiceClient(t),
		sqlmanager.NewMockSqlManagerClient(t),
		&fakeELicense{isValid: true},
	)
	require.NotNil(t, a)
}

func Test_Activity_EELicense_Skip(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activity := New(
		mgmtv1alpha1connect.NewMockJobServiceClient(t),
		mgmtv1alpha1connect.NewMockConnectionServiceClient(t),
		sqlmanager.NewMockSqlManagerClient(t),
		&fakeELicense{isValid: false},
	)
	env.RegisterActivity(activity)

	val, err := env.ExecuteActivity(activity.RunJobHooksByTiming, &RunJobHooksByTimingRequest{
		JobId:  uuid.NewString(),
		Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC,
	})
	require.NoError(t, err)
	res := &RunJobHooksByTimingResponse{}
	err = val.Get(res)
	require.NoError(t, err)
	require.Equal(t, uint(0), res.ExecCount)
}

func Test_Activity_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	jobId := uuid.NewString()
	connId := uuid.NewString()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetActiveJobHooksByTimingRequest]) (*connect.Response[mgmtv1alpha1.GetActiveJobHooksByTimingResponse], error) {
			if r.Msg.GetJobId() == jobId && r.Msg.Timing == mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC {
				return connect.NewResponse(&mgmtv1alpha1.GetActiveJobHooksByTimingResponse{
					Hooks: []*mgmtv1alpha1.JobHook{
						{
							Id:       uuid.NewString(),
							Name:     "test-1",
							JobId:    jobId,
							Enabled:  true,
							Priority: 0,
							Config: &mgmtv1alpha1.JobHookConfig{
								Config: &mgmtv1alpha1.JobHookConfig_Sql{
									Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
										Query:        "truncate table public.users",
										ConnectionId: connId,
										Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
											Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
										},
									},
								},
							},
						},
						{
							Id:       uuid.NewString(),
							Name:     "test-2",
							JobId:    jobId,
							Enabled:  true,
							Priority: 0,
							Config: &mgmtv1alpha1.JobHookConfig{
								Config: &mgmtv1alpha1.JobHookConfig_Sql{
									Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
										Query:        "truncate table public.pets",
										ConnectionId: connId,
										Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
											Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
										},
									},
								},
							},
						},
					},
				}), nil
			}
			return nil, connect.NewError(connect.CodeNotFound, errors.New("invalid test input"))
		},
	))
	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			if r.Msg.GetId() == connId {
				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
					Connection: &mgmtv1alpha1.Connection{
						Id: connId,
						// leaving remaining impl out as it's not needed for this test due to mocking
					},
				}), nil
			}
			return nil, connect.NewError(connect.CodeNotFound, errors.New("invalid test input"))
		},
	))
	srv := startHTTPServer(t, mux)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	mockSqlMgrClient := sqlmanager.NewMockSqlManagerClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)

	mockSqlMgrClient.On("NewSqlConnection", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		sqlmanager.NewPostgresSqlConnection(mockSqlDb), nil,
	)
	mockSqlDb.On("Exec", mock.Anything, mock.Anything).Twice().Return(nil)
	mockSqlDb.On("Close").Once().Return(nil)

	activity := New(jobclient, connclient, mockSqlMgrClient, &fakeELicense{isValid: true})
	env.RegisterActivity(activity)

	val, err := env.ExecuteActivity(activity.RunJobHooksByTiming, &RunJobHooksByTimingRequest{
		JobId:  jobId,
		Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC,
	})
	require.NoError(t, err)
	res := &RunJobHooksByTimingResponse{}
	err = val.Get(res)
	require.NoError(t, err)
	require.Equal(t, uint(2), res.ExecCount)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
