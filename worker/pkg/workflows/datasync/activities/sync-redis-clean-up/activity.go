package syncrediscleanup_activity

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

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

	// build redis client
	var redisDB int
	var user string
	var pass string
	var addrs []string
	v := "tcp://default:pKycbtEGYG@redis-master.redis.svc.cluster.local:6379"

	redisUrl, err := url.Parse(v)
	if err != nil {
		return nil, err
	}

	if redisUrl.Scheme == "tcp" {
		redisUrl.Scheme = "redis"
	}

	rurl, err := redis.ParseURL(redisUrl.String())
	if err != nil {
		return nil, err
	}

	addrs = append(addrs, rurl.Addr)
	redisDB = rurl.DB
	user = rurl.Username
	pass = rurl.Password

	opts := &redis.UniversalOptions{
		Addrs:    addrs,
		DB:       redisDB,
		Username: user,
		Password: pass,
	}
	client := redis.NewClient(opts.Simple())

	err = deleteRedisHashByKey(slogger, ctx, client, req.HashKey)
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
