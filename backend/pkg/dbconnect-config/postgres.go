package dbconnectconfig

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func NewFromPostgresConnection(config *mgmtv1alpha1.ConnectionConfig_PgConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
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
			driver:      postgresDriver,
			host:        cc.Connection.Host,
			port:        &cc.Connection.Port,
			Database:    &cc.Connection.Name,
			User:        cc.Connection.User,
			Pass:        cc.Connection.Pass,
			queryParams: query,
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
			driver:      postgresDriver,
			host:        host,
			port:        shared.Ptr(int32(port)), //nolint:gosec // Ignoring for now
			Database:    shared.Ptr(strings.TrimPrefix(u.Path, "/")),
			User:        user,
			Pass:        pass,
			queryParams: query,
		}, nil
	default:
		return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
	}
}
