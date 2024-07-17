package datasync_workflow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresTestContainer struct {
	pool          *pgxpool.Pool
	testcontainer *testpg.PostgresContainer
}
type PostgresTest struct {
	container *PostgresTestContainer

	sourcePool *pgxpool.Pool
	targetPool *pgxpool.Pool

	sourceDsn string
	targetDsn string

	databases []string
}

type MysqlTestContainer struct {
	pool      *sql.DB
	container *testmysql.MySQLContainer
	dsn       string
	close     func()
}

type MysqlTest struct {
	source *MysqlTestContainer
	target *MysqlTestContainer
}

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	mysql    *MysqlTest
	postgres *PostgresTest

	redisUrl       string
	rediscontainer *redis.RedisContainer
}

func (s *IntegrationTestSuite) SetupPostgres() {
	pgcontainer, err := testpg.Run(
		s.ctx,
		"postgres:15",
		postgres.WithDatabase("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	s.postgres = &PostgresTest{
		container: &PostgresTestContainer{
			testcontainer: pgcontainer,
		},
	}
	connstr, err := pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	s.postgres.databases = []string{"datasync_source", "datasync_target"}
	pool, err := pgxpool.New(s.ctx, connstr)
	if err != nil {
		panic(err)
	}
	s.postgres.container.pool = pool

	s.T().Logf("creating databases. %+v \n", s.postgres.databases)
	for _, db := range s.postgres.databases {
		_, err = s.postgres.container.pool.Exec(s.ctx, fmt.Sprintf("CREATE DATABASE %s;", db))
		if err != nil {
			panic(err)
		}
	}

	srcUrl, err := getDbPgUrl(connstr, "datasync_source", "disable")
	if err != nil {
		panic(err)
	}
	s.postgres.sourceDsn = srcUrl
	sourceConn, err := pgxpool.New(s.ctx, s.postgres.sourceDsn)
	if err != nil {
		panic(err)
	}
	s.postgres.sourcePool = sourceConn

	targetUrl, err := getDbPgUrl(connstr, "datasync_target", "disable")
	if err != nil {
		panic(err)
	}
	s.postgres.targetDsn = targetUrl
	targetConn, err := pgxpool.New(s.ctx, s.postgres.targetDsn)
	if err != nil {
		panic(err)
	}
	s.postgres.targetPool = targetConn
}

func (s *IntegrationTestSuite) SetupMysql() {
	s.ctx = context.Background()

	s.mysql = &MysqlTest{}

	sourcecontainer, err := createMysqlTestContainer(s.ctx, "datasync", "root", "pass-source")
	if err != nil {
		panic(err)
	}
	s.mysql.source = sourcecontainer

	targetcontainer, err := createMysqlTestContainer(s.ctx, "datasync", "root", "pass-target")
	if err != nil {
		panic(err)
	}
	s.mysql.target = targetcontainer
}

func createMysqlTestContainer(
	ctx context.Context,
	database, username, password string,
) (*MysqlTestContainer, error) {
	container, err := testmysql.Run(ctx,
		"mysql:8.0.36",
		testmysql.WithDatabase(database),
		testmysql.WithUsername(username),
		testmysql.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server").
				WithOccurrence(1).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}
	connstr, err := container.ConnectionString(ctx, "multiStatements=true")
	if err != nil {
		panic(err)
	}
	pool, err := sql.Open(sqlmanager_shared.MysqlDriver, connstr)
	if err != nil {
		panic(err)
	}
	return &MysqlTestContainer{
		pool:      pool,
		dsn:       connstr,
		container: container,
		close: func() {
			if pool != nil {
				pool.Close()
			}
		},
	}, nil
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	s.SetupPostgres()
	s.SetupMysql()

	// redis
	redisContainer, err := redis.Run(
		s.ctx,
		"docker.io/redis:7",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		panic(err)
	}
	s.rediscontainer = redisContainer
	s.redisUrl, err = redisContainer.ConnectionString(s.ctx)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) RunPostgresSqlFiles(pool *pgxpool.Pool, testFolder string, files []string) {
	s.T().Logf("running postgres sql file. folder: %s \n", testFolder)
	for _, file := range files {
		sql, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			panic(err)
		}
		_, err = pool.Exec(s.ctx, string(sql))
		if err != nil {
			panic(err)
		}
	}
}

func (s *IntegrationTestSuite) RunMysqlSqlFiles(pool *sql.DB, testFolder string, files []string) {
	s.T().Logf("running mysql sql file. folder: %s \n", testFolder)
	for _, file := range files {
		sql, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			panic(err)
		}

		_, err = pool.ExecContext(s.ctx, string(sql))
		if err != nil {
			panic(err)
		}
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	// postgres
	for _, db := range s.postgres.databases {
		_, err := s.postgres.container.pool.Exec(s.ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", db))
		if err != nil {
			panic(err)
		}
	}
	if s.postgres.sourcePool != nil {
		s.postgres.sourcePool.Close()
	}
	if s.postgres.targetPool != nil {
		s.postgres.targetPool.Close()
	}
	if s.postgres.container.pool != nil {
		s.postgres.container.pool.Close()
	}
	if s.postgres.container.testcontainer != nil {
		err := s.postgres.container.testcontainer.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}

	// mysql
	s.mysql.source.close()
	s.mysql.target.close()
	if s.mysql.source.container != nil {
		err := s.mysql.source.container.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
	if s.mysql.target.container != nil {
		err := s.mysql.target.container.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}

	// redis
	if s.rediscontainer != nil {
		if err := s.rediscontainer.Terminate(s.ctx); err != nil {
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

func getDbPgUrl(dburl, database, sslmode string) (string, error) {
	u, err := url.Parse(dburl)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return "", fmt.Errorf("unable to parse postgres url [%s]: %w", urlErr.Op, urlErr.Err)
		}
		return "", fmt.Errorf("unable to parse postgres url: %w", err)
	}

	u.Path = database
	query := u.Query()
	query.Add("sslmode", sslmode)
	return u.String(), nil
}
