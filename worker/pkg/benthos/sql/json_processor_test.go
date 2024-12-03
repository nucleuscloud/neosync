package neosync_benthos_sql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_convertStringToBit(t *testing.T) {
	t.Run("8 bits", func(t *testing.T) {
		got, err := convertStringToBit("10101010")
		require.NoError(t, err)
		expected := []byte{170}
		require.Equalf(t, expected, got, "got %v, want %v", got, expected)
	})

	t.Run("1 bit", func(t *testing.T) {
		got, err := convertStringToBit("1")
		require.NoError(t, err)
		expected := []byte{1}
		require.Equalf(t, expected, got, "got %v, want %v", got, expected)
	})

	t.Run("16 bits", func(t *testing.T) {
		got, err := convertStringToBit("1010101010101010")
		require.NoError(t, err)
		expected := []byte{170, 170}
		require.Equalf(t, expected, got, "got %v, want %v", got, expected)
	})

	t.Run("24 bits", func(t *testing.T) {
		got, err := convertStringToBit("101010101111111100000000")
		require.NoError(t, err)
		expected := []byte{170, 255, 0}
		require.Equalf(t, expected, got, "got %v, want %v", got, expected)
	})

	t.Run("invalid binary string", func(t *testing.T) {
		_, err := convertStringToBit("102")
		require.Error(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := convertStringToBit("")
		require.Error(t, err)
	})
}
