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
	t.Parallel()
	t.Run("string", func(t *testing.T) {
		t.Parallel()
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

		runTestProgram(t, wrappedCode, propertyPath, "hello world")
	})
	t.Run("number", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleGenerateFunction("return 123;")
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

		runTestProgram(t, wrappedCode, propertyPath, int64(123))
	})

	t.Run("boolean", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleGenerateFunction("return true;")
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

		runTestProgram(t, wrappedCode, propertyPath, true)
	})

	t.Run("object", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleGenerateFunction("return {a: 1, b: 2};")
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

		runTestProgram(t, wrappedCode, propertyPath, map[string]any{"a": int64(1), "b": int64(2)})
	})

	t.Run("array", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleGenerateFunction("return [1, 2, 3];")
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

		runTestProgram(t, wrappedCode, propertyPath, []any{int64(1), int64(2), int64(3)})
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleGenerateFunction("return null;")
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

		runTestProgram(t, wrappedCode, propertyPath, nil)
	})
	t.Run("undefined", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleGenerateFunction("return undefined;")
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

		runTestProgram(t, wrappedCode, propertyPath, nil)
	})
}

func Test_GetSingleTransformFunction(t *testing.T) {
	t.Parallel()
	t.Run("string", func(t *testing.T) {
		t.Parallel()
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

		runTestProgram(t, wrappedCode, propertyPath, "hello world")
	})

	t.Run("number", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleTransformFunction("return value + 1;")
		require.NotEmpty(t, code)
		require.NotEmpty(t, propertyPath)

		wrappedCode := fmt.Sprintf(`
		let programOutput = undefined;
		const benthos = {
			v0_msg_as_structured: () => ({%q: 123}),
		};
		const neosync = {
			patchStructuredMessage: (val) => {
				programOutput = val;
			}
		};
		%s
		`, propertyPath, code)

		runTestProgram(t, wrappedCode, propertyPath, int64(124))
	})

	t.Run("boolean", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleTransformFunction("return !value;")
		require.NotEmpty(t, code)
		require.NotEmpty(t, propertyPath)

		wrappedCode := fmt.Sprintf(`
		let programOutput = undefined;
		const benthos = {
			v0_msg_as_structured: () => ({%q: true}),
		};
		const neosync = {
			patchStructuredMessage: (val) => {
				programOutput = val;
			}
		};
		%s
		`, propertyPath, code)

		runTestProgram(t, wrappedCode, propertyPath, false)
	})

	t.Run("object", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleTransformFunction("return { ...value, c: 3 };")
		require.NotEmpty(t, code)
		require.NotEmpty(t, propertyPath)

		wrappedCode := fmt.Sprintf(`
		let programOutput = undefined;
		const benthos = {
			v0_msg_as_structured: () => ({%q: {a: 1, b: 2}}),
		};
		const neosync = {
			patchStructuredMessage: (val) => {
				programOutput = val;
			}
		};
		%s
		`, propertyPath, code)

		runTestProgram(t, wrappedCode, propertyPath, map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)})
	})

	t.Run("array", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleTransformFunction("return [...value, 3];")
		require.NotEmpty(t, code)
		require.NotEmpty(t, propertyPath)

		wrappedCode := fmt.Sprintf(`
		let programOutput = undefined;
		const benthos = {
			v0_msg_as_structured: () => ({%q: [1, 2]}),
		};
		const neosync = {
			patchStructuredMessage: (val) => {
				programOutput = val;
			}
		};
		%s
		`, propertyPath, code)

		runTestProgram(t, wrappedCode, propertyPath, []any{int64(1), int64(2), int64(3)})
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleTransformFunction("return value;")
		require.NotEmpty(t, code)
		require.NotEmpty(t, propertyPath)

		wrappedCode := fmt.Sprintf(`
		let programOutput = undefined;
		const benthos = {
			v0_msg_as_structured: () => ({%q: null}),
		};
		const neosync = {
			patchStructuredMessage: (val) => {
				programOutput = val;
			}
		};
		%s
		`, propertyPath, code)

		runTestProgram(t, wrappedCode, propertyPath, nil)
	})
	t.Run("undefined", func(t *testing.T) {
		t.Parallel()
		code, propertyPath := GetSingleTransformFunction("return value;")
		require.NotEmpty(t, code)
		require.NotEmpty(t, propertyPath)

		wrappedCode := fmt.Sprintf(`
		let programOutput = undefined;
		const benthos = {
			v0_msg_as_structured: () => ({%q: undefined}),
		};
		const neosync = {
			patchStructuredMessage: (val) => {
				programOutput = val;
			}
		};
		%s
		`, propertyPath, code)

		runTestProgram(t, wrappedCode, propertyPath, nil)
	})
}

func Test_convertJsObjPathToOptionalChain(t *testing.T) {
	require.Equal(t, "address", convertJsObjPathToOptionalChain("address"))
	require.Equal(t, "address?.['city']", convertJsObjPathToOptionalChain("address.city"))
	require.Equal(t, "address?.['city']?.['state']", convertJsObjPathToOptionalChain("address.city.state"))
}

func runTestProgram(t testing.TB, code string, propertyPath string, expectedOutput any) {
	t.Helper()
	program, err := goja.Compile("test.js", code, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, expectedOutput, outputMap[propertyPath])
}
