package integration_tests

import (
	"context"
	"fmt"
	"slices"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	testdata_primarykeytransformer "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata/primary-key-transformer"

	"connectrpc.com/connect"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	tcredis "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/redis"
	pg_edgecases "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/edgecases"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata"
	pg_alltypes "github.com/nucleuscloud/neosync/worker/pkg/integration_tests/testdata/postgres/all-types"
	"github.com/stretchr/testify/require"
)

const neosyncDbMigrationsPath = "../../../backend/sql/postgresql/schema"

func Test_Workflow(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()

	neosyncApi, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t, tcneosyncapi.WithMigrationsDirectory(neosyncDbMigrationsPath))
	if err != nil {
		t.Fatal(err)
	}

	connclient := neosyncApi.UnauthdClients.Connections
	jobclient := neosyncApi.UnauthdClients.Jobs
	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnauthdClients.Users)
	dbManagers := tcworkflow.NewTestDatabaseManagers(t)
	testdataFolder := "../../../../../internal/testutil/testdata/postgres"
	localTestdataFolder := "./testdata/postgres"

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		postgres, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
		if err != nil {
			t.Fatal(err)
		}
		sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-source", postgres.Source.URL)
		destConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-dest", postgres.Target.URL)

		t.Run("all_types", func(t *testing.T) {
			t.Parallel()
			// setup
			capitalSchema := "CaPiTaL"
			alltypesSchema := "alltypes"
			schemas := []string{alltypesSchema, capitalSchema}
			err := postgres.Source.RunCreateStmtsInSchema(ctx, &localTestdataFolder, []string{"all-types/create-tables.sql"}, alltypesSchema)
			require.NoError(t, err)
			err = postgres.Source.RunCreateStmtsInSchema(ctx, &testdataFolder, []string{"edgecases/create-tables.sql"}, capitalSchema)
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
				rowCount int64
			}{
				{schema: alltypesSchema, table: "all_postgres_types", rowCount: 2},
				{schema: alltypesSchema, table: "array_types", rowCount: 1},
				{schema: alltypesSchema, table: "time_time", rowCount: 2},
				{schema: alltypesSchema, table: "json_data", rowCount: 12},
				{schema: capitalSchema, table: "badname", rowCount: 5},
				{schema: capitalSchema, table: "bad_name_123!@#", rowCount: 5},
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
		})

		t.Run("Primary_Key_Transformers", func(t *testing.T) {
			t.Parallel()
			redis, err := tcredis.NewRedisTestContainer(ctx)
			require.NoError(t, err)

			tt := testdata_primarykeytransformer.GetSyncTest()
			// setup
			err = postgres.Source.RunSqlFiles(ctx, &tt.Folder, tt.SourceFilePaths)
			require.NoError(t, err)
			err = postgres.Target.RunSqlFiles(ctx, &tt.Folder, tt.TargetFilePaths)
			require.NoError(t, err)
			neosyncApi.MockTemporalForCreateJob("test-postgres-sync")

			schemas := []*mgmtv1alpha1.PostgresSourceSchemaOption{}
			subsetMap := map[string]*mgmtv1alpha1.PostgresSourceSchemaOption{}
			for table, where := range tt.SubsetMap {
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
			if tt.JobOptions != nil {
				if tt.JobOptions.SubsetByForeignKeyConstraints {
					subsetByForeignKeyConstraints = true
				}
				destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
							InitTableSchema: tt.JobOptions.InitSchema,
							TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
								TruncateBeforeInsert: tt.JobOptions.Truncate,
							},
							SkipForeignKeyViolations: tt.JobOptions.SkipForeignKeyViolations,
						},
					},
				}
			}

			job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
				AccountId: accountId,
				JobName:   tt.Name,
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId:                  sourceConn.Id,
								Schemas:                       schemas,
								SubsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
							},
						},
					},
				},
				Destinations: []*mgmtv1alpha1.CreateJobDestination{
					{
						ConnectionId: destConn.Id,
						Options:      destinationOptions,
					},
				},
				Mappings:           tt.JobMappings,
				VirtualForeignKeys: tt.VirtualForeignKeys,
			}))
			require.NoError(t, err)

			testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers, tcworkflow.WithRedis(redis.URL))
			testworkflow.RequireActivitiesCompletedSuccessfully(t)
			testworkflow.ExecuteTestDataSyncWorkflow(job.Msg.GetJob().GetId())
			require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
			err = testworkflow.TestEnv.GetWorkflowError()
			require.NoError(t, err, "Received Temporal Workflow Error", "testName", tt.Name)

			for table, expected := range tt.Expected {
				rows := postgres.Target.DB.QueryRow(ctx, fmt.Sprintf("select count(*) from %s;", table))
				var rowCount int
				err = rows.Scan(&rowCount)
				require.NoError(t, err)
				require.Equalf(t, expected.RowCount, rowCount, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
			}

			keys, err := testworkflow.Redisclient.Keys(ctx, "*").Result()
			if err != nil {
				t.Fatal(err)
			}
			require.Emptyf(t, keys, "Redis keys should be empty")

			// tear down
			err = postgres.Source.RunSqlFiles(ctx, &tt.Folder, []string{"teardown.sql"})
			require.NoError(t, err)
			err = postgres.Target.RunSqlFiles(ctx, &tt.Folder, []string{"teardown.sql"})
			require.NoError(t, err)

			t.Cleanup(func() {
				err := redis.TearDown(ctx)
				require.NoError(t, err)
			})
		})

		t.Cleanup(func() {
			err := postgres.TearDown(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Cleanup(func() {
		err = neosyncApi.TearDown(ctx)
		if err != nil {
			panic(err)
		}
	})
}
