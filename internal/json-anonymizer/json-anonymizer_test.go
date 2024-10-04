package jsonanonymizer

import (
	"encoding/json"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_NewAnonymizer(t *testing.T) {
	t.Run("Initialize with no options", func(t *testing.T) {
		anonymizer, err := NewAnonymizer()
		require.Error(t, err)
		require.Nil(t, anonymizer)
	})

	t.Run("Initialize with transformer mappings", func(t *testing.T) {
		mappings := map[string]*mgmtv1alpha1.TransformerConfig{
			".field1": {
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{},
			},
		}
		anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings))
		require.NoError(t, err)
		require.NotNil(t, anonymizer)
		require.Equal(t, mappings, anonymizer.transformerMappings)
		require.NotEmpty(t, anonymizer.transformerExecutors)
	})

	t.Run("Initialize with default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults))
		require.NoError(t, err)
		require.NotNil(t, anonymizer)
		require.Equal(t, defaults, anonymizer.defaultTransformers)
		require.NotNil(t, anonymizer.defaultTransformerExecutor)
	})

	t.Run("Initialize with halt on failure", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults), WithHaltOnFailure(true))
		require.NoError(t, err)
		require.NotNil(t, anonymizer)
		require.True(t, anonymizer.haltOnFailure)
	})
}

func Test_AnonymizeJSONObjects(t *testing.T) {
	t.Run("Anonymize with transformer mappings", func(t *testing.T) {
		mappings := map[string]*mgmtv1alpha1.TransformerConfig{
			".name": {
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings))
		require.NoError(t, err)

		input := `{"name": "John Doe", "age": 30}`
		output, anonErrors := anonymizer.AnonymizeJSONObjects([]string{input})
		require.Empty(t, anonErrors)

		var result map[string]any
		err = json.Unmarshal([]byte(output[0]), &result)
		require.NoError(t, err)
		require.NotEqual(t, "John Doe", result["name"])
		require.Equal(t, float64(30), result["age"])
	})

	t.Run("Anonymize with default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults))
		require.NoError(t, err)

		input := `{"name": "John Doe", "city": "New York"}`
		output, anonErrors := anonymizer.AnonymizeJSONObjects([]string{input})
		require.Empty(t, anonErrors)

		var result map[string]any
		err = json.Unmarshal([]byte(output[0]), &result)
		require.NoError(t, err)
		require.NotEqual(t, "John Doe", result["name"])
		require.NotEqual(t, "New York", result["city"])
	})

	t.Run("Anonymize with invalid JSON", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults))
		require.NoError(t, err)

		input := `invalid json`
		_, anonErrors := anonymizer.AnonymizeJSONObjects([]string{input})
		require.NotEmpty(t, anonErrors)
	})
}

func Test_GenerateFunctionName(t *testing.T) {
	t.Run("Generate unique function names", func(t *testing.T) {
		name1 := generateFunctionName("field1")
		name2 := generateFunctionName("field2")
		require.NotEqual(t, name1, name2)
	})

	t.Run("Function name starts with a letter", func(t *testing.T) {
		name := generateFunctionName("field")
		require.Regexp(t, "^[a-z]", name)
	})
}

func Test_DerefPointer(t *testing.T) {
	t.Run("Deref string pointer", func(t *testing.T) {
		str := "test"
		ptr := &str
		result := derefPointer(ptr)
		require.Equal(t, str, result)
	})

	t.Run("Deref nil pointer", func(t *testing.T) {
		var ptr *string
		result := derefPointer(ptr)
		require.Nil(t, result)
	})

	t.Run("Deref non-pointer", func(t *testing.T) {
		value := 42
		result := derefPointer(value)
		require.Equal(t, value, result)
	})
}

func Test_InitTransformerExecutors(t *testing.T) {
	t.Run("Initialize valid transformer", func(t *testing.T) {
		mappings := map[string]*mgmtv1alpha1.TransformerConfig{
			"field1": {
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		executors, err := initTransformerExecutors(mappings)
		require.NoError(t, err)
		require.Len(t, executors, 1)
		require.NotNil(t, executors["field1"])
	})

	t.Run("Initialize invalid transformer", func(t *testing.T) {
		mappings := map[string]*mgmtv1alpha1.TransformerConfig{
			"field1": {
				Config: nil,
			},
		}
		_, err := initTransformerExecutors(mappings)
		require.Error(t, err)
	})
}

func Test_InitDefaultTransformerExecutors(t *testing.T) {
	t.Run("Initialize all default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
				},
			},
			N: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{},
				},
			},
			Boolean: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
			},
		}
		executors, err := initDefaultTransformerExecutors(defaults)
		require.NoError(t, err)
		require.NotNil(t, executors.S)
		require.NotNil(t, executors.N)
		require.NotNil(t, executors.Boolean)
	})

	t.Run("Initialize partial default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
				},
			},
		}
		executors, err := initDefaultTransformerExecutors(defaults)
		require.NoError(t, err)
		require.NotNil(t, executors.S)
		require.Nil(t, executors.N)
		require.Nil(t, executors.Boolean)
	})
}
