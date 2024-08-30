package integrationtests_test

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) createPersonalAccount(
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
) string {
	s.T().Helper()
	resp, err := userclient.SetPersonalAccount(s.ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.AccountId
}

func requireNoErrResp[T any](t testing.TB, resp *connect.Response[T], err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func requireErrResp[T any](t testing.TB, resp *connect.Response[T], err error) {
	t.Helper()
	require.Error(t, err)
	require.Nil(t, resp)
}

func requireConnectError(t testing.TB, err error, code connect.Code) {
	t.Helper()
	connectErr, ok := err.(*connect.Error)
	require.True(t, ok, fmt.Sprintf("error was not connect error %T", err))
	require.Equal(t, code, connectErr.Code())
}
