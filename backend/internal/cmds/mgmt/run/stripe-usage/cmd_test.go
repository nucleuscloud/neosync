package run_stripe_usage_cmd

import (
	"testing"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_getEventTimestamp(t *testing.T) {
	t.Run("past date", func(t *testing.T) {
		actual := getEventTimestamp(&mgmtv1alpha1.Date{
			Year:  2024,
			Month: 9,
			Day:   23,
		})
		t.Log(uint64(1727135999), actual)
		require.Equal(t, uint64(1727092800), actual)
	})
	t.Run("future date", func(t *testing.T) {
		now := time.Now().UTC()
		actual := getEventTimestamp(&mgmtv1alpha1.Date{
			Year:  uint32(now.Year()),
			Month: uint32(now.Month()),
			Day:   uint32(now.Day()),
		})
		require.GreaterOrEqual(t, uint64(now.Unix()), actual)
	})
}

func Test_getEventId(t *testing.T) {
	actual := getEventId("myaccountid", &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23})
	require.Equal(t, "myaccountid-2024-09-23", actual)
}

func Test_formatDate(t *testing.T) {
	type testcase struct {
		input    *mgmtv1alpha1.Date
		expected string
	}

	testcases := []testcase{
		{&mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, "2024-09-23"},
		{&mgmtv1alpha1.Date{Year: 2024, Month: 2, Day: 29}, "2024-02-29"},
		{&mgmtv1alpha1.Date{Year: 2024, Month: 12, Day: 25}, "2024-12-25"},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			actual := formatDate(tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_getIngestDate(t *testing.T) {
	t.Run("both", func(t *testing.T) {
		actual, err := getIngestDate("2024-03-01", "-24h")
		require.NoError(t, err)
		require.Equal(t, 0, compareDates(&mgmtv1alpha1.Date{
			Year:  2024,
			Month: 2,
			Day:   29,
		}, actual), "dates were not the same")
	})

	t.Run("only ingest no offset", func(t *testing.T) {
		actual, err := getIngestDate("2024-03-01", "")
		require.NoError(t, err)
		require.Equal(t, 0, compareDates(&mgmtv1alpha1.Date{
			Year:  2024,
			Month: 3,
			Day:   1,
		}, actual), "dates were not the same")
	})

	t.Run("no ingest, but offset", func(t *testing.T) {
		today := time.Now().UTC()
		yesterday := today.Add(-24 * time.Hour)
		actual, err := getIngestDate("", "-24h")
		require.NoError(t, err)
		require.Equal(t, 0, compareDates(&mgmtv1alpha1.Date{
			Year:  uint32(yesterday.Year()),
			Month: uint32(yesterday.Month()),
			Day:   uint32(yesterday.Day()),
		}, actual), "dates were not the same")
	})

	t.Run("defaults", func(t *testing.T) {
		today := time.Now().UTC()
		actual, err := getIngestDate("", "")
		require.NoError(t, err)
		require.Equal(t, 0, compareDates(&mgmtv1alpha1.Date{
			Year:  uint32(today.Year()),
			Month: uint32(today.Month()),
			Day:   uint32(today.Day()),
		}, actual), "dates were not the same")
	})

	t.Run("invalid ingest date", func(t *testing.T) {
		actual, err := getIngestDate("2024-03-033", "")
		require.Error(t, err)
		require.Nil(t, actual)
	})

	t.Run("invalid offset", func(t *testing.T) {
		actual, err := getIngestDate("2024-09-23", "12zzh")
		require.Error(t, err)
		require.Nil(t, actual)
	})
}

func Test_compareDates(t *testing.T) {
	type testcase struct {
		d1       *mgmtv1alpha1.Date
		d2       *mgmtv1alpha1.Date
		expected int
		name     string
	}
	testcases := []testcase{
		{&mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, 0, "equal"},
		{&mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 22}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, -1, "day before"},
		{&mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 24}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, 1, "day after"},

		{&mgmtv1alpha1.Date{Year: 2023, Month: 9, Day: 23}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, -1, "year before"},
		{&mgmtv1alpha1.Date{Year: 2025, Month: 9, Day: 23}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, 1, "year after"},

		{&mgmtv1alpha1.Date{Year: 2024, Month: 8, Day: 23}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, -1, "month before"},
		{&mgmtv1alpha1.Date{Year: 2024, Month: 10, Day: 23}, &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, 1, "month after"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := compareDates(tc.d1, tc.d2)
			require.Equal(t, tc.expected, actual)
		})
	}
}
