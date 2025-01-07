package dynamodb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	"github.com/stretchr/testify/require"
)

func Test_UnmarshalAttributeValueMap(t *testing.T) {
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
		"Num":    123.45,
		"Bool":   true,
		"Bin":    []byte("BinaryValue"),
		"StrSet": []any{"Str1", "Str2"},
		"NumSet": []any{int64(1), int64(2), int64(3)},
		"BinSet": [][]byte{[]byte("Bin1"), []byte("Bin2")},
		"Map": map[string]any{
			"NestedStr": "NestedStringValue",
			"NestedNum": 456.78,
		},
		"List": []any{"ListStr", 789.01},
		"Null": nil,
	}

	mapper := NewDynamoBuilder()

	actual, keyTypeMap, err := mapper.MapRecordWithKeyType(input)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(actual, expected), fmt.Sprintf("expected %v, got %v", expected, actual))
	require.Equal(t, keyTypeMap["StrSet"], neosync_types.StringSet)
	require.Equal(t, keyTypeMap["NumSet"], neosync_types.NumberSet)
}

func Test_ParseAttributeValue(t *testing.T) {
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
			expected: [][]byte{{1, 2, 3}, {4, 5, 6}},
		},
		{
			name: "List",
			input: &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberS{Value: "test"},
				&types.AttributeValueMemberN{Value: "123"},
			}},
			expected: []any{"test", int64(123)},
		},
		{
			name: "Map",
			input: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"key1": &types.AttributeValueMemberS{Value: "value1"},
				"key2": &types.AttributeValueMemberN{Value: "456"},
			}},
			expected: map[string]any{"key1": "value1", "key2": int64(456)},
		},
		{
			name:     "Number",
			input:    &types.AttributeValueMemberN{Value: "789"},
			expected: int64(789),
		},
		{
			name:     "Number Set",
			input:    &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}},
			expected: []any{int64(1), int64(2), int64(3)},
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
			ktm := map[string]neosync_types.KeyType{}
			actual := parseAttributeValue(tt.name, tt.input, ktm)
			require.True(t, reflect.DeepEqual(actual, tt.expected), fmt.Sprintf("expected %v %v, got %v %v", tt.expected, reflect.TypeOf(tt.expected), actual, reflect.TypeOf(actual)))
		})
	}
}
