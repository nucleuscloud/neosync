package integrationtests_test

import (
	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_TransformersService_GetSystemTransformers() {
	resp, err := s.unauthdClients.transformers.GetSystemTransformers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformersRequest{}))
	requireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetTransformers())
}

func (s *IntegrationTestSuite) Test_TransformersService_GetSystemTransformersBySource_Ok() {
	t := s.T()
	resp, err := s.unauthdClients.transformers.GetSystemTransformerBySource(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
	}))
	requireNoErrResp(t, resp, err)
	transformer := resp.Msg.GetTransformer()
	require.NotNil(t, transformer)
	require.Equal(t, transformer.GetSource(), mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL)
}

func (s *IntegrationTestSuite) Test_TransformersService_GetSystemTransformersBySource_NotFound() {
	t := s.T()
	resp, err := s.unauthdClients.transformers.GetSystemTransformerBySource(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
	}))
	requireErrResp(t, resp, err)
	requireConnectError(s.T(), err, connect.CodeNotFound)
}
