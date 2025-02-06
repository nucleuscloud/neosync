package benthos_metrics

import (
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

type intCtorExpanded interface {
	Incr(i int64)
}

type intGagExpanded interface {
	Set(i int64)
}

func Test_RegisterOtelMetricsExporter(t *testing.T) {
	testenv := service.NewEnvironment()
	meter := getTestMeter(t, "test")

	err := RegisterOtelMetricsExporter(testenv, meter)
	assert.NoError(t, err)
}

func getTestMeter(t testing.TB, name string) metric.Meter {
	t.Helper()
	provider := metricsdk.NewMeterProvider()
	return provider.Meter(name)
}

func TestOtelMetrics(t *testing.T) {
	meter := getTestMeter(t, "test")
	nm := newOtlpMetricsExporter(meter, nil)

	ctr := nm.NewCounterCtor("counterone")()
	ctr.Incr(10)
	ctr.Incr(11)

	gge := nm.NewGaugeCtor("gaugeone")()
	gge.Set(12)

	tmr := nm.NewTimerCtor("timerone")()
	tmr.Timing(13)

	ctrTwo := nm.NewCounterCtor("countertwo", "label1")
	ctrTwo("value1").Incr(10)
	ctrTwo("value2").Incr(11)
	ctrTwo("value3").(intCtorExpanded).Incr(10)

	ggeTwo := nm.NewGaugeCtor("gaugetwo", "label2")
	ggeTwo("value3").Set(12)

	ggeThree := nm.NewGaugeCtor("gaugethree")()
	ggeThree.(intGagExpanded).Set(11)

	tmrTwo := nm.NewTimerCtor("timertwo", "label3", "label4")
	tmrTwo("value4", "value5").Timing(13)
}

func TestOtelHistMetrics(t *testing.T) {
	meter := getTestMeter(t, "test")
	nm := newOtlpMetricsExporter(meter, nil)

	applyTestMetrics(nm)

	tmr := nm.NewTimerCtor("timerone")()
	tmr.Timing(13)
	tmrTwo := nm.NewTimerCtor("timertwo", "label3", "label4")
	tmrTwo("value4", "value5").Timing(14)
}

func applyTestMetrics(nm *otlpMetricsExporter) {
	ctr := nm.NewCounterCtor("counterone")()
	ctr.Incr(10)
	ctr.Incr(11)

	gge := nm.NewGaugeCtor("gaugeone")()
	gge.Set(12)

	ctrTwo := nm.NewCounterCtor("countertwo", "label1")
	ctrTwo("value1").Incr(10)
	ctrTwo("value2").Incr(11)

	ggeTwo := nm.NewGaugeCtor("gaugetwo", "label2")
	ggeTwo("value3").Set(12)
}
