package integrationtest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tcmssql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/sqlserver"
	testutil_testdata "github.com/nucleuscloud/neosync/internal/testutil/testdata"
	mssql_alltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/mssql/alltypes"
	mssql_commerce "github.com/nucleuscloud/neosync/internal/testutil/testdata/mssql/commerce"
	"github.com/stretchr/testify/require"
)

const (
	mssqlTestdataFolder string = "../../../testutil/testdata/mssql"
)

func createMssqlSyncJob(
	t *testing.T,
	ctx context.Context,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	config *createJobConfig,
) *mgmtv1alpha1.Job {
	schemas := []*mgmtv1alpha1.MssqlSourceSchemaOption{}
	subsetMap := map[string]*mgmtv1alpha1.MssqlSourceSchemaOption{}
	for table, where := range config.SubsetMap {
		schema, table := sqlmanager_shared.SplitTableKey(table)
		if _, exists := subsetMap[schema]; !exists {
			subsetMap[schema] = &mgmtv1alpha1.MssqlSourceSchemaOption{
				Schema: schema,
				Tables: []*mgmtv1alpha1.MssqlSourceTableOption{},
			}
		}
		w := where
		subsetMap[schema].Tables = append(subsetMap[schema].Tables, &mgmtv1alpha1.MssqlSourceTableOption{
			Table:       table,
			WhereClause: &w,
		})
	}

	for _, s := range subsetMap {
		schemas = append(schemas, s)
	}

	var subsetByForeignKeyConstraints bool
	destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_MssqlOptions{
			MssqlOptions: &mgmtv1alpha1.MssqlDestinationConnectionOptions{},
		},
	}
	if config.JobOptions != nil {
		if config.JobOptions.SubsetByForeignKeyConstraints {
			subsetByForeignKeyConstraints = true
		}
		destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_MssqlOptions{
				MssqlOptions: &mgmtv1alpha1.MssqlDestinationConnectionOptions{
					InitTableSchema: config.JobOptions.InitSchema,
					TruncateTable: &mgmtv1alpha1.MssqlTruncateTableConfig{
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
				Config: &mgmtv1alpha1.JobSourceOptions_Mssql{
					Mssql: &mgmtv1alpha1.MssqlSourceConnectionOptions{
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

func test_mssql_types(
	t *testing.T,
	ctx context.Context,
	mssql *tcmssql.MssqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "alltypes"
	err := mssql.Source.RunCreateStmtsInSchema(ctx, mssqlTestdataFolder, []string{"alltypes/create-tables.sql"}, schema)
	require.NoError(t, err)
	err = mssql.Target.CreateSchemas(ctx, []string{schema})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-mssql-sync")

	alltypesMappings := mssql_alltypes.GetDefaultSyncJobMappings(schema)

	job := createMssqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mssql_all_types",
		JobMappings: alltypesMappings,
		JobOptions: &TestJobOptions{
			Truncate:   true,
			InitSchema: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithValidEELicense())
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mssql_all_types")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mssql_all_types")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: schema, table: "alldatatypes", rowCount: 1},
	}

	for _, expected := range expectedResults {
		rowCount, err := mssql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mssql_all_types Table: %s", expected.table))
	}

	testutil_testdata.VerifySQLTableColumnValues(t, ctx, mssql.Source.DB, mssql.Target.DB, schema, "alldatatypes", sqlmanager_shared.MssqlDriver, "id")

	// TODO: Tear down, fix schema dropping issue. No way to force drop schemas in MSSQL.
	// err = mssql.Source.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
	// err = mssql.Target.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
}

func test_mssql_cross_schema_foreign_keys(
	t *testing.T,
	ctx context.Context,
	mssql *tcmssql.MssqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	testdataFolder := mssqlTestdataFolder + "/commerce"
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	err := mssql.Source.CreateSchemas(ctx, []string{"sales", "production"})
	require.NoError(t, err)
	err = mssql.Source.RunSqlFiles(ctx, &testdataFolder, []string{"create-tables.sql"})
	require.NoError(t, err)
	err = mssql.Target.CreateSchemas(ctx, []string{"sales", "production"})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-mssql-sync")

	mappings := mssql_commerce.GetDefaultSyncJobMappings()

	job := createMssqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mssql_cross_schema_foreign_keys",
		JobMappings: mappings,
		JobOptions: &TestJobOptions{
			Truncate:   true,
			InitSchema: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithValidEELicense())
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mssql_cross_schema_foreign_keys")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mssql_cross_schema_foreign_keys")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: "production", table: "categories", rowCount: 7},
		{schema: "production", table: "brands", rowCount: 9},
		{schema: "production", table: "products", rowCount: 18},
		{schema: "production", table: "stocks", rowCount: 32},
		{schema: "production", table: "identities", rowCount: 5},
		{schema: "sales", table: "customers", rowCount: 15},
		{schema: "sales", table: "stores", rowCount: 3},
		{schema: "sales", table: "staffs", rowCount: 10},
		{schema: "sales", table: "orders", rowCount: 13},
		{schema: "sales", table: "order_items", rowCount: 26},
	}

	for _, expected := range expectedResults {
		rowCount, err := mssql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mssql_cross_schema_foreign_keys Table: %s", expected.table))
	}

	// TODO: Tear down, fix schema dropping issue. No way to force drop schemas in MSSQL.
	// err = mssql.Source.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
	// err = mssql.Target.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
}

func test_mssql_subset(
	t *testing.T,
	ctx context.Context,
	mssql *tcmssql.MssqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	testdataFolder := mssqlTestdataFolder + "/commerce"
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	err := mssql.Source.CreateSchemas(ctx, []string{"sales_subset", "production_subset"})
	require.NoError(t, err)
	err = createCommerceTables(ctx, mssql.Source, &testdataFolder, []string{"create-tables.sql"}, "subset")
	require.NoError(t, err)
	err = mssql.Target.CreateSchemas(ctx, []string{"sales_subset", "production_subset"})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-mssql-sync")

	mappings := mssql_commerce.GetDefaultSyncJobMappings()
	updatedMappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range mappings {
		updatedMappings = append(updatedMappings, &mgmtv1alpha1.JobMapping{
			Schema:      fmt.Sprintf("%s_subset", jm.Schema),
			Table:       jm.Table,
			Column:      jm.Column,
			Transformer: jm.Transformer,
		})
	}

	subsetMappings := map[string]string{
		"production_subset.products": "product_id in (1, 4, 8, 6)",
		"sales_subset.customers":     "customer_id in (1, 4, 8, 6)",
	}

	job := createMssqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mssql_subset",
		JobMappings: updatedMappings,
		SubsetMap:   subsetMappings,
		JobOptions: &TestJobOptions{
			Truncate:                      true,
			InitSchema:                    true,
			SubsetByForeignKeyConstraints: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithValidEELicense())
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mssql_subset")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mssql_subset")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: "production_subset", table: "categories", rowCount: 7},
		{schema: "production_subset", table: "brands", rowCount: 9},
		{schema: "production_subset", table: "products", rowCount: 4},
		{schema: "production_subset", table: "stocks", rowCount: 10},
		{schema: "production_subset", table: "identities", rowCount: 5},
		{schema: "sales_subset", table: "customers", rowCount: 4},
		{schema: "sales_subset", table: "stores", rowCount: 3},
		{schema: "sales_subset", table: "staffs", rowCount: 10},
		{schema: "sales_subset", table: "orders", rowCount: 4},
		{schema: "sales_subset", table: "order_items", rowCount: 2},
	}

	for _, expected := range expectedResults {
		rowCount, err := mssql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mssql_subset Table: %s", expected.table))
	}

	// TODO: Tear down, fix schema dropping issue. No way to force drop schemas in MSSQL.
	// err = mssql.Source.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
	// err = mssql.Target.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
}

func test_mssql_identity_columns(
	t *testing.T,
	ctx context.Context,
	mssql *tcmssql.MssqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	testdataFolder := mssqlTestdataFolder + "/commerce"
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	err := mssql.Source.CreateSchemas(ctx, []string{"sales_identity", "production_identity"})
	require.NoError(t, err)
	err = createCommerceTables(ctx, mssql.Source, &testdataFolder, []string{"create-tables.sql"}, "identity")
	require.NoError(t, err)
	err = mssql.Target.CreateSchemas(ctx, []string{"sales_identity", "production_identity"})
	require.NoError(t, err)
	neosyncApi.MockTemporalForCreateJob("test-mssql-sync")

	mappings := mssql_commerce.GetDefaultSyncJobMappings()
	tableColTypeMap := mssql_commerce.GetTableColumnTypeMap()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range mappings {
		colTypeMap, ok := tableColTypeMap[fmt.Sprintf("%s.%s", jm.Schema, jm.Table)]
		if ok {
			t, ok := colTypeMap[jm.Column]
			if ok && strings.HasPrefix(t, "INTIDENTITY") {
				updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
					Schema:      fmt.Sprintf("%s_identity", jm.Schema),
					Table:       jm.Table,
					Column:      jm.Column,
					Transformer: getDefaultTransformerConfig(),
				})
				continue
			}
		}
		jm.Schema = fmt.Sprintf("%s_identity", jm.Schema)
		updatedJobmappings = append(updatedJobmappings, jm)
	}

	job := createMssqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mssql_identity_columns",
		JobMappings: updatedJobmappings,
		JobOptions: &TestJobOptions{
			// Truncate:   true,
			InitSchema: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, WithValidEELicense())
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mssql_identity_columns")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mssql_identity_columns")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: "production_identity", table: "categories", rowCount: 7},
		{schema: "production_identity", table: "brands", rowCount: 9},
		{schema: "production_identity", table: "products", rowCount: 18},
		{schema: "production_identity", table: "stocks", rowCount: 32},
		{schema: "production_identity", table: "identities", rowCount: 5},
		{schema: "sales_identity", table: "customers", rowCount: 15},
		{schema: "sales_identity", table: "stores", rowCount: 3},
		{schema: "sales_identity", table: "staffs", rowCount: 10},
		{schema: "sales_identity", table: "orders", rowCount: 13},
		{schema: "sales_identity", table: "order_items", rowCount: 26},
	}

	for _, expected := range expectedResults {
		rowCount, err := mssql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mssql_identity_columns Table: %s", expected.table))
	}

	// TODO: Tear down, fix schema dropping issue. No way to force drop schemas in MSSQL.
	// err = mssql.Source.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
	// err = mssql.Target.DropSchemas(ctx, []string{schema})
	// require.NoError(t, err)
}

func getDefaultTransformerConfig() *mgmtv1alpha1.JobMappingTransformer {
	return &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
				GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
			},
		},
	}
}

func createCommerceTables(ctx context.Context, mssql *tcmssql.MssqlTestContainer, folder *string, files []string, schemaSuffix string) error {
	for _, file := range files {
		filePath := file
		if folder != nil && *folder != "" {
			filePath = fmt.Sprintf("./%s/%s", *folder, file)
		}
		sqlStr, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		updatedSql := strings.ReplaceAll(string(sqlStr), "sales.", fmt.Sprintf("sales_%s.", schemaSuffix))
		updatedSql = strings.ReplaceAll(updatedSql, "production.", fmt.Sprintf("production_%s.", schemaSuffix))
		_, err = mssql.DB.ExecContext(ctx, updatedSql)
		if err != nil {
			return fmt.Errorf("unable to exec SQL when running MsSQL SQL files: %w", err)
		}
	}
	return nil
}
