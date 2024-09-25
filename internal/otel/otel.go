package neosyncotel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type TracerProvider interface {
	trace.TracerProvider

	Shutdown(context.Context) error
}

type MeterProvider interface {
	metric.MeterProvider

	Shutdown(context.Context) error
}

type SetupConfig struct {
	TraceProviders []TracerProvider
	MeterProviders []MeterProvider
	// If provided, configures the global text map propagator
	TextMapPropagator propagation.TextMapPropagator
	// Configures the global otel logger
	Logger logr.Logger
}

func SetupOtelSdk(config *SetupConfig) func(context.Context) error {
	otel.SetLogger(config.Logger)

	if config.TextMapPropagator != nil {
		otel.SetTextMapPropagator(config.TextMapPropagator)
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	for _, tp := range config.TraceProviders {
		if tp != nil {
			shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
		}
	}
	for _, mp := range config.MeterProviders {
		if mp != nil {
			shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
		}
	}

	return shutdown
}

func NewDefaultPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

type TraceProviderConfig struct {
	Exporter string
	Opts     TraceExporterOpts
}

func NewTraceProvider(ctx context.Context, config *TraceProviderConfig) (*tracesdk.TracerProvider, error) {
	exporter, err := getTraceExporter(ctx, config.Exporter, config.Opts)
	if err != nil {
		return nil, err
	}
	if exporter == nil {
		return nil, nil
	}
	provider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter, tracesdk.WithBatchTimeout(5*time.Second)),
	)
	return provider, nil
}

type TraceExporterOpts struct {
	Otlp    []otlptracegrpc.Option
	Console []stdouttrace.Option
}

const (
	otlpExporter    = "otlp"
	consoleExporter = "console"
	noneExporter    = "none"
)

func getTraceExporter(ctx context.Context, exporter string, opts TraceExporterOpts) (tracesdk.SpanExporter, error) {
	switch exporter {
	case otlpExporter:
		return otlptracegrpc.New(ctx, opts.Otlp...)
	case consoleExporter:
		return stdouttrace.New(opts.Console...)
	case noneExporter:
		return nil, nil
	default:
		return nil, fmt.Errorf("this tracer exporter is not currently supported %q: %w", exporter, errors.ErrUnsupported)
	}
}

type MeterProviderConfig struct {
	Exporter   string
	Opts       MeterExporterOpts
	AppVersion string
}

func NewMeterProvider(ctx context.Context, config *MeterProviderConfig) (*metricsdk.MeterProvider, error) {
	exporter, err := getMeterExporter(ctx, config.Exporter, config.Opts)
	if err != nil {
		return nil, err
	}
	if exporter == nil {
		return nil, nil
	}
	reader := metricsdk.WithReader(metricsdk.NewPeriodicReader(exporter))
	attributes := []attribute.KeyValue{semconv.ServiceVersion(config.AppVersion)}
	res := resource.NewWithAttributes(semconv.SchemaURL, attributes...)
	provider := metricsdk.NewMeterProvider(reader, metricsdk.WithResource(res))
	return provider, nil
}

type MeterExporterOpts struct {
	Otlp    []otlpmetricgrpc.Option
	Console []stdoutmetric.Option
}

func getMeterExporter(ctx context.Context, exporter string, opts MeterExporterOpts) (metricsdk.Exporter, error) {
	switch exporter {
	case otlpExporter:
		return otlpmetricgrpc.New(ctx, opts.Otlp...)
	case consoleExporter:
		return stdoutmetric.New(opts.Console...)
	case noneExporter:
		return nil, nil
	default:
		return nil, fmt.Errorf("this meter exporter is not currently supported %q: %w", exporter, errors.ErrUnsupported)
	}
}

func WithDefaultDeltaTemporalitySelector() otlpmetricgrpc.Option {
	return otlpmetricgrpc.WithTemporalitySelector(func(ik metricsdk.InstrumentKind) metricdata.Temporality {
		// Delta Temporality causes metrics to be reset after some time.
		// We are using this today for benthos metrics so that they don't persist indefinitely in the time series database
		return metricdata.DeltaTemporality
	})
}

func withCumulativeTemporalitySelector() otlpmetricgrpc.Option {
	return otlpmetricgrpc.WithTemporalitySelector(func(ik metricsdk.InstrumentKind) metricdata.Temporality {
		return metricdata.CumulativeTemporality
	})
}

type OtelEnvConfig struct {
	IsEnabled      bool
	ServiceVersion string

	TraceExporter string
	MeterExporter string
}

func GetOtelConfigFromViperEnv() OtelEnvConfig {
	return OtelEnvConfig{
		IsEnabled:      getIsOtelEnabled(),
		ServiceVersion: getAppVersion(),
		TraceExporter:  getTracerExporter(),
		MeterExporter:  getMetricsExporter(),
	}
}

func getIsOtelEnabled() bool {
	isDisabledStr := viper.GetString("OTEL_SDK_DISABLED")
	if isDisabledStr == "" {
		return false
	}
	return !viper.GetBool("OTEL_SDK_DISABLED")
}

func getAppVersion() string {
	return viper.GetString("OTEL_SERVICE_VERSION")
}

func getTracerExporter() string {
	exporter := viper.GetString("OTEL_TRACES_EXPORTER")
	if exporter == "" {
		return otlpExporter
	}
	return exporter
}

func getMetricsExporter() string {
	exporter := viper.GetString("OTEL_METRICS_EXPORTER")
	if exporter == "" {
		return otlpExporter
	}
	return exporter
}

// This will be used to test sending benthos metrics with cumulative temporality instead of delta for better prometheus compatibility
func GetBenthosMetricTemporalityOption() otlpmetricgrpc.Option {
	temporality := viper.GetString("BENTHOS_METER_TEMPORALITY")
	if temporality == "" || temporality == "delta" {
		return WithDefaultDeltaTemporalitySelector()
	}
	return withCumulativeTemporalitySelector()
}
