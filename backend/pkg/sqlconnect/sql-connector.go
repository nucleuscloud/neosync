package sqlconnect

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	dbconnectconfig "github.com/nucleuscloud/neosync/backend/pkg/dbconnect-config"
	tun "github.com/nucleuscloud/neosync/internal/sshtunnel"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mssqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mysqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/postgrestunconnector"
	"golang.org/x/crypto/ssh"
)

// interface used by SqlConnector to abstract away the opening and closing of a sqldb that includes tunnelingff
type SqlDbContainer interface {
	Open() (SqlDBTX, error)
	Close() error
}

type SqlDBTX interface {
	mysql_queries.DBTX

	PingContext(context.Context) error
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

type SqlConnectorOption func(*sqlConnectorOptions)

type sqlConnectorOptions struct {
	mysqlDisableParseTime bool
	postgresDriver        string
}

// WithMysqlParseTimeDisabled disables MySQL time parsing
func WithMysqlParseTimeDisabled() SqlConnectorOption {
	return func(opts *sqlConnectorOptions) {
		opts.mysqlDisableParseTime = true
	}
}

// WithPostgresDriver overrides default postgres driver
func WithDefaultPostgresDriver() SqlConnectorOption {
	return func(opts *sqlConnectorOptions) {
		opts.postgresDriver = "postgres"
	}
}

type SqlConnector interface {
	NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger, opts ...SqlConnectorOption) (SqlDbContainer, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) NewDbFromConnectionConfig(cc *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger, opts ...SqlConnectorOption) (SqlDbContainer, error) {
	if cc == nil {
		return nil, errors.New("connectionConfig was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}

	options := sqlConnectorOptions{
		postgresDriver: "pgx",
	}
	for _, opt := range opts {
		opt(&options)
	}

	dbconnopts, err := getConnectionOptsFromConnectionConfig(cc)
	if err != nil {
		return nil, err
	}

	switch config := cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if config.PgConfig.GetClientTls() != nil {
			_, err := clienttls.UpsertCLientTlsFiles(config.PgConfig.GetClientTls())
			if err != nil {
				return nil, fmt.Errorf("unable to upsert client tls files: %w", err)
			}
		}
		connDetails, err := dbconnectconfig.NewFromPostgresConnection(config, connectionTimeout, logger)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		if config.PgConfig.GetTunnel() != nil {
			return newStdlibConnectorContainer(
				getTunnelConnectorFn(
					config.PgConfig.GetTunnel(),
					func(dialer tun.Dialer) (driver.Connector, func(), error) {
						return postgrestunconnector.New(dialer, dsn)
					},
					logger,
				),
				dbconnopts,
			), nil
		} else {
			return newStdlibContainer(options.postgresDriver, dsn, dbconnopts), nil
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		connDetails, err := dbconnectconfig.NewFromMysqlConnection(config, connectionTimeout, logger, options.mysqlDisableParseTime)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		if config.MysqlConfig.GetTunnel() != nil {
			return newStdlibConnectorContainer(
				getTunnelConnectorFn(
					config.MysqlConfig.GetTunnel(),
					func(dialer tun.Dialer) (driver.Connector, func(), error) {
						return mysqltunconnector.New(dialer, dsn)
					},
					logger,
				),
				dbconnopts,
			), nil
		}
		return newStdlibContainer("mysql", dsn, dbconnopts), nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		connDetails, err := dbconnectconfig.NewFromMssqlConnection(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		if config.MssqlConfig.GetTunnel() != nil {
			return newStdlibConnectorContainer(
				getTunnelConnectorFn(
					config.MssqlConfig.GetTunnel(),
					func(dialer tun.Dialer) (driver.Connector, func(), error) {
						return mssqltunconnector.New(dialer, dsn)
					},
					logger,
				),
				dbconnopts,
			), nil
		}
		return newStdlibContainer("sqlserver", dsn, dbconnopts), nil
	default:
		return nil, fmt.Errorf("unsupported connection: %T", config)
	}
}

func getTunnelConnectorFn(
	tunnel *mgmtv1alpha1.SSHTunnel,
	getConnector func(dialer tun.Dialer) (driver.Connector, func(), error),
	logger *slog.Logger,
) func() (driver.Connector, func(), error) {
	return func() (driver.Connector, func(), error) {
		cfg, err := getTunnelConfig(tunnel)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to construct ssh tunnel config: %w", err)
		}
		logger.Debug("constructed tunnel config")
		dialer := tun.NewLazySSHDialer(cfg.Addr, cfg.ClientConfig)
		conn, cleanup, err := getConnector(dialer)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to build db connector: %w", err)
		}
		logger.Debug("built database connector with ssh dialer")
		wrappedCleanup := func() {
			logger.Debug("cleaning up tunnel connector")
			cleanup()
			logger.Debug("connector cleanup completed")
			if err := dialer.Close(); err != nil {
				logger.Error(fmt.Errorf("encountered error when closing ssh dialer: %w", err).Error())
			}
			logger.Debug("tunnel connector cleanup completed")
		}
		return conn, wrappedCleanup, nil
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
			Timeout:         10 * time.Second, // todo: make configurable
		},
	}, nil
}

func getSshAddr(tunnel *mgmtv1alpha1.SSHTunnel) string {
	host := tunnel.GetHost()
	port := tunnel.GetPort()
	if port > 0 {
		return fmt.Sprintf("%s:%d", host, port)
	}
	return host
}

func newStdlibConnectorContainer(getter func() (driver.Connector, func(), error), connopts *DbConnectionOptions) *stdlibConnectorContainer {
	return &stdlibConnectorContainer{getter: getter, connopts: connopts}
}

type stdlibConnectorContainer struct {
	db      *sql.DB
	mu      sync.Mutex
	cleanup func()

	getter   func() (driver.Connector, func(), error)
	connopts *DbConnectionOptions
}

func (s *stdlibConnectorContainer) Open() (SqlDBTX, error) {
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
	return s.db, err
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

func newStdlibContainer(drvr, dsn string, connOpts *DbConnectionOptions) *stdlibContainer {
	return &stdlibContainer{driver: drvr, dsn: dsn, connopts: connOpts}
}

type stdlibContainer struct {
	db *sql.DB
	mu sync.Mutex

	driver   string
	dsn      string
	connopts *DbConnectionOptions
}

func (s *stdlibContainer) Open() (SqlDBTX, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := sql.Open(s.driver, s.dsn)
	if err != nil {
		return nil, err
	}
	setConnectionOpts(db, s.connopts)
	s.db = db
	return db, nil
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

func (s *stdlibContainer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db := s.db
	if db == nil {
		return nil
	}
	s.db = nil
	return db.Close()
}

type ConnectionDetails struct {
	dbconnectconfig.DbConnectConfig
	MaxConnectionLimit *int32
}

func (c *ConnectionDetails) String() string {
	return c.DbConnectConfig.String()
}

type ClientCertConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}
