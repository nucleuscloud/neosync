package v1alpha1_metricsservice

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"

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

	queryLabels := map[string]string{}

	switch identifier := req.Msg.Identifier.(type) {
	case *mgmtv1alpha1.GetMetricCountRequest_AccountId:
		if _, err := s.verifyUserInAccount(ctx, identifier.AccountId); err != nil {
			return nil, err
		}
		queryLabels["neosyncAccountId"] = identifier.AccountId
		// get metrics by account id
	case *mgmtv1alpha1.GetMetricCountRequest_JobId:
		jobResp, err := s.jobservice.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: identifier.JobId}))
		if err != nil {
			return nil, err
		}
		queryLabels["neosyncAccountId"] = jobResp.Msg.GetJob().GetAccountId()
		queryLabels["neosyncJobId"] = jobResp.Msg.GetJob().GetId()
		// get metrics by job id
	case *mgmtv1alpha1.GetMetricCountRequest_RunId:
		jrResp, err := s.jobservice.GetJobRun(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRunRequest{JobRunId: identifier.RunId}))
		if err != nil {
			return nil, err
		}
		// dont really need to add account id here since it's implied by the job id
		queryLabels["neosyncJobId"] = jrResp.Msg.GetJobRun().GetJobId()
		queryLabels["neosyncRunId"] = jrResp.Msg.GetJobRun().GetId()

		// get metrics by run id
	default:
		return nil, nucleuserrors.NewBadRequest("must provide a valid identifier to proceed")
	}

	query := fmt.Sprintf("%s{%s}", req.Msg.GetMetric(), toPrometheusLabels(queryLabels))

	queryResponse, warnings, err := s.prometheusclient.QueryRange(ctx, query, promv1.Range{})
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

		usage := map[string]int64{}
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
			usage[stream.Metric.String()] = latest
		}

		var total uint64
		for _, val := range usage {
			total += uint64(val)
		}
		return connect.NewResponse(&mgmtv1alpha1.GetMetricCountResponse{Count: total}), nil

	default:
		return nil, fmt.Errorf("this method does not support query responses of type: %s", queryResponse.Type())
	}
}

func toPrometheusLabels(input map[string]string) string {
	pieces := []string{}

	for key, value := range input {
		pieces = append(pieces, fmt.Sprintf("%s=%q", key, value))
	}

	return strings.Join(pieces, ",")
}
