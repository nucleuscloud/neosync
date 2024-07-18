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
	pool      *pgxpool.Pool
	container *testpg.PostgresContainer
	url       string
}
type PostgresTest struct {
	pool          *pgxpool.Pool
	testcontainer *testpg.PostgresContainer

	source *PostgresTestContainer
	target *PostgresTestContainer

	databases []string
}

type MysqlTestContainer struct {
	pool      *sql.DB
	container *testmysql.MySQLContainer
	url       string
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
		testcontainer: pgcontainer,
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
	s.postgres.pool = pool

	s.T().Logf("creating databases. %+v \n", s.postgres.databases)
	for _, db := range s.postgres.databases {
		_, err = s.postgres.pool.Exec(s.ctx, fmt.Sprintf("CREATE DATABASE %s;", db))
		if err != nil {
			panic(err)
		}
	}

	srcUrl, err := getDbPgUrl(connstr, "datasync_source", "disable")
	if err != nil {
		panic(err)
	}
	s.postgres.source = &PostgresTestContainer{
		url: srcUrl,
	}
	sourceConn, err := pgxpool.New(s.ctx, s.postgres.source.url)
	if err != nil {
		panic(err)
	}
	s.postgres.source.pool = sourceConn

	targetUrl, err := getDbPgUrl(connstr, "datasync_target", "disable")
	if err != nil {
		panic(err)
	}
	s.postgres.target = &PostgresTestContainer{
		url: targetUrl,
	}
	targetConn, err := pgxpool.New(s.ctx, s.postgres.target.url)
	if err != nil {
		panic(err)
	}
	s.postgres.target.pool = targetConn
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
	containerPort, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return nil, err
	}
	containerHost, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	connUrl := fmt.Sprintf("mysql://%s:%s@%s:%s/%s?multiStatements=true", username, password, containerHost, containerPort.Port(), database)
	return &MysqlTestContainer{
		pool:      pool,
		url:       connUrl,
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
		sqlStr, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			panic(err)
		}
		_, err = pool.Exec(s.ctx, string(sqlStr))
		if err != nil {
			panic(err)
		}
	}
}

func (s *IntegrationTestSuite) RunMysqlSqlFiles(pool *sql.DB, testFolder string, files []string) {
	s.T().Logf("running mysql sql file. folder: %s \n", testFolder)
	for _, file := range files {
		sqlStr, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			panic(err)
		}

		_, err = pool.ExecContext(s.ctx, string(sqlStr))
		if err != nil {
			panic(err)
		}
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	// postgres
	for _, db := range s.postgres.databases {
		_, err := s.postgres.pool.Exec(s.ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", db))
		if err != nil {
			panic(err)
		}
	}
	if s.postgres.source.pool != nil {
		s.postgres.source.pool.Close()
	}
	if s.postgres.target.pool != nil {
		s.postgres.target.pool.Close()
	}
	if s.postgres.pool != nil {
		s.postgres.pool.Close()
	}
	if s.postgres.testcontainer != nil {
		err := s.postgres.testcontainer.Terminate(s.ctx)
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
