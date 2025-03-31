package neosync_functions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	functions, err := Get(nil)
	require.NoError(t, err)
	require.NotEmpty(t, functions)
}

func Test_setNestedProperty(t *testing.T) {
	t.Run("Set simple property", func(t *testing.T) {
		obj := make(map[string]any)
		setNestedProperty(obj, "name", "John")
		require.Equal(t, "John", obj["name"])
	})

	t.Run("Set nested property", func(t *testing.T) {
		obj := make(map[string]any)
		setNestedProperty(obj, "user.name", "Alice")
		require.Equal(t, "Alice", obj["user"].(map[string]any)["name"])
	})

	t.Run("Set deeply nested property", func(t *testing.T) {
		obj := make(map[string]any)
		setNestedProperty(obj, "user.profile.age", 30)
		require.Equal(t, 30, obj["user"].(map[string]any)["profile"].(map[string]any)["age"])
	})

	t.Run("Override existing property", func(t *testing.T) {
		obj := map[string]any{"user": map[string]any{"name": "Bob"}}
		setNestedProperty(obj, "user.name", "Charlie")
		require.Equal(t, "Charlie", obj["user"].(map[string]any)["name"])
	})

	t.Run("Set property in existing nested structure", func(t *testing.T) {
		obj := map[string]any{"user": map[string]any{"profile": map[string]any{}}}
		setNestedProperty(obj, "user.profile.email", "test@example.com")
		require.Equal(t, "test@example.com", obj["user"].(map[string]any)["profile"].(map[string]any)["email"])
	})

	t.Run("Set nestnd property in empty structure", func(t *testing.T) {
		obj := map[string]any{}
		setNestedProperty(obj, "user.profile.email", "test@example.com")
		require.Equal(t, "test@example.com", obj["user"].(map[string]any)["profile"].(map[string]any)["email"])
	})
}
