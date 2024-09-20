package sqlscanners

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_BitString_Scan(t *testing.T) {
	t.Run("Nil Value", func(t *testing.T) {
		var b BitString
		err := b.Scan(nil)
		require.NoError(t, err)
		require.False(t, b.IsValid)
	})

	t.Run("[]byte BIT(1) to BIT(8)", func(t *testing.T) {
		var b BitString
		err := b.Scan([]byte{5})
		require.NoError(t, err)
		require.True(t, b.IsValid)
		require.Equal(t, []byte{5}, b.Bytes)
		require.Equal(t, "101", b.BitString)
	})

	t.Run("[]byte BIT(9) to BIT(64)", func(t *testing.T) {
		var b BitString
		err := b.Scan([]byte{1, 2})
		require.NoError(t, err)
		require.True(t, b.IsValid)
		require.Equal(t, []byte{1, 2}, b.Bytes)
		require.Equal(t, "100000010", b.BitString)
	})

	t.Run("int64", func(t *testing.T) {
		var b BitString
		err := b.Scan(int64(5))
		require.NoError(t, err)
		require.True(t, b.IsValid)
		require.Equal(t, []byte{5}, b.Bytes)
		require.Equal(t, "101", b.BitString)
	})

	t.Run("uint64", func(t *testing.T) {
		var b BitString
		err := b.Scan(uint64(5))
		require.NoError(t, err)
		require.True(t, b.IsValid)
		require.Equal(t, []byte{5}, b.Bytes)
		require.Equal(t, "101", b.BitString)
	})

	t.Run("string", func(t *testing.T) {
		var b BitString
		err := b.Scan("101")
		require.NoError(t, err)
		require.True(t, b.IsValid)
		require.Equal(t, []byte{5}, b.Bytes)
		require.Equal(t, "101", b.BitString)
	})

	t.Run("Invalid Type", func(t *testing.T) {
		var b BitString
		err := b.Scan(3.14)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot scan type float64 into BitString")
	})

	t.Run("Invalid Binary String", func(t *testing.T) {
		var b BitString
		err := b.Scan("12")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid binary string")
	})
}

func Test_BitString_Value(t *testing.T) {
	t.Run("Valid BitString", func(t *testing.T) {
		b := BitString{IsValid: true, Bytes: []byte{5}, BitString: "101"}
		v, err := b.Value()
		require.NoError(t, err)
		require.Equal(t, driver.Value([]byte{5}), v)
	})

	t.Run("Invalid BitString", func(t *testing.T) {
		b := BitString{IsValid: false}
		v, err := b.Value()
		require.NoError(t, err)
		require.Nil(t, v)
	})
}

func Test_BitString_String(t *testing.T) {
	t.Run("Valid BitString", func(t *testing.T) {
		b := BitString{IsValid: true, BitString: "101"}
		require.Equal(t, "101", b.String())
	})

	t.Run("Invalid BitString", func(t *testing.T) {
		b := BitString{IsValid: false}
		require.Equal(t, "", b.String())
	})
}
