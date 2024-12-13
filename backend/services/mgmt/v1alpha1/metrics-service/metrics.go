package v1alpha1_metricsservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
)

const (
	// When querying by multiple dates in PromQL, this is used in the regex to allow doing an OR search on the various dates being queried
	metricDateSeparator = "|"
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
	start := metrics.DateToTime(req.Msg.GetStart())
	end := metrics.ToEndOfDay(metrics.DateToTime(req.Msg.GetEnd()))

	if start.After(end) {
		return nil, nucleuserrors.NewBadRequest("start must not be before end")
	}

	timeDiff := end.Sub(start)
	if timeDiff > timeLimit {
		return nil, nucleuserrors.NewBadRequest("duration between start and end must not exceed 60 days")
	}

	queryLabels := metrics.MetricLabels{
		metrics.NewNotEqLabel(metrics.IsUpdateConfigLabel, "true"), // we want to always exclude update configs
		metrics.NewRegexMatchLabel(metrics.NeosyncDateLabel, strings.Join(metrics.GenerateMonthRegexRange(req.Msg.GetStart(), req.Msg.GetEnd()), metricDateSeparator)),
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	switch identifier := req.Msg.Identifier.(type) {
	case *mgmtv1alpha1.GetDailyMetricCountRequest_AccountId:
		if err := user.EnforceAccount(ctx, userdata.NewIdentifier(identifier.AccountId), rbac.AccountAction_View); err != nil {
			return nil, err
		}
		queryLabels = append(queryLabels, metrics.NewEqLabel(metrics.AccountIdLabel, identifier.AccountId))
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

	query, err := metrics.GetPromQueryFromMetric(
		req.Msg.GetMetric(),
		queryLabels,
		metrics.CalculatePromLookbackDuration(req.Msg.GetStart(), req.Msg.GetEnd()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}
	results, _, err := metrics.GetDailyUsageFromProm(ctx, s.prometheusclient, query, end, logger)
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
	start := metrics.DateToTime(req.Msg.GetStartDay())
	end := metrics.ToEndOfDay(metrics.DateToTime(req.Msg.GetEndDay()))

	if start.After(end) {
		return nil, nucleuserrors.NewBadRequest("start must not be before end")
	}

	timeDiff := end.Sub(start)
	if timeDiff > timeLimit {
		return nil, nucleuserrors.NewBadRequest("duration between start and end must not exceed 60 days")
	}

	queryLabels := metrics.MetricLabels{
		metrics.NewNotEqLabel(metrics.IsUpdateConfigLabel, "true"), // we want to always exclude update configs
		metrics.NewRegexMatchLabel(metrics.NeosyncDateLabel, strings.Join(metrics.GenerateMonthRegexRange(req.Msg.GetStartDay(), req.Msg.GetEndDay()), metricDateSeparator)),
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	switch identifier := req.Msg.Identifier.(type) {
	case *mgmtv1alpha1.GetMetricCountRequest_AccountId:
		if err := user.EnforceAccount(ctx, userdata.NewIdentifier(identifier.AccountId), rbac.AccountAction_View); err != nil {
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

	query, err := metrics.GetPromQueryFromMetric(
		req.Msg.GetMetric(),
		queryLabels,
		metrics.CalculatePromLookbackDuration(req.Msg.GetStartDay(), req.Msg.GetEndDay()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to compute valid prometheus query: %w", err)
	}

	totalCount, err := metrics.GetTotalUsageFromProm(ctx, s.prometheusclient, query, end, logger)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.GetMetricCountResponse{Count: uint64(totalCount)}), nil
}
