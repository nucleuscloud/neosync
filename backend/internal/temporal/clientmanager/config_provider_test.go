package clientmanager

import (
	"context"
	"testing"

	"github.com/google/uuid"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	accountUuid = uuid.NewString()
)

func Test_DbConfigProvider_GetConfig(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		mockdb := NewMockDB(t)
		provider := NewDBConfigProvider(&TemporalConfig{
			Url:              "foo",
			Namespace:        "bar",
			SyncJobQueueName: "baz",
		}, mockdb, db_queries.NewMockDBTX(t))

		mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).
			Return(&pg_models.TemporalConfig{}, nil).Once()

		output, err := provider.GetConfig(context.Background(), accountUuid)
		require.NoError(t, err)
		require.Equal(t, &TemporalConfig{
			Url:              "foo",
			Namespace:        "bar",
			SyncJobQueueName: "baz",
		}, output)
	})

	t.Run("override account info", func(t *testing.T) {
		mockdb := NewMockDB(t)
		provider := NewDBConfigProvider(&TemporalConfig{
			Url:              "foo",
			Namespace:        "bar",
			SyncJobQueueName: "baz",
		}, mockdb, db_queries.NewMockDBTX(t))

		mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).
			Return(&pg_models.TemporalConfig{
				Url:              "foo1",
				Namespace:        "bar1",
				SyncJobQueueName: "baz1",
			}, nil).Once()

		output, err := provider.GetConfig(context.Background(), accountUuid)
		require.NoError(t, err)
		require.Equal(t, &TemporalConfig{
			Url:              "foo1",
			Namespace:        "bar1",
			SyncJobQueueName: "baz1",
		}, output)
	})
}
