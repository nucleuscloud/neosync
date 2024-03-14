package v1alpha1_metricsservice

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"

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

	query, err := getPromQueryFromMetric(req.Msg.GetMetric(), queryLabels)
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}
	queryResponse, warnings, err := s.prometheusclient.QueryRange(ctx, query, promv1.Range{
		Start: start,
		End:   end,
		Step:  getStepByRange(start, end),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to query prometheus for metrics: %w", err)
	}
	for _, warning := range warnings {
		logger.Warn(fmt.Sprintf("[PROMETHEUS]: %s", warning))
	}

	switch queryResponse.Type() {
	case model.ValMatrix:
		matrix, ok := queryResponse.(model.Matrix)
		if !ok {
			return nil, fmt.Errorf("unable to convert query response to model.Matrix, received type: %T", queryResponse)
		}
		usage, err := getDailyUsageFromMatrix(matrix)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&mgmtv1alpha1.GetDailyMetricCountResponse{
			Results: usage,
		}), nil
	default:
		return nil, fmt.Errorf("this method does not support query responses of type: %s", queryResponse.Type())
	}
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

	if req.Msg.GetStart() == nil || req.Msg.GetEnd() == nil {
		return nil, nucleuserrors.NewBadRequest("must provide a start and end time")
	}
	if req.Msg.GetMetric() == mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_UNSPECIFIED {
		return nil, nucleuserrors.NewBadRequest("must provide a metric name")
	}
	start := req.Msg.GetStart()
	end := req.Msg.GetEnd()

	if start.AsTime().After(end.AsTime()) {
		return nil, nucleuserrors.NewBadRequest("start must not be before end")
	}

	timeDiff := end.AsTime().Sub(start.AsTime())
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

	query, err := getPromQueryFromMetric(req.Msg.GetMetric(), queryLabels)
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}

	queryResponse, warnings, err := s.prometheusclient.QueryRange(ctx, query, promv1.Range{
		Start: start.AsTime(),
		End:   end.AsTime(),
		Step:  getStepByRange(start.AsTime(), end.AsTime()),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to query prometheus for metrics: %w", err)
	}
	for _, warning := range warnings {
		logger.Warn(fmt.Sprintf("[PROMETHEUS]: %s", warning))
	}

	switch queryResponse.Type() {
	case model.ValMatrix:
		matrix, ok := queryResponse.(model.Matrix)
		if !ok {
			return nil, fmt.Errorf("unable to convert query response to model.Matrix, received type: %T", queryResponse)
		}

		usage, err := getUsageFromMatrix(matrix)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&mgmtv1alpha1.GetMetricCountResponse{Count: sumUsage(usage)}), nil

	default:
		return nil, fmt.Errorf("this method does not support query responses of type: %s", queryResponse.Type())
	}
}

func getPromQueryFromMetric(
	metric mgmtv1alpha1.RangedMetricName,
	labels metrics.MetricLabels,
) (string, error) {
	metricName, err := getMetricNameFromEnum(metric)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s{%s}", metricName, labels.ToPromQueryString()), nil
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

func getStepByRange(start, end time.Time) time.Duration {
	diff := end.Sub(start)

	diffDays := int(diff.Hours() / 24)
	diffHours := int(diff.Hours())

	switch {
	case diffHours < 24:
		return 1 * time.Minute
	case diffDays >= 0 && diffDays <= 15:
		return 1 * time.Hour
	default:
		return 1 * time.Hour
	}
}

func getDailyUsageFromMatrix(matrix model.Matrix) ([]*mgmtv1alpha1.DayResult, error) {
	output := []*mgmtv1alpha1.DayResult{}
	for _, stream := range matrix {
		var latest int64
		var ts time.Time
		for _, value := range stream.Values {
			converted, err := strconv.ParseInt(value.Value.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert metric value to int64: %w", err)
			}
			if converted > latest {
				latest = converted
				ts = value.Timestamp.Time()
			}
		}
		if latest < 0 {
			return nil, fmt.Errorf("received a metric count that was less than 0")
		}
		output = append(output, &mgmtv1alpha1.DayResult{Date: ptr(timeToDate(ts)), Count: uint64(latest)})
	}
	squished := squishDayResults(output)
	// the results must be sorted as they come out of order from prometheus
	sort.Slice(squished, func(i, j int) bool {
		// Compare years
		if squished[i].Date.Year != squished[j].Date.Year {
			return squished[i].Date.Year < squished[j].Date.Year
		}
		// Years are equal, compare months
		if squished[i].Date.Month != squished[j].Date.Month {
			return squished[i].Date.Month < squished[j].Date.Month
		}
		// Both years and months are equal, compare days
		return squished[i].Date.Day < squished[j].Date.Day
	})
	return squished, nil
}

// combines counts where date is the day and returns a squished list with the original order retained
func squishDayResults(input []*mgmtv1alpha1.DayResult) []*mgmtv1alpha1.DayResult {
	dayMap := map[string]uint64{}
	for _, result := range input {
		dayMap[toDateKey(result.GetDate())] += result.GetCount()
	}
	output := []*mgmtv1alpha1.DayResult{}
	for _, result := range input {
		key := toDateKey(result.GetDate())
		if count, ok := dayMap[key]; ok {
			output = append(output, &mgmtv1alpha1.DayResult{Date: result.GetDate(), Count: count})
			delete(dayMap, key)
		}
	}
	return output
}

func ptr[T any](val T) *T {
	return &val
}

func timeToDate(t time.Time) mgmtv1alpha1.Date {
	return mgmtv1alpha1.Date{
		Year:  uint32(t.Year()),
		Month: uint32(t.Month()),
		Day:   uint32(t.Day()),
	}
}

func toDateKey(day *mgmtv1alpha1.Date) string {
	return fmt.Sprintf("%d_%d_%d", day.Day, day.Month, day.Year)
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

func getUsageFromMatrix(matrix model.Matrix) (map[string]uint64, error) {
	usage := map[string]uint64{}
	for _, stream := range matrix {
		usage[stream.Metric.String()] = 0

		var latest int64
		for _, value := range stream.Values {
			converted, err := strconv.ParseInt(value.Value.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert metric value to int64: %w", err)
			}
			if converted > latest {
				latest = converted
			}
		}
		if latest < 0 {
			return nil, fmt.Errorf("received a metric count that was less than 0")
		}
		usage[stream.Metric.String()] = uint64(latest)
	}
	return usage, nil
}

func sumUsage(usage map[string]uint64) uint64 {
	var total uint64
	for _, val := range usage {
		total += val
	}
	return total
}
