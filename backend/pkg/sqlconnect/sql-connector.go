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
	dbconnectconfig "github.com/nucleuscloud/neosync/backend/pkg/dbconnect-config"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	tun "github.com/nucleuscloud/neosync/internal/sshtunnel"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mssqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mysqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/postgrestunconnector"
	"golang.org/x/crypto/ssh"
)

// interface used by SqlConnector to abstract away the opening and closing of a sqldb that includes tunneling
type SqlDbContainer interface {
	Open() (SqlDBTX, error)
	Close() error
}

type SqlDBTX interface {
	mysql_queries.DBTX

	PingContext(context.Context) error
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

type SqlConnector interface {
	NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (SqlDbContainer, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) NewDbFromConnectionConfig(cc *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (SqlDbContainer, error) {
	if cc == nil {
		return nil, errors.New("connectionConfig was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}

	// todo: need to handle client tls with new setup

	switch config := cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		connDetails, err := dbconnectconfig.NewFromPostgresConnection(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		if config.PgConfig.GetTunnel() != nil {
			return newStdlibConnectorContainer(
				func() (driver.Connector, func(), error) {
					tunnelcfg, err := getTunnelConfig(config.PgConfig.GetTunnel())
					if err != nil {
						return nil, nil, err
					}
					dialer := tun.NewLazySSHDialer(tunnelcfg.Addr, tunnelcfg.ClientConfig)
					return postgrestunconnector.New(dialer, dsn)
				},
			), nil
		} else {
			return newStdlibContainer("pgx", dsn), nil
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		connDetails, err := dbconnectconfig.NewFromMysqlConnection(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		if config.MysqlConfig.GetTunnel() != nil {
			return newStdlibConnectorContainer(
				func() (driver.Connector, func(), error) {
					tunnelcfg, err := getTunnelConfig(config.MysqlConfig.GetTunnel())
					if err != nil {
						return nil, nil, err
					}
					dialer := tun.NewLazySSHDialer(tunnelcfg.Addr, tunnelcfg.ClientConfig)
					return mysqltunconnector.New(dialer, dsn)
				},
			), nil
		}
		return newStdlibContainer("mysql", dsn), nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		connDetails, err := dbconnectconfig.NewFromMssqlConnection(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		dsn := connDetails.String()

		if config.MssqlConfig.GetTunnel() != nil {
			return newStdlibConnectorContainer(
				func() (driver.Connector, func(), error) {
					tunnelcfg, err := getTunnelConfig(config.MssqlConfig.GetTunnel())
					if err != nil {
						return nil, nil, err
					}
					dialer := tun.NewLazySSHDialer(tunnelcfg.Addr, tunnelcfg.ClientConfig)
					return mssqltunconnector.New(dialer, dsn)
				},
			), nil
		}
		return newStdlibContainer("sqlserver", dsn), nil
	default:
		return nil, fmt.Errorf("unsupported connection: %T", config)
	}
}

type tunnelConfig struct {
	Addr         string
	ClientConfig *ssh.ClientConfig
}

func getTunnelConfig(tunnel *mgmtv1alpha1.SSHTunnel) (*tunnelConfig, error) {
	var hostcallback ssh.HostKeyCallback
	if tunnel.GetKnownHostPublicKey() != "" {
		publickey, err := sshtunnel.ParseSshKey(tunnel.GetKnownHostPublicKey())
		if err != nil {
			return nil, fmt.Errorf("unable to parse ssh known host public key: %w", err)
		}
		hostcallback = ssh.FixedHostKey(publickey)
	} else {
		hostcallback = ssh.InsecureIgnoreHostKey() //nolint:gosec // the user has chosen to not provide a known host public key
	}
	authmethod, err := sshtunnel.GetTunnelAuthMethodFromSshConfig(tunnel.GetAuthentication())
	if err != nil {
		return nil, fmt.Errorf("unable to parse ssh auth method: %w", err)
	}

	return &tunnelConfig{
		Addr: getSshAddr(tunnel),
		ClientConfig: &ssh.ClientConfig{
			User:            tunnel.GetUser(),
			Auth:            []ssh.AuthMethod{authmethod},
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

func newStdlibConnectorContainer(getter func() (driver.Connector, func(), error)) *stdlibConnectorContainer {
	return &stdlibConnectorContainer{getter: getter}
}

type stdlibConnectorContainer struct {
	db      *sql.DB
	mu      sync.Mutex
	cleanup func()

	getter func() (driver.Connector, func(), error)
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

func newStdlibContainer(drvr, dsn string) *stdlibContainer {
	return &stdlibContainer{driver: drvr, dsn: dsn}
}

type stdlibContainer struct {
	db *sql.DB
	mu sync.Mutex

	driver string
	dsn    string
}

func (s *stdlibContainer) Open() (SqlDBTX, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := sql.Open(s.driver, s.dsn)
	if err != nil {
		return nil, err
	}
	s.db = db
	return db, nil
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
	dbconnectconfig.GeneralDbConnectConfig
	MaxConnectionLimit *int32
}

func (c *ConnectionDetails) String() string {
	return c.GeneralDbConnectConfig.String()
}

type ClientCertConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}

// Method for retrieving connection details, including tunneling information.
// // Only use if requiring direct access to the SSH Tunnel, otherwise the SqlConnector should be used instead.
// func getConnectionDetails(
// 	c *mgmtv1alpha1.ConnectionConfig,
// 	connectionTimeout *uint32,
// 	handleClientTlsConfig clienttls.ClientTlsFileHandler,
// 	logger *slog.Logger,
// ) (*ConnectionDetails, error) {
// 	if c == nil {
// 		return nil, errors.New("connection config was nil, expected *mgmtv1alpha1.ConnectionConfig")
// 	}
// 	switch config := c.Config.(type) {
// 	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
// 		var maxConnLimit *int32
// 		if config.PgConfig.ConnectionOptions != nil {
// 			maxConnLimit = config.PgConfig.ConnectionOptions.MaxConnectionLimit
// 		}
// 		if config.PgConfig.GetClientTls() != nil {
// 			_, err := handleClientTlsConfig(config.PgConfig.GetClientTls())
// 			if err != nil {
// 				return nil, err
// 			}
// 		}

// 		connDetails, err := dbconnectconfig.NewFromPostgresConnection(config, connectionTimeout)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return &ConnectionDetails{
// 			GeneralDbConnectConfig: *connDetails,
// 			MaxConnectionLimit:     maxConnLimit,
// 		}, nil
// 	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
// 		var maxConnLimit *int32
// 		if config.MysqlConfig.ConnectionOptions != nil {
// 			maxConnLimit = config.MysqlConfig.ConnectionOptions.MaxConnectionLimit
// 		}

// 		connDetails, err := dbconnectconfig.NewFromMysqlConnection(config, connectionTimeout)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return &ConnectionDetails{
// 			GeneralDbConnectConfig: *connDetails,
// 			MaxConnectionLimit:     maxConnLimit,
// 		}, nil
// 	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
// 		var maxConnLimit *int32
// 		if config.MssqlConfig.GetConnectionOptions() != nil {
// 			maxConnLimit = config.MssqlConfig.GetConnectionOptions().MaxConnectionLimit
// 		}
// 		connDetails, err := dbconnectconfig.NewFromMssqlConnection(config, connectionTimeout)
// 		if err != nil {
// 			return nil, fmt.Errorf("unable to compile connection details for mssql connection: %w", err)
// 		}
// 		return &ConnectionDetails{
// 			GeneralDbConnectConfig: *connDetails,
// 			MaxConnectionLimit:     maxConnLimit,
// 		}, nil
// 	default:
// 		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("this connection config (%T) is not currently supported", config))
// 	}
// }
