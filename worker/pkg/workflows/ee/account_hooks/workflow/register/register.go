package accounthook_workflow_register

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	execute_hook_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/activities/execute"
	hooks_by_event_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/activities/hooks-by-event"
	accounthook_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/ee/account_hooks/workflow"
)

type Worker interface {
	RegisterWorkflow(workflow any)
	RegisterActivity(activity any)
}

func Register(w Worker, accounthookclient mgmtv1alpha1connect.AccountHookServiceClient) {
	hooksByEventActivity := hooks_by_event_activity.New(accounthookclient)
	executeHookActivity := execute_hook_activity.New(accounthookclient)

	w.RegisterWorkflow(accounthook_workflow.ProcessAccountHook)
	w.RegisterActivity(hooksByEventActivity.GetAccountHooksByEvent)
	w.RegisterActivity(executeHookActivity.ExecuteAccountHook)
}
