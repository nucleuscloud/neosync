package benthos_metrics

import (
	"context"

	"github.com/redpanda-data/benthos/v4/public/service"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func RegisterOtelMetricsExporter(env *service.Environment, meter metric.Meter) error {
	return env.RegisterMetricsExporter(
		"otel_collector",
		service.NewConfigSpec(),
		func(conf *service.ParsedConfig, log *service.Logger) (service.MetricsExporter, error) {
			return newOtlpMetricsExporter(meter, log), nil
		},
	)
}

var _ service.MetricsExporter = &otlpMetricsExporter{}

type otlpMetricsExporter struct {
	logger  *service.Logger
	meter   metric.Meter
	buckets []float64
}

func newOtlpMetricsExporter(meter metric.Meter, logger *service.Logger) *otlpMetricsExporter {
	return &otlpMetricsExporter{
		logger:  logger,
		meter:   meter,
		buckets: []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
	}
}

type otlpCounter struct {
	counter metric.Int64Counter
	labels  []attribute.KeyValue
}

func (c *otlpCounter) Incr(count int64) {
	c.counter.Add(context.Background(), count, metric.WithAttributes(c.labels...))
}

func (om *otlpMetricsExporter) NewCounterCtor(
	path string,
	labelNames ...string,
) service.MetricsExporterCounterCtor {
	return func(labelValues ...string) service.MetricsExporterCounter {
		var attrs []attribute.KeyValue
		for idx, label := range labelNames {
			attrs = append(attrs, attribute.String(label, labelValues[idx]))
		}

		counter, err := om.meter.Int64Counter(path)
		if err != nil {
			om.logger.Error(err.Error())
			return nil
		}
		return &otlpCounter{
			counter: counter,
			labels:  attrs,
		}
	}
}

type otlpTimer struct {
	timer  metric.Int64Histogram
	labels []attribute.KeyValue
}

func (c *otlpTimer) Timing(delta int64) {
	c.timer.Record(context.Background(), delta, metric.WithAttributes(c.labels...))
}

func (om *otlpMetricsExporter) NewTimerCtor(
	path string,
	labelNames ...string,
) service.MetricsExporterTimerCtor {
	return func(labelValues ...string) service.MetricsExporterTimer {
		var attrs []attribute.KeyValue
		for idx, label := range labelNames {
			attrs = append(attrs, attribute.String(label, labelValues[idx]))
		}

		timer, err := om.meter.Int64Histogram(
			path,
			metric.WithExplicitBucketBoundaries(om.buckets...),
		)
		if err != nil {
			om.logger.Error(err.Error())
			return nil
		}
		return &otlpTimer{
			timer:  timer,
			labels: attrs,
		}
	}
}

type otlpGauge struct {
	gauge     metric.Int64ObservableGauge
	gaugeChan chan int64
	lables    []attribute.KeyValue
}

func (c *otlpGauge) Set(value int64) {
	c.gaugeChan <- value
}

func (om *otlpMetricsExporter) NewGaugeCtor(
	path string,
	labelNames ...string,
) service.MetricsExporterGaugeCtor {
	return func(labelValues ...string) service.MetricsExporterGauge {
		var attrs []attribute.KeyValue
		for idx, label := range labelNames {
			attrs = append(attrs, attribute.String(label, labelValues[idx]))
		}

		gauge, err := om.meter.Int64ObservableGauge(
			path,
		)
		if err != nil {
			om.logger.Error(err.Error())
			return nil
		}

		gaugeChan := make(chan int64, 10)
		latestChan := make(chan int64, 1)
		latestChan <- 0
		_, err = om.meter.RegisterCallback(
			func(ctx context.Context, o metric.Observer) error {
				var last int64
				if len(gaugeChan) == 0 {
					last = <-latestChan
					latestChan <- last
					o.ObserveInt64(gauge, last)
					return nil
				}
				var value int64 = 0
				for len(gaugeChan) > 0 {
					value = <-gaugeChan
				}
				o.ObserveInt64(gauge, value)
				<-latestChan
				latestChan <- value
				return nil
			},
			gauge,
		)
		if err != nil {
			om.logger.Error(err.Error())
			return nil
		}
		return &otlpGauge{
			gauge:     gauge,
			gaugeChan: gaugeChan,
			lables:    attrs,
		}
	}
}

func (om *otlpMetricsExporter) Close(ctx context.Context) error {
	return nil
}
