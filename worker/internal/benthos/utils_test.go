package neosync_benthos

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuildBenthosTable(t *testing.T) {
	assert.Equal(t, BuildBenthosTable("public", "users"), "public.users", "Joins schema and table with a dot")
	assert.Equal(t, BuildBenthosTable("", "users"), "users", "Handles an empty schema")
}

func Test_IsCriticalError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"Partial match", "pq: duplicate key value violates unique constraint jobs_pkey", true},
		{"Match with different case", "pq: insert or update on table \"employees\" violates foreign key constraint", true},
		{"No match", "connection timed out", false},
		{"Unrelated error message", "could not connect to server: Connection refused", false},
		{"Unrelated error message", "too many clients already", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsCriticalError(tt.errMsg)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_ContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Hello, World", "world", true},
		{"Go is fun", "FUN", true},
		{"Case-INSENSITIVE", "case-insensitive", true},
		{"Test", "best", false},
		{"Partial", "art", true},
		{"", "", true},
		{"", "non-empty", false},
		{"non-empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+" contains "+tt.substr, func(t *testing.T) {
			actual := containsIgnoreCase(tt.s, tt.substr)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_HashBenthosCacheKey(t *testing.T) {
	tests := []struct {
		jobId, runId, table, col string
		expected                 string
	}{
		{"job1", "run1", "table1", "col1", ToSha256("job1.run1.table1.col1")},
		{"", "", "", "", ToSha256("...")},
		{"job2", "run2", "table2", "", ToSha256("job2.run2.table2.")},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s.%s.%s.%s", tt.jobId, tt.runId, tt.table, tt.col), func(t *testing.T) {
			actual := HashBenthosCacheKey(tt.jobId, tt.runId, tt.table, tt.col)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
