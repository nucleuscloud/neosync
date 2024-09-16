package syncactivityopts_activity

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_New(t *testing.T) {
	a := New(mgmtv1alpha1connect.NewMockJobServiceClient(t))
	require.NotNil(t, a)
}

func Test_Activity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	syncActivityJobId := uuid.NewString()
	generatedRequestedJobId := uuid.NewString()
	aiGeneratedRequestedJobId := uuid.NewString()

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetJobProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
			if r.Msg.GetId() == syncActivityJobId {
				return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
					Job: &mgmtv1alpha1.Job{
						Id:        syncActivityJobId,
						AccountId: accountId,
						SyncOptions: &mgmtv1alpha1.ActivityOptions{
							ScheduleToCloseTimeout: shared.Ptr(int64(1)),
							StartToCloseTimeout:    shared.Ptr(int64(2)),
							RetryPolicy: &mgmtv1alpha1.RetryPolicy{
								MaximumAttempts: shared.Ptr(int32(3)),
							},
						},
						WorkflowOptions: &mgmtv1alpha1.WorkflowOptions{
							RunTimeout: shared.Ptr(int64(4)),
						},
					},
				}), nil
			}
			if r.Msg.GetId() == generatedRequestedJobId {
				return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
					Job: &mgmtv1alpha1.Job{
						Id:        syncActivityJobId,
						AccountId: accountId,
						Source: &mgmtv1alpha1.JobSource{
							Options: &mgmtv1alpha1.JobSourceOptions{
								Config: &mgmtv1alpha1.JobSourceOptions_Generate{
									Generate: &mgmtv1alpha1.GenerateSourceOptions{
										Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
											{
												Schema: "foo",
												Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
													{
														Table:    "1",
														RowCount: 1,
													},
													{
														Table:    "2",
														RowCount: 2,
													},
												},
											},
											{
												Schema: "bar",
												Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
													{
														Table:    "3",
														RowCount: 3,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}), nil
			}
			if r.Msg.GetId() == aiGeneratedRequestedJobId {
				return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
					Job: &mgmtv1alpha1.Job{
						Id:        syncActivityJobId,
						AccountId: accountId,
						Source: &mgmtv1alpha1.JobSource{
							Options: &mgmtv1alpha1.JobSourceOptions{
								Config: &mgmtv1alpha1.JobSourceOptions_AiGenerate{
									AiGenerate: &mgmtv1alpha1.AiGenerateSourceOptions{
										Schemas: []*mgmtv1alpha1.AiGenerateSourceSchemaOption{
											{
												Schema: "foo",
												Tables: []*mgmtv1alpha1.AiGenerateSourceTableOption{
													{
														Table:    "1",
														RowCount: 11,
													},
													{
														Table:    "2",
														RowCount: 22,
													},
												},
											},
											{
												Schema: "bar",
												Tables: []*mgmtv1alpha1.AiGenerateSourceTableOption{
													{
														Table:    "3",
														RowCount: 33,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}), nil
			}
			return nil, fmt.Errorf("invalid test job id")
		},
	))
	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	activity := New(jobclient)
	env.RegisterActivity(activity.RetrieveActivityOptions)

	t.Run("sync activity options", func(t *testing.T) {
		val, err := env.ExecuteActivity(activity.RetrieveActivityOptions, &RetrieveActivityOptionsRequest{JobId: syncActivityJobId})
		require.NoError(t, err)
		res := &RetrieveActivityOptionsResponse{}
		err = val.Get(res)
		require.NoError(t, err)
		require.Nil(t, res.RequestedRecordCount)
		require.Equal(t, accountId, res.AccountId)
		require.NotNil(t, res.SyncActivityOptions)
	})
	t.Run("generate requested row count", func(t *testing.T) {
		val, err := env.ExecuteActivity(activity.RetrieveActivityOptions, &RetrieveActivityOptionsRequest{JobId: generatedRequestedJobId})
		require.NoError(t, err)
		res := &RetrieveActivityOptionsResponse{}
		err = val.Get(res)
		require.NoError(t, err)
		require.Equal(t, accountId, res.AccountId)
		require.NotNil(t, res.RequestedRecordCount)
		require.Equal(t, uint64(6), *res.RequestedRecordCount)
	})
	t.Run("aigenerate requested row count", func(t *testing.T) {
		val, err := env.ExecuteActivity(activity.RetrieveActivityOptions, &RetrieveActivityOptionsRequest{JobId: aiGeneratedRequestedJobId})
		require.NoError(t, err)
		res := &RetrieveActivityOptionsResponse{}
		err = val.Get(res)
		require.NoError(t, err)
		require.Equal(t, accountId, res.AccountId)
		require.NotNil(t, res.RequestedRecordCount)
		require.Equal(t, uint64(66), *res.RequestedRecordCount)
	})
}

func Test_getSyncActivityOptionsFromJob(t *testing.T) {
	defaultOpts := &workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Minute, RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1}}
	type testcase struct {
		name     string
		input    *mgmtv1alpha1.Job
		expected *workflow.ActivityOptions
	}
	tests := []testcase{
		{name: "nil sync opts", input: &mgmtv1alpha1.Job{}, expected: defaultOpts},
		{name: "custom start to close timeout", input: &mgmtv1alpha1.Job{
			SyncOptions: &mgmtv1alpha1.ActivityOptions{
				StartToCloseTimeout: shared.Ptr(int64(2)),
			},
		}, expected: &workflow.ActivityOptions{StartToCloseTimeout: 2, RetryPolicy: defaultOpts.RetryPolicy, HeartbeatTimeout: 1 * time.Minute}},
		{name: "custom schedule to close timeout", input: &mgmtv1alpha1.Job{
			SyncOptions: &mgmtv1alpha1.ActivityOptions{
				ScheduleToCloseTimeout: shared.Ptr(int64(2)),
			},
		}, expected: &workflow.ActivityOptions{ScheduleToCloseTimeout: 2, RetryPolicy: defaultOpts.RetryPolicy, HeartbeatTimeout: 1 * time.Minute}},
		{name: "custom retry policy", input: &mgmtv1alpha1.Job{
			SyncOptions: &mgmtv1alpha1.ActivityOptions{
				RetryPolicy: &mgmtv1alpha1.RetryPolicy{
					MaximumAttempts: shared.Ptr(int32(2)),
				},
			},
		}, expected: &workflow.ActivityOptions{StartToCloseTimeout: defaultOpts.StartToCloseTimeout, RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 2}, HeartbeatTimeout: 1 * time.Minute}},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", t.Name(), test.name), func(t *testing.T) {
			output := getSyncActivityOptionsFromJob(test.input)
			assert.Equal(t, test.expected, output)
		})
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

func Test_zeroToNilPointer(t *testing.T) {
	val := uint64(1)

	resp := zeroToNilPointer(uint64(1))
	require.NotNil(t, resp)
	require.Equal(t, val, *resp)

	require.Nil(t, zeroToNilPointer(uint64(0)))
	require.Nil(t, zeroToNilPointer(int64(-1)))
}
