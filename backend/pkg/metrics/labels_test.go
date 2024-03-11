package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewNotEqLabel(t *testing.T) {
	label := NewNotEqLabel("foo", "bar")
	assert.Equal(t, `foo!="bar"`, label.ToPromQueryString())
}

func Test_metricLabel_ToPromQueryString(t *testing.T) {
	label := NewEqLabel("foo", "bar")
	assert.Equal(t, `foo="bar"`, label.ToPromQueryString())
}

func Test_metricLabels_ToPromQueryString(t *testing.T) {
	labels := MetricLabels{
		{Key: "foo", Value: "bar"},
		{Key: "foo2", Value: "bar2"},
	}
	assert.Equal(
		t,
		`foo="bar",foo2="bar2"`,
		labels.ToPromQueryString(),
	)
}

func Test_metricLabel_ToBenthosMeta(t *testing.T) {
	label := NewEqLabel("foo", "bar")
	assert.Equal(t, `meta foo = "bar"`, label.ToBenthosMeta())
}

func Test_metricLabels_ToBenthosMeta(t *testing.T) {
	labels := MetricLabels{
		{Key: "foo", Value: "bar"},
		{Key: "foo2", Value: "bar2"},
	}
	assert.Equal(
		t,
		strings.TrimSpace(`
meta foo = "bar"
meta foo2 = "bar2"
`),
		labels.ToBenthosMeta(),
	)
}
