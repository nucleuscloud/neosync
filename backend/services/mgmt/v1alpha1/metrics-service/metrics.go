package v1alpha1_metricsservice

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"golang.org/x/sync/errgroup"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func (s *Service) GetDailyMetricCount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetDailyMetricCountRequest],
) (*connect.Response[mgmtv1alpha1.GetDailyMetricCountResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	if req.Msg.GetStart() == nil || req.Msg.GetEnd() == nil {
		return nil, nucleuserrors.NewBadRequest("must provide a start and end time")
	}
	if req.Msg.GetMetric() == mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_UNSPECIFIED {
		return nil, nucleuserrors.NewBadRequest("must provide a metric name")
	}
	start := dateToTime(req.Msg.GetStart())
	end := toEndOfDay(dateToTime(req.Msg.GetEnd()))

	if start.After(end) {
		return nil, nucleuserrors.NewBadRequest("start must not be before end")
	}

	timeDiff := end.Sub(start)
	if timeDiff > timeLimit {
		return nil, nucleuserrors.NewBadRequest("duration between start and end must not exceed 60 days")
	}

	queryLabels := metrics.MetricLabels{
		metrics.NewNotEqLabel(metrics.IsUpdateConfigLabel, "true"), // we want to always exclude update configs
	}

	switch identifier := req.Msg.Identifier.(type) {
	case *mgmtv1alpha1.GetDailyMetricCountRequest_AccountId:
		if _, err := s.verifyUserInAccount(ctx, identifier.AccountId); err != nil {
			return nil, err
		}
		queryLabels = append(queryLabels, metrics.NewEqLabel(metrics.AccountIdLabel, identifier.AccountId))
	case *mgmtv1alpha1.GetDailyMetricCountRequest_JobId:
		jobResp, err := s.jobservice.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: identifier.JobId}))
		if err != nil {
			return nil, err
		}
		queryLabels = append(
			queryLabels,
			metrics.NewEqLabel(metrics.AccountIdLabel, jobResp.Msg.GetJob().GetAccountId()),
			metrics.NewEqLabel(metrics.JobIdLabel, jobResp.Msg.GetJob().GetId()),
		)
	case *mgmtv1alpha1.GetDailyMetricCountRequest_RunId:
		jrResp, err := s.jobservice.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{JobRunId: identifier.RunId}))
		if err != nil {
			return nil, err
		}
		// dont really need to add account id here since it's implied by the job id
		queryLabels = append(
			queryLabels,
			metrics.NewEqLabel(metrics.JobIdLabel, jrResp.Msg.GetJobRun().GetJobId()),
			metrics.NewEqLabel(metrics.TemporalWorkflowId, jrResp.Msg.GetJobRun().GetId()),
		)
	default:
		return nil, nucleuserrors.NewBadRequest("must provide a valid identifier to proceed")
	}

	query, err := getPromQueryFromMetric(req.Msg.GetMetric(), queryLabels, "1d")
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}
	results, _, err := getDailyUsageFromProm(ctx, s.prometheusclient, query, start, end, logger)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.GetDailyMetricCountResponse{Results: results}), nil
}

var (
	// NeosyncCloud currently limits prom metric storage to 60 days
	// todo: expose as env var if we want to change this per environment
	timeLimit = 60 * 24 * time.Hour
)

func (s *Service) GetMetricCount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetMetricCountRequest],
) (*connect.Response[mgmtv1alpha1.GetMetricCountResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	if req.Msg.GetStartDay() == nil || req.Msg.GetEndDay() == nil {
		return nil, nucleuserrors.NewBadRequest("must provide a start and end time")
	}
	if req.Msg.GetMetric() == mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_UNSPECIFIED {
		return nil, nucleuserrors.NewBadRequest("must provide a metric name")
	}
	start := dateToTime(req.Msg.GetStartDay())
	end := toEndOfDay(dateToTime(req.Msg.GetEndDay()))

	if start.After(end) {
		return nil, nucleuserrors.NewBadRequest("start must not be before end")
	}

	timeDiff := end.Sub(start)
	if timeDiff > timeLimit {
		return nil, nucleuserrors.NewBadRequest("duration between start and end must not exceed 60 days")
	}

	queryLabels := metrics.MetricLabels{
		metrics.NewNotEqLabel(metrics.IsUpdateConfigLabel, "true"), // we want to always exclude update configs
	}

	switch identifier := req.Msg.Identifier.(type) {
	case *mgmtv1alpha1.GetMetricCountRequest_AccountId:
		if _, err := s.verifyUserInAccount(ctx, identifier.AccountId); err != nil {
			return nil, err
		}
		queryLabels = append(queryLabels, metrics.NewEqLabel(metrics.AccountIdLabel, identifier.AccountId))
	case *mgmtv1alpha1.GetMetricCountRequest_JobId:
		jobResp, err := s.jobservice.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: identifier.JobId}))
		if err != nil {
			return nil, err
		}
		queryLabels = append(
			queryLabels,
			metrics.NewEqLabel(metrics.AccountIdLabel, jobResp.Msg.GetJob().GetAccountId()),
			metrics.NewEqLabel(metrics.JobIdLabel, jobResp.Msg.GetJob().GetId()),
		)
	case *mgmtv1alpha1.GetMetricCountRequest_RunId:
		jrResp, err := s.jobservice.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{JobRunId: identifier.RunId}))
		if err != nil {
			return nil, err
		}
		// dont really need to add account id here since it's implied by the job id
		queryLabels = append(
			queryLabels,
			metrics.NewEqLabel(metrics.JobIdLabel, jrResp.Msg.GetJobRun().GetJobId()),
			metrics.NewEqLabel(metrics.TemporalWorkflowId, jrResp.Msg.GetJobRun().GetId()),
		)
	default:
		return nil, nucleuserrors.NewBadRequest("must provide a valid identifier to proceed")
	}

	dayWindow := daysBetween(start, end)
	query, err := getPromQueryFromMetric(req.Msg.GetMetric(), queryLabels, fmt.Sprintf("%dd", dayWindow))
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}

	totalCount, err := getTotalUsageFromProm(ctx, s.prometheusclient, query, end, logger)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.GetMetricCountResponse{Count: uint64(totalCount)}), nil
}

func daysBetween(start, end time.Time) int {
	// Ensure the times are in UTC to avoid timezone issues
	start = start.UTC()
	end = end.UTC()

	// Calculate the difference in days
	duration := end.Sub(start)
	days := int(duration.Hours()/24) + 1
	// Convert the number of days to a string and return
	return days
}

func getPromQueryFromMetric(
	metric mgmtv1alpha1.RangedMetricName,
	labels metrics.MetricLabels,
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

const usageDateFormat = "2006-01-02"

func getDailyUsageFromProm(ctx context.Context, api promv1.API, query string, start, end time.Time, logger *slog.Logger) ([]*mgmtv1alpha1.DayResult, float64, error) {
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

func getTotalUsageFromProm(ctx context.Context, api promv1.API, query string, dayEnd time.Time, logger *slog.Logger) (float64, error) {
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

func timeToDate(t time.Time) mgmtv1alpha1.Date {
	return mgmtv1alpha1.Date{
		Year:  uint32(t.Year()),
		Month: uint32(t.Month()),
		Day:   uint32(t.Day()),
	}
}

func dateToTime(d *mgmtv1alpha1.Date) time.Time {
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
func toEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
}
