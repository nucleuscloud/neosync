package mongomanager

import (
	"math/big"
	"testing"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_UnmarshalPrimitives(t *testing.T) {
	input := map[string]any{
		"string": "test",
		"int":    42,
		"bool":   true,
	}
	expectedOutput := map[string]any{
		"string": "test",
		"int":    42,
		"bool":   true,
	}
	expectedKTM := map[string]neosync_types.KeyType{}

	t.Run("Basic types", func(t *testing.T) {
		output, ktm := UnmarshalPrimitives(input)
		require.Equal(t, expectedOutput, output)
		require.Equal(t, expectedKTM, ktm)
	})

	dec128 := primitive.NewDecimal128(3, 14159)
	objectId := primitive.NewObjectID()
	input = map[string]any{
		"decimal":   dec128,
		"binary":    primitive.Binary{Data: []byte("test")},
		"objectID":  objectId,
		"timestamp": primitive.Timestamp{T: 1, I: 1},
	}
	expectedOutput = map[string]any{
		"decimal":   getBigFloat(dec128.String()),
		"binary":    primitive.Binary{Data: []byte("test")},
		"objectID":  objectId,
		"timestamp": primitive.Timestamp{T: 1, I: 1},
	}
	expectedKTM = map[string]neosync_types.KeyType{
		"decimal":   neosync_types.Decimal128,
		"binary":    neosync_types.Binary,
		"objectID":  neosync_types.ObjectID,
		"timestamp": neosync_types.Timestamp,
	}

	t.Run("BSON types", func(t *testing.T) {
		output, ktm := UnmarshalPrimitives(input)
		require.Equal(t, expectedOutput, output)
		require.Equal(t, expectedKTM, ktm)
	})
}

func getBigFloat(v string) *big.Float {
	f, _, _ := big.ParseFloat(v, 10, 128, big.ToNearestEven)
	return f
}

func Test_ParsePrimitives(t *testing.T) {
	objectId := primitive.NewObjectID()
	dec128 := primitive.NewDecimal128(3, 14159)
	testCases := []struct {
		name        string
		key         string
		value       any
		expectedKTM map[string]neosync_types.KeyType
		expected    any
	}{
		{
			name:        "Decimal128",
			key:         "decimal",
			value:       dec128,
			expectedKTM: map[string]neosync_types.KeyType{"decimal": neosync_types.Decimal128},
			expected:    getBigFloat(dec128.String()),
		},
		{
			name:        "Binary",
			key:         "binary",
			value:       primitive.Binary{Data: []byte("test")},
			expectedKTM: map[string]neosync_types.KeyType{"binary": neosync_types.Binary},
			expected:    primitive.Binary{Data: []byte("test")},
		},
		{
			name:        "ObjectID",
			key:         "objectID",
			value:       objectId,
			expectedKTM: map[string]neosync_types.KeyType{"objectID": neosync_types.ObjectID},
			expected:    objectId,
		},
		{
			name:        "Timestamp",
			key:         "timestamp",
			value:       primitive.Timestamp{T: 1, I: 1},
			expectedKTM: map[string]neosync_types.KeyType{"timestamp": neosync_types.Timestamp},
			expected:    primitive.Timestamp{T: 1, I: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ktm := make(map[string]neosync_types.KeyType)
			result := ParsePrimitives(tc.key, tc.value, ktm)
			require.Equal(t, tc.expectedKTM, ktm)
			require.Equal(t, tc.expected, result)
		})
	}
}

func Test_MarshalToBSONValue(t *testing.T) {
	objId, _ := primitive.ObjectIDFromHex("5f63e6f0d51b0d0001c1b0a1")
	dec128, _ := primitive.ParseDecimal128("123.45")
	testCases := []struct {
		name        string
		key         string
		value       any
		keyTypeMap  map[string]neosync_types.KeyType
		expected    any
		expectError bool
	}{
		{
			name:        "String to ObjectID",
			key:         "id",
			value:       "5f63e6f0d51b0d0001c1b0a1",
			keyTypeMap:  map[string]neosync_types.KeyType{"id": neosync_types.ObjectID},
			expected:    objId,
			expectError: false,
		},
		{
			name:        "Float64 to Decimal128",
			key:         "amount",
			value:       getBigFloat(dec128.String()),
			keyTypeMap:  map[string]neosync_types.KeyType{"amount": neosync_types.Decimal128},
			expected:    dec128,
			expectError: false,
		},
		{
			name:        "Int to Timestamp",
			key:         "timestamp",
			value:       int(1630000000),
			keyTypeMap:  map[string]neosync_types.KeyType{"timestamp": neosync_types.Timestamp},
			expected:    primitive.Timestamp{T: uint32(1630000000), I: 1},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MarshalToBSONValue(tc.key, tc.value, tc.keyTypeMap)
			require.Equal(t, tc.expected, result)
		})
	}
}

func Test_MarshalJSONToBSONDocument(t *testing.T) {
	objId, _ := primitive.ObjectIDFromHex("5f63e6f0d51b0d0001c1b0a1")
	dec128, _ := primitive.ParseDecimal128("123.45")
	testCases := []struct {
		name        string
		input       any
		keyTypeMap  map[string]neosync_types.KeyType
		expected    bson.M
		expectError bool
	}{
		{
			name: "Basic JSON",
			input: map[string]any{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
			},
			keyTypeMap: map[string]neosync_types.KeyType{},
			expected: bson.M{
				"name":  bson.E{Key: "name", Value: "John Doe"},
				"age":   bson.E{Key: "age", Value: 30},
				"email": bson.E{Key: "email", Value: "john@example.com"},
			},
			expectError: false,
		},
		{
			name: "JSON with BSON types",
			input: map[string]any{
				"id":        "5f63e6f0d51b0d0001c1b0a1",
				"amount":    getBigFloat(dec128.String()),
				"timestamp": 1630000000,
			},
			keyTypeMap: map[string]neosync_types.KeyType{
				"id":        neosync_types.ObjectID,
				"amount":    neosync_types.Decimal128,
				"timestamp": neosync_types.Timestamp,
			},
			expected: bson.M{
				"id":        bson.E{Key: "id", Value: objId},
				"amount":    bson.E{Key: "amount", Value: dec128},
				"timestamp": bson.E{Key: "timestamp", Value: primitive.Timestamp{T: 1630000000, I: 1}},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MarshalJSONToBSONDocument(tc.input, tc.keyTypeMap)

			require.Len(t, result, len(tc.expected))
			for _, elem := range result {
				expectedElem := tc.expected[elem.Key]
				require.Equal(t, expectedElem, elem)
			}
		})
	}
}

func TestToUint32(t *testing.T) {
	testCases := []struct {
		name        string
		input       any
		expected    uint32
		expectError bool
	}{
		{"Valid int", 42, 42, false},
		{"Valid float32", float32(42.0), 42, false},
		{"Valid float64", 42.0, 42, false},
		{"Negative int", -1, 0, true},
		{"Out of range uint64", uint64(1 << 33), 0, true},
		{"String", "42", 42, false},
		{"String", "othor", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toUint32(tc.input)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}
