package genbenthosconfigs_activity

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

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
