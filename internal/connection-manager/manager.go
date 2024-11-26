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
	sessionMu  *sync.RWMutex

	connMap map[string]T
	connMu  *sync.RWMutex

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

// Causes the ConnectionManager to automatically close connections when the last session is released
// If this is not provided, connections will be held on to until Shutdown() or Reaper() is enabled
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
		shutdown:           make(chan any),
		sessionMu:          &sync.RWMutex{},
		connMu:             &sync.RWMutex{},
	}
}

func (c *ConnectionManager[T]) GetConnection(
	session string,
	connection ConnectionInput,
	logger *slog.Logger,
) (T, error) {
	// First check if we need to create a new connection
	c.sessionMu.RLock()
	c.connMu.RLock()
	existingDb, exists := c.connMap[connection.GetId()]
	if exists {
		// Check if session binding exists
		_, sessionExists := c.sessionMap[session]
		c.sessionMu.RUnlock()
		c.connMu.RUnlock()

		if !sessionExists {
			return c.handleExistingConnection(session, connection.GetId(), existingDb)
		}
		return existingDb, nil
	}
	c.sessionMu.RUnlock()
	c.connMu.RUnlock()

	// Need write locks to create new connection
	c.sessionMu.Lock()
	c.connMu.Lock()
	defer c.connMu.Unlock()
	defer c.sessionMu.Unlock()

	// Check again under write lock to prevent duplicate creation
	if existingDb, exists := c.connMap[connection.GetId()]; exists {
		if _, ok := c.sessionMap[session]; !ok {
			c.sessionMap[session] = make(map[string]struct{})
		}
		c.sessionMap[session][connection.GetId()] = struct{}{}
		return existingDb, nil
	}

	// Create new connection
	connectionClient, err := c.connectionProvider.GetConnectionClient(connection.GetConnectionConfig())
	if err != nil {
		var result T
		return result, err
	}

	// Store new connection and bind session
	c.connMap[connection.GetId()] = connectionClient
	if _, ok := c.sessionMap[session]; !ok {
		c.sessionMap[session] = make(map[string]struct{})
	}
	c.sessionMap[session][connection.GetId()] = struct{}{}

	return connectionClient, nil
}

func (c *ConnectionManager[T]) handleExistingConnection(session, connId string, client T) (T, error) {
	c.sessionMu.Lock()
	if _, ok := c.sessionMap[session]; !ok {
		c.sessionMap[session] = make(map[string]struct{})
	}
	c.sessionMap[session][connId] = struct{}{}
	c.sessionMu.Unlock()
	return client, nil
}

func (c *ConnectionManager[T]) ReleaseSession(session string) bool {
	c.sessionMu.Lock()
	connMap, ok := c.sessionMap[session]
	if !ok || len(connMap) == 0 {
		c.sessionMu.Unlock()
		return false
	}

	sessionConnIds := getConnectionIds(connMap)

	delete(c.sessionMap, session)
	remainingConns := getUniqueConnectionIdsFromSessions(c.sessionMap)
	c.sessionMu.Unlock()

	if c.config.closeOnRelease {
		c.closeSpecificConnections(sessionConnIds, remainingConns)
	}
	return true
}

func (c *ConnectionManager[T]) closeSpecificConnections(candidateConnIds []string, remainingConns map[string]struct{}) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	for _, connId := range candidateConnIds {
		if _, stillInUse := remainingConns[connId]; !stillInUse {
			if dbConn, exists := c.connMap[connId]; exists {
				if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
					slog.Error(fmt.Sprintf("unable to close client connection during release: %s", err.Error()))
				}
				delete(c.connMap, connId)
			}
		}
	}
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
			c.cleanUnusedConnections()
		}
	}
}

func (c *ConnectionManager[T]) hardClose() {
	// Acquire locks in consistent order
	c.sessionMu.Lock()
	c.connMu.Lock()

	// Close all connections
	for connId, dbConn := range c.connMap {
		if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
			slog.Error(fmt.Sprintf("unable to fully close client connection during hard close: %s", err.Error()))
		}
		delete(c.connMap, connId)
	}

	// Clear all sessions
	for sessionId := range c.sessionMap {
		delete(c.sessionMap, sessionId)
	}

	c.connMu.Unlock()
	c.sessionMu.Unlock()
}

func (c *ConnectionManager[T]) cleanUnusedConnections() {
	// Get session info first
	c.sessionMu.RLock()
	sessionConnections := getUniqueConnectionIdsFromSessions(c.sessionMap)
	c.sessionMu.RUnlock()

	// Then handle connections
	c.connMu.Lock()
	for connId, dbConn := range c.connMap {
		if _, ok := sessionConnections[connId]; !ok {
			if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
				slog.Error(fmt.Sprintf("unable to fully close client connection during cleanup: %s", err.Error()))
			}
			delete(c.connMap, connId)
		}
	}
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
