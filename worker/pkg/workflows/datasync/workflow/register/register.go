package datasync_workflow_register

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	posttablesync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/post-table-sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	datasync_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow"
	"github.com/redis/go-redis/v9"
)

type Worker interface {
	RegisterWorkflow(workflow any)
	RegisterActivity(activity any)
}

func Register(
	w Worker,
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanager *sql_manager.SqlManager,
	redisconfig *neosync_redis.RedisConfig,
	eelicense license.EEInterface,
	redisclient redis.UniversalClient,
	isOtelEnabled bool,
	pageLimit int,
	postgresSchemaDrift bool,
) {
	genbenthosActivity := genbenthosconfigs_activity.New(
		jobclient,
		connclient,
		transformerclient,
		sqlmanager,
		redisconfig,
		isOtelEnabled,
		pageLimit,
	)

	retrieveActivityOpts := syncactivityopts_activity.New(jobclient, postgresSchemaDrift)
	accountStatusActivity := accountstatus_activity.New(userclient)
	runPostTableSyncActivity := posttablesync_activity.New(jobclient, sqlmanager, connclient)
	jobhookByTimingActivity := jobhooks_by_timing_activity.New(jobclient, connclient, sqlmanager, eelicense)
	redisCleanUpActivity := syncrediscleanup_activity.New(redisclient)

	wf := datasync_workflow.New(eelicense)

	w.RegisterWorkflow(wf.Workflow)
	w.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	w.RegisterActivity(redisCleanUpActivity.DeleteRedisHash)
	w.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	w.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	w.RegisterActivity(runPostTableSyncActivity.RunPostTableSync)
	w.RegisterActivity(jobhookByTimingActivity.RunJobHooksByTiming)
}
