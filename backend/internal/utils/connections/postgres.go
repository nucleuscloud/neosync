package connections

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type ConnectConfig struct {
	Host     string
	Port     int32
	Database string
	User     string
	Pass     string
	SslMode  *string
}

func GetPostgresUrl(cfg *ConnectConfig) string {
	dburl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Pass,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	if cfg.SslMode != nil && *cfg.SslMode != "" {
		dburl = fmt.Sprintf("%s?sslmode=%s", dburl, *cfg.SslMode)
	}
	return dburl
}

func ParsePostgresUrl(connStr string) (*ConnectConfig, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, err
	}

	pass, _ := u.User.Password()
	host, port, _ := net.SplitHostPort(u.Host)
	m, _ := url.ParseQuery(u.RawQuery)

	portInt, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return nil, err
	}

	var sslmode *string
	_, ok := m["sslmode"]
	if ok {
		sslmode = &m["sslmode"][0]
	}

	return &ConnectConfig{
		User:     u.User.Username(),
		Pass:     pass,
		Host:     host,
		Port:     int32(portInt),
		Database: strings.Replace(u.Path, "/", "", 1),
		SslMode:  sslmode,
	}, nil
}
