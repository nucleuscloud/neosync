package transformer_utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsZeroValue(t *testing.T) {
	t.Run("Primitive Types", func(t *testing.T) {
		t.Run("Integer", func(t *testing.T) {
			require.True(t, IsZeroValue(0))
			require.False(t, IsZeroValue(1))
			require.False(t, IsZeroValue(-1))
		})

		t.Run("Float", func(t *testing.T) {
			require.True(t, IsZeroValue(0.0))
			require.False(t, IsZeroValue(0.1))
			require.False(t, IsZeroValue(-0.1))
		})

		t.Run("Boolean", func(t *testing.T) {
			require.True(t, IsZeroValue(false))
			require.False(t, IsZeroValue(true))
		})

		t.Run("String", func(t *testing.T) {
			require.True(t, IsZeroValue(""))
			require.False(t, IsZeroValue("hello"))
		})
	})

	t.Run("Complex Types", func(t *testing.T) {
		t.Run("Slice", func(t *testing.T) {
			require.True(t, IsZeroValue([]int(nil)))
			require.True(t, IsZeroValue([]int{}))
			require.False(t, IsZeroValue([]int{1, 2, 3}))
		})

		t.Run("Map", func(t *testing.T) {
			require.True(t, IsZeroValue(map[string]int(nil)))
			require.True(t, IsZeroValue(map[string]int{}))
			require.False(t, IsZeroValue(map[string]int{"a": 1}))
		})

		t.Run("Struct", func(t *testing.T) {
			type TestStruct struct {
				A int
				B string
			}
			require.True(t, IsZeroValue(TestStruct{}))
			require.False(t, IsZeroValue(TestStruct{A: 1}))
			require.False(t, IsZeroValue(TestStruct{B: "hello"}))
		})

		t.Run("Pointer", func(t *testing.T) {
			var nilPtr *int
			value := 5
			ptr := &value
			require.True(t, IsZeroValue(nilPtr))
			require.False(t, IsZeroValue(ptr))
		})
	})

	t.Run("Custom Types", func(t *testing.T) {
		type CustomInt int
		require.True(t, IsZeroValue(CustomInt(0)))
		require.False(t, IsZeroValue(CustomInt(1)))

		type CustomStruct struct {
			Value int
		}
		require.True(t, IsZeroValue(CustomStruct{}))
		require.False(t, IsZeroValue(CustomStruct{Value: 1}))
	})
}
