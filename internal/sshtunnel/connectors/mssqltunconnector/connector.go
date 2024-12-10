package mssqltunconnector

import (
	"crypto/tls"
	"database/sql/driver"
	"fmt"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
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

	params, err := msdsn.Parse(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse mssql dsn as valid dsn: %w", err)
	}

	if cfg.tlsConfig != nil {
		params.TLSConfig = cfg.tlsConfig
	}

	connector := mssql.NewConnectorConfig(params)

	if cfg.dialer != nil {
		connector.Dialer = mssql.Dialer(cfg.dialer)
	}

	return &Connector{Connector: connector}, func() {}, nil
}
