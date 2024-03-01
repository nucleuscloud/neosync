package sync_activity

import (
	"log/slog"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/errgroup"
)

func Test_NewConnectionTunnelManager(t *testing.T) {
	assert.NotNil(t, NewConnectionTunnelManager(nil))
}

func Test_ConnectionTunnelManager_GetConnectionString(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockSqlProvider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	connstr, err := mgr.GetConnectionString("111", conn, slog.Default())
	assert.NoError(t, err)
	assert.Equal(t, "postgres://foo:bar@localhost:5432/test", connstr)
}

func Test_ConnectionTunnelManager_GetConnectionString_Unique_Conns(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

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

	mockSqlProvider.On("GetConnectionDetails", conn1.ConnectionConfig, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)

	conn2Pg := getPgGenDbConfig(t)
	conn2Pg.User = "foo2"
	mockSqlProvider.On("GetConnectionDetails", conn2.ConnectionConfig, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: conn2Pg,
		}, nil)

	connstr, err := mgr.GetConnectionString("111", conn1, slog.Default())
	assert.NoError(t, err)
	assert.Equal(t, "postgres://foo:bar@localhost:5432/test", connstr)
	connstr, err = mgr.GetConnectionString("111", conn2, slog.Default())
	assert.NoError(t, err)
	assert.Equal(t, "postgres://foo2:bar@localhost:5432/test", connstr)
}

func Test_ConnectionTunnelManager_GetConnectionString_Parallel_Sessions_Same_Connection(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockSqlProvider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
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
	assert.NoError(t, err)

	mockSqlProvider.AssertNumberOfCalls(t, "GetConnectionDetails", 1)
}

func Test_ConnectionTunnelManager_GetConnection(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockSqlProvider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	mockSqlProvider.On("DbOpen", "postgres", "postgres://foo:bar@localhost:5432/test").
		Return(NewMocksqlDbtx(t), nil)
	db, err := mgr.GetConnection("111", conn, slog.Default())
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func Test_ConnectionTunnelManager_GetConnection_Parallel_Sessions_Same_Connection(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockSqlProvider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	mockSqlProvider.On("DbOpen", "postgres", "postgres://foo:bar@localhost:5432/test").
		Return(NewMocksqlDbtx(t), nil)

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
	assert.NoError(t, err)
	mockSqlProvider.AssertNumberOfCalls(t, "GetConnectionDetails", 1)
	mockSqlProvider.AssertNumberOfCalls(t, "DbOpen", 1)
}

func Test_ConnectionTunnelManager_ReleaseSession(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

	assert.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockSqlProvider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	_, err := mgr.GetConnectionString("111", conn, slog.Default())
	assert.NoError(t, err)

	assert.True(t, mgr.ReleaseSession("111"), "released an existing session")
}

func Test_ConnectionTunnelManager_close(t *testing.T) {
	mockSqlProvider := NewMocksqlProvider(t)
	mgr := NewConnectionTunnelManager(mockSqlProvider)

	assert.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockSqlProvider.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
		Return(&sqlconnect.ConnectionDetails{
			GeneralDbConnectConfig: getPgGenDbConfig(t),
		}, nil)
	mockDb := NewMocksqlDbtx(t)
	mockSqlProvider.On("DbOpen", "postgres", "postgres://foo:bar@localhost:5432/test").
		Return(mockDb, nil)
	mockDb.On("Close").Return(nil)

	_, err := mgr.GetConnection("111", conn, slog.Default())
	assert.NoError(t, err)

	assert.NotEmpty(t, mgr.connDetailsMap, "has an active connection")
	assert.NotEmpty(t, mgr.connMap, "has an active connection")
	mgr.close()
	assert.NotEmpty(t, mgr.connDetailsMap, "not empty due to active session")
	assert.NotEmpty(t, mgr.connMap, "not empty due to active session")
	assert.True(t, mgr.ReleaseSession("111"), "released an existing session")
	mgr.close()
	assert.Empty(t, mgr.connDetailsMap, "now empty due to no active sessions")
	assert.Empty(t, mgr.connMap, "now empty due to no active sessions")
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
	assert.Len(t, output, 3)
	assert.Contains(t, output, "111")
	assert.Contains(t, output, "222")
	assert.Contains(t, output, "333")
}

func Test_getDriverFromConnection(t *testing.T) {
	driver, err := getDriverFromConnection(nil)
	assert.Error(t, err)
	assert.Empty(t, driver)

	driver, err = getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
	}})
	assert.NoError(t, err)
	assert.Equal(t, "postgres", driver)

	driver, err = getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{},
	}})
	assert.NoError(t, err)
	assert.Equal(t, "mysql", driver)

	driver, err = getDriverFromConnection(&mgmtv1alpha1.Connection{ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{},
	}})
	assert.Error(t, err)
	assert.Empty(t, driver)
}

func getPgGenDbConfig(t *testing.T) sqlconnect.GeneralDbConnectConfig {
	t.Helper()
	return sqlconnect.GeneralDbConnectConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		User:     "foo",
		Pass:     "bar",
	}
}
