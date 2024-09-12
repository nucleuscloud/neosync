package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CreateMultiDimSlice(t *testing.T) {
	t.Run("1D Slice of Ints", func(t *testing.T) {
		dims := []int{3}
		elements := []any{1, 2, 3}
		result := CreateMultiDimSlice(dims, elements)
		expected := []any{1, 2, 3}
		require.Equal(t, expected, result)
	})

	t.Run("2D Slice of Ints", func(t *testing.T) {
		dims := []int{2, 3}
		elements := []any{1, 2, 3, 4, 5, 6}
		result := CreateMultiDimSlice(dims, elements)
		expected := [][]any{{1, 2, 3}, {4, 5, 6}}
		require.Equal(t, expected, result)
	})

	t.Run("3D Slice of Ints", func(t *testing.T) {
		dims := []int{2, 2, 2}
		elements := []any{1, 2, 3, 4, 5, 6, 7, 8}
		result := CreateMultiDimSlice(dims, elements)
		expected := [][][]any{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}
		require.Equal(t, expected, result)
	})

	t.Run("2D Slice of Strings", func(t *testing.T) {
		dims := []int{2, 2}
		elements := []any{"a", "b", "c", "d"}
		result := CreateMultiDimSlice(dims, elements)
		expected := [][]any{{"a", "b"}, {"c", "d"}}
		require.Equal(t, expected, result)
	})

	t.Run("Empty Dims", func(t *testing.T) {
		dims := []int{}
		elements := []any{42}
		result := CreateMultiDimSlice(dims, elements)
		require.Equal(t, []any{42}, result)
	})

	t.Run("1D Slice with Single Element", func(t *testing.T) {
		dims := []int{1}
		elements := []any{42}
		result := CreateMultiDimSlice(dims, elements)
		expected := []any{42}
		require.Equal(t, expected, result)
	})

	t.Run("3D Slice with Mixed Types", func(t *testing.T) {
		dims := []int{2, 2, 2}
		elements := []any{1, "a", true, 3.14, 0, 'b', []int{1, 2}, map[string]int{"x": 1}}
		result := CreateMultiDimSlice(dims, elements)
		expected := [][][]any{
			{{1, "a"}, {true, 3.14}},
			{{0, 'b'}, {[]int{1, 2}, map[string]int{"x": 1}}},
		}
		require.Equal(t, expected, result)
	})
}

func Test_FormatPgArrayLiteral(t *testing.T) {
	t.Run("Empty array", func(t *testing.T) {
		input := []any{}
		expected := "ARRAY[]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "Empty array should be formatted correctly")
	})

	t.Run("1D array of integers", func(t *testing.T) {
		input := []any{1, 2, 3}
		expected := "ARRAY[1,2,3]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "1D array of integers should be formatted correctly")
	})

	t.Run("1D array of strings", func(t *testing.T) {
		input := []any{"a", "b", "c"}
		expected := "ARRAY['a','b','c']"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "1D array of strings should be formatted correctly")
	})

	t.Run("2D array of integers", func(t *testing.T) {
		input := []any{[]any{1, 2}, []any{3, 4}}
		expected := "ARRAY[[1,2],[3,4]]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "2D array of integers should be formatted correctly")
	})
	t.Run("2D array of integers", func(t *testing.T) {
		input := [][]any{{1, 2}, {3, 4}}
		expected := "ARRAY[[1,2],[3,4]]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "2D array of integers should be formatted correctly")
	})
	t.Run("4D array of integers", func(t *testing.T) {
		input := [][][]any{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}
		expected := "ARRAY[[[1,2],[3,4]],[[5,6],[7,8]]]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "2D array of integers should be formatted correctly")
	})

	t.Run("3D array of integers", func(t *testing.T) {
		input := []any{[]any{[]any{1, 2}, []any{3, 4}}, []any{[]any{5, 6}, []any{7, 8}}}
		expected := "ARRAY[[[1,2],[3,4]],[[5,6],[7,8]]]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "3D array of integers should be formatted correctly")
	})

	t.Run("Mixed types array", func(t *testing.T) {
		input := []any{1, "a", true, 3.14}
		expected := "ARRAY[1,'a',true,3.14]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "Array with mixed types should be formatted correctly")
	})

	t.Run("Array with nested mixed types", func(t *testing.T) {
		input := []any{[]any{1, "a"}, []any{true, 3.14}}
		expected := "ARRAY[[1,'a'],[true,3.14]]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "Array with nested mixed types should be formatted correctly")
	})

	t.Run("Array with null values", func(t *testing.T) {
		input := []any{1, nil, 3}
		expected := "ARRAY[1,<nil>,3]"
		result := FormatPgArrayLiteral(input, "")
		require.Equal(t, expected, result, "Array with null values should be formatted correctly")
	})

	// maps
	t.Run("Array of key-value pairs", func(t *testing.T) {
		input := []any{
			map[string]any{"age": 30, "city": "New York"},
			map[string]any{"age": 25, "city": "San Francisco"},
		}
		expected := `ARRAY['{"age":30,"city":"New York"}','{"age":25,"city":"San Francisco"}']::json[]`
		result := FormatPgArrayLiteral(input, "json[]")
		require.Equal(t, expected, result, "Array of key-value pairs should be formatted correctly")
	})

	t.Run("Array with nested key-value pairs", func(t *testing.T) {
		input := []any{
			"simple string",
			map[string]any{
				"name": "John",
				"details": map[string]any{
					"age":  "30",
					"city": "New York",
				},
			},
		}
		expected := `ARRAY['simple string','{"details":{"age":"30","city":"New York"},"name":"John"}']::jsonb[]`
		result := FormatPgArrayLiteral(input, "jsonb[]")
		require.Equal(t, expected, result, "Array with nested key-value pairs should be formatted correctly")
	})
}
