package sqlmanager_mysql

import (
	"context"

	_ "github.com/go-sql-driver/mysql"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	querier    mysql_queries.Querier
	containers *tcmysql.MysqlTestSyncContainer
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := tcmysql.NewMysqlTestSyncContainer(s.ctx, []tcmysql.Option{}, []tcmysql.Option{})
	if err != nil {
		panic(err)
	}
	s.containers = container
	s.querier = mysql_queries.New()
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	err := s.containers.Target.RunSqlFiles(s.ctx, nil, []string{"testdata/init.sql"})
	if err != nil {
		panic(err)
	}
	err = s.containers.Source.RunSqlFiles(s.ctx, nil, []string{"testdata/setup.sql"})
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	err := s.containers.Source.RunSqlFiles(s.ctx, nil, []string{"testdata/teardown.sql"})
	if err != nil {
		panic(err)
	}
	err = s.containers.Target.RunSqlFiles(s.ctx, nil, []string{"testdata/teardown.sql"})
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.containers != nil {
		if s.containers.Source != nil {
			err := s.containers.Source.TearDown(s.ctx)
			if err != nil {
				panic(err)
			}
		}
		if s.containers.Target != nil {
			err := s.containers.Target.TearDown(s.ctx)
			if err != nil {
				panic(err)
			}
		}
	}
}

// func TestIntegrationTestSuite(t *testing.T) {
// 	ok := testutil.ShouldRunIntegrationTest()
// 	if !ok {
// 		return
// 	}
// 	suite.Run(t, new(IntegrationTestSuite))
// }
