package sqlmanager

import (
	context "context"
	"fmt"
	slog "log/slog"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/suite"

	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"

	_ "github.com/microsoft/go-mssqldb"
)

type MssqlIntegrationTestSuite struct {
	suite.Suite

	container *testmssql.MSSQLServerContainer

	ctx context.Context

	sqlmanager SqlManagerClient

	// mysql cfg
	conncfg *mgmtv1alpha1.MssqlConnectionConfig
	// mgmt connection
	mgmtconn *mgmtv1alpha1.Connection
}

func (s *MssqlIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := testmssql.Run(s.ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		testmssql.WithAcceptEULA(),
	)
	if err != nil {
		panic(fmt.Errorf("unable to start container: %w", err))
	}
	connstr, err := container.ConnectionString(s.ctx)
	if err != nil {
		panic(fmt.Errorf("unable to get mssql connection str: %w", err))
	}

	queryvals := url.Values{}
	queryvals.Add("database", "master")

	connstr += queryvals.Encode()

	s.conncfg = &mgmtv1alpha1.MssqlConnectionConfig{
		ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
			Url: connstr,
		},
	}
	s.mgmtconn = &mgmtv1alpha1.Connection{
		Id: uuid.NewString(),
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
				MssqlConfig: s.conncfg,
			},
		},
	}
}

func (s *MssqlIntegrationTestSuite) SetupTest() {
	s.sqlmanager = NewSqlManager(nil, nil, nil, nil, &sync.Map{}, mssql_queries.New(), &sqlconnect.SqlOpenConnector{})
}

func (s *MssqlIntegrationTestSuite) TearDownTest() {
	if s.sqlmanager != nil {
		s.sqlmanager = nil
	}
}

func (s *MssqlIntegrationTestSuite) TearDownSuite() {
	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestMssqlIntegrationTestSuite(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(MssqlIntegrationTestSuite))
}

func (s *MssqlIntegrationTestSuite) Test_NewPooledSqlDb() {
	t := s.T()

	conn, err := s.sqlmanager.NewPooledSqlDb(s.ctx, slog.Default(), s.mgmtconn)
	requireNoConnErr(t, conn, err)
	requireValidDatabase(t, s.ctx, conn, "sqlserver", "SELECT 1")
	conn.Db.Close()
}

func (s *MssqlIntegrationTestSuite) Test_NewSqlDb() {
	t := s.T()

	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDb(s.ctx, slog.Default(), s.mgmtconn, &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "sqlserver", "SELECT 1")
	conn.Db.Close()
}

func (s *MssqlIntegrationTestSuite) Test_NewSqlDbFromUrl() {
	t := s.T()
	conn, err := s.sqlmanager.NewSqlDbFromUrl(s.ctx, "sqlserver", s.conncfg.GetUrl())
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "sqlserver", "SELECT 1")
	conn.Db.Close()
}

func (s *MssqlIntegrationTestSuite) Test_NewSqlDbFromConnectionConfig() {
	t := s.T()
	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDbFromConnectionConfig(s.ctx, slog.Default(), s.mgmtconn.GetConnectionConfig(), &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "sqlserver", "SELECT 1")
	conn.Db.Close()
}
