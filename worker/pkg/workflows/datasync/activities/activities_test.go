package datasync_activities

import (
	"math"
	"strings"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
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

func TestClampInt(t *testing.T) {
	assert.Equal(t, clampInt(0, 1, 2), 1)
	assert.Equal(t, clampInt(1, 1, 2), 1)
	assert.Equal(t, clampInt(2, 1, 2), 2)
	assert.Equal(t, clampInt(3, 1, 2), 2)
	assert.Equal(t, clampInt(1, 1, 1), 1)

	assert.Equal(t, clampInt(1, 3, 2), 3, "low is evaluated first, order is relevant")

}

func TestComputeMaxPgBatchCount(t *testing.T) {
	assert.Equal(t, computeMaxPgBatchCount(65535), 1)
	assert.Equal(t, computeMaxPgBatchCount(65536), 1, "anything over max should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(math.MaxInt), 1, "anything over pgmax should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(1), 65535)
	assert.Equal(t, computeMaxPgBatchCount(0), 65535)
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
	}, &SyncMetadata{Schema: "public", Table: "test"})
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
	}, &SyncMetadata{Schema: "public", Table: "test"})
	require.NoError(t, err)
	res := &SyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
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

func strPtr(val string) *string {
	return &val
}

func boolPtr(val bool) *bool {
	return &val
}
