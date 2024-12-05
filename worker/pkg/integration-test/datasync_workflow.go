package integrationtest

import (
	"fmt"
	"testing"
	"time"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	"github.com/nucleuscloud/neosync/internal/testutil"
	neosync_redis "github.com/nucleuscloud/neosync/worker/internal/redis"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	posttablesync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/post-table-sync"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

type Option func(*TestWorkflowEnv)

type TestWorkflowEnv struct {
	neosyncApi    *tcneosyncapi.NeosyncApiTestClient
	redisconfig   *shared.RedisConfig
	redisclient   redis.UniversalClient
	fakeEELicense *testutil.FakeEELicense
	TestEnv       *testsuite.TestWorkflowEnvironment
}

// WithRedis creates redis client with provided URL
func WithRedis(url string) Option {
	return func(c *TestWorkflowEnv) {
		c.redisconfig = &shared.RedisConfig{
			Url:  url,
			Kind: "simple",
			Tls: &shared.RedisTlsConfig{
				Enabled: false,
			},
		}
	}
}

// WithValidEELicense creates a valid enterprise edition license
func WithValidEELicense() Option {
	return func(c *TestWorkflowEnv) {
		c.fakeEELicense = testutil.NewFakeEELicense(testutil.WithIsValid())
	}
}

// NewTestDataSyncWorkflowEnv creates and configures a new test datasync workflow environment
func NewTestDataSyncWorkflowEnv(
	t testing.TB,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	dbManagers *TestDatabaseManagers,
	opts ...Option,
) *TestWorkflowEnv {
	t.Helper()

	workflowEnv := &TestWorkflowEnv{
		neosyncApi:    neosyncApi,
		fakeEELicense: testutil.NewFakeEELicense(),
	}

	for _, opt := range opts {
		opt(workflowEnv)
	}

	redisclient, err := neosync_redis.GetRedisClient(workflowEnv.redisconfig)
	if err != nil {
		t.Fatal(err)
	}
	workflowEnv.redisclient = redisclient

	connclient := neosyncApi.UnauthdClients.Connections
	jobclient := neosyncApi.UnauthdClients.Jobs
	transformerclient := neosyncApi.UnauthdClients.Transformers
	userclient := neosyncApi.UnauthdClients.Users

	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestWorkflowEnvironment()

	genbenthosActivity := genbenthosconfigs_activity.New(
		jobclient,
		connclient,
		transformerclient,
		dbManagers.SqlManager,
		workflowEnv.redisconfig,
		false,
	)

	var activityMeter metric.Meter
	syncActivity := sync_activity.New(connclient, jobclient, dbManagers.SqlConnManager, dbManagers.MongoConnManager, activityMeter, sync_activity.NewBenthosStreamManager())
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, dbManagers.SqlManager, workflowEnv.fakeEELicense)
	jobhookTimingActivity := jobhooks_by_timing_activity.New(jobclient, connclient, dbManagers.SqlManager, workflowEnv.fakeEELicense)
	accountStatusActivity := accountstatus_activity.New(userclient)
	posttableSyncActivity := posttablesync_activity.New(jobclient, dbManagers.SqlManager, connclient)
	redisCleanUpActivity := syncrediscleanup_activity.New(workflowEnv.redisclient)

	env.RegisterWorkflow(datasync_workflow.Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(redisCleanUpActivity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	env.RegisterActivity(jobhookTimingActivity.RunJobHooksByTiming)
	env.RegisterActivity(posttableSyncActivity.RunPostTableSync)
	env.SetTestTimeout(600 * time.Second)

	workflowEnv.TestEnv = env

	return workflowEnv
}

// ExecuteTestDataSyncWorkflow starts the test workflow with the given job ID
func (w *TestWorkflowEnv) ExecuteTestDataSyncWorkflow(jobId string) {
	w.TestEnv.SetStartWorkflowOptions(client.StartWorkflowOptions{ID: jobId})
	w.TestEnv.ExecuteWorkflow(datasync_workflow.Workflow, &datasync_workflow.WorkflowRequest{JobId: jobId})
}

// RequireActivitiesCompletedSuccessfully verifies all activities completed without errors
// NOTE: this should be called before ExecuteTestDataSyncWorkflow
func (w *TestWorkflowEnv) RequireActivitiesCompletedSuccessfully(t testing.TB) {
	w.TestEnv.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result converter.EncodedValue, err error) {
		require.NoError(t, err, "Activity %s failed", activityInfo.ActivityType.Name)
		if activityInfo.ActivityType.Name == "RunPostTableSync" && result.HasValue() {
			var postTableSyncResp posttablesync_activity.RunPostTableSyncResponse
			decodeErr := result.Get(&postTableSyncResp)
			require.NoError(t, decodeErr, "Failed to decode result for activity %s", activityInfo.ActivityType.Name)
			require.Emptyf(t, postTableSyncResp.Errors, "Post table sync activity returned errors: %v", formatPostTableSyncErrors(postTableSyncResp.Errors))
		}
	})
}

func formatPostTableSyncErrors(errors []*posttablesync_activity.PostTableSyncError) []string {
	formatted := []string{}
	for _, err := range errors {
		for _, e := range err.Errors {
			formatted = append(formatted, fmt.Sprintf("statement: %s  error: %s", e.Statement, e.Error))
		}
	}
	return formatted
}

// TestDatabaseManagers holds managers for supported connection types
type TestDatabaseManagers struct {
	SqlConnManager   *connectionmanager.ConnectionManager[neosync_benthos_sql.SqlDbtx]
	SqlManager       *sql_manager.SqlManager
	MongoConnManager *connectionmanager.ConnectionManager[neosync_benthos_mongodb.MongoClient]
}

// NewTestDatabaseManagers creates and configures database connection managers for testing
func NewTestDatabaseManagers(t testing.TB) *TestDatabaseManagers {
	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithReaperPoll(10*time.Second))
	go sqlconnmanager.Reaper(testutil.GetConcurrentTestLogger(t))
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider())
	go mongoconnmanager.Reaper(testutil.GetConcurrentTestLogger(t))

	t.Cleanup(func() {
		sqlconnmanager.Shutdown(testutil.GetConcurrentTestLogger(t))
		mongoconnmanager.Shutdown(testutil.GetConcurrentTestLogger(t))
	})

	sqlmanager := sql_manager.NewSqlManager(
		sql_manager.WithConnectionManager(sqlconnmanager),
	)
	return &TestDatabaseManagers{
		SqlConnManager:   sqlconnmanager,
		SqlManager:       sqlmanager,
		MongoConnManager: mongoconnmanager,
	}
}
