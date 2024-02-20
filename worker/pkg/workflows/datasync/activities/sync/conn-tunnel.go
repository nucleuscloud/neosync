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
	connMap     sync.Map
	connCounter sync.Map
}

func (c *ConnectionTunnelManager) GetConnectionString(
	connection *mgmtv1alpha1.Connection,
	logger *slog.Logger,
) (string, error) {
	loadedDetails, ok := c.load(connection.Id)
	if ok {
		return loadedDetails.GeneralDbConnectConfig.String(), nil
	}

	details, err := sqlconnect.GetConnectionDetails(connection.ConnectionConfig, shared.Ptr(uint32(5)), logger)
	if err != nil {
		return "", err
	}
	if details.Tunnel == nil {
		return details.GeneralDbConnectConfig.String(), nil
	}
	ready, err := details.Tunnel.Start()
	if err != nil {
		return "", fmt.Errorf("unable to start ssh tunnel: %w", err)
	}
	<-ready
	details.GeneralDbConnectConfig.Host = details.Tunnel.Local.Host
	details.GeneralDbConnectConfig.Port = int32(details.Tunnel.Local.Port)
	c.connMap.Store(connection.Id, details)
	logger.Debug(
		"ssh tunnel is ready, updated configuration host and port",
		"host", details.Tunnel.Local.Host,
		"port", details.Tunnel.Local.Port,
	)
	go c.reaper(connection.Id)
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
	count, ok := c.getCurrentCount(connectionId)
	if !ok {
		count = 1
	} else {
		count += 1
	}
	c.connCounter.Store(connectionId, count)
}
func (c *ConnectionTunnelManager) decrementCount(connectionId string) {
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

func (c *ConnectionTunnelManager) reaper(connectionId string) {
	for {
		select {
		case <-time.After(1 * time.Minute):
			count, ok := c.getCurrentCount(connectionId)
			if !ok || count <= 0 {
				c.close(connectionId)
				return
			}
			// connection is still open and active, wait more time
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
