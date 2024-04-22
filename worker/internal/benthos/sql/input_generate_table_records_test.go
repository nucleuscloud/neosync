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

func TestCombineRowLists(t *testing.T) {
	// Test case 1: Empty input
	emptyResult := combineRowLists([][]map[string]any{})
	assert.Empty(t, emptyResult)

	// Test case 2: Single row with single map
	singleRowSingleMap := [][]map[string]any{{{"a": 1, "b": 2}}}
	singleResult := combineRowLists(singleRowSingleMap)
	expectedSingleResult := []map[string]any{{"a": 1, "b": 2}}
	assert.Equal(t, expectedSingleResult, singleResult)

	// Test case 3: Single row with multiple maps
	singleRowMultipleMaps := [][]map[string]any{{{"a": 1}}, {{"b": 2}}, {{"c": 3}}}
	multipleResult := combineRowLists(singleRowMultipleMaps)
	expectedMultipleResult := []map[string]any{{"a": 1, "b": 2, "c": 3}}
	assert.Equal(t, expectedMultipleResult, multipleResult)

	// Test case 4: Multiple rows with multiple maps
	multipleRowsMultipleMaps := [][]map[string]any{
		{{"a": 1}, {"b": 2}, {"c": 3}},
		{{"d": 4}, {"e": 5}, {"f": 6}},
		{{"g": 7}, {"h": 8}, {"i": 9}},
	}
	multipleRowsResult := combineRowLists(multipleRowsMultipleMaps)
	expectedMultipleRowsResult := []map[string]any{
		{"a": 1, "d": 4, "g": 7},
		{"b": 2, "e": 5, "h": 8},
		{"c": 3, "f": 6, "i": 9},
	}
	assert.Equal(t, expectedMultipleRowsResult, multipleRowsResult)
}
