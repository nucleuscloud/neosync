package sqlmanager_postgres

import (
	"context"
	"os"
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
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	dburl := os.Getenv("TEST_DB_URL")
	if dburl == "" {
		pgcontainer, err := testpg.RunContainer(s.ctx,
			testcontainers.WithImage("postgres:15"),
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
		dburl = connstr
	}

	setupSql, err := os.ReadFile("../../scripts/integration/setup.sql")
	if err != nil {
		panic(err)
	}
	s.setupSql = string(setupSql)

	teardownSql, err := os.ReadFile("../../scripts/integration/teardown.sql")
	if err != nil {
		panic(err)
	}
	s.teardownSql = string(teardownSql)

	pool, err := pgxpool.New(s.ctx, dburl)
	if err != nil {
		panic(err)
	}
	s.pgpool = pool
	s.querier = pg_queries.New()
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.pgpool.Exec(s.ctx, s.setupSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	_, err := s.pgpool.Exec(s.ctx, s.teardownSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
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
	// shouldRun := os.Getenv("NEO_INTEGRATION_TESTS")
	// if shouldRun != "1" {
	// 	return
	// }
	suite.Run(t, new(IntegrationTestSuite))
}
