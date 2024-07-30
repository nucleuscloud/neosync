package sqlmanager_mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite

	querier mysql_queries.Querier
	pool    mysql_queries.DBTX
	close   func()

	setupSql    string
	teardownSql string

	ctx context.Context

	mysqlcontainer *testmysql.MySQLContainer
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	mysqlcontainer, err := testmysql.Run(s.ctx,
		"mysql:8.0.36",
		testmysql.WithDatabase("foo"),
		testmysql.WithUsername("root"),
		testmysql.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server").
				WithOccurrence(1).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	s.mysqlcontainer = mysqlcontainer
	connstr, err := mysqlcontainer.ConnectionString(s.ctx, "multiStatements=true")
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

	pool, err := sql.Open(sqlmanager_shared.MysqlDriver, connstr)
	if err != nil {
		panic(err)
	}
	s.pool = pool
	s.querier = mysql_queries.New()
	s.close = func() {
		if pool != nil {
			pool.Close()
		}
	}
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.pool.ExecContext(s.ctx, s.setupSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	_, err := s.pool.ExecContext(s.ctx, s.teardownSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.close()
	}
	if s.mysqlcontainer != nil {
		err := s.mysqlcontainer.Terminate(s.ctx)
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

func (s *IntegrationTestSuite) buildTable(schema, tableName string) string {
	return fmt.Sprintf("%s.%s", schema, tableName)
}
