package testcontainers_postgres

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

type PostgresTestSyncContainer struct {
	Source *PostgresTestContainer
	Target *PostgresTestContainer
}

func NewPostgresTestSyncContainer(ctx context.Context, sourceOpts, destOpts []Option) (*PostgresTestSyncContainer, error) {
	tc := &PostgresTestSyncContainer{}
	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		p, err := NewPostgresTestContainer(ctx, sourceOpts...)
		if err != nil {
			return err
		}
		tc.Source = p
		return nil
	})

	errgrp.Go(func() error {
		p, err := NewPostgresTestContainer(ctx, destOpts...)
		if err != nil {
			return err
		}
		tc.Target = p
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return tc, nil
}

func (p *PostgresTestSyncContainer) TearDown(ctx context.Context) error {
	if p.Source != nil {
		err := p.Source.TearDown(ctx)
		if err != nil {
			return err
		}
	}
	if p.Target != nil {
		err := p.Target.TearDown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type pgTestContainerConfig struct {
	database string
	username string
	password string
	useTls   bool
}

// Holds the PostgreSQL test container and connection pool.
type PostgresTestContainer struct {
	DB            *pgxpool.Pool
	URL           string
	TestContainer *testpg.PostgresContainer

	cfg *pgTestContainerConfig
}

// Option is a functional option for configuring the Postgres Test Container
type Option func(*pgTestContainerConfig)

// NewPostgresTestContainer initializes a new Postgres Test Container with functional options
func NewPostgresTestContainer(ctx context.Context, opts ...Option) (*PostgresTestContainer, error) {
	p := &pgTestContainerConfig{
		database: "testdb",
		username: "postgres",
		password: "pass",
		useTls:   false,
	}
	for _, opt := range opts {
		opt(p)
	}
	return setup(ctx, p)
}

// Sets test container database
func WithDatabase(database string) Option {
	return func(a *pgTestContainerConfig) {
		a.database = database
	}
}

// Sets test container database
func WithUsername(username string) Option {
	return func(a *pgTestContainerConfig) {
		a.username = username
	}
}

// Sets test container database
func WithPassword(password string) Option {
	return func(a *pgTestContainerConfig) {
		a.password = password
	}
}

func WithTls() Option {
	return func(mtc *pgTestContainerConfig) {
		mtc.useTls = true
	}
}

// Creates and starts a PostgreSQL test container and sets up the connection.
func setup(ctx context.Context, cfg *pgTestContainerConfig) (*PostgresTestContainer, error) {
	tcopts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(cfg.database),
		postgres.WithUsername(cfg.username),
		postgres.WithPassword(cfg.password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(20 * time.Second),
		),
	}
	if cfg.useTls {
		clientCertPaths, err := testutil.GetTlsCertificatePaths()
		if err != nil {
			return nil, err
		}
		tcopts = append(
			tcopts,
			testutil.WithCmd([]string{
				"postgres",
				"-c", "fsync=off",
				"-c", "ssl=on",
				"-c", "ssl_cert_file=/var/lib/postgresql/ssl/server.crt",
				"-c", "ssl_key_file=/var/lib/postgresql/ssl/server.key",
				"-c", "ssl_ca_file=/var/lib/postgresql/ssl/root.crt",
			}),
			testutil.WithFiles([]testcontainers.ContainerFile{
				{
					HostFilePath:      clientCertPaths.ServerCertPath,
					ContainerFilePath: "/var/lib/postgresql/ssl/server.crt",
					FileMode:          0644,
				},
				{
					HostFilePath:      clientCertPaths.ServerKeyPath,
					ContainerFilePath: "/var/lib/postgresql/ssl/server.key",
					FileMode:          0600,
				},
				{
					HostFilePath:      clientCertPaths.RootCertPath,
					ContainerFilePath: "/var/lib/postgresql/ssl/root.crt",
					FileMode:          0644,
				},
			}),
			testcontainers.WithStartupCommand(testcontainers.NewRawCommand([]string{
				"chown", "postgres:postgres", "/var/lib/postgresql/ssl/server.key",
			})),
		)
	}
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15",
		tcopts...,
	)
	if err != nil {
		return nil, err
	}

	connstrArgs := []string{}

	if cfg.useTls {
		connstrArgs = append(connstrArgs, "sslmode=verify-full")
	} else {
		connstrArgs = append(connstrArgs, "sslmode=disable")
	}

	connStr, err := pgContainer.ConnectionString(ctx, connstrArgs...)
	if err != nil {
		return nil, err
	}

	pgxCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	if cfg.useTls {
		serverHost, err := pgContainer.Host(ctx)
		if err != nil {
			return nil, err
		}
		tlsConfig, err := testutil.GetClientTlsConfig(serverHost)
		if err != nil {
			return nil, err
		}
		pgxCfg.ConnConfig.TLSConfig = tlsConfig
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, err
	}

	return &PostgresTestContainer{
		DB:            pool,
		URL:           connStr,
		TestContainer: pgContainer,

		cfg: cfg,
	}, nil
}

func (m *PostgresTestContainer) GetClientTlsConfig(ctx context.Context) (*tls.Config, error) {
	if !m.cfg.useTls {
		return nil, errors.New("tls is not enabled on this test container")
	}

	serverHost, err := m.TestContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	return testutil.GetClientTlsConfig(serverHost)
}

// Closes the connection pool and terminates the container.
func (p *PostgresTestContainer) TearDown(ctx context.Context) error {
	if p.DB != nil {
		p.DB.Close()
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
func (p *PostgresTestContainer) RunSqlFiles(ctx context.Context, folder *string, files []string) error {
	for _, file := range files {
		filePath := file
		if folder != nil && *folder != "" {
			filePath = fmt.Sprintf("./%s/%s", *folder, file)
		}
		sqlStr, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		_, err = p.DB.Exec(ctx, string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec sql when running postgres sql files: %w", err)
		}
	}
	return nil
}

// Creates schema and sets search_path to schema before running SQL files
func (p *PostgresTestContainer) RunCreateStmtsInSchema(ctx context.Context, folder string, files []string, schema string) error {
	for _, file := range files {
		filePath := file
		if folder != "" {
			filePath = fmt.Sprintf("./%s/%s", folder, file)
		}
		sqlStr, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		setSchemaSql := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q; \n SET search_path TO %q; \n", schema, schema)
		_, err = p.DB.Exec(ctx, setSchemaSql+string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec postgres create stmts in schema: %w", err)
		}
	}
	return nil
}

func (p *PostgresTestContainer) CreateSchemas(ctx context.Context, schemas []string) error {
	for _, schema := range schemas {
		_, err := p.DB.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q;", schema))
		if err != nil {
			return fmt.Errorf("unable to create schema %s: %w", schema, err)
		}
	}
	return nil
}

func (p *PostgresTestContainer) DropSchemas(ctx context.Context, schemas []string) error {
	for _, schema := range schemas {
		_, err := p.DB.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE;", schema))
		if err != nil {
			return fmt.Errorf("unable to drop schema %s: %w", schema, err)
		}
	}
	return nil
}

func (p *PostgresTestContainer) GetTableRowCount(ctx context.Context, schema, table string) (int, error) {
	rows := p.DB.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %q.%q;", schema, table))
	var count int
	err := rows.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
