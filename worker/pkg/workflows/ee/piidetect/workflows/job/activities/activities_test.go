package piidetect_job_activities

import (
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tmprl "go.temporal.io/sdk/client"
	tmprl_mocks "go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

func Test_GetPiiDetectJobDetails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	jobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	connClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	connBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	tmprlScheduleClient := tmprl_mocks.NewScheduleClient(t)

	activities := New(jobClient, connClient, connBuilder, tmprlScheduleClient)
	env.RegisterActivity(activities.GetPiiDetectJobDetails)

	t.Run("successfully gets pii detect job details", func(t *testing.T) {
		jobId := "test-job-id"
		accountId := "test-account-id"
		sourceConnId := "test-conn-id"
		piiDetectConfig := &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{}

		jobClient.EXPECT().GetJob(mock.Anything, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		})).Return(
			connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
				Job: &mgmtv1alpha1.Job{
					AccountId: accountId,
					Source: &mgmtv1alpha1.JobSource{
						Options: &mgmtv1alpha1.JobSourceOptions{
							Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
								Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
									ConnectionId: sourceConnId,
								},
							},
						},
					},
					JobType: &mgmtv1alpha1.JobTypeConfig{
						JobType: &mgmtv1alpha1.JobTypeConfig_PiiDetect{
							PiiDetect: piiDetectConfig,
						},
					},
				},
			}),
			nil,
		)

		val, err := env.ExecuteActivity(activities.GetPiiDetectJobDetails, &GetPiiDetectJobDetailsRequest{
			JobId: jobId,
		})
		require.NoError(t, err)
		resp := &GetPiiDetectJobDetailsResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.Equal(t, accountId, resp.AccountId)
		require.Equal(t, sourceConnId, resp.SourceConnectionId)
		require.Equal(t, piiDetectConfig, resp.PiiDetectConfig)
		jobClient.AssertExpectations(t)
	})
}

func Test_GetTablesToPiiScan(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	jobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	connClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	connBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	dataConn := connectiondata.NewMockConnectionDataService(t)
	tmprlScheduleClient := tmprl_mocks.NewScheduleClient(t)
	activities := New(jobClient, connClient, connBuilder, tmprlScheduleClient)
	env.RegisterActivity(activities.GetTablesToPiiScan)

	t.Run("successfully gets tables with include filter", func(t *testing.T) {
		connId := "test-conn-id"
		dbColumnSchemas := []*mgmtv1alpha1.DatabaseColumn{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table1", Column: "col2"},
			{Schema: "schema2", Table: "table2", Column: "col1"},
		}

		connClient.EXPECT().GetConnection(mock.Anything, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: connId,
		})).Return(
			connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
				Connection: &mgmtv1alpha1.Connection{},
			}),
			nil,
		)

		connBuilder.EXPECT().NewDataConnection(mock.Anything, mock.Anything).Return(dataConn, nil)
		dataConn.EXPECT().GetSchema(mock.Anything, &mgmtv1alpha1.ConnectionSchemaConfig{}).Return(dbColumnSchemas, nil)

		filter := &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter{
			Mode: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_Include{
				Include: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TablePatterns{
					Schemas: []string{"schema1"},
				},
			},
		}

		val, err := env.ExecuteActivity(activities.GetTablesToPiiScan, &GetTablesToPiiScanRequest{
			SourceConnectionId: connId,
			Filter:             filter,
		})

		require.NoError(t, err)
		resp := &GetTablesToPiiScanResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.Len(t, resp.Tables, 1)
		require.Equal(t, "schema1", resp.Tables[0].Schema)
		require.Equal(t, "table1", resp.Tables[0].Table)
		// Verify fingerprint matches expected value for schema1.table1 with columns col1,col2
		expectedFingerprint := getTableColumnFingerprint("schema1", "table1", []string{"col1", "col2"})
		require.Equal(t, expectedFingerprint, resp.Tables[0].Fingerprint)

		connClient.AssertExpectations(t)
		connBuilder.AssertExpectations(t)
		dataConn.AssertExpectations(t)
	})
}

func Test_SaveJobPiiDetectReport(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	jobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	connClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	connBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	tmprlScheduleClient := tmprl_mocks.NewScheduleClient(t)

	activities := New(jobClient, connClient, connBuilder, tmprlScheduleClient)
	env.RegisterActivity(activities.SaveJobPiiDetectReport)

	t.Run("successfully saves pii detect report", func(t *testing.T) {
		accountId := "test-account-id"
		jobId := "test-job-id"
		report := &JobPiiDetectReport{
			SuccessfulTableReports: []*TableReport{
				{
					TableSchema: "schema1",
					TableName:   "table1",
					ReportKey:   &mgmtv1alpha1.RunContextKey{AccountId: accountId, JobRunId: "run1"},
				},
			},
		}

		jobClient.EXPECT().SetRunContext(mock.Anything, mock.Anything).Return(
			connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}),
			nil,
		)

		val, err := env.ExecuteActivity(activities.SaveJobPiiDetectReport, &SaveJobPiiDetectReportRequest{
			AccountId: accountId,
			JobId:     jobId,
			Report:    report,
		})

		require.NoError(t, err)
		resp := &SaveJobPiiDetectReportResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.NotNil(t, resp.Key)
		require.Equal(t, accountId, resp.Key.AccountId)
		require.Equal(t, "default-test-workflow-id", resp.Key.JobRunId)
		require.Equal(t, "test-job-id--job-pii-report", resp.Key.ExternalId)

		jobClient.AssertExpectations(t)
	})
}

func Test_GetLastSuccessfulWorkflowId(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	jobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	connClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	connBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	tmprlScheduleClient := tmprl_mocks.NewScheduleClient(t)
	scheduleHandle := tmprl_mocks.NewScheduleHandle(t)

	activities := New(jobClient, connClient, connBuilder, tmprlScheduleClient)
	env.RegisterActivity(activities.GetLastSuccessfulWorkflowId)

	t.Run("successfully gets last successful workflow id", func(t *testing.T) {
		accountId := "test-account-id"
		jobId := "test-job-id"
		workflowId1 := "workflow-1"
		workflowId2 := "workflow-2"

		// Mock schedule handle to return workflow IDs
		tmprlScheduleClient.On("GetHandle", mock.Anything, jobId).Return(scheduleHandle).Once()
		scheduleHandle.On("Describe", mock.Anything).Return(&tmprl.ScheduleDescription{
			Info: tmprl.ScheduleInfo{
				RecentActions: []tmprl.ScheduleActionResult{
					{
						StartWorkflowResult: &tmprl.ScheduleWorkflowExecution{
							WorkflowID: workflowId1,
						},
						ActualTime: time.Now(),
					},
					{
						StartWorkflowResult: &tmprl.ScheduleWorkflowExecution{
							WorkflowID: workflowId2,
						},
						ActualTime: time.Now().Add(-1 * time.Hour),
					},
				},
			},
		}, nil).Once()

		// Mock successful job report for first workflow
		jobClient.EXPECT().GetRunContext(mock.Anything, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				AccountId:  accountId,
				JobRunId:   workflowId1,
				ExternalId: "test-job-id--job-pii-report",
			},
		})).Return(connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
			Value: []byte(`{"successfulTableReports":[{"tableSchema":"schema1","tableName":"table1"}]}`),
		}), nil).Once()

		val, err := env.ExecuteActivity(activities.GetLastSuccessfulWorkflowId, &GetLastSuccessfulWorkflowIdRequest{
			AccountId: accountId,
			JobId:     jobId,
		})

		require.NoError(t, err)
		resp := &GetLastSuccessfulWorkflowIdResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.NotNil(t, resp.WorkflowId)
		require.Equal(t, workflowId1, *resp.WorkflowId)
	})

	t.Run("returns nil when no successful runs found", func(t *testing.T) {
		accountId := "test-account-id"
		jobId := "test-job-id"
		workflowId := "workflow-1"

		// Mock schedule handle to return workflow IDs
		tmprlScheduleClient.On("GetHandle", mock.Anything, jobId).Return(scheduleHandle).Once()
		scheduleHandle.On("Describe", mock.Anything).Return(&tmprl.ScheduleDescription{
			Info: tmprl.ScheduleInfo{
				RecentActions: []tmprl.ScheduleActionResult{
					{
						StartWorkflowResult: &tmprl.ScheduleWorkflowExecution{
							WorkflowID: workflowId,
						},
						ActualTime: time.Now(),
					},
				},
			},
		}, nil).Once()

		// Mock job report with empty value to simulate no successful tables
		jobClient.EXPECT().GetRunContext(mock.Anything, connect.NewRequest(&mgmtv1alpha1.GetRunContextRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				AccountId:  accountId,
				JobRunId:   workflowId,
				ExternalId: "test-job-id--job-pii-report",
			},
		})).Return(connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
			Value: []byte(`{"successfulTableReports":[]}`),
		}), nil).Once()

		val, err := env.ExecuteActivity(activities.GetLastSuccessfulWorkflowId, &GetLastSuccessfulWorkflowIdRequest{
			AccountId: accountId,
			JobId:     jobId,
		})

		require.NoError(t, err)
		resp := &GetLastSuccessfulWorkflowIdResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.Nil(t, resp.WorkflowId)
	})

	t.Run("handles error when getting recent runs", func(t *testing.T) {
		accountId := "test-account-id"
		jobId := "test-job-id"

		// Mock schedule handle to return error
		tmprlScheduleClient.On("GetHandle", mock.Anything, jobId).Return(scheduleHandle).Once()
		scheduleHandle.On("Describe", mock.Anything).Return(nil, errors.New("schedule error")).Once()

		val, err := env.ExecuteActivity(activities.GetLastSuccessfulWorkflowId, &GetLastSuccessfulWorkflowIdRequest{
			AccountId: accountId,
			JobId:     jobId,
		})

		require.NoError(t, err)
		resp := &GetLastSuccessfulWorkflowIdResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.Nil(t, resp.WorkflowId)
	})

	t.Run("handles error when getting job report", func(t *testing.T) {
		accountId := "test-account-id"
		jobId := "test-job-id"
		workflowId := "workflow-1"

		// Mock schedule handle to return workflow IDs
		tmprlScheduleClient.On("GetHandle", mock.Anything, jobId).Return(scheduleHandle).Once()
		scheduleHandle.On("Describe", mock.Anything).Return(&tmprl.ScheduleDescription{
			Info: tmprl.ScheduleInfo{
				RecentActions: []tmprl.ScheduleActionResult{
					{
						StartWorkflowResult: &tmprl.ScheduleWorkflowExecution{
							WorkflowID: workflowId,
						},
						ActualTime: time.Now(),
					},
				},
			},
		}, nil).Once()

		// Mock job report to return error
		jobClient.EXPECT().GetRunContext(mock.Anything, mock.Anything).Return(nil, errors.New("report error")).Once()

		val, err := env.ExecuteActivity(activities.GetLastSuccessfulWorkflowId, &GetLastSuccessfulWorkflowIdRequest{
			AccountId: accountId,
			JobId:     jobId,
		})

		require.NoError(t, err)
		resp := &GetLastSuccessfulWorkflowIdResponse{}
		err = val.Get(resp)
		require.NoError(t, err)
		require.Nil(t, resp.WorkflowId)
	})
}
