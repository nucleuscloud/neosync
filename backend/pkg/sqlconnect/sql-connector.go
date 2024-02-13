package sqlconnect

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
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
	details, err := GetConnectionDetails(connectionConfig, connectionTimeout, logger)
	if err != nil {
		return nil, err
	}
	return newSqlDb(details), nil
}

func (rc *SqlOpenConnector) NewPgPoolFromConnectionConfig(pgconfig *mgmtv1alpha1.PostgresConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (PgPoolContainer, error) {
	if pgconfig == nil {
		return nil, errors.New("pgconfig was nil, expected *mgmtv1alpha1.PostgresConnectionConfig")
	}
	details, err := GetConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
			PgConfig: pgconfig,
		},
	}, connectionTimeout, logger)
	if err != nil {
		return nil, err
	}
	return newPgPool(details), nil
}

type ConnectionDetails struct {
	GeneralDbConnectConfig

	Tunnel *sshtunnel.Sshtunnel
}

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
	localhost      = "localhost"
	randomPort     = 0
)

// Method for retrieving connection details, including tunneling information.
// Only use if requiring direct access to the SSH Tunnel, otherwise the SqlConnector should be used instead.
func GetConnectionDetails(c *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (*ConnectionDetails, error) {
	if c == nil {
		return nil, errors.New("connection config was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}
	switch config := c.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if config.PgConfig.Tunnel != nil {
			destination, err := getEndpointFromPgConnectionConfig(config)
			if err != nil {
				return nil, err
			}
			authmethod, err := getTunnelAuthMethodFromSshConfig(config.PgConfig.GetTunnel().GetAuthentication())
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
				logger,
			)
			connDetails, err := getGeneralDbConnectConfigFromPg(config, connectionTimeout)
			if err != nil {
				return nil, err
			}
			connDetails.Host = localhost
			connDetails.Port = randomPort
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
			}, nil
		}

		connDetails, err := getGeneralDbConnectConfigFromPg(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		return &ConnectionDetails{
			GeneralDbConnectConfig: *connDetails,
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if config.MysqlConfig.Tunnel != nil {
			destination, err := getEndpointFromMysqlConnectionConfig(config)
			if err != nil {
				return nil, err
			}
			authmethod, err := getTunnelAuthMethodFromSshConfig(config.MysqlConfig.Tunnel.Authentication)
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
				logger,
			)
			connDetails, err := getGeneralDbConnectionConfigFromMysql(config, connectionTimeout)
			if err != nil {
				return nil, err
			}
			connDetails.Host = localhost
			connDetails.Port = randomPort
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
			}, nil
		}

		connDetails, err := getGeneralDbConnectionConfigFromMysql(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		return &ConnectionDetails{
			GeneralDbConnectConfig: *connDetails,
		}, nil
	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}

// Auth Method is optional and will return nil if there is no valid method.
// Will only return error if unable to parse the private key into an auth method
func getTunnelAuthMethodFromSshConfig(auth *mgmtv1alpha1.SSHAuthentication) (ssh.AuthMethod, error) {
	if auth == nil {
		return nil, nil
	}
	switch config := auth.AuthConfig.(type) {
	case *mgmtv1alpha1.SSHAuthentication_Passphrase:
		return ssh.Password(config.Passphrase.Value), nil
	case *mgmtv1alpha1.SSHAuthentication_PrivateKey:
		authMethod, err := sshtunnel.GetPrivateKeyAuthMethod([]byte(config.PrivateKey.Value), config.PrivateKey.Passphrase)
		if err != nil {
			return nil, err
		}
		return authMethod, nil
	default:
		return nil, nil
	}
}

func getEndpointFromPgConnectionConfig(config *mgmtv1alpha1.ConnectionConfig_PgConfig) (*sshtunnel.Endpoint, error) {
	switch cc := config.PgConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		return sshtunnel.NewEndpointWithUser(cc.Connection.Host, int(cc.Connection.Port), cc.Connection.User), nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		details, err := getGeneralDbConnectConfigFromPg(config, nil)
		if err != nil {
			return nil, err
		}
		return sshtunnel.NewEndpointWithUser(details.Host, int(details.Port), details.User), nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}

func getEndpointFromMysqlConnectionConfig(config *mgmtv1alpha1.ConnectionConfig_MysqlConfig) (*sshtunnel.Endpoint, error) {
	switch cc := config.MysqlConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		return sshtunnel.NewEndpointWithUser(cc.Connection.Host, int(cc.Connection.Port), cc.Connection.User), nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		details, err := getGeneralDbConnectionConfigFromMysql(config, nil)
		if err != nil {
			return nil, err
		}
		return sshtunnel.NewEndpointWithUser(details.Host, int(details.Port), details.User), nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}

type GeneralDbConnectConfig struct {
	Driver string

	Host     string
	Port     int32
	Database string
	User     string
	Pass     string

	Protocol *string

	QueryParams url.Values
}

func (g *GeneralDbConnectConfig) String() string {
	if g.Driver == postgresDriver {
		u := url.URL{
			Scheme: "postgres",
			Host:   fmt.Sprintf("%s:%d", g.Host, g.Port),
			Path:   g.Database,
		}

		// Add user info
		if g.User != "" || g.Pass != "" {
			u.User = url.UserPassword(g.User, g.Pass)
		}
		u.RawQuery = g.QueryParams.Encode()
		return u.String()
	}
	if g.Driver == mysqlDriver {
		protocol := "tcp"
		if g.Protocol != nil {
			protocol = *g.Protocol
		}
		address := fmt.Sprintf("(%s:%d)", g.Host, g.Port)

		// User info
		userInfo := url.UserPassword(g.User, g.Pass).String()

		// Base DSN
		dsn := fmt.Sprintf("%s@%s%s/%s", userInfo, protocol, address, g.Database)

		// Append query parameters if any
		if len(g.QueryParams) > 0 {
			query := g.QueryParams.Encode()
			dsn += "?" + query
		}
		return dsn
	}
	return ""
}

func getGeneralDbConnectionConfigFromMysql(config *mgmtv1alpha1.ConnectionConfig_MysqlConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
	switch cc := config.MysqlConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		query := url.Values{}
		if connectionTimeout != nil {
			query.Add("timeout", fmt.Sprintf("%ds", *connectionTimeout))
		}
		return &GeneralDbConnectConfig{
			Driver:      mysqlDriver,
			Host:        cc.Connection.Host,
			Port:        cc.Connection.Port,
			Database:    cc.Connection.Name,
			User:        cc.Connection.User,
			Pass:        cc.Connection.Pass,
			Protocol:    &cc.Connection.Protocol,
			QueryParams: query,
		}, nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		return nil, nucleuserrors.NewNotImplemented("not currently implemented")
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid mysql connection")
	}
}

func getGeneralDbConnectConfigFromPg(config *mgmtv1alpha1.ConnectionConfig_PgConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
	switch cc := config.PgConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		query := url.Values{}
		if cc.Connection.SslMode != nil {
			query.Add("sslmode", *cc.Connection.SslMode)
		}
		if connectionTimeout != nil {
			query.Add("connect_timeout", fmt.Sprintf("%d", *connectionTimeout))
		}
		return &GeneralDbConnectConfig{
			Driver:      postgresDriver,
			Host:        cc.Connection.Host,
			Port:        cc.Connection.Port,
			Database:    cc.Connection.Name,
			User:        cc.Connection.User,
			Pass:        cc.Connection.Pass,
			QueryParams: query,
		}, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		u, err := url.Parse(cc.Url)
		if err != nil {
			return nil, err
		}

		// Extract user info
		user := u.User.Username()
		pass, ok := u.User.Password()
		if !ok {
			return nil, errors.New("unable to get password for pg string")
		}

		// Extract host and port
		host, portStr := u.Hostname(), u.Port()

		// Convert port to integer
		var port int64
		if portStr != "" {
			port, err = strconv.ParseInt(portStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %v", err)
			}
		}
		return &GeneralDbConnectConfig{
			Driver:      postgresDriver,
			Host:        host,
			Port:        int32(port),
			Database:    strings.TrimPrefix(u.Path, "/"),
			User:        user,
			Pass:        pass,
			QueryParams: u.Query(),
		}, nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}
