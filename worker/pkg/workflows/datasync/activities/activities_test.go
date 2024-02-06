package datasync_activities

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/javascript"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestAreMappingsSubsetOfSchemas(t *testing.T) {
	ok := areMappingsSubsetOfSchemas(
		map[string]map[string]struct{}{
			"public.users": {
				"id":         struct{}{},
				"created_by": struct{}{},
				"updated_by": struct{}{},
			},
			"neosync_api.accounts": {
				"id": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.True(t, ok, "job mappings are a subset of the present database schemas")

	ok = areMappingsSubsetOfSchemas(
		map[string]map[string]struct{}{
			"public.users": {
				"id": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id2"},
		},
	)
	assert.False(t, ok, "job mappings contain mapping that is not in the source schema")

	ok = areMappingsSubsetOfSchemas(
		map[string]map[string]struct{}{
			"public.users": {
				"id": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.False(t, ok, "job mappings contain more mappings than are present in the source schema")
}

func TestShouldHaltOnSchemaAddition(t *testing.T) {
	ok := shouldHaltOnSchemaAddition(
		map[string]map[string]struct{}{
			"public.users": {
				"id":         struct{}{},
				"created_by": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.False(t, ok, "job mappings are valid set of database schemas")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]struct{}{
			"public.users": {
				"id":         struct{}{},
				"created_by": struct{}{},
			},
			"neosync_api.accounts": {
				"id": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.True(t, ok, "job mappings are missing database schema mappings")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]struct{}{
			"public.users": {
				"id":         struct{}{},
				"created_by": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
		},
	)
	assert.True(t, ok, "job mappings are missing table column")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]struct{}{
			"public.users": {
				"id":         struct{}{},
				"created_by": struct{}{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "updated_by"},
		},
	)
	assert.True(t, ok, "job mappings have same column count, but missing specific column")
}

func Test_Sync_Run_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := &Activities{}
	env.RegisterActivity(activities)

	val, err := env.ExecuteActivity(activities.Sync, &SyncRequest{
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &WorkflowMetadata{WorkflowId: "workflow-id", RunId: "run-id"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Fake_Mutation_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := &Activities{}
	env.RegisterActivity(activities)

	val, err := env.ExecuteActivity(activities.Sync, &SyncRequest{
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &WorkflowMetadata{WorkflowId: "workflow-id", RunId: "RunId"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Sync_Run_Success_Javascript(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := &Activities{}
	env.RegisterActivity(activities)

	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	val, err := env.ExecuteActivity(activities.Sync, &SyncRequest{
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &WorkflowMetadata{WorkflowId: "workflow-id", RunId: "run-id"})
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

	activities := &Activities{}
	env.RegisterActivity(activities)

	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	val, err := env.ExecuteActivity(activities.Sync, &SyncRequest{
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
	}, &SyncMetadata{Schema: "public", Table: "test"}, &WorkflowMetadata{WorkflowId: "workflow-id", RunId: "run-id"})
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

func Test_buildProcessorConfigsMutation(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	ctx := context.Background()

	output, err := bbuilder.buildProcessorConfigs(ctx, nil)
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id"},
	})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{}},
	})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "passthrough"}},
	})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "null", Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
		{Schema: "public", Table: "users", Column: "name", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "null", Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
	})

	assert.Nil(t, err)

	assert.Equal(t, *output[0].Mutation, "root.id = null\nroot.name = null")

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_email",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain: true,
					PreserveLength: false,
				},
			},
		},
	}

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "email", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}})

	assert.Nil(t, err)
	assert.Equal(t, *output[0].Mutation, `root.email = transform_email(email:this.email,preserve_domain:true,preserve_length:false)`)

	output, err = bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "i_do_not_exist", Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
	})
	assert.Error(t, err)
	assert.Empty(t, output)

}

const code = `var payload = value+=" hello";return payload;`

func Test_buildProcessorConfigsJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	ctx := context.Background()

	col := "address"

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code,
				},
			},
		},
	}

	res, err := bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: col, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}})

	assert.NoError(t, err)
	assert.Equal(t, `
(() => {

function fn_address(value, input){
  var payload = value+=" hello";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["address"] = fn_address(input["address"], input);
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

const col = "name"

func Test_buildProcessorConfigsJavascriptMultiLineScript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	ctx := context.Background()

	code :=
		`var payload = value+=" hello";
  payload.replace("hello","newHello");
  return payload;`

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code,
				},
			},
		},
	}

	res, err := bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: col, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}})

	assert.NoError(t, err)
	assert.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";
  payload.replace("hello","newHello");
  return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

func Test_buildProcessorConfigsJavascriptMultiple(t *testing.T) {

	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	ctx := context.Background()

	code2 := `var payload = value*2;return payload;`
	col2 := "age"

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code,
				},
			},
		},
	}

	jsT2 := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code2,
				},
			},
		},
	}

	res, err := bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: col, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}},
		{Schema: "public", Table: "users", Column: col2, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT2.Source, Config: jsT2.Config}}})

	assert.NoError(t, err)
	assert.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";return payload;
};


function fn_age(value, input){
  var payload = value*2;return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
output["age"] = fn_age(input["age"], input);
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

func Test_ShouldProcessColumnTrue(t *testing.T) {

	val := &mgmtv1alpha1.JobMappingTransformer{
		Source: "generate_email",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		},
	}

	res := shouldProcessColumn(val)
	assert.Equal(t, true, res)
}

func Test_ShouldProcessColumnFalse(t *testing.T) {

	val := &mgmtv1alpha1.JobMappingTransformer{
		Source: "passthrough",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		},
	}

	res := shouldProcessColumn(val)
	assert.Equal(t, false, res)
}

func Test_ConstructJsFunction(t *testing.T) {

	col := "col"

	res := constructJsFunction(code, col)
	assert.Equal(t, `
function fn_col(value, input){
  var payload = value+=" hello";return payload;
};
`, res)
}

func Test_ConstructBenthosJsProcessor(t *testing.T) {

	jsFunctions := []string{}
	benthosOutputs := []string{}

	benthosOutput := constructBenthosOutput(col)
	jsFunction := constructJsFunction(code, col)
	benthosOutputs = append(benthosOutputs, benthosOutput)

	jsFunctions = append(jsFunctions, jsFunction)

	res := constructBenthosJsProcessor(jsFunctions, benthosOutputs)

	assert.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
benthos.v0_msg_set_structured(output);
})();`, res)
}

func Test_ConstructBenthosOutput(t *testing.T) {

	col := "col"

	res := constructBenthosOutput(col)

	assert.Equal(t, `output["col"] = fn_col(input["col"], input);`, res)
}

func Test_buildProcessorConfigsJavascriptEmpty(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: ``,
				},
			},
		},
	}

	resp, err := bbuilder.buildProcessorConfigs(ctx, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}})

	assert.NoError(t, err)
	assert.Empty(t, resp)

}

func Test_convertUserDefinedFunctionConfig(t *testing.T) {

	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	ctx := context.Background()

	mockTransformerClient.On(
		"GetUserDefinedTransformerById",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
			TransformerId: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformerByIdResponse{
		Transformer: &mgmtv1alpha1.UserDefinedTransformer{
			Id:          "123",
			Name:        "stage",
			Description: "description",
			DataType:    "string",
			Source:      "transform_email",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain: true,
						PreserveLength: false,
					},
				},
			},
		},
	}), nil)

	jmt := &mgmtv1alpha1.JobMappingTransformer{
		Source: "transform_email",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig{
				UserDefinedTransformerConfig: &mgmtv1alpha1.UserDefinedTransformerConfig{
					Id: "123",
				},
			},
		},
	}

	expected := &mgmtv1alpha1.JobMappingTransformer{
		Source: "transform_email",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain: true,
					PreserveLength: false,
				},
			},
		},
	}

	resp, err := bbuilder.convertUserDefinedFunctionConfig(ctx, jmt)
	assert.NoError(t, err)
	assert.Equal(t, resp, expected)

}

//nolint:all
func MockJobMappingTransformer(source, transformerId string) db_queries.NeosyncApiTransformer {

	return db_queries.NeosyncApiTransformer{
		Source:            source,
		TransformerConfig: &pg_models.TransformerConfigs{},
	}
}

func Test_buildPlainInsertArgs(t *testing.T) {
	assert.Empty(t, buildPlainInsertArgs(nil))
	assert.Empty(t, buildPlainInsertArgs([]string{}))
	assert.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), "root = [this.foo, this.bar, this.baz]")
}

func Test_buildPlainColumns(t *testing.T) {
	assert.Empty(t, buildPlainColumns(nil))
	assert.Empty(t, buildPlainColumns([]*mgmtv1alpha1.JobMapping{}))
	assert.Equal(
		t,
		buildPlainColumns([]*mgmtv1alpha1.JobMapping{
			{Column: "foo"},
			{Column: "bar"},
			{Column: "baz"},
		}),
		[]string{"foo", "bar", "baz"},
	)
}

func Test_splitTableKey(t *testing.T) {
	schema, table := splitTableKey("foo")
	assert.Equal(t, schema, "public")
	assert.Equal(t, table, "foo")

	schema, table = splitTableKey("neosync.foo")
	assert.Equal(t, schema, "neosync")
	assert.Equal(t, table, "foo")
}

func Test_buildBenthosS3Credentials(t *testing.T) {
	assert.Nil(t, buildBenthosS3Credentials(nil))

	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{}),
		&neosync_benthos.AwsCredentials{},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{Profile: strPtr("foo")}),
		&neosync_benthos.AwsCredentials{Profile: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{AccessKeyId: strPtr("foo")}),
		&neosync_benthos.AwsCredentials{Id: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SecretAccessKey: strPtr("foo")}),
		&neosync_benthos.AwsCredentials{Secret: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SessionToken: strPtr("foo")}),
		&neosync_benthos.AwsCredentials{Token: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{FromEc2Role: boolPtr(true)}),
		&neosync_benthos.AwsCredentials{FromEc2Role: true},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleArn: strPtr("foo")}),
		&neosync_benthos.AwsCredentials{Role: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleExternalId: strPtr("foo")}),
		&neosync_benthos.AwsCredentials{RoleExternalId: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{
			Profile:         strPtr("profile"),
			AccessKeyId:     strPtr("access-key"),
			SecretAccessKey: strPtr("secret"),
			SessionToken:    strPtr("session"),
			FromEc2Role:     boolPtr(false),
			RoleArn:         strPtr("role"),
			RoleExternalId:  strPtr("foo"),
		}),
		&neosync_benthos.AwsCredentials{
			Profile:        "profile",
			Id:             "access-key",
			Secret:         "secret",
			Token:          "session",
			FromEc2Role:    false,
			Role:           "role",
			RoleExternalId: "foo",
		},
	)
}

func Test_getPgDsn(t *testing.T) {
	dsn, err := getPgDsn(nil)
	assert.Error(t, err)
	assert.Empty(t, dsn)

	dsn, err = getPgDsn(&mgmtv1alpha1.PostgresConnectionConfig{})
	assert.Error(t, err)
	assert.Empty(t, dsn)

	dsn, err = getPgDsn(&mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{},
	})
	assert.Nil(t, err)
	assert.Empty(t, dsn)

	dsn, err = getPgDsn(&mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{Url: "foo"},
	})
	assert.Nil(t, err)
	assert.Equal(t, dsn, "foo")

	dsn, err = getPgDsn(&mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{},
	})
	assert.Error(t, err)
	assert.Empty(t, dsn)

	dsn, err = getPgDsn(&mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
			Connection: &mgmtv1alpha1.PostgresConnection{},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, dsn, "postgres://:@:0/")

	sslMode := "disable"
	dsn, err = getPgDsn(&mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
			Connection: &mgmtv1alpha1.PostgresConnection{
				User:    "my-user",
				Pass:    "my-pass",
				SslMode: &sslMode,
				Host:    "localhost",
				Port:    5432,
				Name:    "neosync",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, dsn, "postgres://my-user:my-pass@localhost:5432/neosync?sslmode=disable")
}

func Test_getMysqlDsn(t *testing.T) {
	dsn, err := getMysqlDsn(nil)
	assert.Error(t, err)
	assert.Empty(t, dsn)

	dsn, err = getMysqlDsn(&mgmtv1alpha1.MysqlConnectionConfig{})
	assert.Error(t, err)
	assert.Empty(t, dsn)

	dsn, err = getMysqlDsn(&mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{},
	})
	assert.Nil(t, err)
	assert.Empty(t, dsn)

	dsn, err = getMysqlDsn(&mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{Url: "foo"},
	})
	assert.Nil(t, err)
	assert.Equal(t, dsn, "foo")

	dsn, err = getMysqlDsn(&mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{},
	})
	assert.Error(t, err)
	assert.Empty(t, dsn)

	dsn, err = getMysqlDsn(&mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
			Connection: &mgmtv1alpha1.MysqlConnection{},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, dsn, ":@(:0)/")

	dsn, err = getMysqlDsn(&mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
			Connection: &mgmtv1alpha1.MysqlConnection{
				User:     "my-user",
				Pass:     "my-pass",
				Protocol: "tcp",
				Host:     "localhost",
				Port:     5432,
				Name:     "neosync",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, dsn, "my-user:my-pass@tcp(localhost:5432)/neosync")
}

func strPtr(val string) *string {
	return &val
}

func boolPtr(val bool) *bool {
	return &val
}

func Test_computeMutationFunction_null(t *testing.T) {
	val, err := computeMutationFunction(
		&mgmtv1alpha1.JobMapping{
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Source: "null",
			},
		})
	assert.NoError(t, err)
	assert.Equal(t, val, "null")
}

// nolint
func Test_sha256Hash_transformer_string(t *testing.T) {

	mapping := `root = this.bytes().hash("sha256").encode("hex")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the sha256 transformer")

	val := "hello"
	res, err := ex.Query(val)
	assert.NoError(t, err)

	// hash the value
	bites := []byte(val)
	hasher := sha256.New()
	_, err = hasher.Write(bites)
	assert.NoError(t, err)

	// compute sha256 checksum and encode it into a hex string
	hashed := hasher.Sum(nil)
	var buf bytes.Buffer
	e := hex.NewEncoder(&buf)
	_, err = e.Write(hashed)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, res, buf.String())
}

// nolint
func Test_sha256Hash_transformer_int64(t *testing.T) {

	mapping := `root = this.bytes().hash("sha256").encode("hex")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the sha256 transformer")

	val := 20
	res, err := ex.Query(val)
	assert.NoError(t, err)

	// hash the value
	bites := strconv.AppendInt(nil, int64(val), 10)
	hasher := sha256.New()
	_, err = hasher.Write(bites)
	assert.NoError(t, err)

	// compute sha256 checksum and encode it into a hex string
	hashed := hasher.Sum(nil)
	var buf bytes.Buffer
	e := hex.NewEncoder(&buf)
	_, err = e.Write(hashed)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, res, buf.String())
}

// nolint
func Test_sha256Hash_transformer_float(t *testing.T) {

	mapping := `root = this.bytes().hash("sha256").encode("hex")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the sha256 transformer")

	val := 20.39
	res, err := ex.Query(val)
	assert.NoError(t, err)

	// hash the value
	bites := strconv.AppendFloat(nil, val, 'g', -1, 64)
	hasher := sha256.New()
	_, err = hasher.Write(bites)
	assert.NoError(t, err)

	// compute sha256 checksum and encode it into a hex string
	hashed := hasher.Sum(nil)
	var buf bytes.Buffer
	e := hex.NewEncoder(&buf)
	_, err = e.Write(hashed)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, res, buf.String())
}

func Test_TransformerStringLint(t *testing.T) {

	col := "email"

	transformers := []*mgmtv1alpha1.SystemTransformer{
		{

			Name: "generate_email",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
					GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
				},
			},
		},
		{
			Name: "transform_email",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain: false,
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "generate_bool",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		},
		{
			Name: "generate_card_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
					GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
						ValidLuhn: true,
					},
				},
			},
		},
		{
			Name: "generate_city",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
					GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
				},
			},
		},
		{
			Name: "generate_e164_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInternationalPhoneNumberConfig{
					GenerateInternationalPhoneNumberConfig: &mgmtv1alpha1.GenerateInternationalPhoneNumber{
						Min: 9,
						Max: 15,
					},
				},
			},
		},
		{
			Name: "generate_first_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			},
		},
		{
			Name: "generate_float64",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						RandomizeSign: true,
						Min:           1.00,
						Max:           100.00,
					},
				},
			},
		},
		{
			Name: "generate_full_address",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
					GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
				},
			},
		},
		{
			Name: "generate_full_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		},
		{
			Name: "generate_gender",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
					GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
						Abbreviate: false,
					},
				},
			},
		},
		{
			Name: "generate_int64_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
					GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
				},
			},
		},
		{
			Name: "generate_int64",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						RandomizeSign: true,
						Min:           1,
						Max:           4,
					},
				},
			},
		},
		{
			Name: "generate_last_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
					GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
				},
			},
		},
		{
			Name: "generate_sha256hash",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
					GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
				},
			},
		},
		{
			Name: "generate_ssn",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
					GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
				},
			},
		},
		{
			Name: "generate_state",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
					GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
				},
			},
		},
		{
			Name: "generate_street_address",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
					GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
				},
			},
		},
		{
			Name: "generate_string_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
					GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
						Min: 10,
						Max: 10,
					},
				},
			},
		},
		{
			Name: "generate_random_string",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateRandomStringConfig{
					GenerateRandomStringConfig: &mgmtv1alpha1.GenerateRandomString{
						Min: 2,
						Max: 7,
					},
				},
			},
		},
		{
			Name: "generate_unixtimestamp",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
					GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
				},
			},
		},
		{
			Name: "generate_username",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
					GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
				},
			},
		},
		{
			Name: "generate_utctimestamp",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
					GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
				},
			},
		},
		{
			Name: "generate_uuid",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: true,
					},
				},
			},
		},
		{
			Name: "generate_zipcode",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
					GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
				},
			},
		},
		{
			Name: "transform_e164_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
					TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "transform_first_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "transform_float64",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
					TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
						RandomizationRangeMin: 20.00,
						RandomizationRangeMax: 50.00,
					},
				},
			},
		},
		{
			Name: "transform_full_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "transform_int64_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
					TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "transform_int64",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
					TransformInt64Config: &mgmtv1alpha1.TransformInt64{
						RandomizationRangeMin: 20,
						RandomizationRangeMax: 50,
					},
				},
			},
		},
		{
			Name: "transform_last_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
					TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "transform_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
					TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
						PreserveLength: false,
						IncludeHyphens: false,
					},
				},
			},
		},
		{
			Name: "transform_string",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Name: "generate_categorical",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
					GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
						Categories: "value1,value2",
					},
				},
			},
		},
		{
			Name: "transform_character_scramble",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
				},
			},
		},
	}

	for _, transformer := range transformers {

		val, err := computeMutationFunction(
			&mgmtv1alpha1.JobMapping{
				Column: col,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: transformer.Name,
					Config: transformer.Config,
				},
			})

		assert.NoError(t, err)

		_, err = bloblang.Parse(val)
		assert.NoError(t, err, "transformer lint failed, check that the transformer string is being constructed correctly.")
	}
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
				StartToCloseTimeout: ptr(int64(2)),
			},
		}, expected: &workflow.ActivityOptions{StartToCloseTimeout: 2, RetryPolicy: defaultOpts.RetryPolicy}},
		{name: "custom schedule to close timeout", input: &mgmtv1alpha1.Job{
			SyncOptions: &mgmtv1alpha1.ActivityOptions{
				ScheduleToCloseTimeout: ptr(int64(2)),
			},
		}, expected: &workflow.ActivityOptions{ScheduleToCloseTimeout: 2, RetryPolicy: defaultOpts.RetryPolicy}},
		{name: "custom retry policy", input: &mgmtv1alpha1.Job{
			SyncOptions: &mgmtv1alpha1.ActivityOptions{
				RetryPolicy: &mgmtv1alpha1.RetryPolicy{
					MaximumAttempts: ptr(int32(2)),
				},
			},
		}, expected: &workflow.ActivityOptions{StartToCloseTimeout: defaultOpts.StartToCloseTimeout, RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 2}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := getSyncActivityOptionsFromJob(test.input)
			assert.Equal(t, test.expected, output)
		})
	}
}
