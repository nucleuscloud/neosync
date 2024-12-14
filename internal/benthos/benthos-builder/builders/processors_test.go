package benthosbuilder_builders

import (
	"context"
	"fmt"
	"testing"

	"github.com/dop251/goja"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
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
			actual := sanitizeJsFunctionName(tt.input)
			if actual != tt.expected {
				t.Errorf("sanitizeJsFunctionName(%q) = %q; expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func Test_buildProcessorConfigsJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: `return "hello " + value;`,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(
		ctx, mockTransformerClient,
		[]*mgmtv1alpha1.JobMapping{
			{
				Schema: "public", Table: "users", Column: "address",
				Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config},
			}},
		map[string]*sqlmanager_shared.DatabaseSchemaRow{},
		map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, nil,
		tabledependency.NewRunConfig("", tabledependency.RunTypeInsert, nil, nil, nil, []string{"address"}, nil, nil, false),
		nil,
		[]string{},
	)

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotNil(t, res[0].NeosyncJavascript)
	require.NotNil(t, res[0].NeosyncJavascript.Code)

	wrappedCode := fmt.Sprintf(`
let programOutput = undefined;
const benthos = {
  v0_msg_as_structured: () => ({address: "world"}),
};
const neosync = {
  patchStructuredMessage: (val) => {
    programOutput = val;
  }
};
%s
	`, res[0].NeosyncJavascript.Code)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello world", outputMap["address"])
}

func Test_buildProcessorConfigsGenerateJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()
	genCode := `return "hello world";`

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
					Code: genCode,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(
		ctx, mockTransformerClient,
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "test",
				Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config},
			}},
		map[string]*sqlmanager_shared.DatabaseSchemaRow{},
		map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, nil,
		tabledependency.NewRunConfig("", tabledependency.RunTypeInsert, nil, nil, nil, []string{"test"}, nil, nil, false),
		nil,
		[]string{},
	)

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotNil(t, res[0].NeosyncJavascript)
	require.NotNil(t, res[0].NeosyncJavascript.Code)

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
	`, res[0].NeosyncJavascript.Code)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello world", outputMap["test"])
}

func Test_buildProcessorConfigsJavascriptMultiple(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	nameCol := "name"
	ageCol := "age"

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: `return "hello " + value;`,
				},
			},
		},
	}

	jsT2 := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: `return value + 2;`,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(
		ctx, mockTransformerClient,
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: nameCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config}},
			{Schema: "public", Table: "users", Column: ageCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT2.Config}}},
		map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, nil,
		tabledependency.NewRunConfig("", tabledependency.RunTypeInsert, nil, nil, nil, []string{nameCol, ageCol}, nil, nil, false),
		nil,
		[]string{},
	)

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotNil(t, res[0].NeosyncJavascript)
	require.NotNil(t, res[0].NeosyncJavascript.Code)

	wrappedCode := fmt.Sprintf(`
let programOutput = undefined;
const benthos = {
  v0_msg_as_structured: () => ({"name": "world", "age": 2}),
};
const neosync = {
  patchStructuredMessage: (val) => {
    programOutput = val;
  }
};
%s
	`, res[0].NeosyncJavascript.Code)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello world", outputMap["name"])
	require.Equal(t, int64(4), outputMap["age"])
}

func Test_buildProcessorConfigsTransformAndGenerateJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	nameCol := "name"
	col2 := "test"

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: `return "hello " + value;`,
				},
			},
		},
	}

	jsT2 := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
					Code: `return "test";`,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(
		ctx, mockTransformerClient,
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: nameCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config}},
			{Schema: "public", Table: "users", Column: col2, Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT2.Config}}},
		map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, nil,
		tabledependency.NewRunConfig("", tabledependency.RunTypeInsert, nil, nil, nil, []string{nameCol, col2}, nil, nil, false),
		nil,
		[]string{},
	)

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotNil(t, res[0].NeosyncJavascript)
	require.NotNil(t, res[0].NeosyncJavascript.Code)

	wrappedCode := fmt.Sprintf(`
let programOutput = undefined;
const benthos = {
  v0_msg_as_structured: () => ({"name": "world"}),
};
const neosync = {
  patchStructuredMessage: (val) => {
    programOutput = val;
  }
};
%s
	`, res[0].NeosyncJavascript.Code)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello world", outputMap[nameCol])
	require.Equal(t, "test", outputMap[col2])
}

func Test_buildProcessorConfigsJavascript_DeepKeys(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: `return "hello " + value;`,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(
		ctx, mockTransformerClient,
		[]*mgmtv1alpha1.JobMapping{
			{
				Schema: "public", Table: "users", Column: "foo.bar.baz",
				Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config},
			}},
		map[string]*sqlmanager_shared.DatabaseSchemaRow{},
		map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, nil,
		tabledependency.NewRunConfig("", tabledependency.RunTypeInsert, nil, nil, nil, []string{"foo.bar.baz"}, nil, nil, false),
		nil,
		[]string{},
	)

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotNil(t, res[0].NeosyncJavascript)
	require.NotNil(t, res[0].NeosyncJavascript.Code)

	wrappedCode := fmt.Sprintf(`
let programOutput = undefined;
const benthos = {
  v0_msg_as_structured: () => ({foo: {bar: {baz: "world"}}}),
};
const neosync = {
  patchStructuredMessage: (val) => {
    programOutput = val;
  }
};
%s
	`, res[0].NeosyncJavascript.Code)

	program, err := goja.Compile("test.js", wrappedCode, true)
	require.NoError(t, err)
	rt := goja.New()
	_, err = rt.RunProgram(program)
	require.NoError(t, err)
	programOutput := rt.Get("programOutput").Export()
	require.NotNil(t, programOutput)
	outputMap, ok := programOutput.(map[string]any)
	require.True(t, ok)
	require.NotNil(t, outputMap)
	require.Equal(t, "hello world", outputMap["foo.bar.baz"])
}

func Test_shouldProcessColumn(t *testing.T) {
	t.Run("no - passthrough", func(t *testing.T) {
		actual := shouldProcessColumn(&mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
			},
		})
		require.False(t, actual)
	})
	t.Run("no - nil", func(t *testing.T) {
		actual := shouldProcessColumn(nil)
		require.False(t, actual)
	})
	t.Run("yes", func(t *testing.T) {
		actual := shouldProcessColumn(&mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
			},
		})
		require.True(t, actual)
	})
}

func Test_shouldProcessStrict(t *testing.T) {
	t.Run("no - passthrough", func(t *testing.T) {
		actual := shouldProcessStrict(&mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
			},
		})
		require.False(t, actual)
	})
	t.Run("no - default", func(t *testing.T) {
		actual := shouldProcessStrict(&mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{},
			},
		})
		require.False(t, actual)
	})
	t.Run("no - null", func(t *testing.T) {
		actual := shouldProcessStrict(&mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{},
			},
		})
		require.False(t, actual)
	})
	t.Run("no - nil", func(t *testing.T) {
		actual := shouldProcessStrict(nil)
		require.False(t, actual)
	})
	t.Run("yes", func(t *testing.T) {
		actual := shouldProcessStrict(&mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
			},
		})
		require.True(t, actual)
	})
}
