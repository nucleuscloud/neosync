package posttablesync_activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_New(t *testing.T) {
	sqlmanagerMock := mockSqlManager()
	a := New(mgmtv1alpha1connect.NewMockJobServiceClient(t), sqlmanagerMock, mgmtv1alpha1connect.NewMockConnectionServiceClient((t)))
	require.NotNil(t, a)
}

func Test_Activity_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	name := "public.users.insert"
	destConnId := "c9b6ce58-5c8e-4dce-870d-96841b19d988"
	configs := mockPostTableSyncConfigs(name, destConnId)
	configBits, err := json.Marshal(configs)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			if r.Msg.GetId().GetAccountId() == accountId && r.Msg.GetId().GetExternalId() == shared.GetPostTableSyncConfigExternalId(name) {
				return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{
					Value: configBits,
				}), nil
			}
			return nil, errors.New("invalid test account id")
		},
	))
	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			if r.Msg.GetId() == destConnId {
				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
					Connection: &mgmtv1alpha1.Connection{
						Id:   destConnId,
						Name: "source",
						ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
										Url: "url",
									},
								},
							},
						},
					},
				}), nil
			}
			return nil, fmt.Errorf("unknown test connection")
		},
	))
	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlmanagerMock := mockSqlManager()

	activity := New(jobclient, sqlmanagerMock, connclient)

	env.RegisterActivity(activity)

	val, err := env.ExecuteActivity(activity.RunPostTableSync, &RunPostTableSyncRequest{Name: name, AccountId: accountId})
	require.NoError(t, err)
	res := &RunPostTableSyncResponse{}
	err = val.Get(res)
	require.NoError(t, err)
}

func Test_Activity_RunContextNotFound(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	name := "public.users.insert"
	destConnId := "c9b6ce58-5c8e-4dce-870d-96841b19d988"

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			return nil, errors.New("no run context exists")
		},
	))
	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			if r.Msg.GetId() == destConnId {
				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
					Connection: &mgmtv1alpha1.Connection{
						Id:   destConnId,
						Name: "source",
						ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
										Url: "url",
									},
								},
							},
						},
					},
				}), nil
			}
			return nil, fmt.Errorf("unknown test connection")
		},
	))
	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlmanagerMock := mockSqlManager()

	activity := New(jobclient, sqlmanagerMock, connclient)

	env.RegisterActivity(activity)

	_, err := env.ExecuteActivity(activity.RunPostTableSync, &RunPostTableSyncRequest{Name: name, AccountId: accountId})
	require.NoError(t, err)
}

func Test_Activity_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	accountId := uuid.NewString()
	name := "public.users.insert"
	destConnId := "c9b6ce58-5c8e-4dce-870d-96841b19d988"

	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			return nil, errors.New("other error")
		},
	))
	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			if r.Msg.GetId() == destConnId {
				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
					Connection: &mgmtv1alpha1.Connection{
						Id:   destConnId,
						Name: "source",
						ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
										Url: "url",
									},
								},
							},
						},
					},
				}), nil
			}
			return nil, fmt.Errorf("unknown test connection")
		},
	))
	srv := startHTTPServer(t, mux)

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	sqlmanagerMock := mockSqlManager()

	activity := New(jobclient, sqlmanagerMock, connclient)

	env.RegisterActivity(activity)

	_, err := env.ExecuteActivity(activity.RunPostTableSync, &RunPostTableSyncRequest{Name: name, AccountId: accountId})
	require.Error(t, err)
}

func mockPostTableSyncConfigs(name, destConnId string) map[string]*shared.PostTableSyncConfig {
	configs := map[string]*shared.PostTableSyncConfig{}
	destConfigs := map[string]*shared.PostTableSyncDestConfig{}
	destConfigs[destConnId] = &shared.PostTableSyncDestConfig{
		Statements: []string{"reset-sequence"},
	}
	configs[name] = &shared.PostTableSyncConfig{
		DestinationConfigs: destConfigs,
	}
	return configs
}

func mockSqlManager() *sqlmanager.SqlManager {
	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}
	mssqlpoolmap := &sync.Map{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mssqlquerier := mssql_queries.New()
	sqlconnector := &sqlconnect.SqlOpenConnector{}
	return sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlconnector)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
