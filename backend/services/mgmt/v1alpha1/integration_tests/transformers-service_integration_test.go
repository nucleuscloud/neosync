package integrationtests_test

import (
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_TransformersService_GetSystemTransformers() {
	resp, err := s.OSSUnauthenticatedLicensedClients.Transformers().GetSystemTransformers(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformersRequest{}))
	integrationtests_test.RequireNoErrResp(s.T(), resp, err)
	require.NotEmpty(s.T(), resp.Msg.GetTransformers())
}

func (s *IntegrationTestSuite) Test_TransformersService_GetSystemTransformersBySource() {
	t := s.T()
	t.Run("ok", func(t *testing.T) {
		resp, err := s.OSSUnauthenticatedLicensedClients.Transformers().GetSystemTransformerBySource(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		transformer := resp.Msg.GetTransformer()
		require.NotNil(t, transformer)
		require.Equal(t, transformer.GetSource(), mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL)
	})
	t.Run("not_found", func(t *testing.T) {
		resp, err := s.OSSUnauthenticatedLicensedClients.Transformers().GetSystemTransformerBySource(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
		}))
		integrationtests_test.RequireErrResp(t, resp, err)
		integrationtests_test.RequireConnectError(t, err, connect.CodeNotFound)
	})
}

func (s *IntegrationTestSuite) Test_TransformersService_GetTransformPiiRecognizers() {
	t := s.T()

	t.Run("ok", func(t *testing.T) {
		allowed := []string{"foo", "bar"}
		s.Mocks.Presidio.Entities.On("GetSupportedentitiesWithResponse", mock.Anything, mock.Anything).
			Once().
			Return(&presidioapi.GetSupportedentitiesResponse{
				JSON200: &allowed,
			}, nil)

		accountId := integrationtests_test.CreatePersonalAccount(s.ctx, t, s.OSSUnauthenticatedLicensedClients.Users())
		resp, err := s.OSSUnauthenticatedLicensedClients.Transformers().GetTransformPiiEntities(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetTransformPiiEntitiesRequest{
			AccountId: accountId,
		}))
		integrationtests_test.RequireNoErrResp(t, resp, err)
		recognizers := resp.Msg.GetEntities()
		require.Equal(t, allowed, recognizers)
	})
}
