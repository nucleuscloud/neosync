package neosync_benthos_dynamodb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	input := &dynamodbInput{client: mockClient, table: "foo"}

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
	input := &dynamodbInput{client: mockClient, table: "foo"}

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

func TestAttributeValueToStandardValue(t *testing.T) {
	tests := []struct {
		name     string
		input    types.AttributeValue
		expected any
	}{
		{
			name:     "Binary",
			input:    &types.AttributeValueMemberB{Value: []byte{1, 2, 3}},
			expected: []byte{1, 2, 3},
		},
		{
			name:     "Boolean",
			input:    &types.AttributeValueMemberBOOL{Value: true},
			expected: true,
		},
		{
			name:     "Binary Set",
			input:    &types.AttributeValueMemberBS{Value: [][]byte{{1, 2, 3}, {4, 5, 6}}},
			expected: []any{[]byte{1, 2, 3}, []byte{4, 5, 6}},
		},
		{
			name: "List",
			input: &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberS{Value: "test"},
				&types.AttributeValueMemberN{Value: "123"},
			}},
			expected: []any{"test", "123"},
		},
		{
			name: "Map",
			input: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"key1": &types.AttributeValueMemberS{Value: "value1"},
				"key2": &types.AttributeValueMemberN{Value: "456"},
			}},
			expected: map[string]any{"key1": "value1", "key2": "456"},
		},
		{
			name:     "Number",
			input:    &types.AttributeValueMemberN{Value: "789"},
			expected: "789",
		},
		{
			name:     "Number Set",
			input:    &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}},
			expected: []any{"1", "2", "3"},
		},
		{
			name:     "Null",
			input:    &types.AttributeValueMemberNULL{Value: true},
			expected: nil,
		},
		{
			name:     "String",
			input:    &types.AttributeValueMemberS{Value: "hello"},
			expected: "hello",
		},
		{
			name:     "String Set",
			input:    &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}},
			expected: []any{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := attributeValueToStandardValue(tt.input)
			require.True(t, reflect.DeepEqual(actual, tt.expected), fmt.Sprintf("expected %v, got %v", tt.expected, actual))
		})
	}
}

func TestAttributeValueMapToStandardJSON(t *testing.T) {
	input := map[string]types.AttributeValue{
		"PK":     &types.AttributeValueMemberS{Value: "PrimaryKey"},
		"SK":     &types.AttributeValueMemberS{Value: "SortKey"},
		"Str":    &types.AttributeValueMemberS{Value: "StringValue"},
		"Num":    &types.AttributeValueMemberN{Value: "123.45"},
		"Bool":   &types.AttributeValueMemberBOOL{Value: true},
		"Bin":    &types.AttributeValueMemberB{Value: []byte("BinaryValue")},
		"StrSet": &types.AttributeValueMemberSS{Value: []string{"Str1", "Str2"}},
		"NumSet": &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}},
		"BinSet": &types.AttributeValueMemberBS{Value: [][]byte{[]byte("Bin1"), []byte("Bin2")}},
		"Map": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"NestedStr": &types.AttributeValueMemberS{Value: "NestedStringValue"},
			"NestedNum": &types.AttributeValueMemberN{Value: "456.78"},
		}},
		"List": &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "ListStr"},
			&types.AttributeValueMemberN{Value: "789.01"},
		}},
		"Null": &types.AttributeValueMemberNULL{Value: true},
	}

	expected := map[string]any{
		"PK":     "PrimaryKey",
		"SK":     "SortKey",
		"Str":    "StringValue",
		"Num":    "123.45",
		"Bool":   true,
		"Bin":    []byte("BinaryValue"),
		"StrSet": []any{"Str1", "Str2"},
		"NumSet": []any{"1", "2", "3"},
		"BinSet": []any{[]byte("Bin1"), []byte("Bin2")},
		"Map": map[string]any{
			"NestedStr": "NestedStringValue",
			"NestedNum": "456.78",
		},
		"List": []any{"ListStr", "789.01"},
		"Null": nil,
	}

	actual := attributeValueMapToStandardJSON(input)
	require.True(t, reflect.DeepEqual(actual, expected), fmt.Sprintf("expected %v, got %v", expected, actual))
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
