package sqlmanager_mssql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/errgroup"

	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/suite"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
)

type IntegrationTestSuite struct {
	suite.Suite

	initSql     string
	setupSql    string
	teardownSql string

	ctx context.Context

	source *mssqlTestContainer
	target *mssqlTestContainer
}

type mssqlTestContainer struct {
	pool      *sql.DB
	querier   mssql_queries.Querier
	container *testmssql.MSSQLServerContainer
	// url       string
	close func()
}

type mssqlTest struct {
	source *mssqlTestContainer
	target *mssqlTestContainer
}

func (s *IntegrationTestSuite) SetupMssql() (*mssqlTest, error) {
	var source *mssqlTestContainer
	var target *mssqlTestContainer

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		sourcecontainer, err := createMssqlTestContainer(s.ctx, "datasync", "pass-source")
		if err != nil {
			return err
		}
		source = sourcecontainer
		return nil
	})

	errgrp.Go(func() error {
		targetcontainer, err := createMssqlTestContainer(s.ctx, "datasync", "pass-target")
		if err != nil {
			return err
		}
		target = targetcontainer
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return &mssqlTest{
		source: source,
		target: target,
	}, nil
}

func createMssqlTestContainer(
	ctx context.Context,
	database, password string,
) (*mssqlTestContainer, error) {
	container, err := testmssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		testmssql.WithAcceptEULA(),
		testmssql.WithPassword(password),
		// testcontainers.WithWaitStrategy(
		// 	wait.ForLog("port: 3306  MySQL Community Server").
		// 		WithOccurrence(1).WithStartupTimeout(20*time.Second),
		// ),
	)
	if err != nil {
		return nil, err
	}
	connstr, err := container.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	pool, err := sql.Open(sqlmanager_shared.MssqlDriver, connstr)
	if err != nil {
		panic(err)
	}
	// containerPort, err := container.MappedPort(ctx, "1433/tcp")
	// if err != nil {
	// 	return nil, err
	// }
	// containerHost, err := container.Host(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// connUrl := fmt.Sprintf("mysql://%s:%s@%s:%s/%s?multiStatements=true", username, password, containerHost, containerPort.Port(), database)
	return &mssqlTestContainer{
		pool:    pool,
		querier: mssql_queries.New(),
		// url:       connUrl,
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

	m, err := s.SetupMssql()
	if err != nil {
		panic(err)
	}
	s.source = m.source
	s.target = m.target

	initSql, err := os.ReadFile("./testdata/init.sql")
	if err != nil {
		panic(err)
	}
	s.initSql = string(initSql)

	setupSql, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		panic(err)
	}
	s.setupSql = string(setupSql)

	teardownSql, err := os.ReadFile("./testdata/teardown.sql")
	if err != nil {
		panic(err)
	}
	s.teardownSql = string(teardownSql)
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.target.pool.ExecContext(s.ctx, s.initSql)
	if err != nil {
		panic(err)
	}
	_, err = s.source.pool.ExecContext(s.ctx, s.setupSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	_, err := s.target.pool.ExecContext(s.ctx, s.teardownSql)
	if err != nil {
		panic(err)
	}
	_, err = s.source.pool.ExecContext(s.ctx, s.teardownSql)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.source.pool != nil {
		s.source.close()
	}
	if s.target.pool != nil {
		s.target.close()
	}
	if s.source != nil {
		err := s.source.container.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
	if s.target != nil {
		err := s.target.container.Terminate(s.ctx)
		if err != nil {
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

//nolint:unparam
func (s *IntegrationTestSuite) buildTable(schema, tableName string) string {
	return fmt.Sprintf("%s.%s", schema, tableName)
}
