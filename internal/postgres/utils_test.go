package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// func TestConvertPgArrayStringToSlice(t *testing.T) {
// 	t.Run("Empty array", func(t *testing.T) {
// 		result := ConvertPgArrayToSlice("{}")
// 		require.Equal(t, []any{}, result)
// 	})

// 	t.Run("Integer array", func(t *testing.T) {
// 		result := ConvertPgArrayToSlice("{1,2,3}")
// 		require.Equal(t, []any{"1", "2", "3"}, result)
// 	})

// 	t.Run("String array", func(t *testing.T) {
// 		result := ConvertPgArrayToSlice("{foo,bar,baz}")
// 		require.Equal(t, []any{"foo", "bar", "baz"}, result)
// 	})

// 	t.Run("Mixed array", func(t *testing.T) {
// 		result := ConvertPgArrayToSlice("{1,foo,2,bar}")
// 		require.Equal(t, []any{"1", "foo", "2", "bar"}, result)
// 	})

// 	t.Run("Non-string input", func(t *testing.T) {
// 		result := ConvertPgArrayToSlice(123)
// 		require.Equal(t, []any{}, result)
// 	})
// 	t.Run("Nested Array", func(t *testing.T) {
// 		result := ConvertPgArrayToSlice("{{1,2,3}, {4,5,6}}")
// 		require.Equal(t, [][]any{{1, 2, 3}, {4, 5, 6}}, result)
// 	})
// }

func TestConvertSliceToPgString(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		result := ConvertSliceToPgString([]any{})
		require.Equal(t, "{}", result)
	})

	t.Run("Integer slice", func(t *testing.T) {
		result := ConvertSliceToPgString([]any{1, 2, 3})
		require.Equal(t, "{1,2,3}", result)
	})

	t.Run("String slice", func(t *testing.T) {
		result := ConvertSliceToPgString([]any{"foo", "bar", "baz"})
		require.Equal(t, "{foo,bar,baz}", result)
	})

	t.Run("Mixed slice", func(t *testing.T) {
		result := ConvertSliceToPgString([]any{1, "foo", 2, "bar"})
		require.Equal(t, "{1,foo,2,bar}", result)
	})
}

func TestToPgTypes(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		result := ToPgTypes([]any{})
		require.Equal(t, []any{}, result)
	})

	t.Run("No slices", func(t *testing.T) {
		result := ToPgTypes([]any{1, "foo", true})
		require.Equal(t, []any{1, "foo", true}, result)
	})

	t.Run("With slices", func(t *testing.T) {
		result := ToPgTypes([]any{1, []any{2, 3}, "foo", []any{"bar", "baz"}})
		require.Equal(t, []any{1, "{2,3}", "foo", "{bar,baz}"}, result)
	})

	t.Run("Mixed types", func(t *testing.T) {
		result := ToPgTypes([]any{1, []any{2, "foo"}, true, []any{3.14, false}})
		require.Equal(t, []any{1, "{2,foo}", true, "{3.14,false}"}, result)
	})
}
