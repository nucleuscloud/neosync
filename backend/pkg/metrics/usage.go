package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func GetDailyUsageFromProm(
	ctx context.Context,
	api promv1.API,
	query string,
	ts time.Time,
	logger *slog.Logger,
) ([]*mgmtv1alpha1.DayResult, float64, error) {
	result, warnings, err := api.Query(ctx, query, ts)
	if err != nil {
		return nil, -1, fmt.Errorf("error querying Prometheus for daily usage: %w", err)
	}
	if len(warnings) > 0 {
		logger.Warn(fmt.Sprintf("[PROMETHEUS]: %v", warnings))
	}

	vector, ok := result.(model.Vector)
	if !ok {
		return nil, -1, fmt.Errorf("error casting prometheus query result to model.Vector. Got %T", result)
	}

	dailyTotals := map[string]float64{}
	for _, sample := range vector {
		day := getDayFromMetric(sample.Metric, sample.Timestamp.Time().UTC())
		value := float64(sample.Value)
		dailyTotals[day] += value
	}

	dailyResults := []*mgmtv1alpha1.DayResult{}
	var dates []string
	for day := range dailyTotals {
		dates = append(dates, day)
	}
	slices.SortFunc(dates, sortUsageDates)

	var overallTotal float64
	for _, day := range dates {
		date, err := time.Parse(NeosyncDateFormat, day)
		if err != nil {
			return nil, -1, fmt.Errorf("unable to convert day back to usage date (%q) format (%q): %w", date, NeosyncDateFormat, err)
		}
		mgmtDate := timeToDate(date)
		dailyResults = append(dailyResults, &mgmtv1alpha1.DayResult{
			Date:  &mgmtDate,
			Count: uint64(dailyTotals[day]),
		})
		overallTotal += dailyTotals[day]
	}

	return dailyResults, overallTotal, nil
}

func getDayFromMetric(metric model.Metric, metricTimestamp time.Time) string {
	metricVal, ok := metric[NeosyncDateLabel]
	if ok && metricVal.IsValid() {
		return string(metricVal)
	}
	return metricTimestamp.Format(NeosyncDateFormat)
}

// Plugs in to slices.SortFunc
func sortUsageDates(a, b string) int {
	dateA, errA := time.Parse(NeosyncDateFormat, a)
	dateB, errB := time.Parse(NeosyncDateFormat, b)

	// If both dates are invalid, maintain their original order
	if errA != nil && errB != nil {
		return 0
	}

	// If only one date is invalid, consider it "greater" (sort it to the end)
	if errA != nil {
		return 1
	}
	if errB != nil {
		return -1
	}

	// If both dates are valid, compare them
	if dateA.Before(dateB) {
		return -1
	}
	if dateA.After(dateB) {
		return 1
	}
	return 0
}

func GetTotalUsageFromProm(ctx context.Context, api promv1.API, query string, dayEnd time.Time, logger *slog.Logger) (float64, error) {
	var overallTotal float64

	result, warnings, err := api.Query(ctx, query, dayEnd)
	if err != nil {
		return -1, fmt.Errorf("error querying Prometheus for date %s: %w", dayEnd, err)
	}
	if len(warnings) > 0 {
		logger.Warn(fmt.Sprintf("[PROMETHEUS]: %v", warnings))
	}

	// Process the results for this day
	vector, ok := result.(model.Vector)
	if !ok {
		return -1, fmt.Errorf("error casting result to model.Vector for date %s", dayEnd)
	}

	for _, sample := range vector {
		value := float64(sample.Value)
		overallTotal += value
	}

	return overallTotal, nil
}

func GetPromQueryFromMetric(
	metric mgmtv1alpha1.RangedMetricName,
	labels MetricLabels,
	timeWindow string,
) (string, error) {
	metricName, err := getMetricNameFromEnum(metric)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sum(max_over_time(%s{%s}[%s])) by (%s)", metricName, labels.ToPromQueryString(), timeWindow, NeosyncDateLabel), nil
}

const (
	inputReceivedTotalMetric = "input_received_total"
)

func getMetricNameFromEnum(metric mgmtv1alpha1.RangedMetricName) (string, error) {
	switch metric {
	case mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED:
		return inputReceivedTotalMetric, nil
	default:
		return "", fmt.Errorf("invalid metric name")
	}
}

func timeToDate(t time.Time) mgmtv1alpha1.Date {
	return mgmtv1alpha1.Date{
		Year:  uint32(t.Year()),  //nolint:gosec // Ignoring for now
		Month: uint32(t.Month()), //nolint:gosec // Ignoring for now
		Day:   uint32(t.Day()),   //nolint:gosec // Ignoring for now
	}
}
