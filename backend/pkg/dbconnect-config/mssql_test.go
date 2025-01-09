package dbconnectconfig

import (
	"fmt"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
)

func Test_NewFromMssqlConnection(t *testing.T) {

	t.Run("URL from Env", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			viper.Set(fmt.Sprintf("%s%s", userDefinedEnvPrefix, "MSSQL_URL"), "sqlserver://test-user:test-pass@localhost:1433/myinstance?database=master")
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_UrlFromEnv{
							UrlFromEnv: "USER_DEFINED_MSSQL_URL",
						},
					},
				},
				&testConnectionTimeout,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, "sqlserver://test-user:test-pass@localhost:1433/myinstance?connection+timeout=5&database=master", actual.String())
			assert.Equal(t, "test-user", actual.GetUser())
		})
		t.Run("error_no_prefix", func(t *testing.T) {
			viper.Set("MSSQL_URL_NO_PREFIX", "sqlserver://test-user:test-pass@localhost:1433/myinstance?database=master")

			_, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_UrlFromEnv{
							UrlFromEnv: "MSSQL_URL_NO_PREFIX",
						},
					},
				},
				&testConnectionTimeout,
			)
			require.Error(t, err)
		})
	})

	t.Run("URL", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: "sqlserver://test-user:test-pass@localhost:1433/myinstance?database=master",
						},
					},
				},
				&testConnectionTimeout,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"sqlserver://test-user:test-pass@localhost:1433/myinstance?connection+timeout=5&database=master",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})
		t.Run("ok_no_timeout", func(t *testing.T) {
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: "sqlserver://test-user:test-pass@localhost:1433/myinstance?database=master",
						},
					},
				},
				nil,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"sqlserver://test-user:test-pass@localhost:1433/myinstance?database=master",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})
		t.Run("ok_user_provided_timeout", func(t *testing.T) {
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: "sqlserver://test-user:test-pass@localhost:1433/myinstance?connection+timeout=10&database=master",
						},
					},
				},
				&testConnectionTimeout,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"sqlserver://test-user:test-pass@localhost:1433/myinstance?connection+timeout=10&database=master",
				actual.String(),
			)
			assert.Equal(t, "test-user", actual.GetUser())
		})
		t.Run("ok_strong_password", func(t *testing.T) {
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: "sqlserver://sa:myStr0ngP%40assword@localhost:1433/myinstance?database=master",
						},
					},
				},
				&testConnectionTimeout,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"sqlserver://sa:myStr0ngP%40assword@localhost:1433/myinstance?connection+timeout=5&database=master",
				actual.String(),
			)
			assert.Equal(t, "sa", actual.GetUser())
		})
		t.Run("ok_no_instance", func(t *testing.T) {
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: "sqlserver://sa:myStr0ngP%40assword@localhost:1433?database=master",
						},
					},
				},
				&testConnectionTimeout,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"sqlserver://sa:myStr0ngP%40assword@localhost:1433?connection+timeout=5&database=master",
				actual.String(),
			)
			assert.Equal(t, "sa", actual.GetUser())
		})
		t.Run("ok_no_instance_no_port", func(t *testing.T) {
			actual, err := NewFromMssqlConnection(
				&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
					MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
							Url: "sqlserver://sa:myStr0ngP%40assword@localhost?database=master",
						},
					},
				},
				&testConnectionTimeout,
			)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(
				t,
				"sqlserver://sa:myStr0ngP%40assword@localhost?connection+timeout=5&database=master",
				actual.String(),
			)
			assert.Equal(t, "sa", actual.GetUser())
		})
	})
}

func ptr[T any](val T) *T {
	return &val
}
