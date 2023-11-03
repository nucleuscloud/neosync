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
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// Define your mock objects
type MockScheduleClient struct {
	mock.Mock
}

type MockHandle struct {
	mock.Mock
}

// Define the behavior of the MockHandle's Describe method
func (m *MockHandle) Describe(ctx context.Context) (*temporal.Schedule, error) {
	args := m.Called(ctx)
	return args.Get(0).(*temporal.Schedule), args.Error(1)
}

// Define the behavior of the MockScheduleClient's GetHandle method
func (m *MockScheduleClient) GetHandle(ctx context.Context, jobID string) *MockHandle {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*MockHandle)
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
	destConnAssociation1 := mockJobDestConnAssociation(job1.ID, destConn1.ID)
	destConnAssociation2 := mockJobDestConnAssociation(job2.ID, destConn2.ID)
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
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID)
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
func Test_GetJobStatus(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, mock.Anything).Return(job, nil)

	// Mock the Temporal client and its methods
	mockScheduleClient := new(MockScheduleClient)
	mockHandle := new(MockHandle)

	// Setup expectations for the ScheduleClient and its methods
	mockScheduleClient.On("GetHandle", mock.Anything, jobId).Return(mockHandle)
	mockHandle.On("Describe", mock.Anything).Return(&temporal.ScheduleDescription{
		Schedule: temporal.Schedule{
			State: &temporal.ScheduleState{
				Paused: true,
			},
		},
	}, nil)

	m.TemporalWfManagerMock.On("GetWorkflowClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(mockScheduleClient, nil)

	resp, err := m.Service.GetJobStatus(context.Background(), &connect.Request[mgmtv1alpha1.GetJobStatusRequest]{
		Msg: &mgmtv1alpha1.GetJobStatusRequest{
			JobId: jobId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
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

func mockUserAccountCalls(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	mockIsUserInAccount(userAccountServiceMock, isInAccount)
	userAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
}

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
		ConnectionOptions:  &jsonmodels.JobSourceOptions{},
		Mappings:           []*jsonmodels.JobMapping{},
	}

}

func mockJobDestConnAssociation(jobUuid, connectionUuid pgtype.UUID) db_queries.NeosyncApiJobDestinationConnectionAssociation {
	uuid, _ := nucleusdb.ToUuid(uuid.NewString())
	timestamp := pgtype.Timestamp{
		Time: time.Now(),
	}
	return db_queries.NeosyncApiJobDestinationConnectionAssociation{
		ID:           uuid,
		JobID:        jobUuid,
		CreatedAt:    timestamp,
		UpdatedAt:    timestamp,
		ConnectionID: connectionUuid,
		Options:      &jsonmodels.JobDestinationOptions{},
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
		ConnectionConfig: &jsonmodels.ConnectionConfig{
			PgConfig: &jsonmodels.PostgresConnectionConfig{
				Connection: &jsonmodels.PostgresConnection{
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
