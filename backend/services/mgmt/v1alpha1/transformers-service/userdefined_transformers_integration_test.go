package v1alpha1_transformersservice

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
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

func (s *IntegrationTestSuite) Test_DeleteUserDefinedTransformer_Empty() {
	resp, err := s.transformerclient.DeleteUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.DeleteUserDefinedTransformerRequest{
		TransformerId: uuid.NewString(),
	}))
	requireNoErrResp(s.T(), resp, err)
}

func (s *IntegrationTestSuite) Test_DeleteUserDefinedTransformer_Ok() {
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

	resp, err := s.transformerclient.DeleteUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.DeleteUserDefinedTransformerRequest{
		TransformerId: createResp.Msg.GetTransformer().GetId(),
	}))
	requireNoErrResp(s.T(), resp, err)

	getResp, err := s.transformerclient.GetUserDefinedTransformerById(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
		TransformerId: createResp.Msg.GetTransformer().GetId(),
	}))
	requireErrResp(s.T(), getResp, err)
	requireConnectError(s.T(), err, connect.CodeNotFound)
}

func (s *IntegrationTestSuite) Test_UpdateUserDefinedTransformer_NotFound() {
	updatedResp, err := s.transformerclient.UpdateUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.UpdateUserDefinedTransformerRequest{
		TransformerId: uuid.NewString(),
		Name:          "my-test-transformer2",
		Description:   "test description",
	}))
	requireErrResp(s.T(), updatedResp, err)
	requireConnectError(s.T(), err, connect.CodeNotFound)
}

func (s *IntegrationTestSuite) Test_UpdateUserDefinedTransformer_Ok() {
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
	transId := createResp.Msg.GetTransformer().GetId()

	updatedResp, err := s.transformerclient.UpdateUserDefinedTransformer(s.ctx, connect.NewRequest(&mgmtv1alpha1.UpdateUserDefinedTransformerRequest{
		TransformerId: transId,
		Name:          "my-test-transformer2",
		Description:   "test description",
	}))
	requireNoErrResp(s.T(), updatedResp, err)
	require.Equal(s.T(), "my-test-transformer2", updatedResp.Msg.GetTransformer().GetName())
	require.Equal(s.T(), "test description", updatedResp.Msg.GetTransformer().GetDescription())
}

func (s *IntegrationTestSuite) Test_IsTransformerNameAvailable_Yes() {
	accountId := s.createPersonalAccount(s.userclient)

	resp, err := s.transformerclient.IsTransformerNameAvailable(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsTransformerNameAvailableRequest{
		AccountId:       accountId,
		TransformerName: "foo",
	}))
	requireNoErrResp(s.T(), resp, err)
	require.True(s.T(), resp.Msg.GetIsAvailable())
}

func (s *IntegrationTestSuite) Test_IsTransformerNameAvailable_No() {
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

	resp, err := s.transformerclient.IsTransformerNameAvailable(s.ctx, connect.NewRequest(&mgmtv1alpha1.IsTransformerNameAvailableRequest{
		AccountId:       accountId,
		TransformerName: createResp.Msg.GetTransformer().GetName(),
	}))
	requireNoErrResp(s.T(), resp, err)
	require.False(s.T(), resp.Msg.GetIsAvailable())
}

func (s *IntegrationTestSuite) Test_ValidateUserJavascriptCode() {
	errgrp, errctx := errgroup.WithContext(s.ctx)

	errgrp.Go(func() error {
		resp, err := s.transformerclient.ValidateUserJavascriptCode(errctx, connect.NewRequest(&mgmtv1alpha1.ValidateUserJavascriptCodeRequest{
			Code: `return "hello world";`,
		}))
		assertNoErrResp(s.T(), resp, err)
		assert.True(s.T(), resp.Msg.GetValid())
		return nil
	})

	errgrp.Go(func() error {
		resp, err := s.transformerclient.ValidateUserJavascriptCode(errctx, connect.NewRequest(&mgmtv1alpha1.ValidateUserJavascriptCodeRequest{
			Code: `return "hello world`,
		}))
		assertNoErrResp(s.T(), resp, err)
		assert.False(s.T(), resp.Msg.GetValid())
		return nil
	})

	err := errgrp.Wait()
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) Test_ValidateUserRegexCode() {
	errgrp, errctx := errgroup.WithContext(s.ctx)

	errgrp.Go(func() error {
		resp, err := s.transformerclient.ValidateUserRegexCode(errctx, connect.NewRequest(&mgmtv1alpha1.ValidateUserRegexCodeRequest{
			UserProvidedRegex: `\s`,
		}))
		assertNoErrResp(s.T(), resp, err)
		assert.True(s.T(), resp.Msg.GetValid())
		return nil
	})

	errgrp.Go(func() error {
		resp, err := s.transformerclient.ValidateUserRegexCode(errctx, connect.NewRequest(&mgmtv1alpha1.ValidateUserRegexCodeRequest{
			UserProvidedRegex: `(abc**+[0-9]`,
		}))
		assertNoErrResp(s.T(), resp, err)
		assert.False(s.T(), resp.Msg.GetValid())
		return nil
	})

	err := errgrp.Wait()
	require.NoError(s.T(), err)
}
