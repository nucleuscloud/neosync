package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	tcmongodb "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mongodb"
)

func createMongodbSyncJob(
	t *testing.T,
	ctx context.Context,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	config *createJobConfig,
) *mgmtv1alpha1.Job {
	destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_MongodbOptions{
			MongodbOptions: &mgmtv1alpha1.MongoDBDestinationConnectionOptions{},
		},
	}

	job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: config.AccountId,
		JobName:   config.JobName,
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Mongodb{
					Mongodb: &mgmtv1alpha1.MongoDBSourceConnectionOptions{
						ConnectionId: config.SourceConn.Id,
					},
				},
			},
		},
		Destinations: []*mgmtv1alpha1.CreateJobDestination{
			{
				ConnectionId: config.DestConn.Id,
				Options:      destinationOptions,
			},
		},
		Mappings: config.JobMappings,
	}))
	require.NoError(t, err)

	return job.Msg.GetJob()
}

func test_mongodb_alltypes(
	t *testing.T,
	ctx context.Context,
	mongo *tcmongodb.MongoDBTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	neosyncApi.MockTemporalForCreateJob("test-mongodb-sync")
	dbName := "data"
	collectionName := "test-sync"
	docs := getMongodbAllTypesTestData()

	count, err := mongo.Source.InsertMongoDbRecords(ctx, dbName, collectionName, docs)
	require.NoError(t, err)
	require.Greater(t, count, 0)

	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "data",
			Table:  collectionName,
			Column: "string",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
		{
			Schema: "data",
			Table:  collectionName,
			Column: "bool",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
	}

	job := createMongodbSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mongo_all_types",
		JobMappings: mappings,
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mongo_all_types")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mongo_all_types")

	col := mongo.Target.Client.Database(dbName).Collection(collectionName)
	cursor, err := col.Find(ctx, bson.D{})
	require.NoError(t, err)
	var results []bson.M
	for cursor.Next(ctx) {
		var doc bson.M
		err = cursor.Decode(&doc)
		require.NoError(t, err)
		results = append(results, doc)
	}
	cursor.Close(ctx)
	require.Equal(t, 1, len(results), fmt.Sprintf("Test: mongo_all_types collection: %s", collectionName))
	err = cleanupMongodb(ctx, mongo, dbName, collectionName)
	require.NoError(t, err)
}

func test_mongodb_transform(
	t *testing.T,
	ctx context.Context,
	mongo *tcmongodb.MongoDBTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	neosyncApi.MockTemporalForCreateJob("test-mongodb-sync")
	dbName := "data"
	collectionName := "test-sync-transform"
	docs := getMongodbAllTypesTestData()

	count, err := mongo.Source.InsertMongoDbRecords(ctx, dbName, collectionName, docs)
	require.NoError(t, err)
	require.Greater(t, count, 0)

	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "data",
			Table:  collectionName,
			Column: "string",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
						TransformStringConfig: &mgmtv1alpha1.TransformString{
							PreserveLength: gotypeutil.ToPtr(true),
						},
					},
				},
			},
		},
		{
			Schema: "data",
			Table:  collectionName,
			Column: "embedded_document.name",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
						GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
					},
				},
			},
		},
		{
			Schema: "data",
			Table:  collectionName,
			Column: "decimal128",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
						TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
							RandomizationRangeMin: gotypeutil.ToPtr(float64(0)),
							RandomizationRangeMax: gotypeutil.ToPtr(float64(300)),
						},
					},
				},
			},
		},
		{
			Schema: "data",
			Table:  collectionName,
			Column: "int64",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
						TransformInt64Config: &mgmtv1alpha1.TransformInt64{
							RandomizationRangeMin: gotypeutil.ToPtr(int64(0)),
							RandomizationRangeMax: gotypeutil.ToPtr(int64(300)),
						},
					},
				},
			},
		},
		{
			Schema: "data",
			Table:  collectionName,
			Column: "timestamp",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
						GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
					},
				},
			},
		},
	}

	job := createMongodbSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mongo_transform",
		JobMappings: mappings,
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mongo_transform")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mongo_transform")

	col := mongo.Target.Client.Database(dbName).Collection(collectionName)
	cursor, err := col.Find(ctx, bson.D{})
	require.NoError(t, err)
	var results []bson.M
	for cursor.Next(ctx) {
		var doc bson.M
		err = cursor.Decode(&doc)
		require.NoError(t, err)
		results = append(results, doc)
	}
	cursor.Close(ctx)
	require.Equal(t, 1, len(results), fmt.Sprintf("Test: mongo_transform collection: %s", collectionName))
	err = cleanupMongodb(ctx, mongo, dbName, collectionName)
	require.NoError(t, err)
}

func cleanupMongodb(ctx context.Context, mongo *tcmongodb.MongoDBTestSyncContainer, dbName, collectionName string) error {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error { return mongo.Source.DropMongoDbCollection(errctx, dbName, collectionName) })
	errgrp.Go(func() error { return mongo.Target.DropMongoDbCollection(errctx, dbName, collectionName) })
	err := errgrp.Wait()
	return err
}

func getMongodbAllTypesTestData() []any {
	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "string", Value: "Hello, MongoDB!"},
		{Key: "bool", Value: true},
		{Key: "int32", Value: int32(42)},
		{Key: "int64", Value: int64(92233720)},
		{Key: "double", Value: 3.14159},
		{Key: "decimal128", Value: primitive.NewDecimal128(3, 14159)},
		{Key: "date", Value: primitive.NewDateTimeFromTime(time.Now())},
		{Key: "timestamp", Value: primitive.Timestamp{T: 1645553494, I: 1}},
		{Key: "null", Value: primitive.Null{}},
		{Key: "regex", Value: primitive.Regex{Pattern: "^test", Options: "i"}},
		{Key: "array", Value: bson.A{"apple", "banana", "cherry"}},
		{Key: "embedded_document", Value: bson.D{
			{Key: "name", Value: "John Doe"},
			{Key: "age", Value: 30},
		}},
		{Key: "binary", Value: primitive.Binary{Subtype: 0x80, Data: []byte("binary data")}},
		{Key: "undefined", Value: primitive.Undefined{}},
		{Key: "object_id", Value: primitive.NewObjectID()},
		{Key: "min_key", Value: primitive.MinKey{}},
		{Key: "max_key", Value: primitive.MaxKey{}},
	}
	return []any{doc}
}
