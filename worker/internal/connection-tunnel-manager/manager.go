package connectiontunnelmanager

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type ConnectionProvider[T any, TConfig any] interface {
	GetConnectionDetails(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (ConnectionDetails, error)
	GetConnectionClient(driver string, connectionString string, opts TConfig) (T, error)
	GetConnectionClientConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig) (TConfig, error)
	CloseClientConnection(client T) error
}

type ConnectionTunnelManager[T any, TConfig any] struct {
	connectionProvider ConnectionProvider[T, TConfig]

	connDetailsMap map[string]ConnectionDetails
	connDetailsMu  sync.RWMutex

	sessionMap map[string]map[string]struct{}
	sessionMu  sync.RWMutex

	connMap map[string]T
	connMu  sync.RWMutex

	shutdown chan any
}

// type ConnectionDetails struct {
// 	GeneralDbConnectConfig sqlconnect.GeneralDbConnectConfig

// 	Tunnel *sshtunnel.Sshtunnel
// }

type ConnectionDetails interface {
	String() string
	GetTunnel() *sshtunnel.Sshtunnel
}

type Interface[T any] interface {
	GetConnectionString(session string, connection *mgmtv1alpha1.Connection, logger *slog.Logger) (string, error)
	GetConnection(session string, connection *mgmtv1alpha1.Connection, logger *slog.Logger) (T, error)

	ReleaseSession(session string) bool
	Shutdown()
	Reaper()
}

var _ Interface[any] = &ConnectionTunnelManager[any, any]{} // enforces ConnectionTunnelManager always conforms to the interface

func NewConnectionTunnelManager[T any, TConfig any](connectionProvider ConnectionProvider[T, TConfig]) *ConnectionTunnelManager[T, TConfig] {
	return &ConnectionTunnelManager[T, TConfig]{
		connectionProvider: connectionProvider,
		sessionMap:         map[string]map[string]struct{}{},
		connDetailsMap:     map[string]ConnectionDetails{},
		connMap:            map[string]T{},
	}
}

func (c *ConnectionTunnelManager[T, TConfig]) GetConnectionString(
	session string,
	connection *mgmtv1alpha1.Connection,
	logger *slog.Logger,
) (string, error) {
	c.connDetailsMu.RLock()
	loadedDetails, ok := c.connDetailsMap[connection.Id]

	if ok {
		c.bindSession(session, connection.Id)
		c.connDetailsMu.RUnlock()
		return loadedDetails.String(), nil
		// return loadedDetails.GeneralDbConnectConfig.String(), nil
	}
	c.connDetailsMu.RUnlock()
	c.connDetailsMu.Lock()
	defer c.connDetailsMu.Unlock()

	loadedDetails, ok = c.connDetailsMap[connection.Id]
	if ok {
		c.bindSession(session, connection.Id)
		return loadedDetails.String(), nil
		// return loadedDetails.GeneralDbConnectConfig.String(), nil
	}

	details, err := c.connectionProvider.GetConnectionDetails(connection.ConnectionConfig, shared.Ptr(uint32(5)), logger)
	if err != nil {
		return "", err
	}
	tunnel := details.GetTunnel()
	if tunnel == nil {
		c.bindSession(session, connection.Id)
		c.connDetailsMap[connection.Id] = details
		return details.String(), nil
	}
	ready, err := tunnel.Start(logger)
	if err != nil {
		return "", fmt.Errorf("unable to start ssh tunnel: %w", err)
	}
	<-ready // this isn't great as it will block all other requests until this tunnel is ready
	// localhost, localport := tunnel.GetLocalHostPort()
	// details.GeneralDbConnectConfig.Host = localhost
	// details.GeneralDbConnectConfig.Port = int32(localport)
	// logger.Debug(
	// 	"ssh tunnel is ready, updated configuration host and port",
	// 	"host", localhost,
	// 	"port", localport,
	// )
	c.connDetailsMap[connection.Id] = details
	c.bindSession(session, connection.Id)
	return details.String(), nil
}

func (c *ConnectionTunnelManager[T, TConfig]) GetConnection(
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

	connectionString, err := c.GetConnectionString(session, connection, logger)
	if err != nil {
		var result T
		return result, err
	}
	driver, err := getDriverFromConnection(connection)
	if err != nil {
		var result T
		return result, err
	}

	connClientConfig, err := c.connectionProvider.GetConnectionClientConfig(connection.GetConnectionConfig())
	if err != nil {
		var result T
		return result, err
	}

	connectionClient, err := c.connectionProvider.GetConnectionClient(driver, connectionString, connClientConfig)
	if err != nil {
		var result T
		return result, err
	}

	c.connMap[connection.Id] = connectionClient
	c.bindSession(session, connection.Id)
	return connectionClient, nil
}

func (c *ConnectionTunnelManager[T, TConfig]) ReleaseSession(session string) bool {
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

func (c *ConnectionTunnelManager[T, TConfig]) bindSession(session, connectionId string) {
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

func (c *ConnectionTunnelManager[T, TConfig]) Shutdown() {
	c.shutdown <- struct{}{}
}

func (c *ConnectionTunnelManager[T, TConfig]) Reaper() {
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

func (c *ConnectionTunnelManager[T, TConfig]) hardClose() {
	c.connMu.Lock()
	c.connDetailsMu.Lock()
	c.sessionMu.Lock()
	for connId, dbConn := range c.connMap {
		err := c.connectionProvider.CloseClientConnection(dbConn)
		if err != nil {
			slog.Error(fmt.Sprintf("unable to fully close client connection during hard close: %s", err.Error()))
		}
		delete(c.connMap, connId)
	}

	for connId, details := range c.connDetailsMap {
		tunnel := details.GetTunnel()
		if tunnel != nil {
			tunnel.Close()
		}
		delete(c.connDetailsMap, connId)
	}

	for sessionId := range c.sessionMap {
		delete(c.sessionMap, sessionId)
	}
	c.connMu.Unlock()
	c.connDetailsMu.Unlock()
	c.sessionMu.Unlock()
}

func (c *ConnectionTunnelManager[T, TConfig]) close() {
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

	c.connDetailsMu.Lock()
	c.sessionMu.Lock()
	sessionConnections = getUniqueConnectionIdsFromSessions(c.sessionMap)
	for connId, details := range c.connDetailsMap {
		if _, ok := sessionConnections[connId]; !ok {
			tunnel := details.GetTunnel()
			if tunnel != nil {
				tunnel.Close()
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
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return "mongodb", nil
	}
	return "", errors.New("unsupported connection type when computing driver")
}
