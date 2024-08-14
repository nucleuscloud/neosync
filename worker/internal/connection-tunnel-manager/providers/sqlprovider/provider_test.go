package sqlprovider

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_getMaxConnectionLimitFromConnection(t *testing.T) {
	var nilInt32 *int32
	maxConnLimit := int32(50)

	t.Run("postgres", func(t *testing.T) {
		actual := getMaxConnectionLimitFromConnection(nil)
		require.Empty(t, actual)

		actual = getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		})
		require.Equal(t, nilInt32, actual)

		actual = getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionOptions: &mgmtv1alpha1.SqlConnectionOptions{
						MaxConnectionLimit: &maxConnLimit,
					},
				},
			},
		})
		require.Equal(t, &maxConnLimit, actual)
	})

	t.Run("mysql", func(t *testing.T) {
		actual := getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{},
		})
		require.Equal(t, nilInt32, actual)

		actual = getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
					ConnectionOptions: &mgmtv1alpha1.SqlConnectionOptions{
						MaxConnectionLimit: &maxConnLimit,
					},
				},
			},
		})
		require.Equal(t, &maxConnLimit, actual)
	})

	t.Run("mssql", func(t *testing.T) {
		actual := getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{},
		})
		require.Equal(t, nilInt32, actual)

		actual = getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
				MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
					ConnectionOptions: &mgmtv1alpha1.SqlConnectionOptions{
						MaxConnectionLimit: &maxConnLimit,
					},
				},
			},
		})
		require.Equal(t, &maxConnLimit, actual)
	})

	t.Run("awss3", func(t *testing.T) {
		actual := getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{},
		})
		require.Empty(t, actual)
	})
}
