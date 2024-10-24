package sqlmanager

import (
	context "context"
	slog "log/slog"
	"sync"
	"testing"

	"github.com/google/uuid"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/suite"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
)

type MysqlIntegrationTestSuite struct {
	suite.Suite

	mysqlcontainer *tcmysql.MysqlTestContainer

	ctx context.Context

	sqlmanager SqlManagerClient

	// mysql cfg
	conncfg *mgmtv1alpha1.MysqlConnectionConfig
	// mgmt connection
	mgmtconn *mgmtv1alpha1.Connection
}

func (s *MysqlIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := tcmysql.NewMysqlTestContainer(s.ctx)
	if err != nil {
		panic(err)
	}
	s.mysqlcontainer = container

	s.conncfg = &mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
			Url: container.URL,
		},
	}
	s.mgmtconn = &mgmtv1alpha1.Connection{
		Id: uuid.NewString(),
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
				MysqlConfig: s.conncfg,
			},
		},
	}
}

func (s *MysqlIntegrationTestSuite) SetupTest() {
	s.sqlmanager = NewSqlManager(nil, nil, &sync.Map{}, mysql_queries.New(), nil, nil, &sqlconnect.SqlOpenConnector{})
}

func (s *MysqlIntegrationTestSuite) TearDownTest() {
	if s.sqlmanager != nil {
		s.sqlmanager = nil
	}
}

func (s *MysqlIntegrationTestSuite) TearDownSuite() {
	if s.mysqlcontainer != nil {
		err := s.mysqlcontainer.TearDown(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestMysqlIntegrationTestSuite(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(MysqlIntegrationTestSuite))
}

func (s *MysqlIntegrationTestSuite) Test_NewPooledSqlDb() {
	t := s.T()

	conn, err := s.sqlmanager.NewPooledSqlDb(s.ctx, slog.Default(), s.mgmtconn)
	requireNoConnErr(t, conn, err)
	requireValidDatabase(t, s.ctx, conn, "mysql", "SELECT 1")
	conn.Db.Close()
}

func (s *MysqlIntegrationTestSuite) Test_NewSqlDb() {
	t := s.T()

	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDb(s.ctx, slog.Default(), s.mgmtconn, &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "mysql", "SELECT 1")
	conn.Db.Close()
}

func (s *MysqlIntegrationTestSuite) Test_NewSqlDbFromUrl() {
	t := s.T()
	conn, err := s.sqlmanager.NewSqlDbFromUrl(s.ctx, "mysql", s.mysqlcontainer.URL)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "mysql", "SELECT 1")
	conn.Db.Close()
}

func (s *MysqlIntegrationTestSuite) Test_NewSqlDbFromConnectionConfig() {
	t := s.T()
	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDbFromConnectionConfig(s.ctx, slog.Default(), s.mgmtconn.GetConnectionConfig(), &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "mysql", "SELECT 1")
	conn.Db.Close()
}
