package metrics

import (
	"testing"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_GenerateMonthRegexRange(t *testing.T) {
	tests := []struct {
		name      string
		startDate *mgmtv1alpha1.Date
		endDate   *mgmtv1alpha1.Date
		expected  []string
	}{
		{
			name:      "Same month",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 30},
			expected:  []string{"2024-09-.*"},
		},
		{
			name:      "Two consecutive months",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 15},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 10, Day: 15},
			expected:  []string{"2024-09-.*", "2024-10-.*"},
		},
		{
			name:      "Span across year",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 12, Day: 15},
			endDate:   &mgmtv1alpha1.Date{Year: 2025, Month: 1, Day: 15},
			expected:  []string{"2024-12-.*", "2025-01-.*"},
		},
		{
			name:      "Multiple months in same year",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 1, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 3, Day: 31},
			expected:  []string{"2024-01-.*", "2024-02-.*", "2024-03-.*"},
		},
		{
			name:      "Entire year",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 1, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 12, Day: 31},
			expected: []string{
				"2024-01-.*", "2024-02-.*", "2024-03-.*", "2024-04-.*",
				"2024-05-.*", "2024-06-.*", "2024-07-.*", "2024-08-.*",
				"2024-09-.*", "2024-10-.*", "2024-11-.*", "2024-12-.*",
			},
		},
		{
			name:      "Multiple years",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 11, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2025, Month: 2, Day: 28},
			expected:  []string{"2024-11-.*", "2024-12-.*", "2025-01-.*", "2025-02-.*"},
		},
		{
			name:      "Start date after end date",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 10, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 30},
			expected:  []string{},
		},
		{
			name:      "Single day",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 15},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 15},
			expected:  []string{"2024-09-.*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateMonthRegexRange(tt.startDate, tt.endDate)
			require.Equal(t, tt.expected, result, "The generated patterns do not match the expected output")
		})
	}
}

func Test_CalculatePromLookbackDuration(t *testing.T) {
	tests := []struct {
		name      string
		startDate *mgmtv1alpha1.Date
		endDate   *mgmtv1alpha1.Date
		expected  string
	}{
		{
			name:      "Same day",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 24},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 24},
			expected:  "1d",
		},
		{
			name:      "One day difference",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 24},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 25},
			expected:  "2d",
		},
		{
			name:      "One month difference",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2024, Month: 10, Day: 1},
			expected:  "31d",
		},
		{
			name:      "One year difference",
			startDate: &mgmtv1alpha1.Date{Year: 2024, Month: 1, Day: 1},
			endDate:   &mgmtv1alpha1.Date{Year: 2025, Month: 1, Day: 1},
			expected:  "367d", // 2024 is a leap year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePromLookbackDuration(tt.startDate, tt.endDate)
			if result != tt.expected {
				t.Errorf("CalculatePromLookbackDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_daysBetween(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "Same day",
			start:    time.Date(2024, 9, 24, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 9, 24, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "One day difference",
			start:    time.Date(2024, 9, 24, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 9, 25, 0, 0, 0, 0, time.UTC),
			expected: 2,
		},
		{
			name:     "One month difference",
			start:    time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			expected: 31,
		},
		{
			name:     "One year difference (leap year)",
			start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 367,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := daysBetween(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("daysBetween() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_DateToTime(t *testing.T) {
	tests := []struct {
		name     string
		input    *mgmtv1alpha1.Date
		expected time.Time
	}{
		{
			name:     "Full date specified",
			input:    &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 24},
			expected: time.Date(2024, time.September, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Zero year defaults to year 1",
			input:    &mgmtv1alpha1.Date{Year: 0, Month: 9, Day: 24},
			expected: time.Date(1, time.September, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Zero month defaults to January",
			input:    &mgmtv1alpha1.Date{Year: 2024, Month: 0, Day: 24},
			expected: time.Date(2024, time.January, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Zero day defaults to 1st of the month",
			input:    &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 0},
			expected: time.Date(2024, time.September, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "All zeros default to Jan 1, year 1",
			input:    &mgmtv1alpha1.Date{Year: 0, Month: 0, Day: 0},
			expected: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateToTime(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("DateToTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_ToEndOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "UTC time",
			input:    time.Date(2024, 9, 24, 12, 30, 45, 0, time.UTC),
			expected: time.Date(2024, 9, 24, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "Non-UTC time",
			input:    time.Date(2024, 9, 24, 12, 30, 45, 0, time.FixedZone("EST", -5*60*60)),
			expected: time.Date(2024, 9, 24, 23, 59, 59, 999999999, time.FixedZone("EST", -5*60*60)),
		},
		{
			name:     "Already at end of day",
			input:    time.Date(2024, 9, 24, 23, 59, 59, 999999999, time.UTC),
			expected: time.Date(2024, 9, 24, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "Start of day",
			input:    time.Date(2024, 9, 24, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 9, 24, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToEndOfDay(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("ToEndOfDay() = %v, want %v", result, tt.expected)
			}
			if result.Location() != tt.input.Location() {
				t.Errorf("ToEndOfDay() timezone = %v, want %v", result.Location(), tt.input.Location())
			}
		})
	}
}
