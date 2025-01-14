package integrationtests_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Services(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()
	api, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t)
	if err != nil {
		t.Fatalf("unable to create neosync api test client: %v", err)
	}
	_ = api

	t.Run("user-account-service", func(t *testing.T) {
		t.Run("GetAccountOnboardingConfig", func(t *testing.T) {
			t.Run("ok", func(t *testing.T) {
				accountId := integrationtests_test.CreatePersonalAccount(ctx, t, api.OSSUnauthenticatedLicensedClients.Users())

				resp, err := api.OSSUnauthenticatedLicensedClients.Users().GetAccountOnboardingConfig(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: accountId}))
				integrationtests_test.RequireNoErrResp(t, resp, err)
				onboardingConfig := resp.Msg.GetConfig()
				require.NotNil(t, onboardingConfig)

				require.False(t, onboardingConfig.GetHasCompletedOnboarding())
			})

			t.Run("no-account", func(t *testing.T) {
				resp, err := api.OSSUnauthenticatedLicensedClients.Users().GetAccountOnboardingConfig(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountOnboardingConfigRequest{AccountId: uuid.NewString()}))
				integrationtests_test.RequireErrResp(t, resp, err)
				integrationtests_test.RequireConnectError(t, err, connect.CodePermissionDenied)
			})
		})

		t.Run("SetAccountOnboardingConfig", func(t *testing.T) {
			t.Run("ok", func(t *testing.T) {
				accountId := integrationtests_test.CreatePersonalAccount(ctx, t, api.OSSUnauthenticatedLicensedClients.Users())

				resp, err := api.OSSUnauthenticatedLicensedClients.Users().SetAccountOnboardingConfig(ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{
					AccountId: accountId, Config: &mgmtv1alpha1.AccountOnboardingConfig{
						HasCompletedOnboarding: true,
					}},
				))
				integrationtests_test.RequireNoErrResp(t, resp, err)

				onboardingConfig := resp.Msg.GetConfig()
				require.NotNil(t, onboardingConfig)

				require.True(t, onboardingConfig.GetHasCompletedOnboarding())
			})
			t.Run("no-account", func(t *testing.T) {
				resp, err := api.OSSUnauthenticatedLicensedClients.Users().SetAccountOnboardingConfig(ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: uuid.NewString(), Config: &mgmtv1alpha1.AccountOnboardingConfig{}}))
				integrationtests_test.RequireErrResp(t, resp, err)
				integrationtests_test.RequireConnectError(t, err, connect.CodePermissionDenied)
			})
			t.Run("no-config", func(t *testing.T) {
				accountId := integrationtests_test.CreatePersonalAccount(ctx, t, api.OSSUnauthenticatedLicensedClients.Users())

				resp, err := api.OSSUnauthenticatedLicensedClients.Users().SetAccountOnboardingConfig(ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountOnboardingConfigRequest{AccountId: accountId, Config: nil}))
				integrationtests_test.RequireNoErrResp(t, resp, err)
				onboardingConfig := resp.Msg.GetConfig()
				require.NotNil(t, onboardingConfig)

				require.False(t, onboardingConfig.GetHasCompletedOnboarding())
			})
		})

		t.Run("GetAccountTemporalConfig", func(t *testing.T) {
			t.Run("ok", func(t *testing.T) {
				accountId := integrationtests_test.CreatePersonalAccount(ctx, t, api.OSSUnauthenticatedLicensedClients.Users())

				api.Mocks.TemporalConfigProvider.On("GetConfig", mock.Anything, mock.Anything).
					Return(validTemporalConfig, nil)

				resp, err := api.OSSUnauthenticatedLicensedClients.Users().GetAccountTemporalConfig(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
				integrationtests_test.RequireNoErrResp(t, resp, err)

				tc := resp.Msg.GetConfig()
				require.NotNil(t, tc)

				require.Equal(t, validTemporalConfig.Namespace, tc.GetNamespace())
				require.Equal(t, validTemporalConfig.SyncJobQueueName, tc.GetSyncJobQueueName())
				require.Equal(t, validTemporalConfig.Url, tc.GetUrl())
			})

			t.Run("no-account", func(t *testing.T) {
				resp, err := api.OSSUnauthenticatedLicensedClients.Users().GetAccountTemporalConfig(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: uuid.NewString()}))
				integrationtests_test.RequireErrResp(t, resp, err)
				integrationtests_test.RequireConnectError(t, err, connect.CodePermissionDenied)
			})

			t.Run("neosync-cloud", func(t *testing.T) {
				userclient := api.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
				accountId := integrationtests_test.CreatePersonalAccount(ctx, t, userclient)

				resp, err := userclient.GetAccountTemporalConfig(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountTemporalConfigRequest{AccountId: accountId}))
				integrationtests_test.RequireErrResp(t, resp, err)
				integrationtests_test.RequireConnectError(t, err, connect.CodeUnimplemented)
			})
		})

		t.Run("SetAccountTemporalConfig", func(t *testing.T) {

		})
	})

}
