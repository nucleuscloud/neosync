package dbconnectconfig

import (
	"fmt"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:test-pass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})

		t.Run("ok_disable_parse_time", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
							Connection: mysqlconnectionFixture,
						},
					},
				},
				&testConnectionTimeout,
				testutil.GetTestLogger(t),
				true,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:test-pass@tcp(localhost:3309)/mydb?multiStatements=true&timeout=5s",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:test-pass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"specialuser!*-:46!ZfMv3@Uh8*-<@@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
			assert.Equal(t, "specialuser!*-", actual.GetUser())
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})
		t.Run("ok_with_attributes", func(t *testing.T) {
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "mysql://test-user:testpass@localhost:3309/mydb?ssl-mode=preferred",
						},
					},
				},
				&testConnectionTimeout,
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s&ssl-mode=preferred",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
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
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})
	})

	t.Run("URL from Env", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			viper.Set(fmt.Sprintf("%s%s", userDefinedEnvPrefix, "MYSQL_URL"), "mysql://test-user:testpass@localhost:3309/mydb")
			actual, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_UrlFromEnv{
							UrlFromEnv: "USER_DEFINED_MYSQL_URL",
						},
					},
				},
				&testConnectionTimeout,
				testutil.GetTestLogger(t),
				false,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"test-user:testpass@tcp(localhost:3309)/mydb?multiStatements=true&parseTime=true&timeout=5s",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})
		t.Run("error_no_prefix", func(t *testing.T) {
			viper.Set("MYSQL_URL_NO_PREFIX", "mysql://test-user:testpass@localhost:3309/mydb")

			_, err := NewFromMysqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_UrlFromEnv{
							UrlFromEnv: "MYSQL_URL_NO_PREFIX",
						},
					},
				},
				&testConnectionTimeout,
				testutil.GetTestLogger(t),
				false,
			)
			require.Error(t, err)
		})
	})
}
