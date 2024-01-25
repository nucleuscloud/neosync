package connections

import (
	"fmt"
	"net/url"
)

type MysqlConnectConfig struct {
	Host              string
	Port              int32
	Database          string
	Username          string
	Password          string
	Protocol          string
	ConnectionTimeout *uint32
}

func GetMysqlUrl(cfg *MysqlConnectConfig) string {
	// Escape credentials properly
	userInfo := url.UserPassword(cfg.Username, cfg.Password).String()

	// Start constructing the DSN
	dsn := fmt.Sprintf(
		"%s@%s(%s:%d)/%s",
		userInfo,
		cfg.Protocol,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Add connection timeout if it's specified
	if cfg.ConnectionTimeout != nil {
		timeoutValue := fmt.Sprintf("?timeout=%ds", *cfg.ConnectionTimeout)
		dsn += timeoutValue
	}

	return dsn
}
