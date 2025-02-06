package integrationtest

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
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	testutil_testdata "github.com/nucleuscloud/neosync/internal/testutil/testdata"
	mysql_alltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/mysql/alltypes"
	mysql_composite_keys "github.com/nucleuscloud/neosync/internal/testutil/testdata/mysql/composite-keys"
	mysql_edgecases "github.com/nucleuscloud/neosync/internal/testutil/testdata/mysql/edgecases"
	mysql_human_resources "github.com/nucleuscloud/neosync/internal/testutil/testdata/mysql/humanresources"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const (
	mysqlTestdataFolder string = "../../../testutil/testdata/mysql"
)

func createMysqlSyncJob(
	t *testing.T,
	ctx context.Context,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	config *createJobConfig,
) *mgmtv1alpha1.Job {
	schemas := []*mgmtv1alpha1.MysqlSourceSchemaOption{}
	subsetMap := map[string]*mgmtv1alpha1.MysqlSourceSchemaOption{}
	for table, where := range config.SubsetMap {
		schema, table := sqlmanager_shared.SplitTableKey(table)
		if _, exists := subsetMap[schema]; !exists {
			subsetMap[schema] = &mgmtv1alpha1.MysqlSourceSchemaOption{
				Schema: schema,
				Tables: []*mgmtv1alpha1.MysqlSourceTableOption{},
			}
		}
		w := where
		subsetMap[schema].Tables = append(subsetMap[schema].Tables, &mgmtv1alpha1.MysqlSourceTableOption{
			Table:       table,
			WhereClause: &w,
		})
	}

	for _, s := range subsetMap {
		schemas = append(schemas, s)
	}

	destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
			MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{},
		},
	}
	if config.JobOptions != nil {
		onConflict := &mgmtv1alpha1.MysqlOnConflictConfig{}
		if config.JobOptions.OnConflictDoUpdate {
			onConflict.Strategy = &mgmtv1alpha1.MysqlOnConflictConfig_Update{}
		}
		destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
				MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
					InitTableSchema: config.JobOptions.InitSchema,
					TruncateTable: &mgmtv1alpha1.MysqlTruncateTableConfig{
						TruncateBeforeInsert: config.JobOptions.Truncate,
					},
					OnConflict:               onConflict,
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
				Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
					Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
						ConnectionId:                  config.SourceConn.Id,
						Schemas:                       schemas,
						SubsetByForeignKeyConstraints: config.JobOptions.SubsetByForeignKeyConstraints,
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

func test_mysql_types(
	t *testing.T,
	ctx context.Context,
	mysql *tcmysql.MysqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	alltypesSchema := "alltypes"

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return mysql.Source.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
	})
	errgrp.Go(func() error { return mysql.Target.CreateDatabases(errctx, []string{alltypesSchema}) })
	err := errgrp.Wait()
	require.NoError(t, err)

	neosyncApi.MockTemporalForCreateJob("test-mysql-sync")

	alltypesMappings := mysql_alltypes.GetDefaultSyncJobMappings(alltypesSchema)

	job := createMysqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mysql_all_types",
		JobMappings: alltypesMappings,
		JobOptions: &TestJobOptions{
			Truncate:   true,
			InitSchema: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mysql_all_types")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mysql_all_types")

	expectedResults := []struct {
		schema   string
		table    string
		rowCount int
	}{
		{schema: alltypesSchema, table: "all_data_types", rowCount: 2},
		{schema: alltypesSchema, table: "json_data", rowCount: 12},
		{schema: alltypesSchema, table: "generated_table", rowCount: 10},
	}

	for _, expected := range expectedResults {
		rowCount, err := mysql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mysql_all_types Table: %s", expected.table))
	}

	testutil_testdata.VerifySQLTableColumnValues(t, ctx, mysql.Source.DB, mysql.Target.DB, alltypesSchema, "all_data_types", sqlmanager_shared.MysqlDriver, "id")
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, mysql.Source.DB, mysql.Target.DB, alltypesSchema, "json_data", sqlmanager_shared.MysqlDriver, "id")
	testutil_testdata.VerifySQLTableColumnValues(t, ctx, mysql.Source.DB, mysql.Target.DB, alltypesSchema, "generated_table", sqlmanager_shared.MysqlDriver, "id")

	// tear down
	err = cleanupMysqlDatabases(ctx, mysql, []string{alltypesSchema})
	require.NoError(t, err)
}

func test_mysql_edgecases(
	t *testing.T,
	ctx context.Context,
	mysql *tcmysql.MysqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "mysqledgecases"
	schema2 := "mysqledgecasesother"

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return mysql.Source.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"edgecases/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return mysql.Source.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"edgecases/create-tables.sql"}, schema2)
	})
	errgrp.Go(func() error { return mysql.Target.CreateDatabases(errctx, []string{schema, schema2}) })
	err := errgrp.Wait()
	require.NoError(t, err)

	neosyncApi.MockTemporalForCreateJob("test-mysql-sync")

	mappings := mysql_edgecases.GetDefaultSyncJobMappings(schema)
	mappings2 := mysql_edgecases.GetDefaultSyncJobMappings(schema2)

	job := createMysqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mysql_edgecases",
		JobMappings: slices.Concat(mappings, mappings2),
		JobOptions: &TestJobOptions{
			Truncate:   true,
			InitSchema: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mysql_edgecases")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mysql_edgecases")

	expectedResults := []struct {
		table    string
		rowCount int
	}{
		{table: "container", rowCount: 5},
		{table: "container_status", rowCount: 5},
		{table: "container", rowCount: 5},
		{table: "users", rowCount: 5},
		{table: "unique_emails", rowCount: 5},
		{table: "unique_emails_and_usernames", rowCount: 5},
		{table: "t1", rowCount: 5},
		{table: "t2", rowCount: 5},
		{table: "t3", rowCount: 5},
		{table: "parent1", rowCount: 5},
		{table: "child1", rowCount: 5},
		{table: "t4", rowCount: 5},
		{table: "t5", rowCount: 5},
		{table: "employee_log", rowCount: 5},
		{table: "custom_table", rowCount: 5},
		{table: "tablewithcount", rowCount: 5},
	}

	// check schema1
	for _, expected := range expectedResults {
		rowCount, err := mysql.Target.GetTableRowCount(ctx, schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mysql_edgecases schema: %s Table: %s", schema, expected.table))
	}

	// check schema2
	for _, expected := range expectedResults {
		rowCount, err := mysql.Target.GetTableRowCount(ctx, schema2, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mysql_edgecases schema: %s Table: %s", schema2, expected.table))
	}

	// tear down
	err = cleanupMysqlDatabases(ctx, mysql, []string{schema, schema2})
	require.NoError(t, err)
}

func test_mysql_composite_keys(
	t *testing.T,
	ctx context.Context,
	mysql *tcmysql.MysqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "mysqlcompositekeys"

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return mysql.Source.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"composite-keys/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return mysql.Target.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"composite-keys/create-tables.sql"}, schema)
	})
	err := errgrp.Wait()
	require.NoError(t, err)

	neosyncApi.MockTemporalForCreateJob("test-mysql-sync")

	mappings := mysql_composite_keys.GetDefaultSyncJobMappings(schema)

	job := createMysqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mysql_composite_keys",
		JobMappings: mappings,
		JobOptions: &TestJobOptions{
			Truncate:   true,
			InitSchema: false,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mysql_composite_keys")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mysql_composite_keys")

	expectedResults := []struct {
		table    string
		rowCount int
	}{
		{table: "order_details", rowCount: 10},
		{table: "orders", rowCount: 10},
		{table: "order_shipping", rowCount: 10},
		{table: "shipping_status", rowCount: 10},
	}

	for _, expected := range expectedResults {
		rowCount, err := mysql.Target.GetTableRowCount(ctx, schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mysql_composite_keys schema: %s Table: %s", schema, expected.table))
	}

	// tear down
	err = cleanupMysqlDatabases(ctx, mysql, []string{schema})
	require.NoError(t, err)
}

func test_mysql_on_conflict_do_update(
	t *testing.T,
	ctx context.Context,
	mysql *tcmysql.MysqlTestSyncContainer,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()
	schema := "human_resources"

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return mysql.Source.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"humanresources/create-tables.sql"}, schema)
	})
	errgrp.Go(func() error {
		return mysql.Target.RunCreateStmtsInDatabase(errctx, mysqlTestdataFolder, []string{"humanresources/create-tables.sql"}, schema)
	})
	err := errgrp.Wait()
	require.NoError(t, err)

	// update the source data to be different from target data
	updateStmt := `
	UPDATE human_resources.regions 
	SET region_name = CASE region_id
			WHEN 1 THEN 'Modified Europe'
			WHEN 2 THEN 'Modified Americas'
			WHEN 3 THEN 'Modified Asia'
			WHEN 4 THEN 'Modified Africa'
	END
	WHERE region_id IN (1,2,3,4)`
	_, err = mysql.Source.DB.ExecContext(ctx, updateStmt)
	require.NoError(t, err)

	neosyncApi.MockTemporalForCreateJob("test-mysql-sync")

	mappings := mysql_human_resources.GetDefaultSyncJobMappings(schema)

	job := createMysqlSyncJob(t, ctx, jobclient, &createJobConfig{
		AccountId:   accountId,
		SourceConn:  sourceConn,
		DestConn:    destConn,
		JobName:     "mysql_human_resources",
		JobMappings: mappings,
		JobOptions: &TestJobOptions{
			Truncate:           false,
			InitSchema:         false,
			OnConflictDoUpdate: true,
		},
	})

	testworkflow := NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
	testworkflow.RequireActivitiesCompletedSuccessfully(t)
	testworkflow.ExecuteTestDataSyncWorkflow(job.GetId())
	require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mysql_human_resources")
	err = testworkflow.TestEnv.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error: mysql_human_resources")

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
		rowCount, err := mysql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mysql_human_resources Table: %s", expected.table))
	}

	testutil_testdata.VerifySQLTableColumnValues(t, ctx, mysql.Source.DB, mysql.Target.DB, schema, "regions", sqlmanager_shared.MysqlDriver, "region_id")

	// tear down
	err = cleanupMysqlDatabases(ctx, mysql, []string{schema})
	require.NoError(t, err)
}

func cleanupMysqlDatabases(ctx context.Context, mysql *tcmysql.MysqlTestSyncContainer, databases []string) error {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error { return mysql.Source.DropDatabases(errctx, databases) })
	errgrp.Go(func() error { return mysql.Target.DropDatabases(errctx, databases) })
	return errgrp.Wait()
}
