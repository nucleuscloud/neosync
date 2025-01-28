package sqlconnect

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	dbconnectconfig "github.com/nucleuscloud/neosync/backend/pkg/dbconnect-config"
	"github.com/nucleuscloud/neosync/backend/pkg/sqldbtx"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlretry"
	tun "github.com/nucleuscloud/neosync/internal/sshtunnel"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mssqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mysqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/postgrestunconnector"
	"golang.org/x/crypto/ssh"
)

// interface used by SqlConnector to abstract away the opening and closing of a sqldb that includes tunnelingff
type SqlDbContainer interface {
	Open() (sqldbtx.DBTX, error)
	Close() error
}

type SqlConnectorOption func(*sqlConnectorOptions)

type sqlConnectorOptions struct {
	mysqlDisableParseTime bool

	connectionTimeoutSeconds *uint32
}

// WithMysqlParseTimeDisabled disables MySQL time parsing
func WithMysqlParseTimeDisabled() SqlConnectorOption {
	return func(opts *sqlConnectorOptions) {
		opts.mysqlDisableParseTime = true
	}
}

// Provide an integer number that corresponds to the number of seconds to wait before timing out attempting to connect.
// Ex: 10 == 10 seconds
func WithConnectionTimeout(timeoutSeconds uint32) SqlConnectorOption {
	return func(sco *sqlConnectorOptions) {
		sco.connectionTimeoutSeconds = &timeoutSeconds
	}
}

type SqlConnector interface {
	NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, logger *slog.Logger, opts ...SqlConnectorOption) (SqlDbContainer, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) NewDbFromConnectionConfig(cc *mgmtv1alpha1.ConnectionConfig, logger *slog.Logger, opts ...SqlConnectorOption) (SqlDbContainer, error) {
	if cc == nil {
		return nil, errors.New("connectionConfig was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}

	options := sqlConnectorOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	dbconnopts, err := getConnectionOptsFromConnectionConfig(cc)
	if err != nil {
		return nil, err
	}

	switch config := cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		connDetails, err := dbconnectconfig.NewFromPostgresConnection(config, options.connectionTimeoutSeconds, logger)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		return newStdlibConnectorContainer(
			getPgConnectorFn(dsn, config.PgConfig, logger),
			dbconnopts,
			logger,
		), nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		connDetails, err := dbconnectconfig.NewFromMysqlConnection(config, options.connectionTimeoutSeconds, logger, options.mysqlDisableParseTime)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		return newStdlibConnectorContainer(
			getMysqlConnectorFn(dsn, config.MysqlConfig, logger),
			dbconnopts,
			logger,
		), nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		connDetails, err := dbconnectconfig.NewFromMssqlConnection(config, options.connectionTimeoutSeconds)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		return newStdlibConnectorContainer(
			getMssqlConnectorFn(dsn, config.MssqlConfig, logger),
			dbconnopts,
			logger,
		), nil
	default:
		return nil, fmt.Errorf("unsupported connection: %T", config)
	}
}

func getPgConnectorFn(dsn string, config *mgmtv1alpha1.PostgresConnectionConfig, logger *slog.Logger) stdlibConnectorGetter {
	return func() (driver.Connector, func(), error) {
		connectorOpts := []postgrestunconnector.Option{
			postgrestunconnector.WithLogger(logger),
		}
		closers := []func(){}

		if config.GetClientTls() != nil {
			tlsConfig, err := getTLSConfig(config.GetClientTls(), logger)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to construct postgres client tls config: %w", err)
			}
			logger.Debug("constructed postgres client tls config")
			connectorOpts = append(connectorOpts, postgrestunconnector.WithTLSConfig(tlsConfig))
		}
		if config.GetTunnel() != nil {
			cfg, err := getTunnelConfig(config.GetTunnel())
			if err != nil {
				return nil, nil, fmt.Errorf("unable to construct postgres client tunnel config: %w", err)
			}
			logger.Debug("constructed postgres tunnel config")
			dialer := tun.NewLazySSHDialer(cfg.Addr, cfg.ClientConfig, tun.DefaultSSHDialerConfig(), logger)
			connectorOpts = append(connectorOpts, postgrestunconnector.WithDialer(dialer))
			closers = append(closers, func() {
				logger.Debug("closing postgres ssh dialer")
				err := dialer.Close()
				if err != nil {
					logger.Error(fmt.Sprintf("unable to close postgres dialer: %s", err.Error()))
				}
			})
		}
		connector, closer, err := postgrestunconnector.New(dsn, connectorOpts...)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to construct postgres connector: %w", err)
		}
		logger.Debug("built postgres database connector")
		closers = append(closers, closer)

		reverseCloser := func() {
			for i := len(closers) - 1; i >= 0; i-- {
				closers[i]()
			}
		}
		return connector, reverseCloser, nil
	}
}

func getMysqlConnectorFn(dsn string, config *mgmtv1alpha1.MysqlConnectionConfig, logger *slog.Logger) stdlibConnectorGetter {
	return func() (driver.Connector, func(), error) {
		connectorOpts := []mysqltunconnector.Option{}
		closers := []func(){}

		if config.GetClientTls() != nil {
			tlsConfig, err := getTLSConfig(config.GetClientTls(), logger)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to construct mysql client tls config: %w", err)
			}
			logger.Debug("constructed mysql client tls config")
			connectorOpts = append(connectorOpts, mysqltunconnector.WithTLSConfig(tlsConfig))
		}
		if config.GetTunnel() != nil {
			cfg, err := getTunnelConfig(config.GetTunnel())
			if err != nil {
				return nil, nil, fmt.Errorf("unable to construct mysql client tunnel config: %w", err)
			}
			logger.Debug("constructed mysql tunnel config")
			dialer := tun.NewLazySSHDialer(cfg.Addr, cfg.ClientConfig, tun.DefaultSSHDialerConfig(), logger)
			connectorOpts = append(connectorOpts, mysqltunconnector.WithDialer(dialer))
			closers = append(closers, func() {
				logger.Debug("closing mysql ssh dialer")
				err := dialer.Close()
				if err != nil {
					logger.Error(fmt.Sprintf("unable to close mysql dialer: %s", err.Error()))
				}
			})
		}
		connector, closer, err := mysqltunconnector.New(dsn, connectorOpts...)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to construct mysql connector: %w", err)
		}
		logger.Debug("built mysql database connector")
		closers = append(closers, closer)

		reverseCloser := func() {
			for i := len(closers) - 1; i >= 0; i-- {
				closers[i]()
			}
		}
		return connector, reverseCloser, nil
	}
}

func getMssqlConnectorFn(dsn string, config *mgmtv1alpha1.MssqlConnectionConfig, logger *slog.Logger) stdlibConnectorGetter {
	return func() (driver.Connector, func(), error) {
		connectorOpts := []mssqltunconnector.Option{}
		closers := []func(){}

		if config.GetClientTls() != nil {
			tlsConfig, err := getTLSConfig(config.GetClientTls(), logger)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to construct mssql client tls config: %w", err)
			}
			logger.Debug("constructed mssql client tls config")
			connectorOpts = append(connectorOpts, mssqltunconnector.WithTLSConfig(tlsConfig))
		}
		if config.GetTunnel() != nil {
			cfg, err := getTunnelConfig(config.GetTunnel())
			if err != nil {
				return nil, nil, fmt.Errorf("unable to construct mssql tunnel config: %w", err)
			}
			logger.Debug("constructed mssql tunnel config")
			dialer := tun.NewLazySSHDialer(cfg.Addr, cfg.ClientConfig, tun.DefaultSSHDialerConfig(), logger)
			connectorOpts = append(connectorOpts, mssqltunconnector.WithDialer(dialer))
			closers = append(closers, func() {
				logger.Debug("closing mssql ssh dialer")
				err := dialer.Close()
				if err != nil {
					logger.Error(fmt.Sprintf("unable to close mssql dialer: %s", err.Error()))
				}
			})
		}
		connector, closer, err := mssqltunconnector.New(dsn, connectorOpts...)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to construct mssql connector: %w", err)
		}
		logger.Debug("built mssql database connector")
		closers = append(closers, closer)

		reverseCloser := func() {
			for i := len(closers) - 1; i >= 0; i-- {
				closers[i]()
			}
		}
		return connector, reverseCloser, nil
	}
}

func getConnectionOptsFromConnectionConfig(cc *mgmtv1alpha1.ConnectionConfig) (*DbConnectionOptions, error) {
	switch config := cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return sqlConnOptsToDbConnOpts(config.MysqlConfig.GetConnectionOptions())
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return sqlConnOptsToDbConnOpts(config.PgConfig.GetConnectionOptions())
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return sqlConnOptsToDbConnOpts(config.MssqlConfig.GetConnectionOptions())
	default:
		return sqlConnOptsToDbConnOpts(&mgmtv1alpha1.SqlConnectionOptions{})
	}
}

func sqlConnOptsToDbConnOpts(co *mgmtv1alpha1.SqlConnectionOptions) (*DbConnectionOptions, error) {
	if co == nil {
		co = &mgmtv1alpha1.SqlConnectionOptions{}
	}
	var connMaxIdleTime *time.Duration
	if co.GetMaxIdleDuration() != "" {
		duration, err := time.ParseDuration(co.GetMaxIdleDuration())
		if err != nil {
			return nil, fmt.Errorf("max idle duration is not a valid Go duration string: %w", err)
		}
		connMaxIdleTime = &duration
	}
	var connMaxLifetime *time.Duration
	if co.GetMaxOpenDuration() != "" {
		duration, err := time.ParseDuration(co.GetMaxOpenDuration())
		if err != nil {
			return nil, fmt.Errorf("max open duration is not a vlaid Go duration string: %w", err)
		}
		connMaxLifetime = &duration
	}
	return &DbConnectionOptions{
		MaxOpenConns:    convertInt32PtrToIntPtr(co.MaxConnectionLimit),
		MaxIdleConns:    convertInt32PtrToIntPtr(co.MaxIdleConnections),
		ConnMaxIdleTime: connMaxIdleTime,
		ConnMaxLifetime: connMaxLifetime,
	}, nil
}

func convertInt32PtrToIntPtr(input *int32) *int {
	if input == nil {
		return nil
	}
	value := int(*input)
	return &value
}

type tunnelConfig struct {
	Addr         string
	ClientConfig *ssh.ClientConfig
}

func getTunnelConfig(tunnel *mgmtv1alpha1.SSHTunnel) (*tunnelConfig, error) {
	var hostcallback ssh.HostKeyCallback
	if tunnel.GetKnownHostPublicKey() != "" {
		publickey, err := tun.ParseSshKey(tunnel.GetKnownHostPublicKey())
		if err != nil {
			return nil, fmt.Errorf("unable to parse ssh known host public key: %w", err)
		}
		hostcallback = ssh.FixedHostKey(publickey)
	} else {
		hostcallback = ssh.InsecureIgnoreHostKey() //nolint:gosec // the user has chosen to not provide a known host public key
	}
	authmethod, err := tun.GetTunnelAuthMethodFromSshConfig(tunnel.GetAuthentication())
	if err != nil {
		return nil, fmt.Errorf("unable to parse ssh auth method: %w", err)
	}

	authmethods := []ssh.AuthMethod{}
	if authmethod != nil {
		authmethods = append(authmethods, authmethod)
	}

	return &tunnelConfig{
		Addr: getSshAddr(tunnel),
		ClientConfig: &ssh.ClientConfig{
			User:            tunnel.GetUser(),
			Auth:            authmethods,
			HostKeyCallback: hostcallback,
			Timeout:         15 * time.Second, // todo: make configurable
		},
	}, nil
}

func getSshAddr(tunnel *mgmtv1alpha1.SSHTunnel) string {
	host := tunnel.GetHost()
	port := tunnel.GetPort()
	if port > 0 {
		return net.JoinHostPort(host, strconv.FormatInt(int64(port), 10))
	}
	return host
}

type stdlibConnectorGetter func() (driver.Connector, func(), error)

func newStdlibConnectorContainer(
	getter stdlibConnectorGetter,
	connopts *DbConnectionOptions,
	logger *slog.Logger,
) *stdlibConnectorContainer {
	return &stdlibConnectorContainer{getter: getter, connopts: connopts, logger: logger}
}

type stdlibConnectorContainer struct {
	db      *sql.DB
	mu      sync.Mutex
	cleanup func()

	getter   stdlibConnectorGetter
	connopts *DbConnectionOptions
	logger   *slog.Logger
}

func (s *stdlibConnectorContainer) Open() (sqldbtx.DBTX, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	connector, cleanup, err := s.getter()
	if err != nil {
		return nil, err
	}
	s.cleanup = cleanup
	db := sql.OpenDB(connector)
	setConnectionOpts(db, s.connopts)
	s.db = db
	return sqlretry.NewDefault(
		s.db,
		s.logger,
	), err
}
func (s *stdlibConnectorContainer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db := s.db
	cleanup := s.cleanup
	if cleanup != nil {
		defer cleanup()
	}
	if db == nil {
		return nil
	}
	s.db = nil
	s.cleanup = nil
	return db.Close()
}

type DbConnectionOptions struct {
	MaxOpenConns *int
	MaxIdleConns *int

	ConnMaxIdleTime *time.Duration
	ConnMaxLifetime *time.Duration
}

func setConnectionOpts(db *sql.DB, connopts *DbConnectionOptions) {
	if connopts != nil {
		if connopts.ConnMaxIdleTime != nil {
			db.SetConnMaxIdleTime(*connopts.ConnMaxIdleTime)
		}
		if connopts.ConnMaxLifetime != nil {
			db.SetConnMaxLifetime(*connopts.ConnMaxLifetime)
		}
		if connopts.MaxIdleConns != nil {
			db.SetMaxIdleConns(*connopts.MaxIdleConns)
		}
		if connopts.MaxOpenConns != nil {
			db.SetMaxOpenConns(*connopts.MaxOpenConns)
		}
	}
}

type ConnectionDetails struct {
	dbconnectconfig.DbConnectConfig
	MaxConnectionLimit *int32
}

func (c *ConnectionDetails) String() string {
	return c.DbConnectConfig.String()
}

// getTLSConfig converts a ClientTlsConfig proto message to a *tls.Config
func getTLSConfig(cfg *mgmtv1alpha1.ClientTlsConfig, logger *slog.Logger) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Configure root CA cert if provided
	rootCert := cfg.GetRootCert()
	if rootCert != "" {
		logger.Debug("root cert provided, adding to rootcas for client tls connection")
		rootCertPool := x509.NewCertPool()
		if !rootCertPool.AppendCertsFromPEM([]byte(rootCert)) {
			return nil, fmt.Errorf("failed to append root certificate")
		}
		tlsConfig.RootCAs = rootCertPool
	}

	// Configure client certificate if both cert and key are provided
	clientCert := cfg.GetClientCert()
	clientKey := cfg.GetClientKey()
	if clientCert != "" && clientKey != "" {
		logger.Debug("client cert and key provided, adding to certificates for client tls connection")
		cert, err := tls.X509KeyPair([]byte(cfg.GetClientCert()), []byte(cfg.GetClientKey()))
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if clientCert != "" || clientKey != "" {
		// If only one of cert or key is provided, return an error
		return nil, fmt.Errorf("both client certificate and key must be provided")
	}

	serverName := cfg.GetServerName()
	if serverName != "" {
		logger.Debug("server name provided, added to certificates for client tls connection")
		tlsConfig.ServerName = serverName
	}

	return tlsConfig, nil
}
