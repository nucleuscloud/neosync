package integrationtests_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/require"
)

func CreatePersonalAccount(
	ctx context.Context,
	t *testing.T,
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
) string {
	resp, err := userclient.SetPersonalAccount(ctx, connect.NewRequest(&mgmtv1alpha1.SetPersonalAccountRequest{}))
	RequireNoErrResp(t, resp, err)
	return resp.Msg.AccountId
}

func CreatePostgresConnection(
	ctx context.Context,
	t *testing.T,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId string,
	name string,
	pgurl string,
) *mgmtv1alpha1.Connection {
	resp, err := connclient.CreateConnection(
		ctx,
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
	RequireNoErrResp(t, resp, err)
	return resp.Msg.GetConnection()
}

func CreateMysqlConnection(
	ctx context.Context,
	t *testing.T,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId string,
	name string,
	mysqlurl string,
) *mgmtv1alpha1.Connection {
	resp, err := connclient.CreateConnection(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
			AccountId: accountId,
			Name:      name,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: mysqlurl,
						},
					},
				},
			},
		}),
	)
	RequireNoErrResp(t, resp, err)
	return resp.Msg.GetConnection()
}

func CreateMssqlConnection(
	ctx context.Context,
	t *testing.T,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId string,
	name string,
	mssqlurl string,
) *mgmtv1alpha1.Connection {
	resp, err := connclient.CreateConnection(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
			AccountId: accountId,
			Name:      name,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: mssqlurl,
						},
					},
				},
			},
		}),
	)
	RequireNoErrResp(t, resp, err)
	return resp.Msg.GetConnection()
}

func CreateS3Connection(
	ctx context.Context,
	t *testing.T,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId, name string,
	bucket string,
	region *string,
) *mgmtv1alpha1.Connection {
	resp, err := connclient.CreateConnection(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
			AccountId: accountId,
			Name:      name,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
					AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
						Bucket:      bucket,
						PathPrefix:  nil,
						Region:      region,
						Endpoint:    nil,
						Credentials: nil,
					},
				},
			},
		}),
	)
	RequireNoErrResp(t, resp, err)
	return resp.Msg.GetConnection()
}

func SetUser(ctx context.Context, t *testing.T, client mgmtv1alpha1connect.UserAccountServiceClient) string {
	resp, err := client.SetUser(ctx, connect.NewRequest(&mgmtv1alpha1.SetUserRequest{}))
	RequireNoErrResp(t, resp, err)
	return resp.Msg.GetUserId()
}

func CreateTeamAccount(ctx context.Context, t *testing.T, client mgmtv1alpha1connect.UserAccountServiceClient, name string) string {
	resp, err := client.CreateTeamAccount(ctx, connect.NewRequest(&mgmtv1alpha1.CreateTeamAccountRequest{Name: name}))
	RequireNoErrResp(t, resp, err)
	return resp.Msg.AccountId
}

func RequireNoErrResp[T any](t testing.TB, resp *connect.Response[T], err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func RequireErrResp[T any](t testing.TB, resp *connect.Response[T], err error) {
	t.Helper()
	require.Error(t, err)
	require.Nil(t, resp)
}
