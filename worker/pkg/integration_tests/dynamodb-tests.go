package integration_tests

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyntypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	tcdynamodb "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/dynamodb"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func createDynamodbSyncJob(
	t *testing.T,
	ctx context.Context,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	config *createJobConfig,
	sourceTableName, destTableName string,
) *mgmtv1alpha1.Job {
	sourceTableOpts := []*mgmtv1alpha1.DynamoDBSourceTableOption{}
	for table, where := range config.SubsetMap {
		where := where
		sourceTableOpts = append(sourceTableOpts, &mgmtv1alpha1.DynamoDBSourceTableOption{
			Table:       table,
			WhereClause: &where,
		})
	}

	destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_DynamodbOptions{
			DynamodbOptions: &mgmtv1alpha1.DynamoDBDestinationConnectionOptions{
				TableMappings: []*mgmtv1alpha1.DynamoDBDestinationTableMapping{{SourceTable: sourceTableName, DestinationTable: destTableName}},
			},
		},
	}

	job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: config.AccountId,
		JobName:   config.JobName,
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Dynamodb{
					Dynamodb: &mgmtv1alpha1.DynamoDBSourceConnectionOptions{
						ConnectionId: config.SourceConn.Id,
						Tables:       sourceTableOpts,
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

func test_dynamodb_alltypes(
	t *testing.T,
	ctx context.Context,
	dynamo *tcdynamodb.DynamoDBTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	neosyncApi.MockTemporalForCreateJob("test-dynamodb-sync")
	tableName := "test-all-types"
	primaryKey := "id"

	err := createDynamodbTables(ctx, dynamo, tableName, primaryKey)
	if err != nil {
		t.Fatal(err)
	}

	testData := getAllTypesTestData()

	err = dynamo.Source.InsertDynamoDBRecords(ctx, tableName, testData)
	if err != nil {
		t.Fatal(err)
	}

	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "aws",
			Table:  tableName,
			Column: "id",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
		{
			Schema: "aws",
			Table:  tableName,
			Column: "a",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
	}

	job := createDynamodbSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "all_types",
		JobMappings: mappings,
	}, tableName, tableName)

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: all_types")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: all_types")

	// Verify data was synced
	out, err := dynamo.Target.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &tableName,
	})
	require.NoError(t, err)
	require.Equal(t, int32(4), out.Count)
	err = cleanupDynamodbTables(ctx, dynamo, tableName)
	require.NoError(t, err)
}

func test_dynamodb_subset(
	t *testing.T,
	ctx context.Context,
	dynamo *tcdynamodb.DynamoDBTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	neosyncApi.MockTemporalForCreateJob("test-dynamodb-sync")
	tableName := "test-subset"
	primaryKey := "id"

	err := createDynamodbTables(ctx, dynamo, tableName, primaryKey)
	if err != nil {
		t.Fatal(err)
	}

	testData := getAllTypesTestData()

	err = dynamo.Source.InsertDynamoDBRecords(ctx, tableName, testData)
	if err != nil {
		t.Fatal(err)
	}

	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "aws",
			Table:  tableName,
			Column: "id",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
		{
			Schema: "aws",
			Table:  tableName,
			Column: "a",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
	}
	subsetMap := map[string]string{
		tableName: "id = '1'",
	}

	job := createDynamodbSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     tableName,
		SubsetMap:   subsetMap,
		JobMappings: mappings,
	}, tableName, tableName)

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: subset")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: subset")

	// Verify data was synced
	out, err := dynamo.Target.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &tableName,
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), out.Count)
	err = cleanupDynamodbTables(ctx, dynamo, tableName)
	require.NoError(t, err)
}

func test_dynamodb_default_transformers(
	t *testing.T,
	ctx context.Context,
	dynamo *tcdynamodb.DynamoDBTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	neosyncApi.MockTemporalForCreateJob("test-dynamodb-sync")
	tableName := "test-default-transformers"
	primaryKey := "id"

	err := createDynamodbTables(ctx, dynamo, tableName, primaryKey)
	if err != nil {
		t.Fatal(err)
	}

	testData := getAllTypesTestData()
	err = dynamo.Source.InsertDynamoDBRecords(ctx, tableName, testData)
	if err != nil {
		t.Fatal(err)
	}

	mappings := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "aws",
			Table:  tableName,
			Column: "id",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
		{
			Schema: "aws",
			Table:  tableName,
			Column: "a",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
	}
	jobOpts := TestJobOptions{
		DefaultTransformers: &DefaultTransformers{
			Boolean: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
				},
			},
			Number: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
						TransformInt64Config: &mgmtv1alpha1.TransformInt64{RandomizationRangeMin: gotypeutil.ToPtr(int64(10)), RandomizationRangeMax: gotypeutil.ToPtr(int64(1000))},
					},
				},
			},
			String: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
						TransformStringConfig: &mgmtv1alpha1.TransformString{},
					},
				},
			},
			Byte: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
				},
			},
		},
	}

	job := createDynamodbSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     tableName,
		JobOptions:  &jobOpts,
		JobMappings: mappings,
	}, tableName, tableName)

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: default_transformers")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: default_transformers")

	// Verify data was synced
	out, err := dynamo.Target.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &tableName,
	})
	require.NoError(t, err)
	require.Equal(t, int32(4), out.Count)
	// tear down
	err = cleanupDynamodbTables(ctx, dynamo, tableName)
	require.NoError(t, err)
}

func cleanupDynamodbTables(ctx context.Context, dynamo *tcdynamodb.DynamoDBTestSyncContainer, tableName string) error {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error { return dynamo.Source.DestroyDynamoDbTable(errctx, tableName) })
	errgrp.Go(func() error { return dynamo.Target.DestroyDynamoDbTable(errctx, tableName) })
	return errgrp.Wait()
}

func createDynamodbTables(ctx context.Context, dynamo *tcdynamodb.DynamoDBTestSyncContainer, tableName string, primaryKey string) error {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error { return dynamo.Source.SetupDynamoDbTable(errctx, tableName, primaryKey) })
	errgrp.Go(func() error { return dynamo.Target.SetupDynamoDbTable(errctx, tableName, primaryKey) })
	return errgrp.Wait()
}

func getAllTypesTestData() []map[string]dyntypes.AttributeValue {
	return []map[string]dyntypes.AttributeValue{
		{
			"id": &dyntypes.AttributeValueMemberS{Value: "1"},
			"a":  &dyntypes.AttributeValueMemberBOOL{Value: true},
			"NestedMap": &dyntypes.AttributeValueMemberM{
				Value: map[string]dyntypes.AttributeValue{
					"Level1": &dyntypes.AttributeValueMemberM{
						Value: map[string]dyntypes.AttributeValue{
							"Level2": &dyntypes.AttributeValueMemberM{
								Value: map[string]dyntypes.AttributeValue{
									"Attribute1": &dyntypes.AttributeValueMemberS{Value: "Value1"},
									"NumberSet":  &dyntypes.AttributeValueMemberNS{Value: []string{"1", "2", "3"}},
									"BinaryData": &dyntypes.AttributeValueMemberB{Value: []byte("U29tZUJpbmFyeURhdGE=")},
									"Level3": &dyntypes.AttributeValueMemberM{
										Value: map[string]dyntypes.AttributeValue{
											"Attribute2": &dyntypes.AttributeValueMemberS{Value: "Value2"},
											"StringSet":  &dyntypes.AttributeValueMemberSS{Value: []string{"Item1", "Item2", "Item3"}},
											"BinarySet": &dyntypes.AttributeValueMemberBS{
												Value: [][]byte{
													[]byte("U29tZUJpbmFyeQ=="),
													[]byte("QW5vdGhlckJpbmFyeQ=="),
												},
											},
											"Level4": &dyntypes.AttributeValueMemberM{
												Value: map[string]dyntypes.AttributeValue{
													"Attribute3":     &dyntypes.AttributeValueMemberS{Value: "Value3"},
													"Boolean":        &dyntypes.AttributeValueMemberBOOL{Value: true},
													"MoreBinaryData": &dyntypes.AttributeValueMemberB{Value: []byte("TW9yZUJpbmFyeURhdGE=")},
													"MoreBinarySet": &dyntypes.AttributeValueMemberBS{
														Value: [][]byte{
															[]byte("TW9yZUJpbmFyeQ=="),
															[]byte("QW5vdGhlck1vcmVCaW5hcnk="),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"id": &dyntypes.AttributeValueMemberS{Value: "2"},
			"a":  &dyntypes.AttributeValueMemberBOOL{Value: false},
		},
		{
			"id":   &dyntypes.AttributeValueMemberS{Value: "3"},
			"name": &dyntypes.AttributeValueMemberS{Value: "test3"},
		},
		{
			"id":   &dyntypes.AttributeValueMemberS{Value: "4"},
			"name": &dyntypes.AttributeValueMemberS{Value: "test4"},
		},
	}
}
