package dbconnectconfig

import (
	"fmt"
	"net/url"
)

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
	mssqlDriver    = "sqlserver"
)

type GeneralDbConnectConfig struct {
	driver string

	Host string
	Port *int32
	// For mssql this is actually the path..the database is provided as a query parameter
	Database *string
	User     string
	Pass     string

	Protocol *string

	queryParams url.Values
}

func (g *GeneralDbConnectConfig) GetDriver() string {
	return g.driver
}

func (g *GeneralDbConnectConfig) String() string {
	if g.driver == postgresDriver {
		u := url.URL{
			Scheme: "postgres",
			Host:   buildDbUrlHost(g.Host, g.Port),
		}
		if g.Database != nil {
			u.Path = *g.Database
		}

		// Add user info
		if g.User != "" || g.Pass != "" {
			u.User = url.UserPassword(g.User, g.Pass)
		}
		u.RawQuery = g.queryParams.Encode()
		return u.String()
	}
	if g.driver == mysqlDriver {
		protocol := "tcp"
		if g.Protocol != nil {
			protocol = *g.Protocol
		}
		address := fmt.Sprintf("(%s)", buildDbUrlHost(g.Host, g.Port))

		// User info
		// dont use url.UserPassword as it escapes the password
		// host and password should not be escaped. even if they contain special characters
		userInfo := g.User
		if g.Pass != "" {
			userInfo += ":" + g.Pass
		}
		// Base DSN
		dsn := fmt.Sprintf("%s@%s%s", userInfo, protocol, address)
		if g.Database != nil {
			dsn = fmt.Sprintf("%s/%s", dsn, *g.Database)
		}

		// Append query parameters if any
		if len(g.queryParams) > 0 {
			query := g.queryParams.Encode()
			dsn += "?" + query
		}
		return dsn
	}
	if g.driver == mssqlDriver {
		u := url.URL{
			Scheme: mssqlDriver,
			Host:   buildDbUrlHost(g.Host, g.Port),
		}
		if g.Database != nil {
			u.Path = *g.Database
		}
		// Add user info
		if g.User != "" || g.Pass != "" {
			u.User = url.UserPassword(g.User, g.Pass)
		}
		u.RawQuery = g.queryParams.Encode()
		return u.String()
	}
	return ""
}

func buildDbUrlHost(host string, port *int32) string {
	if port != nil {
		return fmt.Sprintf("%s:%d", host, *port)
	}
	return host
}
