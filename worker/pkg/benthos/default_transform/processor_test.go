package neosync_benthos_defaulttransform

import (
	"testing"

	transformer_executor "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformer_executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_transformRoot(t *testing.T) {
	tests := []struct {
		name           string
		input          any
		path           string
		mappedKeys     map[string]struct{}
		expectedOutput any
		expectError    bool
	}{
		{
			name:           "Simple string",
			input:          "hello",
			path:           "",
			mappedKeys:     map[string]struct{}{},
			expectedOutput: "transformed_hello",
			expectError:    false,
		},
		{
			name:           "Mapped key",
			input:          "hello",
			path:           "mapped_key",
			mappedKeys:     map[string]struct{}{"mapped_key": {}},
			expectedOutput: "hello",
			expectError:    false,
		},
		{
			name: "Nested map",
			input: map[string]any{
				"key1": "value1",
				"key2": map[string]any{
					"nested": "nestedvalue",
				},
			},
			path:       "",
			mappedKeys: map[string]struct{}{},
			expectedOutput: map[string]any{
				"key1": "transformed_value1",
				"key2": map[string]any{
					"nested": "transformed_nestedvalue",
				},
			},
			expectError: false,
		},
		{
			name:           "Slice of any",
			input:          []any{"hello", 123, true},
			path:           "",
			mappedKeys:     map[string]struct{}{},
			expectedOutput: []any{"transformed_hello", 246, false},
			expectError:    false,
		},
		{
			name:           "Byte slice",
			input:          []byte("hello"),
			path:           "",
			mappedKeys:     map[string]struct{}{},
			expectedOutput: []byte("transformed_hello"),
			expectError:    false,
		},
		{
			name:           "Boolean",
			input:          true,
			path:           "",
			mappedKeys:     map[string]struct{}{},
			expectedOutput: false, // Assuming the mock transformer inverts booleans
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProcessor := createMockProcessor(tt.mappedKeys)
			output, err := mockProcessor.transformRoot(tt.path, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func createMockProcessor(mappedKeys map[string]struct{}) *defaultTransformerProcessor {
	return &defaultTransformerProcessor{
		mappedKeys: mappedKeys,
		defaultTransformersInitMap: map[primitiveType]*transformer_executor.TransformerExecutor{
			String: {
				Mutate: func(value any, opts any) (any, error) {
					return "transformed_" + value.(string), nil
				},
			},
			Number: {
				Mutate: func(value any, opts any) (any, error) {
					return value.(int) * 2, nil
				},
			},
			Boolean: {
				Mutate: func(value any, opts any) (any, error) {
					return !value.(bool), nil
				},
			},
			Byte: {
				Mutate: func(value any, opts any) (any, error) {
					return []byte("transformed_" + string(value.([]byte))), nil
				},
			},
		},
	}
}
