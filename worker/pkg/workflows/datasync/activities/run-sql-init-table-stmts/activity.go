package runsqlinittablestmts_activity

import (
	"context"
	"time"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
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

	pgpoolmap := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.New()
	mysqlpoolmap := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.New()

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpClient,
		neosyncUrl,
	)

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpClient,
		neosyncUrl,
	)
	builder := newInitStatementBuilder(
		pgpoolmap,
		pgquerier,
		mysqlpoolmap,
		mysqlquerier,
		jobclient,
		connclient,
		&sqlconnect.SqlOpenConnector{},
	)
	slogger := logger_utils.NewJsonSLogger().With(
		"jobId", req.JobId,
		"WorkflowID", req.WorkflowId,
	)
	return builder.RunSqlInitTableStatements(ctx, req, slogger)
}
