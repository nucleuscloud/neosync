package neosync_benthos_dynamodb

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/warpstreamlabs/bento/public/service"
)

type mockDynamoDB struct {
	dynamoDBAPI
	fn      func(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	batchFn func(*dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error)
}

func (m *mockDynamoDB) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.fn(params)
}

func (m *mockDynamoDB) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	return m.batchFn(params)
}

func testDDBOWriter(t *testing.T, conf string) *dynamoDBWriter {
	t.Helper()

	pConf, err := dynamoOutputConfigSpec().ParseYAML(conf, nil)
	require.NoError(t, err)

	dConf, err := ddboConfigFromParsed(pConf)
	require.NoError(t, err)

	w, err := newDynamoDBWriter(dConf, service.MockResources())
	require.NoError(t, err)

	return w
}

func TestDynamoDBHappy(t *testing.T) {
	db := testDDBOWriter(t, `
table: FooTable
string_columns:
  id: ${!json("id")}
  content: ${!json("content")}
`)

	var request map[string][]types.WriteRequest

	db.client = &mockDynamoDB{
		fn: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			t.Error("not expected")
			return nil, errors.New("not implemented")
		},
		batchFn: func(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
			request = input.RequestItems
			return &dynamodb.BatchWriteItemOutput{}, nil
		},
	}

	require.NoError(t, db.WriteBatch(context.Background(), service.MessageBatch{
		service.NewMessage([]byte(`{"id":"foo","content":"foo stuff"}`)),
		service.NewMessage([]byte(`{"id":"bar","content":"bar stuff"}`)),
	}))

	expected := map[string][]types.WriteRequest{
		"FooTable": {
			types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{
							Value: "foo",
						},
						"content": &types.AttributeValueMemberS{
							Value: "foo stuff",
						},
					},
				},
			},
			types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{
							Value: "bar",
						},
						"content": &types.AttributeValueMemberS{
							Value: "bar stuff",
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, request)
}

func TestDynamoDBSadToGood(t *testing.T) {
	t.Parallel()

	db := testDDBOWriter(t, `
table: FooTable
string_columns:
  id: ${!json("id")}
  content: ${!json("content")}
backoff:
  max_elapsed_time: 100ms
`)

	var batchRequest []types.WriteRequest
	var requests []*dynamodb.PutItemInput

	db.client = &mockDynamoDB{
		fn: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			requests = append(requests, input)
			return nil, nil
		},
		batchFn: func(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
			if len(batchRequest) > 0 {
				t.Error("not expected")
				return nil, errors.New("not implemented")
			}
			if request, ok := input.RequestItems["FooTable"]; ok {
				items := make([]types.WriteRequest, len(request))
				copy(items, request)
				batchRequest = items
			} else {
				t.Error("missing FooTable")
			}
			return &dynamodb.BatchWriteItemOutput{}, errors.New("woop")
		},
	}

	require.NoError(t, db.WriteBatch(context.Background(), service.MessageBatch{
		service.NewMessage([]byte(`{"id":"foo","content":"foo stuff"}`)),
		service.NewMessage([]byte(`{"id":"bar","content":"bar stuff"}`)),
		service.NewMessage([]byte(`{"id":"baz","content":"baz stuff"}`)),
	}))

	batchExpected := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "foo"},
					"content": &types.AttributeValueMemberS{Value: "foo stuff"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "bar"},
					"content": &types.AttributeValueMemberS{Value: "bar stuff"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "baz"},
					"content": &types.AttributeValueMemberS{Value: "baz stuff"},
				},
			},
		},
	}

	assert.Equal(t, batchExpected, batchRequest)

	expected := []*dynamodb.PutItemInput{
		{
			TableName: aws.String("FooTable"),
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "foo"},
				"content": &types.AttributeValueMemberS{Value: "foo stuff"},
			},
		},
		{
			TableName: aws.String("FooTable"),
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "bar"},
				"content": &types.AttributeValueMemberS{Value: "bar stuff"},
			},
		},
		{
			TableName: aws.String("FooTable"),
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "baz"},
				"content": &types.AttributeValueMemberS{Value: "baz stuff"},
			},
		},
	}

	assert.Equal(t, expected, requests)
}

func TestDynamoDBSadToGoodBatch(t *testing.T) {
	t.Parallel()

	db := testDDBOWriter(t, `
table: FooTable
string_columns:
  id: ${!json("id")}
  content: ${!json("content")}
`)

	var requests [][]types.WriteRequest

	db.client = &mockDynamoDB{
		fn: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			t.Error("not expected")
			return nil, errors.New("not implemented")
		},
		batchFn: func(input *dynamodb.BatchWriteItemInput) (output *dynamodb.BatchWriteItemOutput, err error) {
			if len(requests) == 0 {
				output = &dynamodb.BatchWriteItemOutput{
					UnprocessedItems: map[string][]types.WriteRequest{
						"FooTable": {
							{
								PutRequest: &types.PutRequest{
									Item: map[string]types.AttributeValue{
										"id":      &types.AttributeValueMemberS{Value: "bar"},
										"content": &types.AttributeValueMemberS{Value: "bar stuff"},
									},
								},
							},
						},
					},
				}
			} else {
				output = &dynamodb.BatchWriteItemOutput{}
			}
			if request, ok := input.RequestItems["FooTable"]; ok {
				items := make([]types.WriteRequest, len(request))
				copy(items, request)
				requests = append(requests, items)
			} else {
				t.Error("missing FooTable")
			}
			return
		},
	}

	require.NoError(t, db.WriteBatch(context.Background(), service.MessageBatch{
		service.NewMessage([]byte(`{"id":"foo","content":"foo stuff"}`)),
		service.NewMessage([]byte(`{"id":"bar","content":"bar stuff"}`)),
		service.NewMessage([]byte(`{"id":"baz","content":"baz stuff"}`)),
	}))

	expected := [][]types.WriteRequest{
		{
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "foo"},
						"content": &types.AttributeValueMemberS{Value: "foo stuff"},
					},
				},
			},
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "bar"},
						"content": &types.AttributeValueMemberS{Value: "bar stuff"},
					},
				},
			},
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "baz"},
						"content": &types.AttributeValueMemberS{Value: "baz stuff"},
					},
				},
			},
		},
		{
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "bar"},
						"content": &types.AttributeValueMemberS{Value: "bar stuff"},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, requests)
}

func TestDynamoDBSad(t *testing.T) {
	t.Parallel()

	db := testDDBOWriter(t, `
table: FooTable
string_columns:
  id: ${!json("id")}
  content: ${!json("content")}
`)

	var batchRequest []types.WriteRequest
	var requests []*dynamodb.PutItemInput

	barErr := errors.New("dont like bar")

	db.client = &mockDynamoDB{
		fn: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			if len(requests) < 3 {
				requests = append(requests, input)
			}
			if input.Item["id"].(*types.AttributeValueMemberS).Value == "bar" {
				return nil, barErr
			}
			return nil, nil
		},
		batchFn: func(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
			if len(batchRequest) > 0 {
				t.Error("not expected")
				return nil, errors.New("not implemented")
			}
			if request, ok := input.RequestItems["FooTable"]; ok {
				items := make([]types.WriteRequest, len(request))
				copy(items, request)
				batchRequest = items
			} else {
				t.Error("missing FooTable")
			}
			return &dynamodb.BatchWriteItemOutput{}, errors.New("woop")
		},
	}

	msg := service.MessageBatch{
		service.NewMessage([]byte(`{"id":"foo","content":"foo stuff"}`)),
		service.NewMessage([]byte(`{"id":"bar","content":"bar stuff"}`)),
		service.NewMessage([]byte(`{"id":"baz","content":"baz stuff"}`)),
	}

	expErr := service.NewBatchError(msg, errors.New("woop"))
	err := expErr.Failed(1, barErr)
	require.NotNil(t, err)
	require.Equal(t, expErr, db.WriteBatch(context.Background(), msg))

	batchExpected := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "foo"},
					"content": &types.AttributeValueMemberS{Value: "foo stuff"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "bar"},
					"content": &types.AttributeValueMemberS{Value: "bar stuff"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "baz"},
					"content": &types.AttributeValueMemberS{Value: "baz stuff"},
				},
			},
		},
	}

	assert.Equal(t, batchExpected, batchRequest)

	expected := []*dynamodb.PutItemInput{
		{
			TableName: aws.String("FooTable"),
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "foo"},
				"content": &types.AttributeValueMemberS{Value: "foo stuff"},
			},
		},
		{
			TableName: aws.String("FooTable"),
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "bar"},
				"content": &types.AttributeValueMemberS{Value: "bar stuff"},
			},
		},
		{
			TableName: aws.String("FooTable"),
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "baz"},
				"content": &types.AttributeValueMemberS{Value: "baz stuff"},
			},
		},
	}

	assert.Equal(t, expected, requests)
}

func TestDynamoDBSadBatch(t *testing.T) {
	t.Parallel()

	db := testDDBOWriter(t, `
table: FooTable
string_columns:
  id: ${!json("id")}
  content: ${!json("content")}
`)

	var requests [][]types.WriteRequest

	db.client = &mockDynamoDB{
		fn: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			t.Error("not expected")
			return nil, errors.New("not implemented")
		},
		batchFn: func(input *dynamodb.BatchWriteItemInput) (output *dynamodb.BatchWriteItemOutput, err error) {
			output = &dynamodb.BatchWriteItemOutput{
				UnprocessedItems: map[string][]types.WriteRequest{
					"FooTable": {
						{
							PutRequest: &types.PutRequest{
								Item: map[string]types.AttributeValue{
									"id":      &types.AttributeValueMemberS{Value: "bar"},
									"content": &types.AttributeValueMemberS{Value: "bar stuff"},
								},
							},
						},
					},
				},
			}
			if len(requests) < 2 {
				if request, ok := input.RequestItems["FooTable"]; ok {
					items := make([]types.WriteRequest, len(request))
					copy(items, request)
					requests = append(requests, items)
				} else {
					t.Error("missing FooTable")
				}
			}
			return
		},
	}

	msg := service.MessageBatch{
		service.NewMessage([]byte(`{"id":"foo","content":"foo stuff"}`)),
		service.NewMessage([]byte(`{"id":"bar","content":"bar stuff"}`)),
		service.NewMessage([]byte(`{"id":"baz","content":"baz stuff"}`)),
	}

	require.Equal(t, errors.New("failed to set 1 items"), db.WriteBatch(context.Background(), msg))

	expected := [][]types.WriteRequest{
		{
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "foo"},
						"content": &types.AttributeValueMemberS{Value: "foo stuff"},
					},
				},
			},
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "bar"},
						"content": &types.AttributeValueMemberS{Value: "bar stuff"},
					},
				},
			},
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "baz"},
						"content": &types.AttributeValueMemberS{Value: "baz stuff"},
					},
				},
			},
		},
		{
			{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":      &types.AttributeValueMemberS{Value: "bar"},
						"content": &types.AttributeValueMemberS{Value: "bar stuff"},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, requests)
}

func Test_MarshalToAttributeValue(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		root       any
		keyTypeMap map[string]neosync_types.KeyType
		want       types.AttributeValue
	}{
		{
			name:       "String",
			key:        "StrKey",
			root:       "value",
			keyTypeMap: map[string]neosync_types.KeyType{},
			want:       &types.AttributeValueMemberS{Value: "value"},
		},
		{
			name:       "Number",
			key:        "NumKey",
			root:       123,
			keyTypeMap: map[string]neosync_types.KeyType{},
			want:       &types.AttributeValueMemberN{Value: "123"},
		},
		{
			name:       "Boolean",
			key:        "BoolKey",
			root:       true,
			keyTypeMap: map[string]neosync_types.KeyType{},
			want:       &types.AttributeValueMemberBOOL{Value: true},
		},
		{
			name:       "Null",
			key:        "NullKey",
			root:       nil,
			keyTypeMap: map[string]neosync_types.KeyType{},
			want:       &types.AttributeValueMemberNULL{Value: true},
		},
		{
			name:       "StringSet",
			key:        "SSKey",
			root:       []string{"a", "b"},
			keyTypeMap: map[string]neosync_types.KeyType{"SSKey": neosync_types.StringSet},
			want:       &types.AttributeValueMemberSS{Value: []string{"a", "b"}},
		},
		{
			name:       "NumberSet",
			key:        "NSKey",
			root:       []int{1, 2},
			keyTypeMap: map[string]neosync_types.KeyType{"NSKey": neosync_types.NumberSet},
			want:       &types.AttributeValueMemberNS{Value: []string{"1", "2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := marshalToAttributeValue(tt.key, tt.root, tt.keyTypeMap)
			require.Equalf(t, tt.want, got, fmt.Sprintf("MarshalToAttributeValue() = %v, want %v", got, tt.want))
		})
	}
}

func Test_FormatFloat(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"Integer", 123.0, "123.0"},
		{"Decimal", 123.456, "123.456"},
		{"Many decimal places", 123.4567890, "123.4568"},
		{"Trailing zeros", 123.4000, "123.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFloat(tt.input)
			require.Equal(t, tt.want, got, fmt.Sprintf("formatFloat() = %v, want %v", got, tt.want))
		})
	}
}

func Test_ConvertToStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []string
		wantErr bool
	}{
		{"String slice", []string{"a", "b", "c"}, []string{"a", "b", "c"}, false},
		{"Int slice", []int{1, 2, 3}, []string{"1", "2", "3"}, false},
		{"Float slice", []float64{1.12, 2.0, 3.43}, []string{"1.12", "2.0", "3.43"}, false},
		{"Mixed slice", []any{"a", 1, true}, []string{"a", "1", "true"}, false},
		{"Not a slice", "not a slice", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToStringSlice(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equalf(t, tt.want, got, fmt.Sprintf("convertToStringSlice() = %v, want %v", got, tt.want))
			}
		})
	}
}

func Test_AnyToString(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"String", "hello", "hello"},
		{"Int", 123, "123"},
		{"Float", 123.456, "123.456"},
		{"Boolean", true, "true"},
		{"Byte slice", []byte("hello"), "hello"},
		{"Nil", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := anyToString(tt.input)
			require.Equalf(t, tt.want, got, fmt.Sprintf("anyToString() = %v, want %v", got, tt.want))
		})
	}
}
