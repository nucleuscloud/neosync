package sqlprovider

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_getMaxConnectionLimitFromConnection(t *testing.T) {
	var nilInt32 *int32
	maxConnLimit := int32(50)

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

	actual = getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
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

	actual = getMaxConnectionLimitFromConnection(&mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{},
	})
	require.Empty(t, actual)
}
