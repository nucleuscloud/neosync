package transformers

import (
	"context"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTransformPiiTextApi struct {
	transformFn func(ctx context.Context, config *mgmtv1alpha1.TransformPiiText, value string) (string, error)
}

func (m *mockTransformPiiTextApi) Transform(ctx context.Context, config *mgmtv1alpha1.TransformPiiText, value string) (string, error) {
	return m.transformFn(ctx, config, value)
}

func TestRegisterTransformPiiText(t *testing.T) {
	env := bloblang.NewEnvironment()
	mockApi := &mockTransformPiiTextApi{
		transformFn: func(ctx context.Context, config *mgmtv1alpha1.TransformPiiText, value string) (string, error) {
			return "anonymized: " + value, nil
		},
	}

	err := RegisterTransformPiiText(env, mockApi)
	require.NoError(t, err)

	t.Run("basic usage", func(t *testing.T) {
		mapping := `transform_pii_text("sensitive data")`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with value param", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data")`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with score_threshold", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", score_threshold: 0.7)`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with language", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", language: "en")`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with allowed_phrases", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", allowed_phrases: ["data"])`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with allowed_entities", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", allowed_entities: ["PERSON", "EMAIL"])`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with default_anonymizer", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", default_anonymizer: {"replace": {"value": "REDACTED"}})`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with deny_recognizers", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", deny_recognizers: [{"name": "test", "deny_words": ["sensitive"]}])`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with entity_anonymizers", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", entity_anonymizers: {"PERSON": {"replace": {"new_value": "PERSON"}}})`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})
	t.Run("with entity_anonymizers using redact", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", entity_anonymizers: {"PERSON": {"redact": {}}})`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with entity_anonymizers using mask", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", entity_anonymizers: {"PERSON": {"mask": {"masking_char": "*", "chars_to_mask": 4, "from_end": true}}})`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with entity_anonymizers using hash", func(t *testing.T) {
		mapping := `transform_pii_text(value: "sensitive data", entity_anonymizers: {"PERSON": {"hash": {"algo": 1}}})`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query(nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})

	t.Run("with value from context", func(t *testing.T) {
		mapping := `transform_pii_text(value: this)`
		exec, err := env.Parse(mapping)
		require.NoError(t, err)

		res, err := exec.Query("sensitive data")
		require.NoError(t, err)
		require.NotNil(t, res)
		output, ok := res.(*string)
		require.True(t, ok)
		assert.Equal(t, "anonymized: sensitive data", *output)
	})
}
