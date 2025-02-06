package dbconnectconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/spf13/viper"
)

const (
	userDefinedEnvPrefix = "USER_DEFINED_"
)

type mysqlConnectConfig struct {
	dsn  string
	user string
}

var _ DbConnectConfig = (*mysqlConnectConfig)(nil)

func (m *mysqlConnectConfig) String() string {
	return m.dsn
}
func (m *mysqlConnectConfig) GetUser() string {
	return m.user
}

func NewFromMysqlConnection(
	config *mgmtv1alpha1.ConnectionConfig_MysqlConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
	mysqlDisableParseTime bool,
) (DbConnectConfig, error) {
	parseTime := !mysqlDisableParseTime
	switch cc := config.MysqlConfig.GetConnectionConfig().(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		cfg := mysql.NewConfig()
		cfg.DBName = cc.Connection.GetName()
		cfg.Addr = cc.Connection.GetHost()
		if cc.Connection.GetPort() > 0 {
			cfg.Addr += fmt.Sprintf(":%d", cc.Connection.GetPort())
		}
		cfg.User = cc.Connection.GetUser()
		cfg.Passwd = cc.Connection.GetPass()
		if connectionTimeout != nil {
			cfg.Timeout = time.Duration(*connectionTimeout) * time.Second
		}
		cfg.Net = cc.Connection.GetProtocol()
		cfg.MultiStatements = true
		cfg.ParseTime = parseTime

		return &mysqlConnectConfig{dsn: cfg.FormatDSN(), user: cfg.User}, nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url, *mgmtv1alpha1.MysqlConnectionConfig_UrlFromEnv:
		var mysqlurl string
		if config.MysqlConfig.GetUrl() != "" {
			mysqlurl = config.MysqlConfig.GetUrl()
		} else if config.MysqlConfig.GetUrlFromEnv() != "" {
			if !strings.HasPrefix(config.MysqlConfig.GetUrlFromEnv(), userDefinedEnvPrefix) {
				return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("to source a url from an environment variable, the variable must have a prefix of %s", userDefinedEnvPrefix))
			}
			mysqlurl = viper.GetString(config.MysqlConfig.GetUrlFromEnv())
		}

		cfg, err := mysql.ParseDSN(mysqlurl)
		if err != nil {
			logger.Warn(fmt.Sprintf("failed to parse mysql url as DSN: %v", err))
			uriConfig, err := url.Parse(mysqlurl)
			if err != nil {
				var urlErr *url.Error
				if errors.As(err, &urlErr) {
					return nil, fmt.Errorf("unable to parse mysql url [%s]: %w", urlErr.Op, urlErr.Err)
				}
				return nil, fmt.Errorf("unable to parse mysql url: %w", err)
			}
			cfg = mysql.NewConfig()
			cfg.Net = "tcp"
			cfg.DBName = strings.TrimPrefix(uriConfig.Path, "/")
			cfg.Addr = uriConfig.Host
			cfg.User = uriConfig.User.Username()
			if passwd, ok := uriConfig.User.Password(); ok {
				cfg.Passwd = passwd
			}

			if connectionTimeout != nil {
				cfg.Timeout = time.Duration(*connectionTimeout) * time.Second
			}
			cfg.MultiStatements = true
			cfg.ParseTime = parseTime

			if uriConfig.RawQuery != "" {
				cfg.Params = make(map[string]string)
				for k, values := range uriConfig.Query() {
					for _, value := range values {
						cfg.Params[k] = value
					}
				}
			}
			return &mysqlConnectConfig{dsn: cfg.FormatDSN(), user: cfg.User}, nil
		}

		if cfg.Timeout == 0 && connectionTimeout != nil {
			cfg.Timeout = time.Duration(*connectionTimeout) * time.Second
		}
		cfg.MultiStatements = true
		cfg.ParseTime = parseTime
		return &mysqlConnectConfig{dsn: cfg.FormatDSN(), user: cfg.User}, nil
	default:
		return nil, fmt.Errorf("unsupported mysql connection config: %T", cc)
	}
}
