package integrationtests_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_AccountHooksService_GetActiveAccountHooksByEvent() {
	t := s.T()
	ctx := s.ctx

	t.Run("OSS-unlicensed-unimplemented", func(t *testing.T) {
		client := s.OSSUnauthenticatedUnlicensedClients.AccountHooks()
		t.Run("GetAccountHooks", func(t *testing.T) {
			resp, err := client.GetAccountHooks(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHooksRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("GetAccountHook", func(t *testing.T) {
			resp, err := client.GetAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("CreateAccountHook", func(t *testing.T) {
			resp, err := client.CreateAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateAccountHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("DeleteAccountHook", func(t *testing.T) {
			resp, err := client.DeleteAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteAccountHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("IsAccountHookNameAvailable", func(t *testing.T) {
			resp, err := client.IsAccountHookNameAvailable(ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountHookNameAvailableRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("UpdateAccountHook", func(t *testing.T) {
			resp, err := client.UpdateAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.UpdateAccountHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("SetAccountHookEnabled", func(t *testing.T) {
			resp, err := client.SetAccountHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountHookEnabledRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("GetActiveAccountHooksByEvent", func(t *testing.T) {
			resp, err := client.GetActiveAccountHooksByEvent(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveAccountHooksByEventRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
	})

	t.Run("Cloud", func(t *testing.T) {
		client := s.NeosyncCloudAuthenticatedLicensedClients.AccountHooks(integrationtests_test.WithUserId(testAuthUserId))
		s.setUser(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)))

		t.Run("GetAccountHooks", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

			resp, err := client.GetAccountHooks(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHooksRequest{AccountId: accountId}))
			requireNoErrResp(t, resp, err)
			require.ElementsMatch(t, []*mgmtv1alpha1.AccountHook{createdHook}, resp.Msg.Hooks)
		})

		t.Run("GetAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

			resp, err := client.GetAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHookRequest{Id: createdHook.Id}))
			requireNoErrResp(t, resp, err)
			require.Equal(t, createdHook, resp.Msg.Hook)
		})

		t.Run("CreateAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)
		})

		t.Run("DeleteAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())

			t.Run("ok", func(t *testing.T) {
				createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

				resp, err := client.DeleteAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteAccountHookRequest{Id: createdHook.Id}))
				requireNoErrResp(t, resp, err)

				getResp, err := client.GetAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHookRequest{Id: createdHook.Id}))
				requireErrResp(t, getResp, err)
				requireConnectError(t, err, connect.CodeNotFound)
			})

			t.Run("non_existent", func(t *testing.T) {
				resp, err := client.DeleteAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteAccountHookRequest{Id: uuid.NewString()}))
				requireNoErrResp(t, resp, err)
			})
		})

		t.Run("IsAccountHookNameAvailable", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())

			t.Run("yes", func(t *testing.T) {
				resp, err := client.IsAccountHookNameAvailable(ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountHookNameAvailableRequest{
					AccountId: accountId,
					Name:      "test-hook",
				}))
				requireNoErrResp(t, resp, err)
				require.True(t, resp.Msg.IsAvailable)
			})
			t.Run("no", func(t *testing.T) {
				createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

				resp, err := client.IsAccountHookNameAvailable(ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountHookNameAvailableRequest{
					AccountId: accountId,
					Name:      createdHook.Name,
				}))
				requireNoErrResp(t, resp, err)
				require.False(t, resp.Msg.IsAvailable)
			})
		})

		t.Run("SetAccountHookEnabled", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())

			t.Run("ok", func(t *testing.T) {
				createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

				resp, err := client.SetAccountHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountHookEnabledRequest{
					Id:      createdHook.Id,
					Enabled: false,
				}))
				requireNoErrResp(t, resp, err)
				require.False(t, resp.Msg.GetHook().GetEnabled())

				resp, err = client.SetAccountHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetAccountHookEnabledRequest{
					Id:      createdHook.Id,
					Enabled: true,
				}))
				requireNoErrResp(t, resp, err)
				require.True(t, resp.Msg.GetHook().GetEnabled())
			})
		})

		t.Run("GetActiveAccountHooksByEvent", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)
			createdHook2 := s.createAccountHook(ctx, t, client, accountId, "test-hook-2", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED, mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED}, true)
			s.createAccountHook(ctx, t, client, accountId, "test-hook-3", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED}, false)
			createdHook4 := s.createAccountHook(ctx, t, client, accountId, "test-hook-4", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED}, true)

			t.Run("ok", func(t *testing.T) {

				resp, err := client.GetActiveAccountHooksByEvent(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveAccountHooksByEventRequest{
					AccountId: accountId,
					Event:     mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED,
				}))
				requireNoErrResp(t, resp, err)
				require.ElementsMatch(t, []*mgmtv1alpha1.AccountHook{createdHook, createdHook2, createdHook4}, resp.Msg.GetHooks())
			})

			t.Run("wildcard", func(t *testing.T) {
				resp, err := client.GetActiveAccountHooksByEvent(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveAccountHooksByEventRequest{
					AccountId: accountId,
					Event:     mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED,
				}))
				requireNoErrResp(t, resp, err)
				require.Len(t, resp.Msg.GetHooks(), 2)
				require.ElementsMatch(t, []*mgmtv1alpha1.AccountHook{createdHook2, createdHook4}, resp.Msg.GetHooks())
			})
		})

		t.Run("UpdateAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			createdHook := s.createAccountHook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

			t.Run("ok", func(t *testing.T) {
				resp, err := client.UpdateAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.UpdateAccountHookRequest{
					Id:          createdHook.Id,
					Name:        "test-hook-updated",
					Description: "updated hook",
					Events:      []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED},
					Enabled:     true,
					Config: &mgmtv1alpha1.AccountHookConfig{
						Config: &mgmtv1alpha1.AccountHookConfig_Webhook{
							Webhook: &mgmtv1alpha1.AccountHookConfig_WebHook{
								Url:                    "https://example2.com",
								Secret:                 "foo-updated",
								DisableSslVerification: true,
							},
						},
					},
				}))
				requireNoErrResp(t, resp, err)
				updatedHook := resp.Msg.GetHook()
				require.Equal(t, "test-hook-updated", updatedHook.GetName())
				require.Equal(t, "updated hook", updatedHook.GetDescription())
				require.ElementsMatch(t, []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED}, updatedHook.GetEvents())
				require.True(t, updatedHook.GetEnabled())
				require.Equal(t, "https://example2.com", updatedHook.GetConfig().GetWebhook().GetUrl())
				require.Equal(t, "foo-updated", updatedHook.GetConfig().GetWebhook().GetSecret())
				require.True(t, updatedHook.GetConfig().GetWebhook().GetDisableSslVerification())
			})
		})
	})
}

func (s *IntegrationTestSuite) createAccountHook(
	ctx context.Context,
	t testing.TB,
	client mgmtv1alpha1connect.AccountHookServiceClient,
	accountId string,
	name string,
	events []mgmtv1alpha1.AccountHookEvent,
	enabled bool,

) *mgmtv1alpha1.AccountHook {
	createResp, err := client.CreateAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateAccountHookRequest{
		AccountId: accountId,
		Hook: &mgmtv1alpha1.NewAccountHook{
			Name:        name,
			Description: "created hook",
			Events:      events,
			Enabled:     enabled,
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Webhook{
					Webhook: &mgmtv1alpha1.AccountHookConfig_WebHook{
						Url:                    "https://example.com",
						Secret:                 "foo",
						DisableSslVerification: false,
					},
				},
			},
		},
	},
	),
	)
	requireNoErrResp(t, createResp, err)
	return createResp.Msg.GetHook()
}
