package javascript_userland

import (
	"fmt"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

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

func Test_GetSingleGenerateFunction(t *testing.T) {
	code, propertyPath := GetSingleGenerateFunction("return 'hello world';")
	require.NotEmpty(t, code)
	require.NotEmpty(t, propertyPath)

	wrappedCode := fmt.Sprintf(`
let programOutput = undefined;
const benthos = {
  v0_msg_as_structured: () => ({}),
};
const neosync = {
  patchStructuredMessage: (val) => {
    programOutput = val;
  }
};
%s
`, code)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello world", outputMap[propertyPath])
}

func Test_GetSingleTransformFunction(t *testing.T) {
	code, propertyPath := GetSingleTransformFunction("return 'hello ' + value;")
	require.NotEmpty(t, code)
	require.NotEmpty(t, propertyPath)

	wrappedCode := fmt.Sprintf(`
let programOutput = undefined;
const benthos = {
  v0_msg_as_structured: () => ({%q: "world"}),
};
const neosync = {
  patchStructuredMessage: (val) => {
    programOutput = val;
  }
};
%s
`, propertyPath, code)

	fmt.Println(wrappedCode)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello world", outputMap[propertyPath])
}

func Test_convertJsObjPathToOptionalChain(t *testing.T) {
	require.Equal(t, "address", convertJsObjPathToOptionalChain("address"))
	require.Equal(t, "address?.['city']", convertJsObjPathToOptionalChain("address.city"))
	require.Equal(t, "address?.['city']?.['state']", convertJsObjPathToOptionalChain("address.city.state"))
}
