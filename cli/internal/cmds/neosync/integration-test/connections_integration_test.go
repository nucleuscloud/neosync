package integrationtest

import (
	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) TestConnections_List() {
	resp, err := s.neosyncApi.UnathdClients.Users.SetPersonalAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	require.NoError(s.T(), err)
	connResp, err := s.neosyncApi.UnathdClients.Connections.CreateConnection(s.ctx, connect.NewRequest[mgmtv1alpha1.CreateConnectionRequest](&mgmtv1alpha1.CreateConnectionRequest{
		AccountId: resp.Msg.AccountId,
		Name:      "source",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
						Url: s.postgres.URL,
					},
				},
			},
		},
	}))
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), connResp.Msg.GetConnection())
}
