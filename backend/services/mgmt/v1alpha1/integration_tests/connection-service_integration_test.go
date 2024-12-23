package integrationtests_test

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_ConnectionService_IsConnectionNameAvailable_Available() {
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	resp, err := s.OSSUnauthenticatedLicensedClients.Connections().IsConnectionNameAvailable(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.IsConnectionNameAvailableRequest{
			AccountId:      accountId,
			ConnectionName: "foo",
		}),
	)
	requireNoErrResp(s.T(), resp, err)
	require.True(s.T(), resp.Msg.GetIsAvailable())
}

func (s *IntegrationTestSuite) Test_ConnectionService_IsConnectionNameAvailable_NotAvailable() {
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
	s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "foo", "test-url")

	resp, err := s.OSSUnauthenticatedLicensedClients.Connections().IsConnectionNameAvailable(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.IsConnectionNameAvailableRequest{
			AccountId:      accountId,
			ConnectionName: "foo",
		}),
	)
	requireNoErrResp(s.T(), resp, err)
	require.False(s.T(), resp.Msg.GetIsAvailable())
}

func (s *IntegrationTestSuite) Test_ConnectionService_CheckConnectionConfig() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	conn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "foo", s.Pgcontainer.URL)

	t.Run("valid-pg-connstr", func(t *testing.T) {
		t.Parallel()

		resp, err := s.OSSUnauthenticatedLicensedClients.Connections().CheckConnectionConfig(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.CheckConnectionConfigRequest{
				ConnectionConfig: conn.GetConnectionConfig(),
			}),
		)
		requireNoErrResp(t, resp, err)
		require.True(t, resp.Msg.GetIsConnected())
		require.Empty(t, resp.Msg.GetConnectionError())
	})
}

func (s *IntegrationTestSuite) Test_ConnectionService_CreateConnection() {
	t := s.T()

	t.Run("OSS Unauthenticated Licensed", func(t *testing.T) {
		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
		client := s.OSSUnauthenticatedLicensedClients.Connections()
		t.Run("postgres-success", func(t *testing.T) {
			resp, err := client.CreateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
					AccountId: accountId,
					Name:      uuid.NewString(),
					ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
						Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
							PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
								ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
									Url: s.Pgcontainer.URL,
								},
							},
						},
					},
				}),
			)
			requireNoErrResp(t, resp, err)
		})
	})
	t.Run("OSS Authenticated Licensed", func(t *testing.T) {
		userclient := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		integrationtests_test.SetUser(s.ctx, t, userclient)
		accountId := s.createPersonalAccount(s.ctx, userclient)
		client := s.OSSAuthenticatedLicensedClients.Connections(integrationtests_test.WithUserId(testAuthUserId))
		t.Run("postgres-success", func(t *testing.T) {
			resp, err := client.CreateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
					AccountId: accountId,
					Name:      uuid.NewString(),
					ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
						Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
							PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
								ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
									Url: s.Pgcontainer.URL,
								},
							},
						},
					},
				}),
			)
			requireNoErrResp(t, resp, err)
		})
	})
	t.Run("OSS Unauthenticated Unlicensed", func(t *testing.T) {
		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedUnlicensedClients.Users())
		client := s.OSSUnauthenticatedUnlicensedClients.Connections()
		t.Run("postgres-success", func(t *testing.T) {
			resp, err := client.CreateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
					AccountId: accountId,
					Name:      uuid.NewString(),
					ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
						Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
							PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
								ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
									Url: s.Pgcontainer.URL,
								},
							},
						},
					},
				}),
			)
			requireNoErrResp(t, resp, err)
		})
		t.Run("aws-s3-failure", func(t *testing.T) {
			resp, err := client.CreateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
					AccountId: accountId,
					Name:      uuid.NewString(),
					ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
						Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
							AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
								Bucket: "foo",
							},
						},
					},
				}),
			)
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodePermissionDenied)
		})
		t.Run("gcp-cloudstorage-failure", func(t *testing.T) {
			resp, err := client.CreateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
					AccountId: accountId,
					Name:      uuid.NewString(),
					ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
						Config: &mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig{
							GcpCloudstorageConfig: &mgmtv1alpha1.GcpCloudStorageConnectionConfig{
								Bucket: "foo",
							},
						},
					},
				}),
			)
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodePermissionDenied)
		})
	})
}

func (s *IntegrationTestSuite) Test_ConnectionService_UpdateConnection() {
	t := s.T()

	t.Run("OSS Unauthenticated Licensed", func(t *testing.T) {
		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
		client := s.OSSUnauthenticatedLicensedClients.Connections()
		t.Run("postgres-success", func(t *testing.T) {
			conn := s.createPostgresConnection(client, accountId, uuid.NewString(), s.Pgcontainer.URL)

			updatedName := uuid.NewString()
			resp, err := client.UpdateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
					Id:               conn.GetId(),
					Name:             updatedName,
					ConnectionConfig: conn.GetConnectionConfig(),
				}),
			)
			requireNoErrResp(t, resp, err)
			require.Equal(t, updatedName, resp.Msg.GetConnection().GetName())
		})
	})

	t.Run("OSS Authenticated Licensed", func(t *testing.T) {
		userclient := s.OSSAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		integrationtests_test.SetUser(s.ctx, t, userclient)
		accountId := s.createPersonalAccount(s.ctx, userclient)
		client := s.OSSAuthenticatedLicensedClients.Connections(integrationtests_test.WithUserId(testAuthUserId))
		t.Run("postgres-success", func(t *testing.T) {
			conn := s.createPostgresConnection(client, accountId, uuid.NewString(), s.Pgcontainer.URL)

			updatedName := uuid.NewString()
			resp, err := client.UpdateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
					Id:               conn.GetId(),
					Name:             updatedName,
					ConnectionConfig: conn.GetConnectionConfig(),
				}),
			)
			requireNoErrResp(t, resp, err)
			require.Equal(t, updatedName, resp.Msg.GetConnection().GetName())
		})
	})

	t.Run("OSS Unauthenticated Unlicensed", func(t *testing.T) {
		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedUnlicensedClients.Users())
		client := s.OSSUnauthenticatedUnlicensedClients.Connections()
		t.Run("postgres-success", func(t *testing.T) {
			conn := s.createPostgresConnection(client, accountId, uuid.NewString(), s.Pgcontainer.URL)

			updatedName := uuid.NewString()
			resp, err := client.UpdateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
					Id:               conn.GetId(),
					Name:             updatedName,
					ConnectionConfig: conn.GetConnectionConfig(),
				}),
			)
			requireNoErrResp(t, resp, err)
			require.Equal(t, updatedName, resp.Msg.GetConnection().GetName())
		})

		t.Run("aws-s3-failure", func(t *testing.T) {
			conn := s.createPostgresConnection(client, accountId, uuid.NewString(), s.Pgcontainer.URL)

			conn.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
					AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
						Bucket: "foo",
					},
				},
			}

			resp, err := client.UpdateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
					Id:               conn.GetId(),
					ConnectionConfig: conn.GetConnectionConfig(),
				}),
			)
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodePermissionDenied)
		})

		t.Run("aws-gcp-cloudstorage-failure", func(t *testing.T) {
			conn := s.createPostgresConnection(client, accountId, uuid.NewString(), s.Pgcontainer.URL)

			conn.ConnectionConfig = &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig{
					GcpCloudstorageConfig: &mgmtv1alpha1.GcpCloudStorageConnectionConfig{
						Bucket: "foo",
					},
				},
			}

			resp, err := client.UpdateConnection(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
					Id:               conn.GetId(),
					ConnectionConfig: conn.GetConnectionConfig(),
				}),
			)
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodePermissionDenied)
		})
	})
}

func (s *IntegrationTestSuite) Test_ConnectionService_GetConnection() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	conn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "foo", s.Pgcontainer.URL)

	resp, err := s.OSSUnauthenticatedLicensedClients.Connections().GetConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: conn.GetId(),
		}),
	)
	requireNoErrResp(t, resp, err)
	require.NotNil(t, resp.Msg.GetConnection())
}

func (s *IntegrationTestSuite) Test_ConnectionService_GetConnections() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "foo", s.Pgcontainer.URL)

	resp, err := s.OSSUnauthenticatedLicensedClients.Connections().GetConnections(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionsRequest{
			AccountId: accountId,
		}),
	)
	requireNoErrResp(t, resp, err)
	require.NotEmpty(t, resp.Msg.GetConnections())
}

func (s *IntegrationTestSuite) Test_ConnectionService_DeleteConnection() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	conn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "foo", s.Pgcontainer.URL)

	resp, err := s.OSSUnauthenticatedLicensedClients.Connections().GetConnections(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionsRequest{
			AccountId: accountId,
		}),
	)
	requireNoErrResp(t, resp, err)
	require.NotEmpty(t, resp.Msg.GetConnections())

	resp2, err := s.OSSUnauthenticatedLicensedClients.Connections().DeleteConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.DeleteConnectionRequest{
			Id: conn.GetId(),
		}),
	)
	requireNoErrResp(t, resp2, err)

	// again to test idempotency
	resp2, err = s.OSSUnauthenticatedLicensedClients.Connections().DeleteConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.DeleteConnectionRequest{
			Id: conn.GetId(),
		}),
	)
	requireNoErrResp(t, resp2, err)
}

func (s *IntegrationTestSuite) Test_ConnectionService_CheckSqlQuery() {
	t := s.T()
	accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

	conn := s.createPostgresConnection(s.OSSUnauthenticatedLicensedClients.Connections(), accountId, "foo", s.Pgcontainer.URL)

	resp, err := s.OSSUnauthenticatedLicensedClients.Connections().CheckSqlQuery(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.CheckSqlQueryRequest{
			Id:    conn.GetId(),
			Query: "SELECT 1",
		}),
	)
	requireNoErrResp(t, resp, err)
	require.True(t, resp.Msg.GetIsValid())
	require.Empty(t, resp.Msg.GetErorrMessage())
}
