package v1alpha1_jobservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

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

// GetJobRuns
func Test_GetJobRuns_ByJobId(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	temporalClientMock := new(MockTemporalClient)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
	runId := uuid.NewString()
	workflowId := uuid.NewString()
	now := time.Now()
	payload, _ := converter.GetDefaultDataConverter().ToPayload(jobId)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.TemporalWfManagerMock.On("GetWorkflowClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(temporalClientMock, nil)
	m.QuerierMock.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, job.AccountID).Return(&pg_models.TemporalConfig{}, nil)

	workflows := []*workflowpb.WorkflowExecutionInfo{{
		Execution: &common.WorkflowExecution{
			WorkflowId: workflowId,
			RunId:      runId,
		},
	}}
	temporalClientMock.On("ListWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.ListWorkflowExecutionsResponse{
		Executions: workflows,
	}, nil)

	temporalClientMock.On("DescribeWorkflowExecution", mock.Anything, workflowId, runId).Return(&workflowservice.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &workflowpb.WorkflowExecutionInfo{
			CloseTime: &now,
			StartTime: &now,
			Type: &common.WorkflowType{
				Name: "name",
			},
			Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			SearchAttributes: &common.SearchAttributes{
				IndexedFields: map[string]*common.Payload{
					"TemporalScheduledById": payload,
				},
			},
			Execution: &common.WorkflowExecution{
				WorkflowId: workflowId,
			},
		},
		PendingActivities: []*workflowpb.PendingActivityInfo{},
	}, nil)

	resp, err := m.Service.GetJobRuns(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRunsRequest]{
		Msg: &mgmtv1alpha1.GetJobRunsRequest{
			Id: &mgmtv1alpha1.GetJobRunsRequest_JobId{
				JobId: jobId,
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, len(workflows), len(resp.Msg.JobRuns))
}

func Test_GetJobRuns_ByAccountId(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	temporalClientMock := new(MockTemporalClient)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
	runId := uuid.NewString()
	workflowId := uuid.NewString()
	workflowExecutionMock := getWorkflowExecutionMock(jobId, workflowId)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobsByAccount", mock.Anything, mock.Anything, accountUuid).Return([]db_queries.NeosyncApiJob{job}, nil)
	m.TemporalWfManagerMock.On("GetWorkflowClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(temporalClientMock, nil)
	m.QuerierMock.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, job.AccountID).Return(&pg_models.TemporalConfig{}, nil)

	workflows := []*workflowpb.WorkflowExecutionInfo{{
		Execution: &common.WorkflowExecution{
			WorkflowId: workflowId,
			RunId:      runId,
		},
	}}
	temporalClientMock.On("ListWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.ListWorkflowExecutionsResponse{
		Executions: workflows,
	}, nil)

	temporalClientMock.On("DescribeWorkflowExecution", mock.Anything, workflowId, runId).Return(workflowExecutionMock, nil)

	resp, err := m.Service.GetJobRuns(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRunsRequest]{
		Msg: &mgmtv1alpha1.GetJobRunsRequest{
			Id: &mgmtv1alpha1.GetJobRunsRequest_AccountId{
				AccountId: mockAccountId,
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, len(workflows), len(resp.Msg.JobRuns))
}

// GetJobRun
func Test_GetJobRun(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	temporalClientMock := new(MockTemporalClient)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
	runId := uuid.NewString()
	workflowId := uuid.NewString()
	workflowExecutionMock := getWorkflowExecutionMock(jobId, workflowId)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.TemporalWfManagerMock.On("GetWorkflowClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(temporalClientMock, nil)
	m.QuerierMock.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, job.AccountID).Return(&pg_models.TemporalConfig{}, nil)

	temporalClientMock.On("DescribeWorkflowExecution", mock.Anything, workflowId, runId).Return(workflowExecutionMock, nil)

	resp, err := m.Service.GetJobRun(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRunRequest]{
		Msg: &mgmtv1alpha1.GetJobRunRequest{
			JobRunId: runId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// GetJobRunEvents

// CreateJobRun
func Test_CreateJobRun(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(MockScheduleHandle)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, nucleusdb.UUIDString(job.ID), mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Trigger", mock.Anything, temporalclient.ScheduleTriggerOptions{}).Return(nil)

	resp, err := m.Service.CreateJobRun(context.Background(), &connect.Request[mgmtv1alpha1.CreateJobRunRequest]{
		Msg: &mgmtv1alpha1.CreateJobRunRequest{
			JobId: nucleusdb.UUIDString(job.ID),
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// CancelJobRun

// DeleteJobRun

func getWorkflowExecutionMock(jobId, workflowId string) *workflowservice.DescribeWorkflowExecutionResponse {
	now := time.Now()
	payload, _ := converter.GetDefaultDataConverter().ToPayload(jobId)

	return &workflowservice.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &workflowpb.WorkflowExecutionInfo{
			CloseTime: &now,
			StartTime: &now,
			Type: &common.WorkflowType{
				Name: "name",
			},
			Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			SearchAttributes: &common.SearchAttributes{
				IndexedFields: map[string]*common.Payload{
					"TemporalScheduledById": payload,
				},
			},
			Execution: &common.WorkflowExecution{
				WorkflowId: workflowId,
			},
		},
		PendingActivities: []*workflowpb.PendingActivityInfo{},
	}

}
