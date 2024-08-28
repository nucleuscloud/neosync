package sqlconnect

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	dbconnectconfig "github.com/nucleuscloud/neosync/backend/pkg/dbconnect-config"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	"golang.org/x/crypto/ssh"
)

type SqlDBTX interface {
	mysql_queries.DBTX

	PingContext(context.Context) error
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// Allows instantiating a sql db or pg pool container that includes SSH tunneling if the config requires it
type SqlConnector interface {
	NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (SqlDbContainer, error)
	NewPgPoolFromConnectionConfig(pgconfig *mgmtv1alpha1.PostgresConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (PgPoolContainer, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (SqlDbContainer, error) {
	if connectionConfig == nil {
		return nil, errors.New("connectionConfig was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}

	details, err := GetConnectionDetails(connectionConfig, connectionTimeout, clienttls.UpsertCLientTlsFiles, logger)
	if err != nil {
		return nil, err
	}

	return newSqlDb(details, logger), nil
}

func (rc *SqlOpenConnector) NewPgPoolFromConnectionConfig(pgconfig *mgmtv1alpha1.PostgresConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (PgPoolContainer, error) {
	if pgconfig == nil {
		return nil, errors.New("pgconfig was nil, expected *mgmtv1alpha1.PostgresConnectionConfig")
	}
	details, err := GetConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
			PgConfig: pgconfig,
		},
	}, connectionTimeout, clienttls.UpsertCLientTlsFiles, logger)
	if err != nil {
		return nil, err
	}
	return newPgPool(details, logger), nil
}

type ConnectionDetails struct {
	dbconnectconfig.GeneralDbConnectConfig
	MaxConnectionLimit *int32

	Tunnel *sshtunnel.Sshtunnel
}

func (c *ConnectionDetails) GetTunnel() *sshtunnel.Sshtunnel {
	return c.Tunnel
}

func (c *ConnectionDetails) String() string {
	if c.Tunnel != nil {
		// todo: would be great to check if tunnel has been started...
		localhost, port := c.Tunnel.GetLocalHostPort()
		c.GeneralDbConnectConfig.SetHost(localhost)
		c.GeneralDbConnectConfig.SetPort(int32(port)) //nolint:gosec // Ignoring for now
	}
	return c.GeneralDbConnectConfig.String()
}

type ClientCertConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}

const (
	localhost  = "localhost"
	randomPort = 0
)

// Method for retrieving connection details, including tunneling information.
// Only use if requiring direct access to the SSH Tunnel, otherwise the SqlConnector should be used instead.
func GetConnectionDetails(
	c *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	handleClientTlsConfig clienttls.ClientTlsFileHandler,
	logger *slog.Logger,
) (*ConnectionDetails, error) {
	if c == nil {
		return nil, errors.New("connection config was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}
	switch config := c.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		var maxConnLimit *int32
		if config.PgConfig.ConnectionOptions != nil {
			maxConnLimit = config.PgConfig.ConnectionOptions.MaxConnectionLimit
		}
		if config.PgConfig.GetClientTls() != nil {
			_, err := handleClientTlsConfig(config.PgConfig.GetClientTls())
			if err != nil {
				return nil, err
			}
		}
		if config.PgConfig.Tunnel != nil {
			destination, err := getEndpointFromPgConnectionConfig(config)
			if err != nil {
				return nil, err
			}
			authmethod, err := sshtunnel.GetTunnelAuthMethodFromSshConfig(config.PgConfig.GetTunnel().GetAuthentication())
			if err != nil {
				return nil, err
			}
			var publickey ssh.PublicKey
			if config.PgConfig.Tunnel.KnownHostPublicKey != nil {
				publickey, err = sshtunnel.ParseSshKey(*config.PgConfig.Tunnel.KnownHostPublicKey)
				if err != nil {
					return nil, err
				}
			}
			tunnel := sshtunnel.New(
				sshtunnel.NewEndpointWithUser(config.PgConfig.Tunnel.GetHost(), int(config.PgConfig.Tunnel.GetPort()), config.PgConfig.Tunnel.GetUser()),
				authmethod,
				destination,
				sshtunnel.NewEndpoint(localhost, randomPort),
				1,
				publickey,
			)
			connDetails, err := dbconnectconfig.NewFromPostgresConnection(config, connectionTimeout)
			if err != nil {
				return nil, err
			}
			portValue := int32(randomPort)
			connDetails.SetHost(localhost)
			connDetails.SetPort(portValue)
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
				MaxConnectionLimit:     maxConnLimit,
			}, nil
		}

		connDetails, err := dbconnectconfig.NewFromPostgresConnection(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		return &ConnectionDetails{
			GeneralDbConnectConfig: *connDetails,
			MaxConnectionLimit:     maxConnLimit,
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		var maxConnLimit *int32
		if config.MysqlConfig.ConnectionOptions != nil {
			maxConnLimit = config.MysqlConfig.ConnectionOptions.MaxConnectionLimit
		}
		if config.MysqlConfig.Tunnel != nil {
			destination, err := getEndpointFromMysqlConnectionConfig(config)
			if err != nil {
				return nil, err
			}
			authmethod, err := sshtunnel.GetTunnelAuthMethodFromSshConfig(config.MysqlConfig.Tunnel.Authentication)
			if err != nil {
				return nil, err
			}
			var publickey ssh.PublicKey
			if config.MysqlConfig.Tunnel.KnownHostPublicKey != nil {
				publickey, err = sshtunnel.ParseSshKey(*config.MysqlConfig.Tunnel.KnownHostPublicKey)
				if err != nil {
					return nil, err
				}
			}
			tunnel := sshtunnel.New(
				sshtunnel.NewEndpointWithUser(config.MysqlConfig.Tunnel.GetHost(), int(config.MysqlConfig.Tunnel.GetPort()), config.MysqlConfig.Tunnel.GetUser()),
				authmethod,
				destination,
				sshtunnel.NewEndpoint(localhost, randomPort),
				1,
				publickey,
			)

			connDetails, err := dbconnectconfig.NewFromMysqlConnection(config, connectionTimeout)
			if err != nil {
				return nil, err
			}

			portValue := int32(randomPort)
			connDetails.SetHost(localhost)
			connDetails.SetPort(portValue)
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
				MaxConnectionLimit:     maxConnLimit,
			}, nil
		}

		connDetails, err := dbconnectconfig.NewFromMysqlConnection(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		return &ConnectionDetails{
			GeneralDbConnectConfig: *connDetails,
			MaxConnectionLimit:     maxConnLimit,
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		var maxConnLimit *int32
		if config.MssqlConfig.GetConnectionOptions() != nil {
			maxConnLimit = config.MssqlConfig.GetConnectionOptions().MaxConnectionLimit
		}
		if config.MssqlConfig.GetTunnel() != nil {
			destination, err := getEndpointFromMssqlConnectionConfig(config)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve tunnel endpoint for mssql: %w", err)
			}
			authmethod, err := sshtunnel.GetTunnelAuthMethodFromSshConfig(config.MssqlConfig.GetTunnel().GetAuthentication())
			if err != nil {
				return nil, fmt.Errorf("unable to compile auth method for ssh tunneling for mssql: %w", err)
			}
			var publickey ssh.PublicKey
			if config.MssqlConfig.GetTunnel().GetKnownHostPublicKey() != "" {
				publickey, err = sshtunnel.ParseSshKey(config.MssqlConfig.GetTunnel().GetKnownHostPublicKey())
				if err != nil {
					return nil, fmt.Errorf("unable to parse provided known host public key for mssql tunnel: %w", err)
				}
			}
			tunnel := sshtunnel.New(
				sshtunnel.NewEndpointWithUser(config.MssqlConfig.GetTunnel().GetHost(), int(config.MssqlConfig.GetTunnel().GetPort()), config.MssqlConfig.GetTunnel().GetUser()),
				authmethod,
				destination,
				sshtunnel.NewEndpoint(localhost, randomPort),
				1,
				publickey,
			)

			connDetails, err := dbconnectconfig.NewFromMssqlConnection(config, connectionTimeout)
			if err != nil {
				return nil, fmt.Errorf("unable to compile connection details for mssql tunnel connection: %w", err)
			}

			portValue := int32(randomPort)
			connDetails.SetHost(localhost)
			connDetails.SetPort(portValue)
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
				MaxConnectionLimit:     maxConnLimit,
			}, nil
		}
		connDetails, err := dbconnectconfig.NewFromMssqlConnection(config, connectionTimeout)
		if err != nil {
			return nil, fmt.Errorf("unable to compile connection details for mssql connection: %w", err)
		}
		return &ConnectionDetails{
			GeneralDbConnectConfig: *connDetails,
			MaxConnectionLimit:     maxConnLimit,
		}, nil
	default:
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("this connection config (%T) is not currently supported", config))
	}
}

func getEndpointFromPgConnectionConfig(config *mgmtv1alpha1.ConnectionConfig_PgConfig) (*sshtunnel.Endpoint, error) {
	switch cc := config.PgConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		return sshtunnel.NewEndpointWithUser(cc.Connection.Host, int(cc.Connection.Port), cc.Connection.User), nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		details, err := dbconnectconfig.NewFromPostgresConnection(config, nil)
		if err != nil {
			return nil, err
		}
		port := 0
		if details.GetPort() != nil {
			port = int(*details.GetPort())
		}
		return sshtunnel.NewEndpointWithUser(details.GetHost(), port, details.GetUser()), nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}

func getEndpointFromMysqlConnectionConfig(config *mgmtv1alpha1.ConnectionConfig_MysqlConfig) (*sshtunnel.Endpoint, error) {
	switch cc := config.MysqlConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		return sshtunnel.NewEndpointWithUser(cc.Connection.Host, int(cc.Connection.Port), cc.Connection.User), nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		details, err := dbconnectconfig.NewFromMysqlConnection(config, nil)
		if err != nil {
			return nil, err
		}
		port := 0
		if details.GetPort() != nil {
			port = int(*details.GetPort())
		}
		return sshtunnel.NewEndpointWithUser(details.GetHost(), port, details.GetUser()), nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid mysql connection")
	}
}

func getEndpointFromMssqlConnectionConfig(config *mgmtv1alpha1.ConnectionConfig_MssqlConfig) (*sshtunnel.Endpoint, error) {
	switch cc := config.MssqlConfig.GetConnectionConfig().(type) {
	case *mgmtv1alpha1.MssqlConnectionConfig_Url:
		details, err := dbconnectconfig.NewFromMssqlConnection(config, nil)
		if err != nil {
			return nil, err
		}
		port := 0
		if details.GetPort() != nil {
			port = int(*details.GetPort())
		}
		return sshtunnel.NewEndpointWithUser(details.GetHost(), port, details.GetUser()), nil
	default:
		return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("must provide valid mssql connection: %T", cc))
	}
}
