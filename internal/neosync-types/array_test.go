package neosynctypes

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func Test_NeosyncArray(t *testing.T) {
	t.Run("NewNeosyncArray initializes with correct defaults", func(t *testing.T) {
		elements := []NeosyncAdapter{
			&Interval{},
			&Interval{},
		}
		array, err := NewNeosyncArray(elements)
		require.NoError(t, err)
		require.NotNil(t, array)
		require.Equal(t, NeosyncArrayId, array.Neosync.TypeId)
		require.Equal(t, LatestVersion, array.GetVersion())
		require.Equal(t, len(elements), len(array.Elements))
	})

	t.Run("NewNeosyncArray accepts options", func(t *testing.T) {
		elements := []NeosyncAdapter{&Interval{}}
		customVersion := Version(1)
		array, err := NewNeosyncArray(elements, WithVersion(customVersion))
		require.NoError(t, err)
		require.Equal(t, customVersion, array.GetVersion())
	})

	t.Run("NewNeosyncArray with invalid version returns error", func(t *testing.T) {
		elements := []NeosyncAdapter{&Interval{}}
		invalidVersion := Version(999)
		array, err := NewNeosyncArray(elements, WithVersion(invalidVersion))
		require.Error(t, err)
		require.Nil(t, array)
		require.Contains(t, err.Error(), "invalid Neosync Type version")
	})

	t.Run("ScanPgx handles slice type mismatch", func(t *testing.T) {
		array, err := NewNeosyncArray([]NeosyncAdapter{&Interval{}})
		require.NoError(t, err)
		err = array.ScanPgx("not a slice")
		require.Error(t, err)
	})

	t.Run("ScanPgx handles length mismatch", func(t *testing.T) {
		array, err := NewNeosyncArray([]NeosyncAdapter{&Interval{}})
		require.NoError(t, err)
		err = array.ScanPgx([]interface{}{1, 2}) // More elements than adapters
		require.Error(t, err)
		require.Contains(t, err.Error(), "length mismatch")
	})

	t.Run("ScanPgx handles valid intervals", func(t *testing.T) {
		interval1 := &Interval{}
		interval2 := &Interval{}
		array, err := NewNeosyncArray([]NeosyncAdapter{interval1, interval2})
		require.NoError(t, err)

		pgIntervals := []*pgtype.Interval{
			{
				Microseconds: 1000000,
				Days:         7,
				Months:       1,
				Valid:        true,
			},
			{
				Microseconds: 2000000,
				Days:         14,
				Months:       2,
				Valid:        true,
			},
		}

		err = array.ScanPgx(pgIntervals)
		require.NoError(t, err)

		scannedInterval1 := array.Elements[0].(*Interval)
		require.Equal(t, int64(1000000), scannedInterval1.Microseconds)
		require.Equal(t, int32(7), scannedInterval1.Days)
		require.Equal(t, int32(1), scannedInterval1.Months)

		scannedInterval2 := array.Elements[1].(*Interval)
		require.Equal(t, int64(2000000), scannedInterval2.Microseconds)
		require.Equal(t, int32(14), scannedInterval2.Days)
		require.Equal(t, int32(2), scannedInterval2.Months)
	})

	t.Run("ValuePgx returns correct slice of values", func(t *testing.T) {
		interval1 := &Interval{Microseconds: 1000000, Days: 7, Months: 1}
		interval2 := &Interval{Microseconds: 2000000, Days: 14, Months: 2}
		array, err := NewNeosyncArray([]NeosyncAdapter{interval1, interval2})
		require.NoError(t, err)

		value, err := array.ValuePgx()
		require.NoError(t, err)

		values, ok := value.(pq.GenericArray)
		require.True(t, ok)

		arrValues, err := values.Value()
		require.NoError(t, err)
		require.Equal(t, `{"1 mon 7 day 00:00:01","2 mon 14 day 00:00:02"}`, arrValues)
	})

	t.Run("ScanJson handles type mismatch", func(t *testing.T) {
		array, err := NewNeosyncArray([]NeosyncAdapter{&Interval{}})
		require.NoError(t, err)
		err = array.ScanJson("not an array")
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not a slice")
	})

	t.Run("ScanJson handles length mismatch", func(t *testing.T) {
		array, err := NewNeosyncArray([]NeosyncAdapter{&Interval{}})
		require.NoError(t, err)
		err = array.ScanJson([]interface{}{1, 2}) // More elements than adapters
		require.Error(t, err)
		require.Contains(t, err.Error(), "length mismatch")
	})

	t.Run("ScanJson handles valid input", func(t *testing.T) {
		interval1 := &Interval{}
		interval2 := &Interval{}
		array, err := NewNeosyncArray([]NeosyncAdapter{interval1, interval2})
		require.NoError(t, err)

		jsonData := []any{
			[]byte(`{"microseconds": 1000000, "days": 7, "months": 1}`),
			[]byte(`{"microseconds": 2000000, "days": 14, "months": 2}`),
		}

		err = array.ScanJson(jsonData)
		require.NoError(t, err)

		scannedInterval1 := array.Elements[0].(*Interval)
		require.Equal(t, int64(1000000), scannedInterval1.Microseconds)
		require.Equal(t, int32(7), scannedInterval1.Days)
		require.Equal(t, int32(1), scannedInterval1.Months)

		scannedInterval2 := array.Elements[1].(*Interval)
		require.Equal(t, int64(2000000), scannedInterval2.Microseconds)
		require.Equal(t, int32(14), scannedInterval2.Days)
		require.Equal(t, int32(2), scannedInterval2.Months)
	})

	t.Run("ValueJson returns correct array of values", func(t *testing.T) {
		interval1 := &Interval{Microseconds: 1000000, Days: 7, Months: 1}
		interval2 := &Interval{Microseconds: 2000000, Days: 14, Months: 2}
		array, err := NewNeosyncArray([]NeosyncAdapter{interval1, interval2})
		require.NoError(t, err)

		value, err := array.ValueJson()
		require.NoError(t, err)

		values, ok := value.([]any)
		require.True(t, ok)
		require.Len(t, values, 2)

		// Each value should be []byte from json.Marshal
		jsonBytes1, ok := values[0].([]byte)
		require.True(t, ok)
		jsonBytes2, ok := values[1].([]byte)
		require.True(t, ok)

		var result1 Interval
		err = json.Unmarshal(jsonBytes1, &result1)
		require.NoError(t, err)
		require.Equal(t, int64(1000000), result1.Microseconds)
		require.Equal(t, int32(7), result1.Days)
		require.Equal(t, int32(1), result1.Months)

		var result2 Interval
		err = json.Unmarshal(jsonBytes2, &result2)
		require.NoError(t, err)
		require.Equal(t, int64(2000000), result2.Microseconds)
		require.Equal(t, int32(14), result2.Days)
		require.Equal(t, int32(2), result2.Months)
	})

	t.Run("Version getters and setters work correctly", func(t *testing.T) {
		array, err := NewNeosyncArray([]NeosyncAdapter{&Interval{}})
		require.NoError(t, err)

		originalVersion := array.GetVersion()
		newVersion := Version(2)
		array.setVersion(newVersion)

		require.NotEqual(t, originalVersion, array.GetVersion())
		require.Equal(t, newVersion, array.GetVersion())
	})
}
