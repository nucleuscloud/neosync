package v1alpha1_jobservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// GetJobRuns
func Test_GetJobRuns_ByJobId(t *testing.T) {
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
	m.TemporalWfManagerMock.On("GetTemporalConfigByAccount", mock.Anything, mockAccountId).Return(&pg_models.TemporalConfig{
		Namespace:        "default",
		SyncJobQueueName: "sync-job",
		Url:              "localhost:7233",
	}, nil)

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
	m.TemporalWfManagerMock.On("GetTemporalConfigByAccount", mock.Anything, mockAccountId).Return(&pg_models.TemporalConfig{}, nil)

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
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
	runId := uuid.NewString()
	workflowId := uuid.NewString()
	workflowExecutionMock := getWorkflowExecutionMock(jobId, workflowId)
	workflows := []*workflowpb.WorkflowExecutionInfo{{
		Execution: &common.WorkflowExecution{
			WorkflowId: workflowId,
			RunId:      runId,
		},
	}}

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	mockGetVerifiedJobRun(m.TemporalWfManagerMock, accountUuid, temporalClientMock, workflows)
	temporalClientMock.On("DescribeWorkflowExecution", mock.Anything, workflowId, runId).Return(workflowExecutionMock, nil)

	resp, err := m.Service.GetJobRun(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRunRequest]{
		Msg: &mgmtv1alpha1.GetJobRunRequest{
			JobRunId:  runId,
			AccountId: mockAccountId,
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
func Test_CancelJobRun(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	temporalClientMock := new(MockTemporalClient)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	runId := uuid.NewString()
	workflowId := uuid.NewString()
	workflows := []*workflowpb.WorkflowExecutionInfo{{
		Execution: &common.WorkflowExecution{
			WorkflowId: workflowId,
			RunId:      runId,
		},
	}}

	mockGetVerifiedJobRun(m.TemporalWfManagerMock, accountUuid, temporalClientMock, workflows)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	temporalClientMock.On("CancelWorkflow", mock.Anything, workflowId, runId).Return(nil)

	resp, err := m.Service.CancelJobRun(context.Background(), &connect.Request[mgmtv1alpha1.CancelJobRunRequest]{
		Msg: &mgmtv1alpha1.CancelJobRunRequest{
			JobRunId:  runId,
			AccountId: mockAccountId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func mockGetVerifiedJobRun(
	temporalWfManagerMock *clientmanager.MockTemporalClientManagerClient,
	accountUuid pgtype.UUID,
	temporalClientMock *MockTemporalClient,
	workflowsMock []*workflowpb.WorkflowExecutionInfo,
) {
	temporalWfManagerMock.On("DoesAccountHaveTemporalWorkspace", mock.Anything, nucleusdb.UUIDString(accountUuid), mock.Anything).Return(true, nil)
	temporalWfManagerMock.On("GetTemporalConfigByAccount", mock.Anything, nucleusdb.UUIDString(accountUuid)).Return(&pg_models.TemporalConfig{
		Namespace:        "default",
		SyncJobQueueName: "sync-job",
		Url:              "localhost:7233",
	}, nil)
	temporalWfManagerMock.On("GetWorkflowClientByAccount", mock.Anything, nucleusdb.UUIDString(accountUuid), mock.Anything).Return(temporalClientMock, nil)
	temporalClientMock.On("ListWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.ListWorkflowExecutionsResponse{
		Executions: workflowsMock,
	}, nil)
}

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
