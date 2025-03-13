package piidetect_job_activities

import (
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_GetPiiDetectJobDetails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	jobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	connClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	connBuilder := connectiondata.NewMockConnectionDataBuilder(t)

	activities := New(jobClient, connClient, connBuilder)
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

	activities := New(jobClient, connClient, connBuilder)
	env.RegisterActivity(activities.GetTablesToPiiScan)

	t.Run("successfully gets tables with include filter", func(t *testing.T) {
		connId := "test-conn-id"
		tables := []connectiondata.TableIdentifier{
			{Schema: "schema1", Table: "table1"},
			{Schema: "schema2", Table: "table2"},
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
		dataConn.EXPECT().GetAllTables(mock.Anything).Return(tables, nil)

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

	activities := New(jobClient, connClient, connBuilder)
	env.RegisterActivity(activities.SaveJobPiiDetectReport)

	t.Run("successfully saves pii detect report", func(t *testing.T) {
		accountId := "test-account-id"
		jobId := "test-job-id"
		report := &JobPiiDetectReport{
			SuccessfulTableKeys: []*mgmtv1alpha1.RunContextKey{
				{AccountId: accountId, JobRunId: "run1"},
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
