package datasync_workflow

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite

	pgpool       *pgxpool.Pool
	sourcePgPool *pgxpool.Pool
	targetPgPool *pgxpool.Pool

	sourceDsn string
	targetDsn string

	querier pg_queries.Querier

	ctx context.Context

	pgcontainer *testpg.PostgresContainer
	databases   []string

	redisUrl       string
	rediscontainer *redis.RedisContainer
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	testDbUrl := os.Getenv("TEST_DB_URL")
	dburl := fmt.Sprintf("%s?sslmode=disable", testDbUrl)
	if testDbUrl == "" {
		pgcontainer, err := testpg.RunContainer(s.ctx,
			testcontainers.WithImage("postgres:15"),
			postgres.WithDatabase("datasync"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).WithStartupTimeout(5*time.Second),
			),
		)
		if err != nil {
			panic(err)
		}
		s.pgcontainer = pgcontainer
		connstr, err := pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
		if err != nil {
			panic(err)
		}
		dburl = connstr
	}

	s.databases = []string{"datasync_source", "datasync_target"}
	pool, err := pgxpool.New(s.ctx, dburl)
	if err != nil {
		panic(err)
	}

	s.pgpool = pool
	for _, db := range s.databases {
		_, err = s.pgpool.Exec(s.ctx, fmt.Sprintf("CREATE DATABASE %s;", db))
		if err != nil {
			panic(err)
		}
	}

	s.sourceDsn = strings.Replace(dburl, "datasync", "datasync_source", 1)
	sourceConn, err := pgxpool.New(s.ctx, s.sourceDsn)
	if err != nil {
		panic(err)
	}
	s.sourcePgPool = sourceConn

	s.targetDsn = strings.Replace(dburl, "datasync", "datasync_target", 1)
	targetConn, err := pgxpool.New(s.ctx, s.targetDsn)
	if err != nil {
		panic(err)
	}
	s.targetPgPool = targetConn

	s.querier = pg_queries.New()

	// redis
	redisUrl := os.Getenv("TEST_REDIS_URL")
	if redisUrl == "" {
		redisContainer, err := redis.RunContainer(s.ctx,
			testcontainers.WithImage("docker.io/redis:7"),
			redis.WithSnapshotting(10, 1),
			redis.WithLogLevel(redis.LogLevelVerbose),
		)
		if err != nil {
			panic(err)
		}
		s.rediscontainer = redisContainer
		s.redisUrl, err = redisContainer.ConnectionString(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func (s *IntegrationTestSuite) SetupTestByFolder(testFolder string) {
	s.T().Logf("setting up test. folder: %s \n", testFolder)
	setupSourceSql, err := os.ReadFile(fmt.Sprintf("./testdata/%s/source-setup.sql", testFolder))
	if err != nil {
		panic(err)
	}
	_, err = s.sourcePgPool.Exec(s.ctx, string(setupSourceSql))
	if err != nil {
		panic(err)
	}
	setupTargetSql, err := os.ReadFile(fmt.Sprintf("./testdata/%s/target-setup.sql", testFolder))
	if err != nil {
		panic(err)
	}
	_, err = s.targetPgPool.Exec(s.ctx, string(setupTargetSql))
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTestByFolder(testFolder string) {
	s.T().Logf("tearing down test. folder: %s \n", testFolder)
	teardownSql, err := os.ReadFile(fmt.Sprintf("./testdata/%s/teardown.sql", testFolder))
	if err != nil {
		panic(err)
	}
	_, err = s.targetPgPool.Exec(s.ctx, string(teardownSql))
	if err != nil {
		panic(err)
	}
	_, err = s.sourcePgPool.Exec(s.ctx, string(teardownSql))
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	for _, db := range s.databases {
		_, err := s.pgpool.Exec(s.ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", db))
		if err != nil {
			panic(err)
		}
	}
	if s.sourcePgPool != nil {
		s.sourcePgPool.Close()
	}
	if s.targetPgPool != nil {
		s.targetPgPool.Close()
	}
	if s.pgpool != nil {
		s.pgpool.Close()
	}
	if s.pgcontainer != nil {
		err := s.pgcontainer.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
	if s.rediscontainer != nil {
		if err := s.rediscontainer.Terminate(s.ctx); err != nil {
			panic(err)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}
