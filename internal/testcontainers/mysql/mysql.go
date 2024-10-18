package testcontainers_mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Holds the MySQL test container and connection pool.
type MysqlTestContainer struct {
	Pool          *sql.DB
	URL           string
	TestContainer *testmysql.MySQLContainer
	database      string
	password      string
	username      string
}

// Option is a functional option for configuring the Anonymizer
type Option func(*MysqlTestContainer)

// NewAnonymizer initializes a new Anonymizer with functional options
func NewMysqlTestContainer(ctx context.Context, opts ...Option) (*MysqlTestContainer, error) {
	m := &MysqlTestContainer{
		database: "testdb",
		username: "root",
		password: "pass",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m.setup(ctx)
}

// Sets test container database
func WithDatabase(database string) Option {
	return func(a *MysqlTestContainer) {
		a.database = database
	}
}

// Sets test container database
func WithUsername(username string) Option {
	return func(a *MysqlTestContainer) {
		a.username = username
	}
}

// Sets test container database
func WithPassword(password string) Option {
	return func(a *MysqlTestContainer) {
		a.password = password
	}
}

// Creates and starts a MySQL test container and sets up the connection.
func (m *MysqlTestContainer) setup(ctx context.Context) (*MysqlTestContainer, error) {
	mysqlContainer, err := mysql.Run(
		ctx,
		"mysql:8.0.36",
		mysql.WithDatabase(m.database),
		mysql.WithUsername(m.username),
		mysql.WithPassword(m.password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server").WithOccurrence(1).WithStartupTimeout(20*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := mysqlContainer.ConnectionString(ctx, "multiStatements=true&parseTime=true")
	if err != nil {
		return nil, err
	}

	pool, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	return &MysqlTestContainer{
		Pool:          pool,
		URL:           connStr,
		TestContainer: mysqlContainer,
	}, nil
}

// Closes the connection pool and terminates the container.
func (m *MysqlTestContainer) TearDown(ctx context.Context) error {
	if m.Pool != nil {
		m.Pool.Close()
	}

	if m.TestContainer != nil {
		err := m.TestContainer.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("failed to terminate MySQL container: %w", err)
		}
	}

	return nil
}

// Executes SQL files within the test container
func (m *MysqlTestContainer) RunSqlFiles(ctx context.Context, testFolder string, files []string) error {
	for _, file := range files {
		sqlStr, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			return err
		}
		_, err = m.Pool.ExecContext(ctx, string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec SQL when running MySQL SQL files: %w", err)
		}
	}
	return nil
}
