package sync_activity

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func NewConnectionTunnelManager() *ConnectionTunnelManager {
	return &ConnectionTunnelManager{}
}

type ConnectionTunnelManager struct {
	connMap       sync.Map
	connLockMxMap sync.Map

	connCounter    sync.Map
	counterLockMap sync.Map
}

// Purpose of this function is to return a connection string that can be used by
// a database client to connect to a tunneled instance of a database using a localhost port.
// Primarily used by Benthos since we don't have the ability to directly use a pg client
func (c *ConnectionTunnelManager) GetConnectionString(
	connection *mgmtv1alpha1.Connection,
	logger *slog.Logger,
) (string, error) {
	loadedDetails, ok := c.load(connection.Id)
	if ok {
		c.incrementCount(connection.Id)
		return loadedDetails.GeneralDbConnectConfig.String(), nil
	}

	mu, _ := c.connLockMxMap.LoadOrStore(connection.Id, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	loadDetails, ok := c.load(connection.Id)
	if ok {
		c.incrementCount(connection.Id)
		return loadDetails.GeneralDbConnectConfig.String(), nil
	}

	details, err := sqlconnect.GetConnectionDetails(connection.ConnectionConfig, shared.Ptr(uint32(5)), logger)
	if err != nil {
		return "", err
	}
	if details.Tunnel == nil {
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
	c.connMap.Store(connection.Id, details)
	c.incrementCount(connection.Id)
	logger.Debug(
		"ssh tunnel is ready, updated configuration host and port",
		"host", localhost,
		"port", localport,
	)
	go c.reaper(connection.Id, 1*time.Minute)
	return details.GeneralDbConnectConfig.String(), nil
}

func (c *ConnectionTunnelManager) load(connectionId string) (*sqlconnect.ConnectionDetails, bool) {
	t, ok := c.connMap.Load(connectionId)
	if !ok {
		return nil, false
	}
	tunnel, ok := t.(*sqlconnect.ConnectionDetails)
	return tunnel, ok
}

func (c *ConnectionTunnelManager) Release(connectionId string) {
	c.decrementCount(connectionId)
}

func (c *ConnectionTunnelManager) incrementCount(connectionId string) {
	mu, _ := c.counterLockMap.LoadOrStore(connectionId, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	count, ok := c.getCurrentCount(connectionId)
	if !ok {
		count = 1
	} else {
		count += 1
	}
	c.connCounter.Store(connectionId, count)
}
func (c *ConnectionTunnelManager) decrementCount(connectionId string) {
	mu, _ := c.counterLockMap.LoadOrStore(connectionId, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	count, ok := c.getCurrentCount(connectionId)
	if !ok {
		count = 0
	} else {
		count -= 1
	}
	c.connCounter.Store(connectionId, count)
}

func (c *ConnectionTunnelManager) getCurrentCount(connectionId string) (int, bool) {
	val, ok := c.connCounter.Load(connectionId)
	if !ok {
		return 0, false
	}
	count, ok := val.(int)
	return count, ok
}

func (c *ConnectionTunnelManager) reaper(connectionId string, initialDelay time.Duration) {
	time.Sleep(initialDelay)
	for {
		time.Sleep(1 * time.Minute)
		count, ok := c.getCurrentCount(connectionId)
		if !ok || count <= 0 {
			c.close(connectionId)
			return
		}
	}
}

func (c *ConnectionTunnelManager) close(connectionId string) {
	details, ok := c.load(connectionId)
	if !ok {
		return
	}
	c.connMap.Delete(connectionId)
	c.connCounter.Delete(connectionId)
	if details.Tunnel != nil {
		details.Tunnel.Close()
	}
}
