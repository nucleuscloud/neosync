package connections

import (
	"fmt"
	"net/url"
)

type PostgresConnectConfig struct {
	Host              string
	Port              int32
	Database          string
	User              string
	Pass              string
	SslMode           *string
	ConnectionTimeout *uint32
}

func GetPostgresUrl(cfg *PostgresConnectConfig) string {
	u := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}

	// Add user info
	if cfg.User != "" || cfg.Pass != "" {
		u.User = url.UserPassword(cfg.User, cfg.Pass)
	}

	// Build query parameters
	query := url.Values{}
	if cfg.SslMode != nil {
		query.Add("sslmode", *cfg.SslMode)
	}
	if cfg.ConnectionTimeout != nil {
		query.Add("connect_timeout", fmt.Sprintf("%d", *cfg.ConnectionTimeout))
	}
	u.RawQuery = query.Encode()

	return u.String()
}
