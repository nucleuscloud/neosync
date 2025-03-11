package schemainit_workflow_register

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	initschema_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/schemainit/activities/init-schema"
	schemainit_workflow "github.com/nucleuscloud/neosync/worker/pkg/workflows/schemainit/workflow"
)

type Worker interface {
	RegisterWorkflow(workflow any)
	RegisterActivity(activity any)
}

func Register(
	w Worker,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	sqlmanager *sql_manager.SqlManager,
	eelicense license.EEInterface,
) {
	runSqlInitTableStatements := initschema_activity.New(jobclient, connclient, sqlmanager, eelicense)
	siWf := schemainit_workflow.New()
	w.RegisterWorkflow(siWf.SchemaInit)
	w.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
}
