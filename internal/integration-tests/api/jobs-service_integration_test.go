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
	piidetect_table_activities "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/piidetect/workflows/table/activities"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetJobs_Empty() {
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Jobs().GetJobs(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetJobsRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)
	jobs := resp.Msg.GetJobs()
	require.Empty(s.T(), jobs)
}

func (s *IntegrationTestSuite) Test_CreateJob_Ok() {
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
	srcconn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "source", "test")
	destconn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "dest", "test2")

	s.MockTemporalForCreateJob("test-id")

	resp, err := s.OSSUnauthenticatedLicensedClients.Jobs().CreateJob(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: accountId,
		JobName:   "test",
		Mappings:  []*mgmtv1alpha1.JobMapping{},
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
					Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
						ConnectionId: srcconn.GetId(),
					},
				},
			},
		},
		Destinations: []*mgmtv1alpha1.CreateJobDestination{
			{
				ConnectionId: destconn.GetId(),
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{},
					},
				}},
		},
		InitiateJobRun: false,
	}))
	requireNoErrResp(s.T(), resp, err)
	require.NotNil(s.T(), resp.Msg.GetJob())
}

func (s *IntegrationTestSuite) Test_JobService_JobHooks() {
	t := s.T()
	ctx := s.ctx

	t.Run("OSS-unlicensed-unimplemented", func(t *testing.T) {
		client := s.OSSUnauthenticatedUnlicensedClients.Jobs()
		t.Run("GetJobHooks", func(t *testing.T) {
			resp, err := client.GetJobHooks(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobHooksRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("GetJobHook", func(t *testing.T) {
			resp, err := client.GetJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("CreateJobHook", func(t *testing.T) {
			resp, err := client.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("DeleteJobHook", func(t *testing.T) {
			resp, err := client.DeleteJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteJobHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("IsJobHookNameAvailable", func(t *testing.T) {
			resp, err := client.IsJobHookNameAvailable(ctx, connect.NewRequest(&mgmtv1alpha1.IsJobHookNameAvailableRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("UpdateJobHook", func(t *testing.T) {
			resp, err := client.UpdateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.UpdateJobHookRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
		t.Run("SetJobHookEnabled", func(t *testing.T) {
			resp, err := client.SetJobHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetJobHookEnabledRequest{}))
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeUnimplemented)
		})
	})

	t.Run("Cloud", func(t *testing.T) {
		client := s.NeosyncCloudAuthenticatedLicensedClients.Jobs(integrationtests_test.WithUserId(testAuthUserId))
		s.setUser(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)))
		accountId := s.createPersonalAccount(ctx, s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId)))

		srcconn := s.createPostgresConnection(s.NeosyncCloudAuthenticatedLicensedClients.Connections(integrationtests_test.WithUserId(testAuthUserId)), accountId, "source", "test")
		destconn := s.createPostgresConnection(s.NeosyncCloudAuthenticatedLicensedClients.Connections(integrationtests_test.WithUserId(testAuthUserId)), accountId, "dest", "test2")

		s.MockTemporalForCreateJob("test-id")
		jobResp, err := client.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
			JobName:   "cloud-testjob-1",
			AccountId: accountId,
			Mappings:  []*mgmtv1alpha1.JobMapping{},
			Source: &mgmtv1alpha1.JobSource{
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
						Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
							ConnectionId: srcconn.GetId(),
						},
					},
				},
			},
			Destinations: []*mgmtv1alpha1.CreateJobDestination{
				{
					ConnectionId: destconn.GetId(),
					Options: &mgmtv1alpha1.JobDestinationOptions{
						Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
							PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{},
						},
					},
				},
			},
		}))
		requireNoErrResp(t, jobResp, err)

		t.Run("GetJobHooks", func(t *testing.T) {
			createdHook := s.createSqlJobHook(ctx, t, client, "getjobhooks-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
			})

			t.Run("ok", func(t *testing.T) {
				resp, err := client.GetJobHooks(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobHooksRequest{
					JobId: createdHook.GetJobId(),
				}))
				requireNoErrResp(t, resp, err)
				hooks := resp.Msg.GetHooks()
				require.NotEmpty(t, hooks)
			})
		})

		t.Run("GetJobHook", func(t *testing.T) {
			createdHook := s.createSqlJobHook(ctx, t, client, "getjobhook-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
			})

			t.Run("ok", func(t *testing.T) {
				resp, err := client.GetJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobHookRequest{
					Id: createdHook.GetId(),
				}))
				requireNoErrResp(t, resp, err)
				hook := resp.Msg.GetHook()
				require.NotNil(t, hook)
				require.Equal(t, createdHook.GetId(), hook.GetId())
			})
			t.Run("not_found", func(t *testing.T) {
				resp, err := client.GetJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobHookRequest{
					Id: uuid.NewString(),
				}))
				requireErrResp(t, resp, err)
				requireConnectError(t, err, connect.CodeNotFound)
			})
		})

		t.Run("CreateJobHook", func(t *testing.T) {
			t.Run("ok", func(t *testing.T) {
				resp, err := client.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{
					JobId: jobResp.Msg.GetJob().GetId(),
					Hook: &mgmtv1alpha1.NewJobHook{
						Name:        "createjobhook-1",
						Description: "createjobhook ok",
						Enabled:     true,
						Priority:    0,
						Config: &mgmtv1alpha1.JobHookConfig{
							Config: &mgmtv1alpha1.JobHookConfig_Sql{
								Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
									Query:        "foo",
									ConnectionId: srcconn.GetId(),
									Timing:       &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{}},
								},
							},
						},
					},
				}))
				requireNoErrResp(t, resp, err)
			})

			t.Run("job_not_found", func(t *testing.T) {
				resp, err := client.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{
					JobId: uuid.NewString(), // job id does not exist
					Hook: &mgmtv1alpha1.NewJobHook{
						Name:        "createjobhook-2",
						Description: "createjobhook job not found",
						Enabled:     true,
						Priority:    0,
						Config: &mgmtv1alpha1.JobHookConfig{
							Config: &mgmtv1alpha1.JobHookConfig_Sql{
								Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
									Query:        "foo",
									ConnectionId: srcconn.GetId(),
									Timing:       &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{}},
								},
							},
						},
					},
				}))
				requireErrResp(t, resp, err)
				requireConnectError(t, err, connect.CodeNotFound)
			})
			t.Run("connection_not_in_job", func(t *testing.T) {
				resp, err := client.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{
					JobId: jobResp.Msg.GetJob().GetId(),
					Hook: &mgmtv1alpha1.NewJobHook{
						Name:        "createjobhook-3",
						Description: "createjobhook connection not in job",
						Enabled:     true,
						Priority:    0,
						Config: &mgmtv1alpha1.JobHookConfig{
							Config: &mgmtv1alpha1.JobHookConfig_Sql{
								Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
									Query:        "foo",
									ConnectionId: uuid.NewString(), // job does not have this connection id
									Timing:       &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{}},
								},
							},
						},
					},
				}))
				requireErrResp(t, resp, err)
				requireConnectError(t, err, connect.CodeInvalidArgument)
			})
			t.Run("invalid_timing", func(t *testing.T) {
				resp, err := client.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{
					JobId: jobResp.Msg.GetJob().GetId(),
					Hook: &mgmtv1alpha1.NewJobHook{
						Name:        "createjobhook-4",
						Description: "createjobhook bad timing",
						Enabled:     true,
						Priority:    0,
						Config: &mgmtv1alpha1.JobHookConfig{
							Config: &mgmtv1alpha1.JobHookConfig_Sql{
								Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
									Query:        "foo",
									ConnectionId: srcconn.GetId(),
									Timing:       &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{Timing: nil},
								},
							},
						},
					},
				}))
				requireErrResp(t, resp, err)
				requireConnectError(t, err, connect.CodeUnknown)
			})
		})

		t.Run("DeleteJobHook", func(t *testing.T) {
			t.Run("ok", func(t *testing.T) {
				createdHook := s.createSqlJobHook(ctx, t, client, "deletejobhook-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
					Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
				})
				resp, err := client.DeleteJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteJobHookRequest{Id: createdHook.GetId()}))
				requireNoErrResp(t, resp, err)

				getResp, err := client.GetJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobHookRequest{Id: createdHook.GetId()}))
				requireErrResp(t, getResp, err)
				requireConnectError(t, err, connect.CodeNotFound)
			})
			t.Run("non_existent", func(t *testing.T) {
				resp, err := client.DeleteJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteJobHookRequest{Id: uuid.NewString()}))
				requireNoErrResp(t, resp, err)
			})
		})

		t.Run("IsJobHookNameAvailable", func(t *testing.T) {
			t.Run("yes", func(t *testing.T) {
				resp, err := client.IsJobHookNameAvailable(ctx, connect.NewRequest(&mgmtv1alpha1.IsJobHookNameAvailableRequest{
					JobId: jobResp.Msg.GetJob().GetId(),
					Name:  "isjobhooknameavailable-1",
				}))
				requireNoErrResp(t, resp, err)
				require.True(t, resp.Msg.GetIsAvailable())
			})
			t.Run("no", func(t *testing.T) {
				createdHook := s.createSqlJobHook(ctx, t, client, "isjobhooknameavail-2", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
					Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
				})

				resp, err := client.IsJobHookNameAvailable(ctx, connect.NewRequest(&mgmtv1alpha1.IsJobHookNameAvailableRequest{
					JobId: jobResp.Msg.GetJob().GetId(),
					Name:  createdHook.GetName(),
				}))
				requireNoErrResp(t, resp, err)
				require.False(t, resp.Msg.GetIsAvailable())
			})
		})

		t.Run("SetJobHookEnabled", func(t *testing.T) {
			createdHook := s.createSqlJobHook(ctx, t, client, "setjobhookenabled-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
			})
			require.True(t, createdHook.GetEnabled())
			resp, err := client.SetJobHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetJobHookEnabledRequest{
				Id:      createdHook.GetId(),
				Enabled: false,
			}))
			requireNoErrResp(t, resp, err)
			require.False(t, resp.Msg.GetHook().GetEnabled())
			resp, err = client.SetJobHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetJobHookEnabledRequest{
				Id:      createdHook.GetId(),
				Enabled: true,
			}))
			requireNoErrResp(t, resp, err)
			require.True(t, resp.Msg.GetHook().GetEnabled())
		})

		t.Run("UpdateJobHook", func(t *testing.T) {
			createdHook := s.createSqlJobHook(ctx, t, client, "updatejobhook-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
			})
			resp, err := client.UpdateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.UpdateJobHookRequest{
				Id:          createdHook.GetId(),
				Name:        fmt.Sprintf("%s-updated", createdHook.GetName()),
				Description: fmt.Sprintf("%s-updated", createdHook.GetDescription()),
				Enabled:     !createdHook.GetEnabled(),
				Priority:    createdHook.GetPriority() + 1,
				Config: &mgmtv1alpha1.JobHookConfig{
					Config: &mgmtv1alpha1.JobHookConfig_Sql{
						Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
							Query:        "foobar",
							ConnectionId: destconn.GetId(),
							Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
								Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PostSync{},
							},
						},
					},
				},
			}))
			requireNoErrResp(t, resp, err)
			updatedHook := resp.Msg.GetHook()
			require.Equal(t, fmt.Sprintf("%s-updated", createdHook.GetName()), updatedHook.GetName())
			require.Equal(t, fmt.Sprintf("%s-updated", createdHook.GetDescription()), updatedHook.GetDescription())
			require.Equal(t, !createdHook.GetEnabled(), updatedHook.GetEnabled())
			require.Equal(t, createdHook.GetPriority()+1, updatedHook.GetPriority())
			sqlhook := updatedHook.GetConfig().GetSql()
			require.NotNil(t, sqlhook)
			require.Equal(t, "foobar", sqlhook.GetQuery())
			require.Equal(t, destconn.GetId(), sqlhook.GetConnectionId())
			require.NotNil(t, sqlhook.GetTiming().GetPostSync())
		})

		t.Run("GetActiveJobHooksByTiming", func(t *testing.T) {
			createdPreSyncHook := s.createSqlJobHook(ctx, t, client, "getactivejobhooksbytiming-pre", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
			})
			createdPostSyncHook := s.createSqlJobHook(ctx, t, client, "getactivejobhooksbytiming-post", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), true, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PostSync{},
			})
			disabledHook := s.createSqlJobHook(ctx, t, client, "getactivejobhooksbytiming-disabled", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), false, &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PostSync{},
			})
			t.Run("unspecified", func(t *testing.T) {
				resp, err := client.GetActiveJobHooksByTiming(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveJobHooksByTimingRequest{
					JobId:  jobResp.Msg.GetJob().GetId(),
					Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_UNSPECIFIED,
				}))
				requireNoErrResp(t, resp, err)
				require.NotEmpty(t, resp.Msg.GetHooks())
				hasDisabledHook := false
				for _, hook := range resp.Msg.GetHooks() {
					if hook.GetId() == disabledHook.GetId() {
						hasDisabledHook = true
						break
					}
				}
				require.False(t, hasDisabledHook, "GetActiveHooksByTiming should never return disabled hooks!")
			})

			t.Run("presync", func(t *testing.T) {
				resp, err := client.GetActiveJobHooksByTiming(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveJobHooksByTimingRequest{
					JobId:  jobResp.Msg.GetJob().GetId(),
					Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC,
				}))
				requireNoErrResp(t, resp, err)
				require.NotEmpty(t, resp.Msg.GetHooks())
				hasCreatedHook := false
				for _, hook := range resp.Msg.GetHooks() {
					if hook.GetId() == createdPreSyncHook.GetId() {
						hasCreatedHook = true
						break
					}
				}
				require.True(t, hasCreatedHook)
			})

			t.Run("postsync", func(t *testing.T) {
				resp, err := client.GetActiveJobHooksByTiming(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveJobHooksByTimingRequest{
					JobId:  jobResp.Msg.GetJob().GetId(),
					Timing: mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_POSTSYNC,
				}))
				requireNoErrResp(t, resp, err)
				require.NotEmpty(t, resp.Msg.GetHooks())
				hasCreatedHook := false
				for _, hook := range resp.Msg.GetHooks() {
					if hook.GetId() == createdPostSyncHook.GetId() {
						hasCreatedHook = true
						break
					}
				}
				require.True(t, hasCreatedHook)
			})
		})
	})
}

func (s *IntegrationTestSuite) createSqlJobHook(
	ctx context.Context,
	t testing.TB,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	name string,
	jobId string,
	connectionId string,
	enabled bool,
	timing *mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing,
) *mgmtv1alpha1.JobHook {
	createResp, err := jobclient.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{
		JobId: jobId,
		Hook: &mgmtv1alpha1.NewJobHook{
			Name:        name,
			Description: "sql job hook test",
			Enabled:     enabled,
			Priority:    100,
			Config: &mgmtv1alpha1.JobHookConfig{
				Config: &mgmtv1alpha1.JobHookConfig_Sql{
					Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
						Query:        "truncate table public.users;",
						ConnectionId: connectionId,
						Timing:       timing,
					},
				},
			},
		},
	}))
	requireNoErrResp(t, createResp, err)
	return createResp.Msg.GetHook()
}

func (s *IntegrationTestSuite) createPostgresConnection(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId string,
	name string,
	pgurl string,
) *mgmtv1alpha1.Connection {
	resp, err := connclient.CreateConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
			AccountId: accountId,
			Name:      name,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: pgurl,
						},
					},
				},
			},
		}),
	)
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.GetConnection()
}

func (s *IntegrationTestSuite) Test_ValidateSchema() {
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
	destconn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "dest", s.Pgcontainer.URL)

	s.T().Run("MissingTables", func(t *testing.T) {
		Mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "public",
				Table:  "addresses",
				Column: "order_id",
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					},
				},
			},
			{
				Schema: "public",
				Table:  "customers",
				Column: "id",
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					},
				},
			},
		}

		resp, err := s.OSSUnauthenticatedLicensedClients.Jobs().ValidateSchema(s.ctx, connect.NewRequest(&mgmtv1alpha1.ValidateSchemaRequest{
			ConnectionId: destconn.GetId(),
			Mappings:     Mappings,
		}))
		requireNoErrResp(t, resp, err)
		require.Len(t, resp.Msg.MissingTables, 2)
	})

	s.T().Run("Ok", func(t *testing.T) {
		Mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "neosync_api",
				Table:  "users",
				Column: "id",
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					},
				},
			},
		}

		resp, err := s.OSSUnauthenticatedLicensedClients.Jobs().ValidateSchema(s.ctx, connect.NewRequest(&mgmtv1alpha1.ValidateSchemaRequest{
			ConnectionId: destconn.GetId(),
			Mappings:     Mappings,
		}))
		requireNoErrResp(t, resp, err)
		require.Len(t, resp.Msg.MissingTables, 0)
	})
}

func (s *IntegrationTestSuite) Test_GetPiiDetectionReport() {
	userclient := s.OSSUnauthenticatedLicensedClients.Users()
	s.setUser(s.ctx, userclient)
	accountId := s.createPersonalAccount(s.ctx, userclient)

	connclient := s.OSSUnauthenticatedLicensedClients.Connections()
	jobclient := s.OSSUnauthenticatedLicensedClients.Jobs()

	srcconn := s.createPostgresConnection(connclient, accountId, "pii-detect-src", "postgres://postgres:postgres@localhost:5432/postgres")

	s.MockTemporalForCreateJob("test-id")
	jobResp, err := jobclient.CreateJob(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: accountId,
		JobName:   "pii-detection-test" + uuid.NewString(),
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
					Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
						ConnectionId: srcconn.GetId(),
					},
				},
			},
		},
		JobType: &mgmtv1alpha1.JobTypeConfig{
			JobType: &mgmtv1alpha1.JobTypeConfig_PiiDetect{
				PiiDetect: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect{
					DataSampling: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_DataSampling{
						IsEnabled: true,
					},
					TableScanFilter: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter{Mode: &mgmtv1alpha1.JobTypeConfig_JobTypePiiDetect_TableScanFilter_IncludeAll{}},
				},
			},
		},
	}))
	requireNoErrResp(s.T(), jobResp, err)

	jobId := jobResp.Msg.GetJob().GetId()

	s.T().Run("found", func(t *testing.T) {
		jobRunId := fmt.Sprintf("%s-%s", jobId, time.Now().Format(time.RFC3339))

		report := piidetect_table_activities.TableReport{
			TableSchema: "public",
			TableName:   "users",
			ColumnReports: []piidetect_table_activities.ColumnReport{
				{
					ColumnName: "age",
					Report: piidetect_table_activities.CombinedPiiDetectReport{
						Regex: &piidetect_table_activities.RegexPiiDetectReport{
							Category: piidetect_table_activities.PiiCategoryPersonal,
						},
					},
				},
			},
		}
		reportBytes, err := json.Marshal(report)
		require.NoError(s.T(), err)

		setResp, err := jobclient.SetRunContext(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetRunContextRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				AccountId:  accountId,
				JobRunId:   jobRunId,
				ExternalId: piidetect_table_activities.BuildTableReportExternalId("public", "users"),
			},
			Value: reportBytes,
		}))
		requireNoErrResp(t, setResp, err)

		s.MockTemporalForDescribeWorkflowExecution(accountId, jobId, jobRunId, "JobPiiDetect")

		getResp, err := jobclient.GetPiiDetectionReport(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetPiiDetectionReportRequest{
			JobRunId:  jobRunId,
			AccountId: accountId,
		}))
		requireNoErrResp(t, getResp, err)
		require.NotNil(t, getResp.Msg.GetReport())
		tables := getResp.Msg.GetReport().GetTables()
		require.Len(t, tables, 1)

		table := tables[0]
		require.Equal(t, "public", table.Schema)
		require.Equal(t, "users", table.Table)
		require.Len(t, table.Columns, 1)

		columnReport := table.Columns[0]
		require.Equal(t, "age", columnReport.Column)
		require.Equal(t, piidetect_table_activities.PiiCategoryPersonal.String(), columnReport.RegexReport.Category)
	})

	s.T().Run("empty", func(t *testing.T) {
		jobRunId := fmt.Sprintf("%s-%s-empty", jobId, time.Now().Format(time.RFC3339))
		s.MockTemporalForDescribeWorkflowExecution(accountId, jobId, jobRunId, "JobPiiDetect")

		getResp, err := jobclient.GetPiiDetectionReport(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetPiiDetectionReportRequest{
			JobRunId:  jobRunId,
			AccountId: accountId,
		}))
		requireNoErrResp(t, getResp, err)
		require.Empty(t, getResp.Msg.GetReport().GetTables())
	})
}
