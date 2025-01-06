package mongodb

import (
	"math/big"
	"testing"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	"github.com/stretchr/testify/require"
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

	mapper := NewMongoBuilder()

	t.Run("Basic types", func(t *testing.T) {
		output, ktm, err := mapper.MapRecordWithKeyType(input)
		require.NoError(t, err)
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
		output, ktm, err := mapper.MapRecordWithKeyType(input)
		require.NoError(t, err)
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
			result := parsePrimitives(tc.key, tc.value, ktm)
			require.Equal(t, tc.expectedKTM, ktm)
			require.Equal(t, tc.expected, result)
		})
	}
}
