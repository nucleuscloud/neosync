package dbconnectconfig

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func NewFromMssqlConnection(config *mgmtv1alpha1.ConnectionConfig_MssqlConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
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
			port = shared.Ptr(int32(parsedPort)) //nolint:gosec // Ignoring for now
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
			driver:      mssqlDriver,
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
