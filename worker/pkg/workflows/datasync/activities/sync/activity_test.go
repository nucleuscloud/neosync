package sync_activity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.temporal.io/sdk/testsuite"
)

func Test_Sync_RunContext_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := NewBenthosStreamManager()

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
			}
			return nil, errors.New("invalid test account id")
		},
	))
	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)

	activity := New(nil, jobclient, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)

	env.RegisterActivity(activity.Sync)

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		AccountId: accountId,
		Name:      "test",
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Run_No_BenthosConfig(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := NewBenthosStreamManager()

	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)

	env.RegisterActivity(activity.Sync)

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err)
	require.Nil(t, val)
}

func Test_Sync_Run_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)

	env.RegisterActivity(activity.Sync)

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(`
input:
  generate:
    count: 1
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  stdout:
    codec: lines
`),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Run_Metrics_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	meterProvider := metricsdk.NewMeterProvider()
	meter := meterProvider.Meter("test")
	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, meter, benthosStreamManager, true)

	env.RegisterActivity(activity.Sync)

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(`
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
`),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Fake_Mutation_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)
	env.RegisterActivity(activity.Sync)

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(`
input:
  generate:
    count: 1
    interval: ""
    mapping: 'root = { "name": "nick" }'
pipeline:
  threads: 1
  processors:
    - mutation: |
        root.name = fake("first_name")
output:
  label: ""
  stdout:
    codec: lines
`),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Run_Success_Javascript(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)
	env.RegisterActivity(activity.Sync)

	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(fmt.Sprintf(`
input:
  generate:
    mapping: root = {"name":"evis"}
    interval: 1s
    count: 1
pipeline:
  processors:
    - neosync_javascript:
        code: |
          (() => {
          function fn_name(value, input){
          var a = value + "test";
          return a };
          const input = benthos.v0_msg_as_structured();
          const output = { ...input };
          output["name"] = fn_name(input["name"], input);
          benthos.v0_msg_set_structured(output);
          })();
output:
  label: ""
  file:
    path:  %s
    codec: lines
`, tmpFile.Name())),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	assert.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	assert.NoError(t, err)

	stdoutBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read from temp file: %v", err)
	}
	stringResult := string(stdoutBytes)

	returnValue := strings.TrimSpace(stringResult) // remove new line at the end of the stdout line

	assert.Equal(t, `{"name":"evistest"}`, returnValue)
}

func Test_Sync_Run_Success_MutataionAndJavascript(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)
	env.RegisterActivity(activity.Sync)

	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	val, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(fmt.Sprintf(`
input:
  generate:
    mapping: root = {"name":"evis"}
    interval: 1s
    count: 1
pipeline:
  processors:
    - mutation:
        root.name = this.name.reverse()
    - neosync_javascript:
        code: |
          (() => {
          function fn1(value, input){
          var a = value + "test";
          return a };
          const input = benthos.v0_msg_as_structured();
          const output = { ...input };
          output["name"] = fn1(input["name"], input);
          benthos.v0_msg_set_structured(output);
          })();
output:
  label: ""
  file:
    path:  %s
    codec: lines
  `, tmpFile.Name())),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	assert.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	assert.NoError(t, err)

	stdoutBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read from temp file: %v", err)
	}
	stringResult := string(stdoutBytes)

	returnValue := strings.TrimSpace(stringResult) // remove new line at the end of the stdout line

	assert.Equal(t, `{"name":"sivetest"}`, returnValue)
}

func Test_Sync_Run_Processor_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)

	env.RegisterActivity(activity.Sync)

	_, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(`
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
`),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err, "error was nil when it should be present")
}

func Test_Sync_Run_Output_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	mockBenthosStreamManager := NewMockBenthosStreamManagerClient(t)
	mockBenthosStream := NewMockBenthosStreamClient(t)
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, mockBenthosStreamManager, true)

	env.RegisterActivity(activity.Sync)

	mockBenthosStreamManager.On("NewBenthosStreamFromBuilder", mock.Anything).Return(mockBenthosStream, nil)
	errmsg := "duplicate key value violates unique constraint"
	mockBenthosStream.On("Run", mock.Anything).Return(errors.New(errmsg))
	mockBenthosStream.On("StopWithin", mock.Anything).Return(nil).Maybe()

	_, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(`
input:
  generate:
    count: 1000
    interval: ""
    mapping: 'root = { "name": "nick" }'
pipeline:
  threads: 1
  processors: []
output:
  label: ""
  error:
     error_msg: ${! meta("fallback_error")}
`),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err, "error was nil when it should be present")
	require.Contains(t, err.Error(), "activity error")
}

func Test_Sync_Run_ActivityStop_MockBenthos(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	mockBenthosStreamManager := NewMockBenthosStreamManagerClient(t)
	mockBenthosStream := NewMockBenthosStreamClient(t)
	config := strings.TrimSpace(`
input:
  generate:
    count: 10000
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  stdout:
    codec: lines
`)

	mockBenthosStreamManager.On("NewBenthosStreamFromBuilder", mock.Anything).Return(mockBenthosStream, nil)
	mockBenthosStream.On("Run", mock.Anything).After(5 * time.Second).Return(nil)
	mockBenthosStream.On("StopWithin", mock.Anything).Return(nil)

	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, mockBenthosStreamManager, true)
	env.RegisterActivity(activity.Sync)

	stopCh := make(chan struct{})
	env.SetWorkerStopChannel(stopCh)

	go func() {
		time.Sleep(300 * time.Millisecond)
		close(stopCh)
	}()

	_, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: config,
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "received worker stop signal")
	mockBenthosStream.AssertCalled(t, "StopWithin", mock.Anything)
}

func Test_Sync_Run_ActivityWorkerStop(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	benthosStreamManager := NewBenthosStreamManager()
	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, benthosStreamManager, true)

	env.RegisterActivity(activity.Sync)
	stopCh := make(chan struct{})
	env.SetWorkerStopChannel(stopCh)

	go func() {
		// Close the channel to simulate sending a stop signal
		time.Sleep(210 * time.Millisecond)
		close(stopCh)
	}()

	_, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: strings.TrimSpace(`
input:
  generate:
    count: 100000
    interval: ""
    mapping: 'root = { "id": uuid_v4() }'
output:
  label: ""
  drop: {}
`),
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "received worker stop signal")
}

func Test_Sync_Run_BenthosError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	mockBenthosStreamManager := NewMockBenthosStreamManagerClient(t)
	mockBenthosStream := NewMockBenthosStreamClient(t)
	config := strings.TrimSpace(`
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

	mockBenthosStreamManager.On("NewBenthosStreamFromBuilder", mock.Anything).Return(mockBenthosStream, nil)
	errmsg := "benthos error"
	mockBenthosStream.On("Run", mock.Anything).Return(errors.New(errmsg))
	mockBenthosStream.On("StopWithin", mock.Anything).Return(nil).Maybe()

	activity := New(nil, nil, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, nil, nil, mockBenthosStreamManager, true)

	env.RegisterActivity(activity.Sync)
	_, err := env.ExecuteActivity(activity.Sync, &SyncRequest{
		BenthosConfig: config,
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

func Test_syncMapToStringMap(t *testing.T) {
	syncmap := sync.Map{}

	syncmap.Store("foo", "bar")
	syncmap.Store("bar", "baz")
	syncmap.Store(1, "2")
	syncmap.Store("3", 4)

	out := syncMapToStringMap(&syncmap)
	assert.Len(t, out, 2)
	assert.Equal(t, out["foo"], "bar")
	assert.Equal(t, out["bar"], "baz")

	assert.Empty(t, syncMapToStringMap(nil))
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
