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
	GetConnectionClient(connectionConfig *mgmtv1alpha1.ConnectionConfig, logger *slog.Logger) (T, error)
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

	isReaping bool
}

type ConnectionInput interface {
	GetId() string
	GetConnectionConfig() *mgmtv1alpha1.ConnectionConfig
}

type Interface[T any] interface {
	GetConnection(session SessionInterface, connection ConnectionInput, logger *slog.Logger) (T, error)
	ReleaseSession(session SessionInterface, logger *slog.Logger) bool
	Shutdown(logger *slog.Logger)
	Reaper(logger *slog.Logger)
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
		shutdown:           make(chan any, 1),
		mu:                 &sync.Mutex{},
		isReaping:          false,
	}
}

func (c *ConnectionManager[T]) GetConnection(
	session SessionInterface,
	connection ConnectionInput,
	logger *slog.Logger,
) (T, error) {
	groupId := session.Group()
	sessionId := session.Name()

	logger = logger.With("session", session.String())

	c.mu.Lock()
	defer c.mu.Unlock()

	if groupConns, exists := c.groupConnMap[groupId]; exists {
		if existingDb, exists := groupConns[connection.GetId()]; exists {
			logger.Debug("found existing connection for the session group")
			c.ensureSessionMapsExist(groupId, sessionId)
			c.groupSessionMap[groupId][sessionId][connection.GetId()] = struct{}{}
			return existingDb, nil
		}
	}

	logger.Debug("no cached connection found, creating new connection client")

	// Create new connection
	connectionClient, err := c.connectionProvider.GetConnectionClient(connection.GetConnectionConfig(), logger)
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

func (c *ConnectionManager[T]) ReleaseSession(session SessionInterface, logger *slog.Logger) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	groupId := session.Group()
	sessionId := session.Name()
	logger = logger.With("session", session.String())
	groupSessions, groupExists := c.groupSessionMap[groupId]
	if !groupExists {
		logger.Debug("session group not found during release")
		return false
	}

	sessionConns, sessionExists := groupSessions[sessionId]
	if !sessionExists || len(sessionConns) == 0 {
		logger.Debug("session not found in group during release")
		return false
	}

	sessionConnIds := getConnectionIds(sessionConns)
	delete(groupSessions, sessionId)

	// If this was the last session in the group, clean up the group
	if len(groupSessions) == 0 {
		logger.Debug("cleaning up group, last session found")
		delete(c.groupSessionMap, groupId)
	}

	if c.config.closeOnRelease {
		logger.Debug("close on release is enabled, pruning connections that are not bound to any sessions in the group")
		remainingConns := getUniqueConnectionIdsFromGroupSessions(groupSessions)
		c.closeSpecificGroupConnections(groupId, sessionConnIds, remainingConns, logger)
	}
	return true
}

// does not handle locks as it assumes the parent caller holds the lock
func (c *ConnectionManager[T]) closeSpecificGroupConnections(groupId string, candidateConnIds []string, remainingConns map[string]struct{}, logger *slog.Logger) {
	groupConns, exists := c.groupConnMap[groupId]
	if !exists {
		return
	}

	for _, connId := range candidateConnIds {
		if _, stillInUse := remainingConns[connId]; !stillInUse {
			if dbConn, exists := groupConns[connId]; exists {
				logger.Debug(fmt.Sprintf("closing connection %q", connId))
				if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
					logger.Error(fmt.Sprintf("unable to close client connection during release: %s", err.Error()))
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

func (c *ConnectionManager[T]) Shutdown(logger *slog.Logger) {
	c.shutdown <- struct{}{}
	if !c.isReaping {
		logger.Debug("reaper is not turned on, hard closing")
		c.hardClose(logger)
	} else {
		logger.Debug("sent shutdown signal to reaper")
	}
}

func (c *ConnectionManager[T]) Reaper(logger *slog.Logger) {
	for {
		select {
		case <-c.shutdown:
			c.hardClose(logger)
			return
		case <-time.After(c.config.reapDuration):
			c.cleanUnusedConnections(logger)
		}
	}
}

func (c *ConnectionManager[T]) cleanUnusedConnections(logger *slog.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	groupSessionConnections := make(map[string]map[string]struct{})
	for groupId, sessions := range c.groupSessionMap {
		groupSessionConnections[groupId] = getUniqueConnectionIdsFromGroupSessions(sessions)

		groupSessions := []string{}
		for session := range sessions {
			groupSessions = append(groupSessions, session)
		}
		logger.Debug(fmt.Sprintf("[ConnectionManager][Reaper] group %q with sessions %s", groupId, strings.Join(groupSessions, ",")))
	}

	for groupId, groupConns := range c.groupConnMap {
		logger.Debug(fmt.Sprintf("[ConnectionManager][Reaper] checking group %q with %d connection(s)", groupId, len(groupConns)))
		sessionConns := groupSessionConnections[groupId]
		for connId, dbConn := range groupConns {
			logger.Debug(fmt.Sprintf("[ConnectionManager][Reaper] checking group %q for connection %q", groupId, connId))
			if _, ok := sessionConns[connId]; !ok {
				logger.Debug(fmt.Sprintf("[ConnectionManager][Reaper] closing client connection: %q in group %q", connId, groupId))
				if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
					logger.Warn(fmt.Sprintf("[ConnectionManager][Reaper] unable to fully close client connection %q in group %q during cleanup: %s", connId, groupId, err.Error()))
				}
				delete(groupConns, connId)
			}
		}
		if len(groupConns) == 0 {
			delete(c.groupConnMap, groupId)
		}
	}
}

func (c *ConnectionManager[T]) hardClose(logger *slog.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all connections in all groups
	for groupId, groupConns := range c.groupConnMap {
		for connId, dbConn := range groupConns {
			if err := c.connectionProvider.CloseClientConnection(dbConn); err != nil {
				logger.Error(fmt.Sprintf("unable to fully close client connection during hard close: %s", err.Error()))
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
