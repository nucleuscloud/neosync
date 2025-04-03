package testcontainers_redis

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Holds the Redis test container and connection string.
type RedisTestContainer struct {
	URL           string
	TestContainer *testredis.RedisContainer
}

// Option is a functional option for configuring the Redis Test Container
type Option func(*RedisTestContainer)

// NewRedisTestContainer initializes a new Redis Test Container with functional options
func NewRedisTestContainer(ctx context.Context, opts ...Option) (*RedisTestContainer, error) {
	r := &RedisTestContainer{}
	for _, opt := range opts {
		opt(r)
	}
	return r.Setup(ctx)
}

// Creates and starts a Redis test container
func (r *RedisTestContainer) Setup(ctx context.Context) (*RedisTestContainer, error) {
	redisContainer, err := testredis.Run(
		ctx,
		"docker.io/redis:7",
		testredis.WithSnapshotting(10, 1),
		testredis.WithLogLevel(testredis.LogLevelVerbose),
		testcontainers.WithWaitStrategy(
			wait.ForLog("* Ready to accept connections"),
			wait.ForExposedPort(),
		),
	)
	if err != nil {
		return nil, err
	}
	redisUrl, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &RedisTestContainer{
		URL:           redisUrl,
		TestContainer: redisContainer,
	}, nil
}

// Terminates the container.
func (r *RedisTestContainer) TearDown(ctx context.Context) error {
	if r.TestContainer != nil {
		if err := r.TestContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}
	return nil
}
