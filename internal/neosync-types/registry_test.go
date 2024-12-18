package neosynctypes

import (
	"testing"

	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
)

func Test_TypeRegistry(t *testing.T) {
	t.Run("NewTypeRegistry initializes with default types", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		require.NotNil(t, registry)
		require.NotNil(t, registry.types)

		interval, err := registry.New(NeosyncIntervalId, LatestVersion)
		require.NoError(t, err)
		require.NotNil(t, interval)

		array, err := registry.New(NeosyncArrayId, LatestVersion)
		require.NoError(t, err)
		require.NotNil(t, array)
	})

	t.Run("Register adds new type successfully", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		customTypeId := "custom-type"
		customVersion := Version(1)

		registry.Register(customTypeId, customVersion, func() (NeosyncAdapter, error) {
			return NewInterval(WithVersion(customVersion))
		})

		obj, err := registry.New(customTypeId, customVersion)
		require.NoError(t, err)
		require.NotNil(t, obj)
	})

	t.Run("New returns error for unknown type", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		unknownTypeId := "unknown-type"

		obj, err := registry.New(unknownTypeId, LatestVersion)
		require.Error(t, err)
		require.Nil(t, obj)
		require.Contains(t, err.Error(), "unknown type ID")
	})

	t.Run("New falls back to latest version when specific version not found", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		nonExistentVersion := Version(999)

		obj, err := registry.New(NeosyncIntervalId, nonExistentVersion)
		require.NoError(t, err)
		require.NotNil(t, obj)
	})

	t.Run("Unmarshal handles invalid JSON", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		invalidJSON := []byte("invalid json")

		result, err := registry.Unmarshal(invalidJSON)
		require.NoError(t, err)
		require.Equal(t, invalidJSON, result)
	})

	t.Run("Unmarshal handles regular JSON without _neosync field", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		regularJSON := []byte(`{"key": "value"}`)

		result, err := registry.Unmarshal(regularJSON)
		require.NoError(t, err)
		require.Equal(t, regularJSON, result)
	})

	t.Run("Unmarshal handles NeosyncInterval type", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		intervalJSON := []byte(`{
			"_neosync": {
				"type_id": "NEOSYNC_INTERVAL",
				"version": 1
			},
			"microseconds": 1,
			"days": 10,
			"months": 0
		}`)

		result, err := registry.Unmarshal(intervalJSON)
		require.NoError(t, err)

		interval, ok := result.(NeosyncAdapter)
		require.True(t, ok)
		require.NotNil(t, interval)
	})

	t.Run("Unmarshal handles NeosyncInterval type in map[string]any", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		intervalMap := map[string]any{
			"_neosync": map[string]any{
				"type_id": "NEOSYNC_INTERVAL",
				"version": 1,
			},
			"microseconds": 1,
			"days":         10,
			"months":       0,
		}

		result, err := registry.Unmarshal(intervalMap)
		require.NoError(t, err)

		interval, ok := result.(NeosyncAdapter)
		require.True(t, ok)
		require.NotNil(t, interval)
	})

	t.Run("Unmarshal handles NeosyncArray type", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		arrayJSON := []byte(`{
			"_neosync": {
				"type_id": "NEOSYNC_ARRAY",
				"version": 1
			},
			"elements": [
				{
					"_neosync": {
						"type_id": "NEOSYNC_INTERVAL",
						"version": 1
					},
					"microseconds": 1,
					"days": 10,
					"months": 0
				},
				{
					"_neosync": {
						"type_id": "NEOSYNC_INTERVAL",
						"version": 1
					},
					"microseconds": 11,
					"days": 20,
					"months": 0
				}
			]
		}`)

		result, err := registry.Unmarshal(arrayJSON)
		require.NoError(t, err)

		array, ok := result.(*NeosyncArray)
		require.True(t, ok)
		require.NotNil(t, array)
		require.Len(t, array.Elements, 2)

		for _, element := range array.Elements {
			require.NotNil(t, element)
			_, ok := element.(NeosyncAdapter)
			require.True(t, ok)
		}
	})

	t.Run("Unmarshal returns error for invalid array elements", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		invalidArrayJSON := []byte(`{
			"_neosync": {
				"type_id": "NEOSYNC_ARRAY",
				"version": 1
			},
			"elements": "not-an-array"
		}`)

		result, err := registry.Unmarshal(invalidArrayJSON)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "invalid array elements")
	})

	t.Run("Unmarshal returns error for non-NeosyncAdapter array elements", func(t *testing.T) {
		registry := NewTypeRegistry(testutil.GetTestLogger(t))
		invalidElementJSON := []byte(`{
			"_neosync": {
				"type_id": "NEOSYNC_ARRAY",
				"version": 1
			},
			"elements": [
				{"not": "a neosync adapter"}
			]
		}`)

		result, err := registry.Unmarshal(invalidElementJSON)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "not a NeosyncAdapter")
	})
}
