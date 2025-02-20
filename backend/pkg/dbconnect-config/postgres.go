package dbconnectconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/spf13/viper"
)

const postgresScheme = "postgres"

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

		// For both postgres and pgx drivers, the URL scheme (protocol) should always be "postgres"
		pgurl := url.URL{
			Scheme: postgresScheme,
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
		if connectionTimeout != nil {
			query.Set("connect_timeout", fmt.Sprintf("%d", *connectionTimeout))
		}
		pgurl.RawQuery = query.Encode()

		return &pgConnectConfig{url: pgurl.String(), user: getUserFromInfo(pgurl.User)}, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url, *mgmtv1alpha1.PostgresConnectionConfig_UrlFromEnv:
		var pgurl string
		if config.PgConfig.GetUrl() != "" {
			pgurl = config.PgConfig.GetUrl()
		} else if config.PgConfig.GetUrlFromEnv() != "" {
			if !strings.HasPrefix(config.PgConfig.GetUrlFromEnv(), userDefinedEnvPrefix) {
				return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("to source a url from an environment variable, the variable must have a prefix of %s", userDefinedEnvPrefix))
			}
			pgurl = viper.GetString(config.PgConfig.GetUrlFromEnv())
		}

		uriconfig, err := GetPostgresUri(pgurl)
		if err != nil {
			return nil, err
		}
		query := uriconfig.Query()
		if !query.Has("connect_timeout") && connectionTimeout != nil {
			query.Set("connect_timeout", fmt.Sprintf("%d", *connectionTimeout))
		}
		uriconfig.RawQuery = query.Encode()
		return &pgConnectConfig{url: uriconfig.String(), user: getUserFromInfo(uriconfig.User)}, nil
	default:
		return nil, fmt.Errorf("unsupported pg connection config: %T", cc)
	}
}

func getUserFromInfo(u *url.Userinfo) string {
	if u == nil {
		return ""
	}
	return u.Username()
}

func GetPostgresUri(pgurl string) (*url.URL, error) {
	uriconfig, err := url.Parse(pgurl)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return nil, fmt.Errorf("unable to parse postgres url [%s]: %w", urlErr.Op, urlErr.Err)
		}
		return nil, fmt.Errorf("unable to parse postgres url: %w", err)
	}
	return uriconfig, nil
}
