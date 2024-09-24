package metrics

import (
	"fmt"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

// For a given date range, returns the PromQL Month Regex Patterns.
//
// Returns:
//
// []string{"2024-09-.*", "2024-10-.*"}
func GenerateMonthRegexRange(startDate, endDate *mgmtv1alpha1.Date) []string {
	start := time.Date(int(startDate.Year), time.Month(startDate.Month), int(startDate.Day), 0, 0, 0, 0, time.UTC)
	end := time.Date(int(endDate.Year), time.Month(endDate.Month), int(endDate.Day), 0, 0, 0, 0, time.UTC)

	patterns := []string{}
	has := map[string]any{}
	current := start

	for !current.After(end) {
		// Creates patterns that look like this: 2024-09.*
		pattern := fmt.Sprintf("%04d-%02d-.*", current.Year(), current.Month())
		if _, ok := has[pattern]; !ok {
			patterns = append(patterns, pattern)
			has[pattern] = struct{}{}
		}
		current = current.AddDate(0, 1, 0) // Move to the next month
	}

	return patterns
}

// For a given date range, returns the lookback period duration to be plugged in to the PromQL Query
//
// Returns:
//
// 2024-09-14, 2024-09-15 == 2d
func CalculatePromLookbackDuration(startDate, endDate *mgmtv1alpha1.Date) string {
	start := time.Date(int(startDate.Year), time.Month(startDate.Month), int(startDate.Day), 0, 0, 0, 0, time.UTC)
	end := time.Date(int(endDate.Year), time.Month(endDate.Month), int(endDate.Day), 0, 0, 0, 0, time.UTC)

	days := daysBetween(start, end)

	return fmt.Sprintf("%dd", days)
}

func daysBetween(start, end time.Time) int {
	// Calculate the difference in days
	duration := end.Sub(start)
	days := int(duration.Hours()/24) + 1
	// Convert the number of days to a string and return
	return days
}

func DateToTime(d *mgmtv1alpha1.Date) time.Time {
	year := int(d.Year)
	if year == 0 {
		year = 1 // default to year 1 if unspecified
	}
	month := time.Month(d.Month)
	if month == 0 {
		month = time.January // default to January if unspecified
	}
	day := int(d.Day)
	if day == 0 {
		day = 1 // default to first of the month if unspecified
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func ToEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}
