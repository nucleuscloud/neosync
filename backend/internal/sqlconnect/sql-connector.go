package sqlconnect

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	"golang.org/x/crypto/ssh"
)

type SqlConnector interface {
	NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (*SqlDb, error)
	NewPgPoolFromConnectionConfig(pgconfig *mgmtv1alpha1.PostgresConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (*PgPool, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) NewDbFromConnectionConfig(connectionConfig *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (*SqlDb, error) {
	return &SqlDb{
		connectionConfig:  connectionConfig,
		logger:            logger,
		connectionTimeout: connectionTimeout,
	}, nil
}

func (rc *SqlOpenConnector) NewPgPoolFromConnectionConfig(pgconfig *mgmtv1alpha1.PostgresConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (*PgPool, error) {
	return &PgPool{
		connectionConfig:  pgconfig,
		logger:            logger,
		connectionTimeout: connectionTimeout,
	}, nil
}

type PgPool struct {
	pool *pgxpool.Pool

	connectionConfig *mgmtv1alpha1.PostgresConnectionConfig
	tunnel           *sshtunnel.Sshtunnel
	logger           *slog.Logger

	connectionTimeout *uint32
}

func (s *PgPool) Open(ctx context.Context) (*pgxpool.Pool, error) {
	details, err := getConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
			PgConfig: s.connectionConfig,
		},
	}, s.connectionTimeout, s.logger)
	if err != nil {
		return nil, err
	}
	if details.tunnel != nil {
		ready, err := details.tunnel.Start()
		if err != nil {
			return nil, err
		}
		<-ready
		newPort := int32(details.tunnel.Local.Port)
		details.GeneralDbConnectConfig.Port = newPort
		db, err := pgxpool.New(ctx, details.GeneralDbConnectConfig.String())
		if err != nil {
			return nil, err
		}
		s.pool = db
		s.tunnel = details.tunnel
		return db, nil
	}

	db, err := pgxpool.New(ctx, details.GeneralDbConnectConfig.String())
	if err != nil {
		return nil, err
	}
	s.pool = db
	return db, nil
}

func (s *PgPool) Close() {
	if s.pool == nil {
		return
	}
	db := s.pool
	s.pool = nil
	db.Close()
	if s.tunnel != nil {
		tunnel := s.tunnel
		s.tunnel = nil
		tunnel.Close()
	}
}

type SqlDb struct {
	db *sql.DB

	connectionConfig *mgmtv1alpha1.ConnectionConfig
	tunnel           *sshtunnel.Sshtunnel
	logger           *slog.Logger

	connectionTimeout *uint32
}

func (s *SqlDb) Open() (*sql.DB, error) {
	details, err := getConnectionDetails(s.connectionConfig, s.connectionTimeout, s.logger)
	if err != nil {
		return nil, err
	}
	if details.tunnel != nil {
		ready, err := details.tunnel.Start()
		if err != nil {
			return nil, err
		}
		<-ready

		newPort := int32(details.tunnel.Local.Port)
		details.GeneralDbConnectConfig.Port = newPort
		db, err := sql.Open(details.GeneralDbConnectConfig.Driver, details.GeneralDbConnectConfig.String())
		if err != nil {
			return nil, err
		}
		s.db = db
		s.tunnel = details.tunnel
		return db, nil
	}
	db, err := sql.Open(details.GeneralDbConnectConfig.Driver, details.GeneralDbConnectConfig.String())
	s.db = db
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *SqlDb) Close() error {
	if s.db == nil {
		return nil
	}
	db := s.db
	s.db = nil
	err := db.Close()
	if s.tunnel != nil {
		s.tunnel.Close()
		s.tunnel = nil
	}
	return err
}

type connectionDetails struct {
	// ConnectionString string
	GeneralDbConnectConfig

	tunnel *sshtunnel.Sshtunnel
}

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
)

func getConnectionDetails(c *mgmtv1alpha1.ConnectionConfig, connectionTimeout *uint32, logger *slog.Logger) (*connectionDetails, error) {

	switch config := c.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if config.PgConfig.Tunnel != nil {
			destination, err := getEndpointFromPgConnectionConfig(config)
			if err != nil {
				return nil, err
			}
			authmethod, err := getTunnelAuthMethodFromSshConfig(config.PgConfig.Tunnel.Authentication)
			if err != nil {
				return nil, err
			}
			tunnel := sshtunnel.New(
				sshtunnel.NewEndpointWithUser(config.PgConfig.Tunnel.GetHost(), int(config.PgConfig.Tunnel.GetPort()), config.PgConfig.Tunnel.GetUser()),
				authmethod,
				destination,
				sshtunnel.NewEndpoint("localhost", 0),
				1,
				logger,
			)
			connDetails, err := GetGeneralDbConnectConfigFromPg(config)
			if err != nil {
				return nil, err
			}
			connDetails.Host = "localhost"
			connDetails.Port = 0
			return &connectionDetails{
				tunnel:                 tunnel,
				GeneralDbConnectConfig: *connDetails,
			}, nil
		}

		connDetails, err := GetGeneralDbConnectConfigFromPg(config)
		if err != nil {
			return nil, err
		}
		return &connectionDetails{
			GeneralDbConnectConfig: *connDetails,
		}, nil

		// switch connectionConfig := config.PgConfig.ConnectionConfig.(type) {
		// case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		// 	connStr := connections.GetPostgresUrl(&connections.PostgresConnectConfig{
		// 		Host:              connectionConfig.Connection.Host,
		// 		Port:              connectionConfig.Connection.Port,
		// 		Database:          connectionConfig.Connection.Name,
		// 		User:              connectionConfig.Connection.User,
		// 		Pass:              connectionConfig.Connection.Pass,
		// 		SslMode:           connectionConfig.Connection.SslMode,
		// 		ConnectionTimeout: connectionTimeout,
		// 	})
		// 	connectionString = &connStr
		// case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		// 	connectionString = &connectionConfig.Url
		// default:
		// 	return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
		// }

		// TODO
	// case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
	// 	var connectionString *string
	// 	switch connectionConfig := config.MysqlConfig.ConnectionConfig.(type) {
	// 	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
	// 		connStr := connections.GetMysqlUrl(&connections.MysqlConnectConfig{
	// 			Host:              connectionConfig.Connection.Host,
	// 			Port:              connectionConfig.Connection.Port,
	// 			Database:          connectionConfig.Connection.Name,
	// 			Username:          connectionConfig.Connection.User,
	// 			Password:          connectionConfig.Connection.Pass,
	// 			Protocol:          connectionConfig.Connection.Protocol,
	// 			ConnectionTimeout: connectionTimeout,
	// 		})
	// 		connectionString = &connStr
	// 	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
	// 		connectionString = &connectionConfig.Url
	// 	default:
	// 		return nil, nucleuserrors.NewBadRequest("must provide valid mysql connection")
	// 	}
	// 	return &connectionDetails{ConnectionString: *connectionString, ConnectionDriver: mysqlDriver}, nil
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
		return getEndpointFromPostgresUrl(cc.Url)
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}

func getEndpointFromPostgresUrl(dsn string) (*sshtunnel.Endpoint, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	// Extract user info
	user := u.User.Username()

	// Extract host and port
	host, portStr := u.Hostname(), u.Port()

	// Convert port to integer
	var port int
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %v", err)
		}
	}

	return sshtunnel.NewEndpointWithUser(host, port, user), nil
}

type GeneralDbConnectConfig struct {
	Driver string

	Host     string
	Port     int32
	Database string
	User     string
	Pass     string

	QueryParams url.Values
}

func (g *GeneralDbConnectConfig) String() string {
	if g.Driver == "postgres" {
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
	return ""
}

func GetGeneralDbConnectConfigFromPg(config *mgmtv1alpha1.ConnectionConfig_PgConfig) (*GeneralDbConnectConfig, error) {
	switch cc := config.PgConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		query := url.Values{}
		if cc.Connection.SslMode != nil {
			query.Add("sslmode", *cc.Connection.SslMode)
		}
		// if cc.Connection..ConnectionTimeout != nil {
		// 	query.Add("connect_timeout", fmt.Sprintf("%d", *cfg.ConnectionTimeout))
		// }
		return &GeneralDbConnectConfig{
			Driver:      "postgres",
			Host:        cc.Connection.Host,
			Port:        cc.Connection.Port,
			Database:    cc.Connection.Name,
			User:        cc.Connection.User,
			Pass:        cc.Connection.Pass,
			QueryParams: url.Values{},
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
			Driver:      "postgres",
			Host:        host,
			Port:        int32(port),
			Database:    u.Path,
			User:        user,
			Pass:        pass,
			QueryParams: u.Query(),
		}, nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}
