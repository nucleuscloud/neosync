package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/sync/errgroup"
)

const usageDateFormat = "2006-01-02"

func GetDailyUsageFromProm(ctx context.Context, api promv1.API, query string, start, end time.Time, logger *slog.Logger) ([]*mgmtv1alpha1.DayResult, float64, error) {
	var dailyResults []*mgmtv1alpha1.DayResult
	dailyTotals := make(map[string]float64)
	var overallTotal float64

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(10)
	mu := sync.Mutex{}
	// Iterate through each day in the range
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		d := d
		errgrp.Go(func() error {
			dayStart := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Nanosecond)
			result, warnings, err := api.Query(errctx, query, dayEnd)
			if err != nil {
				return fmt.Errorf("error querying Prometheus for date %s: %w", d.Format(usageDateFormat), err)
			}
			if len(warnings) > 0 {
				logger.Warn(fmt.Sprintf("[PROMETHEUS]: %v", warnings))
			}

			// Process the results for this day
			vector, ok := result.(model.Vector)
			if !ok {
				return fmt.Errorf("error casting result to model.Vector for date %s", d.Format(usageDateFormat))
			}

			for _, sample := range vector {
				// using the dayStart timestamp here instead of the sample TS because sometimes it comes back as UTC for the next day
				// which throws off the count. Using the dayStart we know that it will
				day := dayStart.Format(usageDateFormat)
				value := float64(sample.Value)
				mu.Lock()
				dailyTotals[day] += value
				mu.Unlock()
			}
			return nil
		})
	}
	err := errgrp.Wait()
	if err != nil {
		return nil, -1, err
	}

	var dates []string
	for day := range dailyTotals {
		dates = append(dates, day)
	}
	sort.Strings(dates)

	for _, day := range dates {
		date, err := time.Parse(usageDateFormat, day)
		if err != nil {
			return nil, 0, err
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
	return fmt.Sprintf("sum(max_over_time(%s{%s}[%s]))", metricName, labels.ToPromQueryString(), timeWindow), nil
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
