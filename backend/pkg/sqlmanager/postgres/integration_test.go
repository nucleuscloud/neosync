package sqlmanager_postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/suite"
)

var (
	testdataFolder = "testdata"
)

type IntegrationTestSuite struct {
	suite.Suite

	db      *sql.DB
	querier pg_queries.Querier

	ctx context.Context

	pgcontainer *tcpostgres.PostgresTestContainer

	schema string
}

func (s *IntegrationTestSuite) buildTable(tableName string) string {
	return fmt.Sprintf("%s.%s", s.schema, tableName)
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.schema = "sqlmanagerpostgres@special"

	pgcontainer, err := tcpostgres.NewPostgresTestContainer(s.ctx)
	if err != nil {
		panic(err)
	}
	s.pgcontainer = pgcontainer

	db, err := sql.Open(sqlmanager_shared.PostgresDriver, s.pgcontainer.URL)
	if err != nil {
		panic(err)
	}
	s.db = db
	s.querier = pg_queries.New()
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	err := s.pgcontainer.RunSqlFiles(s.ctx, &testdataFolder, []string{"setup.sql"})
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	err := s.pgcontainer.RunSqlFiles(s.ctx, &testdataFolder, []string{"teardown.sql"})
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	err := s.pgcontainer.TearDown(s.ctx)
	if err != nil {
		panic(err)
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}
