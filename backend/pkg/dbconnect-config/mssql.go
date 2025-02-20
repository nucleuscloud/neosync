package dbconnectconfig

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/spf13/viper"
)

type mssqlConnectConfig struct {
	url  string
	user string
}

var _ DbConnectConfig = (*mssqlConnectConfig)(nil)

func (m *mssqlConnectConfig) String() string {
	return m.url
}
func (m *mssqlConnectConfig) GetUser() string {
	return m.user
}

func NewFromMssqlConnection(
	config *mgmtv1alpha1.ConnectionConfig_MssqlConfig,
	connectionTimeout *uint32,
) (DbConnectConfig, error) {
	switch cc := config.MssqlConfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MssqlConnectionConfig_Url, *mgmtv1alpha1.MssqlConnectionConfig_UrlFromEnv:
		var mssqlurl string
		if config.MssqlConfig.GetUrl() != "" {
			mssqlurl = config.MssqlConfig.GetUrl()
		} else if config.MssqlConfig.GetUrlFromEnv() != "" {
			if !strings.HasPrefix(config.MssqlConfig.GetUrlFromEnv(), userDefinedEnvPrefix) {
				return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("to source a url from an environment variable, the variable must have a prefix of %s", userDefinedEnvPrefix))
			}
			mssqlurl = viper.GetString(config.MssqlConfig.GetUrlFromEnv())
		}

		uriconfig, err := GetMssqlUri(mssqlurl)
		if err != nil {
			return nil, err
		}

		query := uriconfig.Query()

		if !query.Has("connection timeout") && connectionTimeout != nil {
			query.Add("connection timeout", fmt.Sprintf("%d", *connectionTimeout))
		}
		uriconfig.RawQuery = query.Encode()

		return &mssqlConnectConfig{url: uriconfig.String(), user: getUserFromInfo(uriconfig.User)}, nil
	default:
		return nil, nucleuserrors.NewBadRequest(fmt.Sprintf("must provide valid mssql connection: %T", cc))
	}
}

func GetMssqlUri(mssqlurl string) (*url.URL, error) {
	uriconfig, err := url.Parse(mssqlurl)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return nil, fmt.Errorf("unable to parse mssql url [%s]: %w", urlErr.Op, urlErr.Err)
		}
		return nil, fmt.Errorf("unable to parse mssql url: %w", err)
	}
	return uriconfig, nil
}
