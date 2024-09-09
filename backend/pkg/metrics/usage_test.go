package metrics

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func Test_getPromQueryFromMetric(t *testing.T) {
	output, err := GetPromQueryFromMetric(
		mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		MetricLabels{NewEqLabel("foo", "bar"), NewEqLabel("foo2", "bar2")},
		"1d",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(
		t,
		`sum(max_over_time(input_received_total{foo="bar",foo2="bar2"}[1d]))`,
		output,
	)
}

func Test_getPromQueryFromMetric_Invalid_Metric(t *testing.T) {
	output, err := GetPromQueryFromMetric(
		mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_UNSPECIFIED,
		MetricLabels{NewEqLabel("foo", "bar"), NewEqLabel("foo2", "bar2")},
		"1d",
	)
	assert.Error(t, err)
	assert.Empty(t, output)
}
