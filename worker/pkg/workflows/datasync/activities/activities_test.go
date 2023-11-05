package datasync_activities

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
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

func Test_buildProcessorMutation(t *testing.T) {
	output, err := buildProcessorMutation(nil)
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorMutation([]*mgmtv1alpha1.JobMapping{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorMutation([]*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id"},
	})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorMutation([]*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.Transformer{}},
	})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorMutation([]*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.Transformer{Value: "passthrough"}},
	})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorMutation([]*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.Transformer{Value: "null"}},
		{Schema: "public", Table: "users", Column: "name", Transformer: &mgmtv1alpha1.Transformer{Value: "null"}},
	})
	assert.Nil(t, err)
	assert.Equal(t, output, "root.id = null\nroot.name = null")

	output, err = buildProcessorMutation([]*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.Transformer{Value: "i_do_not_exist"}},
	})
	assert.Error(t, err)
	assert.Empty(t, output)
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
			Transformer: &mgmtv1alpha1.Transformer{
				Value: "null",
			},
		})
	assert.NoError(t, err)
	assert.Equal(t, val, "null")
}

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
