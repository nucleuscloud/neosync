package sqlconnect

import (
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

	mssqlconnection = "sqlserver://sa:YourStrong@Passw0rd@localhost:1433?database=master"

	tunnel = &mgmtv1alpha1.SSHTunnel{
		Host:               "localhost",
		Port:               2222,
		User:               "foo",
		KnownHostPublicKey: nil,
		Authentication: &mgmtv1alpha1.SSHAuthentication{
			AuthConfig: &mgmtv1alpha1.SSHAuthentication_Passphrase{
				Passphrase: &mgmtv1alpha1.SSHPassphrase{
					Value: "foo",
				},
			},
		},
	}
)

func Test_NewDbFromConnectionConfig(t *testing.T) {
	connector := &SqlOpenConnector{}
	t.Run("mysql", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
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
	})

	t.Run("mysql tunnel", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
						Connection: mysqlconnection,
					},
					Tunnel: tunnel,
				},
			},
		}, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, sqldb)
	})

	t.Run("pg", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: pgconnection,
					},
				},
			},
		}, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, sqldb)
	})

	t.Run("pg tunnel", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: pgconnection,
					},
					Tunnel: tunnel,
				},
			},
		}, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, sqldb)
	})

	t.Run("mssql", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
				MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
						Url: mssqlconnection,
					},
				},
			},
		}, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, sqldb)
	})

	t.Run("mssql tunnel", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
				MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
						Url: mssqlconnection,
					},
					Tunnel: tunnel,
				},
			},
		}, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, sqldb)
	})

	t.Run("invalid", func(t *testing.T) {
		sqldb, err := connector.NewDbFromConnectionConfig(nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, sqldb)
	})
}

func ptr[T any](val T) *T {
	return &val
}
