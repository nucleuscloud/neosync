package schemainit_workflow

import (
	"testing"
	"time"

	initschema_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/schemainit/activities/init-schema"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_SchemaInit_Workflow(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	tsWf := New()
	env.RegisterWorkflow(tsWf.SchemaInit)

	var initSchemaActivity *initschema_activity.Activity
	env.OnActivity(initSchemaActivity.RunSqlInitTableStatements, mock.Anything, mock.Anything).
		Return(&initschema_activity.RunSqlInitTableStatementsResponse{}, nil)

	options := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	request := &SchemaInitRequest{
		AccountId:                 "account1",
		JobId:                     "job1",
		JobRunId:                  "jobrun1",
		SchemaInitActivityOptions: &options,
	}

	env.ExecuteWorkflow(tsWf.SchemaInit, request)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
