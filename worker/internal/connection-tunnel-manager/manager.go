package connectiontunnelmanager

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type ConnectionProvider[T any] interface {
	GetConnectionClient(connectionConfig *mgmtv1alpha1.ConnectionConfig) (T, error)
	CloseClientConnection(client T) error
}

type ConnectionTunnelManager[T any] struct {
	connectionProvider ConnectionProvider[T]

	sessionMap map[string]map[string]struct{}
	sessionMu  sync.RWMutex

	connMap map[string]T
	connMu  sync.RWMutex

	shutdown chan any
}

type ConnectionDetails interface {
	String() string
}

type Interface[T any] interface {
	GetConnection(session string, connection *mgmtv1alpha1.Connection, logger *slog.Logger) (T, error)

	ReleaseSession(session string) bool
	Shutdown()
	Reaper()
}

var _ Interface[any] = &ConnectionTunnelManager[any]{} // enforces ConnectionTunnelManager always conforms to the interface

func NewConnectionTunnelManager[T any](connectionProvider ConnectionProvider[T]) *ConnectionTunnelManager[T] {
	return &ConnectionTunnelManager[T]{
		connectionProvider: connectionProvider,
		sessionMap:         map[string]map[string]struct{}{},
		connMap:            map[string]T{},
	}
}

func (c *ConnectionTunnelManager[T]) GetConnection(
	session string,
	connection *mgmtv1alpha1.Connection,
	logger *slog.Logger,
) (T, error) {
	c.connMu.RLock()
	existingDb, ok := c.connMap[connection.Id]
	if ok {
		c.bindSession(session, connection.Id)
		c.connMu.RUnlock()
		return existingDb, nil
	}
	c.connMu.RUnlock()
	c.connMu.Lock()
	defer c.connMu.Unlock()

	existingDb, ok = c.connMap[connection.Id]
	if ok {
		c.bindSession(session, connection.Id)
		return existingDb, nil
	}

	connectionClient, err := c.connectionProvider.GetConnectionClient(connection.GetConnectionConfig())
	if err != nil {
		var result T
		return result, err
	}

	c.connMap[connection.Id] = connectionClient
	c.bindSession(session, connection.Id)
	return connectionClient, nil
}

func (c *ConnectionTunnelManager[T]) ReleaseSession(session string) bool {
	c.sessionMu.RLock()
	connMap, ok := c.sessionMap[session]
	if !ok || len(connMap) == 0 {
		c.sessionMu.RUnlock()
		return false
	}
	c.sessionMu.RUnlock()
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	connMap, ok = c.sessionMap[session]
	if !ok || len(connMap) == 0 {
		return false
	}
	delete(c.sessionMap, session)
	return true
}

func (c *ConnectionTunnelManager[T]) bindSession(session, connectionId string) {
	c.sessionMu.RLock()
	connmap, ok := c.sessionMap[session]
	if ok {
		if _, ok := connmap[connectionId]; ok {
			c.sessionMu.RUnlock()
			return
		}
	}
	c.sessionMu.RUnlock()
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if _, ok := c.sessionMap[session]; !ok {
		c.sessionMap[session] = map[string]struct{}{}
	}
	c.sessionMap[session][connectionId] = struct{}{}
}

func (c *ConnectionTunnelManager[T]) Shutdown() {
	c.shutdown <- struct{}{}
}

func (c *ConnectionTunnelManager[T]) Reaper() {
	for {
		select {
		case <-c.shutdown:
			c.hardClose()
			return
		case <-time.After(1 * time.Minute):
			c.close()
		}
	}
}

func (c *ConnectionTunnelManager[T]) hardClose() {
	c.connMu.Lock()
	c.sessionMu.Lock()
	for connId, dbConn := range c.connMap {
		err := c.connectionProvider.CloseClientConnection(dbConn)
		if err != nil {
			slog.Error(fmt.Sprintf("unable to fully close client connection during hard close: %s", err.Error()))
		}
		delete(c.connMap, connId)
	}

	for sessionId := range c.sessionMap {
		delete(c.sessionMap, sessionId)
	}
	c.connMu.Unlock()
	c.sessionMu.Unlock()
}

func (c *ConnectionTunnelManager[T]) close() {
	c.connMu.Lock()
	c.sessionMu.Lock()
	sessionConnections := getUniqueConnectionIdsFromSessions(c.sessionMap)
	for connId, dbConn := range c.connMap {
		if _, ok := sessionConnections[connId]; !ok {
			err := c.connectionProvider.CloseClientConnection(dbConn)
			if err != nil {
				slog.Error(fmt.Sprintf("unable to fully close client connection during close: %s", err.Error()))
			}
			delete(c.connMap, connId)
		}
	}
	c.sessionMu.Unlock()
	c.connMu.Unlock()
}

func getUniqueConnectionIdsFromSessions(sessionMap map[string]map[string]struct{}) map[string]struct{} {
	connSet := map[string]struct{}{}
	for _, sessConnSet := range sessionMap {
		for key := range sessConnSet {
			connSet[key] = struct{}{}
		}
	}
	return connSet
}
