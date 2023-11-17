package v1alpha1_jobservice

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/workflowservice/v1"
	temporal "go.temporal.io/sdk/client"
)

const (
	anonymousUserId    = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider   = "test-provider"
	mockUserId         = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId      = "5629813e-1a35-4874-922c-9827d85f0378"
	mockConnectionName = "test-conn"
	mockConnectionId   = "884765c6-1708-488d-b03a-70a02b12c81e"
)

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

func (m *MockScheduleHandle) Backfill(ctx context.Context, options temporal.ScheduleBackfillOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Update(ctx context.Context, options temporal.ScheduleUpdateOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Describe(ctx context.Context) (*temporal.ScheduleDescription, error) {
	args := m.Called(ctx)
	return args.Get(0).(*temporal.ScheduleDescription), args.Error(1)
}

func (m *MockScheduleHandle) Trigger(ctx context.Context, options temporal.ScheduleTriggerOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Pause(ctx context.Context, options temporal.SchedulePauseOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockScheduleHandle) Unpause(ctx context.Context, options temporal.ScheduleUnpauseOptions) error {
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
	Handle temporal.ScheduleHandle
}

func (_m *MockScheduleClient) Create(ctx context.Context, options temporal.ScheduleOptions) (temporal.ScheduleHandle, error) {
	args := _m.Called(ctx, options)
	if h := args.Get(0); h != nil {
		return h.(temporal.ScheduleHandle), args.Error(1)
	}
	return nil, args.Error(1)
}

func (_m *MockScheduleClient) List(ctx context.Context, options temporal.ScheduleListOptions) (temporal.ScheduleListIterator, error) {
	return nil, nil
}

func (_m *MockScheduleClient) GetHandle(ctx context.Context, scheduleID string) temporal.ScheduleHandle {
	args := _m.Called(ctx, scheduleID)
	if h := args.Get(0); h != nil {
		return h.(temporal.ScheduleHandle)
	}
	return nil
}

// GetJobs
func Test_GetJobs_UnauthorizedUser(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, false)

	resp, err := m.Service.GetJobs(context.Background(), &connect.Request[mgmtv1alpha1.GetJobsRequest]{
		Msg: &mgmtv1alpha1.GetJobsRequest{
			AccountId: mockAccountId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_GetJobs(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString())
	job2 := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn1 := getConnectionMock(mockAccountId, "test-1", nil)
	destConn2 := getConnectionMock(mockAccountId, "test-2", nil)
	destConnAssociation1 := mockJobDestConnAssociation(job1.ID, destConn1.ID, &pg_models.JobDestinationOptions{})
	destConnAssociation2 := mockJobDestConnAssociation(job2.ID, destConn2.ID, &pg_models.JobDestinationOptions{})
	m.QuerierMock.On("GetJobsByAccount", context.Background(), mock.Anything, accountUuid).Return([]db_queries.NeosyncApiJob{job1, job2}, nil)
	m.QuerierMock.On("GetJobConnectionDestinationsByJobIds", context.Background(), mock.Anything, []pgtype.UUID{job1.ID, job2.ID}).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation1, destConnAssociation2}, nil)
	resp, err := m.Service.GetJobs(context.Background(), &connect.Request[mgmtv1alpha1.GetJobsRequest]{
		Msg: &mgmtv1alpha1.GetJobsRequest{
			AccountId: mockAccountId,
		},
	})

	jobActualMap := map[string]*mgmtv1alpha1.Job{}
	for _, job := range resp.Msg.GetJobs() {
		jobActualMap[job.Id] = job
	}

	job1Id := nucleusdb.UUIDString(job1.ID)
	job2Id := nucleusdb.UUIDString(job2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetJobs()))
	assert.NotNil(t, jobActualMap[job1Id])
	assert.NotNil(t, jobActualMap[job2Id])
	assert.Equal(t, nucleusdb.UUIDString(destConn1.ID), jobActualMap[job1Id].Destinations[0].ConnectionId)
	assert.Equal(t, nucleusdb.UUIDString(destConn2.ID), jobActualMap[job2Id].Destinations[0].ConnectionId)
}

func Test_GetJobs_MissingDestinations(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString())
	m.QuerierMock.On("GetJobsByAccount", context.Background(), mock.Anything, accountUuid).Return([]db_queries.NeosyncApiJob{job1}, nil)
	m.QuerierMock.On("GetJobConnectionDestinationsByJobIds", context.Background(), mock.Anything, []pgtype.UUID{job1.ID}).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	resp, err := m.Service.GetJobs(context.Background(), &connect.Request[mgmtv1alpha1.GetJobsRequest]{
		Msg: &mgmtv1alpha1.GetJobsRequest{
			AccountId: mockAccountId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.GetJobs()))
	assert.Empty(t, resp.Msg.Jobs[0].Destinations)
}

// GetJob
func Test_GetJob(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn := getConnectionMock(mockAccountId, "test-1", nil)
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)
	jobId := nucleusdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_GetJob_MissingDestinations(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	jobId := nucleusdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Msg.Job.Destinations)
}

func Test_GetJob_UnauthorizedUser(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, false)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	jobId := nucleusdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_GetJob_NotFound(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	var nilJob db_queries.NeosyncApiJob
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, mock.Anything).Return(nilJob, sql.ErrNoRows)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, mock.Anything).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: mockAccountId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// GetJobStatus
func Test_GetJobStatus_Paused(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, mock.Anything).Return(job, nil)

	mockHandle := new(MockScheduleHandle)

	mockHandle.On("Describe", mock.Anything).Return(&temporal.ScheduleDescription{
		Schedule: temporal.Schedule{
			State: &temporal.ScheduleState{
				Paused: true,
			},
		},
	}, nil)

	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)

	resp, err := m.Service.GetJobStatus(context.Background(), &connect.Request[mgmtv1alpha1.GetJobStatusRequest]{
		Msg: &mgmtv1alpha1.GetJobStatusRequest{
			JobId: jobId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mgmtv1alpha1.JobStatus(3), resp.Msg.Status)
}

func Test_GetJobStatus_Enabled(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)

	mockHandle := new(MockScheduleHandle)

	mockHandle.On("Describe", mock.Anything).Return(&temporal.ScheduleDescription{
		Schedule: temporal.Schedule{
			State: &temporal.ScheduleState{
				Paused: false,
			},
		},
	}, nil)

	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)

	resp, err := m.Service.GetJobStatus(context.Background(), &connect.Request[mgmtv1alpha1.GetJobStatusRequest]{
		Msg: &mgmtv1alpha1.GetJobStatusRequest{
			JobId: jobId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, mgmtv1alpha1.JobStatus(1), resp.Msg.Status)
}

// GetJobStatuses
func Test_GetJobStatuses(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(MockScheduleHandle)
	mockScheduleClient := new(MockScheduleClient)
	mockScheduleClient.Handle = mockHandle

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())

	m.QuerierMock.On("GetJobsByAccount", mock.Anything, mock.Anything, job.AccountID).Return([]db_queries.NeosyncApiJob{job}, nil)
	m.TemporalWfManagerMock.On("GetScheduleClientByAccount", mock.Anything, mock.Anything, mock.Anything).Return(mockScheduleClient, nil)
	mockScheduleClient.On("GetHandle", mock.Anything, mock.Anything).Return(mockHandle)

	mockHandle.On("Describe", mock.Anything).Return(&temporal.ScheduleDescription{
		Schedule: temporal.Schedule{
			State: &temporal.ScheduleState{
				Paused: false,
			},
		},
	}, nil)

	resp, err := m.Service.GetJobStatuses(context.Background(), &connect.Request[mgmtv1alpha1.GetJobStatusesRequest]{
		Msg: &mgmtv1alpha1.GetJobStatusesRequest{
			AccountId: mockAccountId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Msg.Statuses))
	assert.Equal(t, mgmtv1alpha1.JobStatus(1), resp.Msg.Statuses[0].Status)

}

// // GetJobRecentRuns
// func Test_GetJobRecentRuns(t *testing.T) {
// 	m := createServiceMock(t, &Config{IsAuthEnabled: true})

// 	mockIsUserInAccount(m.UserAccountServiceMock, true)
// 	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
// 	jobId := nucleusdb.UUIDString(job.ID)

// 	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)

// 	mockHandle := new(MockScheduleHandle)
// 	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)

// 	mockHandle.On("Describe", mock.Anything).Return(&temporal.ScheduleDescription{
// 		Info: ScheduleInfo{
// 			RecentActions: []ScheduleActionResult{

// 			},
// 		},
// 	}, nil)

// 	resp, err := m.Service.GetJobStatus(context.Background(), &connect.Request[mgmtv1alpha1.GetJobStatusRequest]{
// 		Msg: &mgmtv1alpha1.GetJobStatusRequest{
// 			JobId: jobId,
// 		},
// 	})

// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.Equal(t, mgmtv1alpha1.JobStatus(1), resp.Msg.Status)
// }

// CreateJob
func Test_CreateJob(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)
	mockHandle := new(MockScheduleHandle)
	mockScheduleClient := new(MockScheduleClient)
	mockScheduleClient.Handle = mockHandle

	cronSchedule := "* * * * *"
	whereClause := "where"
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString())
	srcConn := getConnectionMock(mockAccountId, "test-4", nil)
	destConn := getConnectionMock(mockAccountId, "test-1", nil)
	destConnAssociation := mockJobDestConnAssociation(job1.ID, destConn.ID, &pg_models.JobDestinationOptions{
		PostgresOptions: &pg_models.PostgresDestinationOptions{
			TruncateTableConfig: &pg_models.PostgresTruncateTableConfig{
				TruncateBeforeInsert: true,
				TruncateCascade:      true,
			},
			InitTableSchema: true,
		},
	})

	cron := pgtype.Text{}
	cron.Scan(cronSchedule)
	destinationParams := []db_queries.CreateJobConnectionDestinationsParams{
		{JobID: job1.ID, ConnectionID: destConn.ID, Options: &pg_models.JobDestinationOptions{
			PostgresOptions: &pg_models.PostgresDestinationOptions{
				TruncateTableConfig: &pg_models.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					TruncateCascade:      true,
				},
				InitTableSchema: true,
			},
		}},
	}

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	m.DbtxMock.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTx.On("Commit", mock.Anything).Return(nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)
	m.QuerierMock.On("GetConnectionById", mock.Anything, mock.Anything, srcConn.ID).Return(srcConn, nil)
	m.QuerierMock.On("GetConnectionById", mock.Anything, mock.Anything, destConn.ID).Return(destConn, nil)
	m.QuerierMock.On("GetTemporalConfigByAccount", mock.Anything, mock.Anything, accountUuid).Return(&pg_models.TemporalConfig{Namespace: "namespace"}, nil)
	mockNamespaceClient := new(MockNamespaceClient)
	m.TemporalWfManagerMock.On("GetNamespaceClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(mockNamespaceClient, nil)
	mockNamespaceClient.On("Describe", mock.Anything, "namespace").Return(&workflowservice.DescribeNamespaceResponse{}, nil)
	m.TemporalWfManagerMock.On("GetScheduleClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(mockScheduleClient, nil)
	mockScheduleClient.On("Create", mock.Anything, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Trigger", mock.Anything, mock.Anything).Return(nil)
	mockHandle.On("GetID").Return(nucleusdb.UUIDString(job1.ID))

	m.QuerierMock.On("CreateJob", mock.Anything, mockTx, db_queries.CreateJobParams{
		Name:               job1.Name,
		AccountID:          accountUuid,
		Status:             int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
		CronSchedule:       cron,
		ConnectionSourceID: srcConn.ID,
		ConnectionOptions: &pg_models.JobSourceOptions{
			PostgresOptions: &pg_models.PostgresSourceOptions{
				HaltOnNewColumnAddition: true,
				Schemas: []*pg_models.PostgresSourceSchemaOption{
					{Schema: "schema-1", Tables: []*pg_models.PostgresSourceTableOption{
						{Table: "table-1", WhereClause: &whereClause},
					}},
					{Schema: "schema-2", Tables: []*pg_models.PostgresSourceTableOption{
						{Table: "table-2", WhereClause: &whereClause},
					}},
				},
			},
		},
		Mappings: []*pg_models.JobMapping{
			{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &pg_models.Transformer{
				Value:  "passthrough",
				Config: &pg_models.TransformerConfigs{},
			}},
			{Schema: "schema-2", Table: "table-2", Column: "col", Transformer: &pg_models.Transformer{
				Value:  "passthrough",
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		CreatedByID: userUuid,
		UpdatedByID: userUuid,
	}).Return(job1, nil)
	m.QuerierMock.On("CreateJobConnectionDestinations", mock.Anything, mockTx, destinationParams).Return(int64(1), nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job1.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)

	resp, err := m.Service.CreateJob(context.Background(), &connect.Request[mgmtv1alpha1.CreateJobRequest]{
		Msg: &mgmtv1alpha1.CreateJobRequest{
			AccountId:      mockAccountId,
			JobName:        job1.Name,
			CronSchedule:   &cronSchedule,
			InitiateJobRun: true,
			Source: &mgmtv1alpha1.JobSource{
				ConnectionId: nucleusdb.UUIDString(srcConn.ID),
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresSourceConnectionOptions{
							HaltOnNewColumnAddition: true,
							Schemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
								{Schema: "schema-1", Tables: []*mgmtv1alpha1.PostgresSourceTableOption{
									{Table: "table-1", WhereClause: &whereClause},
								}},
								{Schema: "schema-2", Tables: []*mgmtv1alpha1.PostgresSourceTableOption{
									{Table: "table-2", WhereClause: &whereClause},
								}},
							},
						},
					},
				},
			},
			Destinations: []*mgmtv1alpha1.CreateJobDestination{
				{ConnectionId: nucleusdb.UUIDString(destConn.ID), Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: true,
								Cascade:              true,
							},
							InitTableSchema: true,
						},
					},
				}},
			},

			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.Transformer{
					Value:  "passthrough",
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "schema-2", Table: "table-2", Column: "col", Transformer: &mgmtv1alpha1.Transformer{
					Value:  "passthrough",
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// CreateJobDestinationConnections
func Test_CreateJobDestinationConnections(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(MockScheduleHandle)
	mockScheduleClient := new(MockScheduleClient)
	mockScheduleClient.Handle = mockHandle

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn := getConnectionMock(mockAccountId, "test-1", nil)
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetConnectionsByIds", mock.Anything, mock.Anything, []pgtype.UUID{destConn.ID}).Return([]db_queries.NeosyncApiConnection{destConn}, nil)
	m.QuerierMock.On("CreateJobConnectionDestinations", mock.Anything, mock.Anything, []db_queries.CreateJobConnectionDestinationsParams{
		{
			JobID:        job.ID,
			ConnectionID: destConn.ID,
			Options: &pg_models.JobDestinationOptions{
				PostgresOptions: &pg_models.PostgresDestinationOptions{
					TruncateTableConfig: &pg_models.PostgresTruncateTableConfig{
						TruncateBeforeInsert: true,
						TruncateCascade:      true,
					},
					InitTableSchema: true,
				},
			},
		},
	}).Return(int64(1), nil)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)

	jobId := nucleusdb.UUIDString(job.ID)

	resp, err := m.Service.CreateJobDestinationConnections(context.Background(), &connect.Request[mgmtv1alpha1.CreateJobDestinationConnectionsRequest]{
		Msg: &mgmtv1alpha1.CreateJobDestinationConnectionsRequest{
			JobId: jobId,
			Destinations: []*mgmtv1alpha1.CreateJobDestination{{
				ConnectionId: nucleusdb.UUIDString(destConn.ID),
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: true,
								Cascade:              true,
							},
							InitTableSchema: true,
						},
					},
				},
			}},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_CreateJobDestinationConnections_ConnectionNotInAccount(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(MockScheduleHandle)
	mockScheduleClient := new(MockScheduleClient)
	mockScheduleClient.Handle = mockHandle

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn := getConnectionMock(uuid.NewString(), "test-1", nil)
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetConnectionsByIds", mock.Anything, mock.Anything, []pgtype.UUID{destConn.ID}).Return([]db_queries.NeosyncApiConnection{destConn}, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)

	jobId := nucleusdb.UUIDString(job.ID)

	resp, err := m.Service.CreateJobDestinationConnections(context.Background(), &connect.Request[mgmtv1alpha1.CreateJobDestinationConnectionsRequest]{
		Msg: &mgmtv1alpha1.CreateJobDestinationConnectionsRequest{
			JobId: jobId,
			Destinations: []*mgmtv1alpha1.CreateJobDestination{{
				ConnectionId: nucleusdb.UUIDString(destConn.ID),
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: true,
								Cascade:              true,
							},
							InitTableSchema: true,
						},
					},
				},
			}},
		},
	})

	m.QuerierMock.AssertNotCalled(t, "CreateJobConnectionDestinations", mock.Anything, mock.Anything, mock.Anything)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// UpdateJobSchedule

func Test_UpdateJobSchedule(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn := getConnectionMock(mockAccountId, "test-1", nil)
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	cronSchedule := "* * * * *"
	cron := pgtype.Text{}
	cron.Scan(cronSchedule)
	jobId := nucleusdb.UUIDString(job.ID)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	m.DbtxMock.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTx.On("Commit", mock.Anything).Return(nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("UpdateJobSchedule", mock.Anything, mock.Anything, db_queries.UpdateJobScheduleParams{
		ID:           job.ID,
		CronSchedule: cron,
		UpdatedByID:  userUuid,
	}).Return(job, nil)
	mockHandle := new(MockScheduleHandle)
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Update", mock.Anything, mock.Anything).Return(nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)

	resp, err := m.Service.UpdateJobSchedule(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobScheduleRequest]{
		Msg: &mgmtv1alpha1.UpdateJobScheduleRequest{
			Id:           jobId,
			CronSchedule: &cronSchedule,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

type serviceMocks struct {
	Service                     *Service
	DbtxMock                    *nucleusdb.MockDBTX
	QuerierMock                 *db_queries.MockQuerier
	UserAccountServiceMock      *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceClientMock *mgmtv1alpha1connect.MockConnectionServiceClient
	TemporalWfManagerMock       *clientmanager.MockTemporalClientManagerClient
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTemporalWfManager := clientmanager.NewMockTemporalClientManagerClient(t)

	service := New(config, nucleusdb.New(mockDbtx, mockQuerier), mockTemporalWfManager, mockConnectionService, mockUserAccountService)

	return &serviceMocks{
		Service:                     service,
		DbtxMock:                    mockDbtx,
		QuerierMock:                 mockQuerier,
		UserAccountServiceMock:      mockUserAccountService,
		ConnectionServiceClientMock: mockConnectionService,
		TemporalWfManagerMock:       mockTemporalWfManager,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

//nolint:all
func mockUserAccountCalls(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	mockIsUserInAccount(userAccountServiceMock, isInAccount)
	userAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
}

//nolint:all
func mockJob(accountId, userId, srcConnId string) db_queries.NeosyncApiJob {
	id, _ := nucleusdb.ToUuid(uuid.NewString())
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(userId)
	srcConnUuid, _ := nucleusdb.ToUuid(srcConnId)
	currentTime := time.Now()
	var timestamp pgtype.Timestamp
	timestamp.Time = currentTime
	return db_queries.NeosyncApiJob{
		ID:                 id,
		AccountID:          accountUuid,
		CreatedAt:          timestamp,
		UpdatedAt:          timestamp,
		CreatedByID:        userUuid,
		UpdatedByID:        userUuid,
		Status:             0,
		Name:               "some-name",
		ConnectionSourceID: srcConnUuid,
		ConnectionOptions:  &pg_models.JobSourceOptions{},
		Mappings:           []*pg_models.JobMapping{},
	}

}

func mockJobDestConnAssociation(jobUuid, connectionUuid pgtype.UUID, options *pg_models.JobDestinationOptions) db_queries.NeosyncApiJobDestinationConnectionAssociation {
	idUuid, _ := nucleusdb.ToUuid(uuid.NewString())
	timestamp := pgtype.Timestamp{
		Time: time.Now(),
	}
	return db_queries.NeosyncApiJobDestinationConnectionAssociation{
		ID:           idUuid,
		JobID:        jobUuid,
		CreatedAt:    timestamp,
		UpdatedAt:    timestamp,
		ConnectionID: connectionUuid,
		Options:      options,
	}
}

func getConnectionMock(accountId, name string, id *pgtype.UUID) db_queries.NeosyncApiConnection {
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)

	currentTime := time.Now()
	var timestamp pgtype.Timestamp
	timestamp.Time = currentTime

	sslMode := "disable"

	connUuid, _ := nucleusdb.ToUuid(uuid.NewString())
	if id != nil {
		connUuid = *id
	}
	return db_queries.NeosyncApiConnection{
		AccountID:   accountUuid,
		Name:        name,
		ID:          connUuid,
		CreatedByID: userUuid,
		UpdatedByID: userUuid,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		ConnectionConfig: &pg_models.ConnectionConfig{
			PgConfig: &pg_models.PostgresConnectionConfig{
				Connection: &pg_models.PostgresConnection{
					Host:    "host",
					Port:    5432,
					Name:    "database",
					User:    "user",
					Pass:    "topsecret",
					SslMode: &sslMode,
				},
			},
		},
	}
}
