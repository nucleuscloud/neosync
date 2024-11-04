package clientmanager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DescribeSchedulesResponse struct {
	Schedule *temporalclient.ScheduleDescription
	Error    error
}

type Interface interface {
	DoesAccountHaveNamespace(ctx context.Context, accountId string, logger *slog.Logger) (bool, error)
	GetSyncJobTaskQueue(ctx context.Context, accountId string, logger *slog.Logger) (string, error)

	CreateSchedule(ctx context.Context, accountId string, opts *temporalclient.ScheduleOptions, logger *slog.Logger) (string, error)
	TriggerSchedule(ctx context.Context, accountId string, scheduleId string, opts *temporalclient.ScheduleTriggerOptions, logger *slog.Logger) error
	PauseSchedule(ctx context.Context, accountId string, scheduleId string, opts *temporalclient.SchedulePauseOptions, logger *slog.Logger) error
	UnpauseSchedule(ctx context.Context, accountId string, scheduleId string, opts *temporalclient.ScheduleUnpauseOptions, logger *slog.Logger) error
	UpdateSchedule(ctx context.Context, accountId string, scheduleId string, opts *temporalclient.ScheduleUpdateOptions, logger *slog.Logger) error
	DescribeSchedule(ctx context.Context, accountId string, scheduleId string, logger *slog.Logger) (*temporalclient.ScheduleDescription, error)
	DescribeSchedules(ctx context.Context, accountId string, scheduleIds []string, logger *slog.Logger) ([]*DescribeSchedulesResponse, error)
	DeleteSchedule(ctx context.Context, accountId string, scheduleId string, logger *slog.Logger) error
	GetWorkflowExecutionById(ctx context.Context, accountId string, workflowId string, logger *slog.Logger) (*workflowpb.WorkflowExecutionInfo, error)
	DeleteWorkflowExecution(ctx context.Context, accountId string, workflowId string, logger *slog.Logger) error
	GetWorkflowExecutionsByScheduleIds(ctx context.Context, accountId string, scheduleIds []string, logger *slog.Logger) ([]*workflowpb.WorkflowExecutionInfo, error)
	DescribeWorklowExecution(ctx context.Context, accountId string, workflowId string, logger *slog.Logger) (*workflowservice.DescribeWorkflowExecutionResponse, error)
	CancelWorkflow(ctx context.Context, accountId string, workflowId string, logger *slog.Logger) error
	TerminateWorkflow(ctx context.Context, accountId string, workflowId string, logger *slog.Logger) error
	GetWorkflowHistory(ctx context.Context, accountId string, workflowId string, logger *slog.Logger) (temporalclient.HistoryEventIterator, error)
}

var _ Interface = (*ClientManager)(nil)

type ClientManager struct {
	configProvider ConfigProvider
	clientFactory  ClientFactory
	clientCache    *ClientCache
}

func NewClientManager(
	configProvider ConfigProvider,
	clientFactory ClientFactory,
) *ClientManager {
	return &ClientManager{
		configProvider: configProvider,
		clientFactory:  clientFactory,
		clientCache:    NewClientCache(),
	}
}

func (m *ClientManager) getClients(ctx context.Context, accountId string, logger *slog.Logger) (*clientHandle, error) {
	config, err := m.configProvider.GetConfig(ctx, accountId)
	if err != nil {
		return nil, fmt.Errorf("failed to get temporal config: %w", err)
	}

	if config.Namespace == "" {
		return nil, fmt.Errorf("temporal namespace not configured")
	}

	return m.clientCache.getOrCreateClient(ctx, config, m.clientFactory, logger)
}

func (m *ClientManager) DoesAccountHaveNamespace(
	ctx context.Context,
	accountID string,
	logger *slog.Logger,
) (bool, error) {
	clients, err := m.getClients(ctx, accountID, logger)
	if err != nil {
		return false, err
	}
	defer clients.Release()

	_, err = clients.NamespaceClient().Describe(ctx, clients.config.Namespace)
	if err != nil {
		if _, ok := err.(*serviceerror.NamespaceNotFound); ok {
			logger.Warn("temporal namespace not found")
			return false, nil
		}
		return false, fmt.Errorf("failed to describe namespace: %w", err)
	}

	return true, nil
}

func (m *ClientManager) GetSyncJobTaskQueue(ctx context.Context, accountId string, logger *slog.Logger) (string, error) {
	config, err := m.configProvider.GetConfig(ctx, accountId)
	if err != nil {
		return "", fmt.Errorf("failed to get temporal config: %w", err)
	}
	return config.SyncJobQueueName, nil
}

func (m *ClientManager) CreateSchedule(
	ctx context.Context,
	accountID string,
	opts *temporalclient.ScheduleOptions,
	logger *slog.Logger,
) (string, error) {
	schedclient, closeClient, err := m.createScheduleClient(ctx, accountID, logger)
	if err != nil {
		return "", err
	}
	handle, err := schedclient.Create(ctx, *opts)
	if err != nil {
		return "", err
	}
	defer closeClient()
	return handle.GetID(), nil
}

func (m *ClientManager) TriggerSchedule(
	ctx context.Context,
	accountID string,
	id string,
	opts *temporalclient.ScheduleTriggerOptions,
	logger *slog.Logger,
) error {
	schedclient, closeClient, err := m.createScheduleClient(ctx, accountID, logger)
	if err != nil {
		return err
	}
	defer closeClient()
	handle := schedclient.GetHandle(ctx, id)
	return handle.Trigger(ctx, *opts)
}

func (m *ClientManager) PauseSchedule(
	ctx context.Context,
	accountId string,
	id string,
	opts *temporalclient.SchedulePauseOptions,
	logger *slog.Logger,
) error {
	schedclient, closeFn, err := m.createScheduleClient(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer closeFn()
	handle := schedclient.GetHandle(ctx, id)
	return handle.Pause(ctx, *opts)
}

func (m *ClientManager) UnpauseSchedule(
	ctx context.Context,
	accountId string,
	id string,
	opts *temporalclient.ScheduleUnpauseOptions,
	logger *slog.Logger,
) error {
	schedclient, closeFn, err := m.createScheduleClient(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer closeFn()
	handle := schedclient.GetHandle(ctx, id)
	return handle.Unpause(ctx, *opts)
}

func (m *ClientManager) UpdateSchedule(
	ctx context.Context,
	accountId string,
	id string,
	opts *temporalclient.ScheduleUpdateOptions,
	logger *slog.Logger,
) error {
	schedclient, closeFn, err := m.createScheduleClient(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer closeFn()
	handle := schedclient.GetHandle(ctx, id)
	return handle.Update(ctx, *opts)
}

func (m *ClientManager) DescribeSchedule(
	ctx context.Context,
	accountId string,
	id string,
	logger *slog.Logger,
) (*temporalclient.ScheduleDescription, error) {
	schedclient, closeFn, err := m.createScheduleClient(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	defer closeFn()
	handle := schedclient.GetHandle(ctx, id)
	return handle.Describe(ctx)
}

func (m *ClientManager) DescribeSchedules(
	ctx context.Context,
	accountId string,
	ids []string,
	logger *slog.Logger,
) ([]*DescribeSchedulesResponse, error) {
	output := make([]*DescribeSchedulesResponse, len(ids))

	schedclient, closeFn, err := m.createScheduleClient(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	wg := sync.WaitGroup{}
	wg.Add(len(ids))
	for idx, id := range ids {
		idx := idx
		id := id
		go func() {
			defer wg.Done()
			handle := schedclient.GetHandle(ctx, id)
			schedule, err := handle.Describe(ctx)
			output[idx] = &DescribeSchedulesResponse{Schedule: schedule, Error: err}
		}()
	}

	wg.Wait()

	return output, nil
}

func (m *ClientManager) DeleteSchedule(
	ctx context.Context,
	accountId string,
	id string,
	logger *slog.Logger,
) error {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer clients.Release()

	logger.Debug(fmt.Sprintf("removing schedule %q workflow executions", id))
	err = m.deleteWorkflows(
		ctx,
		clients.WorkflowClient(),
		clients.config.Namespace,
		func(ctx context.Context, namespace string) ([]*workflowpb.WorkflowExecutionInfo, error) {
			return getWorfklowsByScheduleIds(ctx, clients.WorkflowClient(), namespace, []string{id})
		},
	)
	if err != nil {
		return fmt.Errorf("unable to delete all workflows when removing schedule: %w", err)
	}

	svc := clients.WorkflowClient().WorkflowService()
	logger.Debug(fmt.Sprintf("removing schedule %q", id))
	_, err = svc.DeleteSchedule(ctx, &workflowservice.DeleteScheduleRequest{Namespace: clients.config.Namespace, ScheduleId: id})
	if err != nil && isGrpcNotFoundError(err) {
		logger.Debug("schedule was not found when issuing delete")
		return nil
	}
	return err
}

func (m *ClientManager) GetWorkflowExecutionsByScheduleIds(
	ctx context.Context,
	accountId string,
	scheduleIds []string,
	logger *slog.Logger,
) ([]*workflowpb.WorkflowExecutionInfo, error) {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	defer clients.Release()

	return getWorfklowsByScheduleIds(ctx, clients.WorkflowClient(), clients.config.Namespace, scheduleIds)
}

func (m *ClientManager) GetWorkflowExecutionById(
	ctx context.Context,
	accountId string,
	workflowId string,
	logger *slog.Logger,
) (*workflowpb.WorkflowExecutionInfo, error) {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	defer clients.Release()

	return getLatestWorfkow(ctx, clients.WorkflowClient(), clients.config.Namespace, workflowId)
}

func getLatestWorfkow(
	ctx context.Context,
	client temporalclient.Client,
	namespace string,
	workflowId string,
) (*workflowpb.WorkflowExecutionInfo, error) {
	resp, err := client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     fmt.Sprintf("WorkflowId = %q", workflowId),
		PageSize:  1,
	})
	if err != nil {
		return nil, err
	}
	executions := resp.GetExecutions()
	if len(executions) == 0 {
		return nil, nucleuserrors.NewNotFound(fmt.Sprintf("workflow not found for %q", workflowId))
	}
	return executions[0], nil
}

func (m *ClientManager) DescribeWorklowExecution(
	ctx context.Context,
	accountId string,
	workflowId string,
	logger *slog.Logger,
) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	defer clients.Release()

	wf, err := getLatestWorfkow(ctx, clients.WorkflowClient(), clients.config.Namespace, workflowId)
	if err != nil {
		return nil, err
	}
	return clients.WorkflowClient().DescribeWorkflowExecution(ctx, wf.GetExecution().GetWorkflowId(), wf.GetExecution().GetRunId())
}

func (m *ClientManager) DeleteWorkflowExecution(
	ctx context.Context,
	accountId string,
	workflowId string,
	logger *slog.Logger,
) error {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer clients.Release()

	err = m.deleteWorkflows(
		ctx,
		clients.WorkflowClient(),
		clients.config.Namespace,
		func(ctx context.Context, namespace string) ([]*workflowpb.WorkflowExecutionInfo, error) {
			// todo: should technically paginate this, but the amount of workflows + unique run ids should be only ever 1
			resp, err := clients.WorkflowClient().ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
				Namespace: namespace,
				Query:     fmt.Sprintf("WorkflowId = %q", workflowId),
			})
			if err != nil {
				return nil, err
			}
			return resp.GetExecutions(), nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *ClientManager) deleteWorkflows(
	ctx context.Context,
	client temporalclient.Client,
	namespace string,
	getWorkflowExecs func(ctx context.Context, ns string) ([]*workflowpb.WorkflowExecutionInfo, error),
) error {
	workflowExecs, err := getWorkflowExecs(ctx, namespace)
	if err != nil {
		return err
	}

	svc := client.WorkflowService()

	errgrp, ctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(10)
	for _, wf := range workflowExecs {
		wf := wf
		errgrp.Go(func() error {
			_, err := svc.DeleteWorkflowExecution(ctx, &workflowservice.DeleteWorkflowExecutionRequest{
				Namespace:         namespace,
				WorkflowExecution: wf.GetExecution(),
			})
			return err
		})
	}
	return errgrp.Wait()
}

func (m *ClientManager) CancelWorkflow(
	ctx context.Context,
	accountId string,
	workflowId string,
	logger *slog.Logger,
) error {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer clients.Release()

	wf, err := getLatestWorfkow(ctx, clients.WorkflowClient(), clients.config.Namespace, workflowId)
	if err != nil {
		return err
	}
	return clients.WorkflowClient().CancelWorkflow(ctx, wf.GetExecution().GetWorkflowId(), wf.GetExecution().GetRunId())
}

func (m *ClientManager) TerminateWorkflow(
	ctx context.Context,
	accountId string,
	workflowId string,
	logger *slog.Logger,
) error {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return err
	}
	defer clients.Release()

	wf, err := getLatestWorfkow(ctx, clients.WorkflowClient(), clients.config.Namespace, workflowId)
	if err != nil {
		return err
	}
	return clients.WorkflowClient().TerminateWorkflow(ctx, wf.GetExecution().GetWorkflowId(), wf.GetExecution().GetRunId(), "terminated by user")
}

func (m *ClientManager) GetWorkflowHistory(
	ctx context.Context,
	accountId string,
	workflowId string,
	logger *slog.Logger,
) (temporalclient.HistoryEventIterator, error) {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	defer clients.Release()

	wf, err := getLatestWorfkow(ctx, clients.WorkflowClient(), clients.config.Namespace, workflowId)
	if err != nil {
		return nil, err
	}
	return clients.WorkflowClient().GetWorkflowHistory(
		ctx,
		wf.GetExecution().GetWorkflowId(),
		wf.GetExecution().GetRunId(),
		false,
		enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT,
	), nil
}

func getWorfklowsByScheduleIds(
	ctx context.Context,
	client temporalclient.Client,
	namespace string,
	scheduleIds []string,
) ([]*workflowpb.WorkflowExecutionInfo, error) {
	if len(scheduleIds) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf("TemporalScheduledById IN (%s)", getScheduleIdsForQuery(scheduleIds))

	executions := []*workflowpb.WorkflowExecutionInfo{}
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      20,
			NextPageToken: nextPageToken,
			Query:         query,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve workflow executions: %w", err)
		}
		executions = append(executions, resp.GetExecutions()...)
		nextPageToken = resp.GetNextPageToken()
	}
	return executions, nil
}

func getScheduleIdsForQuery(scheduleIds []string) string {
	formatted := make([]string, len(scheduleIds))
	for idx := range scheduleIds {
		formatted[idx] = fmt.Sprintf("%q", scheduleIds[idx])
	}
	return strings.Join(formatted, ", ")
}

func (m *ClientManager) createScheduleClient(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.ScheduleClient, func(), error) {
	clients, err := m.getClients(ctx, accountId, logger)
	if err != nil {
		return nil, nil, err
	}

	return clients.WorkflowClient().ScheduleClient(), func() {
		clients.WorkflowClient().Close()
	}, nil
}

func (m *ClientManager) getNamespace(
	ctx context.Context,
	accountId string,
) (string, error) {
	config, err := m.configProvider.GetConfig(ctx, accountId)
	if err != nil {
		return "", fmt.Errorf("failed to get temporal config: %w", err)
	}

	if config.Namespace == "" {
		return "", errors.New("temporal namespace not configured")
	}
	return config.Namespace, nil
}

func isGrpcNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Convert error to gRPC status
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// Check if the error code is NotFound
	return st.Code() == codes.NotFound
}
