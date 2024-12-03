package connectionmanager

import (
	"fmt"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_NewConnectionTunnelManager(t *testing.T) {
	require.NotNil(t, NewConnectionManager[any](nil))
}

func Test_ConnectionTunnelManager_GetConnectionClient(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(&struct{}{}, nil)

	db, err := mgr.GetConnection(NewSession("111"), conn, testutil.GetTestLogger(t))
	require.NoError(t, err)
	require.NotNil(t, db)
}

func Test_ConnectionTunnelManager_GetConnection_Parallel_Sessions_Same_Connection(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	cc := &mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
	}
	conn := &mgmtv1alpha1.Connection{
		Id:               "1",
		ConnectionConfig: cc,
	}

	provider.On("GetConnectionClient", cc, mock.Anything).Return(&struct{}{}, nil)

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		_, err := mgr.GetConnection(NewSession("111"), conn, testutil.GetTestLogger(t))
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnection(NewSession("222"), conn, testutil.GetTestLogger(t))
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetConnection(NewSession("333"), conn, testutil.GetTestLogger(t))
		return err
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	provider.AssertNumberOfCalls(t, "GetConnectionClient", 1)
}

func Test_ConnectionTunnelManager_ReleaseSession(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	require.False(t, mgr.ReleaseSession(NewSession("111"), testutil.GetTestLogger(t)), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionClient", mock.Anything, mock.Anything).
		Return(&struct{}{}, nil)
	_, err := mgr.GetConnection(NewSession("111"), conn, testutil.GetTestLogger(t))
	require.NoError(t, err)

	require.True(t, mgr.ReleaseSession(NewSession("111"), testutil.GetTestLogger(t)), "released an existing session")
}

func Test_ConnectionTunnelManager_cleanUnusedConnections(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	require.False(t, mgr.ReleaseSession(NewSession("111"), testutil.GetTestLogger(t)), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockDb := &struct{}{}

	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	_, err := mgr.GetConnection(NewSession("111"), conn, testutil.GetTestLogger(t))
	require.NoError(t, err)

	require.NotEmpty(t, mgr.groupConnMap, "has an active connection")
	mgr.cleanUnusedConnections(testutil.GetTestLogger(t))
	require.NotEmpty(t, mgr.groupConnMap, "not empty due to active session")
	require.True(t, mgr.ReleaseSession(NewSession("111"), testutil.GetTestLogger(t)), "released an existing session")
	mgr.cleanUnusedConnections(testutil.GetTestLogger(t))
	require.Empty(t, mgr.groupConnMap, "now empty due to no active sessions")
}

func Test_ConnectionTunnelManager_hardClose(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	session := NewSession("111")
	require.False(t, mgr.ReleaseSession(session, testutil.GetTestLogger(t)), "currently no session")

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	mockDb := struct{}{}
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	_, err := mgr.GetConnection(session, conn, testutil.GetTestLogger(t))
	require.NoError(t, err)

	require.NotEmpty(t, mgr.groupConnMap, "has an active connection")
	mgr.hardClose(testutil.GetTestLogger(t))
	require.Empty(t, mgr.groupConnMap, "now empty due to no active sessions")
}

func Test_getUniqueConnectionIdsFromSessions(t *testing.T) {
	output := getUniqueConnectionIdsFromGroupSessions(
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

func Test_ConnectionManager_CloseOnRelease_Option(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider, WithCloseOnRelease())

	// Create two different mock DBs
	mockDb1 := &struct{}{}
	mockDb2 := &struct{}{}

	// First call returns mockDb1, second call returns mockDb2
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb1, nil).Once()
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb2, nil).Once()
	provider.On("CloseClientConnection", mockDb1).Return(nil).Once()
	provider.On("CloseClientConnection", mockDb2).Return(nil).Once()

	conn1 := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}
	conn2 := &mgmtv1alpha1.Connection{
		Id: "2",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	session1 := NewSession("session1")
	session2 := NewSession("session2")

	// Create two sessions using the same connection
	_, err := mgr.GetConnection(session1, conn1, testutil.GetTestLogger(t))
	require.NoError(t, err)
	_, err = mgr.GetConnection(session2, conn1, testutil.GetTestLogger(t))
	require.NoError(t, err)

	// Create one session using a different connection
	_, err = mgr.GetConnection(session1, conn2, testutil.GetTestLogger(t))
	require.NoError(t, err)

	// Release session1 - conn1 should stay alive due to session2, but conn2 should be closed
	require.True(t, mgr.ReleaseSession(session1, testutil.GetTestLogger(t)))
	require.Len(t, mgr.groupConnMap[""], 1)                     // Only conn1 should still exist
	provider.AssertNumberOfCalls(t, "CloseClientConnection", 1) // Only conn2 should be closed

	// Release session2 - conn1 should now be closed
	require.True(t, mgr.ReleaseSession(session2, testutil.GetTestLogger(t)))
	require.Empty(t, mgr.groupConnMap) // All connections should be closed
	provider.AssertNumberOfCalls(t, "CloseClientConnection", 2)
}

func Test_ConnectionManager_Concurrent_Sessions_Different_Connections(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	mockDb := &struct{}{}
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb, nil)

	connections := make([]*mgmtv1alpha1.Connection, 10)
	for i := 0; i < 10; i++ {
		connections[i] = &mgmtv1alpha1.Connection{
			Id: fmt.Sprintf("conn-%d", i),
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
			},
		}
	}

	errgrp := errgroup.Group{}
	for i := 0; i < 10; i++ {
		conn := connections[i]
		errgrp.Go(func() error {
			_, err := mgr.GetConnection(NewSession(fmt.Sprintf("session-%s", conn.Id)), conn, testutil.GetTestLogger(t))
			return err
		})
	}
	err := errgrp.Wait()
	require.NoError(t, err)
	require.Len(t, mgr.groupConnMap[""], 10)
	require.Len(t, mgr.groupSessionMap[""], 10)
}

func Test_ConnectionManager_Error_During_Connection(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("connection error"))

	_, err := mgr.GetConnection(NewSession("session1"), conn, testutil.GetTestLogger(t))
	require.Error(t, err)
	require.Empty(t, mgr.groupConnMap)
	require.Empty(t, mgr.groupSessionMap)
}

func Test_ConnectionManager_Error_During_Close(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider, WithCloseOnRelease())

	mockDb := &struct{}{}
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(fmt.Errorf("close error"))

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	_, err := mgr.GetConnection(NewSession("session1"), conn, testutil.GetTestLogger(t))
	require.NoError(t, err)

	// Even with close error, session and connection should be removed
	require.True(t, mgr.ReleaseSession(NewSession("session1"), testutil.GetTestLogger(t)))
	require.Empty(t, mgr.groupConnMap)
	require.Empty(t, mgr.groupSessionMap)
}

func Test_ConnectionManager_Concurrent_GetConnection_And_Release(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider, WithCloseOnRelease())

	mockDb := &struct{}{}
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	conn := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	// Run 100 goroutines that repeatedly get and release connections
	errgrp := errgroup.Group{}
	for i := 0; i < 100; i++ {
		sessionId := fmt.Sprintf("session-%d", i)
		errgrp.Go(func() error {
			for j := 0; j < 10; j++ {
				_, err := mgr.GetConnection(NewSession(sessionId), conn, testutil.GetTestLogger(t))
				if err != nil {
					return err
				}
				mgr.ReleaseSession(NewSession(sessionId), testutil.GetTestLogger(t))
			}
			return nil
		})
	}
	err := errgrp.Wait()
	require.NoError(t, err)
}

func Test_ConnectionManager_Reaper_With_Active_Sessions(t *testing.T) {
	provider := NewMockConnectionProvider[any](t)
	mgr := NewConnectionManager(provider)

	mockDb := &struct{}{}
	provider.On("GetConnectionClient", mock.Anything, mock.Anything).Return(mockDb, nil)
	provider.On("CloseClientConnection", mockDb).Return(nil)

	conn1 := &mgmtv1alpha1.Connection{
		Id: "1",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}
	conn2 := &mgmtv1alpha1.Connection{
		Id: "2",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		},
	}

	// Create active session with conn1
	_, err := mgr.GetConnection(NewSession("session1"), conn1, testutil.GetTestLogger(t))
	require.NoError(t, err)

	// Create and release session with conn2
	_, err = mgr.GetConnection(NewSession("session2"), conn2, testutil.GetTestLogger(t))
	require.NoError(t, err)
	require.True(t, mgr.ReleaseSession(NewSession("session2"), testutil.GetTestLogger(t)))

	// Run reaper
	mgr.cleanUnusedConnections(testutil.GetTestLogger(t))

	// Verify conn2 was cleaned up but conn1 remains
	require.Len(t, mgr.groupConnMap, 1)
	_, exists := mgr.groupConnMap[""][conn1.Id]
	require.True(t, exists)
}
