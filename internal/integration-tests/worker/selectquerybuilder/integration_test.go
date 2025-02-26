package selectquerybuilder

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/suite"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
)

type mssqlTest struct {
	pool          *sql.DB
	testcontainer *testmssql.MSSQLServerContainer
}

type IntegrationTestSuite struct {
	suite.Suite

	querier pg_queries.Querier

	setupSql    string
	teardownSql string

	ctx context.Context

	pgcontainer *tcpostgres.PostgresTestContainer

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
		return nil, fmt.Errorf("unable to run mssql server container: %w", err)
	}
	// disabling tls encryption here to fix flaky startup and we also aren't concerned about TLS for a local container
	connstr, err := mssqlcontainer.ConnectionString(s.ctx, "encrypt=disable")
	if err != nil {
		return nil, fmt.Errorf("unable to get mssql connection string: %w", err)
	}
	setupSql, err := os.ReadFile("./testdata/mssql/setup.sql")
	if err != nil {
		return nil, fmt.Errorf("unable to read mssql setup file: %w", err)
	}

	conn, err := sql.Open(sqlmanager_shared.MssqlDriver, connstr)
	if err != nil {
		return nil, fmt.Errorf("unable to open mssql driver: %w", err)
	}

	_, err = conn.ExecContext(s.ctx, string(setupSql))
	if err != nil {
		return nil, fmt.Errorf("unable to exec mssql setup sql: %w", err)
	}

	return &mssqlTest{
		testcontainer: mssqlcontainer,
		pool:          conn,
	}, nil
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.schema = "genbenthosconfigs_querybuilder"

	pgcontainer, err := tcpostgres.NewPostgresTestContainer(s.ctx)
	if err != nil {
		panic(err)
	}
	s.pgcontainer = pgcontainer

	s.setupSql = "testdata/postgres/setup.sql"
	s.teardownSql = "testdata/postgres/teardown.sql"

	s.querier = pg_queries.New()

	mssqlTest, err := s.SetupMssql()
	if err != nil {
		panic(err)
	}
	s.mssql = mssqlTest
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	err := s.pgcontainer.RunSqlFiles(s.ctx, nil, []string{s.setupSql})
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	err := s.pgcontainer.RunSqlFiles(s.ctx, nil, []string{s.teardownSql})
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pgcontainer != nil {
		err := s.pgcontainer.TearDown(s.ctx)
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
	ok := testutil.ShouldRunWorkerIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}
