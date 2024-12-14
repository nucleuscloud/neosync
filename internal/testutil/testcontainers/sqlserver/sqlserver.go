package testcontainers_sqlserver

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mssqltunconnector"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/testcontainers/testcontainers-go"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/wait"
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

type mssqlTestContainerConfig struct {
	database string
	password string
	useTls   bool
}

// Holds the MsSQL test container and connection pool.
type MssqlTestContainer struct {
	DB            *sql.DB
	URL           string
	TestContainer *testmssql.MSSQLServerContainer

	cfg *mssqlTestContainerConfig
}

// Option is a functional option for configuring the MsSQL Test Container
type Option func(*mssqlTestContainerConfig)

// NewMssqlTestContainer initializes a new MsSQL Test Container with functional options
func NewMssqlTestContainer(ctx context.Context, opts ...Option) (*MssqlTestContainer, error) {
	m := &mssqlTestContainerConfig{
		database: "testdb",
		password: "mssqlPASSword1",
		useTls:   false,
	}
	for _, opt := range opts {
		opt(m)
	}
	return setup(ctx, m)
}

// Sets test container database
func WithDatabase(database string) Option {
	return func(a *mssqlTestContainerConfig) {
		a.database = database
	}
}

// Sets test container database
func WithPassword(password string) Option {
	return func(a *mssqlTestContainerConfig) {
		a.password = password
	}
}

func WithTls() Option {
	return func(mtc *mssqlTestContainerConfig) {
		mtc.useTls = true
	}
}

// Creates and starts a MsSQL test container and sets up the connection.
func setup(ctx context.Context, cfg *mssqlTestContainerConfig) (*MssqlTestContainer, error) {
	tcOpts := []testcontainers.ContainerCustomizer{
		testmssql.WithAcceptEULA(),
		testmssql.WithPassword(cfg.password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Recovery is complete.").WithStartupTimeout(20 * time.Second),
		),
	}
	if cfg.useTls {
		mssqlDf, err := testutil.GetMssqlTlsDockerfile()
		if err != nil {
			return nil, err
		}
		tcOpts = append(
			tcOpts,
			testutil.WithDockerFile(mssqlDf),
		)
	}
	mssqlcontainer, err := testmssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-latest", // WithDockerFile overrides the image and updates it to be empty
		tcOpts...,
	)
	if err != nil {
		return nil, err
	}

	connStrArgs := []string{}
	if cfg.useTls {
		connStrArgs = append(connStrArgs, "encrypt=true")
	} else {
		connStrArgs = append(connStrArgs, "encrypt=disable")
	}

	connStr, err := mssqlcontainer.ConnectionString(ctx, connStrArgs...)
	if err != nil {
		return nil, fmt.Errorf("uanble to build mssql conn str: %w", err)
	}

	connectorOpts := []mssqltunconnector.Option{}
	if cfg.useTls {
		serverHost, err := mssqlcontainer.Host(ctx)
		if err != nil {
			return nil, err
		}
		tlsConfig, err := testutil.GetClientTlsConfig(serverHost)
		if err != nil {
			return nil, err
		}

		connectorOpts = append(connectorOpts, mssqltunconnector.WithTLSConfig(tlsConfig))
	}

	connector, cleanup, err := mssqltunconnector.New(connStr, connectorOpts...)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	db := sql.OpenDB(connector)
	defer db.Close()

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE [%s];", cfg.database))
	if err != nil {
		return nil, fmt.Errorf("unable to create mssql database: %w", err)
	}

	queryvals := url.Values{}
	queryvals.Add("database", cfg.database)
	dbConnStr := connStr + "&" + queryvals.Encode() // adding & due to existing query param above

	dbconnector, _, err := mssqltunconnector.New(dbConnStr, connectorOpts...)
	if err != nil {
		return nil, err
	}

	dbConn := sql.OpenDB(dbconnector)
	return &MssqlTestContainer{
		DB:            dbConn,
		URL:           dbConnStr,
		TestContainer: mssqlcontainer,
		cfg:           cfg,
	}, nil
}

func (m *MssqlTestContainer) GetClientTlsConfig(ctx context.Context) (*tls.Config, error) {
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
