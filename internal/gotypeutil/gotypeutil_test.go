package gotypeutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsMultiDimensionalArray(t *testing.T) {
	t.Run("MultiDimensionalArray", func(t *testing.T) {
		input := []any{[]any{1, 2}, []any{3, 4}}
		result := IsMultiDimensionalSlice(input)
		require.True(t, result, "Expected true for multi-dimensional array")
	})

	t.Run("MultiDimensionalArray", func(t *testing.T) {
		input := [][]any{{1, 2}, {}}
		result := IsMultiDimensionalSlice(input)
		require.True(t, result, "Expected true for multi-dimensional array")
	})

	t.Run("SingleDimensionalArray", func(t *testing.T) {
		input := []any{1, 2, 3, 4}
		result := IsMultiDimensionalSlice(input)
		require.False(t, result, "Expected false for single-dimensional array")
	})

	t.Run("EmptyArray", func(t *testing.T) {
		input := []any{}
		result := IsMultiDimensionalSlice(input)
		require.False(t, result, "Expected false for empty array")
	})

	t.Run("NonArrayType", func(t *testing.T) {
		input := "not an array"
		result := IsMultiDimensionalSlice(input)
		require.False(t, result, "Expected false for non-array type")
	})

	t.Run("ArrayWithNonArrayElements", func(t *testing.T) {
		input := []any{1, "string", true}
		result := IsMultiDimensionalSlice(input)
		require.False(t, result, "Expected false for array with non-array elements")
	})

	t.Run("NestedEmptyArray", func(t *testing.T) {
		input := []any{[]any{1, 2}, []any{}}
		result := IsMultiDimensionalSlice(input)
		require.True(t, result, "Expected true for array containing an empty array")
	})
}
