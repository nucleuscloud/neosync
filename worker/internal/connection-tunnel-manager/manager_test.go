package connectiontunnelmanager

import (
	"log/slog"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_NewConnectionTunnelManager(t *testing.T) {
	require.NotNil(t, NewConnectionTunnelManager[any, any](nil))
}

func Test_ConnectionTunnelManager_GetConnectionString(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	connstr, err := mgr.GetConnectionString("111", conn, slog.Default())
	require.NoError(t, err)
	require.Equal(t, "postgres://foo:bar@localhost:5432/test", connstr)
}

func Test_ConnectionTunnelManager_GetConnectionString_Unique_Conns(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	conn1 := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.PostgresConnection{
							Host: "1",
						},
					},
				},
			},
		},
	}
	conn2 := &mgmtv1alpha1.Connection{
		Id: "2",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.PostgresConnection{
							Host: "2",
						},
					},
				},
			},
		},
	}

	provider.On("GetConnectionDetails", conn1.ConnectionConfig, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)

	conn2Pg := getPgGenDbConfig(t)
	conn2Pg.User = "foo2"
	provider.On("GetConnectionDetails", conn2.ConnectionConfig, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: conn2Pg,
		}, nil)

	connstr, err := mgr.GetConnectionString("111", conn1, slog.Default())
	require.NoError(t, err)
	require.Equal(t, "postgres://foo:bar@localhost:5432/test", connstr)
	connstr, err = mgr.GetConnectionString("111", conn2, slog.Default())
	require.NoError(t, err)
	require.Equal(t, "postgres://foo2:bar@localhost:5432/test", connstr)
}

func Test_ConnectionTunnelManager_GetConnectionString_Parallel_Sessions_Same_Connection(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		_, err := mgr.GetConnectionString("111", conn, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnectionString("222", conn, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnectionString("333", conn, slog.Default())
		return err
	})
	err := errgrp.Wait()
	require.NoError(t, err)

	provider.AssertNumberOfCalls(t, "GetConnectionDetails", 1)
}

func Test_ConnectionTunnelManager_GetConnectionClient(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	provider.On("GetConnectionClientConfig", mock.Anything).Return(struct{}{}, nil)
	provider.On("GetConnectionClient", "postgres", "postgres://foo:bar@localhost:5432/test", mock.Anything).Return(neosync_benthos_sql.NewMockSqlDbtx(t), nil)

	db, err := mgr.GetConnection("111", conn, slog.Default())
	require.NoError(t, err)
	require.NotNil(t, db)
}

func Test_ConnectionTunnelManager_GetConnection_Parallel_Sessions_Same_Connection(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	provider.On("GetConnectionClientConfig", mock.Anything).Return(struct{}{}, nil)
	provider.On("GetConnectionClient", "postgres", "postgres://foo:bar@localhost:5432/test", mock.Anything).Return(neosync_benthos_sql.NewMockSqlDbtx(t), nil)

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		_, err := mgr.GetConnection("111", conn, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnection("222", conn, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnection("333", conn, slog.Default())
		return err
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	provider.AssertNumberOfCalls(t, "GetConnectionDetails", 1)
	provider.AssertNumberOfCalls(t, "GetConnectionClient", 1)
}

func Test_ConnectionTunnelManager_ReleaseSession(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	require.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	_, err := mgr.GetConnectionString("111", conn, slog.Default())
	require.NoError(t, err)

	require.True(t, mgr.ReleaseSession("111"), "released an existing session")
}

func Test_ConnectionTunnelManager_close(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	require.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	mockDb := neosync_benthos_sql.NewMockSqlDbtx(t)
	provider.On("GetConnectionClientConfig", mock.Anything).Return(struct{}{}, nil)
	provider.On("GetConnectionClient", "postgres", "postgres://foo:bar@localhost:5432/test", mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	_, err := mgr.GetConnection("111", conn, slog.Default())
	require.NoError(t, err)

	require.NotEmpty(t, mgr.connDetailsMap, "has an active connection")
	require.NotEmpty(t, mgr.connMap, "has an active connection")
	mgr.close()
	require.NotEmpty(t, mgr.connDetailsMap, "not empty due to active session")
	require.NotEmpty(t, mgr.connMap, "not empty due to active session")
	require.True(t, mgr.ReleaseSession("111"), "released an existing session")
	mgr.close()
	require.Empty(t, mgr.connDetailsMap, "now empty due to no active sessions")
	require.Empty(t, mgr.connMap, "now empty due to no active sessions")
}

func Test_ConnectionTunnelManager_hardClose(t *testing.T) {
	provider := NewMockConnectionProvider[any, any](t)
	mgr := NewConnectionTunnelManager(provider)

	require.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	mockDb := neosync_benthos_sql.NewMockSqlDbtx(t)
	provider.On("GetConnectionClientConfig", mock.Anything).Return(struct{}{}, nil)
	provider.On("GetConnectionClient", "postgres", "postgres://foo:bar@localhost:5432/test", mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	_, err := mgr.GetConnection("111", conn, slog.Default())
	require.NoError(t, err)

	require.NotEmpty(t, mgr.connDetailsMap, "has an active connection")
	require.NotEmpty(t, mgr.connMap, "has an active connection")
	mgr.hardClose()
	require.Empty(t, mgr.connDetailsMap, "now empty due to no active sessions")
	require.Empty(t, mgr.connMap, "now empty due to no active sessions")
}

func Test_getUniqueConnectionIdsFromSessions(t *testing.T) {
	output := getUniqueConnectionIdsFromSessions(
		map[string]map[string]struct{}{
			"1": {
				"111": {},
				"222": {},
			},
			"2": {
				"111": {},
				"333": {},
			},
		},
	)
	require.Len(t, output, 3)
	require.Contains(t, output, "111")
	require.Contains(t, output, "222")
	require.Contains(t, output, "333")
}

func Test_getDriverFromConnection(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		driver, err := getDriverFromConnection(nil)
		require.Error(t, err)
		require.Empty(t, driver)
	})

	t.Run("postgres", func(t *testing.T) {
		driver, err := getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		}})
		require.NoError(t, err)
		require.Equal(t, "postgres", driver)
	})

	t.Run("mysql", func(t *testing.T) {
		driver, err := getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{},
		}})
		require.NoError(t, err)
		require.Equal(t, "mysql", driver)
	})

	t.Run("mssql", func(t *testing.T) {
		driver, err := getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{},
		}})
		require.NoError(t, err)
		require.Equal(t, "sqlserver", driver)
	})

	t.Run("unsupported", func(t *testing.T) {
		driver, err := getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{},
		}})
		require.Error(t, err)
		require.Empty(t, driver)
	})

	t.Run("mongo", func(t *testing.T) {
		driver, err := getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MongoConfig{},
		}})
		require.NoError(t, err)
		require.Equal(t, "mongodb", driver)
	})
}

func getPgGenDbConfig(t *testing.T) sqlconnect.GeneralDbConnectConfig {
	t.Helper()
	port := int32(5432)
	db := "test"
	return sqlconnect.GeneralDbConnectConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     &port,
		Database: &db,
		User:     "foo",
		Pass:     "bar",
	}
}
