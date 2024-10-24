package connectiontunnelmanager

import (
	"io"
	"log/slog"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

var (
	discardLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
)

func Test_NewConnectionTunnelManager(t *testing.T) {
	require.NotNil(t, NewConnectionTunnelManager[any](nil))
}

func Test_ConnectionTunnelManager_GetConnectionClient(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionTunnelManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionClient", mock.Anything).Return(&struct{}{}, nil)

	db, err := mgr.GetConnection("111", conn, discardLogger)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func Test_ConnectionTunnelManager_GetConnection_Parallel_Sessions_Same_Connection(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionTunnelManager(provider)

	cc := &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
	}
	conn := &mgmtv1alpha1.Connection{
		Id:               "1",
		ConnectionConfig: cc,
	}

	provider.On("GetConnectionClient", cc).Return(&struct{}{}, nil)

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		_, err := mgr.GetConnection("111", conn, discardLogger)
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnection("222", conn, discardLogger)
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnection("333", conn, discardLogger)
		return err
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	provider.AssertNumberOfCalls(t, "GetConnectionClient", 1)
}

func Test_ConnectionTunnelManager_ReleaseSession(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionTunnelManager(provider)

	require.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionClient", mock.Anything).
		Return(&struct{}{}, nil)
	_, err := mgr.GetConnection("111", conn, discardLogger)
	require.NoError(t, err)

	require.True(t, mgr.ReleaseSession("111"), "released an existing session")
}

func Test_ConnectionTunnelManager_close(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionTunnelManager(provider)

	require.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockDb := &struct{}{}

	provider.On("GetConnectionClient", mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	_, err := mgr.GetConnection("111", conn, discardLogger)
	require.NoError(t, err)

	require.NotEmpty(t, mgr.connMap, "has an active connection")
	mgr.close()
	require.NotEmpty(t, mgr.connMap, "not empty due to active session")
	require.True(t, mgr.ReleaseSession("111"), "released an existing session")
	mgr.close()
	require.Empty(t, mgr.connMap, "now empty due to no active sessions")
}

func Test_ConnectionTunnelManager_hardClose(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionTunnelManager(provider)

	require.False(t, mgr.ReleaseSession("111"), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockDb := struct{}{}
	provider.On("GetConnectionClient", mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	_, err := mgr.GetConnection("111", conn, discardLogger)
	require.NoError(t, err)

	require.NotEmpty(t, mgr.connMap, "has an active connection")
	mgr.hardClose()
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
