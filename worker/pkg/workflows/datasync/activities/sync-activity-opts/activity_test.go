package syncactivityopts_activity

import (
	"fmt"
	"testing"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

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
