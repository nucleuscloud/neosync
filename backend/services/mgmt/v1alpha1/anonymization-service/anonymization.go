package v1alpha_anonymizationservice

import (
	"context"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	jsonanonymizer "github.com/nucleuscloud/neosync/internal/json-anonymizer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	inputMetricStr         = "input_received"
	outputMetricStr        = "output_sent"
	outputBatchCounterStr  = "output_batch_sent" // stream endpint only
	outputErrorsCounterStr = "output_error"
)

func (s *Service) AnonymizeMany(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.AnonymizeManyRequest],
) (*connect.Response[mgmtv1alpha1.AnonymizeManyResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	// add account id to request
	anonymizer, err := jsonanonymizer.NewAnonymizer(
		jsonanonymizer.WithTransformerMappings(req.Msg.TransformerMappings),
		jsonanonymizer.WithDefaultTransformers(req.Msg.DefaultTransformers),
		jsonanonymizer.WithHaltOnFailure(req.Msg.HaltOnFailure),
	)
	if err != nil {
		return nil, err
	}

	var outputBatchCounter, outputCounter metric.Int64Counter
	var labels []attribute.KeyValue
	if s.meter != nil {
		labels, err = getMetricLabels(ctx, "anonymizeMany", neosyncdb.UUIDString(*accountUuid))
		if err != nil {
			return nil, err
		}
		counter, err := s.meter.Int64Counter(inputMetricStr)
		if err != nil {
			return nil, err
		}
		counter.Add(ctx, int64(len(req.Msg.InputData)), metric.WithAttributes(labels...))
		outputCounter, err = s.meter.Int64Counter(outputMetricStr)
		if err != nil {
			return nil, err
		}
		outputBatchCounter, err = s.meter.Int64Counter(outputBatchCounterStr)
		if err != nil {
			return nil, err
		}
	}

	outputData, anonymizeErrors := anonymizer.AnonymizeJSONObjects(req.Msg.InputData)

	if outputCounter != nil && outputBatchCounter != nil {
		anonymizedCounter := 0
		for _, js := range outputData {
			if js != "" {
				anonymizedCounter += 1
			}
		}
		outputCounter.Add(ctx, int64(anonymizedCounter), metric.WithAttributes(labels...))
	}

	errors := []*mgmtv1alpha1.AnonymizeManyErrors{}
	for _, e := range anonymizeErrors {
		errors = append(errors, &mgmtv1alpha1.AnonymizeManyErrors{
			InputIndex:   e.InputIndex,
			ErrorMessage: e.Message,
		})
	}

	return connect.NewResponse(&mgmtv1alpha1.AnonymizeManyResponse{
		OutputData: outputData,
		Errors:     errors,
	}), nil
}

func (s *Service) AnonymizeSingle(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.AnonymizeSingleRequest],
) (*connect.Response[mgmtv1alpha1.AnonymizeSingleResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	anonymizer, err := jsonanonymizer.NewAnonymizer(
		jsonanonymizer.WithTransformerMappings(req.Msg.TransformerMappings),
		jsonanonymizer.WithDefaultTransformers(req.Msg.DefaultTransformers),
	)
	if err != nil {
		return nil, err
	}

	var outputCounter metric.Int64Counter
	var labels []attribute.KeyValue
	if s.meter != nil {
		labels, err = getMetricLabels(ctx, "anonymizeSingle", neosyncdb.UUIDString(*accountUuid))
		if err != nil {
			return nil, err
		}
		counter, err := s.meter.Int64Counter(inputMetricStr)
		if err != nil {
			return nil, err
		}
		counter.Add(ctx, 1, metric.WithAttributes(labels...))
		outputCounter, err = s.meter.Int64Counter(outputMetricStr)
		if err != nil {
			return nil, err
		}
	}

	outputData, err := anonymizer.AnonymizeJSONObject(req.Msg.InputData)
	if err != nil {
		return nil, err
	}

	if outputCounter != nil {
		outputCounter.Add(ctx, 1, metric.WithAttributes(labels...))
	}

	return connect.NewResponse(&mgmtv1alpha1.AnonymizeSingleResponse{
		OutputData: outputData,
	}), nil
}

func getMetricLabels(ctx context.Context, requestName string, accountId string) ([]attribute.KeyValue, error) {
	traceId := getTraceID(ctx)
	return []attribute.KeyValue{
		attribute.String(metrics.AccountIdLabel, accountId),
		attribute.String(metrics.ApiRequestId, traceId),
		attribute.String(metrics.ApiRequestName, requestName),
		attribute.String(metrics.NeosyncDateLabel, time.Now().UTC().Format(metrics.NeosyncDateFormat)),
	}, nil
}

func getTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID()
		return traceID.String()
	}
	return ""
}
