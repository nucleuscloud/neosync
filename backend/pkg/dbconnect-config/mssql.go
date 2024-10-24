package dbconnectconfig

import (
	"errors"
	"fmt"
	"net/url"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
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
	case *mgmtv1alpha1.MssqlConnectionConfig_Url:
		uriconfig, err := url.Parse(cc.Url)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return nil, fmt.Errorf("unable to parse mssql url [%s]: %w", urlErr.Op, urlErr.Err)
			}
			return nil, fmt.Errorf("unable to parse mssql url: %w", err)
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
