package clientmanager

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

func Test_ManagerClient_ClearWorkflowClientByAccount_Idempotent(t *testing.T) {
	mgr := New(&Config{}, NewMockDB(t), nucleusdb.NewMockDBTX(t))

	mgr.ClearWorkflowClientByAccount(context.Background(), "123")
	mgr.ClearWorkflowClientByAccount(context.Background(), "123")
}

func Test_ClearWorkflowClientByAccount_SucceedsIfEvicting(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetWorkflowClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	mgr.ClearWorkflowClientByAccount(context.Background(), accountUuid)
}

func Test_ManagerClient_GetWorkflowClientByAccount(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetWorkflowClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func Test_ManagerClient_GetWorkflowClientByAccount_NoNamespace(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	_, err := mgr.GetWorkflowClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
}

func Test_ManagerClient_GetWorkflowClientByAccount_CacheClient(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetWorkflowClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	client2, err := mgr.GetWorkflowClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.Equal(t, client, client2)
	assert.True(t, mockdb.AssertNumberOfCalls(t, "GetTemporalConfigByAccount", 1))
}

func Test_ManagerClient_GetWorkflowClientByAccount_ConcurrentRequests(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()

	errgrp, errctx := errgroup.WithContext(context.Background())
	errgrp.Go(func() error {
		_, err := mgr.GetWorkflowClientByAccount(errctx, accountUuid, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetWorkflowClientByAccount(errctx, accountUuid, slog.Default())
		return err
	})
	err := errgrp.Wait()
	assert.Nil(t, err)
	assert.True(t, mockdb.AssertNumberOfCalls(t, "GetTemporalConfigByAccount", 1))
}

func Test_ManagerClient_ClearNamespaceClientByAccount_Idempotent(t *testing.T) {
	mgr := New(&Config{}, NewMockDB(t), nucleusdb.NewMockDBTX(t))

	mgr.ClearNamespaceClientByAccount(context.Background(), "123")
	mgr.ClearNamespaceClientByAccount(context.Background(), "123")
}

func Test_ClearNamespaceClientByAccount_SucceedsIfEvicting(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	mgr.ClearNamespaceClientByAccount(context.Background(), accountUuid)
}

func Test_ManagerClient_GetNamespaceClientByAccount(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func Test_ManagerClient_GetNamespaceClientByAccount_NoNamespace(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	_, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
}

func Test_ManagerClient_GetNamespaceClientByAccount_CacheClient(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	client2, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.Equal(t, client, client2)
	assert.True(t, mockdb.AssertNumberOfCalls(t, "GetTemporalConfigByAccount", 1))
}

func Test_ManagerClient_GetNamespaceClientByAccount_ConcurrentRequests(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&jsonmodels.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()

	errgrp, errctx := errgroup.WithContext(context.Background())
	errgrp.Go(func() error {
		_, err := mgr.GetNamespaceClientByAccount(errctx, accountUuid, slog.Default())
		return err
	})
	errgrp.Go(func() error {
		_, err := mgr.GetNamespaceClientByAccount(errctx, accountUuid, slog.Default())
		return err
	})
	err := errgrp.Wait()
	assert.Nil(t, err)
	assert.True(t, mockdb.AssertNumberOfCalls(t, "GetTemporalConfigByAccount", 1))
}
