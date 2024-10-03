package mysqltunconnector

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/nucleuscloud/neosync/internal/sshtunnel"
)

type Connector struct {
	driver.Connector
}

var _ driver.Connector = (*Connector)(nil)

func New(dialer sshtunnel.Dialer, dsn string) (*Connector, func(), error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse mysql dsn: %w", err)
	}

	ogNetwork := cfg.Net
	newNetwork := buildUniqueNetwork(ogNetwork)

	cfg.Net = newNetwork
	mysql.RegisterDialContext(cfg.Net, func(ctx context.Context, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, ogNetwork, addr)
	})

	conn, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		mysql.DeregisterDialContext(newNetwork)
	}

	return &Connector{Connector: conn}, cleanup, nil
}

func buildUniqueNetwork(network string) string {
	return fmt.Sprintf("%s_%s", network, getUniqueIdentifier())
}

func getUniqueIdentifier() string {
	id := uuid.NewString()
	return strings.ReplaceAll(id, "-", "")
}
