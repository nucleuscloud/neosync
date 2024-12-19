package integration_tests

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	tcredis "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/redis"
	pg_alltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/alltypes"
	pg_edgecases "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/edgecases"
	pg_humanresources "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/humanresources"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata"
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
	alltypesSchema := "alltypes"
	err := postgres.Source.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
	require.NoError(t, err)
	err = postgres.Target.CreateSchemas(ctx, []string{alltypesSchema})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	alltypesMappings := pg_alltypes.GetDefaultSyncJobMappings(alltypesSchema)

	job := createPostgresJob(t, ctx, jobclient, &createPostgresJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "all_types",
		JobMappings: alltypesMappings,
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
		{schema: alltypesSchema, table: "time_time", rowCount: 3},
		{schema: alltypesSchema, table: "json_data", rowCount: 12},
		{schema: alltypesSchema, table: "generated_table", rowCount: 4},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: all_types Table: %s", expected.table))
	}

	// tear down
	err = postgres.Source.DropSchemas(ctx, []string{alltypesSchema})
	require.NoError(t, err)
	err = postgres.Target.DropSchemas(ctx, []string{alltypesSchema})
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

func test_postgres_edgecases(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.UnauthdClients.Jobs
	schema := "CaPiTaL"
	err := postgres.Source.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"edgecases/create-tables.sql"}, schema)
	require.NoError(t, err)
	err = postgres.Target.CreateSchemas(ctx, []string{schema})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	edgecasesMappings := pg_edgecases.GetDefaultSyncJobMappings(schema)

	job := createPostgresJob(t, ctx, jobclient, &createPostgresJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: edgecasesMappings,
		JobOptions: &workflow_testdata.TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
	})

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: edgecases")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: edgecases")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "addresses", rowCount: 8},
		{schema: schema, table: "customers", rowCount: 10},
		{schema: schema, table: "orders", rowCount: 10},
		{schema: schema, table: "company", rowCount: 3},
		{schema: schema, table: "department", rowCount: 4},
		{schema: schema, table: "expense_report", rowCount: 2},
		{schema: schema, table: "transaction", rowCount: 3},
		{schema: schema, table: "BadName", rowCount: 5},
		{schema: schema, table: "Bad Name 123!@#", rowCount: 5},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: edgecases Table: %s", expected.table))
	}

	// tear down
	err = postgres.Source.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
	err = postgres.Target.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
}

func test_postgres_virtual_foreign_keys(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.UnauthdClients.Jobs
	schema := "vfk_hr"
	err := postgres.Source.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"humanresources/create-tables.sql"}, schema)
	require.NoError(t, err)
	// only create foreign key constraints in target to test that virtual foreign keys are correct
	err = postgres.Target.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"humanresources/create-tables.sql", "humanresources/create-constraints.sql"}, schema)
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	humanresourcesMappings := pg_humanresources.GetDefaultSyncJobMappings(schema)

	job := createPostgresJob(t, ctx, jobclient, &createPostgresJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: humanresourcesMappings,
		JobOptions: &workflow_testdata.TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      false,
		},
		VirtualForeignKeys: []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "vfk_hr",
				Table:   "countries",
				Columns: []string{"region_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "regions",
					Columns: []string{"region_id"},
				},
			},
			{
				Schema:  "vfk_hr",
				Table:   "departments",
				Columns: []string{"location_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "locations",
					Columns: []string{"location_id"},
				},
			},
			{
				Schema:  "vfk_hr",
				Table:   "dependents",
				Columns: []string{"employee_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "employees",
					Columns: []string{"employee_id"},
				},
			},
			{
				Schema:  "vfk_hr",
				Table:   "employees",
				Columns: []string{"manager_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "employees",
					Columns: []string{"employee_id"},
				},
			},
			{
				Schema:  "vfk_hr",
				Table:   "employees",
				Columns: []string{"department_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "departments",
					Columns: []string{"department_id"},
				},
			},
			{
				Schema:  "vfk_hr",
				Table:   "employees",
				Columns: []string{"job_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "jobs",
					Columns: []string{"job_id"},
				},
			},
			{
				Schema:  "vfk_hr",
				Table:   "locations",
				Columns: []string{"country_id"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "vfk_hr",
					Table:   "countries",
					Columns: []string{"country_id"},
				},
			},
		},
	})

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: edgecases")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: edgecases")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "regions", rowCount: 4},
		{schema: schema, table: "countries", rowCount: 25},
		{schema: schema, table: "locations", rowCount: 7},
		{schema: schema, table: "departments", rowCount: 11},
		{schema: schema, table: "jobs", rowCount: 19},
		{schema: schema, table: "employees", rowCount: 40},
		{schema: schema, table: "dependents", rowCount: 30},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: edgecases Table: %s", expected.table))
	}

	// tear down
	err = postgres.Source.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
	err = postgres.Target.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
}
