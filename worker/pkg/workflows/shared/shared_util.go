package workflow_shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SanitizeWorkflowID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces special characters",
			input:    "public.users@123",
			expected: "public_users_123",
		},
		{
			name:     "keeps valid characters",
			input:    "public-users-123",
			expected: "public-users-123",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeWorkflowID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
