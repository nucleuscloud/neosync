package benthosbuilder_builders

import (
	"context"
	"fmt"
	"testing"

	"github.com/dop251/goja"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	"github.com/stretchr/testify/require"
)

func Test_buildProcessorConfigsJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: `return "hello " + value + " " + input.extra;`,
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
		runconfigs.NewRunConfig(sqlmanager_shared.SchemaTable{}, runconfigs.RunTypeInsert, nil, nil, nil, []string{"address"}, nil, false),
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
  v0_msg_as_structured: () => ({address: "world", extra: "foobar"}),
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
	require.Equal(t, "hello world foobar", outputMap["address"])
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
		runconfigs.NewRunConfig(sqlmanager_shared.SchemaTable{}, runconfigs.RunTypeInsert, nil, nil, nil, []string{"test"}, nil, false),
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
		runconfigs.NewRunConfig(sqlmanager_shared.SchemaTable{}, runconfigs.RunTypeInsert, nil, nil, nil, []string{nameCol, ageCol}, nil, false),
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
		runconfigs.NewRunConfig(sqlmanager_shared.SchemaTable{}, runconfigs.RunTypeInsert, nil, nil, nil, []string{nameCol, col2}, nil, false),
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
		runconfigs.NewRunConfig(sqlmanager_shared.SchemaTable{}, runconfigs.RunTypeInsert, nil, nil, nil, []string{"foo.bar.baz"}, nil, false),
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
