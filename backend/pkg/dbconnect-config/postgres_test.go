package dbconnectconfig

import (
	"net/url"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

var (
	pgconnectionFixture = &mgmtv1alpha1.PostgresConnection{
		Host:    "localhost",
		Port:    5432,
		Name:    "postgres",
		User:    "test-user",
		Pass:    "test-pass",
		SslMode: ptr("verify"),
	}
)

func Test_getGeneralDbConnectConfigFromPg_Connection(t *testing.T) {
	out, err := NewFromPostgresConnection(&mgmtv1alpha1.ConnectionConfig_PgConfig{
		PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
				Connection: pgconnectionFixture,
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		driver:        "postgres",
		host:          "localhost",
		port:          ptr(int32(5432)),
		Database:      ptr("postgres"),
		User:          "test-user",
		Pass:          "test-pass",
		mysqlProtocol: nil,
		queryParams:   url.Values{"sslmode": []string{"verify"}, "connect_timeout": []string{"5"}},
	})
}

func Test_getGeneralDbConnectConfigFromPg_Url(t *testing.T) {
	out, err := NewFromPostgresConnection(&mgmtv1alpha1.ConnectionConfig_PgConfig{
		PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
				Url: "postgres://test-user:test-pass@localhost:5432/postgres?sslmode=verify&connect_timeout=5",
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		driver:        "postgres",
		host:          "localhost",
		port:          ptr(int32(5432)),
		Database:      ptr("postgres"),
		User:          "test-user",
		Pass:          "test-pass",
		mysqlProtocol: nil,
		queryParams:   url.Values{"sslmode": []string{"verify"}, "connect_timeout": []string{"5"}},
	})
}
