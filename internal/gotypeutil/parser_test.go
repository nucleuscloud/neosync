package gotypeutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseStringAsNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
	}{
		{"Valid int", "123", int64(123), false},
		{"Valid float", "123.45", float64(123.45), false},
		{"Invalid number", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStringAsNumber(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_MapToJson(t *testing.T) {
	t.Run("Simple map", func(t *testing.T) {
		m := map[string]any{"name": "John", "age": 30}
		expected := []byte(`{"age":30,"name":"John"}`)

		result, err := MapToJson(m)
		require.NoError(t, err)
		require.JSONEq(t, string(expected), string(result))
	})

	t.Run("Empty map", func(t *testing.T) {
		m := map[string]any{}
		expected := []byte(`{}`)

		result, err := MapToJson(m)
		require.NoError(t, err)
		require.JSONEq(t, string(expected), string(result))
	})

	t.Run("Nested map", func(t *testing.T) {
		m := map[string]any{
			"person": map[string]any{
				"name": "Alice",
				"age":  25,
			},
			"city": "New York",
		}
		expected := []byte(`{"city":"New York","person":{"age":25,"name":"Alice"}}`)

		result, err := MapToJson(m)
		require.NoError(t, err)
		require.JSONEq(t, string(expected), string(result))
	})
}

func Test_JsonToMap(t *testing.T) {
	t.Run("Simple JSON", func(t *testing.T) {
		j := []byte(`{"name":"John","age":30}`)
		expected := map[string]any{"name": "John", "age": float64(30)}

		result, err := JsonToMap(j)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("Empty JSON", func(t *testing.T) {
		j := []byte(`{}`)
		expected := map[string]any{}

		result, err := JsonToMap(j)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("Nested JSON", func(t *testing.T) {
		j := []byte(`{"person":{"name":"Alice","age":25},"city":"New York"}`)
		expected := map[string]any{
			"person": map[string]any{
				"name": "Alice",
				"age":  float64(25),
			},
			"city": "New York",
		}

		result, err := JsonToMap(j)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		j := []byte(`{"name":"John","age":}`)

		_, err := JsonToMap(j)
		require.Error(t, err)
	})
}
