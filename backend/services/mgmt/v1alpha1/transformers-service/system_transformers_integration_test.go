package v1alpha1_transformersservice

import (
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetSystemTransformers() {
	resp, err := s.transformerclient.GetSystemTransformers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformersRequest{}))
	requireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetTransformers())
}

func (s *IntegrationTestSuite) Test_GetSystemTransformersBySource() {
	t := s.T()
	t.Run("ok", func(t *testing.T) {
		resp, err := s.transformerclient.GetSystemTransformerBySource(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
		}))
		requireNoErrResp(t, resp, err)
		transformer := resp.Msg.GetTransformer()
		require.NotNil(t, transformer)
		require.Equal(t, transformer.GetSource(), mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL)
	})

	t.Run("not found", func(t *testing.T) {
		resp, err := s.transformerclient.GetSystemTransformerBySource(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
		}))
		requireErrResp(t, resp, err)
		requireConnectError(s.T(), err, connect.CodeNotFound)
	})
}
