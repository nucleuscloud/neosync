package integrationtests_test

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) createPersonalAccount(
	ctx context.Context,
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
) string {
	s.T().Helper()
	resp, err := userclient.SetPersonalAccount(ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
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

func requireConnectError(t testing.TB, err error, expectedCode connect.Code) {
	t.Helper()
	connectErr, ok := err.(*connect.Error)
	require.True(t, ok, fmt.Sprintf("error was not connect error %T", err))
	require.Equal(t, expectedCode, connectErr.Code(), fmt.Sprintf("%d: %s", connectErr.Code(), connectErr.Message()))
}

func (s *IntegrationTestSuite) setMaxAllowedRecords(
	ctx context.Context,
	accountId string,
	maxAllowed uint64, //nolint:unparam // want to leave it here to allow other options in the future
) error {
	accountUuid, err := neosyncdb.ToUuid(accountId)
	if err != nil {
		return err
	}
	_, err = s.neosyncQuerier.SetAccountMaxAllowedRecords(ctx, s.pgpool, db_queries.SetAccountMaxAllowedRecordsParams{
		MaxAllowedRecords: pgtype.Int8{Int64: int64(maxAllowed), Valid: true},
		AccountId:         accountUuid,
	})
	return err
}

func getAccountIds(t testing.TB, accounts []*mgmtv1alpha1.UserAccount) []string {
	t.Helper()
	output := []string{}
	for _, acc := range accounts {
		output = append(output, acc.GetId())
	}
	return output
}
