package dbconnectconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

type pgConnectConfig struct {
	url  string
	user string
}

var _ DbConnectConfig = (*pgConnectConfig)(nil)

func (m *pgConnectConfig) String() string {
	return m.url
}
func (m *pgConnectConfig) GetUser() string {
	return m.user
}

func NewFromPostgresConnection(
	config *mgmtv1alpha1.ConnectionConfig_PgConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
) (DbConnectConfig, error) {
	switch cc := config.PgConfig.GetConnectionConfig().(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		host := cc.Connection.GetHost()
		if cc.Connection.GetPort() > 0 {
			host += fmt.Sprintf(":%d", cc.Connection.GetPort())
		}

		pgurl := url.URL{
			Scheme: sqlmanager_shared.DefaultPostgresDriver,
			Host:   host,
		}
		if cc.Connection.GetUser() != "" && cc.Connection.GetPass() != "" {
			pgurl.User = url.UserPassword(cc.Connection.GetUser(), cc.Connection.GetPass())
		} else if cc.Connection.GetUser() != "" && cc.Connection.GetPass() == "" {
			pgurl.User = url.User(cc.Connection.GetUser())
		}
		if cc.Connection.GetName() != "" {
			pgurl.Path = cc.Connection.GetName()
		}
		query := url.Values{}
		if cc.Connection.GetSslMode() != "" {
			query.Set("sslmode", cc.Connection.GetSslMode())
		}
		// if config.PgConfig.GetClientTls() != nil {
		// 	query = setPgClientTlsQueryParams(query, config.PgConfig.GetClientTls())
		// }
		if connectionTimeout != nil {
			query.Set("connect_timeout", fmt.Sprintf("%d", *connectionTimeout))
		}
		pgurl.RawQuery = query.Encode()

		return &pgConnectConfig{url: pgurl.String(), user: getUserFromInfo(pgurl.User)}, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		pgurl := cc.Url

		uriconfig, err := url.Parse(pgurl)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return nil, fmt.Errorf("unable to parse postgres url [%s]: %w", urlErr.Op, urlErr.Err)
			}
			return nil, fmt.Errorf("unable to parse postgres url: %w", err)
		}
		query := uriconfig.Query()
		if !query.Has("connect_timeout") && connectionTimeout != nil {
			query.Set("connect_timeout", fmt.Sprintf("%d", *connectionTimeout))
		}
		// todo: move this out of here into the driver
		// if config.PgConfig.GetClientTls() != nil {
		// 	query = setPgClientTlsQueryParams(query, config.PgConfig.GetClientTls())
		// }
		uriconfig.RawQuery = query.Encode()
		return &pgConnectConfig{url: uriconfig.String(), user: getUserFromInfo(uriconfig.User)}, nil
	default:
		return nil, fmt.Errorf("unsupported pg connection config: %T", cc)
	}
}

func setPgClientTlsQueryParams(
	query url.Values,
	cfg *mgmtv1alpha1.ClientTlsConfig,
) url.Values {
	filenames := clienttls.GetClientTlsFileNames(cfg)
	if filenames.RootCert != nil {
		query.Set("sslrootcert", *filenames.RootCert)
	}
	if filenames.ClientCert != nil && filenames.ClientKey != nil {
		query.Set("sslcert", *filenames.ClientCert)
		query.Set("sslkey", *filenames.ClientKey)
	}
	return query
}

func getUserFromInfo(u *url.Userinfo) string {
	if u == nil {
		return ""
	}
	return u.Username()
}
