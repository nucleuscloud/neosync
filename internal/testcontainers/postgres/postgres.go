package testcontainers_postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Holds the PostgreSQL test container and connection pool.
type PostgresTestContainer struct {
	Pool          *pgxpool.Pool
	URL           string
	TestContainer *testpg.PostgresContainer
	ctx           context.Context
	database      string
	username      string
	password      string
}

// Option is a functional option for configuring the Postgres Test Container
type Option func(*PostgresTestContainer)

// NewPostgresTestContainer initializes a new Postgres Test Container with functional options
func NewPostgresTestContainer(ctx context.Context, opts ...Option) (*PostgresTestContainer, error) {
	p := &PostgresTestContainer{
		database: "testdb",
		username: "postrgres",
		password: "pass",
		ctx:      ctx,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p.Setup()
}

// Sets test container database
func WithDatabase(database string) Option {
	return func(a *PostgresTestContainer) {
		a.database = database
	}
}

// Sets test container database
func WithUsername(username string) Option {
	return func(a *PostgresTestContainer) {
		a.username = username
	}
}

// Sets test container database
func WithPassword(password string) Option {
	return func(a *PostgresTestContainer) {
		a.password = password
	}
}

// Creates and starts a PostgreSQL test container and sets up the connection.
func (p *PostgresTestContainer) Setup() (*PostgresTestContainer, error) {
	pgContainer, err := postgres.Run(
		p.ctx,
		"postgres:15",
		postgres.WithDatabase("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(20*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(p.ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(p.ctx, connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresTestContainer{
		Pool:          pool,
		URL:           connStr,
		TestContainer: pgContainer,
	}, nil
}

// Closes the connection pool and terminates the container.
func (p *PostgresTestContainer) TearDown() error {
	if p.Pool != nil {
		p.Pool.Close()
	}

	if p.TestContainer != nil {
		err := p.TestContainer.Terminate(p.ctx)
		if err != nil {
			return fmt.Errorf("failed to terminate postgres container: %w", err)
		}
	}

	return nil
}

// Executes SQL files within the test container
func (p *PostgresTestContainer) RunSqlFiles(testFolder string, files []string) error {
	for _, file := range files {
		sqlStr, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			return err
		}
		_, err = p.Pool.Exec(p.ctx, string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec sql when running postgres sql files: %w", err)
		}
	}
	return nil
}
