package testcontainers_mariadb

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mysqltunconnector"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

type MariaDBTestSyncContainer struct {
	Source *MariaDBTestContainer
	Target *MariaDBTestContainer
}

func NewMariaDBTestSyncContainer(ctx context.Context, sourceOpts, destOpts []Option) (*MariaDBTestSyncContainer, error) {
	tc := &MariaDBTestSyncContainer{}
	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		m, err := NewMariaDBTestContainer(ctx, sourceOpts...)
		if err != nil {
			return err
		}
		tc.Source = m
		return nil
	})

	errgrp.Go(func() error {
		m, err := NewMariaDBTestContainer(ctx, destOpts...)
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

func (m *MariaDBTestSyncContainer) TearDown(ctx context.Context) error {
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

type MariaDBTestContainer struct {
	DB            *sql.DB
	URL           string
	TestContainer *mariadb.MariaDBContainer

	cfg *mariadbTestContainerConfig
}

type mariadbTestContainerConfig struct {
	database string
	password string
	username string
}

// Option is a functional option for configuring the MariaDB Test Container
type Option func(*mariadbTestContainerConfig)

// NewMariaDBTestContainer initializes a new MariaDB Test Container with functional options
func NewMariaDBTestContainer(ctx context.Context, opts ...Option) (*MariaDBTestContainer, error) {
	m := &mariadbTestContainerConfig{
		database: "testdb",
		username: "root",
		password: "pass",
		// useTls:   false,
	}
	for _, opt := range opts {
		opt(m)
	}
	return setup(ctx, m)
}

// Sets test container database
func WithDatabase(database string) Option {
	return func(a *mariadbTestContainerConfig) {
		a.database = database
	}
}

// Sets test container database
func WithUsername(username string) Option {
	return func(a *mariadbTestContainerConfig) {
		a.username = username
	}
}

// Sets test container database
func WithPassword(password string) Option {
	return func(a *mariadbTestContainerConfig) {
		a.password = password
	}
}

// Creates and starts a MariaDB test container and sets up the connection.
func setup(ctx context.Context, cfg *mariadbTestContainerConfig) (*MariaDBTestContainer, error) {
	tcopts := []testcontainers.ContainerCustomizer{
		mariadb.WithDatabase(cfg.database),
		mariadb.WithUsername(cfg.username),
		mariadb.WithPassword(cfg.password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  mariadb.org binary distribution").WithOccurrence(1).WithStartupTimeout(20 * time.Second),
		),
	}
	// if cfg.useTls {
	// 	clientCertPaths, err := testutil.GetTlsCertificatePaths()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tcopts = append(
	// 		tcopts,
	// 		testutil.WithCmd([]string{
	// 			"mysqld",
	// 			"--ssl-ca=/etc/mysql/certs/root.crt",
	// 			"--ssl-cert=/etc/mysql/certs/server.crt",
	// 			"--ssl-key=/etc/mysql/certs/server.key",
	// 			"--require-secure-transport=ON",
	// 			"--tls-version=TLSv1.2,TLSv1.3",
	// 		}),
	// 		testutil.WithFiles([]testcontainers.ContainerFile{
	// 			{
	// 				HostFilePath:      clientCertPaths.ServerCertPath,
	// 				ContainerFilePath: "/etc/mysql/certs/server.crt",
	// 				FileMode:          0644,
	// 			},
	// 			{
	// 				HostFilePath:      clientCertPaths.ServerKeyPath,
	// 				ContainerFilePath: "/etc/mysql/certs/server.key",
	// 				FileMode:          0600,
	// 			},
	// 			{
	// 				HostFilePath:      clientCertPaths.RootCertPath,
	// 				ContainerFilePath: "/etc/mysql/certs/root.crt",
	// 				FileMode:          0644,
	// 			},
	// 		}),
	// 		testcontainers.WithStartupCommand(testcontainers.NewRawCommand([]string{
	// 			"chown", "mysql:mysql", "/etc/mysql/certs/server.key",
	// 		})),
	// 	)
	// }
	mariaDb, err := mariadb.Run(
		ctx,
		"mariadb:11.7.2",
		tcopts...,
	)
	if err != nil {
		return nil, err
	}

	connstrArgs := []string{"multiStatements=true", "parseTime=true"}
	// if cfg.useTls {
	// 	connstrArgs = append(connstrArgs, "tls=true")
	// }

	connStr, err := mariaDb.ConnectionString(ctx, connstrArgs...)
	if err != nil {
		return nil, err
	}

	connectorOpts := []mysqltunconnector.Option{}
	// if cfg.useTls {
	// 	serverHost, err := mariaDb.Host(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tlsConfig, err := testutil.GetClientTlsConfig(serverHost)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	connectorOpts = append(connectorOpts, mysqltunconnector.WithTLSConfig(tlsConfig))
	// }

	connector, _, err := mysqltunconnector.New(connStr, connectorOpts...)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)

	return &MariaDBTestContainer{
		DB:            db,
		URL:           connStr,
		TestContainer: mariaDb,

		cfg: cfg,
	}, nil
}

func (m *MariaDBTestContainer) GetClientTlsConfig(ctx context.Context) (*tls.Config, error) {
	// if !m.cfg.useTls {
	// 	return nil, errors.New("tls is not enabled on this test container")
	// }

	serverHost, err := m.TestContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	return testutil.GetClientTlsConfig(serverHost)
}

// Closes the connection pool and terminates the container.
func (m *MariaDBTestContainer) TearDown(ctx context.Context) error {
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
func (m *MariaDBTestContainer) RunSqlFiles(ctx context.Context, folder *string, files []string) error {
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

// Creates schema and sets USE to schema before running SQL files
func (m *MariaDBTestContainer) RunCreateStmtsInDatabase(ctx context.Context, folder string, files []string, database string) error {
	for _, file := range files {
		filePath := file
		if folder != "" {
			filePath = fmt.Sprintf("./%s/%s", folder, file)
		}
		sqlStr, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		setSchemaSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`; \n USE `%s`; \n", database, database)
		_, err = m.DB.ExecContext(ctx, setSchemaSql+string(sqlStr))
		if err != nil {
			return fmt.Errorf("unable to exec sql when running mysql sql files: %w", err)
		}
	}
	return nil
}

func (m *MariaDBTestContainer) CreateDatabases(ctx context.Context, databases []string) error {
	for _, database := range databases {
		_, err := m.DB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", database))
		if err != nil {
			return fmt.Errorf("unable to create database %s: %w", database, err)
		}
	}
	return nil
}

func (m *MariaDBTestContainer) DropDatabases(ctx context.Context, databases []string) error {
	for _, database := range databases {
		_, err := m.DB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", database))
		if err != nil {
			return fmt.Errorf("unable to drop database %s: %w", database, err)
		}
	}
	return nil
}

func (m *MariaDBTestContainer) GetTableRowCount(ctx context.Context, schema, table string) (int, error) {
	rows := m.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM  `%s`.`%s`;", schema, table))
	var count int
	err := rows.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
