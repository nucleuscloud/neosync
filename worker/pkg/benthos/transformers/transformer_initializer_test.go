package transformers

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_InitializeTransformerByConfigType(t *testing.T) {
	t.Run("PassthroughConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test", nil)
		require.NoError(t, err)
		require.Equal(t, "test", result)
	})

	t.Run("GenerateCategoricalConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
				GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
					Categories: "A,B,C",
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Contains(t, []string{"A", "B", "C"}, result)
	})

	t.Run("GenerateCategoricalConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
				GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateBoolConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, bool(true), result)
	})

	t.Run("TransformStringConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
				TransformStringConfig: &mgmtv1alpha1.TransformString{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test", executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "test", result)
		require.Len(t, result.(string), 4)
	})

	t.Run("TransformInt64Config", func(t *testing.T) {
		min, max := int64(1), int64(100)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
				TransformInt64Config: &mgmtv1alpha1.TransformInt64{
					RandomizationRangeMin: min,
					RandomizationRangeMax: max,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(int64(50), executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.GreaterOrEqual(t, result.(int64), min)
		require.LessOrEqual(t, result.(int64), max)
	})

	t.Run("TransformFullNameConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
				TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("John Doe", executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "John Doe", result)
		require.Len(t, result.(string), 8)
	})

	t.Run("GenerateEmailConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, result)
	})

	t.Run("TransformEmailConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test@example.com", executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^[a-zA-Z0-9._%+-]+@example\.com$`, result)
		require.Len(t, result.(string), 16)
	})

	// Add more subtests for other config types...

	t.Run("UnsupportedConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: nil,
		}
		_, err := InitializeTransformerByConfigType(config)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported transformer")
	})
}
