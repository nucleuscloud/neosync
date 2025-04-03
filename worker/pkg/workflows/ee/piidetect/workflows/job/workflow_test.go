package piidetect_job_workflow

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	accounthook_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow"
	piidetect_job_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/job/activities"
	piidetect_table_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

// pointerToString returns a pointer to the given string.
func pointerToString(s string) *string {
	return &s
}

func Test_JobPiiDetect(t *testing.T) {
	t.Run("successful_workflow_with_multiple_tables", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		// Register workflow
		wf := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
		env.RegisterWorkflow(wf.JobPiiDetect)

		// Register child workflow
		tableWf := piidetect_table_workflow.New()
		env.RegisterWorkflow(tableWf.TablePiiDetect)

		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

		var activities *piidetect_job_activities.Activities

		// Setup GetPiiDetectJobDetails activity expectations
		env.OnActivity(activities.GetPiiDetectJobDetails, mock.Anything, &piidetect_job_activities.GetPiiDetectJobDetailsRequest{
			JobId: "job-123",
		}).Return(&piidetect_job_activities.GetPiiDetectJobDetailsResponse{
			AccountId:          "acc-123",
			SourceConnectionId: "conn-123",
			PiiDetectConfig: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{
				DataSampling: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_DataSampling{
					IsEnabled: true,
				},
				UserPrompt: pointerToString("Please detect PII"),
			},
		}, nil)

		// Setup GetTablesToPiiScan activity expectations
		env.OnActivity(activities.GetTablesToPiiScan, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetTablesToPiiScanResponse{
			Tables: []piidetect_job_activities.TableIdentifierWithFingerprint{
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "users"}, Fingerprint: "fingerprint1"},
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "orders"}, Fingerprint: "fingerprint2"},
			},
		}, nil)

		// Setup child workflow expectations for both tables
		usersKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.users--table-pii-report",
		}
		ordersKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.orders--table-pii-report",
		}

		env.OnWorkflow(tableWf.TablePiiDetect, mock.Anything, mock.Anything).Return(
			func(ctx any, req *piidetect_table_workflow.TablePiiDetectRequest) (*piidetect_table_workflow.TablePiiDetectResponse, error) {
				if req.TableName == "users" {
					return &piidetect_table_workflow.TablePiiDetectResponse{ResultKey: usersKey}, nil
				}
				return &piidetect_table_workflow.TablePiiDetectResponse{ResultKey: ordersKey}, nil
			})

		// Setup SaveJobPiiDetectReport activity expectations
		expectedJobKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "job-pii-report",
		}
		env.OnActivity(activities.SaveJobPiiDetectReport, mock.Anything, mock.Anything, mock.Anything).Return(&piidetect_job_activities.SaveJobPiiDetectReportResponse{
			Key: expectedJobKey,
		}, nil)

		// Execute workflow
		req := &PiiDetectRequest{
			JobId: "job-123",
		}

		var result *PiiDetectResponse
		env.ExecuteWorkflow(wf.JobPiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedJobKey, result.ReportKey)
	})

	t.Run("workflow_with_table_filter", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
		env.RegisterWorkflow(wf.JobPiiDetect)
		tableWf := piidetect_table_workflow.New()
		env.RegisterWorkflow(tableWf.TablePiiDetect)

		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

		var activities *piidetect_job_activities.Activities

		// Setup GetPiiDetectJobDetails with table filter
		env.OnActivity(activities.GetPiiDetectJobDetails, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetPiiDetectJobDetailsResponse{
			AccountId:          "acc-123",
			SourceConnectionId: "conn-123",
			PiiDetectConfig: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{
				TableScanFilter: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter{
					Mode: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_Include{
						Include: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TablePatterns{
							Schemas: []string{"public"},
							Tables: []*mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableIdentifier{
								{Schema: "public", Table: "users"},
							},
						},
					},
				},
			},
		}, nil)

		// Setup GetTablesToPiiScan to return filtered tables
		env.OnActivity(activities.GetTablesToPiiScan, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetTablesToPiiScanResponse{
			Tables: []piidetect_job_activities.TableIdentifierWithFingerprint{
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "users"}, Fingerprint: "fingerprint1"},
			},
		}, nil)

		// Setup child workflow expectations
		usersKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.users--table-pii-report",
		}
		env.OnWorkflow(tableWf.TablePiiDetect, mock.Anything, mock.Anything).Return(&piidetect_table_workflow.TablePiiDetectResponse{
			ResultKey: usersKey,
		}, nil)

		expectedJobKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "job-pii-report",
		}
		env.OnActivity(activities.SaveJobPiiDetectReport, mock.Anything, mock.Anything, mock.Anything).Return(&piidetect_job_activities.SaveJobPiiDetectReportResponse{
			Key: expectedJobKey,
		}, nil)

		req := &PiiDetectRequest{
			JobId: "job-123",
		}

		var result *PiiDetectResponse
		env.ExecuteWorkflow(wf.JobPiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedJobKey, result.ReportKey)
	})

	t.Run("workflow_fails_when_get_job_details_fails", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
		env.RegisterWorkflow(wf.JobPiiDetect)

		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Never()

		var activities *piidetect_job_activities.Activities

		// Setup GetPiiDetectJobDetails to fail
		env.OnActivity(activities.GetPiiDetectJobDetails, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		req := &PiiDetectRequest{
			JobId: "job-123",
		}

		env.ExecuteWorkflow(wf.JobPiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		workflowErr := env.GetWorkflowError()
		require.Error(t, workflowErr)
		assert.Contains(t, workflowErr.Error(), assert.AnError.Error())
	})

	t.Run("workflow_handles_child_workflow_failure", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
		env.RegisterWorkflow(wf.JobPiiDetect)
		tableWf := piidetect_table_workflow.New()
		env.RegisterWorkflow(tableWf.TablePiiDetect)

		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

		var activities *piidetect_job_activities.Activities

		// Setup GetPiiDetectJobDetails
		env.OnActivity(activities.GetPiiDetectJobDetails, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetPiiDetectJobDetailsResponse{
			AccountId:          "acc-123",
			SourceConnectionId: "conn-123",
		}, nil)

		// Setup GetTablesToPiiScan
		env.OnActivity(activities.GetTablesToPiiScan, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetTablesToPiiScanResponse{
			Tables: []piidetect_job_activities.TableIdentifierWithFingerprint{
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "users"}, Fingerprint: "fingerprint1"},
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "orders"}, Fingerprint: "fingerprint2"},
			},
		}, nil)

		// Setup child workflow to fail for one table but succeed for another
		usersKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.users--table-pii-report",
		}
		env.OnWorkflow(tableWf.TablePiiDetect, mock.Anything, mock.Anything).Return(
			func(ctx any, req *piidetect_table_workflow.TablePiiDetectRequest) (*piidetect_table_workflow.TablePiiDetectResponse, error) {
				if req.TableName == "users" {
					return &piidetect_table_workflow.TablePiiDetectResponse{ResultKey: usersKey}, nil
				}
				return nil, temporal.NewApplicationError("table scan failed", "ScanError")
			})

		// Setup SaveJobPiiDetectReport - should still save report with successful table
		expectedJobKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "job-pii-report",
		}
		env.OnActivity(activities.SaveJobPiiDetectReport, mock.Anything, mock.Anything, mock.Anything).Return(&piidetect_job_activities.SaveJobPiiDetectReportResponse{
			Key: expectedJobKey,
		}, nil)

		req := &PiiDetectRequest{
			JobId: "job-123",
		}

		var result *PiiDetectResponse
		env.ExecuteWorkflow(wf.JobPiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedJobKey, result.ReportKey)
	})

	t.Run("workflow_with_incremental_sync", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
		env.RegisterWorkflow(wf.JobPiiDetect)
		tableWf := piidetect_table_workflow.New()
		env.RegisterWorkflow(tableWf.TablePiiDetect)

		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

		var activities *piidetect_job_activities.Activities

		// Setup GetPiiDetectJobDetails with incremental config enabled
		env.OnActivity(activities.GetPiiDetectJobDetails, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetPiiDetectJobDetailsResponse{
			AccountId:          "acc-123",
			SourceConnectionId: "conn-123",
			PiiDetectConfig: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{
				Incremental: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_Incremental{
					IsEnabled: true,
				},
			},
		}, nil)

		// Setup GetLastSuccessfulWorkflowId
		lastWorkflowId := "previous-workflow-123"
		env.OnActivity(activities.GetLastSuccessfulWorkflowId, mock.Anything, &piidetect_job_activities.GetLastSuccessfulWorkflowIdRequest{
			AccountId: "acc-123",
			JobId:     "job-123",
		}).Return(&piidetect_job_activities.GetLastSuccessfulWorkflowIdResponse{
			WorkflowId: &lastWorkflowId,
		}, nil)

		// Setup GetTablesToPiiScan with previous reports and changed tables
		previousReport := &piidetect_job_activities.TableReport{
			TableSchema:     "public",
			TableName:       "unchanged_table",
			ScanFingerprint: "old-fingerprint",
			ReportKey: &mgmtv1alpha1.RunContextKey{
				AccountId:  "acc-123",
				JobRunId:   "job-123",
				ExternalId: "public.unchanged_table--table-pii-report",
			},
		}

		env.OnActivity(activities.GetTablesToPiiScan, mock.Anything, &piidetect_job_activities.GetTablesToPiiScanRequest{
			AccountId:          "acc-123",
			JobId:              "job-123",
			SourceConnectionId: "conn-123",
			IncrementalConfig: &piidetect_job_activities.GetIncrementalTablesConfig{
				LastWorkflowId: lastWorkflowId,
			},
		}).Return(&piidetect_job_activities.GetTablesToPiiScanResponse{
			Tables: []piidetect_job_activities.TableIdentifierWithFingerprint{
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "changed_table"}, Fingerprint: "new-fingerprint"},
			},
			PreviousReports: []*piidetect_job_activities.TableReport{previousReport},
		}, nil)

		// Setup child workflow expectations for changed table
		changedTableKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.changed_table--table-pii-report",
		}
		env.OnWorkflow(tableWf.TablePiiDetect, mock.Anything, mock.Anything).Return(&piidetect_table_workflow.TablePiiDetectResponse{
			ResultKey: changedTableKey,
		}, nil)

		// Setup SaveJobPiiDetectReport - should include both previous and new reports
		expectedJobKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "job-pii-report",
		}
		env.OnActivity(activities.SaveJobPiiDetectReport, mock.Anything, mock.MatchedBy(func(req *piidetect_job_activities.SaveJobPiiDetectReportRequest) bool {
			// Verify that both the unchanged and changed table reports are included
			if len(req.Report.SuccessfulTableReports) != 2 {
				return false
			}
			hasUnchanged := false
			hasChanged := false
			for _, report := range req.Report.SuccessfulTableReports {
				if report.TableName == "unchanged_table" {
					hasUnchanged = true
				}
				if report.TableName == "changed_table" {
					hasChanged = true
				}
			}
			return hasUnchanged && hasChanged
		})).Return(&piidetect_job_activities.SaveJobPiiDetectReportResponse{
			Key: expectedJobKey,
		}, nil)

		req := &PiiDetectRequest{
			JobId: "job-123",
		}

		var result *PiiDetectResponse
		env.ExecuteWorkflow(wf.JobPiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedJobKey, result.ReportKey)
	})

	t.Run("workflow_with_incremental_sync_updates_changed_table", func(t *testing.T) {
		var ts testsuite.WorkflowTestSuite
		env := ts.NewTestWorkflowEnvironment()

		wf := New(testutil.NewFakeEELicense(testutil.WithIsValid()))
		env.RegisterWorkflow(wf.JobPiiDetect)
		tableWf := piidetect_table_workflow.New()
		env.RegisterWorkflow(tableWf.TablePiiDetect)

		env.OnWorkflow(accounthook_workflow.ProcessAccountHook, mock.Anything, mock.Anything).
			Return(&accounthook_workflow.ProcessAccountHookResponse{}, nil).Twice()

		var activities *piidetect_job_activities.Activities

		// Setup GetPiiDetectJobDetails with incremental config enabled
		env.OnActivity(activities.GetPiiDetectJobDetails, mock.Anything, mock.Anything).Return(&piidetect_job_activities.GetPiiDetectJobDetailsResponse{
			AccountId:          "acc-123",
			SourceConnectionId: "conn-123",
			PiiDetectConfig: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{
				Incremental: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_Incremental{
					IsEnabled: true,
				},
			},
		}, nil)

		// Setup GetLastSuccessfulWorkflowId
		lastWorkflowId := "previous-workflow-123"
		env.OnActivity(activities.GetLastSuccessfulWorkflowId, mock.Anything, &piidetect_job_activities.GetLastSuccessfulWorkflowIdRequest{
			AccountId: "acc-123",
			JobId:     "job-123",
		}).Return(&piidetect_job_activities.GetLastSuccessfulWorkflowIdResponse{
			WorkflowId: &lastWorkflowId,
		}, nil)

		// Setup previous report for a table that will be detected as changed
		previousReport := &piidetect_job_activities.TableReport{
			TableSchema:     "public",
			TableName:       "users",
			ScanFingerprint: "old-fingerprint",
			ReportKey: &mgmtv1alpha1.RunContextKey{
				AccountId:  "acc-123",
				JobRunId:   "previous-job",
				ExternalId: "public.users--table-pii-report-old",
			},
		}

		// Setup GetTablesToPiiScan to return the same table with a new fingerprint
		env.OnActivity(activities.GetTablesToPiiScan, mock.Anything, &piidetect_job_activities.GetTablesToPiiScanRequest{
			AccountId:          "acc-123",
			JobId:              "job-123",
			SourceConnectionId: "conn-123",
			IncrementalConfig: &piidetect_job_activities.GetIncrementalTablesConfig{
				LastWorkflowId: lastWorkflowId,
			},
		}).Return(&piidetect_job_activities.GetTablesToPiiScanResponse{
			Tables: []piidetect_job_activities.TableIdentifierWithFingerprint{
				{TableIdentifier: piidetect_job_activities.TableIdentifier{Schema: "public", Table: "users"}, Fingerprint: "new-fingerprint"},
			},
			PreviousReports: []*piidetect_job_activities.TableReport{previousReport},
		}, nil)

		// Setup child workflow expectations for the changed table
		newTableKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "public.users--table-pii-report-new",
		}
		env.OnWorkflow(tableWf.TablePiiDetect, mock.Anything, mock.MatchedBy(func(req *piidetect_table_workflow.TablePiiDetectRequest) bool {
			// Verify that the previous results key is passed to the table workflow
			return req.TableName == "users" &&
				req.PreviousResultsKey != nil &&
				req.PreviousResultsKey.ExternalId == "public.users--table-pii-report-old"
		})).Return(&piidetect_table_workflow.TablePiiDetectResponse{
			ResultKey: newTableKey,
		}, nil)

		// Setup SaveJobPiiDetectReport - should contain only the new report
		expectedJobKey := &mgmtv1alpha1.RunContextKey{
			AccountId:  "acc-123",
			JobRunId:   "job-123",
			ExternalId: "job-pii-report",
		}
		env.OnActivity(activities.SaveJobPiiDetectReport, mock.Anything, mock.MatchedBy(func(req *piidetect_job_activities.SaveJobPiiDetectReportRequest) bool {
			if len(req.Report.SuccessfulTableReports) != 1 {
				return false
			}
			report := req.Report.SuccessfulTableReports[0]
			// Verify that we're using the new report with the new fingerprint
			return report.TableName == "users" &&
				report.ScanFingerprint == "new-fingerprint" &&
				report.ReportKey.ExternalId == "public.users--table-pii-report-new"
		})).Return(&piidetect_job_activities.SaveJobPiiDetectReportResponse{
			Key: expectedJobKey,
		}, nil)

		req := &PiiDetectRequest{
			JobId: "job-123",
		}

		var result *PiiDetectResponse
		env.ExecuteWorkflow(wf.JobPiiDetect, req)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.NotNil(t, result)
		assert.Equal(t, expectedJobKey, result.ReportKey)
	})
}
