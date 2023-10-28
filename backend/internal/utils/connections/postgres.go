package connections

import (
	"fmt"
)

type PostgresConnectConfig struct {
	Host     string
	Port     int32
	Database string
	User     string
	Pass     string
	SslMode  *string
}

func GetPostgresUrl(cfg *PostgresConnectConfig) string {
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
