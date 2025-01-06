package neosync_benthos_dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbmapper "github.com/nucleuscloud/neosync/internal/database-record-mapper/dynamodb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_isTableActive(t *testing.T) {
	type testcase struct {
		input    *dynamodb.DescribeTableOutput
		expected bool
	}

	testcases := []testcase{
		{nil, false},
		{&dynamodb.DescribeTableOutput{}, false},
		{&dynamodb.DescribeTableOutput{Table: nil}, false},
		{&dynamodb.DescribeTableOutput{Table: &types.TableDescription{}}, false},
		{&dynamodb.DescribeTableOutput{Table: &types.TableDescription{TableStatus: types.TableStatusArchived}}, false},
		{&dynamodb.DescribeTableOutput{Table: &types.TableDescription{TableStatus: types.TableStatusActive}}, true},
	}

	for _, tc := range testcases {
		actual := isTableActive(tc.input)
		require.Equal(t, tc.expected, actual)
	}
}

func Test_dynamoDbBatchInput_Connect_Client(t *testing.T) {
	mockClient := NewMockdynamoDBAPIV2(t)
	input := &dynamodbInput{client: mockClient}
	err := input.Connect(context.Background())
	require.NoError(t, err)
}

func Test_dynamoDbBatchInput_ReadBatch_NotConnected(t *testing.T) {
	input := &dynamodbInput{}
	_, _, err := input.ReadBatch(context.Background())
	require.Error(t, err)
	require.Equal(t, service.ErrNotConnected, err)
}

func Test_dynamoDbBatchInput_ReadBatch_EndOfInput(t *testing.T) {
	mockClient := NewMockdynamoDBAPIV2(t)
	input := &dynamodbInput{client: mockClient, done: true}
	_, _, err := input.ReadBatch(context.Background())
	require.Error(t, err)
	require.Equal(t, service.ErrEndOfInput, err)
}

func Test_dynamoDbBatchInput_ReadBatch_SinglePage(t *testing.T) {
	mockClient := NewMockdynamoDBAPIV2(t)
	input := &dynamodbInput{client: mockClient, table: "foo", recordMapper: dynamodbmapper.NewDynamoBuilder()}

	mockClient.On("ExecuteStatement", mock.Anything, mock.Anything).Return(&dynamodb.ExecuteStatementOutput{
		Items: []map[string]types.AttributeValue{
			{"f": &types.AttributeValueMemberBOOL{Value: false}},
			{"g": &types.AttributeValueMemberBOOL{Value: true}},
		},
	}, nil)

	batch, _, err := input.ReadBatch(context.Background())
	require.NoError(t, err)
	require.Len(t, batch, 2)
	require.Nil(t, input.nextToken)
	require.True(t, input.done)
}

func Test_dynamoDbBatchInput_ReadBatch_MultiPage(t *testing.T) {
	mockClient := NewMockdynamoDBAPIV2(t)
	input := &dynamodbInput{client: mockClient, table: "foo", recordMapper: dynamodbmapper.NewDynamoBuilder()}

	mockClient.On("ExecuteStatement", mock.Anything, mock.Anything).Return(&dynamodb.ExecuteStatementOutput{
		Items: []map[string]types.AttributeValue{
			{"f": &types.AttributeValueMemberBOOL{Value: false}},
			{"g": &types.AttributeValueMemberBOOL{Value: true}},
		},
		NextToken: aws.String("foo"),
	}, nil)

	batch, _, err := input.ReadBatch(context.Background())
	require.NoError(t, err)
	require.Len(t, batch, 2)
	require.NotNil(t, input.nextToken)
	require.False(t, input.done)
}

func Test_dynamoDbBatchInput_Close(t *testing.T) {
	mockClient := NewMockdynamoDBAPIV2(t)

	input := &dynamodbInput{}
	err := input.Close(context.Background())
	require.NoError(t, err)

	input.client = mockClient
	err = input.Close(context.Background())
	require.NoError(t, err)
	require.Nil(t, input.client)
}

func Test_buildExecStatement(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		where    *string
		expected string
	}{
		{
			name:     "No Where Clause",
			table:    "users",
			where:    nil,
			expected: `SELECT * FROM "users"`,
		},
		{
			name:     "Empty Where Clause",
			table:    "users",
			where:    func() *string { s := ""; return &s }(),
			expected: `SELECT * FROM "users"`,
		},
		{
			name:     "Valid Where Clause",
			table:    "users",
			where:    func() *string { s := "id = 1"; return &s }(),
			expected: `SELECT * FROM "users" WHERE id = 1`,
		},
		{
			name:     "Another Table with Where Clause",
			table:    "orders",
			where:    func() *string { s := "status = 'shipped'"; return &s }(),
			expected: `SELECT * FROM "orders" WHERE status = 'shipped'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildExecStatement(tt.table, tt.where)
			require.True(t, result == tt.expected, "expected %v, got %v", tt.expected, result)
		})
	}
}

func Test_RegisterDynamoDBInput(t *testing.T) {
	err := RegisterDynamoDbInput(service.NewEmptyEnvironment())
	require.NoError(t, err)
}

func Test_InputBasic_Config(t *testing.T) {
	conf, err := dynamoInputConfigSpec().ParseYAML(`
table: test-table
`, service.NewEmptyEnvironment())
	require.NoError(t, err)
	require.NotNil(t, conf)
}

func Test_InputBasic_Config_Opts(t *testing.T) {
	conf, err := dynamoInputConfigSpec().ParseYAML(`
table: test-table
where: foo = '123'
consistent_read: true
`, service.NewEmptyEnvironment())
	require.NoError(t, err)
	require.NotNil(t, conf)
}

func Test_InputBasic_Config_Creds(t *testing.T) {
	conf, err := dynamoInputConfigSpec().ParseYAML(`
table: test-table
region: us-west-2
endpoint: http://localhost:8000
credentials:
  profile: default
  id: dummyid
  secret: dummysecret
  token: dummytoken
  from_ec2_role: true
  role: my-role
  role_external_id: 123
`, service.NewEmptyEnvironment())
	require.NoError(t, err)
	require.NotNil(t, conf)
}

func Test_Input_AwsCreds(t *testing.T) {
	conf, err := dynamoInputConfigSpec().ParseYAML(`
table: test-table
region: us-west-2
endpoint: http://localhost:8000
credentials:
  profile: default
  id: dummyid
  secret: dummysecret
  token: dummytoken
  from_ec2_role: true
  role: my-role
  role_external_id: 123
`, service.NewEmptyEnvironment())
	require.NoError(t, err)
	require.NotNil(t, conf)

	credsConfig := getAwsCredentialsConfigFromParsedConf(conf)
	require.NotNil(t, credsConfig)
	require.Equal(t, "us-west-2", credsConfig.Region)
	require.Equal(t, "http://localhost:8000", credsConfig.Endpoint)
	require.Equal(t, "default", credsConfig.Profile)
	require.Equal(t, "dummyid", credsConfig.Id)
	require.Equal(t, "dummysecret", credsConfig.Secret)
	require.Equal(t, "dummytoken", credsConfig.Token)
	require.True(t, credsConfig.UseEc2)
	require.Equal(t, "my-role", credsConfig.Role)
	require.Equal(t, "123", credsConfig.RoleExternalId)
	require.Equal(t, "neosync", credsConfig.RoleSessionName)
}
