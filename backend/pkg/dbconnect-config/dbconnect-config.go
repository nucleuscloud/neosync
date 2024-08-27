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

	host string
	port *int32
	// For mssql this is actually the path..the database is provided as a query parameter
	database *string
	user     string
	pass     string

	mysqlProtocol *string

	queryParams url.Values
}

func (g *GeneralDbConnectConfig) GetDriver() string {
	return g.driver
}

func (g *GeneralDbConnectConfig) SetPort(port int32) {
	g.port = &port
}
func (g *GeneralDbConnectConfig) SetHost(host string) {
	g.host = host
}

func (g *GeneralDbConnectConfig) GetPort() *int32 {
	return g.port
}
func (g *GeneralDbConnectConfig) GetHost() string {
	return g.host
}

func (g *GeneralDbConnectConfig) GetUser() string {
	return g.user
}

func (g *GeneralDbConnectConfig) String() string {
	if g.driver == postgresDriver {
		u := url.URL{
			Scheme: "postgres",
			Host:   buildDbUrlHost(g.host, g.port),
		}
		if g.database != nil {
			u.Path = *g.database
		}

		// Add user info
		if g.user != "" || g.pass != "" {
			u.User = url.UserPassword(g.user, g.pass)
		}
		u.RawQuery = g.queryParams.Encode()
		return u.String()
	}
	if g.driver == mysqlDriver {
		protocol := "tcp"
		if g.mysqlProtocol != nil {
			protocol = *g.mysqlProtocol
		}
		address := fmt.Sprintf("(%s)", buildDbUrlHost(g.host, g.port))

		// User info
		// dont use url.UserPassword as it escapes the password
		// host and password should not be escaped. even if they contain special characters
		userInfo := g.user
		if g.pass != "" {
			userInfo += ":" + g.pass
		}
		// Base DSN
		dsn := fmt.Sprintf("%s@%s%s", userInfo, protocol, address)
		if g.database != nil {
			dsn = fmt.Sprintf("%s/%s", dsn, *g.database)
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
			Host:   buildDbUrlHost(g.host, g.port),
		}
		if g.database != nil {
			u.Path = *g.database
		}
		// Add user info
		if g.user != "" || g.pass != "" {
			u.User = url.UserPassword(g.user, g.pass)
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
