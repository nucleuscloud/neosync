package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetPostgresUrl(t *testing.T) {
	assert.Equal(t, GetPostgresUrl(&PostgresConnectConfig{}), "postgres://:@:0/", "should not fail in the default state")

	sslMode := "disable"
	assert.Equal(t, GetPostgresUrl(&PostgresConnectConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "neosync",
		User:     "myuser",
		Pass:     "mypass",
		SslMode:  &sslMode,
	}),
		"postgres://myuser:mypass@localhost:5432/neosync?sslmode=disable",
	)
}
