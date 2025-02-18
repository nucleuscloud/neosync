package tablesync_workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/activities/sync"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// pointerToString returns a pointer to the given string.
func pointerToString(s string) *string {
	return &s
}

// Test_TableSync_SingleIteration verifies that when the activity returns a nil continuation token immediately,
// the workflow completes in a single iteration.
func Test_TableSync_SingleIteration(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	// Register the workflow.
	tsWf := New()
	env.RegisterWorkflow(tsWf.TableSync)

	// Register a fake activity implementation.
	// This activity returns a response with a nil continuation token immediately.
	var syncActivity *sync_activity.Activity
	env.OnActivity(syncActivity.SyncTable, mock.Anything, mock.Anything, mock.Anything).
		Return(&sync_activity.SyncTableResponse{
			ContinuationToken: nil,
		}, nil)

	// Set activity options to be used by the workflow.
	options := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	request := &TableSyncRequest{
		AccountId:           "account1",
		Id:                  "id1",
		JobRunId:            "jobrun1",
		ContinuationToken:   nil,
		SyncActivityOptions: &options,
		TableSchema:         "schema1",
		TableName:           "table1",
	}

	env.ExecuteWorkflow((&Workflow{}).TableSync, request)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result *TableSyncResponse
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)
	require.Equal(t, "schema1", result.Schema)
	require.Equal(t, "table1", result.Table)
}

// Test_TableSync_MultipleIterations simulates the case where the activity returns a non-nil continuation token
// on the first call and nil on the second call, causing one iteration loop.
func Test_TableSync_MultipleIterations(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	tsWf := New()
	env.RegisterWorkflow(tsWf.TableSync)

	// Use a counter so that the first activity call returns a token and the second returns nil.
	callCount := 0
	var syncActivity *sync_activity.Activity
	env.OnActivity(syncActivity.SyncTable, mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncTableRequest, meta *sync_activity.SyncMetadata) (*sync_activity.SyncTableResponse, error) {
			callCount++
			if callCount == 1 {
				// Return a non-nil token.
				return &sync_activity.SyncTableResponse{
					ContinuationToken: pointerToString("token1"),
				}, nil
			}
			// Second call returns nil token to finish the loop.
			return &sync_activity.SyncTableResponse{
				ContinuationToken: nil,
			}, nil
		})

	options := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	request := &TableSyncRequest{
		AccountId:           "account2",
		Id:                  "id2",
		JobRunId:            "jobrun2",
		ContinuationToken:   nil,
		SyncActivityOptions: &options,
		TableSchema:         "schema2",
		TableName:           "table2",
	}

	env.ExecuteWorkflow((&Workflow{}).TableSync, request)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result *TableSyncResponse
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)
	require.Equal(t, "schema2", result.Schema)
	require.Equal(t, "table2", result.Table)
	require.Equal(t, 2, callCount)
}

// Test_TableSync_ContinueAsNew verifies that when the activity always returns a non-nil continuation token,
// after MAX_ITERATIONS the workflow issues a ContinueAsNew error.
func Test_TableSync_ContinueAsNew(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	tsWf := New()
	env.RegisterWorkflow(tsWf.TableSync)

	// The activity always returns a non-nil token.
	var syncActivity *sync_activity.Activity
	env.OnActivity(syncActivity.SyncTable, mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *sync_activity.SyncTableRequest, meta *sync_activity.SyncMetadata) (*sync_activity.SyncTableResponse, error) {
			return &sync_activity.SyncTableResponse{
				ContinuationToken: pointerToString("loop"),
			}, nil
		})

	options := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	request := &TableSyncRequest{
		AccountId:           "account3",
		Id:                  "id3",
		JobRunId:            "jobrun3",
		ContinuationToken:   nil,
		SyncActivityOptions: &options,
		TableSchema:         "schema3",
		TableName:           "table3",
	}

	env.ExecuteWorkflow((&Workflow{}).TableSync, request)
	// Since the activity always returns a token, after MAX_ITERATIONS the workflow should not complete normally.
	err := env.GetWorkflowError()
	require.Error(t, err)

	// Verify that the error is a ContinueAsNewError.
	var continueErr *workflow.ContinueAsNewError
	require.True(t, errors.As(err, &continueErr))
}
