package runsqlinittablestmts_activity

import (
	"context"
	"sync"
	"time"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	logger_utils "github.com/nucleuscloud/neosync/worker/internal/logger"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type RunSqlInitTableStatementsRequest struct {
	JobId      string
	WorkflowId string
}

type RunSqlInitTableStatementsResponse struct {
}

func RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
) (*RunSqlInitTableStatementsResponse, error) {
	logger := log.With(
		activity.GetLogger(ctx),
		"jobId", req.JobId,
		"WorkflowID", req.WorkflowId,
		// "RunID", wfmetadata.RunId,
	)
	_ = logger

	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	neosyncUrl := shared.GetNeosyncUrl()
	httpClient := shared.GetNeosyncHttpClient()

	pgpoolmap := &sync.Map{}
	pgquerier := pg_queries.New()
	mysqlpoolmap := &sync.Map{}
	mysqlquerier := mysql_queries.New()
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, &sqlconnect.SqlOpenConnector{})

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpClient,
		neosyncUrl,
	)

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpClient,
		neosyncUrl,
	)
	builder := newInitStatementBuilder(
		sqlmanager,
		jobclient,
		connclient,
	)
	slogger := logger_utils.NewJsonSLogger().With(
		"jobId", req.JobId,
		"WorkflowID", req.WorkflowId,
	)
	return builder.RunSqlInitTableStatements(ctx, req, slogger)
}
