package v1alpha1_metricsservice

import (
	"context"
	"fmt"
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

func (s *Service) GetRangedMetrics(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetRangedMetricsRequest],
) (*connect.Response[mgmtv1alpha1.GetRangedMetricsResponse], error) {
	return nil, nucleuserrors.NewNotImplemented("this method is not currently implemented")
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
		return 1 * 24 * time.Hour
	}
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
