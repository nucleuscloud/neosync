package integrationtests_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	accounthook_events "github.com/nucleuscloud/neosync/internal/ee/events"
	ee_slack "github.com/nucleuscloud/neosync/internal/ee/slack"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/mock"
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
			createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

			resp, err := client.GetAccountHooks(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHooksRequest{AccountId: accountId}))
			requireNoErrResp(t, resp, err)
			require.ElementsMatch(t, []*mgmtv1alpha1.AccountHook{createdHook}, resp.Msg.Hooks)
		})

		t.Run("GetAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

			resp, err := client.GetAccountHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetAccountHookRequest{Id: createdHook.Id}))
			requireNoErrResp(t, resp, err)
			require.Equal(t, createdHook, resp.Msg.Hook)
		})

		t.Run("CreateAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())
			s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)
		})

		t.Run("DeleteAccountHook", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)), uuid.NewString(), uuid.NewString())

			t.Run("ok", func(t *testing.T) {
				createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

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
				createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

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
				createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

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
			createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)
			createdHook2 := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook-2", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED, mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED}, true)
			s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook-3", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED}, false)
			createdHook4 := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook-4", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED}, true)

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
			createdHook := s.createAccountHook_Webhook(ctx, t, client, accountId, "test-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED}, true)

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

func (s *IntegrationTestSuite) Test_AccountHooksService_Slack() {
	t := s.T()
	ctx := s.ctx

	userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
	userId := s.setUser(ctx, userclient)
	accountId := s.createBilledTeamAccount(ctx, userclient, uuid.NewString(), uuid.NewString())

	hookclient := s.NeosyncCloudAuthenticatedLicensedClients.AccountHooks(integrationtests_test.WithUserId(testAuthUserId))

	t.Run("GetSlackConnectionUrl", func(t *testing.T) {
		mockedurl := "https://example.com"
		s.Mocks.Slackclient.EXPECT().GetAuthorizeUrl(accountId, userId).Return(mockedurl, nil)
		resp, err := hookclient.GetSlackConnectionUrl(ctx, connect.NewRequest(&mgmtv1alpha1.GetSlackConnectionUrlRequest{
			AccountId: accountId,
		}))
		requireNoErrResp(t, resp, err)
		require.Equal(t, mockedurl, resp.Msg.GetUrl())
	})

	t.Run("HandleSlackOAuthCallback", func(t *testing.T) {
		s.Mocks.Slackclient.EXPECT().ValidateState(mock.Anything, mock.Anything, userId, mock.Anything).Return(&ee_slack.OauthState{
			AccountId: accountId,
			UserId:    userId,
			Timestamp: time.Now().UTC().Unix(),
		}, nil)
		s.Mocks.Slackclient.EXPECT().ExchangeCodeForAccessToken(mock.Anything, mock.Anything).Return(&slack.OAuthV2Response{
			AccessToken: "access_token",
		}, nil)
		resp, err := hookclient.HandleSlackOAuthCallback(ctx, connect.NewRequest(&mgmtv1alpha1.HandleSlackOAuthCallbackRequest{
			State: "state",
			Code:  "code",
		}))
		requireNoErrResp(t, resp, err)
	})

	t.Run("TestSlackConnection", func(t *testing.T) {
		t.Run("is configured", func(t *testing.T) {
			s.Mocks.Slackclient.EXPECT().ValidateState(mock.Anything, mock.Anything, userId, mock.Anything).Return(&ee_slack.OauthState{
				AccountId: accountId,
				UserId:    userId,
				Timestamp: time.Now().UTC().Unix(),
			}, nil)
			s.Mocks.Slackclient.EXPECT().ExchangeCodeForAccessToken(mock.Anything, mock.Anything).Return(&slack.OAuthV2Response{
				AccessToken: "access_token",
			}, nil)
			resp, err := hookclient.HandleSlackOAuthCallback(ctx, connect.NewRequest(&mgmtv1alpha1.HandleSlackOAuthCallbackRequest{
				State: "state",
				Code:  "code",
			}))
			requireNoErrResp(t, resp, err)

			t.Run("slack test fails", func(t *testing.T) {
				s.Mocks.Slackclient.EXPECT().Test(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("slack test failed")).Once()

				resp, err := hookclient.TestSlackConnection(ctx, connect.NewRequest(&mgmtv1alpha1.TestSlackConnectionRequest{
					AccountId: accountId,
				}))
				requireNoErrResp(t, resp, err)
				require.Equal(t, "slack test failed", resp.Msg.GetError())
				require.True(t, resp.Msg.GetHasConfiguration())
				require.Nil(t, resp.Msg.GetTestResponse())
			})

			t.Run("slack test succeeds", func(t *testing.T) {
				s.Mocks.Slackclient.EXPECT().Test(mock.Anything, mock.Anything).Return(&slack.AuthTestResponse{
					URL:  "https://example.com",
					Team: "team-id",
				}, nil).Once()

				resp, err := hookclient.TestSlackConnection(ctx, connect.NewRequest(&mgmtv1alpha1.TestSlackConnectionRequest{
					AccountId: accountId,
				}))
				requireNoErrResp(t, resp, err)
				t.Log(resp.Msg.GetTestResponse())
				require.Equal(t, "https://example.com", resp.Msg.GetTestResponse().GetUrl())
				require.Equal(t, "team-id", resp.Msg.GetTestResponse().GetTeam())
				require.True(t, resp.Msg.GetHasConfiguration())
				require.Empty(t, resp.Msg.GetError())
			})
		})

		t.Run("is not configured", func(t *testing.T) {
			accountId := s.createBilledTeamAccount(ctx, userclient, uuid.NewString(), uuid.NewString())

			t.Run("no slack configuration record", func(t *testing.T) {
				resp, err := hookclient.TestSlackConnection(ctx, connect.NewRequest(&mgmtv1alpha1.TestSlackConnectionRequest{
					AccountId: accountId,
				}))
				requireNoErrResp(t, resp, err)
				require.Equal(t, "slack oauth connection not found", resp.Msg.GetError())
				require.False(t, resp.Msg.GetHasConfiguration())
				require.Nil(t, resp.Msg.GetTestResponse())
			})
		})
	})

	t.Run("SendSlackMessage", func(t *testing.T) {
		hook := s.createAccountHook_Slack(ctx, t, hookclient, accountId, "sendslackmessage-hook", []mgmtv1alpha1.AccountHookEvent{mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_UNSPECIFIED}, true)
		t.Run("job run created", func(t *testing.T) {
			s.Mocks.Slackclient.EXPECT().SendMessage(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			event := accounthook_events.NewEvent_JobRunCreated(accountId, uuid.NewString(), uuid.NewString())
			eventbits, err := json.Marshal(event)

			require.NoError(t, err)
			resp, err := hookclient.SendSlackMessage(ctx, connect.NewRequest(&mgmtv1alpha1.SendSlackMessageRequest{
				AccountHookId: hook.Id,
				Event:         eventbits,
			}))
			requireNoErrResp(t, resp, err)
		})
		t.Run("job run succeeded", func(t *testing.T) {
			s.Mocks.Slackclient.EXPECT().SendMessage(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			event := accounthook_events.NewEvent_JobRunSucceeded(accountId, uuid.NewString(), uuid.NewString())
			eventbits, err := json.Marshal(event)

			require.NoError(t, err)
			resp, err := hookclient.SendSlackMessage(ctx, connect.NewRequest(&mgmtv1alpha1.SendSlackMessageRequest{
				AccountHookId: hook.Id,
				Event:         eventbits,
			}))
			requireNoErrResp(t, resp, err)
		})
		t.Run("job run failed", func(t *testing.T) {
			s.Mocks.Slackclient.EXPECT().SendMessage(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			event := accounthook_events.NewEvent_JobRunFailed(accountId, uuid.NewString(), uuid.NewString())
			eventbits, err := json.Marshal(event)

			require.NoError(t, err)
			resp, err := hookclient.SendSlackMessage(ctx, connect.NewRequest(&mgmtv1alpha1.SendSlackMessageRequest{
				AccountHookId: hook.Id,
				Event:         eventbits,
			}))
			requireNoErrResp(t, resp, err)
		})
	})
}

func (s *IntegrationTestSuite) createAccountHook_Webhook(
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
	}))
	requireNoErrResp(t, createResp, err)
	return createResp.Msg.GetHook()
}

func (s *IntegrationTestSuite) createAccountHook_Slack(
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
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						Channel: "channel-id",
					},
				},
			},
		},
	}))
	requireNoErrResp(t, createResp, err)
	return createResp.Msg.GetHook()
}
