package dbconnectconfig

import (
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

func Test_NewFromPostgresConnection(t *testing.T) {
	t.Run("Connection", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: pgconnectionFixture,
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
				"postgres://test-user:test-pass@localhost:5432/postgres?connect_timeout=5&sslmode=verify",
				actual.String(),
			)
		})
		t.Run("ok_no_timeout", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: pgconnectionFixture,
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
				"postgres://test-user:test-pass@localhost:5432/postgres?sslmode=verify",
				actual.String(),
			)
		})
		t.Run("ok_no_port", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: &mgmtv1alpha1.PostgresConnection{
								Host: "localhost",
								// Port:    5432,
								Name:    "postgres",
								User:    "test-user",
								Pass:    "test-pass",
								SslMode: ptr("verify"),
							},
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
				"postgres://test-user:test-pass@localhost/postgres?connect_timeout=5&sslmode=verify",
				actual.String(),
			)
		})
		t.Run("ok_no_pass", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: &mgmtv1alpha1.PostgresConnection{
								Host: "localhost",
								Port: 5432,
								Name: "postgres",
								User: "test-user",
								// Pass:    "test-pass",
								SslMode: ptr("verify"),
							},
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
				"postgres://test-user@localhost:5432/postgres?connect_timeout=5&sslmode=verify",
				actual.String(),
			)
		})
		t.Run("ok_creds", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: &mgmtv1alpha1.PostgresConnection{
								Host: "localhost",
								Port: 5432,
								Name: "postgres",
								// User:    "test-user",
								// Pass:    "test-pass",
								SslMode: ptr("verify"),
							},
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
				"postgres://localhost:5432/postgres?connect_timeout=5&sslmode=verify",
				actual.String(),
			)
		})
	})

	t.Run("URL", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "postgres://test-user:test-pass@localhost:5432/postgres?sslmode=disable",
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
				"postgres://test-user:test-pass@localhost:5432/postgres?connect_timeout=5&sslmode=disable",
				actual.String(),
			)
		})
		t.Run("ok_no_timeout", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "postgres://test-user:test-pass@localhost:5432/postgres",
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
				"postgres://test-user:test-pass@localhost:5432/postgres",
				actual.String(),
			)
		})
		t.Run("ok_user_provided_timeout", func(t *testing.T) {
			actual, err := NewFromPostgresConnection(
				&mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "postgres://test-user:test-pass@localhost:5432/postgres?connect_timeout=10",
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
				"postgres://test-user:test-pass@localhost:5432/postgres?connect_timeout=10",
				actual.String(),
			)
		})
	})
}

// func Test_getGeneralDbConnectConfigFromPg_Connection(t *testing.T) {
// 	out, err := NewFromPostgresConnection(&mgmtv1alpha1.ConnectionConfig_PgConfig{
// 		PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
// 				Connection: pgconnectionFixture,
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.Equal(t, out, &GeneralDbConnectConfig{
// 		driver:        "postgres",
// 		host:          "localhost",
// 		port:          ptr(int32(5432)),
// 		database:      ptr("postgres"),
// 		user:          "test-user",
// 		pass:          "test-pass",
// 		mysqlProtocol: nil,
// 		queryParams:   url.Values{"sslmode": []string{"verify"}, "connect_timeout": []string{"5"}},
// 	})
// }

// func Test_getGeneralDbConnectConfigFromPg_Url(t *testing.T) {
// 	out, err := NewFromPostgresConnection(&mgmtv1alpha1.ConnectionConfig_PgConfig{
// 		PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 			ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
// 				Url: "postgres://test-user:test-pass@localhost:5432/postgres?sslmode=verify&connect_timeout=5",
// 			},
// 		},
// 	}, ptr(uint32(5)))

// 	assert.NoError(t, err)
// 	assert.NotNil(t, out)
// 	assert.Equal(t, out, &GeneralDbConnectConfig{
// 		driver:        "postgres",
// 		host:          "localhost",
// 		port:          ptr(int32(5432)),
// 		database:      ptr("postgres"),
// 		user:          "test-user",
// 		pass:          "test-pass",
// 		mysqlProtocol: nil,
// 		queryParams:   url.Values{"sslmode": []string{"verify"}, "connect_timeout": []string{"5"}},
// 	})
// }
