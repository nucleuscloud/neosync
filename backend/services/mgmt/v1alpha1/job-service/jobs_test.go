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
	destConn := getConnectionMock(mockAccountId, "test-1")
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
	srcConn := getConnectionMock(mockAccountId, "test-4")
	destConn := getConnectionMock(mockAccountId, "test-1")
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
	_ = cron.Scan(cronSchedule)
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
	mockDbTransaction(m.DbtxMock, mockTx)
	m.QuerierMock.On("AreConnectionsInAccount", mock.Anything, mock.Anything, db_queries.AreConnectionsInAccountParams{
		AccountId:     accountUuid,
		ConnectionIds: []pgtype.UUID{destConn.ID},
	}).Return(int64(1), nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: srcConn.ID,
	}).Return(int64(1), nil)
	m.QuerierMock.On("GetConnectionById", mock.Anything, mock.Anything, srcConn.ID).Return(srcConn, nil)
	m.QuerierMock.On("GetConnectionById", mock.Anything, mock.Anything, destConn.ID).Return(destConn, nil)
	m.TemporalWfManagerMock.On("DoesAccountHaveTemporalWorkspace", mock.Anything, mockAccountId, mock.Anything).Return(true, nil)
	m.TemporalWfManagerMock.On("GetTemporalConfigByAccount", mock.Anything, mockAccountId).Return(&pg_models.TemporalConfig{
		Namespace:        "default",
		SyncJobQueueName: "sync-job",
		Url:              "localhost:7233",
	}, nil)
	m.TemporalWfManagerMock.On("GetScheduleClientByAccount", mock.Anything, mockAccountId, mock.Anything).Return(mockScheduleClient, nil)
	mockScheduleClient.On("Create", mock.Anything, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Trigger", mock.Anything, mock.Anything).Return(nil)
	mockHandle.On("GetID").Return(nucleusdb.UUIDString(job1.ID))

	m.QuerierMock.On("CreateJob", mock.Anything, mockTx, db_queries.CreateJobParams{
		Name:         job1.Name,
		AccountID:    accountUuid,
		Status:       int16(mgmtv1alpha1.JobStatus_JOB_STATUS_ENABLED),
		CronSchedule: cron,
		ConnectionOptions: &pg_models.JobSourceOptions{
			PostgresOptions: &pg_models.PostgresSourceOptions{
				ConnectionId:            nucleusdb.UUIDString(srcConn.ID),
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
			{Schema: "schema-1", Table: "table-1", Column: "col", JobMappingTransformer: &pg_models.JobMappingTransformerModel{
				Source: "passthrough",
				Config: &pg_models.TransformerConfigs{},
			}},
			{Schema: "schema-2", Table: "table-2", Column: "col", JobMappingTransformer: &pg_models.JobMappingTransformerModel{
				Source: "passthrough",
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
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
						Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
							ConnectionId:            nucleusdb.UUIDString(srcConn.ID),
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
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: "passthrough",
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "schema-2", Table: "table-2", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: "passthrough",
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
				},
			},
		},
	}).Return(int64(1), nil)

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
	destConn := getConnectionMock(uuid.NewString(), "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})

	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	m.QuerierMock.On("GetConnectionsByIds", mock.Anything, mock.Anything, []pgtype.UUID{destConn.ID}).Return([]db_queries.NeosyncApiConnection{destConn}, nil)

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
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	cronSchedule := "* * * * *"
	cron := pgtype.Text{}
	_ = cron.Scan(cronSchedule)
	jobId := nucleusdb.UUIDString(job.ID)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockDbTransaction(m.DbtxMock, mockTx)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation})
	m.QuerierMock.On("UpdateJobSchedule", mock.Anything, mock.Anything, db_queries.UpdateJobScheduleParams{
		ID:           job.ID,
		CronSchedule: cron,
		UpdatedByID:  userUuid,
	}).Return(job, nil)
	mockHandle := new(MockScheduleHandle)
	m.TemporalWfManagerMock.On("GetScheduleHandleClientByAccount", mock.Anything, mockAccountId, jobId, mock.Anything).Return(mockHandle, nil)
	mockHandle.On("Update", mock.Anything, mock.Anything).Return(nil)

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

// PauseJob
func Test_PauseJob_Pause(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(MockScheduleHandle)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := nucleusdb.UUIDString(job.ID)

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

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_PauseJob_UnPause(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(MockScheduleHandle)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := nucleusdb.UUIDString(job.ID)

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

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// UpdateJobSourceConnection
func Test_UpdateJobSourceConnection_Success(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	conn := getConnectionMock(mockAccountId, "test-1")
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	whereClause := "where1"

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockDbTransaction(m.DbtxMock, mockTx)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{})
	m.QuerierMock.On("UpdateJobSource", mock.Anything, mockTx, db_queries.UpdateJobSourceParams{
		ID: job.ID,
		ConnectionOptions: &pg_models.JobSourceOptions{
			PostgresOptions: &pg_models.PostgresSourceOptions{
				ConnectionId:            nucleusdb.UUIDString(conn.ID),
				HaltOnNewColumnAddition: true,
				Schemas: []*pg_models.PostgresSourceSchemaOption{
					{Schema: "schema-1", Tables: []*pg_models.PostgresSourceTableOption{
						{Table: "table-1", WhereClause: &whereClause},
					}},
				},
			},
		},
		UpdatedByID: userUuid,
	}).Return(job, nil)
	m.QuerierMock.On("UpdateJobMappings", mock.Anything, mockTx, db_queries.UpdateJobMappingsParams{
		ID: job.ID,
		Mappings: []*pg_models.JobMapping{
			{Schema: "schema-1", Table: "table-1", Column: "col", JobMappingTransformer: &pg_models.JobMappingTransformerModel{
				Source: "passthrough",
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		UpdatedByID: userUuid,
	}).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: conn.ID,
	}).Return(int64(1), nil)

	resp, err := m.Service.UpdateJobSourceConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateJobSourceConnectionRequest{
			Id: nucleusdb.UUIDString(job.ID),
			Source: &mgmtv1alpha1.JobSource{
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
						Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
							ConnectionId:            nucleusdb.UUIDString(conn.ID),
							HaltOnNewColumnAddition: true,
							Schemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
								{Schema: "schema-1", Tables: []*mgmtv1alpha1.PostgresSourceTableOption{
									{Table: "table-1", WhereClause: &whereClause},
								}},
							},
						},
					},
				},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: "passthrough",
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// SetJobSourceSqlConnectionSubsets
func Test_SetJobSourceSqlConnectionSubsets_Invalid_Connection_No_ConnectionId(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
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
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// UpdateJobDestinationConnection
func Test_UpdateJobDestinationConnection_Update(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
	destinationId := uuid.NewString()
	destinationUuid, _ := nucleusdb.ToUuid(destinationId)
	connectionId := uuid.NewString()
	connectionUuid, _ := nucleusdb.ToUuid(connectionId)
	updatedOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
			PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
				TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					Cascade:              true,
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
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, jobId, resp.Msg.Job.Id)
}

func Test_UpdateJobDestinationConnection_Create(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	jobId := nucleusdb.UUIDString(job.ID)
	destinationId := uuid.NewString()
	connectionId := uuid.NewString()
	connectionUuid, _ := nucleusdb.ToUuid(connectionId)
	updatedOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
			PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
				TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
					TruncateBeforeInsert: true,
					Cascade:              true,
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

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, jobId, resp.Msg.Job.Id)
}

func Test_DeleteJobDestinationConnection(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString())
	destinationId := uuid.NewString()
	destinationUuid, _ := nucleusdb.ToUuid(destinationId)
	connId, _ := nucleusdb.ToUuid(uuid.NewString())
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

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func Test_DeleteJobDestinationConnection_NotFound(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	destinationId := uuid.NewString()
	destinationUuid, _ := nucleusdb.ToUuid(destinationId)
	var nilDestConnAssociation db_queries.NeosyncApiJobDestinationConnectionAssociation

	m.QuerierMock.On("GetJobConnectionDestination", mock.Anything, mock.Anything, destinationUuid).Return(nilDestConnAssociation, sql.ErrNoRows)

	resp, err := m.Service.DeleteJobDestinationConnection(context.Background(), &connect.Request[mgmtv1alpha1.DeleteJobDestinationConnectionRequest]{
		Msg: &mgmtv1alpha1.DeleteJobDestinationConnectionRequest{
			DestinationId: destinationId,
		},
	})

	m.QuerierMock.AssertNotCalled(t, "RemoveJobConnectionDestination", mock.Anything, mock.Anything, mock.Anything)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// IsJobNameAvailable
func Test_IsJobNameAvailable_Available(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	jobName := "unique_job_name"
	accountId := mockAccountId
	accountUuid, _ := nucleusdb.ToUuid(accountId)

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

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Msg.IsAvailable)
}

func Test_IsJobNameAvailable_NotAvailable(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	jobName := "existing_job_name"
	accountId := mockAccountId
	accountUuid, _ := nucleusdb.ToUuid(accountId)

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

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Msg.IsAvailable)
}

type serviceMocks struct {
	Service                     *Service
	DbtxMock                    *nucleusdb.MockDBTX
	QuerierMock                 *db_queries.MockQuerier
	UserAccountServiceMock      *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceClientMock *mgmtv1alpha1connect.MockConnectionServiceHandler
	TemporalWfManagerMock       *clientmanager.MockTemporalClientManagerClient
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceHandler(t)
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

func mockDbTransaction(dbtxMock *nucleusdb.MockDBTX, txMock *nucleusdb.MockTx) {
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
func mockJob(accountId, userId, srcConnId string) db_queries.NeosyncApiJob {
	id, _ := nucleusdb.ToUuid(uuid.NewString())
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(userId)
	// srcConnUuid, _ := nucleusdb.ToUuid(srcConnId)
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

func getConnectionMock(accountId, name string) db_queries.NeosyncApiConnection {
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)

	currentTime := time.Now()
	var timestamp pgtype.Timestamp
	timestamp.Time = currentTime

	sslMode := "disable"

	connUuid, _ := nucleusdb.ToUuid(uuid.NewString())
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
