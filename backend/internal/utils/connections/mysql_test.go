package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetMysqlUrl(t *testing.T) {
	assert.Equal(t, GetMysqlUrl(&MysqlConnectConfig{}), ":@(:0)/", "should not fail in the default state")

	assert.Equal(t, GetMysqlUrl(&MysqlConnectConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "neosync",
		Username: "myuser",
		Password: "mypass",
		Protocol: "tcp",
	}),
		"myuser:mypass@tcp(localhost:5432)/neosync",
	)
}
