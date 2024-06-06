package sqlconnect

import (
	"log/slog"
	"net/url"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/zeebo/assert"
)

var (
	pgconnection = &mgmtv1alpha1.PostgresConnection{
		Host:    "localhost",
		Port:    5432,
		Name:    "postgres",
		User:    "test-user",
		Pass:    "test-pass",
		SslMode: ptr("verify"),
	}

	mysqlconnection = &mgmtv1alpha1.MysqlConnection{
		Host:     "localhost",
		Port:     3309,
		Name:     "mydb",
		User:     "test-user",
		Pass:     "test-pass",
		Protocol: "tcp",
	}
)

func Test_NewDbFromConnectionConfig(t *testing.T) {
	c := &SqlOpenConnector{}
	sqldb, err := c.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
			MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
				ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
					Connection: mysqlconnection,
				},
			},
		},
	}, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, sqldb)
}

func Test_NewDbFromConnectionConfig_BadConfig(t *testing.T) {
	c := &SqlOpenConnector{}
	sqldb, err := c.NewDbFromConnectionConfig(nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, sqldb)
}

func Test_NewPgPoolFromConnectionConfig(t *testing.T) {
	c := &SqlOpenConnector{}
	sqldb, err := c.NewPgPoolFromConnectionConfig(&mgmtv1alpha1.PostgresConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
			Connection: pgconnection,
		},
	}, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, sqldb)
}

func Test_NewPgPoolFromConnectionConfig_BadConfig(t *testing.T) {
	c := &SqlOpenConnector{}
	sqldb, err := c.NewPgPoolFromConnectionConfig(nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, sqldb)
}

func Test_getGeneralDbConnectConfigFromPg_Connection(t *testing.T) {
	out, err := getGeneralDbConnectConfigFromPg(&mgmtv1alpha1.ConnectionConfig_PgConfig{
		PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
				Connection: pgconnection,
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		Driver:      "postgres",
		Host:        "localhost",
		Port:        5432,
		Database:    "postgres",
		User:        "test-user",
		Pass:        "test-pass",
		Protocol:    nil,
		QueryParams: url.Values{"sslmode": []string{"verify"}, "connect_timeout": []string{"5"}},
	})
}

func Test_getGeneralDbConnectConfigFromPg_Url(t *testing.T) {
	out, err := getGeneralDbConnectConfigFromPg(&mgmtv1alpha1.ConnectionConfig_PgConfig{
		PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
				Url: "postgres://test-user:test-pass@localhost:5432/postgres?sslmode=verify&connect_timeout=5",
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		Driver:      "postgres",
		Host:        "localhost",
		Port:        5432,
		Database:    "postgres",
		User:        "test-user",
		Pass:        "test-pass",
		Protocol:    nil,
		QueryParams: url.Values{"sslmode": []string{"verify"}, "connect_timeout": []string{"5"}},
	})
}

func Test_getGeneralDbConnectionConfigFromMysql_Connection(t *testing.T) {
	out, err := getGeneralDbConnectionConfigFromMysql(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
				Connection: mysqlconnection,
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		Driver:      "mysql",
		Host:        "localhost",
		Port:        3309,
		Database:    "mydb",
		User:        "test-user",
		Pass:        "test-pass",
		Protocol:    ptr("tcp"),
		QueryParams: url.Values{"timeout": []string{"5s"}, "multiStatements": []string{"true"}},
	})
}

func Test_getGeneralDbConnectionConfigFromMysql_Url_mysql(t *testing.T) {
	out, err := getGeneralDbConnectionConfigFromMysql(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysql://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		Driver:      "mysql",
		Host:        "localhost",
		Port:        3306,
		Database:    "mydatabase",
		User:        "myuser",
		Pass:        "mypassword",
		Protocol:    nil,
		QueryParams: url.Values{"ssl": []string{"true"}},
	})
}
func Test_getGeneralDbConnectionConfigFromMysql_Url_mysqlx(t *testing.T) {
	out, err := getGeneralDbConnectionConfigFromMysql(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysqlx://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		Driver:      "mysqlx",
		Host:        "localhost",
		Port:        3306,
		Database:    "mydatabase",
		User:        "myuser",
		Pass:        "mypassword",
		Protocol:    nil,
		QueryParams: url.Values{"ssl": []string{"true"}},
	})
}

func Test_getGeneralDbConnectionConfigFromMysql_Url_Error(t *testing.T) {
	_, err := getGeneralDbConnectionConfigFromMysql(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysql://myuser:mypassword/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.Error(t, err)
}

func Test_getGeneralDbConnectionConfigFromMysql_Url_NoScheme(t *testing.T) {
	_, err := getGeneralDbConnectionConfigFromMysql(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysqlxxx://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.Error(t, err)
}

func Test_getGeneralDbConnectionConfigFromMysql_Url_NoPort(t *testing.T) {
	_, err := getGeneralDbConnectionConfigFromMysql(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysqlxxx://myuser:mypassword@localhost/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.Error(t, err)
}

func Test_GeneralDbConnectionConfig_String(t *testing.T) {
	type testcase struct {
		name     string
		input    GeneralDbConnectConfig
		expected string
	}
	testcases := []testcase{
		{
			name:     "empty",
			input:    GeneralDbConnectConfig{},
			expected: "",
		},
		{
			name: "postgres",
			input: GeneralDbConnectConfig{
				Driver:      "postgres",
				Host:        "localhost",
				Port:        5432,
				Database:    "mydb",
				User:        "test-user",
				Pass:        "test-pass",
				QueryParams: url.Values{"sslmode": []string{"verify"}},
			},
			expected: "postgres://test-user:test-pass@localhost:5432/mydb?sslmode=verify",
		},
		{
			name: "mysql",
			input: GeneralDbConnectConfig{
				Driver:      "mysql",
				Host:        "localhost",
				Port:        3309,
				Database:    "mydb",
				User:        "test-user",
				Pass:        "test-pass",
				Protocol:    ptr("tcp"),
				QueryParams: url.Values{"foo": []string{"bar"}},
			},
			expected: "test-user:test-pass@tcp(localhost:3309)/mydb?foo=bar",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.input.String(), tc.expected)
		})
	}
}

func Test_getConnectionDetails_Pg_NoTunnel(t *testing.T) {
	out, err := GetConnectionDetails(
		&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: pgconnection,
					},
				},
			},
		},
		ptr(uint32(5)),
		nil,
		slog.Default(),
	)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.NotNil(t, out.GeneralDbConnectConfig)
	assert.Nil(t, out.Tunnel)
}

func Test_getConnectionDetails_Pg_Tunnel(t *testing.T) {
	out, err := GetConnectionDetails(
		&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: pgconnection,
					},
					Tunnel: &mgmtv1alpha1.SSHTunnel{
						Host:               "bastion.neosync.dev",
						Port:               22,
						User:               "testuser",
						Authentication:     nil,
						KnownHostPublicKey: nil,
					},
				},
			},
		},
		ptr(uint32(5)),
		nil,
		slog.Default(),
	)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.NotNil(t, out.GeneralDbConnectConfig)
	assert.NotNil(t, out.Tunnel)
	assert.Equal(t, out.GeneralDbConnectConfig.Host, "localhost")
	assert.Equal(t, out.GeneralDbConnectConfig.Port, 0)
}

func Test_getConnectionDetails_Mysql_NoTunnel(t *testing.T) {
	out, err := GetConnectionDetails(
		&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
						Connection: mysqlconnection,
					},
				},
			},
		},
		ptr(uint32(5)),
		nil,
		slog.Default(),
	)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.NotNil(t, out.GeneralDbConnectConfig)
	assert.Nil(t, out.Tunnel)
}

func Test_getConnectionDetails_Mysql_Tunnel(t *testing.T) {
	out, err := GetConnectionDetails(
		&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
						Connection: mysqlconnection,
					},
					Tunnel: &mgmtv1alpha1.SSHTunnel{
						Host:               "bastion.neosync.dev",
						Port:               22,
						User:               "testuser",
						Authentication:     nil,
						KnownHostPublicKey: nil,
					},
				},
			},
		},
		ptr(uint32(5)),
		nil,
		slog.Default(),
	)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.NotNil(t, out.GeneralDbConnectConfig)
	assert.NotNil(t, out.Tunnel)
	assert.Equal(t, out.GeneralDbConnectConfig.Host, "localhost")
	assert.Equal(t, out.GeneralDbConnectConfig.Port, 0)
}

func ptr[T any](val T) *T {
	return &val
}
