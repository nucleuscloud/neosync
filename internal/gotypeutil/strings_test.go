package gotypeutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CaseInsensitiveContains(t *testing.T) {
	t.Run("CaseInsensitiveContains", func(t *testing.T) {
		require.True(t, CaseInsensitiveContains("Hello, World!", "hello"), "Should find lowercase substring")
		require.True(t, CaseInsensitiveContains("Hello, World!", "WORLD"), "Should find uppercase substring")
		require.True(t, CaseInsensitiveContains("Hello, World!", "o, wo"), "Should find mixed case substring")
		require.True(t, CaseInsensitiveContains("Hello, World!", ""), "Should return true for empty substring")
		require.True(t, CaseInsensitiveContains("Hello, World!", "Hello, World!"), "Should find when substring is equal to string")

		require.False(t, CaseInsensitiveContains("Hello, World!", "goodbye"), "Should not find non-existent substring")
		require.False(t, CaseInsensitiveContains("", "test"), "Should return false when string is empty and substring is not")
		require.False(t, CaseInsensitiveContains("Hello", "Hello, World!"), "Should return false when substring is longer than string")

		require.True(t, CaseInsensitiveContains("HeLLo, WoRLD!", "hello, world!"), "Should handle mixed case in both string and substring")
		require.True(t, CaseInsensitiveContains("HELLO", "hello"), "Should handle all uppercase string and lowercase substring")
		require.True(t, CaseInsensitiveContains("hello", "HELLO"), "Should handle all lowercase string and uppercase substring")
	})
}
