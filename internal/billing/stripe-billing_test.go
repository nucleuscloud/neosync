package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_getNextMonthBillingCycleAnchor(t *testing.T) {
	t.Run("next month", func(t *testing.T) {
		date := time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC)
		actual := getNextMonthBillingCycleAnchor(date)
		expected := time.Date(2024, time.May, 1, 0, 0, 0, 0, time.UTC).Unix()
		require.Equal(t, expected, actual)
	})
	t.Run("year", func(t *testing.T) {
		date := time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC)
		actual := getNextMonthBillingCycleAnchor(date)
		expected := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
		require.Equal(t, expected, actual)
	})
	t.Run("timezone", func(t *testing.T) {
		date := time.Date(2024, time.December, 1, 0, 0, 0, 0, time.Local)
		actual := getNextMonthBillingCycleAnchor(date)
		expected := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local).Unix()
		require.Equal(t, expected, actual)
	})
}
