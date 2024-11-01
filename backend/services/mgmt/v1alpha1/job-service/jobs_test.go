package v1alpha1_jobservice

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	pgxmock "github.com/nucleuscloud/neosync/internal/mocks/github.com/jackc/pgx/v5"

	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	temporal "go.temporal.io/sdk/client"
	temporalmocks "go.temporal.io/sdk/mocks"
)

const (
	anonymousUserId    = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider   = "test-provider"
	mockUserId         = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId      = "5629813e-1a35-4874-922c-9827d85f0378"
	mockConnectionName = "test-conn"
	mockConnectionId   = "884765c6-1708-488d-b03a-70a02b12c81e"
)

// GetJobs
func Test_GetJobs_UnauthorizedUser(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, false)

	resp, err := m.Service.GetJobs(context.Background(), &connect.Request[mgmtv1alpha1.GetJobsRequest]{
		Msg: &mgmtv1alpha1.GetJobsRequest{
			AccountId: mockAccountId,
		},
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func Test_GetJobs(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	accountUuid, _ := neosyncdb.ToUuid(mockAccountId)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	job2 := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn1 := getConnectionMock(mockAccountId, "test-1")
	destConn2 := getConnectionMock(mockAccountId, "test-2")
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

	job1Id := neosyncdb.UUIDString(job1.ID)
	job2Id := neosyncdb.UUIDString(job2.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, len(resp.Msg.GetJobs()))
	require.NotNil(t, jobActualMap[job1Id])
	require.NotNil(t, jobActualMap[job2Id])
	require.Equal(t, neosyncdb.UUIDString(destConn1.ID), jobActualMap[job1Id].Destinations[0].ConnectionId)
	require.Equal(t, neosyncdb.UUIDString(destConn2.ID), jobActualMap[job2Id].Destinations[0].ConnectionId)
}

func Test_GetJobs_MissingDestinations(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	accountUuid, _ := neosyncdb.ToUuid(mockAccountId)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	m.QuerierMock.On("GetJobsByAccount", context.Background(), mock.Anything, accountUuid).Return([]db_queries.NeosyncApiJob{job1}, nil)
	m.QuerierMock.On("GetJobConnectionDestinationsByJobIds", context.Background(), mock.Anything, []pgtype.UUID{job1.ID}).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	resp, err := m.Service.GetJobs(context.Background(), &connect.Request[mgmtv1alpha1.GetJobsRequest]{
		Msg: &mgmtv1alpha1.GetJobsRequest{
			AccountId: mockAccountId,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Msg.GetJobs()))
	require.Empty(t, resp.Msg.Jobs[0].Destinations)
}

// GetJob
func Test_GetJob(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)
	jobId := neosyncdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, neosyncdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_GetJob_Supports_WorkerApiKeys(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)
	jobId := neosyncdb.UUIDString(job.ID)

	ctx := context.WithValue(context.Background(), auth_apikey.TokenContextKey{}, &auth_apikey.TokenContextData{
		ApiKeyType: apikey.WorkerApiKey,
	})

	resp, err := m.Service.GetJob(ctx, &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, neosyncdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_GetJob_MissingDestinations(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	jobId := neosyncdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Msg.Job.Destinations)
}

func Test_GetJob_UnauthorizedUser(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, false)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	jobId := neosyncdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	require.Error(t, err)
	require.Nil(t, resp)
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

	require.Error(t, err)
	require.Nil(t, resp)
}

// GetJobStatus
func Test_GetJobStatus_Paused(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	jobId := neosyncdb.UUIDString(job.ID)

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, mock.Anything).Return(job, nil)

	mockHandle := new(temporalmocks.ScheduleHandle)

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

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, mgmtv1alpha1.JobStatus(3), resp.Msg.Status)
}

func Test_GetJobStatus_Enabled(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	jobId := neosyncdb.UUIDString(job.ID)

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)

	mockHandle := new(temporalmocks.ScheduleHandle)

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

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, mgmtv1alpha1.JobStatus(1), resp.Msg.Status)
}

// GetJobStatuses
func Test_GetJobStatuses(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(temporalmocks.ScheduleHandle)
	mockScheduleClient := new(temporalmocks.ScheduleClient)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})

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

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Msg.Statuses))
	require.Equal(t, mgmtv1alpha1.JobStatus(1), resp.Msg.Statuses[0].Status)
}

var (
	cronSchedule = "* * * * *"
	whereClause  = "where"
)

// CreateJobDestinationConnections
func Test_CreateJobDestinationConnections(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})

	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
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
					OnConflictConfig: &pg_models.PostgresOnConflictConfig{
						DoNothing: true,
					},
				},
			},
		},
	}).Return(int64(1), nil)

	jobId := neosyncdb.UUIDString(job.ID)

	resp, err := m.Service.CreateJobDestinationConnections(context.Background(), &connect.Request[mgmtv1alpha1.CreateJobDestinationConnectionsRequest]{
		Msg: &mgmtv1alpha1.CreateJobDestinationConnectionsRequest{
			JobId: jobId,
			Destinations: []*mgmtv1alpha1.CreateJobDestination{{
				ConnectionId: neosyncdb.UUIDString(destConn.ID),
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: true,
								Cascade:              true,
							},
							InitTableSchema: true,
							OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
								DoNothing: true,
							},
						},
					},
				},
			}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, neosyncdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_CreateJobDestinationConnections_ConnectionNotInAccount(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(uuid.NewString(), "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})

	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	m.QuerierMock.On("GetConnectionsByIds", mock.Anything, mock.Anything, []pgtype.UUID{destConn.ID}).Return([]db_queries.NeosyncApiConnection{destConn}, nil)

	jobId := neosyncdb.UUIDString(job.ID)

	resp, err := m.Service.CreateJobDestinationConnections(context.Background(), &connect.Request[mgmtv1alpha1.CreateJobDestinationConnectionsRequest]{
		Msg: &mgmtv1alpha1.CreateJobDestinationConnectionsRequest{
			JobId: jobId,
			Destinations: []*mgmtv1alpha1.CreateJobDestination{{
				ConnectionId: neosyncdb.UUIDString(destConn.ID),
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: true,
								Cascade:              true,
							},
							InitTableSchema: true,
							OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
								DoNothing: true,
							},
						},
					},
				},
			}},
		},
	})

	m.QuerierMock.AssertNotCalled(t, "CreateJobConnectionDestinations", mock.Anything, mock.Anything, mock.Anything)
	require.Error(t, err)
	require.Nil(t, resp)
}

// UpdateJobSchedule
func Test_UpdateJobSchedule(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := pgxmock.NewMockTx(t)

	userUuid, _ := neosyncdb.ToUuid(mockUserId)
	cron := pgtype.Text{}
	_ = cron.Scan(cronSchedule)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), cron)
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})

	jobId := neosyncdb.UUIDString(job.ID)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockDbTransaction(m.DbtxMock, mockTx)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	m.QuerierMock.On("UpdateJobSchedule", mock.Anything, mock.Anything, db_queries.UpdateJobScheduleParams{
		ID:           job.ID,
		CronSchedule: cron,
		UpdatedByID:  userUuid,
	}).Return(job, nil)
	mockHandle := new(temporalmocks.ScheduleHandle)
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Update", mock.Anything, mock.Anything).Return(nil)

	resp, err := m.Service.UpdateJobSchedule(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobScheduleRequest]{
		Msg: &mgmtv1alpha1.UpdateJobScheduleRequest{
			Id:           jobId,
			CronSchedule: &cronSchedule,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, neosyncdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

// PauseJob
func Test_PauseJob_Pause(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(temporalmocks.ScheduleHandle)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := neosyncdb.UUIDString(job.ID)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Pause", mock.Anything, mock.Anything).Return(nil)

	resp, err := m.Service.PauseJob(context.Background(), &connect.Request[mgmtv1alpha1.PauseJobRequest]{
		Msg: &mgmtv1alpha1.PauseJobRequest{
			Id:    jobId,
			Pause: true,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_PauseJob_UnPause(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(temporalmocks.ScheduleHandle)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := neosyncdb.UUIDString(job.ID)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Unpause", mock.Anything, mock.Anything).Return(nil)

	resp, err := m.Service.PauseJob(context.Background(), &connect.Request[mgmtv1alpha1.PauseJobRequest]{
		Msg: &mgmtv1alpha1.PauseJobRequest{
			Id:    jobId,
			Pause: false,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

// SetJobSourceSqlConnectionSubsets
func Test_SetJobSourceSqlConnectionSubsets_Invalid_Connection_No_ConnectionId(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	jobId := neosyncdb.UUIDString(job.ID)
	whereClause := "where2"

	mockUserAccountCalls(m.UserAccountServiceMock, true)

	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)

	resp, err := m.Service.SetJobSourceSqlConnectionSubsets(context.Background(), &connect.Request[mgmtv1alpha1.SetJobSourceSqlConnectionSubsetsRequest]{
		Msg: &mgmtv1alpha1.SetJobSourceSqlConnectionSubsetsRequest{
			Id: jobId,
			Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas{
				Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas_PostgresSubset{
					PostgresSubset: &mgmtv1alpha1.PostgresSourceSchemaSubset{
						PostgresSchemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
							{Schema: "schema-1", Tables: []*mgmtv1alpha1.PostgresSourceTableOption{{
								Table:       "table-1",
								WhereClause: &whereClause,
							}}},
						},
					},
				},
			},
		},
	})

	m.QuerierMock.AssertNotCalled(t, "SetSqlSourceSubsets", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	require.Error(t, err)
	require.Nil(t, resp)
}

// UpdateJobDestinationConnection
func Test_UpdateJobDestinationConnection_Update(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	jobId := neosyncdb.UUIDString(job.ID)
	destinationId := uuid.NewString()
	destinationUuid, _ := neosyncdb.ToUuid(destinationId)
	connectionId := uuid.NewString()
	connectionUuid, _ := neosyncdb.ToUuid(connectionId)
	updatedOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
			PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
				TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					Cascade:              true,
				},
				OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
					DoNothing: true,
				},
			},
		},
	}

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{})
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	m.QuerierMock.On("UpdateJobConnectionDestination", mock.Anything, mock.Anything, db_queries.UpdateJobConnectionDestinationParams{
		ID:           destinationUuid,
		ConnectionID: connectionUuid,
		Options: &pg_models.JobDestinationOptions{
			PostgresOptions: &pg_models.PostgresDestinationOptions{
				TruncateTableConfig: &pg_models.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					TruncateCascade:      true,
				},
				OnConflictConfig: &pg_models.PostgresOnConflictConfig{
					DoNothing: true,
				},
			},
		},
	}).Return(db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)

	resp, err := m.Service.UpdateJobDestinationConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobDestinationConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateJobDestinationConnectionRequest{
			JobId:         jobId,
			DestinationId: destinationId,
			ConnectionId:  connectionId,
			Options:       updatedOptions,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "CreateJobConnectionDestination", mock.Anything, mock.Anything, mock.Anything)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, jobId, resp.Msg.Job.Id)
}

func Test_UpdateJobDestinationConnection_Create(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	jobId := neosyncdb.UUIDString(job.ID)
	destinationId := uuid.NewString()
	connectionId := uuid.NewString()
	connectionUuid, _ := neosyncdb.ToUuid(connectionId)
	updatedOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
			PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
				TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					Cascade:              true,
				},
				OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
					DoNothing: true,
				},
			},
		},
	}
	var nilDestConnAssociation db_queries.NeosyncApiJobDestinationConnectionAssociation

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{})
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	m.QuerierMock.On("UpdateJobConnectionDestination", mock.Anything, mock.Anything, mock.Anything).Return(nilDestConnAssociation, sql.ErrNoRows)
	m.QuerierMock.On("CreateJobConnectionDestination", mock.Anything, mock.Anything, db_queries.CreateJobConnectionDestinationParams{
		JobID:        job.ID,
		ConnectionID: connectionUuid,
		Options: &pg_models.JobDestinationOptions{
			PostgresOptions: &pg_models.PostgresDestinationOptions{
				TruncateTableConfig: &pg_models.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					TruncateCascade:      true,
				},
				OnConflictConfig: &pg_models.PostgresOnConflictConfig{
					DoNothing: true,
				},
			},
		},
	}).Return(db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)

	resp, err := m.Service.UpdateJobDestinationConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobDestinationConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateJobDestinationConnectionRequest{
			JobId:         jobId,
			DestinationId: destinationId,
			ConnectionId:  connectionId,
			Options:       updatedOptions,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, jobId, resp.Msg.Job.Id)
}

func Test_DeleteJobDestinationConnection(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destinationId := uuid.NewString()
	destinationUuid, _ := neosyncdb.ToUuid(destinationId)
	connId, _ := neosyncdb.ToUuid(uuid.NewString())
	destinationConn := mockJobDestConnAssociation(job.ID, connId, nil)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobConnectionDestination", mock.Anything, mock.Anything, destinationUuid).Return(destinationConn, nil)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	m.QuerierMock.On("RemoveJobConnectionDestination", mock.Anything, mock.Anything, destinationUuid).Return(nil)

	resp, err := m.Service.DeleteJobDestinationConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteJobDestinationConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteJobDestinationConnectionRequest{
			DestinationId: destinationId,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_DeleteJobDestinationConnection_NotFound(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	destinationId := uuid.NewString()
	destinationUuid, _ := neosyncdb.ToUuid(destinationId)
	var nilDestConnAssociation db_queries.NeosyncApiJobDestinationConnectionAssociation

	m.QuerierMock.On("GetJobConnectionDestination", mock.Anything, mock.Anything, destinationUuid).Return(nilDestConnAssociation, sql.ErrNoRows)

	resp, err := m.Service.DeleteJobDestinationConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteJobDestinationConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteJobDestinationConnectionRequest{
			DestinationId: destinationId,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "RemoveJobConnectionDestination", mock.Anything, mock.Anything, mock.Anything)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// IsJobNameAvailable
func Test_IsJobNameAvailable_Available(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	jobName := "unique_job_name"
	accountId := mockAccountId
	accountUuid, _ := neosyncdb.ToUuid(accountId)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("IsJobNameAvailable", mock.Anything, mock.Anything, db_queries.IsJobNameAvailableParams{
		AccountId: accountUuid,
		JobName:   jobName,
	}).Return(int64(0), nil)

	resp, err := m.Service.IsJobNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsJobNameAvailableRequest]{
		Msg: &mgmtv1alpha1.IsJobNameAvailableRequest{
			AccountId: accountId,
			Name:      jobName,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Msg.IsAvailable)
}

func Test_IsJobNameAvailable_NotAvailable(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	jobName := "existing_job_name"
	accountId := mockAccountId
	accountUuid, _ := neosyncdb.ToUuid(accountId)

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("IsJobNameAvailable", mock.Anything, mock.Anything, db_queries.IsJobNameAvailableParams{
		AccountId: accountUuid,
		JobName:   jobName,
	}).Return(int64(1), nil)

	resp, err := m.Service.IsJobNameAvailable(context.Background(), &connect.Request[mgmtv1alpha1.IsJobNameAvailableRequest]{
		Msg: &mgmtv1alpha1.IsJobNameAvailableRequest{
			AccountId: accountId,
			Name:      jobName,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Msg.IsAvailable)
}

type serviceMocks struct {
	Service                     *Service
	DbtxMock                    *neosyncdb.MockDBTX
	QuerierMock                 *db_queries.MockQuerier
	UserAccountServiceMock      *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceClientMock *mgmtv1alpha1connect.MockConnectionServiceClient
	TemporalWfManagerMock       *clientmanager.MockTemporalClientManagerClient
	SqlManagerMock              *sql_manager.MockSqlManagerClient
	SqlDbMock                   *sql_manager.MockSqlDatabase
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := neosyncdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTemporalWfManager := clientmanager.NewMockTemporalClientManagerClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	service := New(config, neosyncdb.New(mockDbtx, mockQuerier), mockTemporalWfManager, mockConnectionService, mockUserAccountService, mockSqlManager)

	return &serviceMocks{
		Service:                     service,
		DbtxMock:                    mockDbtx,
		QuerierMock:                 mockQuerier,
		UserAccountServiceMock:      mockUserAccountService,
		ConnectionServiceClientMock: mockConnectionService,
		TemporalWfManagerMock:       mockTemporalWfManager,
		SqlManagerMock:              mockSqlManager,
		SqlDbMock:                   mockSqlDb,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

func mockDbTransaction(dbtxMock *neosyncdb.MockDBTX, txMock *pgxmock.MockTx) {
	dbtxMock.On("Begin", mock.Anything).Return(txMock, nil)
	txMock.On("Commit", mock.Anything).Return(nil)
	txMock.On("Rollback", mock.Anything).Return(nil)
}

//nolint:all
func mockGetJob(
	userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient,
	querierMock *db_queries.MockQuerier,
	jobMock db_queries.NeosyncApiJob,
	destinationMocks []db_queries.NeosyncApiJobDestinationConnectionAssociation,
) {
	mockIsUserInAccount(userAccountServiceMock, true)
	querierMock.On("GetJobById", mock.Anything, mock.Anything, jobMock.ID).Return(jobMock, nil)
	querierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, jobMock.ID).Return(destinationMocks, nil)
}

//nolint:all
func mockUserAccountCalls(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	mockIsUserInAccount(userAccountServiceMock, isInAccount)
	userAccountServiceMock.On("GetUser", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetUserResponse{
		UserId: mockUserId,
	}), nil)
}

//nolint:all
func mockJob(accountId, userId, srcConnId string, cronSchedule pgtype.Text) db_queries.NeosyncApiJob {
	id, _ := neosyncdb.ToUuid(uuid.NewString())
	accountUuid, _ := neosyncdb.ToUuid(accountId)
	userUuid, _ := neosyncdb.ToUuid(userId)
	currentTime := time.Now()
	var timestamp pgtype.Timestamp
	timestamp.Time = currentTime
	return db_queries.NeosyncApiJob{
		ID:                id,
		AccountID:         accountUuid,
		CreatedAt:         timestamp,
		UpdatedAt:         timestamp,
		CreatedByID:       userUuid,
		UpdatedByID:       userUuid,
		Status:            0,
		Name:              "some-name",
		ConnectionOptions: &pg_models.JobSourceOptions{},
		Mappings:          []*pg_models.JobMapping{},
		CronSchedule:      cronSchedule,
	}

}

func mockJobDestConnAssociation(jobUuid, connectionUuid pgtype.UUID, options *pg_models.JobDestinationOptions) db_queries.NeosyncApiJobDestinationConnectionAssociation {
	idUuid, _ := neosyncdb.ToUuid(uuid.NewString())
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

func getConnectionMock(accountId, name string) db_queries.NeosyncApiConnection {
	accountUuid, _ := neosyncdb.ToUuid(accountId)
	userUuid, _ := neosyncdb.ToUuid(mockUserId)

	currentTime := time.Now()
	var timestamp pgtype.Timestamp
	timestamp.Time = currentTime

	sslMode := "disable"

	connUuid, _ := neosyncdb.ToUuid(uuid.NewString())
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

// SetJobWorkflowOptions
func Test_SetJobWorkflowOptions(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := pgxmock.NewMockTx(t)

	userUuid, _ := neosyncdb.ToUuid(mockUserId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := neosyncdb.UUIDString(job.ID)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	mockDbTransaction(m.DbtxMock, mockTx)

	m.QuerierMock.On("SetJobWorkflowOptions", mock.Anything, mock.Anything, db_queries.SetJobWorkflowOptionsParams{
		ID: job.ID,
		WorkflowOptions: &pg_models.WorkflowOptions{
			RunTimeout: ptr(int64(123)),
		},
		UpdatedByID: userUuid,
	}).Return(job, nil)
	mockHandle := new(temporalmocks.ScheduleHandle)
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Update", mock.Anything, mock.Anything).Return(nil)

	resp, err := m.Service.SetJobWorkflowOptions(context.Background(), &connect.Request[mgmtv1alpha1.SetJobWorkflowOptionsRequest]{
		Msg: &mgmtv1alpha1.SetJobWorkflowOptionsRequest{
			Id: jobId,
			WorfklowOptions: &mgmtv1alpha1.WorkflowOptions{
				RunTimeout: ptr(int64(123)),
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

// SetJobSyncOptions
func Test_SetJobSyncOptions(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	userUuid, _ := neosyncdb.ToUuid(mockUserId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := neosyncdb.UUIDString(job.ID)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})

	m.QuerierMock.On("SetJobSyncOptions", mock.Anything, mock.Anything, db_queries.SetJobSyncOptionsParams{
		ID: job.ID,
		SyncOptions: &pg_models.ActivityOptions{
			ScheduleToCloseTimeout: ptr(int64(1234)),
			StartToCloseTimeout:    ptr(int64(123)),
			RetryPolicy: &pg_models.RetryPolicy{
				MaximumAttempts: ptr(int32(222)),
			},
		},
		UpdatedByID: userUuid,
	}).Return(job, nil)

	resp, err := m.Service.SetJobSyncOptions(context.Background(), &connect.Request[mgmtv1alpha1.SetJobSyncOptionsRequest]{
		Msg: &mgmtv1alpha1.SetJobSyncOptionsRequest{
			Id: jobId,
			SyncOptions: &mgmtv1alpha1.ActivityOptions{
				ScheduleToCloseTimeout: ptr(int64(1234)),
				StartToCloseTimeout:    ptr(int64(123)),
				RetryPolicy: &mgmtv1alpha1.RetryPolicy{
					MaximumAttempts: ptr(int32(222)),
				},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_getDurationFromInt(t *testing.T) {
	require.Equal(t, getDurationFromInt(nil), time.Duration(0))
	require.Equal(t, getDurationFromInt(ptr(int64(0))), time.Duration(0))
	require.Equal(t, getDurationFromInt(ptr(int64(1))), time.Duration(1))
}

// ValidateJobMappings
func Test_ValidateJobMappings_NoValidationErrors(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	conn := getConnectionMock(mockAccountId, "test-4")
	connId := neosyncdb.UUIDString(conn.ID)

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: connId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{},
				},
			},
		},
	}), nil)

	m.SqlManagerMock.On("NewSqlDb", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: m.SqlDbMock, Driver: sqlmanager_shared.PostgresDriver}, nil)
	m.SqlDbMock.On("Close").Return(nil)
	m.SqlDbMock.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"public.users":   {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "name": &sqlmanager_shared.DatabaseSchemaRow{IsNullable: true}},
		"public.orders":  {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "buyer_id": &sqlmanager_shared.DatabaseSchemaRow{IsNullable: true}},
		"circle.table_1": {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "table2_id": &sqlmanager_shared.DatabaseSchemaRow{IsNullable: true}},
		"circle.table_2": {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "table1_id": &sqlmanager_shared.DatabaseSchemaRow{}},
	}, nil)
	m.SqlDbMock.On("GetTableConstraintsBySchema", mock.Anything, mock.Anything).Return(&sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.orders":  {{Columns: []string{"buyer_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
			"circle.table_1": {{Columns: []string{"table2_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "circle.table_2", Columns: []string{"id"}}}},
			"circle.table_2": {{Columns: []string{"table1_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "circle.table_1", Columns: []string{"id"}}}},
		},
		PrimaryKeyConstraints: map[string][]string{"public.users": {"id"},
			"public.orders": {"id"}},
	}, nil)

	resp, err := m.Service.ValidateJobMappings(context.Background(), &connect.Request[mgmtv1alpha1.ValidateJobMappingsRequest]{
		Msg: &mgmtv1alpha1.ValidateJobMappingsRequest{
			AccountId:    mockAccountId,
			ConnectionId: connId,
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "public", Table: "orders", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_1", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_1", Column: "table2_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_2", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_2", Column: "table1_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Msg.DatabaseErrors.Errors)
	require.Empty(t, resp.Msg.ColumnErrors)
}

func Test_ValidateJobMappings_ValidationErrors(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	conn := getConnectionMock(mockAccountId, "test-4")
	connId := neosyncdb.UUIDString(conn.ID)

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: connId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{},
				},
			},
		},
	}), nil)

	m.SqlManagerMock.On("NewSqlDb", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: m.SqlDbMock, Driver: sqlmanager_shared.PostgresDriver}, nil)
	m.SqlDbMock.On("Close").Return(nil)
	m.SqlDbMock.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"public.users":   {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "name": &sqlmanager_shared.DatabaseSchemaRow{}},
		"public.orders":  {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "buyer_id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"circle.table_1": {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "table2_id": &sqlmanager_shared.DatabaseSchemaRow{}},
		"circle.table_2": {"id": &sqlmanager_shared.DatabaseSchemaRow{}, "table1_id": &sqlmanager_shared.DatabaseSchemaRow{}},
	}, nil)
	m.SqlDbMock.On("GetTableConstraintsBySchema", mock.Anything, mock.Anything).Return(&sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.orders":  {{Columns: []string{"buyer_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
			"circle.table_1": {{Columns: []string{"table2_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "circle.table_2", Columns: []string{"id"}}}},
			"circle.table_2": {{Columns: []string{"table1_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "circle.table_1", Columns: []string{"id"}}}},
		},
		PrimaryKeyConstraints: map[string][]string{"public.users": {"id"},
			"public.orders": {"id"}},
	}, nil)

	resp, err := m.Service.ValidateJobMappings(context.Background(), &connect.Request[mgmtv1alpha1.ValidateJobMappingsRequest]{
		Msg: &mgmtv1alpha1.ValidateJobMappingsRequest{
			AccountId:    mockAccountId,
			ConnectionId: connId,
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "public", Table: "orders", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "account", Table: "accounts", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_1", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_1", Column: "table2_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_2", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
				{Schema: "circle", Table: "table_2", Column: "table1_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{}},
				}},
			},
		},
	})

	expectedTableError := "Table does not exist [account.accounts]"
	expectedCdError := "Unsupported circular dependency. At least one foreign key in circular dependency must be nullable."

	expectedColErros := []*mgmtv1alpha1.ColumnError{
		{
			Schema: "public",
			Table:  "users",
			Column: "id",
			Errors: []string{
				"Missing required foreign key. Table: public.users  Column: id",
			},
		},
		{
			Schema: "public",
			Table:  "orders",
			Column: "buyer_id",
			Errors: []string{
				"Violates not-null constraint. Missing required column. Table: public.orders  Column: buyer_id",
			},
		},
	}

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.ElementsMatch(t, expectedColErros, resp.Msg.ColumnErrors)
	require.Len(t, resp.Msg.DatabaseErrors.Errors, 2)
	require.Contains(t, resp.Msg.DatabaseErrors.Errors, expectedTableError)
	for _, actualErr := range resp.Msg.DatabaseErrors.Errors {
		if strings.Contains(actualErr, "Unsupported circular dependency") {
			require.Contains(t, actualErr, expectedCdError)
		}
	}
}

func ptr[T any](val T) *T {
	return &val
}
