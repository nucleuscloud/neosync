package sqlmanager

import (
	context "context"
	"fmt"
	slog "log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MysqlIntegrationTestSuite struct {
	suite.Suite

	mysqlcontainer *testmysql.MySQLContainer

	ctx context.Context

	sqlmanager SqlManagerClient

	// mysql cfg
	conncfg *mgmtv1alpha1.MysqlConnectionConfig
	// mgmt connection
	mgmtconn *mgmtv1alpha1.Connection

	// dsn format of connection url
	dsn string
}

func (s *MysqlIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	dbname := "testdb"
	user := "root"
	pass := "test-password"

	container, err := testmysql.Run(s.ctx,
		"mysql:8.0.36",
		testmysql.WithDatabase(dbname),
		testmysql.WithUsername(user),
		testmysql.WithPassword(pass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server").
				WithOccurrence(1).WithStartupTimeout(20*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}

	connstr, err := container.ConnectionString(s.ctx, "multiStatements=true")
	if err != nil {
		panic(err)
	}
	s.dsn = connstr

	s.mysqlcontainer = container

	containerPort, err := container.MappedPort(s.ctx, "3306/tcp")
	if err != nil {
		panic(err)
	}
	containerHost, err := container.Host(s.ctx)
	if err != nil {
		panic(err)
	}

	connUrl := fmt.Sprintf("mysql://%s:%s@%s:%s/%s?multiStatements=true", user, pass, containerHost, containerPort.Port(), dbname)

	s.conncfg = &mgmtv1alpha1.MysqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
			Url: connUrl,
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
		err := s.mysqlcontainer.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestMysqlIntegrationTestSuite(t *testing.T) {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
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
	conn, err := s.sqlmanager.NewSqlDbFromUrl(s.ctx, "mysql", s.dsn) // NewSqlDbFromUrl requires dsn format
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
