package connectionmanager

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

type ConnectionManager[T any] struct {
	connectionProvider ConnectionProvider[T]

	config *managerConfig

	sessionMap map[string]map[string]struct{}
	sessionMu  sync.RWMutex

	connMap map[string]T
	connMu  sync.RWMutex

	shutdown chan any
}

type ConnectionInput interface {
	GetId() string
	GetConnectionConfig() *mgmtv1alpha1.ConnectionConfig
}

type Interface[T any] interface {
	GetConnection(session string, connection ConnectionInput, logger *slog.Logger) (T, error)

	ReleaseSession(session string) bool
	Shutdown()
	Reaper()
}

var _ Interface[any] = &ConnectionManager[any]{}

type managerConfig struct {
	closeOnRelease bool
}

func WithCloseOnRelease() ManagerOption {
	return func(mc *managerConfig) {
		mc.closeOnRelease = true
	}
}

type ManagerOption func(*managerConfig)

func NewConnectionManager[T any](
	connectionProvider ConnectionProvider[T],
	opts ...ManagerOption,
) *ConnectionManager[T] {
	cfg := &managerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &ConnectionManager[T]{
		connectionProvider: connectionProvider,
		sessionMap:         map[string]map[string]struct{}{},
		connMap:            map[string]T{},
		config:             cfg,
	}
}

func (c *ConnectionManager[T]) GetConnection(
	session string,
	connection ConnectionInput,
	logger *slog.Logger,
) (T, error) {
	c.connMu.RLock()
	existingDb, ok := c.connMap[connection.GetId()]
	if ok {
		c.bindSession(session, connection.GetId())
		c.connMu.RUnlock()
		return existingDb, nil
	}
	c.connMu.RUnlock()
	c.connMu.Lock()
	defer c.connMu.Unlock()

	existingDb, ok = c.connMap[connection.GetId()]
	if ok {
		c.bindSession(session, connection.GetId())
		return existingDb, nil
	}

	connectionClient, err := c.connectionProvider.GetConnectionClient(connection.GetConnectionConfig())
	if err != nil {
		var result T
		return result, err
	}

	c.connMap[connection.GetId()] = connectionClient
	c.bindSession(session, connection.GetId())
	return connectionClient, nil
}

func (c *ConnectionManager[T]) ReleaseSession(session string) bool {
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

	sessionConnIds := getConnectionIds(connMap)
	delete(c.sessionMap, session)
	if c.config.closeOnRelease {
		c.closeUnusedConnections(sessionConnIds)
	}
	return true
}

func (c *ConnectionManager[T]) closeUnusedConnections(candidateConnIds []string) {
	remainingConnections := getUniqueConnectionIdsFromSessions(c.sessionMap)

	c.connMu.Lock()
	defer c.connMu.Unlock()

	for _, connId := range candidateConnIds {
		if _, stillInUse := remainingConnections[connId]; !stillInUse {
			if dbConn, exists := c.connMap[connId]; exists {
				if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
					slog.Error(fmt.Sprintf("unable to close client connection during release: %s", err.Error()))
				}
				delete(c.connMap, connId)
			}
		}
	}
}

func (c *ConnectionManager[T]) bindSession(session, connectionId string) {
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

func (c *ConnectionManager[T]) Shutdown() {
	c.shutdown <- struct{}{}
}

func (c *ConnectionManager[T]) Reaper() {
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

func (c *ConnectionManager[T]) hardClose() {
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

func (c *ConnectionManager[T]) close() {
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

func getConnectionIds(connMap map[string]struct{}) []string {
	ids := make([]string, 0, len(connMap))
	for connId := range connMap {
		ids = append(ids, connId)
	}
	return ids
}
