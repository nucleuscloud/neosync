package connections

import "fmt"

type MysqlConnectConfig struct {
	Host     string
	Port     int32
	Database string
	Username string
	Password string
	Protocol string
}

func GetMysqlUrl(cfg *MysqlConnectConfig) string {
	return fmt.Sprintf(
		"%s:%s@%s(%s:%d)/%s",
		cfg.Username,
		cfg.Password,
		cfg.Protocol,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}
