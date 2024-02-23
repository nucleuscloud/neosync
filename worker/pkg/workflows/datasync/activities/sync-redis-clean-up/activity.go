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
)

type DeleteRedisHashRequest struct {
	JobId      string
	WorkflowId string
	HashKey    string
}

type DeleteRedisHashResponse struct {
}

func DeleteRedisHash(
	ctx context.Context,
	req *DeleteRedisHashRequest,
) (*DeleteRedisHashResponse, error) {
	logger := activity.GetLogger(ctx)
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
		"WorkflowID", req.WorkflowId,
		"RedisHashKey", req.HashKey,
	)

	redisClient, err := neosync_redis.GetRedisClient()
	if err != nil {
		return nil, err
	}

	err = deleteRedisHashByKey(slogger, ctx, redisClient, req.HashKey)
	if err != nil {
		return nil, err
	}

	return &DeleteRedisHashResponse{}, nil
}

func deleteRedisHashByKey(logger *slog.Logger, ctx context.Context, client redis.UniversalClient, key string) error {
	err := client.Del(ctx, key).Err()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to delete redis hash: %v", err))
		return err
	}
	return nil
}
