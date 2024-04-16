package neosync_benthos_sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_combineRows(t *testing.T) {
	tests := []struct {
		name     string
		maps     []map[string]any
		expected map[string]any
	}{
		{
			name:     "empty input",
			maps:     []map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "single map",
			maps: []map[string]any{
				{"key1": "value1", "key2": 2},
			},
			expected: map[string]any{"key1": "value1", "key2": 2},
		},
		{
			name: "multiple maps with unique keys",
			maps: []map[string]any{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			expected: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name: "multiple maps with overlapping keys",
			maps: []map[string]any{
				{"key1": "value1", "key2": "value2"},
				{"key2": "newValue2", "key3": 3},
			},
			expected: map[string]any{"key1": "value1", "key2": "newValue2", "key3": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := combineRows(tt.maps)
			assert.Equal(t, actual, tt.expected)
		})
	}
}
