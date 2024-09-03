package querybuilder2

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/microsoft/go-mssqldb"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mssqlTest struct {
	pool          *sql.DB
	testcontainer *testmssql.MSSQLServerContainer
}

type IntegrationTestSuite struct {
	suite.Suite

	pgpool  *pgxpool.Pool
	querier pg_queries.Querier

	setupSql    string
	teardownSql string

	ctx context.Context

	pgcontainer *testpg.PostgresContainer

	schema string

	mssql *mssqlTest
}

func (s *IntegrationTestSuite) SetupMssql() (*mssqlTest, error) {
	mssqlcontainer, err := testmssql.Run(s.ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		testmssql.WithAcceptEULA(),
		testmssql.WithPassword("mssqlPASSword1"),
	)
	if err != nil {
		return nil, err
	}
	connstr, err := mssqlcontainer.ConnectionString(s.ctx)
	if err != nil {
		return nil, err
	}
	setupSql, err := os.ReadFile("./testdata/mssql/setup.sql")
	if err != nil {
		panic(err)
	}

	conn, err := sql.Open(sqlmanager_shared.MssqlDriver, connstr)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(s.ctx, string(setupSql))
	if err != nil {
		return nil, err
	}

	return &mssqlTest{
		testcontainer: mssqlcontainer,
		pool:          conn,
	}, nil
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

	setupSql, err := os.ReadFile("./testdata/postgres/setup.sql")
	if err != nil {
		panic(err)
	}
	s.setupSql = string(setupSql)

	teardownSql, err := os.ReadFile("./testdata/postgres/teardown.sql")
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

	mssqlTest, err := s.SetupMssql()
	if err != nil {
		panic(err)
	}
	s.mssql = mssqlTest
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
	if s.mssql != nil {
		if s.mssql.pool != nil {
			s.mssql.pool.Close()
		}
		if s.mssql.testcontainer != nil {
			err := s.mssql.testcontainer.Terminate(s.ctx)
			if err != nil {
				panic(err)
			}
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
