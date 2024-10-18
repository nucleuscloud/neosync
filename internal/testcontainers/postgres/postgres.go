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
}

// Creates and starts a PostgreSQL test container and sets up the connection.
func (p *PostgresTestContainer) Setup(ctx context.Context) (*PostgresTestContainer, error) {
	pgContainer, err := postgres.Run(
		ctx,
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

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(ctx, connStr)
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
func (p *PostgresTestContainer) TearDown(ctx context.Context) error {
	if p.Pool != nil {
		p.Pool.Close()
	}

	if p.TestContainer != nil {
		err := p.TestContainer.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("failed to terminate postgres container: %w", err)
		}
	}

	return nil
}

// Executes SQL files within the test container
func (p *PostgresTestContainer) RunSqlFiles(ctx context.Context, testFolder string, files []string) error {
	for _, file := range files {
		sqlStr, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			return err
		}
		_, err = p.Pool.Exec(ctx, string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec sql when running postgres sql files: %w", err)
		}
	}
	return nil
}
