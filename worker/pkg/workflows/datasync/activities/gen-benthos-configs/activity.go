package genbenthosconfigs_activity

import (
	"context"
	"log/slog"
	"os"
	"time"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type GenerateBenthosConfigsRequest struct {
	JobId      string
	WorkflowId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*BenthosConfigResponse
}

type BenthosRedisConfig struct {
	Key    string
	Table  string // schema.table
	Column string
}

type BenthosConfigResponse struct {
	Name        string
	DependsOn   []*tabledependency.DependsOn
	Config      *neosync_benthos.BenthosConfig
	TableSchema string
	TableName   string
	Columns     []string

	BenthosDsns []*shared.BenthosDsn
	RedisConfig []*BenthosRedisConfig

	primaryKeys    []string
	excludeColumns []string
	updateConfig   *tabledependency.RunConfig
}

func GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	wfmetadata *shared.WorkflowMetadata,
) (*GenerateBenthosConfigsResponse, error) {
	loggerKeyVals := []any{
		"jobId", req.JobId,
		"WorkflowID", wfmetadata.WorkflowId,
		"RunID", wfmetadata.RunId,
	}
	logger := log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
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

	redisConfig := shared.GetRedisConfig()
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

	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(
		httpClient,
		neosyncUrl,
	)

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpClient,
		neosyncUrl,
	)
	bbuilder := newBenthosBuilder(
		pgpoolmap,
		pgquerier,
		mysqlpoolmap,
		mysqlquerier,
		jobclient,
		connclient,
		transformerclient,
		&sqlconnect.SqlOpenConnector{},
		req.JobId,
		wfmetadata.RunId,
		redisConfig,
	)
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With(loggerKeyVals...)
	return bbuilder.GenerateBenthosConfigs(ctx, req, slogger)
}
