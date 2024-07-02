package clientmanager

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	temporalmocks "go.temporal.io/sdk/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"golang.org/x/sync/errgroup"
)

var (
	defaultTemporalConfig = &DefaultTemporalConfig{
		Url:              "localhost:7233",
		Namespace:        "default",
		SyncJobQueueName: "sync-job",
	}
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
		Namespace:        "",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	_, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
}

func Test_ManagerClient_GetNamespaceClientByAccount_EmptyTemporalConfig(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: defaultTemporalConfig}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func Test_ManagerClient_GetNamespaceClientByAccount_EmptyTemporalConfig_EmptyDefaults(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: nil}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)

	accountUuid := uuid.New().String()
	client, err := mgr.GetNamespaceClientByAccount(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
	assert.Nil(t, client)
}

func Test_ManagerClient_GetNamespaceClientByAccount_CacheClient(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
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

func Test_ManagerClient_GetTemporalConfigByAccount_Db(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	}, nil)

	accountUuid := uuid.New().String()
	tc, err := mgr.GetTemporalConfigByAccount(context.Background(), accountUuid)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	assert.Equal(t, tc, &pg_models.TemporalConfig{
		Namespace:        "foo",
		SyncJobQueueName: "foo-queue",
		Url:              "localhost:7233",
	})
}

func Test_ManagerClient_GetTemporalConfigByAccount_Default(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: defaultTemporalConfig}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)

	accountUuid := uuid.New().String()
	tc, err := mgr.GetTemporalConfigByAccount(context.Background(), accountUuid)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	assert.Equal(t, tc, &pg_models.TemporalConfig{
		Namespace:        defaultTemporalConfig.Namespace,
		SyncJobQueueName: defaultTemporalConfig.SyncJobQueueName,
		Url:              defaultTemporalConfig.Url,
	})
}

func Test_ManagerClient_GetTemporalConfigByAccount_Empty(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: nil}, mockdb, mockdbtx)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)

	accountUuid := uuid.New().String()
	tc, err := mgr.GetTemporalConfigByAccount(context.Background(), accountUuid)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	assert.Equal(t, tc, &pg_models.TemporalConfig{
		Namespace:        "",
		SyncJobQueueName: "",
		Url:              "",
	})
}

func Test_ManagerClient_DoesAccountHaveTemporalWorkspace_TemporalConfig_Err(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: defaultTemporalConfig}, mockdb, mockdbtx)

	accountUuid := uuid.New().String()

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("test"))

	ok, err := mgr.DoesAccountHaveTemporalWorkspace(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
	assert.False(t, ok)
}

func Test_ManagerClient_DoesAccountHaveTemporalWorkspace_TemporalConfig_Empty_Namespace(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: nil}, mockdb, mockdbtx)

	accountUuid := uuid.New().String()

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)

	ok, err := mgr.DoesAccountHaveTemporalWorkspace(context.Background(), accountUuid, slog.Default())
	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_ManagerClient_DoesAccountHaveTemporalWorkspace_Has_Temporal_Namespace(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: defaultTemporalConfig}, mockdb, mockdbtx)

	accountUuid := uuid.New().String()

	mockNsClient := new(temporalmocks.NamespaceClient)
	mgr.nsmap.Store(accountUuid, mockNsClient)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)
	mockNsClient.On("Describe", mock.Anything, defaultTemporalConfig.Namespace).Return(&workflowservice.DescribeNamespaceResponse{}, nil)

	ok, err := mgr.DoesAccountHaveTemporalWorkspace(context.Background(), accountUuid, slog.Default())
	assert.NoError(t, err)
	assert.True(t, ok)
}

func Test_ManagerClient_DoesAccountHaveTemporalWorkspace_Has_Not_Temporal_Namespace(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: defaultTemporalConfig}, mockdb, mockdbtx)

	accountUuid := uuid.New().String()
	mockNsClient := new(temporalmocks.NamespaceClient)
	mgr.nsmap.Store(accountUuid, mockNsClient)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)
	mockNsClient.On("Describe", mock.Anything, defaultTemporalConfig.Namespace).Return(&workflowservice.DescribeNamespaceResponse{}, serviceerror.NewNamespaceNotFound(defaultTemporalConfig.Namespace))

	ok, err := mgr.DoesAccountHaveTemporalWorkspace(context.Background(), accountUuid, slog.Default())
	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_ManagerClient_DoesAccountHaveTemporalWorkspace_Describe_Error(t *testing.T) {
	mockdb := NewMockDB(t)
	mockdbtx := nucleusdb.NewMockDBTX(t)
	mgr := New(&Config{DefaultTemporalConfig: defaultTemporalConfig}, mockdb, mockdbtx)

	accountUuid := uuid.New().String()

	mockNsClient := new(temporalmocks.NamespaceClient)
	mgr.nsmap.Store(accountUuid, mockNsClient)

	mockdb.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, mock.Anything).Return(&pg_models.TemporalConfig{}, nil)
	mockNsClient.On("Describe", mock.Anything, defaultTemporalConfig.Namespace).Return(&workflowservice.DescribeNamespaceResponse{}, serviceerror.NewCanceled("test"))

	ok, err := mgr.DoesAccountHaveTemporalWorkspace(context.Background(), accountUuid, slog.Default())
	assert.Error(t, err)
	assert.False(t, ok)
}
