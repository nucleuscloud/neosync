package jsonanonymizer

import (
	"context"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
)

func Test_NeosyncOperator(t *testing.T) {
	t.Run("Transform", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			operator := newNeosyncOperatorApi(testutil.GetTestLogger(t))
			actual, err := operator.Transform(context.Background(), &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			}, "blah")
			require.NoError(t, err)
			require.NotEmpty(t, actual)
		})
		t.Run("default", func(t *testing.T) {
			operator := newNeosyncOperatorApi(testutil.GetTestLogger(t))
			actual, err := operator.Transform(context.Background(), &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
				},
			}, "")
			require.NoError(t, err)
			require.Empty(t, actual)
		})
	})
}
