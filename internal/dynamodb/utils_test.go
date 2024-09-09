package dynamodb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	"github.com/stretchr/testify/require"
)

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
			actual := ParseAttributeValue(tt.name, tt.input, ktm)
			require.True(t, reflect.DeepEqual(actual, tt.expected), fmt.Sprintf("expected %v %v, got %v %v", tt.expected, reflect.TypeOf(tt.expected), actual, reflect.TypeOf(actual)))
		})
	}
}

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

	actual, keyTypeMap := UnmarshalAttributeValueMap(input)
	require.True(t, reflect.DeepEqual(actual, expected), fmt.Sprintf("expected %v, got %v", expected, actual))
	require.Equal(t, keyTypeMap["StrSet"], neosync_types.StringSet)
	require.Equal(t, keyTypeMap["NumSet"], neosync_types.NumberSet)
}

func Test_UnmarshalDynamoDBItem(t *testing.T) {
	tests := []struct {
		name            string
		input           map[string]any
		wantStandardMap map[string]any
		wantKeyTypeMap  map[string]neosync_types.KeyType
	}{
		{
			name: "Basic test",
			input: map[string]any{
				"StringKey": map[string]any{"S": "value"},
				"NumberKey": map[string]any{"N": "123"},
				"BoolKey":   map[string]any{"BOOL": true},
			},
			wantStandardMap: map[string]any{
				"StringKey": "value",
				"NumberKey": int64(123),
				"BoolKey":   true,
			},
			wantKeyTypeMap: map[string]neosync_types.KeyType{},
		},
		{
			name: "Complex test with sets",
			input: map[string]any{
				"StringSetKey": map[string]any{"SS": []any{"a", "b", "c"}},
				"NumberSetKey": map[string]any{"NS": []any{"1", "2", "3"}},
			},
			wantStandardMap: map[string]any{
				"StringSetKey": []string{"a", "b", "c"},
				"NumberSetKey": []any{int64(1), int64(2), int64(3)},
			},
			wantKeyTypeMap: map[string]neosync_types.KeyType{
				"StringSetKey": neosync_types.StringSet,
				"NumberSetKey": neosync_types.NumberSet,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStandardMap, gotKeyTypeMap := UnmarshalDynamoDBItem(tt.input)
			require.Equal(t, tt.wantStandardMap, gotStandardMap)
			require.Equal(t, tt.wantKeyTypeMap, gotKeyTypeMap)
		})
	}
}

func Test_ParseDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       any
		keyTypeMap  map[string]neosync_types.KeyType
		want        any
		wantKeyType neosync_types.KeyType
	}{
		{"String", "StrKey", map[string]any{"S": "value"}, map[string]neosync_types.KeyType{}, "value", 0},
		{"Number", "NumKey", map[string]any{"N": "123"}, map[string]neosync_types.KeyType{}, int64(123), 0},
		{"Boolean", "BoolKey", map[string]any{"BOOL": true}, map[string]neosync_types.KeyType{}, true, 0},
		{"Null", "NullKey", map[string]any{"NULL": true}, map[string]neosync_types.KeyType{}, nil, 0},
		{"StringSet", "SSKey", map[string]any{"SS": []any{"a", "b"}}, map[string]neosync_types.KeyType{}, []string{"a", "b"}, neosync_types.StringSet},
		{"NumberSet", "NSKey", map[string]any{"NS": []any{"1", "2"}}, map[string]neosync_types.KeyType{}, []any{int64(1), int64(2)}, neosync_types.NumberSet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDynamoDBAttributeValue(tt.key, tt.value, tt.keyTypeMap)
			require.Equalf(t, tt.want, got, fmt.Sprintf("ParseDynamoDBAttributeValue() = %v, want %v", got, tt.want))
			if gotKeyType, ok := tt.keyTypeMap[tt.key]; ok {
				require.Equalf(t, tt.wantKeyType, gotKeyType, fmt.Sprintf("ParseDynamoDBAttributeValue() key type = %v, want %v", gotKeyType, tt.wantKeyType))
			}
		})
	}
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
			got := MarshalToAttributeValue(tt.key, tt.root, tt.keyTypeMap)
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
			got, err := ConvertToStringSlice(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equalf(t, tt.want, got, fmt.Sprintf("ConvertToStringSlice() = %v, want %v", got, tt.want))
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
