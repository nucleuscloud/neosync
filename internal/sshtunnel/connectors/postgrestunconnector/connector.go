package postgrestunconnector

import (
	"context"
	"database/sql/driver"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nucleuscloud/neosync/internal/sshtunnel"
)

type Connector struct {
	connStr string
	driver  driver.Driver
}

var _ driver.Connector = (*Connector)(nil)

func New(
	dialer sshtunnel.Dialer,
	dsn string,
) (*Connector, func(), error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, nil, err
	}
	cfg.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	connStr := stdlib.RegisterConnConfig(cfg)
	cleanup := func() {
		stdlib.UnregisterConnConfig(connStr)
	}

	return &Connector{connStr: connStr, driver: stdlib.GetDefaultDriver()}, cleanup, nil
}

func (c *Connector) Connect(_ context.Context) (driver.Conn, error) {
	return c.driver.Open(c.connStr)
}

func (c *Connector) Driver() driver.Driver {
	return c.driver
}
