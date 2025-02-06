package v1alpha_anonymizationservice

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	jsonanonymizer "github.com/nucleuscloud/neosync/internal/json-anonymizer"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	inputMetricStr        = "input_received"
	outputMetricStr       = "output_sent"
	outputErrorCounterStr = "output_error"
)

func (s *Service) AnonymizeMany(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.AnonymizeManyRequest],
) (*connect.Response[mgmtv1alpha1.AnonymizeManyResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	if !s.cfg.IsNeosyncCloud {
		return nil, nucleuserrors.NewNotImplemented(
			fmt.Sprintf("%s is not implemented in the OSS version of Neosync.", strings.TrimPrefix(mgmtv1alpha1connect.AnonymizationServiceAnonymizeManyProcedure, "/")),
		)
	}

	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceAccountAccess(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
	if err != nil {
		return nil, err
	}
	if account.AccountType == int16(neosyncdb.AccountType_Personal) {
		return nil, nucleuserrors.NewForbidden(
			fmt.Sprintf("%s is not implemented for personal accounts", strings.TrimPrefix(mgmtv1alpha1connect.AnonymizationServiceAnonymizeManyProcedure, "/")),
		)
	}

	for cfg := range getTransformerConfigsToValidate(req.Msg) {
		if err := validateTransformerConfig(cfg); err != nil {
			return nil, err
		}
	}

	requestedCount := uint64(len(req.Msg.InputData))
	resp, err := s.useraccountService.IsAccountStatusValid(ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
		AccountId:            req.Msg.GetAccountId(),
		RequestedRecordCount: &requestedCount,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account status: %w", err)
	}

	if !resp.Msg.IsValid {
		return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("unable to anonymize due to account in invalid state. Reason: %q", *resp.Msg.Reason))
	}

	anonymizer, err := jsonanonymizer.NewAnonymizer(
		jsonanonymizer.WithTransformerMappings(req.Msg.TransformerMappings),
		jsonanonymizer.WithDefaultTransformers(req.Msg.DefaultTransformers),
		jsonanonymizer.WithHaltOnFailure(req.Msg.HaltOnFailure),
		jsonanonymizer.WithConditionalAnonymizeConfig(s.cfg.IsPresidioEnabled, s.analyze, s.anonymize, s.cfg.PresidioDefaultLanguage),
		jsonanonymizer.WithTransformerClient(s.transformerClient),
		jsonanonymizer.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	var outputErrorCounter, outputCounter metric.Int64Counter
	var labels []attribute.KeyValue
	if s.meter != nil {
		labels = getMetricLabels(ctx, "anonymizeMany", req.Msg.GetAccountId())
		counter, err := s.meter.Int64Counter(inputMetricStr)
		if err != nil {
			return nil, err
		}
		counter.Add(ctx, int64(len(req.Msg.InputData)), metric.WithAttributes(labels...))
		outputCounter, err = s.meter.Int64Counter(outputMetricStr)
		if err != nil {
			return nil, err
		}
		outputErrorCounter, err = s.meter.Int64Counter(outputErrorCounterStr)
		if err != nil {
			return nil, err
		}
	}

	outputData, anonymizeErrors := anonymizer.AnonymizeJSONObjects(req.Msg.InputData)

	if outputCounter != nil {
		anonymizedCounter := 0
		for _, js := range outputData {
			if js != "" {
				anonymizedCounter += 1
			}
		}
		outputCounter.Add(ctx, int64(anonymizedCounter), metric.WithAttributes(labels...))
	}

	if outputErrorCounter != nil && len(anonymizeErrors) > 0 {
		outputErrorCounter.Add(ctx, int64(len(anonymizeErrors)), metric.WithAttributes(labels...))
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
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceAccountAccess(ctx, req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	accountUuid, err := neosyncdb.ToUuid(req.Msg.GetAccountId())
	if err != nil {
		return nil, err
	}

	account, err := s.db.Q.GetAccount(ctx, s.db.Db, accountUuid)
	if err != nil {
		return nil, err
	}
	if !s.cfg.IsNeosyncCloud || account.AccountType == int16(neosyncdb.AccountType_Personal) {
		for _, mapping := range req.Msg.GetTransformerMappings() {
			if mapping.GetTransformer().GetTransformPiiTextConfig() != nil {
				return nil, nucleuserrors.NewForbidden("TransformPiiText is not available for use. Please contact us to upgrade your account.")
			}
		}
		defaultTransforms := req.Msg.GetDefaultTransformers()
		if defaultTransforms.GetBoolean().GetTransformPiiTextConfig() != nil ||
			defaultTransforms.GetN().GetTransformPiiTextConfig() != nil ||
			defaultTransforms.GetS().GetTransformPiiTextConfig() != nil {
			return nil, nucleuserrors.NewForbidden("TransformPiiText is not available for use. Please contact us to upgrade your account.")
		}
	}

	for cfg := range getTransformerConfigsToValidate(req.Msg) {
		if err := validateTransformerConfig(cfg); err != nil {
			return nil, err
		}
	}

	requestedCount := uint64(len(req.Msg.InputData))
	resp, err := s.useraccountService.IsAccountStatusValid(ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
		AccountId:            req.Msg.GetAccountId(),
		RequestedRecordCount: &requestedCount,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account status: %w", err)
	}

	if !resp.Msg.IsValid {
		return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("unable to anonymize due to account in invalid state. Reason: %q", *resp.Msg.Reason))
	}

	anonymizer, err := jsonanonymizer.NewAnonymizer(
		jsonanonymizer.WithTransformerMappings(req.Msg.TransformerMappings),
		jsonanonymizer.WithDefaultTransformers(req.Msg.DefaultTransformers),
		jsonanonymizer.WithConditionalAnonymizeConfig(s.cfg.IsPresidioEnabled, s.analyze, s.anonymize, s.cfg.PresidioDefaultLanguage),
		jsonanonymizer.WithTransformerClient(s.transformerClient),
		jsonanonymizer.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	var outputCounter, outputErrorCounter metric.Int64Counter
	var labels []attribute.KeyValue
	if s.meter != nil {
		labels = getMetricLabels(ctx, "anonymizeSingle", req.Msg.GetAccountId())
		counter, err := s.meter.Int64Counter(inputMetricStr)
		if err != nil {
			return nil, err
		}
		counter.Add(ctx, 1, metric.WithAttributes(labels...))
		outputCounter, err = s.meter.Int64Counter(outputMetricStr)
		if err != nil {
			return nil, err
		}
		outputErrorCounter, err = s.meter.Int64Counter(outputErrorCounterStr)
		if err != nil {
			return nil, err
		}
	}

	outputData, err := anonymizer.AnonymizeJSONObject(req.Msg.InputData)
	if err != nil {
		if outputErrorCounter != nil {
			outputErrorCounter.Add(ctx, int64(1), metric.WithAttributes(labels...))
		}
		return nil, err
	}

	if outputCounter != nil {
		outputCounter.Add(ctx, 1, metric.WithAttributes(labels...))
	}

	return connect.NewResponse(&mgmtv1alpha1.AnonymizeSingleResponse{
		OutputData: outputData,
	}), nil
}

func getMetricLabels(ctx context.Context, requestName, accountId string) []attribute.KeyValue {
	requestId := getTraceID(ctx)
	if requestId == "" {
		requestId = uuid.NewString()
	}
	return []attribute.KeyValue{
		attribute.String(metrics.AccountIdLabel, accountId),
		attribute.String(metrics.ApiRequestId, requestId),
		attribute.String(metrics.ApiRequestName, requestName),
		attribute.String(metrics.NeosyncDateLabel, time.Now().UTC().Format(metrics.NeosyncDateFormat)),
	}
}

func getTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID()
		return traceID.String()
	}
	return ""
}

// ensures the transformer config is of a valid configuration
// the main thing it does today is ensure that TransformPiiText is not being used recursively
func validateTransformerConfig(cfg *mgmtv1alpha1.TransformerConfig) error {
	if cfg == nil {
		return fmt.Errorf("transformer config is nil")
	}
	root := cfg.GetTransformPiiTextConfig()
	if root == nil {
		return nil
	}
	defaultAnonymizer := root.GetDefaultAnonymizer()
	if defaultAnonymizer != nil {
		child := defaultAnonymizer.GetTransform().GetConfig().GetTransformPiiTextConfig()
		if child != nil {
			return nucleuserrors.NewBadRequest("found nested TransformPiiText config in default anonymizer. TransformPiiText may not be used deeply nested within itself.")
		}
	}
	entityAnonymizers := root.GetEntityAnonymizers()
	for entity, entityAnonymizer := range entityAnonymizers {
		child := entityAnonymizer.GetTransform().GetConfig().GetTransformPiiTextConfig()
		if child != nil {
			return nucleuserrors.NewBadRequest(fmt.Sprintf("found nested TransformPiiText config in entity (%s) anonymizer. TransformPiiText may not be used deeply nested within itself.", entity))
		}
	}
	return nil
}

type transformerMsgToValidate interface {
	GetDefaultTransformers() *mgmtv1alpha1.DefaultTransformersConfig
	GetTransformerMappings() []*mgmtv1alpha1.TransformerMapping
}

func getTransformerConfigsToValidate(msg transformerMsgToValidate) iter.Seq[*mgmtv1alpha1.TransformerConfig] {
	return func(yield func(*mgmtv1alpha1.TransformerConfig) bool) {
		if msg.GetDefaultTransformers().GetBoolean() != nil {
			if !yield(msg.GetDefaultTransformers().GetBoolean()) {
				return
			}
		}
		if msg.GetDefaultTransformers().GetN() != nil {
			if !yield(msg.GetDefaultTransformers().GetN()) {
				return
			}
		}
		if msg.GetDefaultTransformers().GetS() != nil {
			if !yield(msg.GetDefaultTransformers().GetS()) {
				return
			}
		}

		for _, mapping := range msg.GetTransformerMappings() {
			if mapping.GetTransformer() != nil {
				if !yield(mapping.GetTransformer()) {
					return
				}
			}
		}
	}
}
