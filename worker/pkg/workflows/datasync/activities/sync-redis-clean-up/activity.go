package syncrediscleanup_activity

import (
	"context"
	"fmt"
	"time"

	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	redis "github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	redisclient redis.UniversalClient
}

func New(
	redisclient redis.UniversalClient,
) *Activity {
	return &Activity{
		redisclient: redisclient,
	}
}

type DeleteRedisHashRequest struct {
	JobId   string
	HashKey string
}

type DeleteRedisHashResponse struct {
}

func (a *Activity) DeleteRedisHash(
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

	slogger := temporallogger.NewSlogger(logger)
	slogger = slogger.With(
		"jobId", req.JobId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
		"RedisHashKey", req.HashKey,
	)

	if a.redisclient == nil {
		return nil, fmt.Errorf("missing redis client. this operation requires redis.")
	}
	slogger.Debug("redis client provided")

	err := deleteRedisHashByKey(ctx, a.redisclient, req.HashKey)
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
