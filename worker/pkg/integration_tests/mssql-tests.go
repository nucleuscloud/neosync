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
	tcmssql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/sqlserver"
	mssql_alltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/mssql/alltypes"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata"
	"github.com/stretchr/testify/require"
)

const (
	mssqlTestdataFolder string = "../../../internal/testutil/testdata/mssql"
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
	dbManagers *tcworkflow.TestDatabaseManagers,
	accountId string,
	sourceConn, destConn *mgmtv1alpha1.Connection,
) {
	jobclient := neosyncApi.UnauthdClients.Jobs
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
		JobOptions: &workflow_testdata.TestJobOptions{
			Truncate:   true,
			InitSchema: true,
		},
	})

	testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, tcworkflow.WithValidEELicense())
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
		{schema: schema, table: "all_data_types", rowCount: 2},
	}

	for _, expected := range expectedResults {
		rowCount, err := mssql.Target.GetTableRowCount(ctx, expected.schema, expected.table)
		require.NoError(t, err)
		require.Equalf(t, expected.rowCount, rowCount, fmt.Sprintf("Test: mssql_all_types Table: %s", expected.table))
	}

	// tear down
	err = mssql.Source.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
	err = mssql.Target.DropSchemas(ctx, []string{schema})
	require.NoError(t, err)
}
