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
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
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
	GeneralDbConnectConfig
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
		c.GeneralDbConnectConfig.Host = localhost
		c.GeneralDbConnectConfig.Port = shared.Ptr(int32(port))
	}
	return c.GeneralDbConnectConfig.String()
}

type ClientCertConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
	mssqlDriver    = "sqlserver"
	localhost      = "localhost"
	randomPort     = 0
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
			connDetails, err := getGeneralDbConnectConfigFromPg(config, connectionTimeout)
			if err != nil {
				return nil, err
			}
			portValue := int32(randomPort)
			connDetails.Host = localhost
			connDetails.Port = &portValue
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
				MaxConnectionLimit:     maxConnLimit,
			}, nil
		}

		connDetails, err := getGeneralDbConnectConfigFromPg(config, connectionTimeout)
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

			connDetails, err := getGeneralDbConnectionConfigFromMysql(config, connectionTimeout)
			if err != nil {
				return nil, err
			}

			portValue := int32(randomPort)
			connDetails.Host = localhost
			connDetails.Port = &portValue
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
				MaxConnectionLimit:     maxConnLimit,
			}, nil
		}

		connDetails, err := getGeneralDbConnectionConfigFromMysql(config, connectionTimeout)
		if err != nil {
			return nil, err
		}
		return &ConnectionDetails{
			GeneralDbConnectConfig: *connDetails,
			MaxConnectionLimit:     maxConnLimit,
		}, nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		var maxConnLimit *int32
		if config.MssqlConfig.GetConnectionOptions().MaxConnectionLimit != nil {
			maxConnLimit = config.MssqlConfig.GetConnectionOptions().MaxConnectionLimit
		}

		connDetails, err := getGeneralDbConnectionConfigFromMssql(config, connectionTimeout)
		if err != nil {
			return nil, err
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
		details, err := getGeneralDbConnectConfigFromPg(config, nil)
		if err != nil {
			return nil, err
		}
		port := 0
		if details.Port != nil {
			port = int(*details.Port)
		}
		return sshtunnel.NewEndpointWithUser(details.Host, port, details.User), nil
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
		port := 0
		if details.Port != nil {
			port = int(*details.Port)
		}
		return sshtunnel.NewEndpointWithUser(details.Host, port, details.User), nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid mysql connection")
	}
}

type GeneralDbConnectConfig struct {
	Driver string

	Host string
	Port *int32
	// For mssql this is actually the path..the database is provided as a query parameter
	Database *string
	User     string
	Pass     string

	Protocol *string

	QueryParams url.Values
}

func (g *GeneralDbConnectConfig) String() string {
	if g.Driver == postgresDriver {
		u := url.URL{
			Scheme: "postgres",
			Host:   buildDbUrlHost(g.Host, g.Port),
		}
		if g.Database != nil {
			u.Path = *g.Database
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
		address := fmt.Sprintf("(%s)", buildDbUrlHost(g.Host, g.Port))

		// User info
		userInfo := url.UserPassword(g.User, g.Pass).String()

		// Base DSN
		dsn := fmt.Sprintf("%s@%s%s", userInfo, protocol, address)
		if g.Database != nil {
			dsn = fmt.Sprintf("%s/%s", dsn, *g.Database)
		}

		// Append query parameters if any
		if len(g.QueryParams) > 0 {
			query := g.QueryParams.Encode()
			dsn += "?" + query
		}
		return dsn
	}
	if g.Driver == mssqlDriver {
		u := url.URL{
			Scheme: mssqlDriver,
			Host:   buildDbUrlHost(g.Host, g.Port),
		}
		if g.Database != nil {
			u.Path = *g.Database
		}
		// Add user info
		if g.User != "" || g.Pass != "" {
			u.User = url.UserPassword(g.User, g.Pass)
		}
		u.RawQuery = g.QueryParams.Encode()
		return u.String()
	}
	return ""
}

func buildDbUrlHost(host string, port *int32) string {
	if port != nil {
		return fmt.Sprintf("%s:%d", host, *port)
	}
	return host
}

func getGeneralDbConnectionConfigFromMysql(config *mgmtv1alpha1.ConnectionConfig_MysqlConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
	switch cc := config.MysqlConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		query := url.Values{}
		if connectionTimeout != nil {
			query.Add("timeout", fmt.Sprintf("%ds", *connectionTimeout))
		}
		query.Add("multiStatements", "true")
		return &GeneralDbConnectConfig{
			Driver:      mysqlDriver,
			Host:        cc.Connection.Host,
			Port:        &cc.Connection.Port,
			Database:    &cc.Connection.Name,
			User:        cc.Connection.User,
			Pass:        cc.Connection.Pass,
			Protocol:    &cc.Connection.Protocol,
			QueryParams: query,
		}, nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		// follows the format [scheme://][user[:password]@]<host[:port]|socket>[/schema][?option=value&option=value...]
		// from the format - https://dev.mysql.com/doc/dev/mysqlsh-api-javascript/8.0/classmysqlsh_1_1_shell.html#a639614cf6b980f0d5267cc7057b81012

		u, err := url.Parse(cc.Url)
		if err != nil {
			return nil, err
		}

		// mysqlx is a newer connection protocol meant for more flexible schemas and supports mysqls nosql db capabilities
		// more information here - https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html

		if u.Scheme != "mysql" && u.Scheme != "mysqlx" {
			return nil, fmt.Errorf("scheme is not mysql ,unsupported scheme: %s", u.Scheme)
		}

		var user string
		var pass string

		if u.User != nil {
			user = u.User.Username()
			pass, _ = u.User.Password()
		}

		port := int32(3306)
		if p := u.Port(); p != "" {
			portInt, err := strconv.Atoi(p)
			if err != nil {
				return nil, err
			}

			// #nosec G109
			// this throws a linter error due to strconv.Atoi conversion above from string -> int32
			// mysql ports are unsigned 16-bit numbers so they should never overflow in an in32
			// https://stackoverflow.com/questions/20379491/what-is-the-optimal-way-to-store-port-numbers-in-a-mysql-database#:~:text=Port%20number%20is%20an%20unsinged,highest%20value%20can%20be%2065535.
			// https://downloads.mysql.com/docs/mysql-port-reference-en.pdf
			port = int32(portInt)
		}

		database := strings.TrimPrefix(u.Path, "/")

		query := u.Query()
		if connectionTimeout != nil {
			query.Add("timeout", fmt.Sprintf("%ds", *connectionTimeout))
		}
		query.Add("multiStatements", "true")

		return &GeneralDbConnectConfig{
			Driver:      u.Scheme,
			Host:        u.Hostname(),
			Port:        &port,
			Database:    &database,
			User:        user,
			Pass:        pass,
			Protocol:    nil,
			QueryParams: query,
		}, nil
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
		if config.PgConfig.GetClientTls() != nil {
			filenames := clienttls.GetClientTlsFileNames(config.PgConfig.GetClientTls())
			if filenames.RootCert != nil {
				query.Add("sslrootcert", *filenames.RootCert)
			}
			if filenames.ClientCert != nil && filenames.ClientKey != nil {
				query.Add("sslcert", *filenames.ClientCert)
				query.Add("sslkey", *filenames.ClientKey)
			}
		}
		return &GeneralDbConnectConfig{
			Driver:      postgresDriver,
			Host:        cc.Connection.Host,
			Port:        &cc.Connection.Port,
			Database:    &cc.Connection.Name,
			User:        cc.Connection.User,
			Pass:        cc.Connection.Pass,
			QueryParams: query,
		}, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		u, err := url.Parse(cc.Url)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return nil, fmt.Errorf("unable to parse postgres url [%s]: %w", urlErr.Op, urlErr.Err)
			}
			return nil, fmt.Errorf("unable to parse postgres url: %w", err)
		}

		user := u.User.Username()
		pass, ok := u.User.Password()
		if !ok {
			return nil, errors.New("unable to get password for pg string")
		}

		host, portStr := u.Hostname(), u.Port()

		var port int64
		if portStr != "" {
			port, err = strconv.ParseInt(portStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %w", err)
			}
		} else {
			// default to standard postgres port 5432 if port not provided
			port = int64(5432)
		}
		query := u.Query()
		if config.PgConfig.GetClientTls() != nil {
			filenames := clienttls.GetClientTlsFileNames(config.PgConfig.GetClientTls())
			if filenames.RootCert != nil {
				query.Add("sslrootcert", *filenames.RootCert)
			}
			if filenames.ClientCert != nil && filenames.ClientKey != nil {
				query.Add("sslcert", *filenames.ClientCert)
				query.Add("sslkey", *filenames.ClientKey)
			}
		}
		return &GeneralDbConnectConfig{
			Driver:      postgresDriver,
			Host:        host,
			Port:        shared.Ptr(int32(port)),
			Database:    shared.Ptr(strings.TrimPrefix(u.Path, "/")),
			User:        user,
			Pass:        pass,
			QueryParams: query,
		}, nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}

func getGeneralDbConnectionConfigFromMssql(config *mgmtv1alpha1.ConnectionConfig_MssqlConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
	switch cc := config.MssqlConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MssqlConnectionConfig_Url:
		u, err := url.Parse(cc.Url)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return nil, fmt.Errorf("unable to parse mssql url [%s]: %w", urlErr.Op, urlErr.Err)
			}
			return nil, fmt.Errorf("unable to parse mssql url: %w", err)
		}
		user := u.User.Username()
		pass, _ := u.User.Password()

		host, portStr := u.Hostname(), u.Port()

		query := u.Query()

		var port *int32
		if portStr != "" {
			parsedPort, err := strconv.ParseInt(portStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid port when processing mssql connection url: %w", err)
			}
			port = shared.Ptr(int32(parsedPort))
		}

		var instance *string
		if u.Path != "" {
			trimmed := strings.TrimPrefix(u.Path, "/")
			if trimmed != "" {
				instance = &trimmed
			}
		}

		if connectionTimeout != nil {
			query.Add("connection timeout", fmt.Sprintf("%d", *connectionTimeout))
		}

		return &GeneralDbConnectConfig{
			Driver:      mssqlDriver,
			Host:        host,
			Port:        port,
			Database:    instance,
			User:        user,
			Pass:        pass,
			QueryParams: query,
		}, nil
	default:
		return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("must provide valid mssql connection: %T", cc))
	}
}
