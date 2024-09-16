package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_MyDate_Scan(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var d MyDate
		err := d.Scan(nil)
		require.NoError(t, err)
		require.True(t, d.Time.IsZero())
	})

	t.Run("ByteSlice", func(t *testing.T) {
		var d MyDate
		err := d.Scan([]byte("2023-04-15"))
		require.NoError(t, err)
		require.Equal(t, time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC), d.Time)
	})

	t.Run("String", func(t *testing.T) {
		var d MyDate
		err := d.Scan("2023-04-15 10:30:45")
		require.NoError(t, err)
		require.Equal(t, time.Date(2023, 4, 15, 10, 30, 45, 0, time.UTC), d.Time)
	})

	t.Run("ZeroDate", func(t *testing.T) {
		var d MyDate
		err := d.Scan("0000-00-00")
		require.NoError(t, err)
		require.True(t, d.Time.IsZero())
	})

	t.Run("ZeroDateTime", func(t *testing.T) {
		var d MyDate
		err := d.Scan("0000-00-00 00:00:00")
		require.NoError(t, err)
		require.True(t, d.Time.IsZero())
	})

	t.Run("InvalidType", func(t *testing.T) {
		var d MyDate
		err := d.Scan(123)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot scan type int into MyDate")
	})

	t.Run("InvalidDateFormat", func(t *testing.T) {
		var d MyDate
		err := d.Scan("2023/04/15")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to parse date: 2023/04/15")
	})

	t.Run("ISO8601Format", func(t *testing.T) {
		var d MyDate
		err := d.Scan("2023-04-15T10:30:45Z")
		require.NoError(t, err)
		require.Equal(t, time.Date(2023, 4, 15, 10, 30, 45, 0, time.UTC), d.Time)
	})

	t.Run("DMYFormat", func(t *testing.T) {
		var d MyDate
		err := d.Scan("15/04/2023")
		require.NoError(t, err)
		require.Equal(t, time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC), d.Time)
	})

	t.Run("DMYWithTimeFormat", func(t *testing.T) {
		var d MyDate
		err := d.Scan("15/04/2023 10:30:45")
		require.NoError(t, err)
		require.Equal(t, time.Date(2023, 4, 15, 10, 30, 45, 0, time.UTC), d.Time)
	})

	t.Run("RFC3339Format", func(t *testing.T) {
		var d MyDate
		err := d.Scan("2023-04-15T10:30:45+00:00")
		require.NoError(t, err)
		require.Equal(t, time.Date(2023, 4, 15, 10, 30, 45, 0, time.UTC), d.Time)
	})
}

func Test_MyDate_Value(t *testing.T) {
	t.Run("NonZeroDate", func(t *testing.T) {
		d := MyDate{Time: time.Date(2023, 4, 15, 10, 30, 45, 0, time.UTC)}
		v, err := d.Value()
		require.NoError(t, err)
		require.Equal(t, "2023-04-15 10:30:45", v)
	})

	t.Run("ZeroDate", func(t *testing.T) {
		var d MyDate
		v, err := d.Value()
		require.NoError(t, err)
		require.Nil(t, v)
	})
}
