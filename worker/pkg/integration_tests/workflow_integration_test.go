package datasync_workflow_tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

	// testdata_javascripttransformers "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/javascript-transformers"

	testdata_pgtypes "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/all-types"
	testdata_doublereference "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/double-reference"

	// testdata_circulardependencies "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/circular-dependencies"

	// testdata_subsetting "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/subsetting"
	// testdata_virtualforeignkeys "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/virtual-foreign-keys"
	// testdata_primarykeytransformer "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/primary-key-transformer"
	// testdata_skipfkviolations "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/skip-fk-violations"

	"connectrpc.com/connect"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	tcredis "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/redis"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	temporalmocks "go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

const neosyncDbMigrationsPath = "../../../backend/sql/postgresql/schema"

func getAllPostgresSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}
	drTests := testdata_doublereference.GetSyncTests()
	// vfkTests := testdata_virtualforeignkeys.GetSyncTests()
	// cdTests := testdata_circulardependencies.GetSyncTests()
	// javascriptTests := testdata_javascripttransformers.GetSyncTests()
	// pkTransformationTests := testdata_primarykeytransformer.GetSyncTests()
	// subsettingTests := testdata_subsetting.GetSyncTests()
	pgTypesTests := testdata_pgtypes.GetSyncTests()
	// skipFkViolationTests := testdata_skipfkviolations.GetSyncTests()

	allTests["Double_References"] = drTests
	// allTests["Virtual_Foreign_Keys"] = vfkTests
	// allTests["Circular_Dependencies"] = cdTests
	// allTests["Javascript_Transformers"] = javascriptTests
	// allTests["Primary_Key_Transformers"] = pkTransformationTests
	// allTests["Subsetting"] = subsettingTests
	allTests["PG_Types"] = pgTypesTests
	// allTests["Skip_ForeignKey_Violations"] = skipFkViolationTests
	return allTests
}

func Test_Workflow(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()

	neosyncApi, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t, tcneosyncapi.WithMigrationsDirectory(neosyncDbMigrationsPath))
	if err != nil {
		panic(err)
	}

	redis, err := tcredis.NewRedisTestContainer(ctx)
	if err != nil {
		panic(err)
	}

	connclient := neosyncApi.UnauthdClients.Connections
	jobclient := neosyncApi.UnauthdClients.Jobs
	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnauthdClients.Users)

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		postgres, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
		if err != nil {
			panic(err)
		}
		sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-source", postgres.Source.URL)
		destConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-dest", postgres.Target.URL)
		tests := getAllPostgresSyncTests()

		neosyncApi.Mocks.TemporalClientManager.
			On(
				"DoesAccountHaveNamespace", mock.Anything, mock.Anything, mock.Anything,
			).
			Return(true, nil)
		neosyncApi.Mocks.TemporalClientManager.
			On(
				"GetSyncJobTaskQueue", mock.Anything, mock.Anything, mock.Anything,
			).
			Return("sync-job", nil)
		neosyncApi.Mocks.TemporalClientManager.
			On(
				"CreateSchedule", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			).
			Return("test-id", nil)

		for groupName, group := range tests {
			group := group
			t.Run(groupName, func(t *testing.T) {
				t.Parallel()
				for _, tt := range group {
					t.Run(tt.Name, func(t *testing.T) {
						t.Logf("running integration test: %s \n", tt.Name)
						// setup
						err := postgres.Source.RunSqlFiles(ctx, &tt.Folder, tt.SourceFilePaths)
						require.NoError(t, err)
						err = postgres.Target.RunSqlFiles(ctx, &tt.Folder, tt.TargetFilePaths)
						require.NoError(t, err)

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
						var destinationOptions *mgmtv1alpha1.JobDestinationOptions
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

						env := executeWorkflow(t, neosyncApi, redis.URL, job.Msg.GetJob().GetId())
						require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
						err = env.GetWorkflowError()
						if tt.ExpectError {
							require.Error(t, err, "Did not received Temporal Workflow Error", "testName", tt.Name)
							return
						}
						require.NoError(t, err, "Received Temporal Workflow Error", "testName", tt.Name)

						for table, expected := range tt.Expected {
							rows, err := postgres.Target.DB.Query(ctx, fmt.Sprintf("select * from %s;", table))
							require.NoError(t, err)
							count := 0
							for rows.Next() {
								count++
							}
							require.Equalf(t, expected.RowCount, count, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
						}

						// tear down
						err = postgres.Source.RunSqlFiles(ctx, &tt.Folder, []string{"teardown.sql"})
						require.NoError(t, err)
						err = postgres.Target.RunSqlFiles(ctx, &tt.Folder, []string{"teardown.sql"})
						require.NoError(t, err)
					})
				}
			})
		}

		t.Cleanup(func() {
			err := postgres.TearDown(ctx)
			if err != nil {
				panic(err)
			}
		})
	})

	// t.Run("mysql", func(t *testing.T) {
	// 	t.Parallel()

	// 	t.Run("sync", func(t *testing.T) {
	// 		t.Parallel()
	// 		testdataFolder := "../../../../../internal/testutil/testdata/mysql/humanresources"
	// 		err = mysql.Source.RunSqlFiles(ctx, &testdataFolder, []string{"create-tables.sql"})
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		err = mysql.Target.RunSqlFiles(ctx, &testdataFolder, []string{"create-schema.sql"})
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		sourceConn := tcneosyncapi.CreateMysqlConnection(ctx, t, neosyncApi.UnauthdClients.Connections, accountId, "mysql-source", mysql.Source.URL)
	// 	})

	// 	t.Cleanup(func() {
	// 		err := mysql.TearDown(ctx)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	})
	// })

	t.Cleanup(func() {
		err = neosyncApi.TearDown(ctx)
		if err != nil {
			panic(err)
		}
	})
}

func executeWorkflow(
	t *testing.T,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	redisUrl string,
	jobId string,
) *testsuite.TestWorkflowEnvironment {
	connclient := neosyncApi.UnauthdClients.Connections
	jobclient := neosyncApi.UnauthdClients.Jobs
	transformerclient := neosyncApi.UnauthdClients.Transformers
	userclient := neosyncApi.UnauthdClients.Users
	// sqlconnector := &sqlconnect.SqlOpenConnector{}
	redisconfig := &shared.RedisConfig{
		Url:  redisUrl,
		Kind: "simple",
		Tls: &shared.RedisTlsConfig{
			Enabled: false,
		},
	}
	temporalClientMock := temporalmocks.NewClient(t)
	// pgpoolmap := &sync.Map{}
	// mysqlpoolmap := &sync.Map{}
	// mssqlpoolmap := &sync.Map{}
	// pgquerier := pg_queries.New()
	// mysqlquerier := mysql_queries.New()
	// mssqlquerier := mssql_queries.New()
	// sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlconnector)
	sqlmanager := tcneosyncapi.NewTestSqlManagerClient()

	// temporal workflow
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// register activities
	genbenthosActivity := genbenthosconfigs_activity.New(
		jobclient,
		connclient,
		transformerclient,
		sqlmanager,
		redisconfig,
		false,
	)
	var activityMeter metric.Meter
	disableReaper := true
	syncActivity := sync_activity.New(connclient, jobclient, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, temporalClientMock, activityMeter, sync_activity.NewBenthosStreamManager(), disableReaper)
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager)
	accountStatusActivity := accountstatus_activity.New(userclient)
	env.RegisterWorkflow(datasync_workflow.Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	env.SetTestTimeout(600 * time.Second) // increase the test timeout

	env.ExecuteWorkflow(datasync_workflow.Workflow, &datasync_workflow.WorkflowRequest{JobId: jobId})
	return env
}
