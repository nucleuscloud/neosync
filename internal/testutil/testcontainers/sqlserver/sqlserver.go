package testcontainers_sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
	"golang.org/x/sync/errgroup"
)

type MssqlTestSyncContainer struct {
	Source *MssqlTestContainer
	Target *MssqlTestContainer
}

func NewMssqlTestSyncContainer(ctx context.Context, sourceOpts, destOpts []Option) (*MssqlTestSyncContainer, error) {
	tc := &MssqlTestSyncContainer{}
	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		m, err := NewMssqlTestContainer(ctx, sourceOpts...)
		if err != nil {
			return err
		}
		tc.Source = m
		return nil
	})

	errgrp.Go(func() error {
		m, err := NewMssqlTestContainer(ctx, destOpts...)
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

func (m *MssqlTestSyncContainer) TearDown(ctx context.Context) error {
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

// Holds the MsSQL test container and connection pool.
type MssqlTestContainer struct {
	DB            *sql.DB
	URL           string
	TestContainer *testmssql.MSSQLServerContainer
	database      string
	password      string
}

// Option is a functional option for configuring the MsSQL Test Container
type Option func(*MssqlTestContainer)

// NewMssqlTestContainer initializes a new MsSQL Test Container with functional options
func NewMssqlTestContainer(ctx context.Context, opts ...Option) (*MssqlTestContainer, error) {
	m := &MssqlTestContainer{
		database: "testdb",
		password: "mssqlPASSword1",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m.setup(ctx)
}

// Sets test container database
func WithDatabase(database string) Option {
	return func(a *MssqlTestContainer) {
		a.database = database
	}
}

// Sets test container database
func WithPassword(password string) Option {
	return func(a *MssqlTestContainer) {
		a.password = password
	}
}

// Creates and starts a MsSQL test container and sets up the connection.
func (m *MssqlTestContainer) setup(ctx context.Context) (*MssqlTestContainer, error) {
	mssqlcontainer, err := testmssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		testmssql.WithAcceptEULA(),
		testmssql.WithPassword(m.password),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := mssqlcontainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(sqlmanager_shared.MssqlDriver, connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE [%s];", m.database))
	if err != nil {
		return nil, err
	}

	queryvals := url.Values{}
	queryvals.Add("database", "testdb")
	dbConnStr := connStr + queryvals.Encode()

	dbConn, err := sql.Open(sqlmanager_shared.MssqlDriver, dbConnStr)
	if err != nil {
		return nil, err
	}

	return &MssqlTestContainer{
		DB:            dbConn,
		URL:           dbConnStr,
		TestContainer: mssqlcontainer,
	}, nil
}

// Closes the connection pool and terminates the container.
func (m *MssqlTestContainer) TearDown(ctx context.Context) error {
	if m.DB != nil {
		m.DB.Close()
	}

	if m.TestContainer != nil {
		err := m.TestContainer.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("failed to terminate MsSQL container: %w", err)
		}
	}

	return nil
}

// Executes SQL files within the test container
func (m *MssqlTestContainer) RunSqlFiles(ctx context.Context, folder *string, files []string) error {
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
			return fmt.Errorf("unable to exec SQL when running MsSQL SQL files: %w", err)
		}
	}
	return nil
}
