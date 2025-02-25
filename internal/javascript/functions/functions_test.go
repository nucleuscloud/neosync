package javascript_functions

import (
	"context"
	"log/slog"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestNewFunctionDefinition(t *testing.T) {
	namespace := "test"
	name := "fn"
	ctor := func(r Runner) Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			return nil, nil
		}
	}

	fn := NewFunctionDefinition(namespace, name, ctor)
	require.Equal(t, namespace, fn.Namespace())
	require.Equal(t, name, fn.Name())
	require.NotNil(t, fn.Ctor())
}

func TestParseFunctionArguments(t *testing.T) {
	rt := goja.New()

	t.Run("string argument", func(t *testing.T) {
		var str string
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("test")}}
		err := ParseFunctionArguments(call, &str)
		require.NoError(t, err)
		require.Equal(t, "test", str)
	})

	t.Run("int argument", func(t *testing.T) {
		var num int
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(42)}}
		err := ParseFunctionArguments(call, &num)
		require.NoError(t, err)
		require.Equal(t, 42, num)
	})

	t.Run("map arguments", func(t *testing.T) {
		t.Run("simple map", func(t *testing.T) {
			var m map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(map[string]any{"key": "value"})}}
			err := ParseFunctionArguments(call, &m)
			require.NoError(t, err)
			require.Equal(t, "value", m["key"])
		})

		t.Run("nested map", func(t *testing.T) {
			var m map[string]any
			nestedMap := map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  30,
				},
			}
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nestedMap)}}
			err := ParseFunctionArguments(call, &m)
			require.NoError(t, err)
			require.Equal(t, "John", m["user"].(map[string]any)["name"])
			require.Equal(t, 30, m["user"].(map[string]any)["age"])
		})

		t.Run("empty map", func(t *testing.T) {
			var m map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(map[string]any{})}}
			err := ParseFunctionArguments(call, &m)
			require.NoError(t, err)
			require.Empty(t, m)
		})

		t.Run("nil map", func(t *testing.T) {
			var m map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nil)}}
			err := ParseFunctionArguments(call, &m)
			require.NoError(t, err)
			require.Empty(t, m)
		})
	})

	t.Run("undefined argument", func(t *testing.T) {
		var str string
		call := goja.FunctionCall{Arguments: []goja.Value{goja.Undefined()}}
		err := ParseFunctionArguments(call, &str)
		require.Error(t, err)
		require.Contains(t, err.Error(), "undefined")
	})

	t.Run("too many arguments", func(t *testing.T) {
		var str string
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("test"), rt.ToValue("extra")}}
		err := ParseFunctionArguments(call, &str)
		require.Error(t, err)
		require.Contains(t, err.Error(), "have 2 arguments, but only 1 pointers")
	})
	t.Run("int64 argument", func(t *testing.T) {
		var num int64
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(42)}}
		err := ParseFunctionArguments(call, &num)
		require.NoError(t, err)
		require.Equal(t, int64(42), num)
	})

	t.Run("float64 argument", func(t *testing.T) {
		var num float64
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(42.5)}}
		err := ParseFunctionArguments(call, &num)
		require.NoError(t, err)
		require.Equal(t, 42.5, num)
	})

	t.Run("bool argument", func(t *testing.T) {
		var b bool
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(true)}}
		err := ParseFunctionArguments(call, &b)
		require.NoError(t, err)
		require.True(t, b)
	})

	t.Run("slice argument", func(t *testing.T) {
		var s []any
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue([]any{"test", 42})}}
		err := ParseFunctionArguments(call, &s)
		require.NoError(t, err)
		require.Equal(t, []any{"test", int(42)}, s)
	})

	t.Run("map slice argument", func(t *testing.T) {
		t.Run("valid map slice", func(t *testing.T) {
			var ms []map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue([]map[string]any{{"key": "value"}})}}
			err := ParseFunctionArguments(call, &ms)
			require.NoError(t, err)
			require.Equal(t, []map[string]any{{"key": "value"}}, ms)
		})

		t.Run("from slice of any", func(t *testing.T) {
			var ms []map[string]any
			// Create a []any containing map[string]any elements
			sliceOfAny := []any{map[string]any{"name": "John"}, map[string]any{"name": "Alice"}}
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(sliceOfAny)}}
			err := ParseFunctionArguments(call, &ms)
			require.NoError(t, err)
			require.Equal(t, []map[string]any{{"name": "John"}, {"name": "Alice"}}, ms)
		})

		t.Run("with nil value", func(t *testing.T) {
			var ms []map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nil)}}
			err := ParseFunctionArguments(call, &ms)
			require.NoError(t, err)
			require.Equal(t, []map[string]any{}, ms)
		})

		t.Run("with non-slice value", func(t *testing.T) {
			var ms []map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("not a slice")}}
			err := ParseFunctionArguments(call, &ms)
			require.Error(t, err)
			require.Contains(t, err.Error(), "value is not of type map slice")
		})

		t.Run("with slice containing non-map elements", func(t *testing.T) {
			var ms []map[string]any
			// Create a []any containing a string, which isn't a map[string]any
			sliceWithNonMap := []any{map[string]any{"valid": true}, "not a map"}
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(sliceWithNonMap)}}
			err := ParseFunctionArguments(call, &ms)
			require.Error(t, err)
			require.Contains(t, err.Error(), "value is not of type map slice")
		})
	})

	t.Run("goja.Value argument", func(t *testing.T) {
		var v goja.Value
		expected := rt.ToValue("test")
		call := goja.FunctionCall{Arguments: []goja.Value{expected}}
		err := ParseFunctionArguments(call, &v)
		require.NoError(t, err)
		require.Equal(t, expected, v)
	})

	t.Run("any argument", func(t *testing.T) {
		var a any
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("test")}}
		err := ParseFunctionArguments(call, &a)
		require.NoError(t, err)
		require.Equal(t, "test", a)
	})

	t.Run("unhandled type argument", func(t *testing.T) {
		var c chan int
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("test")}}
		err := ParseFunctionArguments(call, &c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unhandled type")
	})

	t.Run("null value with type error", func(t *testing.T) {
		var s []string // This expects a specific slice type, not just any slice
		call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nil)}}
		err := ParseFunctionArguments(call, &s)
		require.Error(t, err)
		require.Contains(t, err.Error(), "null")
		// Update the expected error message to match what's actually returned
		require.Contains(t, err.Error(), "encountered unhandled type null while trying to parse")
	})

	t.Run("error handling", func(t *testing.T) {
		t.Run("null value with specific type error", func(t *testing.T) {
			var ms []map[string]any
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nil)}}
			err := ParseFunctionArguments(call, &ms)
			require.NoError(t, err) // This should succeed with empty slice
		})

		t.Run("null value with string slice type", func(t *testing.T) {
			var s []string // This expects a specific slice type, not just any slice
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nil)}}
			err := ParseFunctionArguments(call, &s)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unhandled type")
			require.Contains(t, err.Error(), "null")
		})

		t.Run("null value with unhandled type", func(t *testing.T) {
			var c chan int // An unhandled type
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(nil)}}
			err := ParseFunctionArguments(call, &c)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unhandled type")
			require.Contains(t, err.Error(), "null")
		})

		t.Run("non-null value with unhandled type", func(t *testing.T) {
			var c chan int // An unhandled type
			call := goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("test")}}
			err := ParseFunctionArguments(call, &c)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unhandled type")
		})
	})
}

func TestGetTypeString(t *testing.T) {
	rt := goja.New()

	t.Run("null value", func(t *testing.T) {
		nullValue := rt.ToValue(nil)
		typeStr := getTypeString(nullValue)
		require.Equal(t, "null", typeStr)
	})

	t.Run("undefined value", func(t *testing.T) {
		undefinedValue := goja.Undefined()
		typeStr := getTypeString(undefinedValue)
		require.Equal(t, "undefined", typeStr)
	})

	t.Run("string value", func(t *testing.T) {
		stringValue := rt.ToValue("test")
		typeStr := getTypeString(stringValue)
		require.Equal(t, "string", typeStr)
	})

	t.Run("number value", func(t *testing.T) {
		numberValue := rt.ToValue(42)
		typeStr := getTypeString(numberValue)
		require.Equal(t, "int64", typeStr)
	})

	t.Run("object value", func(t *testing.T) {
		objectValue := rt.ToValue(map[string]any{"key": "value"})
		typeStr := getTypeString(objectValue)
		require.Equal(t, "map[string]interface {}", typeStr)
	})
}
