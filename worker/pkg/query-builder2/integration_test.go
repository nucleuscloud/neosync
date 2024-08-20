package querybuilder2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite

	pgpool  *pgxpool.Pool
	querier pg_queries.Querier

	setupSql    string
	teardownSql string

	ctx context.Context

	pgcontainer *testpg.PostgresContainer

	schema string

	wg sync.WaitGroup
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.schema = "genbenthosconfigs_querybuilder"

	pgcontainer, err := testpg.Run(
		s.ctx,
		"postgres:15",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	s.pgcontainer = pgcontainer
	connstr, err := pgcontainer.ConnectionString(s.ctx)
	if err != nil {
		panic(err)
	}

	setupSql, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		panic(err)
	}
	s.setupSql = string(setupSql)

	teardownSql, err := os.ReadFile("./testdata/teardown.sql")
	if err != nil {
		panic(err)
	}
	s.teardownSql = string(teardownSql)

	pool, err := pgxpool.New(s.ctx, connstr)
	if err != nil {
		panic(err)
	}
	s.pgpool = pool
	s.querier = pg_queries.New()
	_, err = s.pgpool.Exec(s.ctx, s.setupSql)
	if err != nil {
		panic(err)
	}
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
}

func (s *IntegrationTestSuite) TearDownTest() {
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("waiting for all tests to complete")
	s.wg.Wait()
	s.T().Log("all tests completed, tearing down suite")
	_, err := s.pgpool.Exec(s.ctx, s.teardownSql)
	if err != nil {
		panic(err)
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
