package dbconnectconfig

import (
	"io"
	"log/slog"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

var (
	mysqlconnectionFixture = &mgmtv1alpha1.MysqlConnection{
		Host:     "localhost",
		Port:     3309,
		Name:     "mydb",
		User:     "test-user",
		Pass:     "test-pass",
		Protocol: "tcp",
	}
	discardLogger         = slog.New(slog.NewTextHandler(io.Discard, nil))
	testConnectionTimeout = uint32(5)
)

func Test_NewFromMysqlConnection(t *testing.T) {
	t.Run("Connection", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
							Connection: mysqlconnectionFixture,
						},
					},
				},
				&testConnectionTimeout,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:test-pass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
		})
		t.Run("ok_no_timeout", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
							Connection: mysqlconnectionFixture,
						},
					},
				},
				nil,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:test-pass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
		})
	})

	t.Run("URL_DSN", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
						},
					},
				},
				&testConnectionTimeout,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
		})
		t.Run("ok_no_timeout", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "test-user:testpass@tcp(localhost:3309)/mydb",
						},
					},
				},
				nil,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
		})
		t.Run("ok_specialchars_userpass", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "specialuser!*-:46!ZfMv3@Uh8*-<@@tcp(localhost:3309)/mydb",
						},
					},
				},
				nil,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"specialuser!*-:46!ZfMv3@Uh8*-<@@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
		})
	})

	t.Run("URL_URI", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "mysql://test-user:testpass@localhost:3309/mydb",
						},
					},
				},
				&testConnectionTimeout,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
		})
		t.Run("ok_no_timeout", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "mysql://test-user:testpass@localhost:3309/mydb",
						},
					},
				},
				nil,
				discardLogger,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
		})
	})
}

// func Test_NewFromMysqlConnection_Url_mysql(t *testing.T) {
// 	out, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
// 				Url: "mysql://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.Equal(t, out, &GeneralDbConnectConfig{
// 		driver:        "mysql",
// 		host:          "localhost",
// 		port:          ptr(int32(3306)),
// 		database:      ptr("mydatabase"),
// 		user:          "myuser",
// 		pass:          "mypassword",
// 		mysqlProtocol: nil,
// 		queryParams:   url.Values{"ssl": []string{"true"}, "multiStatements": []string{"true"}, "timeout": []string{"5s"}, "parseTime": []string{"true"}},
// 	})
// }
// func Test_NewFromMysqlConnection_Url_mysqlx(t *testing.T) {
// 	out, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
// 				Url: "mysqlx://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.Equal(t, out, &GeneralDbConnectConfig{
// 		driver:        "mysqlx",
// 		host:          "localhost",
// 		port:          ptr(int32(3306)),
// 		database:      ptr("mydatabase"),
// 		user:          "myuser",
// 		pass:          "mypassword",
// 		mysqlProtocol: nil,
// 		queryParams:   url.Values{"ssl": []string{"true"}, "multiStatements": []string{"true"}, "timeout": []string{"5s"}, "parseTime": []string{"true"}},
// 	})
// }

// func Test_NewFromMysqlConnection_Url_Error(t *testing.T) {
// 	_, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
// 				Url: "mysql://myuser:mypassword/mydatabase?ssl=true",
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.Error(t, err)
// }

// func Test_NewFromMysqlConnection_Url_NoScheme(t *testing.T) {
// 	_, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
// 				Url: "mysqlxxx://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.Error(t, err)
// }

// func Test_NewFromMysqlConnection_Url_NoPort(t *testing.T) {
// 	_, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
// 		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
// 				Url: "mysqlxxx://myuser:mypassword@localhost/mydatabase?ssl=true",
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.Error(t, err)
// }
