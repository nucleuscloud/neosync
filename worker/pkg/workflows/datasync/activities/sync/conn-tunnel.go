package sync_activity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func NewConnectionTunnelManager() *ConnectionTunnelManager {
	return &ConnectionTunnelManager{
		connMap:        map[string]*sql.DB{},
		sessionMap:     map[string]map[string]struct{}{},
		connDetailsMap: map[string]*sqlconnect.ConnectionDetails{},
		shutdown:       make(chan any),
	}
}

type ConnectionTunnelManager struct {
	// connection id to connection details
	connDetailsMap map[string]*sqlconnect.ConnectionDetails
	connDetailsMu  sync.RWMutex

	// session
	sessionMap map[string]map[string]struct{}
	sessionMu  sync.RWMutex

	// connection id to sql connection
	connMap map[string]*sql.DB
	connMu  sync.RWMutex

	shutdown chan any
}

func (c *ConnectionTunnelManager) LogStats(logger *slog.Logger, level slog.Level) {
	logger.Log(
		context.Background(),
		level,
		"connection manager stats",
		"liveConnections", len(c.connMap),
	)
}

// Purpose of this function is to return a connection string that can be used by
// a database client to connect to a tunneled instance of a database using a localhost port.
// Primarily used by Benthos since we don't have the ability to directly use a pg client
func (c *ConnectionTunnelManager) GetConnectionString(
	session string,
	connection *mgmtv1alpha1.Connection,
	logger *slog.Logger,
) (string, error) {
	c.connDetailsMu.RLock()
	loadedDetails, ok := c.connDetailsMap[connection.Id]

	if ok {
		c.bindSession(session, connection.Id)
		c.connDetailsMu.RUnlock()
		return loadedDetails.GeneralDbConnectConfig.String(), nil
	}
	c.connDetailsMu.RUnlock()
	c.connDetailsMu.Lock()
	defer c.connDetailsMu.Unlock()

	loadedDetails, ok = c.connDetailsMap[connection.Id]
	if ok {
		c.bindSession(session, connection.Id)
		return loadedDetails.GeneralDbConnectConfig.String(), nil
	}

	details, err := sqlconnect.GetConnectionDetails(connection.ConnectionConfig, shared.Ptr(uint32(5)), logger)
	if err != nil {
		return "", err
	}
	if details.Tunnel == nil {
		c.bindSession(session, connection.Id)
		return details.GeneralDbConnectConfig.String(), nil
	}
	ready, err := details.Tunnel.Start(logger)
	if err != nil {
		return "", fmt.Errorf("unable to start ssh tunnel: %w", err)
	}
	<-ready
	localhost, localport := details.Tunnel.GetLocalHostPort()
	details.GeneralDbConnectConfig.Host = localhost
	details.GeneralDbConnectConfig.Port = int32(localport)
	logger.Debug(
		"ssh tunnel is ready, updated configuration host and port",
		"host", localhost,
		"port", localport,
	)
	c.bindSession(session, connection.Id)
	return details.GeneralDbConnectConfig.String(), nil
}

func (c *ConnectionTunnelManager) GetConnection(
	session string,
	connection *mgmtv1alpha1.Connection,
	logger *slog.Logger,
) (*sql.DB, error) {
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

	connectionString, err := c.GetConnectionString(session, connection, logger)
	if err != nil {
		return nil, err
	}
	driver, err := getDriverFromConnection(connection)
	if err != nil {
		return nil, err
	}

	dbconn, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}
	c.connMap[connection.Id] = dbconn
	c.bindSession(session, connection.Id)
	return dbconn, nil
}

func (c *ConnectionTunnelManager) ReleaseSession(session string) {
	c.sessionMu.RLock()
	connMap, ok := c.sessionMap[session]
	if !ok || len(connMap) == 0 {
		c.sessionMu.RUnlock()
		return
	}
	c.sessionMu.RUnlock()
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	connMap, ok = c.sessionMap[session]
	if !ok || len(connMap) == 0 {
		return
	}
	delete(c.sessionMap, session)
}

func (c *ConnectionTunnelManager) bindSession(session, connectionId string) {
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

func (c *ConnectionTunnelManager) Shutdown() {
	c.shutdown <- struct{}{}
}

func (c *ConnectionTunnelManager) Reaper() {
	for {
		select {
		case <-c.shutdown:
			c.close()
			return
		case <-time.After(1 * time.Minute):
			c.close()
		}
	}
}

func (c *ConnectionTunnelManager) close() {
	c.connMu.Lock()
	c.sessionMu.Lock()
	sessionConnections := getUniqueConnectionIdsFromSessions(c.sessionMap)
	for connId, dbConn := range c.connMap {
		if _, ok := sessionConnections[connId]; !ok {
			dbConn.Close()
			delete(c.connMap, connId)
		}
	}
	c.sessionMu.Unlock()
	c.connMu.Unlock()

	c.connDetailsMu.Lock()
	c.sessionMu.Lock()
	sessionConnections = getUniqueConnectionIdsFromSessions(c.sessionMap)
	for connId, details := range c.connDetailsMap {
		if _, ok := sessionConnections[connId]; !ok {
			if details.Tunnel != nil {
				details.Tunnel.Close()
			}
			delete(c.connDetailsMap, connId)
		}
	}
	c.sessionMu.Unlock()
	c.connDetailsMu.Unlock()
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

func getDriverFromConnection(connection *mgmtv1alpha1.Connection) (string, error) {
	if connection == nil {
		return "", errors.New("connection was nil")
	}
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return "mysql", nil
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return "postgres", nil
	}
	return "", errors.New("unsupported connection type when computing driver")
}
