package neosynctypes

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNeosyncDateTime(t *testing.T) {
	t.Run("ScanPgx_Timestamp", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		ts := &pgtype.Timestamp{
			Time:  time.Date(2023, 6, 15, 14, 30, 45, 123456789, time.UTC),
			Valid: true,
		}

		err := dt.ScanPgx(ts)
		assert.NoError(t, err)
		assert.Equal(t, int32(2023), dt.Year)
		assert.Equal(t, uint8(6), dt.Month)
		assert.Equal(t, uint8(15), dt.Day)
		assert.Equal(t, uint8(14), dt.Hour)
		assert.Equal(t, uint8(30), dt.Minute)
		assert.Equal(t, uint8(45), dt.Second)
		assert.Equal(t, int32(123456789), dt.Nano)
		assert.Empty(t, dt.TimeZone)
		assert.False(t, dt.IsBC)
	})

	t.Run("ScanPgx_TimestampBC", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		ts := &pgtype.Timestamp{
			Time:  time.Date(-44, 3, 15, 12, 0, 0, 0, time.UTC),
			Valid: true,
		}

		err := dt.ScanPgx(ts)
		assert.NoError(t, err)
		assert.Equal(t, int32(45), dt.Year)
		assert.Equal(t, uint8(3), dt.Month)
		assert.Equal(t, uint8(15), dt.Day)
		assert.True(t, dt.IsBC)
	})

	t.Run("ScanPgx_Timestamptz", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		loc := time.FixedZone("UTC+1", 3600)
		ts := &pgtype.Timestamptz{
			Time:  time.Date(2023, 12, 25, 23, 59, 59, 999999999, loc),
			Valid: true,
		}

		err := dt.ScanPgx(ts)
		assert.NoError(t, err)
		assert.Equal(t, int32(2023), dt.Year)
		assert.Equal(t, uint8(12), dt.Month)
		assert.Equal(t, uint8(25), dt.Day)
		assert.Equal(t, uint8(23), dt.Hour)
		assert.Equal(t, uint8(59), dt.Minute)
		assert.Equal(t, uint8(59), dt.Second)
		assert.Equal(t, int32(999999999), dt.Nano)
		assert.Equal(t, "UTC+1", dt.TimeZone)
	})

	t.Run("ScanPgx_Date", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		d := &pgtype.Date{
			Time:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			Valid: true,
		}

		err := dt.ScanPgx(d)
		assert.NoError(t, err)
		assert.Equal(t, int32(2000), dt.Year)
		assert.Equal(t, uint8(1), dt.Month)
		assert.Equal(t, uint8(1), dt.Day)
		assert.Equal(t, uint8(0), dt.Hour)
		assert.Equal(t, uint8(0), dt.Minute)
		assert.Equal(t, uint8(0), dt.Second)
		assert.Equal(t, int32(0), dt.Nano)
		assert.Equal(t, "UTC", dt.TimeZone)
	})

	t.Run("ScanPgx_TimeTime", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		loc := time.FixedZone("UTC+2", 7200)
		tt := time.Date(2024, 3, 18, 10, 30, 45, 123456789, loc)

		err := dt.ScanPgx(tt)
		assert.NoError(t, err)
		assert.Equal(t, int32(2024), dt.Year)
		assert.Equal(t, uint8(3), dt.Month)
		assert.Equal(t, uint8(18), dt.Day)
		assert.Equal(t, uint8(10), dt.Hour)
		assert.Equal(t, uint8(30), dt.Minute)
		assert.Equal(t, uint8(45), dt.Second)
		assert.Equal(t, int32(123456789), dt.Nano)
		assert.Equal(t, "UTC+2", dt.TimeZone)
	})

	t.Run("ScanPgx_TimeTime_UTC", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		tt := time.Date(2024, 3, 18, 10, 30, 45, 123456789, time.UTC)

		err := dt.ScanPgx(tt)
		assert.NoError(t, err)
		assert.Equal(t, int32(2024), dt.Year)
		assert.Equal(t, uint8(3), dt.Month)
		assert.Equal(t, uint8(18), dt.Day)
		assert.Equal(t, uint8(10), dt.Hour)
		assert.Equal(t, uint8(30), dt.Minute)
		assert.Equal(t, uint8(45), dt.Second)
		assert.Equal(t, int32(123456789), dt.Nano)
		assert.Empty(t, dt.TimeZone)
	})

	t.Run("ScanPgx_TimeTime_BC", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		tt := time.Date(-44, 3, 15, 12, 0, 0, 0, time.UTC)

		err := dt.ScanPgx(tt)
		assert.NoError(t, err)
		assert.Equal(t, int32(45), dt.Year)
		assert.Equal(t, uint8(3), dt.Month)
		assert.Equal(t, uint8(15), dt.Day)
		assert.Equal(t, uint8(12), dt.Hour)
		assert.Equal(t, uint8(0), dt.Minute)
		assert.Equal(t, uint8(0), dt.Second)
		assert.Equal(t, int32(0), dt.Nano)
		assert.True(t, dt.IsBC)
	})

	t.Run("ValuePgx_Timestamp", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:   2023,
			Month:  7,
			Day:    4,
			Hour:   16,
			Minute: 20,
			Second: 30,
			Nano:   500000000,
		}

		val, err := dt.ValuePgx()
		assert.NoError(t, err)
		ts, ok := val.(*pgtype.Timestamp)
		assert.True(t, ok)
		assert.True(t, ts.Valid)
		assert.Equal(t, time.Date(2023, 7, 4, 16, 20, 30, 500000000, time.UTC), ts.Time)
	})

	t.Run("ValuePgx_TimestampBC", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:  100,
			Month: 1,
			Day:   1,
			IsBC:  true,
		}

		val, err := dt.ValuePgx()
		assert.NoError(t, err)
		ts, ok := val.(*pgtype.Timestamp)
		assert.True(t, ok)
		assert.True(t, ts.Valid)
		assert.Equal(t, time.Date(-99, 1, 1, 0, 0, 0, 0, time.UTC), ts.Time)
	})

	t.Run("ValuePgx_DateBC", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:      100,
			Month:     1,
			Day:       1,
			IsBC:      true,
			Precision: -1,
		}

		val, err := dt.ValuePgx()
		assert.NoError(t, err)
		d, ok := val.(*pgtype.Date)
		assert.True(t, ok)
		assert.True(t, d.Valid)
		assert.Equal(t, time.Date(-99, 1, 1, 0, 0, 0, 0, time.UTC), d.Time)
	})

	t.Run("ValuePgx_TimestamptzBC", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:     100,
			Month:    7,
			Day:      4,
			Hour:     16,
			Minute:   20,
			Second:   30,
			IsBC:     true,
			TimeZone: "America/New_York",
		}

		val, err := dt.ValuePgx()
		assert.NoError(t, err)
		tstz, ok := val.(*pgtype.Timestamptz)
		assert.True(t, ok)
		assert.True(t, tstz.Valid)
		assert.Equal(t, time.Date(-99, 7, 4, 16, 20, 30, 0, time.UTC), tstz.Time)
	})

	t.Run("ValuePgx_Date", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:      2024,
			Month:     2,
			Day:       29,
			Precision: -1,
		}

		val, err := dt.ValuePgx()
		assert.NoError(t, err)
		d, ok := val.(*pgtype.Date)
		assert.True(t, ok)
		assert.True(t, d.Valid)
		assert.Equal(t, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), d.Time)
	})

	t.Run("ValuePgx_Timestamptz", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:     2023,
			Month:    8,
			Day:      15,
			Hour:     13,
			Minute:   45,
			Second:   0,
			TimeZone: "America/New_York",
		}

		val, err := dt.ValuePgx()
		assert.NoError(t, err)
		tstz, ok := val.(*pgtype.Timestamptz)
		assert.True(t, ok)
		assert.True(t, tstz.Valid)
		loc, _ := time.LoadLocation("America/New_York")
		assert.Equal(t, time.Date(2023, 8, 15, 13, 45, 0, 0, loc), tstz.Time)
	})

	t.Run("ValueMysql_DateTime", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:     2023,
			Month:    8,
			Day:      15,
			Hour:     13,
			Minute:   45,
			Second:   30,
			TimeZone: "America/New_York",
		}

		val, err := dt.ValueMysql()
		assert.NoError(t, err)
		assert.Equal(t, "2023-08-15 13:45:30", val)
	})

	t.Run("ValueMysql_Date", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:  2024,
			Month: 2,
			Day:   29,
		}

		val, err := dt.ValueMysql()
		assert.NoError(t, err)
		assert.Equal(t, "2024-02-29", val)
	})

	t.Run("ValueMysql_BC", func(t *testing.T) {
		dt := &NeosyncDateTime{
			Year:     100,
			Month:    7,
			Day:      4,
			Hour:     16,
			Minute:   20,
			Second:   30,
			IsBC:     true,
			TimeZone: "America/New_York",
		}

		val, err := dt.ValueMysql()
		assert.NoError(t, err)
		assert.Equal(t, "-0099-07-04 16:20:30", val)
	})

	t.Run("ScanMysql", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		loc, _ := time.LoadLocation("America/New_York")
		err := dt.ScanMysql(time.Date(2023, 8, 15, 13, 45, 30, 0, loc))
		assert.NoError(t, err)

		assert.Equal(t, int32(2023), dt.Year)
		assert.Equal(t, uint8(8), dt.Month)
		assert.Equal(t, uint8(15), dt.Day)
		assert.Equal(t, uint8(13), dt.Hour)
		assert.Equal(t, uint8(45), dt.Minute)
		assert.Equal(t, uint8(30), dt.Second)
		assert.Equal(t, "America/New_York", dt.TimeZone)
	})

	t.Run("ScanMysql_BC", func(t *testing.T) {
		dt := &NeosyncDateTime{}
		loc, _ := time.LoadLocation("America/New_York")
		err := dt.ScanMysql(time.Date(-99, 7, 4, 16, 20, 30, 0, loc))
		assert.NoError(t, err)

		assert.Equal(t, int32(100), dt.Year)
		assert.Equal(t, uint8(7), dt.Month)
		assert.Equal(t, uint8(4), dt.Day)
		assert.Equal(t, uint8(16), dt.Hour)
		assert.Equal(t, uint8(20), dt.Minute)
		assert.Equal(t, uint8(30), dt.Second)
		assert.Equal(t, "America/New_York", dt.TimeZone)
		assert.True(t, dt.IsBC)
	})
}
