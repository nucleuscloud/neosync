package sqlconnect

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
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

	details, err := GetConnectionDetails(connectionConfig, connectionTimeout, UpsertCLientTlsFiles, logger)
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
	}, connectionTimeout, UpsertCLientTlsFiles, logger)
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

type ClientCertConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
	localhost      = "localhost"
	randomPort     = 0
)

type ClientTlsFileConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}

func UpsertCLientTlsFiles(config *mgmtv1alpha1.ClientTlsConfig) (*ClientTlsFileConfig, error) {
	if config == nil {
		return nil, errors.New("config was nil")
	}

	errgrp := errgroup.Group{}

	filenames := getClientTlsFileNames(config)

	errgrp.Go(func() error {
		if filenames.RootCert == nil {
			return nil
		}
		_, err := os.Stat(*filenames.RootCert)
		if err != nil && !os.IsNotExist(err) {
			return err
		} else if err != nil && os.IsNotExist(err) {
			if err := os.WriteFile(*filenames.RootCert, []byte(config.GetRootCert()), 0600); err != nil {
				return err
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		if filenames.ClientCert != nil && filenames.ClientKey != nil {
			_, err := os.Stat(*filenames.ClientKey)
			if err != nil && !os.IsNotExist(err) {
				return err
			} else if err != nil && os.IsNotExist(err) {
				if err := os.WriteFile(*filenames.ClientKey, []byte(config.GetClientKey()), 0600); err != nil {
					return err
				}
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		if filenames.ClientCert != nil && filenames.ClientKey != nil {
			_, err := os.Stat(*filenames.ClientCert)
			if err != nil && !os.IsNotExist(err) {
				return err
			} else if err != nil && os.IsNotExist(err) {
				if err := os.WriteFile(*filenames.ClientCert, []byte(config.GetClientCert()), 0600); err != nil {
					return err
				}
			}
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return &filenames, nil
}

func getClientTlsFileNames(config *mgmtv1alpha1.ClientTlsConfig) ClientTlsFileConfig {
	if config == nil {
		return ClientTlsFileConfig{}
	}

	basedir := os.TempDir()

	output := ClientTlsFileConfig{}
	if config.GetRootCert() != "" {
		content := hashContent(config.GetRootCert())
		fullpath := filepath.Join(basedir, content)
		output.RootCert = &fullpath
	}
	if config.GetClientCert() != "" && config.GetClientKey() != "" {
		certContent := hashContent(config.GetClientCert())
		certpath := filepath.Join(basedir, certContent)
		keyContent := hashContent(config.GetClientKey())
		keypath := filepath.Join(basedir, keyContent)
		output.ClientCert = &certpath
		output.ClientKey = &keypath
	}
	return output
}

func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// Method for retrieving connection details, including tunneling information.
// Only use if requiring direct access to the SSH Tunnel, otherwise the SqlConnector should be used instead.
func GetConnectionDetails(
	c *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	handleClientTlsConfig func(config *mgmtv1alpha1.ClientTlsConfig) (*ClientTlsFileConfig, error),
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
			)
			connDetails, err := getGeneralDbConnectConfigFromPg(config, connectionTimeout)
			if err != nil {
				return nil, err
			}
			portValue := int32(randomPort)
			connDetails.Host = localhost
			connDetails.Port = portValue
			return &ConnectionDetails{
				Tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
				MaxConnectionLimit:     maxConnLimit,
			}, nil
		}

		if config.PgConfig.GetClientTls() != nil {
			_, err := handleClientTlsConfig(config.PgConfig.GetClientTls())
			if err != nil {
				return nil, err
			}
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
			)

			connDetails, err := getGeneralDbConnectionConfigFromMysql(config, connectionTimeout)
			if err != nil {
				return nil, err
			}

			portValue := int32(randomPort)
			connDetails.Host = localhost
			connDetails.Port = portValue
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
		return nil, nucleuserrors.NewBadRequest("must provide valid mysql connection")
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
		query.Add("multiStatements", "true")
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
			Port:        port,
			Database:    database,
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
			filenames := getClientTlsFileNames(config.PgConfig.GetClientTls())
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
			Port:        cc.Connection.Port,
			Database:    cc.Connection.Name,
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
			filenames := getClientTlsFileNames(config.PgConfig.GetClientTls())
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
			Port:        int32(port),
			Database:    strings.TrimPrefix(u.Path, "/"),
			User:        user,
			Pass:        pass,
			QueryParams: query,
		}, nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}
