package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewNotEqLabel(t *testing.T) {
	label := NewNotEqLabel("foo", "bar")
	assert.Equal(t, `foo!="bar"`, label.ToPromQueryString())
}

func Test_NewEqLabel(t *testing.T) {
	actual := NewEqLabel("foo", "bar")
	require.Equal(t, `foo="bar"`, actual.ToPromQueryString())
}

func Test_NewRegexMatchLabel(t *testing.T) {
	actual := NewRegexMatchLabel("foo", "bar")
	require.Equal(t, `foo=~"bar"`, actual.ToPromQueryString())
}

func Test_MetricLabel_ToPromQueryString(t *testing.T) {
	label := NewEqLabel("foo", "bar")
	assert.Equal(t, `foo="bar"`, label.ToPromQueryString())
}

func Test_MetricLabels_ToPromQueryString(t *testing.T) {
	labels := MetricLabels{
		NewEqLabel("foo", "bar"),
		NewEqLabel("foo2", "bar2"),
	}
	assert.Equal(
		t,
		`foo="bar",foo2="bar2"`,
		labels.ToPromQueryString(),
	)
}

func Test_MetricLabel_ToBenthosMeta(t *testing.T) {
	label := NewEqLabel("foo", "bar")
	assert.Equal(t, `meta foo = "bar"`, label.ToBenthosMeta())
}

func Test_MetricLabels_ToBenthosMeta(t *testing.T) {
	labels := MetricLabels{
		NewEqLabel("foo", "bar"),
		NewEqLabel("foo2", "bar2"),
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
