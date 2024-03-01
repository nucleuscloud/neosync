package sync_activity

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Sync_Run_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activity := New(nil, &sync.Map{}, nil)

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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &shared.WorkflowMetadata{WorkflowId: "workflow-id", RunId: "run-id"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Fake_Mutation_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activity := New(nil, &sync.Map{}, nil)
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &shared.WorkflowMetadata{WorkflowId: "workflow-id", RunId: "RunId"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Run_Success_Javascript(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activity := New(nil, &sync.Map{}, nil)
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
    - javascript:
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &shared.WorkflowMetadata{WorkflowId: "workflow-id", RunId: "run-id"})
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
	activity := New(nil, &sync.Map{}, nil)
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
    - javascript:
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &shared.WorkflowMetadata{WorkflowId: "workflow-id", RunId: "run-id"})
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
