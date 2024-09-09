package v1alpha1_metricsservice

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
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

	query, err := metrics.GetPromQueryFromMetric(req.Msg.GetMetric(), queryLabels, "1d")
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}
	results, _, err := metrics.GetDailyUsageFromProm(ctx, s.prometheusclient, query, start, end, logger)
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
	query, err := metrics.GetPromQueryFromMetric(req.Msg.GetMetric(), queryLabels, fmt.Sprintf("%dd", dayWindow))
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}

	totalCount, err := metrics.GetTotalUsageFromProm(ctx, s.prometheusclient, query, end, logger)
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
