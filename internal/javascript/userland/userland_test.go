package javascript_userland

import "testing"

func Test_sanitizeFunctionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123my Function!", "_123my_Function_"},
		{"validName", "validName"},
		{"name_with_underscores", "name_with_underscores"},
		{"$dollarSign", "$dollarSign"},
		{"invalid-char$", "invalid_char$"},
		{"spaces in name", "spaces_in_name"},
		{"!@#$%^&*()_+=", "___$_________"},
		{"_leadingUnderscore", "_leadingUnderscore"},
		{"$startingDollarSign", "$startingDollarSign"},
		{"endingWithNumber1", "endingWithNumber1"},
		{"functionName123", "functionName123"},
		{"中文字符", "中文字符"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := sanitizeFunctionName(tt.input)
			if actual != tt.expected {
				t.Errorf("sanitizeJsFunctionName(%q) = %q; expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}
