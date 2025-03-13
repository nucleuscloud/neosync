package integrationtest

import (
	"context"
	"database/sql"
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
	testutil_testdata "github.com/nucleuscloud/neosync/internal/testutil/testdata"
	pg_alltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/alltypes"
	pg_complex "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/complex"
	pg_edgecases "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/edgecases"
	pg_foreignkey_violations "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/foreignkey-violations"
	pg_humanresources "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/humanresources"
	pg_subsetting "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/subsetting"
	pg_transformers "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/transformers"
	pg_uuids "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/uuids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const (
	testdataFolder string = "../../../testutil/testdata/postgres"
)

type createJobConfig struct {
	AccountId          string
	SourceConn         *mgmtv1alpha1.Connection
	DestConn           *mgmtv1alpha1.Connection
	JobName            string
	JobMappings        []*mgmtv1alpha1.JobMapping
	VirtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint
	SubsetMap          map[string]string
	JobOptions         *TestJobOptions
}

func createPostgresSyncJob(
	t *testing.T,
	ctx context.Context,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	config *createJobConfig,
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
		var batchSize, maxInFlight *uint32
		if config.JobOptions.BatchSize != nil {
			batchSize = config.JobOptions.BatchSize
		}
		if config.JobOptions.MaxInFlight != nil {
			maxInFlight = config.JobOptions.MaxInFlight
		}
		destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
				PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
					InitTableSchema: config.JobOptions.InitSchema,
					TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
						TruncateBeforeInsert: config.JobOptions.Truncate,
					},
					SkipForeignKeyViolations: config.JobOptions.SkipForeignKeyViolations,
					Batch: &mgmtv1alpha1.BatchConfig{
						Count: batchSize,
					},
					MaxInFlight: maxInFlight,
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
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	alltypesSchema := "alltypes"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
	})
	errgrp.Go(func() error { return postgres.Target.CreateSchemas(errctx, []string{alltypesSchema}) })
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	alltypesMappings := pg_alltypes.GetDefaultSyncJobMappings(alltypesSchema)

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "all_types",
		JobMappings: alltypesMappings,
		JobOptions: &TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
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

	source, err := sql.Open("postgres", postgres.Source.URL)
	require.NoError(t, err)
	defer source.Close()

	target, err := sql.Open("postgres", postgres.Target.URL)
	require.NoError(t, err)
	defer target.Close()

	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, alltypesSchema, "all_data_types", "postgres", []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, alltypesSchema, "json_data", "postgres", []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, alltypesSchema, "array_types", "postgres", []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, alltypesSchema, "generated_table", "postgres", []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, alltypesSchema, "time_time", "postgres", []string{"id"})

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{alltypesSchema})
	require.NoError(t, err)
}

func test_postgres_primary_key_transformations(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	redis *tcredis.RedisTestContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "primary_$key_sdef"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"uuids/create-tables.sql", "humanresources/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Target.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"humanresources/create-tables.sql", "humanresources/create-constraints.sql"}, schema)
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	defaultMappings := pg_uuids.GetDefaultSyncJobMappings(schema)
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

	transformHumanresourcesMappings := pg_humanresources.GetDefaultSyncJobMappings(schema)
	for _, m := range transformHumanresourcesMappings {
		if m.Table == "countries" && m.Column == "country_id" {
			m.Transformer = &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
						TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `if (value == 'US') { return 'SU'; } return value;`},
					},
				},
			}
		}
	}

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "primary_key_transformations",
		JobMappings: slices.Concat(updatedJobmappings, transformHumanresourcesMappings),
		JobOptions: &TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
		VirtualForeignKeys: pg_humanresources.GetVirtualForeignKeys(schema),
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithRedis(redis.URL))
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
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: primary_key_transformations Table: %s", expected.table))
	}

	keys, err := testworkflow.Redisclient.Keys(ctx, "*").Result()
	if err != nil {
		t.Fatal(err)
	}
	require.Emptyf(t, keys, "Redis keys should be empty")

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_edgecases(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "CaPiTaL"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"edgecases/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Target.CreateSchemas(errctx, []string{schema})
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	edgecasesMappings := pg_edgecases.GetDefaultSyncJobMappings(schema)

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: edgecasesMappings,
		JobOptions: &TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
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
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_virtual_foreign_keys(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "vfk_hr"
	subsetSchema := "vfk_hr_subset"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"humanresources/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"humanresources/create-tables.sql"}, subsetSchema)
	})
	// only create foreign key constraints in target to test that virtual foreign keys are correct
	errgrp.Go(func() error {
		return postgres.Target.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"humanresources/create-tables.sql", "humanresources/create-constraints.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Target.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"humanresources/create-tables.sql", "humanresources/create-constraints.sql"}, subsetSchema)
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	humanresourcesMappings := pg_humanresources.GetDefaultSyncJobMappings(schema)
	subsetHumanresourcesMappings := pg_humanresources.GetDefaultSyncJobMappings(subsetSchema)
	virtualForeignKeys := pg_humanresources.GetVirtualForeignKeys(schema)
	subsetVirtualForeignKeys := pg_humanresources.GetVirtualForeignKeys(subsetSchema)

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: slices.Concat(humanresourcesMappings, subsetHumanresourcesMappings),
		SubsetMap: map[string]string{
			fmt.Sprintf("%s.employees", subsetSchema): "first_name = 'Alexander'",
		},
		JobOptions: &TestJobOptions{
			Truncate:                      true,
			TruncateCascade:               true,
			InitSchema:                    false,
			SubsetByForeignKeyConstraints: true,
		},
		VirtualForeignKeys: slices.Concat(virtualForeignKeys, subsetVirtualForeignKeys),
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: virtual-foreign-keys")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: virtual-foreign-keys")

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
		{schema: subsetSchema, table: "regions", rowCount: 4},
		{schema: subsetSchema, table: "countries", rowCount: 25},
		{schema: subsetSchema, table: "locations", rowCount: 7},
		{schema: subsetSchema, table: "departments", rowCount: 11},
		{schema: subsetSchema, table: "dependents", rowCount: 2},
		{schema: subsetSchema, table: "employees", rowCount: 2},
		{schema: subsetSchema, table: "jobs", rowCount: 19},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: virtual-foreign-keys Table: %s", expected.table))
	}

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_javascript_transformers(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	transformersSchema := "transformers"
	generatorsSchema := "generators"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"transformers/create-tables.sql"}, transformersSchema)
	})
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"transformers/create-tables.sql"}, generatorsSchema)
	})
	errgrp.Go(func() error {
		return postgres.Target.CreateSchemas(errctx, []string{transformersSchema, generatorsSchema})
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	transformersMappings := getJsTransformerJobmappings(pg_transformers.GetDefaultSyncJobMappings(transformersSchema))
	generatorsMappings := getJsGeneratorJobmappings(pg_transformers.GetDefaultSyncJobMappings(generatorsSchema))

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "javascript-transformers",
		JobMappings: slices.Concat(transformersMappings, generatorsMappings),
		JobOptions: &TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: javascript-transformers")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: javascript-transformers")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: transformersSchema, table: "transformers", rowCount: 13},
		{schema: generatorsSchema, table: "transformers", rowCount: 13},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: javascript-transformers Table: %s", expected.table))
	}

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{transformersSchema, generatorsSchema})
	require.NoError(t, err)
}

func getJsGeneratorJobmappings(jobmappings []*mgmtv1alpha1.JobMapping) []*mgmtv1alpha1.JobMapping {
	colTransformerMap := map[string]*mgmtv1alpha1.JobMappingTransformer{
		"e164_phone_number":   getJavascriptTransformerConfig("return neosync.generateInternationalPhoneNumber({ min: 9, max: 15});"),
		"email":               getJavascriptTransformerConfig("return neosync.generateEmail({ maxLength: 255});"),
		"str":                 getJavascriptTransformerConfig("return neosync.generateRandomString({ min: 1, max: 50});"),
		"measurement":         getJavascriptTransformerConfig("return neosync.generateFloat64({ min: 3.14, max: 300.10});"),
		"int64":               getJavascriptTransformerConfig("return neosync.generateInt64({ min: 1, max: 50});"),
		"int64_phone_number":  getJavascriptTransformerConfig("return neosync.generateInt64PhoneNumber({});"),
		"string_phone_number": getJavascriptTransformerConfig("return neosync.generateStringPhoneNumber({ min: 1, max: 15});"),
		"first_name":          getJavascriptTransformerConfig("return neosync.generateFirstName({ maxLength: 25});"),
		"last_name":           getJavascriptTransformerConfig("return neosync.generateLastName({ maxLength: 25});"),
		"full_name":           getJavascriptTransformerConfig("return neosync.generateFullName({ maxLength: 25});"),
		"character_scramble":  getJavascriptTransformerConfig("return neosync.generateCity({ maxLength: 100});"),
		"bool":                getJavascriptTransformerConfig("return neosync.generateBool({});"),
		"card_number":         getJavascriptTransformerConfig("return neosync.generateCardNumber({ validLuhn: true });"),
		"categorical":         getJavascriptTransformerConfig("return neosync.generateCategorical({ categories: 'dog,cat,horse'});"),
		"city":                getJavascriptTransformerConfig("return neosync.generateCity({ maxLength: 100 });"),
		"full_address":        getJavascriptTransformerConfig("return neosync.generateFullAddress({ maxLength: 100 });"),
		"gender":              getJavascriptTransformerConfig("return neosync.generateGender({});"),
		"international_phone": getJavascriptTransformerConfig("return neosync.generateInternationalPhoneNumber({ min: 9, max: 14});"),
		"sha256":              getJavascriptTransformerConfig("return neosync.generateSHA256Hash({});"),
		"ssn":                 getJavascriptTransformerConfig("return neosync.generateSSN({});"),
		"state":               getJavascriptTransformerConfig("return neosync.generateState({});"),
		"street_address":      getJavascriptTransformerConfig("return neosync.generateStreetAddress({ maxLength: 100 });"),
		"unix_time":           getJavascriptTransformerConfig("return neosync.generateUnixTimestamp({});"),
		"username":            getJavascriptTransformerConfig("return neosync.generateUsername({ maxLength: 100 });"),
		"utc_timestamp":       getJavascriptTransformerConfig("return neosync.generateUTCTimestamp({});"),
		"uuid":                getJavascriptTransformerConfig("return neosync.generateUUID({});"),
		"zipcode":             getJavascriptTransformerConfig("return neosync.generateZipcode({});"),
	}
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		if _, ok := colTransformerMap[jm.Column]; !ok {
			updatedJobmappings = append(updatedJobmappings, jm)
		} else {
			updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
				Schema:      jm.Schema,
				Table:       jm.Table,
				Column:      jm.Column,
				Transformer: colTransformerMap[jm.Column],
			})
		}
	}
	return updatedJobmappings
}

func getJsTransformerJobmappings(jobmappings []*mgmtv1alpha1.JobMapping) []*mgmtv1alpha1.JobMapping {
	colTransformerMap := map[string]*mgmtv1alpha1.JobMappingTransformer{
		"e164_phone_number":   getJavascriptTransformerConfig("return neosync.transformE164PhoneNumber(value, { preserveLength: true, maxLength: 20});"),
		"email":               getJavascriptTransformerConfig("return neosync.transformEmail(value, { preserveLength: true, maxLength: 255});"),
		"str":                 getJavascriptTransformerConfig("return neosync.transformString(value, { preserveLength: true, maxLength: 30});"),
		"measurement":         getJavascriptTransformerConfig("return neosync.transformFloat64(value, { randomizationRangeMin: 3.14, randomizationRangeMax: 300.10});"),
		"int64":               getJavascriptTransformerConfig("return neosync.transformInt64(value, { randomizationRangeMin: 1, randomizationRangeMax: 300});"),
		"int64_phone_number":  getJavascriptTransformerConfig("return neosync.transformInt64PhoneNumber(value, { preserveLength: true});"),
		"string_phone_number": getJavascriptTransformerConfig("return neosync.transformStringPhoneNumber(value, { preserveLength: true, maxLength: 200});"),
		"first_name":          getJavascriptTransformerConfig("return neosync.transformFirstName(value, { preserveLength: true, maxLength: 25});"),
		"last_name":           getJavascriptTransformerConfig("return neosync.transformLastName(value, { preserveLength: true, maxLength: 25});"),
		"full_name":           getJavascriptTransformerConfig("return neosync.transformFullName(value, { preserveLength: true, maxLength: 25});"),
		"character_scramble":  getJavascriptTransformerConfig("return neosync.transformCharacterScramble(value, { preserveLength: false, maxLength: 100});"),
	}
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
			Schema:      jm.Schema,
			Table:       jm.Table,
			Column:      jm.Column,
			Transformer: colTransformerMap[jm.Column],
		})
	}
	return updatedJobmappings
}

func getJavascriptTransformerConfig(code string) *mgmtv1alpha1.JobMappingTransformer {
	return &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: code},
			},
		},
	}
}

func test_postgres_skip_foreign_keys_violations(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "fk_violations"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"foreignkey-violations/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Target.CreateSchemas(errctx, []string{schema})
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	jobmappings := pg_foreignkey_violations.GetDefaultSyncJobMappings(schema)

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: jobmappings,
		JobOptions: &TestJobOptions{
			Truncate:                 true,
			TruncateCascade:          true,
			InitSchema:               true,
			SkipForeignKeyViolations: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: skip-foreign-keys-violations")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: skip-foreign-keys-violations")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "countries", rowCount: 24},
		{schema: schema, table: "dependents", rowCount: 28},
		{schema: schema, table: "employees", rowCount: 40},
		{schema: schema, table: "locations", rowCount: 4},
		{schema: schema, table: "departments", rowCount: 4},
		{schema: schema, table: "jobs", rowCount: 19},
		{schema: schema, table: "regions", rowCount: 4},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		assert.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: skip-foreign-keys-violations Table: %s", expected.table))
	}

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_foreign_keys_violations_error(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "fk_violations_error"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"foreignkey-violations/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Target.CreateSchemas(errctx, []string{schema})
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	jobmappings := pg_foreignkey_violations.GetDefaultSyncJobMappings(schema)

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: jobmappings,
		JobOptions: &TestJobOptions{
			Truncate:                 true,
			TruncateCascade:          true,
			InitSchema:               true,
			SkipForeignKeyViolations: false,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: foreign-keys-violations-error")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.Error(t, err, "Received Temporal Workflow Error: foreign-keys-violations-error")

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_subsetting(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "subsetting"
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return postgres.Source.RunCreateStmtsInSchema(errctx, testdataFolder, []string{"subsetting/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return postgres.Target.CreateSchemas(errctx, []string{schema})
	})
	err := errgrp.Wait()
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	jobmappings := pg_subsetting.GetDefaultSyncJobMappings(schema)

	subsetMappings := map[string]string{
		"subsetting.users":     "user_id in (1,2,5,6,7,8)",
		"subsetting.test_2_x":  "created > '2023-06-03'",
		"subsetting.test_2_b":  "created > '2023-06-03'",
		"subsetting.addresses": "id in (1,5)",
		"subsetting.division":  "id in (3,5)",
		"subsetting.bosses":    "id in (3,5)",
	}

	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     schema,
		JobMappings: jobmappings,
		SubsetMap:   subsetMappings,
		JobOptions: &TestJobOptions{
			Truncate:                      true,
			TruncateCascade:               true,
			InitSchema:                    true,
			SubsetByForeignKeyConstraints: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithMaxIterations(2), WithPageLimit(3))
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: skip-foreign-keys-violations")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: skip-foreign-keys-violations")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "attachments", rowCount: 6},
		{schema: schema, table: "comments", rowCount: 12},
		{schema: schema, table: "initiatives", rowCount: 6},
		{schema: schema, table: "skills", rowCount: 10},
		{schema: schema, table: "tasks", rowCount: 6},
		{schema: schema, table: "user_skills", rowCount: 6},
		{schema: schema, table: "users", rowCount: 6},
		{schema: schema, table: "test_2_x", rowCount: 3},
		{schema: schema, table: "test_2_b", rowCount: 3},
		{schema: schema, table: "test_2_a", rowCount: 4},
		{schema: schema, table: "test_2_c", rowCount: 2},
		{schema: schema, table: "test_2_d", rowCount: 2},
		{schema: schema, table: "test_2_e", rowCount: 2},
		{schema: schema, table: "orders", rowCount: 2},
		{schema: schema, table: "addresses", rowCount: 2},
		{schema: schema, table: "customers", rowCount: 2},
		{schema: schema, table: "payments", rowCount: 1},
		{schema: schema, table: "division", rowCount: 2},
		{schema: schema, table: "employees", rowCount: 2},
		{schema: schema, table: "projects", rowCount: 2},
		{schema: schema, table: "bosses", rowCount: 2},
		{schema: schema, table: "minions", rowCount: 2},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		assert.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: skip-foreign-keys-violations Table: %s", expected.table))
	}

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_generate_workflow(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "generate"
	err := postgres.Target.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"alltypes/create-tables.sql"}, schema)
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	table := "all_data_types"
	mappings := []*mgmtv1alpha1.JobMapping{
		{Schema: schema, Table: table, Column: "integer_col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{}},
		}},
		{Schema: schema, Table: table, Column: "text_col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{},
			},
		}},
		{Schema: schema, Table: table, Column: "uuid_col", Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{},
			},
		}},
	}

	job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: accountId,
		JobName:   schema,
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Generate{
					Generate: &mgmtv1alpha1.GenerateSourceOptions{
						FkSourceConnectionId: &destConn.Id,
						Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
							{Schema: schema, Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
								{Table: table, RowCount: 10},
							}},
						},
					},
				},
			},
		},
		Destinations: []*mgmtv1alpha1.CreateJobDestination{
			{
				ConnectionId: destConn.Id,
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							InitTableSchema: false,
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: true,
							},
							SkipForeignKeyViolations: false,
						},
					},
				},
			},
		},
		Mappings:           mappings,
		VirtualForeignKeys: nil,
	}))
	require.NoError(t, err)

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.Msg.GetJob().GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: generate")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: generate")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: table, rowCount: 10},
	}

	for _, expected := range expectedResults {
		rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: generate Table: %s", expected.table))
	}

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func cleanupPostgresSchemas(ctx context.Context, postgres *tcpostgres.PostgresTestSyncContainer, schemas []string) error {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error { return postgres.Source.DropSchemas(errctx, schemas) })
	errgrp.Go(func() error { return postgres.Target.DropSchemas(errctx, schemas) })
	return errgrp.Wait()
}

func test_postgres_small_batch_size(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "small_batch"
	err := postgres.Source.RunCreateStmtsInSchema(ctx, testdataFolder, []string{"uuids/create-tables.sql", "humanresources/create-tables.sql", "humanresources/create-constraints.sql"}, schema)
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

	defaultMappings := pg_uuids.GetDefaultSyncJobMappings(schema)
	transformHumanresourcesMappings := pg_humanresources.GetDefaultSyncJobMappings(schema)

	limit := uint32(1)
	job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "tablesync_pages",
		JobMappings: slices.Concat(defaultMappings, transformHumanresourcesMappings),
		JobOptions: &TestJobOptions{
			Truncate:        true,
			TruncateCascade: true,
			InitSchema:      true,
			BatchSize:       &limit,
			MaxInFlight:     &limit,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithPageLimit(5), WithMaxIterations(2))
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: tablesync_pages")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: tablesync_pages")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "store_notifications", rowCount: 20},
		{schema: schema, table: "stores", rowCount: 20},
		{schema: schema, table: "store_customers", rowCount: 20},
		{schema: schema, table: "referral_codes", rowCount: 20},
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
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: tablesync_pages Table: %s", expected.table))
	}

	source, err := sql.Open("postgres", postgres.Source.URL)
	require.NoError(t, err)
	defer source.Close()

	target, err := sql.Open("postgres", postgres.Target.URL)
	require.NoError(t, err)
	defer target.Close()

	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "employees", sqlmanager_shared.PostgresDriver, []string{"employee_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "jobs", sqlmanager_shared.PostgresDriver, []string{"job_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "departments", sqlmanager_shared.PostgresDriver, []string{"department_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "dependents", sqlmanager_shared.PostgresDriver, []string{"dependent_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "countries", sqlmanager_shared.PostgresDriver, []string{"country_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "locations", sqlmanager_shared.PostgresDriver, []string{"location_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "regions", sqlmanager_shared.PostgresDriver, []string{"region_id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "referral_codes", sqlmanager_shared.PostgresDriver, []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "store_customers", sqlmanager_shared.PostgresDriver, []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "store_notifications", sqlmanager_shared.PostgresDriver, []string{"id"})
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, schema, "stores", sqlmanager_shared.PostgresDriver, []string{"id"})

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{schema})
	require.NoError(t, err)
}

func test_postgres_complex(
	t *testing.T,
	ctx context.Context,
	postgres *tcpostgres.PostgresTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	folder := testdataFolder + "/complex"
	err := postgres.Source.RunSqlFiles(ctx, &folder, []string{"create-tables.sql", "inserts.sql"})
	require.NoError(t, err)

	jobmappings := pg_complex.GetDefaultSyncJobMappings()

	t.Run("sync", func(t *testing.T) {
		neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

		job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
			AccountId:   accountId,
			SourceConn:  sourceConn,
			DestConn:    destConn,
			JobName:     "space-mission",
			JobMappings: jobmappings,
			JobOptions: &TestJobOptions{
				Truncate:                      true,
				TruncateCascade:               true,
				InitSchema:                    true,
				SubsetByForeignKeyConstraints: true,
			},
		})

		testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithMaxIterations(10), WithPageLimit(100))
		testworkflow.RequireActivitiesCompletedSuccessfully(t)
		testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
		require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: space-mission")
		err = testworkflow.TestEnv.GetWorkflowError()
		require.NoError(t, err, "Received Temporal Workflow Error: space-mission")

		expectedResults := []struct {
			schema    string
			table     string
			rowCount  int
			idColumns []string
		}{
			{schema: "space_mission", table: "astronauts", rowCount: 10, idColumns: []string{"astronaut_id"}},
			{schema: "space_mission", table: "missions", rowCount: 10, idColumns: []string{"mission_id"}},
			{schema: "space_mission", table: "objectives", rowCount: 10, idColumns: []string{"objective_id"}},
			{schema: "space_mission", table: "capabilities", rowCount: 10, idColumns: []string{"capability_id"}},
			{schema: "space_mission", table: "astronaut_capabilities", rowCount: 10, idColumns: []string{"astronaut_capability_id"}},
			{schema: "space_mission", table: "transmissions", rowCount: 20, idColumns: []string{"transmission_id"}},
			{schema: "space_mission", table: "payloads", rowCount: 10, idColumns: []string{"payload_id"}},
			{schema: "space_mission", table: "crew_assignments", rowCount: 10, idColumns: []string{"crew_assignment_id"}},
			{schema: "space_mission", table: "mission_logs", rowCount: 6, idColumns: []string{"log_id"}},
			{schema: "space_mission", table: "crews", rowCount: 5, idColumns: []string{"crew_id"}},
			{schema: "space_mission", table: "crew_missions", rowCount: 4, idColumns: []string{"crew_mission_id"}},
			{schema: "space_mission", table: "supplies", rowCount: 10, idColumns: []string{"supply_id"}},
			{schema: "space_mission", table: "supply_items", rowCount: 5, idColumns: []string{"supply_item_id"}},
			{schema: "space_mission", table: "spacecraft_class", rowCount: 3, idColumns: []string{"class_id"}},
			{schema: "space_mission", table: "spacecraft", rowCount: 3, idColumns: []string{"spacecraft_id"}},
			{schema: "space_mission", table: "spacecraft_module", rowCount: 3, idColumns: []string{"module_id"}},
			{schema: "space_mission", table: "module_component", rowCount: 3, idColumns: []string{"component_id"}},
			{schema: "space_mission", table: "equipment", rowCount: 3, idColumns: []string{"equipment_id"}},
			{schema: "space_mission", table: "mission_equipment", rowCount: 3, idColumns: []string{"mission_id", "equipment_id"}},
			{schema: "space_mission", table: "equipment_maintenance", rowCount: 3, idColumns: []string{"maintenance_id"}},
			{schema: "space_mission", table: "training_courses", rowCount: 3, idColumns: []string{"course_id"}},
			{schema: "space_mission", table: "course_prerequisites", rowCount: 3, idColumns: []string{"prerequisite_id"}},
			{schema: "space_mission", table: "certifications", rowCount: 3, idColumns: []string{"certification_id"}},
			{schema: "space_mission", table: "astronaut_certifications", rowCount: 4, idColumns: []string{"astronaut_id", "certification_id"}},
			{schema: "space_mission", table: "certification_requirements", rowCount: 3, idColumns: []string{"requirement_id"}},
			{schema: "space_mission", table: "mission_logs_extended", rowCount: 3, idColumns: []string{"log_id"}},
			{schema: "space_mission", table: "communication_channels", rowCount: 3, idColumns: []string{"channel_id"}},
			{schema: "space_mission", table: "mission_communications", rowCount: 3, idColumns: []string{"mission_id", "channel_id"}},
			{schema: "space_mission", table: "message_logs", rowCount: 3, idColumns: []string{"log_id"}},
			{schema: "space_mission", table: "events", rowCount: 9, idColumns: []string{"event_id"}},
			// {schema: "space_mission", table: "system_events", rowCount: 3},
			// {schema: "space_mission", table: "astronaut_events", rowCount: 3},
			// {schema: "space_mission", table: "mission_events", rowCount: 3},
			{schema: "space_mission", table: "telemetry", rowCount: 6, idColumns: []string{"telemetry_id"}},
			{schema: "space_mission", table: "telemetry_2023", rowCount: 2, idColumns: []string{"telemetry_id"}},
			{schema: "space_mission", table: "telemetry_2024", rowCount: 2, idColumns: []string{"telemetry_id"}},
			{schema: "space_mission", table: "telemetry_2025", rowCount: 2, idColumns: []string{"telemetry_id"}},
			{schema: "space_mission", table: "comments", rowCount: 4, idColumns: []string{"comment_id"}},
			{schema: "space_mission", table: "tags", rowCount: 4, idColumns: []string{"tag_id"}},
			{schema: "space_mission", table: "taggables", rowCount: 4, idColumns: []string{"tag_id"}},
			{schema: "space_mission", table: "mission_experiments", rowCount: 9, idColumns: []string{"mission_id", "experiment_id"}},
			{schema: "space_mission", table: "mission_parameters", rowCount: 3, idColumns: []string{"parameter_id"}},
			{schema: "space_mission", table: "skill_groups", rowCount: 6, idColumns: []string{"group_id"}},
			{schema: "space_mission", table: "capability_skill_groups", rowCount: 8, idColumns: []string{"capability_id", "group_id"}},
			{schema: "space_mission", table: "mission_required_skill_groups", rowCount: 7, idColumns: []string{"mission_id", "group_id", "role"}},
			{schema: "space_mission", table: "equipment_compatibility", rowCount: 5, idColumns: []string{"primary_equipment_id", "compatible_equipment_id"}},
			{schema: "space_mission", table: "mission_status_history", rowCount: 8, idColumns: []string{"history_id"}},
			{schema: "space_mission", table: "equipment_status_history", rowCount: 8, idColumns: []string{"history_id"}},
			{schema: "space_mission", table: "astronaut_role_history", rowCount: 5, idColumns: []string{"history_id"}},
			{schema: "space_mission", table: "astronaut_vitals", rowCount: 4, idColumns: []string{"vital_id"}},
			{schema: "scientific_data", table: "experiments", rowCount: 9, idColumns: []string{"experiment_id"}},
			{schema: "scientific_data", table: "samples", rowCount: 9, idColumns: []string{"sample_id"}},
			{schema: "scientific_data", table: "measurements", rowCount: 9, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2022", rowCount: 4, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2022_digital_microscope", rowCount: 2, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2022_mass_spectrometer", rowCount: 1, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2022_other", rowCount: 1, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2023", rowCount: 5, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2023_digital_microscope", rowCount: 2, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2023_mass_spectrometer", rowCount: 2, idColumns: []string{"measurement_id"}},
			{schema: "scientific_data", table: "measurement_2023_other", rowCount: 1, idColumns: []string{"measurement_id"}},
		}

		for _, expected := range expectedResults {
			rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
			require.NoError(t, err)
			assert.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: %s Table: %s", "space-mission", expected.table))
		}

		source, err := sql.Open("postgres", postgres.Source.URL)
		require.NoError(t, err)
		defer source.Close()

		target, err := sql.Open("postgres", postgres.Target.URL)
		require.NoError(t, err)
		defer target.Close()

		for _, e := range expectedResults {
			testutil_testdata.VerifySQLTableColumnValues(t, ctx, source, target, e.schema, e.table, "postgres", e.idColumns)
		}
	})

	t.Run("subset", func(t *testing.T) {
		subsetMappings := map[string]string{
			"space_mission.astronauts":    "astronaut_id in (1,2,3,4,5)",
			"space_mission.missions":      "mission_id < 6",
			"scientific_data.experiments": "experiment_id < 5",
		}

		neosyncApi.MockTemporalForCreateJob("test-postgres-sync")
		job := createPostgresSyncJob(t, ctx, jobclient, &createJobConfig{
			AccountId:   accountId,
			SourceConn:  sourceConn,
			DestConn:    destConn,
			JobName:     "space-mission-subset",
			JobMappings: jobmappings,
			SubsetMap:   subsetMappings,
			JobOptions: &TestJobOptions{
				Truncate:                      true,
				TruncateCascade:               true,
				InitSchema:                    true,
				SubsetByForeignKeyConstraints: true,
			},
		})

		testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithMaxIterations(10), WithPageLimit(100))
		testworkflow.RequireActivitiesCompletedSuccessfully(t)
		testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
		require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: space-mission-subset")
		err = testworkflow.TestEnv.GetWorkflowError()
		require.NoError(t, err, "Received Temporal Workflow Error: space-mission-subset")
		expectedResults := []struct {
			schema   string
			table    string
			rowCount int
		}{
			{schema: "space_mission", table: "astronauts", rowCount: 5},
			{schema: "space_mission", table: "missions", rowCount: 4},
			{schema: "space_mission", table: "objectives", rowCount: 3},
			{schema: "space_mission", table: "capabilities", rowCount: 10},
			{schema: "space_mission", table: "astronaut_capabilities", rowCount: 5},
			{schema: "space_mission", table: "transmissions", rowCount: 8},
			{schema: "space_mission", table: "payloads", rowCount: 3},
			{schema: "space_mission", table: "crew_assignments", rowCount: 3},
			{schema: "space_mission", table: "mission_logs", rowCount: 6},
			{schema: "space_mission", table: "crews", rowCount: 4},
			{schema: "space_mission", table: "crew_missions", rowCount: 3},
			{schema: "space_mission", table: "supplies", rowCount: 3},
			{schema: "space_mission", table: "supply_items", rowCount: 2},
			{schema: "space_mission", table: "spacecraft_class", rowCount: 3},
			{schema: "space_mission", table: "spacecraft", rowCount: 3},
			{schema: "space_mission", table: "spacecraft_module", rowCount: 3},
			{schema: "space_mission", table: "module_component", rowCount: 3},
			{schema: "space_mission", table: "equipment", rowCount: 3},
			{schema: "space_mission", table: "mission_equipment", rowCount: 3},
			{schema: "space_mission", table: "equipment_maintenance", rowCount: 3},
			{schema: "space_mission", table: "training_courses", rowCount: 3},
			{schema: "space_mission", table: "course_prerequisites", rowCount: 3},
			{schema: "space_mission", table: "certifications", rowCount: 3},
			{schema: "space_mission", table: "astronaut_certifications", rowCount: 4},
			{schema: "space_mission", table: "certification_requirements", rowCount: 3},
			{schema: "space_mission", table: "mission_logs_extended", rowCount: 2},
			{schema: "space_mission", table: "communication_channels", rowCount: 3},
			{schema: "space_mission", table: "mission_communications", rowCount: 3},
			{schema: "space_mission", table: "message_logs", rowCount: 2},
			{schema: "space_mission", table: "events", rowCount: 9},
			// {schema: "space_mission", table: "system_events", rowCount: 3},
			// {schema: "space_mission", table: "astronaut_events", rowCount: 3},
			// {schema: "space_mission", table: "mission_events", rowCount: 3},
			{schema: "space_mission", table: "telemetry", rowCount: 6},
			{schema: "space_mission", table: "telemetry_2023", rowCount: 2},
			{schema: "space_mission", table: "telemetry_2024", rowCount: 2},
			{schema: "space_mission", table: "telemetry_2025", rowCount: 2},
			{schema: "space_mission", table: "comments", rowCount: 4},
			{schema: "space_mission", table: "tags", rowCount: 4},
			{schema: "space_mission", table: "taggables", rowCount: 4},
			{schema: "space_mission", table: "mission_experiments", rowCount: 2},
			{schema: "space_mission", table: "mission_parameters", rowCount: 3},
			{schema: "space_mission", table: "skill_groups", rowCount: 6},
			{schema: "space_mission", table: "capability_skill_groups", rowCount: 8},
			{schema: "space_mission", table: "mission_required_skill_groups", rowCount: 7},
			{schema: "space_mission", table: "equipment_compatibility", rowCount: 5},
			{schema: "space_mission", table: "mission_status_history", rowCount: 8},
			{schema: "space_mission", table: "equipment_status_history", rowCount: 8},
			{schema: "space_mission", table: "astronaut_role_history", rowCount: 5},
			{schema: "space_mission", table: "astronaut_vitals", rowCount: 4},
			{schema: "scientific_data", table: "experiments", rowCount: 3},
			{schema: "scientific_data", table: "samples", rowCount: 4},
			{schema: "scientific_data", table: "measurements", rowCount: 4},
			{schema: "scientific_data", table: "measurement_2022", rowCount: 2},
			{schema: "scientific_data", table: "measurement_2022_digital_microscope", rowCount: 1},
			{schema: "scientific_data", table: "measurement_2022_mass_spectrometer", rowCount: 1},
			{schema: "scientific_data", table: "measurement_2022_other", rowCount: 0},
			{schema: "scientific_data", table: "measurement_2023", rowCount: 2},
			{schema: "scientific_data", table: "measurement_2023_digital_microscope", rowCount: 1},
			{schema: "scientific_data", table: "measurement_2023_mass_spectrometer", rowCount: 1},
			{schema: "scientific_data", table: "measurement_2023_other", rowCount: 0},
		}
		for _, expected := range expectedResults {
			rowCount, err := postgres.Target.GetTableRowCount(ctx, expected.schema, expected.table)
			require.NoError(t, err)
			assert.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: space-mission-subset Table: %s", expected.table))
		}
	})

	// tear down
	err = cleanupPostgresSchemas(ctx, postgres, []string{"space_mission", "scientific_data"})
	require.NoError(t, err)
}
