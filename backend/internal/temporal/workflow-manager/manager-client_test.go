package workflowmanager

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/errgroup"
)

func Test_ManagerClient_New(t *testing.T) {
	assert.NotNil(t, New(&Config{}, NewMockDB(t), nucleusdb.NewMockDBTX(t)))
}

func Test_ManagerClient_ClearClientByAccount_Idempotent(t *testing.T) {
	mgr := New(&Config{}, NewMockDB(t), nucleusdb.NewMockDBTX(t))

	mgr.ClearClientByAccount(context.Background(), "123")
	mgr.ClearClientByAccount(context.Background(), "123")
}

func Test_ClearClientByAccount_SucceedsIfEvicting(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	mgr.ClearClientByAccount(context.Background(), accountUuid)
}

func Test_ManagerClient_GetClientByAccount(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func Test_ManagerClient_GetClientByAccount_NoNamespace(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "",
		SyncJobQueueName: "foo-queue",
	}, nil)

	accountUuid := uuid.New().String()
	_, err := mgr.GetClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
}

func Test_ManagerClient_GetClientByAccount_CacheClient(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	client2, err := mgr.GetClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.Equal(t, client, client2)
	assert.True(t, mockdb.AssertNumberOfCalls(t, "GetTemporalConfigByAccount", 1))
}

func Test_ManagerClient_GetClientByAccount_ConcurrentRequests(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
	}, nil)

	accountUuid := uuid.New().String()

	errgrp, errctx := errgroup.WithContext(context.Background())
	errgrp.Go(func() error {
		_, err := mgr.GetClientByAccount(errctx, accountUuid, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetClientByAccount(errctx, accountUuid, slog.Default())
		return err
	})
	err := errgrp.Wait()
	assert.Nil(t, err)
	assert.True(t, mockdb.AssertNumberOfCalls(t, "GetTemporalConfigByAccount", 1))
}
