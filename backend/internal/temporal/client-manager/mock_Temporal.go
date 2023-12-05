//nolint:all
package clientmanager

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// MockTemporalClient is a mock of temporal client interface.
type MockTemporalClient struct {
	mock.Mock
}

func (m *MockTemporalClient) ExecuteWorkflow(ctx context.Context, options temporalclient.StartWorkflowOptions, workflow interface{}, args ...interface{}) (temporalclient.WorkflowRun, error) {
	ret := m.Called(ctx, options, workflow, args)
	return ret.Get(0).(temporalclient.WorkflowRun), ret.Error(1)
}

func (m *MockTemporalClient) GetWorkflow(ctx context.Context, workflowID string, runID string) temporalclient.WorkflowRun {
	ret := m.Called(ctx, workflowID, runID)
	return ret.Get(0).(temporalclient.WorkflowRun)
}

func (m *MockTemporalClient) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	args := m.Called(ctx, workflowID, runID, signalName, arg)
	return args.Error(0)
}

func (m *MockTemporalClient) SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{}, options temporalclient.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (temporalclient.WorkflowRun, error) {
	args := m.Called(ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs)
	return args.Get(0).(temporalclient.WorkflowRun), args.Error(1)
}

func (m *MockTemporalClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	args := m.Called(ctx, workflowID, runID)
	return args.Error(0)
}

func (m *MockTemporalClient) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error {
	args := m.Called(ctx, workflowID, runID, reason, details)
	return args.Error(0)
}

func (m *MockTemporalClient) GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enums.HistoryEventFilterType) temporalclient.HistoryEventIterator {
	args := m.Called(ctx, workflowID, runID, isLongPoll, filterType)
	return args.Get(0).(temporalclient.HistoryEventIterator)
}

func (m *MockTemporalClient) CompleteActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error {
	args := m.Called(ctx, taskToken, result, err)
	return args.Error(0)
}

func (m *MockTemporalClient) CompleteActivityByID(ctx context.Context, namespace, workflowID, runID, activityID string, result interface{}, err error) error {
	args := m.Called(ctx, namespace, workflowID, runID, activityID, result, err)
	return args.Error(0)
}

func (m *MockTemporalClient) RecordActivityHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error {
	args := m.Called(ctx, taskToken, details)
	return args.Error(0)
}

func (m *MockTemporalClient) RecordActivityHeartbeatByID(ctx context.Context, namespace, workflowID, runID, activityID string, details ...interface{}) error {
	args := m.Called(ctx, namespace, workflowID, runID, activityID, details)
	return args.Error(0)
}

func (m *MockTemporalClient) ListClosedWorkflow(ctx context.Context, request *workflowservice.ListClosedWorkflowExecutionsRequest) (*workflowservice.ListClosedWorkflowExecutionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.ListClosedWorkflowExecutionsResponse), args.Error(1)
}

func (m *MockTemporalClient) ListOpenWorkflow(ctx context.Context, request *workflowservice.ListOpenWorkflowExecutionsRequest) (*workflowservice.ListOpenWorkflowExecutionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.ListOpenWorkflowExecutionsResponse), args.Error(1)
}

func (m *MockTemporalClient) ListWorkflow(ctx context.Context, request *workflowservice.ListWorkflowExecutionsRequest) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.ListWorkflowExecutionsResponse), args.Error(1)
}

func (m *MockTemporalClient) ListArchivedWorkflow(ctx context.Context, request *workflowservice.ListArchivedWorkflowExecutionsRequest) (*workflowservice.ListArchivedWorkflowExecutionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.ListArchivedWorkflowExecutionsResponse), args.Error(1)
}

func (m *MockTemporalClient) ScanWorkflow(ctx context.Context, request *workflowservice.ScanWorkflowExecutionsRequest) (*workflowservice.ScanWorkflowExecutionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.ScanWorkflowExecutionsResponse), args.Error(1)
}

func (m *MockTemporalClient) CountWorkflow(ctx context.Context, request *workflowservice.CountWorkflowExecutionsRequest) (*workflowservice.CountWorkflowExecutionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.CountWorkflowExecutionsResponse), args.Error(1)
}

func (m *MockTemporalClient) GetSearchAttributes(ctx context.Context) (*workflowservice.GetSearchAttributesResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(*workflowservice.GetSearchAttributesResponse), args.Error(1)
}

func (m *MockTemporalClient) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	mockArgs := m.Called(ctx, workflowID, runID, queryType, args)
	return mockArgs.Get(0).(converter.EncodedValue), mockArgs.Error(1)
}

func (m *MockTemporalClient) QueryWorkflowWithOptions(ctx context.Context, request *temporalclient.QueryWorkflowWithOptionsRequest) (*temporalclient.QueryWorkflowWithOptionsResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*temporalclient.QueryWorkflowWithOptionsResponse), args.Error(1)
}
func (m *MockTemporalClient) DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	args := m.Called(ctx, workflowID, runID)
	return args.Get(0).(*workflowservice.DescribeWorkflowExecutionResponse), args.Error(1)
}

func (m *MockTemporalClient) DescribeTaskQueue(ctx context.Context, taskqueue string, taskqueueType enums.TaskQueueType) (*workflowservice.DescribeTaskQueueResponse, error) {
	args := m.Called(ctx, taskqueue, taskqueueType)
	return args.Get(0).(*workflowservice.DescribeTaskQueueResponse), args.Error(1)
}

func (m *MockTemporalClient) ResetWorkflowExecution(ctx context.Context, request *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*workflowservice.ResetWorkflowExecutionResponse), args.Error(1)
}

func (m *MockTemporalClient) UpdateWorkerBuildIdCompatibility(ctx context.Context, options *temporalclient.UpdateWorkerBuildIdCompatibilityOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockTemporalClient) GetWorkerBuildIdCompatibility(ctx context.Context, options *temporalclient.GetWorkerBuildIdCompatibilityOptions) (*temporalclient.WorkerBuildIDVersionSets, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(*temporalclient.WorkerBuildIDVersionSets), args.Error(1)
}

func (m *MockTemporalClient) GetWorkerTaskReachability(ctx context.Context, options *temporalclient.GetWorkerTaskReachabilityOptions) (*temporalclient.WorkerTaskReachability, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(*temporalclient.WorkerTaskReachability), args.Error(1)
}

func (m *MockTemporalClient) CheckHealth(ctx context.Context, request *temporalclient.CheckHealthRequest) (*temporalclient.CheckHealthResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*temporalclient.CheckHealthResponse), args.Error(1)
}

func (m *MockTemporalClient) UpdateWorkflow(ctx context.Context, workflowID string, workflowRunID string, updateName string, args ...interface{}) (temporalclient.WorkflowUpdateHandle, error) {
	mockArgs := m.Called(ctx, workflowID, workflowRunID, updateName, args)
	return mockArgs.Get(0).(temporalclient.WorkflowUpdateHandle), mockArgs.Error(1)
}

func (m *MockTemporalClient) UpdateWorkflowWithOptions(ctx context.Context, request *temporalclient.UpdateWorkflowWithOptionsRequest) (temporalclient.WorkflowUpdateHandle, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(temporalclient.WorkflowUpdateHandle), args.Error(1)
}

func (m *MockTemporalClient) GetWorkflowUpdateHandle(ref temporalclient.GetWorkflowUpdateHandleOptions) temporalclient.WorkflowUpdateHandle {
	args := m.Called(ref)
	return args.Get(0).(temporalclient.WorkflowUpdateHandle)
}

func (m *MockTemporalClient) WorkflowService() workflowservice.WorkflowServiceClient {
	args := m.Called()
	return args.Get(0).(workflowservice.WorkflowServiceClient)
}

func (m *MockTemporalClient) OperatorService() operatorservice.OperatorServiceClient {
	args := m.Called()
	return args.Get(0).(operatorservice.OperatorServiceClient)
}

func (m *MockTemporalClient) ScheduleClient() temporalclient.ScheduleClient {
	args := m.Called()
	return args.Get(0).(temporalclient.ScheduleClient)
}

func (m *MockTemporalClient) Close() {
	m.Called()
}

// MockScheduleHandle is a mock of ScheduleHandle interface.
type MockScheduleHandle struct {
	mock.Mock
}

func (m *MockScheduleHandle) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockScheduleHandle) Delete(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockScheduleHandle) Backfill(ctx context.Context, options temporalclient.ScheduleBackfillOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Update(ctx context.Context, options temporalclient.ScheduleUpdateOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Describe(ctx context.Context) (*temporalclient.ScheduleDescription, error) {
	args := m.Called(ctx)
	return args.Get(0).(*temporalclient.ScheduleDescription), args.Error(1)
}

func (m *MockScheduleHandle) Trigger(ctx context.Context, options temporalclient.ScheduleTriggerOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Pause(ctx context.Context, options temporalclient.SchedulePauseOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Unpause(ctx context.Context, options temporalclient.ScheduleUnpauseOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

// MockNamespaceClient is a mock of Namespace Client interface.
type MockNamespaceClient struct {
	mock.Mock
}

func (_m *MockNamespaceClient) Register(ctx context.Context, request *workflowservice.RegisterNamespaceRequest) error {
	ret := _m.Called(ctx, request)
	return ret.Error(0)
}

func (_m *MockNamespaceClient) Describe(ctx context.Context, name string) (*workflowservice.DescribeNamespaceResponse, error) {
	ret := _m.Called(ctx, name)
	return ret.Get(0).(*workflowservice.DescribeNamespaceResponse), ret.Error(1)
}

func (_m *MockNamespaceClient) Update(ctx context.Context, request *workflowservice.UpdateNamespaceRequest) error {
	ret := _m.Called(ctx, request)
	return ret.Error(0)
}

func (_m *MockNamespaceClient) Close() {
	_m.Called()
}

// MockScheduleClient is a mock of Schedule Client interface.
type MockScheduleClient struct {
	mock.Mock
	Handle temporalclient.ScheduleHandle
}

func (_m *MockScheduleClient) Create(ctx context.Context, options temporalclient.ScheduleOptions) (temporalclient.ScheduleHandle, error) {
	args := _m.Called(ctx, options)
	if h := args.Get(0); h != nil {
		return h.(temporalclient.ScheduleHandle), args.Error(1)
	}
	return nil, args.Error(1)
}

func (_m *MockScheduleClient) List(ctx context.Context, options temporalclient.ScheduleListOptions) (temporalclient.ScheduleListIterator, error) {
	return nil, nil
}

func (_m *MockScheduleClient) GetHandle(ctx context.Context, scheduleID string) temporalclient.ScheduleHandle {
	args := _m.Called(ctx, scheduleID)
	if h := args.Get(0); h != nil {
		return h.(temporalclient.ScheduleHandle)
	}
	return nil
}
