package integration_tests

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	tcredis "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/redis"
	pg_edgecases "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/edgecases"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata"
	pg_alltypes "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata/postgres/all-types"
	pg_primarykeytransformer "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata/postgres/primary-key-transformer"
	"github.com/stretchr/testify/require"
)

const (
	testdataFolder      string = "../../../internal/testutil/testdata/postgres"
	localTestdataFolder string = "./testdata/postgres"
)

type createPostgresJobConfig struct {
	AccountId          string
	SourceConn         *mgmtv1alpha1.Connection
	DestConn           *mgmtv1alpha1.Connection
	JobName            string
	JobMappings        []*mgmtv1alpha1.JobMapping
	VirtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint
	SubsetMap          map[string]string
	JobOptions         *workflow_testdata.TestJobOptions
}

func createPostgresJob(
	t *testing.T,
	ctx context.Context,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	config *createPostgresJobConfig,
) *mgmtv1alpha1.Job {
	schemas := []*mgmtv1alpha1.PostgresSourceSchemaOption{}
	subsetMap := map[string]*mgmtv1alpha1.PostgresSourceSchemaOption{}
	for table, where := range config.SubsetMap {
		schema, table := sqlmanager_shared.SplitTableKey(table)
		if _, exists := subsetMap[schema]; !exists {
			subsetMap[schema] = &mgmtv1alpha1.PostgresSourceSchemaOption{
				Schema: schema,
				Tables: []*mgmtv1alpha1.PostgresSourceTableOption{},
			}
		}
		w := where
		subsetMap[schema].Tables = append(subsetMap[schema].Tables, &mgmtv1alpha1.PostgresSourceTableOption{
			Table:       table,
			WhereClause: &w,
		})
	}

	for _, s := range subsetMap {
		schemas = append(schemas, s)
	}

	var subsetByForeignKeyConstraints bool
	destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
			PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{},
		},
	}
	if config.JobOptions != nil {
		if config.JobOptions.SubsetByForeignKeyConstraints {
			subsetByForeignKeyConstraints = true
		}
		destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
				PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
					InitTableSchema: config.JobOptions.InitSchema,
					TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
						TruncateBeforeInsert: config.JobOptions.Truncate,
					},
					SkipForeignKeyViolations: config.JobOptions.SkipForeignKeyViolations,
				},
			},
		}
	}

	job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: config.AccountId,
		JobName:   config.JobName,
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
					Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
						ConnectionId:                  config.SourceConn.Id,
						Schemas:                       schemas,
						SubsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
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
		Mappings:           config.JobMappings,
		VirtualForeignKeys: config.VirtualForeignKeys,
	}))
	require.NoError(t, err)

	return job.Msg.GetJob()
}

func test_postgres_types(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.UnauthdClients.Jobs
	capitalSchema := "CaPiTaL"
	alltypesSchema := "alltypes"
	schemas := []string{alltypesSchema, capitalSchema}
	err := postgres.Source.RunCreateStmtsInSchema(ctx, localTestdataFolder, []string{"all-types/create-tables.sql"}, alltypesSchema)
	require.NoError(t, err)
	err = postgres.Source.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"edgecases/create-tables.sql"}, capitalSchema)
	require.NoError(t, err)
	err = postgres.Target.CreateSchemas(ctx, schemas)
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	edgecasesMappings := pg_edgecases.GetDefaultSyncJobMappings(capitalSchema)
	alltypesMappings := pg_alltypes.GetDefaultSyncJobMappings(alltypesSchema)

	job := createPostgresJob(t, ctx, jobclient, &createPostgresJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "all_types",
		JobMappings: slices.Concat(edgecasesMappings, alltypesMappings),
		JobOptions: &workflow_testdata.TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
	})

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: all_types")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: all_types")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: alltypesSchema, table: "all_data_types", rowCount: 2},
		{schema: alltypesSchema, table: "array_types", rowCount: 1},
		{schema: alltypesSchema, table: "time_time", rowCount: 2},
		{schema: alltypesSchema, table: "json_data", rowCount: 12},
		{schema: capitalSchema, table: "BadName", rowCount: 5},
		{schema: capitalSchema, table: "Bad Name 123!@#", rowCount: 5},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: all_types Table: %s", expected.table))
	}

	// tear down
	err = postgres.Source.DropSchemas(ctx, schemas)
	require.NoError(t, err)
	err = postgres.Target.DropSchemas(ctx, schemas)
	require.NoError(t, err)
}

func test_postgres_primary_key_transformations(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	redis *tcredis.RedisTestContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.UnauthdClients.Jobs
	schema := "primary_$key"
	err := postgres.Source.RunCreateStmtsInSchema(ctx, localTestdataFolder, []string{"primary-key-transformer/create-tables.sql"}, schema)
	require.NoError(t, err)
	err = postgres.Target.CreateSchemas(ctx, []string{schema})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	defaultMappings := pg_primarykeytransformer.GetDefaultSyncJobMappings(schema)
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range defaultMappings {
		if jm.Column != "id" {
			updatedJobmappings = append(updatedJobmappings, jm)
		} else {
			updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
				Schema: jm.Schema,
				Table:  jm.Table,
				Column: jm.Column,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
							GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
								IncludeHyphens: gotypeutil.ToPtr(true),
							},
						},
					},
				},
			})
		}
	}

	job := createPostgresJob(t, ctx, jobclient, &createPostgresJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "primary_key_transformations",
		JobMappings: updatedJobmappings,
		JobOptions: &workflow_testdata.TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
	})

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, tcworkflow.WithRedis(redis.URL))
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: primary_key_transformations")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: all_types")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "store_notifications", rowCount: 20},
		{schema: schema, table: "stores", rowCount: 20},
		{schema: schema, table: "store_customers", rowCount: 20},
		{schema: schema, table: "referral_codes", rowCount: 20},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: primary_key_transformations Table: %s", expected.table))
	}

	keys, err := testworkflow.Redisclient.Keys(ctx, "*").Result()
	if err != nil {
		t.Fatal(err)
	}
	require.Emptyf(t, keys, "Redis keys should be empty")

	// tear down
	err = postgres.Source.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
	err = postgres.Target.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
}
