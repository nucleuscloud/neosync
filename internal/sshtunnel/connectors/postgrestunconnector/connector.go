package postgrestunconnector

import (
	"context"
	"crypto/tls"
	"database/sql/driver"
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	pgxslog "github.com/nucleuscloud/neosync/internal/pgx-slog"
	"github.com/nucleuscloud/neosync/internal/sshtunnel"
)

type Connector struct {
	connStr string
	driver  driver.Driver
}

var _ driver.Connector = (*Connector)(nil)

type Option func(*connectorConfig) error

type connectorConfig struct {
	dialer    sshtunnel.ContextDialer
	tlsConfig *tls.Config
	logger    *slog.Logger
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

func WithLogger(logger *slog.Logger) Option {
	return func(cfg *connectorConfig) error {
		cfg.logger = logger
		return nil
	}
}

func New(
	dsn string,
	opts ...Option,
) (*Connector, func(), error) {
	cfg := &connectorConfig{
		logger: slog.Default(),
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, nil, err
		}
	}

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, nil, err
	}

	if cfg.dialer != nil {
		pgxConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return cfg.dialer.DialContext(ctx, network, addr)
		}
	}
	if cfg.tlsConfig != nil {
		pgxConfig.TLSConfig = cfg.tlsConfig
	}

	pgxConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgxslog.NewLogger(cfg.logger, pgxslog.GetShouldOmitArgs()),
		LogLevel: pgxslog.GetDatabaseLogLevel(),
	}
	// todo: We may need to re-enable this to support pg bouncer
	// pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	// RegisterConnConfig returns unique connection strings, so even if the dsn is used for multiple calls to New()
	// The unregister will not interfere with any other instances of Connector that are using the same input dsn
	connStr := stdlib.RegisterConnConfig(pgxConfig)
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
