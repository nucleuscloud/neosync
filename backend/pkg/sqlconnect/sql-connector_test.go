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

// func Test_getConnectionDetails_Pg_NoTunnel(t *testing.T) {
// 	out, err := GetConnectionDetails(
// 		&mgmtv1alpha1.ConnectionConfig{
// 			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
// 				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
// 						Connection: pgconnection,
// 					},
// 				},
// 			},
// 		},
// 		ptr(uint32(5)),
// 		nil,
// 		slog.Default(),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.NotNil(t, out.GeneralDbConnectConfig)
// 	assert.Nil(t, out.Tunnel)
// }

// func Test_getConnectionDetails_Pg_Tunnel(t *testing.T) {
// 	out, err := GetConnectionDetails(
// 		&mgmtv1alpha1.ConnectionConfig{
// 			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
// 				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
// 						Connection: pgconnection,
// 					},
// 					Tunnel: &mgmtv1alpha1.SSHTunnel{
// 						Host:               "bastion.neosync.dev",
// 						Port:               22,
// 						User:               "testuser",
// 						Authentication:     nil,
// 						KnownHostPublicKey: nil,
// 					},
// 				},
// 			},
// 		},
// 		ptr(uint32(5)),
// 		nil,
// 		slog.Default(),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.NotNil(t, out.GeneralDbConnectConfig)
// 	assert.NotNil(t, out.Tunnel)
// 	assert.Equal(t, out.GeneralDbConnectConfig.GetHost(), "localhost")
// 	assert.Equal(t, *out.GeneralDbConnectConfig.GetPort(), 0)
// }

// func Test_getConnectionDetails_Mysql_NoTunnel(t *testing.T) {
// 	out, err := GetConnectionDetails(
// 		&mgmtv1alpha1.ConnectionConfig{
// 			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
// 						Connection: mysqlconnection,
// 					},
// 				},
// 			},
// 		},
// 		ptr(uint32(5)),
// 		nil,
// 		slog.Default(),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.NotNil(t, out.GeneralDbConnectConfig)
// 	assert.Nil(t, out.Tunnel)
// }

// func Test_getConnectionDetails_Mysql_Tunnel(t *testing.T) {
// 	out, err := GetConnectionDetails(
// 		&mgmtv1alpha1.ConnectionConfig{
// 			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 					ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
// 						Connection: mysqlconnection,
// 					},
// 					Tunnel: &mgmtv1alpha1.SSHTunnel{
// 						Host:               "bastion.neosync.dev",
// 						Port:               22,
// 						User:               "testuser",
// 						Authentication:     nil,
// 						KnownHostPublicKey: nil,
// 					},
// 				},
// 			},
// 		},
// 		ptr(uint32(5)),
// 		nil,
// 		slog.Default(),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.NotNil(t, out.GeneralDbConnectConfig)
// 	assert.NotNil(t, out.Tunnel)
// 	assert.Equal(t, out.GeneralDbConnectConfig.GetHost(), "localhost")
// 	assert.Equal(t, *out.GeneralDbConnectConfig.GetPort(), 0)
// }

// func Test_getConnectionDetails_Mssql_NoTunnel(t *testing.T) {
// 	out, err := GetConnectionDetails(
// 		&mgmtv1alpha1.ConnectionConfig{
// 			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
// 				MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
// 					ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
// 						Url: mssqlconnection,
// 					},
// 				},
// 			},
// 		},
// 		ptr(uint32(5)),
// 		nil,
// 		slog.Default(),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.NotNil(t, out.GeneralDbConnectConfig)
// 	assert.Nil(t, out.Tunnel)
// }

// func Test_getConnectionDetails_Mssql_Tunnel(t *testing.T) {
// 	out, err := GetConnectionDetails(
// 		&mgmtv1alpha1.ConnectionConfig{
// 			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
// 				MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
// 					ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
// 						Url: mssqlconnection,
// 					},
// 					Tunnel: &mgmtv1alpha1.SSHTunnel{
// 						Host:               "bastion.neosync.dev",
// 						Port:               22,
// 						User:               "testuser",
// 						Authentication:     nil,
// 						KnownHostPublicKey: nil,
// 					},
// 				},
// 			},
// 		},
// 		ptr(uint32(5)),
// 		nil,
// 		slog.Default(),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.NotNil(t, out.GeneralDbConnectConfig)
// 	assert.NotNil(t, out.Tunnel)
// 	assert.Equal(t, out.GeneralDbConnectConfig.GetHost(), "localhost")
// 	assert.Equal(t, *out.GeneralDbConnectConfig.GetPort(), 0)
// }

func ptr[T any](val T) *T {
	return &val
}
