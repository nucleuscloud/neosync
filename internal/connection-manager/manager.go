package connectionmanager

import (
	"fmt"
	"log/slog"
	"strings"
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
	config             *managerConfig

	mu *sync.Mutex

	// Map of groupId -> sessionId -> connectionIds
	groupSessionMap map[string]map[string]map[string]struct{}

	// Map of groupId -> connectionId -> connection
	groupConnMap map[string]map[string]T

	shutdown chan any
}

type ConnectionInput interface {
	GetId() string
	GetConnectionConfig() *mgmtv1alpha1.ConnectionConfig
}

type Interface[T any] interface {
	GetConnection(session SessionInterface, connection ConnectionInput, logger *slog.Logger) (T, error)
	ReleaseSession(session SessionInterface) bool
	ReleaseSessionGroup(grouper SessionGroupInterface) bool
	Shutdown()
	Reaper()
}

var _ Interface[any] = &ConnectionManager[any]{}

type managerConfig struct {
	closeOnRelease bool
	reapDuration   time.Duration
}

// When a session is closed, if no other sessions within that group exist, the connection will be closed.
// Otherwise, the connection will not be closed until Shutdown() or a Repear() cycle occurs.
func WithCloseOnRelease() ManagerOption {
	return func(mc *managerConfig) {
		mc.closeOnRelease = true
	}
}

func WithReaperPoll(duration time.Duration) ManagerOption {
	return func(mc *managerConfig) {
		mc.reapDuration = duration
	}
}

type ManagerOption func(*managerConfig)

func NewConnectionManager[T any](
	connectionProvider ConnectionProvider[T],
	opts ...ManagerOption,
) *ConnectionManager[T] {
	cfg := &managerConfig{
		reapDuration: 1 * time.Minute,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &ConnectionManager[T]{
		connectionProvider: connectionProvider,
		groupSessionMap:    map[string]map[string]map[string]struct{}{},
		groupConnMap:       map[string]map[string]T{},
		config:             cfg,
		shutdown:           make(chan any),
		mu:                 &sync.Mutex{},
	}
}

func (c *ConnectionManager[T]) GetConnection(
	session SessionInterface,
	connection ConnectionInput,
	logger *slog.Logger,
) (T, error) {
	groupId := session.Group()
	sessionId := session.Name()

	c.mu.Lock()
	defer c.mu.Unlock()

	if groupConns, exists := c.groupConnMap[groupId]; exists {
		if existingDb, exists := groupConns[connection.GetId()]; exists {
			c.ensureSessionMapsExist(groupId, sessionId)
			c.groupSessionMap[groupId][sessionId][connection.GetId()] = struct{}{}
			return existingDb, nil
		}
	}

	// Create new connection
	connectionClient, err := c.connectionProvider.GetConnectionClient(connection.GetConnectionConfig())
	if err != nil {
		var result T
		return result, err
	}

	// Initialize maps if they don't exist
	c.ensureSessionMapsExist(groupId, sessionId)
	if _, ok := c.groupConnMap[groupId]; !ok {
		c.groupConnMap[groupId] = make(map[string]T)
	}

	// Store new connection and bind session
	c.groupConnMap[groupId][connection.GetId()] = connectionClient
	c.groupSessionMap[groupId][sessionId][connection.GetId()] = struct{}{}

	return connectionClient, nil
}

func (c *ConnectionManager[T]) ensureSessionMapsExist(groupId, sessionId string) {
	if _, ok := c.groupSessionMap[groupId]; !ok {
		c.groupSessionMap[groupId] = make(map[string]map[string]struct{})
	}
	if _, ok := c.groupSessionMap[groupId][sessionId]; !ok {
		c.groupSessionMap[groupId][sessionId] = make(map[string]struct{})
	}
}

func (c *ConnectionManager[T]) ReleaseSession(session SessionInterface) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	groupId := session.Group()
	sessionId := session.Name()
	groupSessions, groupExists := c.groupSessionMap[groupId]
	if !groupExists {
		return false
	}

	sessionConns, sessionExists := groupSessions[sessionId]
	if !sessionExists || len(sessionConns) == 0 {
		return false
	}

	sessionConnIds := getConnectionIds(sessionConns)
	delete(groupSessions, sessionId)

	// If this was the last session in the group, clean up the group
	if len(groupSessions) == 0 {
		delete(c.groupSessionMap, groupId)
	}

	if c.config.closeOnRelease {
		remainingConns := getUniqueConnectionIdsFromGroupSessions(groupSessions)
		c.closeSpecificGroupConnections(groupId, sessionConnIds, remainingConns)
	}
	return true
}

func (c *ConnectionManager[T]) ReleaseSessionGroup(grouper SessionGroupInterface) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	groupId := grouper.Group()

	groupSessions, groupExists := c.groupSessionMap[groupId]
	if !groupExists || len(groupSessions) == 0 {
		return false
	}

	// Get all connection IDs that are in the group
	connIds := make([]string, 0)
	for _, sessionConns := range groupSessions {
		connIds = append(connIds, getConnectionIds(sessionConns)...)
	}

	// Remove all sessions in the group
	delete(c.groupSessionMap, groupId)

	if c.config.closeOnRelease {
		// Since we're removing the entire group, there are no remaining connections
		c.closeSpecificGroupConnections(groupId, connIds, make(map[string]struct{}))
	}

	return len(connIds) > 0
}

// does not handle locks as it assumes the parent caller holds the lock
func (c *ConnectionManager[T]) closeSpecificGroupConnections(groupId string, candidateConnIds []string, remainingConns map[string]struct{}) {
	groupConns, exists := c.groupConnMap[groupId]
	if !exists {
		return
	}

	for _, connId := range candidateConnIds {
		if _, stillInUse := remainingConns[connId]; !stillInUse {
			if dbConn, exists := groupConns[connId]; exists {
				if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
					slog.Error(fmt.Sprintf("unable to close client connection during release: %s", err.Error()))
				}
				delete(groupConns, connId)
			}
		}
	}

	// If this was the last connection in the group, clean up the group
	if len(groupConns) == 0 {
		delete(c.groupConnMap, groupId)
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
		case <-time.After(c.config.reapDuration):
			c.cleanUnusedConnections()
		}
	}
}

func (c *ConnectionManager[T]) cleanUnusedConnections() {
	c.mu.Lock()
	defer c.mu.Unlock()

	groupSessionConnections := make(map[string]map[string]struct{})
	for groupId, sessions := range c.groupSessionMap {
		groupSessionConnections[groupId] = getUniqueConnectionIdsFromGroupSessions(sessions)

		groupSessions := []string{}
		for session := range sessions {
			groupSessions = append(groupSessions, session)
		}
		slog.Debug(fmt.Sprintf("[ConnectionManager][Reaper] group %q with sessions %s", groupId, strings.Join(groupSessions, ",")))
	}

	for groupId, groupConns := range c.groupConnMap {
		slog.Debug(fmt.Sprintf("[ConnectionManager][Reaper] checking group %q with %d connection(s)", groupId, len(groupConns)))
		sessionConns := groupSessionConnections[groupId]
		for connId, dbConn := range groupConns {
			slog.Debug(fmt.Sprintf("[ConnectionManager][Reaper] checking group %q for connection %q", groupId, connId))
			if _, ok := sessionConns[connId]; !ok {
				slog.Debug(fmt.Sprintf("[ConnectionManager][Reaper] closing client connection: %q in group %q", connId, groupId))
				if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
					slog.Warn(fmt.Sprintf("[ConnectionManager][Reaper] unable to fully close client connection %q in group %q during cleanup: %s", connId, groupId, err.Error()))
				}
				delete(groupConns, connId)
			}
		}
		if len(groupConns) == 0 {
			delete(c.groupConnMap, groupId)
		}
	}
}

func (c *ConnectionManager[T]) hardClose() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all connections in all groups
	for groupId, groupConns := range c.groupConnMap {
		for connId, dbConn := range groupConns {
			if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
				slog.Error(fmt.Sprintf("unable to fully close client connection during hard close: %s", err.Error()))
			}
			delete(groupConns, connId)
		}
		delete(c.groupConnMap, groupId)
	}

	// Clear all sessions in all groups
	for groupId := range c.groupSessionMap {
		delete(c.groupSessionMap, groupId)
	}
}

func getUniqueConnectionIdsFromGroupSessions(sessions map[string]map[string]struct{}) map[string]struct{} {
	connSet := map[string]struct{}{}
	for _, sessionConns := range sessions {
		for connId := range sessionConns {
			connSet[connId] = struct{}{}
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
