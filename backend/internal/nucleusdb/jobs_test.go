package nucleusdb

import (
	"context"
	"errors"
	"testing"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/zeebo/assert"
)

const (
	mockJobId  = "9e9ec62a-e0f8-4f0b-bf04-7de327ab6f9c"
	mockConnId = "8e34ce5b-cde8-4fc5-bd0a-7d1e50a4b14f"
)

// CreateJob
func Test_CreateJob(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	jobUuid, _ := ToUuid(mockJobId)
	accountUuid, _ := ToUuid(mockAccountId)
	connUuid, _ := ToUuid(mockConnId)
	ctx := context.Background()
	createJobParams := &db_queries.CreateJobParams{
		Name:      "job-name",
		AccountID: accountUuid,
	}
	destinations := []*CreateJobConnectionDestination{
		{ConnectionId: connUuid, Options: &pg_models.JobDestinationOptions{}},
	}
	destinationParams := []db_queries.CreateJobConnectionDestinationsParams{
		{JobID: jobUuid, ConnectionID: connUuid, Options: &pg_models.JobDestinationOptions{}},
	}

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("CreateJob", ctx, mockTx, *createJobParams).Return(db_queries.NeosyncApiJob{ID: jobUuid}, nil)
	querierMock.On("CreateJobConnectionDestinations", ctx, mockTx, destinationParams).Return(int64(1), nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.CreateJob(context.Background(), createJobParams, destinations)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, jobUuid, resp.ID)
}

func Test_CreateJob_Rollback(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	jobUuid, _ := ToUuid(mockJobId)
	accountUuid, _ := ToUuid(mockAccountId)
	connUuid, _ := ToUuid(mockConnId)
	ctx := context.Background()
	createJobParams := &db_queries.CreateJobParams{
		Name:      "job-name",
		AccountID: accountUuid,
	}
	destinations := []*CreateJobConnectionDestination{
		{ConnectionId: connUuid, Options: &pg_models.JobDestinationOptions{}},
	}
	destinationParams := []db_queries.CreateJobConnectionDestinationsParams{
		{JobID: jobUuid, ConnectionID: connUuid, Options: &pg_models.JobDestinationOptions{}},
	}

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("CreateJob", ctx, mockTx, *createJobParams).Return(db_queries.NeosyncApiJob{ID: jobUuid}, nil)
	querierMock.On("CreateJobConnectionDestinations", ctx, mockTx, destinationParams).Return(int64(1), errors.New("error"))
	mockTx.On("Rollback", ctx).Return(nil)

	resp, err := service.CreateJob(context.Background(), createJobParams, destinations)

	mockTx.AssertNotCalled(t, "Commit", ctx)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// SetSqlSourceSubsets
func Test_SetSqlSourceSubsets(t *testing.T) {
	dbtxMock := NewMockDBTX(t)
	querierMock := db_queries.NewMockQuerier(t)
	mockTx := new(MockTx)

	jobUuid, _ := ToUuid(mockJobId)
	userUuid, _ := ToUuid(mockUserId)
	connUuid, _ := ToUuid(mockConnId)
	ctx := context.Background()
	whereClause := "where"
	schemas := &mgmtv1alpha1.JobSourceSqlSubetSchemas{
		Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas_PostgresSubset{
			PostgresSubset: &mgmtv1alpha1.PostgresSourceSchemaSubset{
				PostgresSchemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
					{Schema: "schema", Tables: []*mgmtv1alpha1.PostgresSourceTableOption{{Table: "table-1", WhereClause: &whereClause}}},
				},
			},
		},
	}

	service := New(dbtxMock, querierMock)

	dbtxMock.On("Begin", ctx).Return(mockTx, nil)
	querierMock.On("GetJobById", ctx, mockTx, jobUuid).Return(db_queries.NeosyncApiJob{ID: jobUuid, ConnectionSourceID: connUuid, ConnectionOptions: &pg_models.JobSourceOptions{}}, nil)
	querierMock.On("UpdateJobSource", ctx, mockTx, db_queries.UpdateJobSourceParams{
		ID:                 jobUuid,
		ConnectionSourceID: connUuid,
		ConnectionOptions: &pg_models.JobSourceOptions{
			PostgresOptions: &pg_models.PostgresSourceOptions{
				Schemas: []*pg_models.PostgresSourceSchemaOption{
					{Schema: "schema", Tables: []*pg_models.PostgresSourceTableOption{{Table: "table-1", WhereClause: &whereClause}}},
				},
			},
		},
		UpdatedByID: userUuid,
	}).Return(db_queries.NeosyncApiJob{}, nil)
	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	err := service.SetSqlSourceSubsets(context.Background(), jobUuid, schemas, userUuid)

	assert.NoError(t, err)
}
