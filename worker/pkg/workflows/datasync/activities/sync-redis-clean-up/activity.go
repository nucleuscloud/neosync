package syncrediscleanup_activity

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	neosync_redis "github.com/nucleuscloud/neosync/worker/internal/redis"
	redis "github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type DeleteRedisHashRequest struct {
	JobId   string
	HashKey string
}

type DeleteRedisHashResponse struct {
}

func DeleteRedisHash(
	ctx context.Context,
	req *DeleteRedisHashRequest,
) (*DeleteRedisHashResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	logger := log.With(
		activity.GetLogger(ctx),
		"jobId", req.JobId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
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

	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slogger = slogger.With(
		"jobId", req.JobId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
		"RedisHashKey", req.HashKey,
	)

	// todo: this should be factored out of here and live on the activity itself
	redisClient, err := neosync_redis.GetRedisClient()
	if err != nil {
		return nil, err
	}
	slogger.Debug("redis client created")

	err = deleteRedisHashByKey(ctx, redisClient, req.HashKey)
	if err != nil {
		return nil, err
	}

	return &DeleteRedisHashResponse{}, nil
}

func deleteRedisHashByKey(ctx context.Context, client redis.UniversalClient, key string) error {
	err := client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete redis hash: %w", err)
	}
	return nil
}
