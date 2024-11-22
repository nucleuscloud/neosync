package integrationtests_test

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_GetJobs_Empty() {
	accountId := s.createPersonalAccount(s.ctx, s.UnauthdClients.Users)

	resp, err := s.UnauthdClients.Jobs.GetJobs(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetJobsRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)
	jobs := resp.Msg.GetJobs()
	require.Empty(s.T(), jobs)
}

func (s *IntegrationTestSuite) Test_CreateJob_Ok() {
	accountId := s.createPersonalAccount(s.ctx, s.UnauthdClients.Users)
	srcconn := s.createPostgresConnection(s.UnauthdClients.Connections, accountId, "source", "test")
	destconn := s.createPostgresConnection(s.UnauthdClients.Connections, accountId, "dest", "test2")

	s.mockTemporalForCreateJob("test-id")

	resp, err := s.UnauthdClients.Jobs.CreateJob(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
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

func (s *IntegrationTestSuite) mockTemporalForCreateJob(returnId string) {
	s.Mocks.TemporalClientManager.
		On(
			"DoesAccountHaveNamespace", mock.Anything, mock.Anything, mock.Anything,
		).
		Return(true, nil).
		Once()
	s.Mocks.TemporalClientManager.
		On(
			"GetSyncJobTaskQueue", mock.Anything, mock.Anything, mock.Anything,
		).
		Return("sync-job", nil).
		Once()
	s.Mocks.TemporalClientManager.
		On(
			"CreateSchedule", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).
		Return(returnId, nil).
		Once()
}

func (s *IntegrationTestSuite) Test_JobService_JobHooks() {
	t := s.T()
	ctx := s.ctx

	t.Run("OSS-unimplemented", func(t *testing.T) {
		client := s.UnauthdClients.Jobs
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
		client := s.NeosyncCloudClients.GetJobClient(testAuthUserId)
		s.setUser(ctx, s.NeosyncCloudClients.GetUserClient(testAuthUserId))
		accountId := s.createPersonalAccount(ctx, s.NeosyncCloudClients.GetUserClient(testAuthUserId))

		srcconn := s.createPostgresConnection(s.NeosyncCloudClients.GetConnectionClient(testAuthUserId), accountId, "source", "test")
		destconn := s.createPostgresConnection(s.NeosyncCloudClients.GetConnectionClient(testAuthUserId), accountId, "dest", "test2")

		s.mockTemporalForCreateJob("test-id")
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
			createdHook := s.createSqlJobHook(ctx, t, client, "getjobhooks-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
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
			createdHook := s.createSqlJobHook(ctx, t, client, "getjobhook-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
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
				createdHook := s.createSqlJobHook(ctx, t, client, "deletejobhook-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
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
				createdHook := s.createSqlJobHook(ctx, t, client, "isjobhooknameavail-2", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
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
			createdHook := s.createSqlJobHook(ctx, t, client, "setjobhookenabled-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
				Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{},
			})
			require.False(t, createdHook.GetEnabled())
			resp, err := client.SetJobHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetJobHookEnabledRequest{
				Id:      createdHook.GetId(),
				Enabled: true,
			}))
			requireNoErrResp(t, resp, err)
			require.True(t, resp.Msg.GetHook().GetEnabled())
			resp, err = client.SetJobHookEnabled(ctx, connect.NewRequest(&mgmtv1alpha1.SetJobHookEnabledRequest{
				Id:      createdHook.GetId(),
				Enabled: false,
			}))
			requireNoErrResp(t, resp, err)
			require.False(t, resp.Msg.GetHook().GetEnabled())
		})

		t.Run("UpdateJobHook", func(t *testing.T) {
			createdHook := s.createSqlJobHook(ctx, t, client, "updatejobhook-1", jobResp.Msg.GetJob().GetId(), srcconn.GetId(), &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{
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
	})
}

func (s *IntegrationTestSuite) createSqlJobHook(
	ctx context.Context,
	t testing.TB,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	name string,
	jobId string,
	connectionId string,
	timing *mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing,
) *mgmtv1alpha1.JobHook {
	createResp, err := jobclient.CreateJobHook(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobHookRequest{
		JobId: jobId,
		Hook: &mgmtv1alpha1.NewJobHook{
			Name:        name,
			Description: "sql job hook test",
			Enabled:     false,
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
