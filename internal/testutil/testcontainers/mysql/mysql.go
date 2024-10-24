package testcontainers_mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

type MysqlTestSyncContainer struct {
	Source *MysqlTestContainer
	Target *MysqlTestContainer
}

func NewMysqlTestSyncContainer(ctx context.Context, sourceOpts, destOpts []Option) (*MysqlTestSyncContainer, error) {
	tc := &MysqlTestSyncContainer{}
	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		m, err := NewMysqlTestContainer(ctx, sourceOpts...)
		if err != nil {
			return err
		}
		tc.Source = m
		return nil
	})

	errgrp.Go(func() error {
		m, err := NewMysqlTestContainer(ctx, destOpts...)
		if err != nil {
			return err
		}
		tc.Target = m
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return tc, nil
}

func (m *MysqlTestSyncContainer) TearDown(ctx context.Context) error {
	if m.Source != nil {
		err := m.Source.TearDown(ctx)
		if err != nil {
			return err
		}
	}
	if m.Target != nil {
		err := m.Target.TearDown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Holds the MySQL test container and connection pool.
type MysqlTestContainer struct {
	DB            *sql.DB
	URL           string
	TestContainer *testmysql.MySQLContainer
	database      string
	password      string
	username      string
}

// Option is a functional option for configuring the Mysql Test Container
type Option func(*MysqlTestContainer)

// NewMysqlTestContainer initializes a new MySQL Test Container with functional options
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

	db, err := sql.Open(sqlmanager_shared.MysqlDriver, connStr)
	if err != nil {
		return nil, err
	}

	return &MysqlTestContainer{
		DB:            db,
		URL:           connStr,
		TestContainer: mysqlContainer,
	}, nil
}

// Closes the connection pool and terminates the container.
func (m *MysqlTestContainer) TearDown(ctx context.Context) error {
	if m.DB != nil {
		m.DB.Close()
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
func (m *MysqlTestContainer) RunSqlFiles(ctx context.Context, folder *string, files []string) error {
	for _, file := range files {
		filePath := file
		if folder != nil && *folder != "" {
			filePath = fmt.Sprintf("./%s/%s", *folder, file)
		}
		sqlStr, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		_, err = m.DB.ExecContext(ctx, string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec SQL when running MySQL SQL files: %w", err)
		}
	}
	return nil
}
