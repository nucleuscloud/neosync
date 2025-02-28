package selectquerybuilder

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/suite"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
)

type mssqlTest struct {
	pool          *sql.DB
	testcontainer *testmssql.MSSQLServerContainer
}

type postgresTest struct {
	pgcontainer       *tcpostgres.PostgresTestContainer
	tableConstraints  *sqlmanager_shared.TableConstraints
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow
	querier           pg_queries.Querier

	teardownSql string
}

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	schema string

	postgres *postgresTest
	mssql    *mssqlTest
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
	err = pgcontainer.RunSqlFiles(s.ctx, nil, []string{"testdata/postgres/setup.sql"})
	if err != nil {
		panic(err)
	}

	sourceConn := &mgmtv1alpha1.Connection{
		Id: "test",
		ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
						Url: pgcontainer.URL,
					},
				},
			},
		},
	}

	logger := testutil.GetTestLogger(s.T())

	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mssqlquerier := mssql_queries.New()

	sqlmanager := sql_manager.NewSqlManager(
		sql_manager.WithPostgresQuerier(pgquerier),
		sql_manager.WithMysqlQuerier(mysqlquerier),
		sql_manager.WithMssqlQuerier(mssqlquerier),
		sql_manager.WithConnectionManagerOpts(connectionmanager.WithCloseOnRelease()),
	)
	db, err := sqlmanager.NewSqlConnection(s.ctx, connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup("test")), sourceConn, logger)
	if err != nil {
		s.T().Fatalf("unable to create sql connection: %s", err)
	}
	defer db.Db().Close()

	constraints, err := db.Db().GetTableConstraintsBySchema(s.ctx, []string{s.schema})
	if err != nil {
		s.T().Fatalf("unable to get table constraints: %s", err)
	}
	groupedColumnInfo, err := db.Db().GetSchemaColumnMap(s.ctx)
	if err != nil {
		s.T().Fatalf("unable to get schema column map: %s", err)
	}

	pgTest := &postgresTest{
		pgcontainer:       pgcontainer,
		tableConstraints:  constraints,
		groupedColumnInfo: groupedColumnInfo,
		teardownSql:       "testdata/postgres/teardown.sql",
		querier:           pgquerier,
	}
	s.postgres = pgTest

	mssqlTest, err := s.SetupMssql()
	if err != nil {
		panic(err)
	}
	s.mssql = mssqlTest

}

func (s *IntegrationTestSuite) TearDownSuite() {
	err := s.postgres.pgcontainer.RunSqlFiles(s.ctx, nil, []string{s.postgres.teardownSql})
	if err != nil {
		panic(err)
	}
	if s.postgres != nil && s.postgres.pgcontainer != nil {
		err := s.postgres.pgcontainer.TearDown(s.ctx)
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
