package mysqltunconnector

import (
	"context"
	"crypto/tls"
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

type Option func(*connectorConfig) error

type connectorConfig struct {
	dialer    sshtunnel.ContextDialer
	tlsConfig *tls.Config
}

// WithDialer sets a custom dialer for the connector
func WithDialer(dialer sshtunnel.ContextDialer) Option {
	return func(cfg *connectorConfig) error {
		cfg.dialer = dialer
		return nil
	}
}

// WithTLSConfig sets TLS configuration for the connector
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *connectorConfig) error {
		cfg.tlsConfig = tlsConfig
		return nil
	}
}

func New(
	dsn string,
	opts ...Option,
) (*Connector, func(), error) {
	cfg := &connectorConfig{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, nil, err
		}
	}

	mysqlCfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse mysql dsn: %w", err)
	}

	ogNetwork := mysqlCfg.Net
	newNetwork := buildUniqueNetwork(ogNetwork)

	mysqlCfg.Net = newNetwork

	var dialer sshtunnel.ContextDialer
	if cfg.dialer != nil {
		dialer = cfg.dialer
	} else {
		dialer = &net.Dialer{}
	}

	mysql.RegisterDialContext(newNetwork, func(ctx context.Context, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, ogNetwork, addr)
	})

	if cfg.tlsConfig != nil {
		mysqlCfg.TLS = cfg.tlsConfig
	}

	conn, err := mysql.NewConnector(mysqlCfg)
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
