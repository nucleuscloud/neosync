package providers

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers/sqlprovider"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func Test_NewProvider(t *testing.T) {
	mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
	mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

	require.NotNil(t, NewProvider(mockMp, mockSp))
}

func Test_Provider_GetConnectionDetails(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		mockMp.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
			Return(&mongoconnect.ConnectionDetails{}, nil)

		result, err := provider.GetConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MongoConfig{},
		}, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
			Return(&sqlconnect.ConnectionDetails{}, nil)

		result, err := provider.GetConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		}, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionDetails", mock.Anything, mock.Anything, mock.Anything).
			Return(&sqlconnect.ConnectionDetails{}, nil)

		result, err := provider.GetConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{},
		}, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func Test_Provider_GetConnectionClient(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		mockMp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(&mongo.Client{}, nil)

		var opts any = struct{}{}
		result, err := provider.GetConnectionClient("mongodb", "test-str", opts)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(mockDbtx, nil)

		opts := &sqlprovider.ConnectionClientConfig{}
		result, err := provider.GetConnectionClient("postgres", "test-str", opts)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("postgres-bad", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		var opts any = struct{}{}
		result, err := provider.GetConnectionClient("postgres", "test-str", opts)
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(mockDbtx, nil)

		opts := &sqlprovider.ConnectionClientConfig{}
		result, err := provider.GetConnectionClient("mysql", "test-str", opts)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("mysql-bad", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		var opts any = struct{}{}
		result, err := provider.GetConnectionClient("mysql", "test-str", opts)
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func Test_Provider_CloseClientConnection(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		mockMp.On("CloseClientConnection", mock.Anything).Return(nil)

		client := &mongo.Client{}
		err := provider.CloseClientConnection(client)
		require.NoError(t, err)
	})

	t.Run("sql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("CloseClientConnection", mock.Anything).Return(nil)

		err := provider.CloseClientConnection(mockDbtx)
		require.NoError(t, err)
	})
}

func Test_Provider_GetConnectionClientConfig(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		cc := &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MongoConfig{},
		}

		var result any = struct{}{}
		mockMp.On("GetConnectionClientConfig", cc).Return(result, nil)

		config, err := provider.GetConnectionClientConfig(cc)
		require.NoError(t, err)
		require.Equal(t, result, config)
	})

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		cc := &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		}

		result := &sqlprovider.ConnectionClientConfig{}
		mockSp.On("GetConnectionClientConfig", cc).Return(result, nil)

		config, err := provider.GetConnectionClientConfig(cc)
		require.NoError(t, err)
		require.Equal(t, result, config)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient, any](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig](t)

		provider := NewProvider(mockMp, mockSp)

		cc := &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{},
		}

		result := &sqlprovider.ConnectionClientConfig{}
		mockSp.On("GetConnectionClientConfig", cc).Return(result, nil)

		config, err := provider.GetConnectionClientConfig(cc)
		require.NoError(t, err)
		require.Equal(t, result, config)
	})
}
