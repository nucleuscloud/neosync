package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"testing"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	promapiv1mock "github.com/nucleuscloud/neosync/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_getPromQueryFromMetric(t *testing.T) {
	output, err := GetPromQueryFromMetric(
		mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		MetricLabels{NewEqLabel("foo", "bar"), NewEqLabel("foo2", "bar2")},
		"1d",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(
		t,
		`sum(max_over_time(input_received_total{foo="bar",foo2="bar2"}[1d])) by (date)`,
		output,
	)
}

func Test_getPromQueryFromMetric_Invalid_Metric(t *testing.T) {
	output, err := GetPromQueryFromMetric(
		mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_UNSPECIFIED,
		MetricLabels{NewEqLabel("foo", "bar"), NewEqLabel("foo2", "bar2")},
		"1d",
	)
	assert.Error(t, err)
	assert.Empty(t, output)
}

func Test_getDayFromMetric(t *testing.T) {
	t.Run("valid metric", func(t *testing.T) {
		actual := getDayFromMetric(model.Metric{NeosyncDateLabel: NeosyncDateFormat}, time.Now())
		require.Equal(t, NeosyncDateFormat, actual)
	})

	t.Run("timestamp fallback", func(t *testing.T) {
		now := time.Now()
		date := timeToDate(now)
		actual := getDayFromMetric(model.Metric{}, now)
		require.Equal(t, formatDate(&date), actual)
	})
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

func Test_sortUsageDates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Already sorted dates",
			input:    []string{"2024-01-01", "2024-01-02", "2024-01-03"},
			expected: []string{"2024-01-01", "2024-01-02", "2024-01-03"},
		},
		{
			name:     "Reverse sorted dates",
			input:    []string{"2024-03-01", "2024-02-01", "2024-01-01"},
			expected: []string{"2024-01-01", "2024-02-01", "2024-03-01"},
		},
		{
			name:     "Mixed order dates",
			input:    []string{"2024-02-15", "2024-01-30", "2024-03-01", "2024-02-01"},
			expected: []string{"2024-01-30", "2024-02-01", "2024-02-15", "2024-03-01"},
		},
		{
			name:     "Dates spanning multiple years",
			input:    []string{"2025-01-01", "2023-12-31", "2024-06-15"},
			expected: []string{"2023-12-31", "2024-06-15", "2025-01-01"},
		},
		{
			name:     "Dates with invalid entries",
			input:    []string{"2024-01-01", "invalid-date", "2023-12-31", "also-invalid"},
			expected: []string{"2023-12-31", "2024-01-01", "invalid-date", "also-invalid"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single date",
			input:    []string{"2024-01-01"},
			expected: []string{"2024-01-01"},
		},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			slices.SortFunc(testcase.input, sortUsageDates)
			require.Equal(t, testcase.expected, testcase.input)
		})
	}
}

func Test_GetDailyUsageFromProm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockapi := promapiv1mock.NewMockAPI(t)
		mockapi.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().
			Return(model.Vector{
				&model.Sample{
					Metric:    model.Metric{NeosyncDateLabel: "2024-09-23"},
					Value:     10,
					Timestamp: model.TimeFromUnix(time.Date(2024, 9, 23, 12, 0, 0, 0, time.UTC).Unix()),
				},
				&model.Sample{
					Metric:    model.Metric{NeosyncDateLabel: "2024-09-24"},
					Value:     15,
					Timestamp: model.TimeFromUnix(time.Date(2024, 9, 24, 12, 0, 0, 0, time.UTC).Unix()),
				},
			}, nil, nil)
		actual, actualTotal, err := GetDailyUsageFromProm(context.Background(), mockapi, "test", time.Now(), slog.Default())
		require.NoError(t, err)
		require.Equal(t, float64(25), actualTotal)

		require.Equal(
			t,
			[]*mgmtv1alpha1.DayResult{
				{Date: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 23}, Count: 10},
				{Date: &mgmtv1alpha1.Date{Year: 2024, Month: 9, Day: 24}, Count: 15},
			},
			actual,
		)
		mockapi.AssertExpectations(t)
	})
}

func Test_GetTotalUsageFromProm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockapi := promapiv1mock.NewMockAPI(t)
		mockapi.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().
			Return(model.Vector{
				&model.Sample{
					Metric:    model.Metric{NeosyncDateLabel: "2024-09-23"},
					Value:     10,
					Timestamp: model.TimeFromUnix(time.Date(2024, 9, 23, 12, 0, 0, 0, time.UTC).Unix()),
				},
				&model.Sample{
					Metric:    model.Metric{NeosyncDateLabel: "2024-09-24"},
					Value:     15,
					Timestamp: model.TimeFromUnix(time.Date(2024, 9, 24, 12, 0, 0, 0, time.UTC).Unix()),
				},
			}, nil, nil)
		actual, err := GetTotalUsageFromProm(context.Background(), mockapi, "test", time.Now(), slog.Default())
		require.NoError(t, err)
		require.Equal(t, float64(25), actual)

		mockapi.AssertExpectations(t)
	})
}

// Formats the day into the Neosync Date Format of YYYY-DD-MM
func formatDate(d *mgmtv1alpha1.Date) string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}
