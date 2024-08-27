package dbconnectconfig

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

func NewFromMysqlConnection(config *mgmtv1alpha1.ConnectionConfig_MysqlConfig, connectionTimeout *uint32) (*GeneralDbConnectConfig, error) {
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
			port = int32(portInt) //nolint:gosec // Ignoring for now
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
