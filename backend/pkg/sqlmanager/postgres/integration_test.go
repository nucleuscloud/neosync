package sqlmanager_postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite

	db      *sql.DB
	querier pg_queries.Querier

	setupSql    string
	teardownSql string

	ctx context.Context

	pgcontainer *testpg.PostgresContainer

	schema string
}

func (s *IntegrationTestSuite) buildTable(tableName string) string {
	return fmt.Sprintf("%s.%s", s.schema, tableName)
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.schema = "sqlmanagerpostgres@special"

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

	db, err := sql.Open("pgx", connstr)
	if err != nil {
		panic(err)
	}
	s.db = db
	s.querier = pg_queries.New()
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.db.ExecContext(s.ctx, s.setupSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	_, err := s.db.ExecContext(s.ctx, s.teardownSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
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
