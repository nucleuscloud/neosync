package metrics

import (
	"fmt"
	"strings"
)

const (
	AccountIdLabel     = "neosyncAccountId"
	JobIdLabel         = "neosyncJobId"
	TemporalWorkflowId = "temporalWorkflowId"
	TemporalRunId      = "temporalRunId"

	TableSchemaLabel    = "tableSchema"
	TableNameLabel      = "tableName"
	JobTypeLabel        = "jobType"
	IsUpdateConfigLabel = "isUpdateConfig"
)

func NewEqLabel(key, value string) MetricLabel {
	return MetricLabel{Key: key, Value: value, Sign: "="}
}

// note: this has only been tested with prometheus and using it with benthos is not currently supported
func NewNotEqLabel(key, value string) MetricLabel {
	return MetricLabel{Key: key, Value: value, Sign: "!="}
}

type MetricLabel struct {
	Key   string
	Value string
	Sign  string
}

func (m *MetricLabel) ToPromQueryString() string {
	return fmt.Sprintf("%s%s%q", m.Key, m.Sign, m.Value)
}

func (m *MetricLabel) ToBenthosMeta() string {
	return fmt.Sprintf("meta %s %s %q", m.Key, m.Sign, m.Value)
}

type MetricLabels []MetricLabel

func (m MetricLabels) ToPromQueryString() string {
	var parts []string
	for _, label := range m {
		parts = append(parts, label.ToPromQueryString())
	}
	return strings.Join(parts, ",")
}

func (m MetricLabels) ToBenthosMeta() string {
	var parts []string
	for _, label := range m {
		parts = append(parts, label.ToBenthosMeta())
	}
	return strings.Join(parts, "\n")
}
