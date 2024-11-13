package providers

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/internal/connection-tunnel-manager"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func Test_NewProvider(t *testing.T) {
	mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
	mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)

	require.NotNil(t, NewProvider(mockMp, mockSp))
}

func Test_Provider_GetConnectionClient(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)

		provider := NewProvider(mockMp, mockSp)

		mockMp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(&mongo.Client{}, nil)

		result, err := provider.GetConnectionClient(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MongoConfig{},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(mockDbtx, nil)

		result, err := provider.GetConnectionClient(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(mockDbtx, nil)

		result, err := provider.GetConnectionClient(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("mssql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("GetConnectionClient", mock.Anything, mock.Anything, mock.Anything).
			Return(mockDbtx, nil)

		result, err := provider.GetConnectionClient(&mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func Test_Provider_CloseClientConnection(t *testing.T) {
	t.Run("mongo", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)

		provider := NewProvider(mockMp, mockSp)

		mockMp.On("CloseClientConnection", mock.Anything).Return(nil)

		client := &mongo.Client{}
		err := provider.CloseClientConnection(client)
		require.NoError(t, err)
	})

	t.Run("sql", func(t *testing.T) {
		t.Parallel()
		mockMp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_mongodb.MongoClient](t)
		mockSp := connectiontunnelmanager.NewMockConnectionProvider[neosync_benthos_sql.SqlDbtx](t)
		mockDbtx := neosync_benthos_sql.NewMockSqlDbtx(t)

		provider := NewProvider(mockMp, mockSp)

		mockSp.On("CloseClientConnection", mock.Anything).Return(nil)

		err := provider.CloseClientConnection(mockDbtx)
		require.NoError(t, err)
	})
}
