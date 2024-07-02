package v1alpha1_jobservice

import (
	"context"
	"database/sql"
	"errors"
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

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
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
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
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

	job1Id := nucleusdb.UUIDString(job1.ID)
	job2Id := nucleusdb.UUIDString(job2.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, len(resp.Msg.GetJobs()))
	require.NotNil(t, jobActualMap[job1Id])
	require.NotNil(t, jobActualMap[job2Id])
	require.Equal(t, nucleusdb.UUIDString(destConn1.ID), jobActualMap[job1Id].Destinations[0].ConnectionId)
	require.Equal(t, nucleusdb.UUIDString(destConn2.ID), jobActualMap[job2Id].Destinations[0].ConnectionId)
}

func Test_GetJobs_MissingDestinations(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
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
	jobId := nucleusdb.UUIDString(job.ID)

	resp, err := m.Service.GetJob(context.Background(), &connect.Request[mgmtv1alpha1.GetJobRequest]{
		Msg: &mgmtv1alpha1.GetJobRequest{
			Id: jobId,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_GetJob_Supports_WorkerApiKeys(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{destConnAssociation}, nil)
	jobId := nucleusdb.UUIDString(job.ID)

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
	require.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_GetJob_MissingDestinations(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockIsUserInAccount(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("GetJobConnectionDestinations", mock.Anything, mock.Anything, job.ID).Return([]db_queries.NeosyncApiJobDestinationConnectionAssociation{}, nil)
	jobId := nucleusdb.UUIDString(job.ID)

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
	jobId := nucleusdb.UUIDString(job.ID)

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
	jobId := nucleusdb.UUIDString(job.ID)

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
	jobId := nucleusdb.UUIDString(job.ID)

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

// CreateJob
func Test_CreateJob(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)
	mockHandle := new(temporalmocks.ScheduleHandle)
	mockScheduleClient := new(temporalmocks.ScheduleClient)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	cron := pgtype.Text{}
	_ = cron.Scan(cronSchedule)

	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString(), cron)
	srcConn := getConnectionMock(mockAccountId, "test-4")
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job1.ID, destConn.ID, &pg_models.JobDestinationOptions{
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
	})

	destinationParams := []db_queries.CreateJobConnectionDestinationsParams{
		{JobID: job1.ID, ConnectionID: destConn.ID, Options: &pg_models.JobDestinationOptions{
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
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
			{Schema: "schema-2", Table: "table-2", Column: "col", JobMappingTransformer: &pg_models.JobMappingTransformerModel{
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
		CreatedByID:        userUuid,
		UpdatedByID:        userUuid,
		WorkflowOptions:    &pg_models.WorkflowOptions{},
		SyncOptions:        &pg_models.ActivityOptions{},
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
							OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
								DoNothing: true,
							},
						},
					},
				}},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "schema-2", Table: "table-2", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
			WorkflowOptions: &mgmtv1alpha1.WorkflowOptions{},
			SyncOptions:     &mgmtv1alpha1.ActivityOptions{},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

// CreateJob
func Test_CreateJob_Schedule_Creation_Error(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)
	mockScheduleClient := new(temporalmocks.ScheduleClient)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	cron := pgtype.Text{}
	_ = cron.Scan(cronSchedule)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString(), cron)
	srcConn := getConnectionMock(mockAccountId, "test-4")
	destConn := getConnectionMock(mockAccountId, "test-1")

	destinationParams := []db_queries.CreateJobConnectionDestinationsParams{
		{JobID: job1.ID, ConnectionID: destConn.ID, Options: &pg_models.JobDestinationOptions{
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
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
			{Schema: "schema-2", Table: "table-2", Column: "col", JobMappingTransformer: &pg_models.JobMappingTransformerModel{
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
		CreatedByID:        userUuid,
		UpdatedByID:        userUuid,
		WorkflowOptions:    &pg_models.WorkflowOptions{},
		SyncOptions:        &pg_models.ActivityOptions{},
	}).Return(job1, nil)
	m.QuerierMock.On("CreateJobConnectionDestinations", mock.Anything, mockTx, destinationParams).Return(int64(1), nil)

	mockScheduleClient.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("test: unable to create temporal schedule"))
	m.QuerierMock.On("RemoveJobById", mock.Anything, mock.Anything, job1.ID).Return(nil) // job deletion succeeds

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
							OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
								DoNothing: true,
							},
						},
					},
				}},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "schema-2", Table: "table-2", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
			WorkflowOptions: &mgmtv1alpha1.WorkflowOptions{},
			SyncOptions:     &mgmtv1alpha1.ActivityOptions{},
		},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "unable to create scheduled job")
	require.ErrorContains(t, err, "test: unable to create temporal schedule")
	require.Nil(t, resp)
}

// CreateJob
func Test_CreateJob_Schedule_Creation_Error_JobCleanup_Error(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)
	mockScheduleClient := new(temporalmocks.ScheduleClient)

	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	cron := pgtype.Text{}
	_ = cron.Scan(cronSchedule)
	job1 := mockJob(mockAccountId, mockUserId, uuid.NewString(), cron)
	srcConn := getConnectionMock(mockAccountId, "test-4")
	destConn := getConnectionMock(mockAccountId, "test-1")

	destinationParams := []db_queries.CreateJobConnectionDestinationsParams{
		{JobID: job1.ID, ConnectionID: destConn.ID, Options: &pg_models.JobDestinationOptions{
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
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
			{Schema: "schema-2", Table: "table-2", Column: "col", JobMappingTransformer: &pg_models.JobMappingTransformerModel{
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
		CreatedByID:        userUuid,
		UpdatedByID:        userUuid,
		WorkflowOptions:    &pg_models.WorkflowOptions{},
		SyncOptions:        &pg_models.ActivityOptions{},
	}).Return(job1, nil)
	m.QuerierMock.On("CreateJobConnectionDestinations", mock.Anything, mockTx, destinationParams).Return(int64(1), nil)

	mockScheduleClient.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("test: unable to create temporal schedule"))
	m.QuerierMock.On("RemoveJobById", mock.Anything, mock.Anything, job1.ID).Return(errors.New("test: unable to remove job")) // job deletion succeeds

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
							OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
								DoNothing: true,
							},
						},
					},
				}},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "schema-2", Table: "table-2", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
			WorkflowOptions: &mgmtv1alpha1.WorkflowOptions{},
			SyncOptions:     &mgmtv1alpha1.ActivityOptions{},
		},
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "unable to create scheduled job")
	require.ErrorContains(t, err, "test: unable to create temporal schedule")
	require.ErrorContains(t, err, "test: unable to remove job")
	require.Nil(t, resp)
}

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
	require.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

func Test_CreateJobDestinationConnections_ConnectionNotInAccount(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
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
	mockTx := new(nucleusdb.MockTx)

	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	cron := pgtype.Text{}
	_ = cron.Scan(cronSchedule)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), cron)
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})

	jobId := nucleusdb.UUIDString(job.ID)

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
	require.Equal(t, nucleusdb.UUIDString(destConn.ID), resp.Msg.Job.Destinations[0].ConnectionId)
}

// PauseJob
func Test_PauseJob_Pause(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(temporalmocks.ScheduleHandle)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
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

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_PauseJob_UnPause(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockHandle := new(temporalmocks.ScheduleHandle)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
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

	require.NoError(t, err)
	require.NotNil(t, resp)
}

// UpdateJobSourceConnection
func Test_UpdateJobSourceConnection_Success(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
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
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		UpdatedByID: userUuid,
	}).Return(job, nil)
	m.QuerierMock.On("UpdateJobVirtualForeignKeys", mock.Anything, mockTx, db_queries.UpdateJobVirtualForeignKeysParams{
		ID:                 job.ID,
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
		UpdatedByID:        userUuid,
	}).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: conn.ID,
	}).Return(int64(1), nil)
	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: mockConnectionId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{},
				},
			},
		},
	}), nil)

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
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_UpdateJobSourceConnection_GenerateSuccess(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	fkString := uuid.NewString()
	k, _ := nucleusdb.ToUuid(fkString)

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockDbTransaction(m.DbtxMock, mockTx)
	mockGetJob(m.UserAccountServiceMock, m.QuerierMock, job, []db_queries.NeosyncApiJobDestinationConnectionAssociation{})
	m.QuerierMock.On("UpdateJobSource", mock.Anything, mockTx, db_queries.UpdateJobSourceParams{
		ID: job.ID,
		ConnectionOptions: &pg_models.JobSourceOptions{
			GenerateOptions: &pg_models.GenerateSourceOptions{
				FkSourceConnectionId: &fkString,
				Schemas: []*pg_models.GenerateSourceSchemaOption{
					{Schema: "schema-1", Tables: []*pg_models.GenerateSourceTableOption{
						{Table: "table-1", RowCount: 1},
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
				Source: int32(mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH),
				Config: &pg_models.TransformerConfigs{},
			}},
		},
		UpdatedByID: userUuid,
	}).Return(job, nil)
	m.QuerierMock.On("UpdateJobVirtualForeignKeys", mock.Anything, mockTx, db_queries.UpdateJobVirtualForeignKeysParams{
		ID:                 job.ID,
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
		UpdatedByID:        userUuid,
	}).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: k,
	}).Return(int64(1), nil)
	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: mockConnectionId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{},
				},
			},
		},
	}), nil)
	resp, err := m.Service.UpdateJobSourceConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateJobSourceConnectionRequest{
			Id: nucleusdb.UUIDString(job.ID),
			Source: &mgmtv1alpha1.JobSource{
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_Generate{
						Generate: &mgmtv1alpha1.GenerateSourceOptions{
							FkSourceConnectionId: &fkString,
							Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
								{Schema: "schema-1", Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
									{Table: "table-1", RowCount: 1},
								}},
							},
						},
					},
				},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_UpdateJobSourceConnection_PgMismatchError(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	conn := getConnectionMock(mockAccountId, "test-1")
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	whereClause := "where1"

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: conn.ID,
	}).Return(int64(1), nil)
	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: mockConnectionId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{},
				},
			},
		},
	}), nil)

	_, err := m.Service.UpdateJobSourceConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateJobSourceConnectionRequest{
			Id: nucleusdb.UUIDString(job.ID),
			Source: &mgmtv1alpha1.JobSource{
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
						Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
							ConnectionId:            nucleusdb.UUIDString(conn.ID),
							HaltOnNewColumnAddition: true,
							Schemas: []*mgmtv1alpha1.MysqlSourceSchemaOption{
								{Schema: "schema-1", Tables: []*mgmtv1alpha1.MysqlSourceTableOption{
									{Table: "table-1", WhereClause: &whereClause},
								}},
							},
						},
					},
				},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	require.Error(t, err)
	m.QuerierMock.AssertNotCalled(t, "UpdateJobSource", mock.Anything, mock.Anything, mock.Anything)
	m.QuerierMock.AssertNotCalled(t, "UpdateJobMappings", mock.Anything, mock.Anything, mock.Anything)
}

func Test_UpdateJobSourceConnection_MysqlMismatchError(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	conn := getConnectionMock(mockAccountId, "test-1")
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	whereClause := "where1"

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: conn.ID,
	}).Return(int64(1), nil)
	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: mockConnectionId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{},
				},
			},
		},
	}), nil)

	_, err := m.Service.UpdateJobSourceConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest]{
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
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	require.Error(t, err)
	m.QuerierMock.AssertNotCalled(t, "UpdateJobSource", mock.Anything, mock.Anything, mock.Anything)
	m.QuerierMock.AssertNotCalled(t, "UpdateJobMappings", mock.Anything, mock.Anything, mock.Anything)
}

func Test_UpdateJobSourceConnection_AwsS3MismatchError(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	conn := getConnectionMock(mockAccountId, "test-1")
	accountUuid, _ := nucleusdb.ToUuid(mockAccountId)
	whereClause := "where1"

	mockUserAccountCalls(m.UserAccountServiceMock, true)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetJobById", mock.Anything, mock.Anything, job.ID).Return(job, nil)
	m.QuerierMock.On("IsConnectionInAccount", mock.Anything, mock.Anything, db_queries.IsConnectionInAccountParams{
		AccountId:    accountUuid,
		ConnectionId: conn.ID,
	}).Return(int64(1), nil)
	m.ConnectionServiceClientMock.On("GetConnection", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: mockConnectionId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
					AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{},
				},
			},
		},
	}), nil)

	_, err := m.Service.UpdateJobSourceConnection(context.Background(), &connect.Request[mgmtv1alpha1.UpdateJobSourceConnectionRequest]{
		Msg: &mgmtv1alpha1.UpdateJobSourceConnectionRequest{
			Id: nucleusdb.UUIDString(job.ID),
			Source: &mgmtv1alpha1.JobSource{
				Options: &mgmtv1alpha1.JobSourceOptions{
					Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
						Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
							ConnectionId:            nucleusdb.UUIDString(conn.ID),
							HaltOnNewColumnAddition: true,
							Schemas: []*mgmtv1alpha1.MysqlSourceSchemaOption{
								{Schema: "schema-1", Tables: []*mgmtv1alpha1.MysqlSourceTableOption{
									{Table: "table-1", WhereClause: &whereClause},
								}},
							},
						},
					},
				},
			},
			Mappings: []*mgmtv1alpha1.JobMapping{
				{Schema: "schema-1", Table: "table-1", Column: "col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
			},
		},
	})

	require.Error(t, err)
	m.QuerierMock.AssertNotCalled(t, "UpdateJobSource", mock.Anything, mock.Anything, mock.Anything)
	m.QuerierMock.AssertNotCalled(t, "UpdateJobMappings", mock.Anything, mock.Anything, mock.Anything)
}

// SetJobSourceSqlConnectionSubsets
func Test_SetJobSourceSqlConnectionSubsets_Invalid_Connection_No_ConnectionId(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
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
	require.Error(t, err)
	require.Nil(t, resp)
}

// UpdateJobDestinationConnection
func Test_UpdateJobDestinationConnection_Update(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
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

	require.NoError(t, err)
	require.NotNil(t, resp)
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
	require.NoError(t, err)
	require.NotNil(t, resp)
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

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Msg.IsAvailable)
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

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Msg.IsAvailable)
}

type serviceMocks struct {
	Service                     *Service
	DbtxMock                    *nucleusdb.MockDBTX
	QuerierMock                 *db_queries.MockQuerier
	UserAccountServiceMock      *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceClientMock *mgmtv1alpha1connect.MockConnectionServiceClient
	TemporalWfManagerMock       *clientmanager.MockTemporalClientManagerClient
	SqlManagerMock              *sql_manager.MockSqlManagerClient
	SqlDbMock                   *sql_manager.MockSqlDatabase
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTemporalWfManager := clientmanager.NewMockTemporalClientManagerClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	service := New(config, nucleusdb.New(mockDbtx, mockQuerier), mockTemporalWfManager, mockConnectionService, mockUserAccountService, mockSqlManager)

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
func mockJob(accountId, userId, srcConnId string, cronSchedule pgtype.Text) db_queries.NeosyncApiJob {
	id, _ := nucleusdb.ToUuid(uuid.NewString())
	accountUuid, _ := nucleusdb.ToUuid(accountId)
	userUuid, _ := nucleusdb.ToUuid(userId)
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

// SetJobWorkflowOptions
func Test_SetJobWorkflowOptions(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})
	mockTx := new(nucleusdb.MockTx)

	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := nucleusdb.UUIDString(job.ID)

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

	userUuid, _ := nucleusdb.ToUuid(mockUserId)
	job := mockJob(mockAccountId, mockUserId, uuid.NewString(), pgtype.Text{})
	destConn := getConnectionMock(mockAccountId, "test-1")
	destConnAssociation := mockJobDestConnAssociation(job.ID, destConn.ID, &pg_models.JobDestinationOptions{})
	jobId := nucleusdb.UUIDString(job.ID)

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
	connId := nucleusdb.UUIDString(conn.ID)

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
	m.SqlDbMock.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sqlmanager_shared.ColumnInfo{
		"public.users":   {"id": &sqlmanager_shared.ColumnInfo{}, "name": &sqlmanager_shared.ColumnInfo{IsNullable: true}},
		"public.orders":  {"id": &sqlmanager_shared.ColumnInfo{}, "buyer_id": &sqlmanager_shared.ColumnInfo{IsNullable: true}},
		"circle.table_1": {"id": &sqlmanager_shared.ColumnInfo{}, "table2_id": &sqlmanager_shared.ColumnInfo{IsNullable: true}},
		"circle.table_2": {"id": &sqlmanager_shared.ColumnInfo{}, "table1_id": &sqlmanager_shared.ColumnInfo{}},
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
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_1", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_1", Column: "table2_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_2", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_2", Column: "table1_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
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
	connId := nucleusdb.UUIDString(conn.ID)

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
	m.SqlDbMock.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sqlmanager_shared.ColumnInfo{
		"public.users":   {"id": &sqlmanager_shared.ColumnInfo{}, "name": &sqlmanager_shared.ColumnInfo{}},
		"public.orders":  {"id": &sqlmanager_shared.ColumnInfo{}, "buyer_id": &sqlmanager_shared.ColumnInfo{}},
		"circle.table_1": {"id": &sqlmanager_shared.ColumnInfo{}, "table2_id": &sqlmanager_shared.ColumnInfo{}},
		"circle.table_2": {"id": &sqlmanager_shared.ColumnInfo{}, "table1_id": &sqlmanager_shared.ColumnInfo{}},
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
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "account", Table: "accounts", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_1", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_1", Column: "table2_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_2", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
				}},
				{Schema: "circle", Table: "table_2", Column: "table1_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					Config: &mgmtv1alpha1.TransformerConfig{},
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
