package piidetect_table_workflow

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	piidetect_table_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_TablePiiDetect(t *testing.T) {
	t.Run("successful_workflow_with_pii_detected", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		// Register workflow
		wf := New()
		env.RegisterWorkflow(wf.TablePiiDetect)

		// Mock activities
		var activities *piidetect_table_activities.Activities

		// Setup GetColumnData activity expectations
		env.OnActivity(activities.GetColumnData, mock.Anything, &piidetect_table_activities.GetColumnDataRequest{
			ConnectionId: "conn-123",
			TableSchema:  "public",
			TableName:    "users",
		}).Return(&piidetect_table_activities.GetColumnDataResponse{
			ColumnData: []*piidetect_table_activities.ColumnData{
				{
					Column:     "email",
					DataType:   "varchar",
					IsNullable: true,
				},
				{
					Column:     "created_at",
					DataType:   "timestamp",
					IsNullable: false,
				},
			},
		}, nil)

		// Setup DetectPiiRegex activity expectations
		env.OnActivity(activities.DetectPiiRegex, mock.Anything, mock.Anything).Return(&piidetect_table_activities.DetectPiiRegexResponse{
			PiiColumns: map[string]piidetect_table_activities.PiiCategory{
				"email": piidetect_table_activities.PiiCategoryContact,
			},
		}, nil)

		// Setup DetectPiiLLM activity expectations
		env.OnActivity(activities.DetectPiiLLM, mock.Anything, mock.Anything).Return(&piidetect_table_activities.DetectPiiLLMResponse{
			PiiColumns: map[string]piidetect_table_activities.LLMPiiDetectReport{
				"email": {
					Category:   piidetect_table_activities.PiiCategoryContact,
					Confidence: 0.95,
				},
			},
		}, nil)

		// Setup SaveTablePiiDetectReport activity expectations
		expectedKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.users--table-pii-report",
		}
		env.OnActivity(activities.SaveTablePiiDetectReport, mock.Anything, mock.Anything, mock.Anything).Return(&piidetect_table_activities.SaveTablePiiDetectReportResponse{
			Key: expectedKey,
		}, nil)

		// Execute workflow
		req := &TablePiiDetectRequest{
			AccountId:        "acc-123",
			JobId:            "job-123",
			ConnectionId:     "conn-123",
			TableSchema:      "public",
			TableName:        "users",
			ShouldSampleData: true,
			UserPrompt:       "Please detect PII",
		}

		var result *TablePiiDetectResponse
		env.ExecuteWorkflow(wf.TablePiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedKey, result.ResultKey)
		assert.Len(t, result.PiiColumns, 1)

		// Verify the combined report for email
		emailReport, exists := result.PiiColumns["email"]
		require.True(t, exists)
		assert.Equal(t, piidetect_table_activities.PiiCategoryContact, *emailReport.Regex)
		require.NotNil(t, emailReport.LLM)
		assert.Equal(t, piidetect_table_activities.PiiCategoryContact, emailReport.LLM.Category)
		assert.Equal(t, 0.95, emailReport.LLM.Confidence)
	})

	t.Run("workflow_with_no_pii_detected", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New()
		env.RegisterWorkflow(wf.TablePiiDetect)

		var activities *piidetect_table_activities.Activities

		// Setup GetColumnData activity expectations
		env.OnActivity(activities.GetColumnData, mock.Anything, mock.Anything).Return(&piidetect_table_activities.GetColumnDataResponse{
			ColumnData: []*piidetect_table_activities.ColumnData{
				{
					Column:     "id",
					DataType:   "uuid",
					IsNullable: false,
				},
				{
					Column:     "created_at",
					DataType:   "timestamp",
					IsNullable: false,
				},
			},
		}, nil)

		// Setup activities to return no PII
		env.OnActivity(activities.DetectPiiRegex, mock.Anything, mock.Anything).Return(&piidetect_table_activities.DetectPiiRegexResponse{
			PiiColumns: map[string]piidetect_table_activities.PiiCategory{},
		}, nil)

		env.OnActivity(activities.DetectPiiLLM, mock.Anything, mock.Anything).Return(&piidetect_table_activities.DetectPiiLLMResponse{
			PiiColumns: map[string]piidetect_table_activities.LLMPiiDetectReport{},
		}, nil)

		expectedKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.users--table-pii-report",
		}
		env.OnActivity(activities.SaveTablePiiDetectReport, mock.Anything, mock.Anything, mock.Anything).Return(&piidetect_table_activities.SaveTablePiiDetectReportResponse{
			Key: expectedKey,
		}, nil)

		req := &TablePiiDetectRequest{
			AccountId:        "acc-123",
			JobId:            "job-123",
			ConnectionId:     "conn-123",
			TableSchema:      "public",
			TableName:        "users",
			ShouldSampleData: false,
		}

		var result *TablePiiDetectResponse
		env.ExecuteWorkflow(wf.TablePiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedKey, result.ResultKey)
		assert.Empty(t, result.PiiColumns)
	})

	t.Run("workflow_fails_when_get_column_data_fails", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New()
		env.RegisterWorkflow(wf.TablePiiDetect)

		var activities *piidetect_table_activities.Activities

		// Setup GetColumnData to fail
		env.OnActivity(activities.GetColumnData, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		req := &TablePiiDetectRequest{
			AccountId:        "acc-123",
			JobId:            "job-123",
			ConnectionId:     "conn-123",
			TableSchema:      "public",
			TableName:        "users",
			ShouldSampleData: false,
		}

		env.ExecuteWorkflow(wf.TablePiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		workflowErr := env.GetWorkflowError()
		require.Error(t, workflowErr)
		assert.Contains(t, workflowErr.Error(), assert.AnError.Error())
	})
}
