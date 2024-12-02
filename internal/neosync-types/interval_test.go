package neosynctypes

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func Test_Interval(t *testing.T) {
	t.Run("NewInterval initializes with correct defaults", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		require.NotNil(t, interval)
		require.Equal(t, NeosyncIntervalId, interval.Neosync.TypeId)
		require.Equal(t, LatestVersion, interval.GetVersion())
	})

	t.Run("NewInterval invalid version", func(t *testing.T) {
		customVersion := Version(99999)
		_, err := NewInterval(WithVersion(customVersion))
		require.Error(t, err)
	})

	t.Run("NewInterval set version", func(t *testing.T) {
		ver := Version(1)
		interval, err := NewInterval(WithVersion(ver))
		require.NoError(t, err)
		require.Equal(t, ver, interval.GetVersion())
	})

	t.Run("ScanPgx handles nil value", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		err = interval.ScanPgx(nil)
		require.NoError(t, err)
	})

	t.Run("ScanPgx handles invalid type", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		err = interval.ScanPgx("not an interval")
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected *pgtype.Interval")
	})

	t.Run("ScanPgx handles invalid pgtype.Interval", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		pgInterval := &pgtype.Interval{Valid: false}
		err = interval.ScanPgx(pgInterval)
		require.NoError(t, err)
		require.Zero(t, interval.Microseconds)
		require.Zero(t, interval.Days)
		require.Zero(t, interval.Months)
	})

	t.Run("ScanPgx correctly scans valid interval", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		pgInterval := &pgtype.Interval{
			Microseconds: 1000000,
			Days:         7,
			Months:       1,
			Valid:        true,
		}
		err = interval.ScanPgx(pgInterval)
		require.NoError(t, err)
		require.Equal(t, int64(1000000), interval.Microseconds)
		require.Equal(t, int32(7), interval.Days)
		require.Equal(t, int32(1), interval.Months)
	})

	t.Run("ValuePgx returns correct pgtype.Interval", func(t *testing.T) {
		interval := &Interval{
			Microseconds: 1000000,
			Days:         7,
			Months:       1,
		}
		value, err := interval.ValuePgx()
		require.NoError(t, err)
		pgInterval, ok := value.(*pgtype.Interval)
		require.True(t, ok)
		require.Equal(t, int64(1000000), pgInterval.Microseconds)
		require.Equal(t, int32(7), pgInterval.Days)
		require.Equal(t, int32(1), pgInterval.Months)
		require.True(t, pgInterval.Valid)
	})

	t.Run("ScanJson handles byte slice", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		jsonData := []byte(`{"microseconds": 1000000, "days": 7, "months": 1}`)
		err = interval.ScanJson(jsonData)
		require.NoError(t, err)
		require.Equal(t, int64(1000000), interval.Microseconds)
		require.Equal(t, int32(7), interval.Days)
		require.Equal(t, int32(1), interval.Months)
	})

	t.Run("ScanJson handles string", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		jsonStr := `{"microseconds": 1000000, "days": 7, "months": 1}`
		err = interval.ScanJson(jsonStr)
		require.NoError(t, err)
		require.Equal(t, int64(1000000), interval.Microseconds)
		require.Equal(t, int32(7), interval.Days)
		require.Equal(t, int32(1), interval.Months)
	})

	t.Run("ScanJson handles unsupported type", func(t *testing.T) {
		interval, err := NewInterval()
		require.NoError(t, err)
		err = interval.ScanJson(123)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported scan type")
	})

	t.Run("ValueJson returns valid JSON", func(t *testing.T) {
		interval := &Interval{
			Microseconds: 1000000,
			Days:         7,
			Months:       1,
		}
		interval.Neosync.TypeId = NeosyncIntervalId
		value, err := interval.ValueJson()
		require.NoError(t, err)
		jsonBytes, ok := value.([]byte)
		require.True(t, ok)

		var unmarshaled map[string]interface{}
		err = json.Unmarshal(jsonBytes, &unmarshaled)
		require.NoError(t, err)
		require.Equal(t, float64(1000000), unmarshaled["microseconds"])
		require.Equal(t, float64(7), unmarshaled["days"])
		require.Equal(t, float64(1), unmarshaled["months"])
	})

	t.Run("NewIntervalFromPgx handles valid interval", func(t *testing.T) {
		pgInterval := &pgtype.Interval{
			Microseconds: 1000000,
			Days:         7,
			Months:       1,
			Valid:        true,
		}
		interval, err := NewIntervalFromPgx(pgInterval)
		require.NoError(t, err)
		require.NotNil(t, interval)
		require.Equal(t, int64(1000000), interval.Microseconds)
		require.Equal(t, int32(7), interval.Days)
		require.Equal(t, int32(1), interval.Months)
	})

	t.Run("NewIntervalFromPgx handles invalid type", func(t *testing.T) {
		interval, err := NewIntervalFromPgx("not an interval")
		require.Error(t, err)
		require.Nil(t, interval)
		require.Contains(t, err.Error(), "expected *pgtype.Interval")
	})

	t.Run("NewIntervalArrayFromPgx handles valid intervals", func(t *testing.T) {
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
		array, err := NewIntervalArrayFromPgx(pgIntervals, nil)
		require.NoError(t, err)
		require.NotNil(t, array)
		require.Len(t, array.Elements, 2)

		interval1, ok := array.Elements[0].(*Interval)
		require.True(t, ok)
		require.Equal(t, int64(1000000), interval1.Microseconds)
		require.Equal(t, int32(7), interval1.Days)
		require.Equal(t, int32(1), interval1.Months)

		interval2, ok := array.Elements[1].(*Interval)
		require.True(t, ok)
		require.Equal(t, int64(2000000), interval2.Microseconds)
		require.Equal(t, int32(14), interval2.Days)
		require.Equal(t, int32(2), interval2.Months)
	})
}
