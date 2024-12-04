package integrationtest

import (
	"testing"
	"time"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	"github.com/nucleuscloud/neosync/internal/testutil"
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
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

func ExecuteTestDataSyncWorkflow(
	t testing.TB,
	neosyncApi *tcneosyncapi.NeosyncApiTestClient,
	redisUrl *string,
	jobId string,
	validEELicense bool,
) *testsuite.TestWorkflowEnvironment {
	t.Helper()
	connclient := neosyncApi.UnauthdClients.Connections
	jobclient := neosyncApi.UnauthdClients.Jobs
	transformerclient := neosyncApi.UnauthdClients.Transformers
	userclient := neosyncApi.UnauthdClients.Users

	var redisconfig *shared.RedisConfig
	if redisUrl != nil && *redisUrl != "" {
		redisconfig = &shared.RedisConfig{
			Url:  *redisUrl,
			Kind: "simple",
			Tls: &shared.RedisTlsConfig{
				Enabled: false,
			},
		}
	}

	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithReaperPoll(10*time.Second))
	go sqlconnmanager.Reaper(testutil.GetTestLogger(t))
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider())
	go mongoconnmanager.Reaper(testutil.GetTestLogger(t))

	sqlmanager := sql_manager.NewSqlManager(
		sql_manager.WithConnectionManager(sqlconnmanager),
	)

	// temporal workflow
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetTestLogger(t)))
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
	syncActivity := sync_activity.New(connclient, jobclient, sqlconnmanager, mongoconnmanager, activityMeter, sync_activity.NewBenthosStreamManager())
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager, &testutil.FakeEELicense{IsValid: validEELicense})
	jobhookTimingActivity := jobhooks_by_timing_activity.New(jobclient, connclient, sqlmanager, &testutil.FakeEELicense{IsValid: validEELicense})
	accountStatusActivity := accountstatus_activity.New(userclient)
	posttableSyncActivity := posttablesync_activity.New(jobclient, sqlmanager, connclient)

	env.RegisterWorkflow(datasync_workflow.Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	env.RegisterActivity(jobhookTimingActivity.RunJobHooksByTiming)
	env.RegisterActivity(posttableSyncActivity.RunPostTableSync)
	env.SetTestTimeout(600 * time.Second) // increase the test timeout

	env.SetStartWorkflowOptions(client.StartWorkflowOptions{ID: jobId})
	env.ExecuteWorkflow(datasync_workflow.Workflow, &datasync_workflow.WorkflowRequest{JobId: jobId})
	return env
}
