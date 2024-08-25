package v1alpha1_transformersservice

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetUserDefinedTransformers_Empty() {
	t := s.T()
	accountId := s.createPersonalAccount(s.userclient)
	resp, err := s.transformerclient.GetUserDefinedTransformers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformersRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(t, resp, err)
	require.Empty(t, resp.Msg.GetTransformers())
}

func (s *IntegrationTestSuite) Test_GetUserDefinedTransformers_NotEmpty() {
	t := s.T()

	accountId := s.createPersonalAccount(s.userclient)
	createResp, err := s.transformerclient.CreateUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateUserDefinedTransformerRequest{
		AccountId:   accountId,
		Name:        "my-test-transformer",
		Description: "this is a test",
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
		TransformerConfig: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
				GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
			},
		},
	}))
	requireNoErrResp(t, createResp, err)

	resp, err := s.transformerclient.GetUserDefinedTransformers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformersRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(t, resp, err)
	udTransformers := resp.Msg.GetTransformers()
	require.NotEmpty(t, udTransformers)
}

func (s *IntegrationTestSuite) Test_GetUserDefinedTransformerById_Found() {
	t := s.T()
	accountId := s.createPersonalAccount(s.userclient)
	createResp, err := s.transformerclient.CreateUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateUserDefinedTransformerRequest{
		AccountId:   accountId,
		Name:        "my-test-transformer",
		Description: "this is a test",
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
		TransformerConfig: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
				GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
			},
		},
	}))
	requireNoErrResp(t, createResp, err)

	resp, err := s.transformerclient.GetUserDefinedTransformerById(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
		TransformerId: createResp.Msg.GetTransformer().GetId(),
	}))
	requireNoErrResp(t, resp, err)
	require.Equal(t, createResp.Msg.GetTransformer().GetId(), resp.Msg.GetTransformer().GetId())

	t.Run("not found", func(t *testing.T) {
		resp, err := s.transformerclient.GetUserDefinedTransformerById(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
			TransformerId: uuid.NewString(),
		}))
		requireErrResp(t, resp, err)
		requireConnectError(t, err, connect.CodeNotFound)
	})
}

func (s *IntegrationTestSuite) Test_GetUserDefinedTransformerById_NotFound() {
	t := s.T()
	resp, err := s.transformerclient.GetUserDefinedTransformerById(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
		TransformerId: uuid.NewString(),
	}))
	requireErrResp(t, resp, err)
	requireConnectError(t, err, connect.CodeNotFound)
}

func (s *IntegrationTestSuite) Test_CreateUserDefinedTransformer() {
	accountId := s.createPersonalAccount(s.userclient)
	createResp, err := s.transformerclient.CreateUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateUserDefinedTransformerRequest{
		AccountId:   accountId,
		Name:        "my-test-transformer",
		Description: "this is a test",
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
		TransformerConfig: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
				GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
			},
		},
	}))
	requireNoErrResp(s.T(), createResp, err)
}
