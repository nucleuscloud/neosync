package sqlmanager_mssql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"

	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmssql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/sqlserver"
	"github.com/stretchr/testify/suite"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
)

type IntegrationTestSuite struct {
	suite.Suite

	sourceSetupStatements []string
	destSetupStatements   []string

	teardownStatements []string

	ctx context.Context

	source *mssqlTestContainer
	target *mssqlTestContainer
}

type mssqlTestContainer struct {
	// master db connection
	masterDb *sql.DB
	// test db connection
	testDb        *sql.DB
	testDbConnStr string
	querier       mssql_queries.Querier
	container     *testmssql.MSSQLServerContainer
	close         func()
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
		sourcecontainer, err := createMssqlTestContainer(s.ctx)
		if err != nil {
			return fmt.Errorf("unable to start mssql source container: %w", err)
		}
		source = sourcecontainer
		return nil
	})

	errgrp.Go(func() error {
		targetcontainer, err := createMssqlTestContainer(s.ctx)
		if err != nil {
			return fmt.Errorf("unable to start mssql dest container: %w", err)
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
) (*mssqlTestContainer, error) {
	container, err := testmssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		testmssql.WithAcceptEULA(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to start container: %w", err)
	}
	connstr, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get mssql connection str: %w", err)
	}

	pool, err := sql.Open(sqlmanager_shared.MssqlDriver, connstr)
	if err != nil {
		return nil, fmt.Errorf("unable to open mssql connection: %w", err)
	}

	queryvals := url.Values{}
	queryvals.Add("database", "testdb")

	return &mssqlTestContainer{
		masterDb:      pool,
		testDbConnStr: connstr + queryvals.Encode(),
		querier:       mssql_queries.New(),
		container:     container,
		close: func() {
			if pool != nil {
				pool.Close()
			}
		},
	}, nil
}

func readSqlFiles(dir string) ([]string, error) {
	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %v", dir, err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Prepare a slice to store results
	sqlContents := make([]string, len(sqlFiles))

	// Use errgroup for concurrent file reading
	var eg errgroup.Group
	var mu sync.Mutex

	for i, file := range sqlFiles {
		i, file := i, file
		eg.Go(func() error {
			content, err := os.ReadFile(filepath.Join(dir, file))
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", file, err)
			}

			mu.Lock()
			sqlContents[i] = string(content)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all goroutines to complete and check for errors
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return sqlContents, nil
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	m, err := s.SetupMssql()
	if err != nil {
		panic(err)
	}
	s.source = m.source
	s.target = m.target

	baseDir := "testdata"

	sourceSetupContents, err := readSqlFiles(filepath.Join(baseDir, "source-setup"))
	if err != nil {
		panic(fmt.Errorf("unable to read source setup files: %w", err))
	}
	s.sourceSetupStatements = sourceSetupContents

	destSetupContents, err := readSqlFiles(filepath.Join(baseDir, "dest-setup"))
	if err != nil {
		panic(fmt.Errorf("unable to read dest setup files: %w", err))
	}
	s.destSetupStatements = destSetupContents

	teardownContents, err := readSqlFiles(filepath.Join(baseDir, "teardown"))
	if err != nil {
		panic(fmt.Errorf("unable to read teardown files: %w", err))
	}
	s.teardownStatements = teardownContents
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	errgrp, errctx := errgroup.WithContext(s.ctx)
	errgrp.Go(func() error {
		for i, stmt := range s.sourceSetupStatements {
			_, err := s.source.masterDb.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing source setup statement %d: %w", i+1, err)
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		for i, stmt := range s.destSetupStatements {
			_, err := s.target.masterDb.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing dest setup statemetn: %d: %w", i+1, err)
			}
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		panic(err)
	}
	// We have to initialize this because every test creates a fresh database so it must be reconnected to.
	sourceTestDbPool, err := sql.Open(sqlmanager_shared.MssqlDriver, s.source.testDbConnStr)
	if err != nil {
		panic(fmt.Errorf("unable to open source testdb mssql connection: %w", err))
	}
	s.source.testDb = sourceTestDbPool

	targetTestDbPool, err := sql.Open(sqlmanager_shared.MssqlDriver, s.target.testDbConnStr)
	if err != nil {
		panic(fmt.Errorf("unable to open target testdb mssql connection: %w", err))
	}
	s.target.testDb = targetTestDbPool
}

func (s *IntegrationTestSuite) TearDownTest() {
	// We have to close these connections prior to tearing down so that the database can be closed cleanly
	if s.target != nil && s.target.testDb != nil {
		s.target.testDb.Close()
		s.target.testDb = nil
	}
	if s.source != nil && s.source.testDb != nil {
		s.source.testDb.Close()
		s.source.testDb = nil
	}

	errgrp, errctx := errgroup.WithContext(s.ctx)
	errgrp.Go(func() error {
		for i, stmt := range s.teardownStatements {
			_, err := s.source.masterDb.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing source teardown statement %d: %w", i+1, err)
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		for i, stmt := range s.teardownStatements {
			_, err := s.target.masterDb.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing dest teardown statemetn: %d: %w", i+1, err)
			}
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.source.masterDb != nil {
		s.source.close()
	}
	if s.target.masterDb != nil {
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
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func buildTable(schema, tableName string) string {
	return fmt.Sprintf("%s.%s", schema, tableName)
}

type Schema struct {
	Name string
}

func setup(ctx context.Context, containers *tcmssql.MssqlTestSyncContainer) error {
	baseDir := "testdata"

	sourceSetupContents, err := readSqlFiles(filepath.Join(baseDir, "source-setup"))
	if err != nil {
		return fmt.Errorf("unable to read source setup files: %w", err)
	}

	destSetupContents, err := readSqlFiles(filepath.Join(baseDir, "dest-setup"))
	if err != nil {
		return fmt.Errorf("unable to read dest setup files: %w", err)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		for i, stmt := range sourceSetupContents {
			_, err := containers.Source.DB.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing source setup statement %d: %w", i+1, err)
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		for i, stmt := range destSetupContents {
			_, err := containers.Target.DB.ExecContext(errctx, stmt)
			if err != nil {
				return fmt.Errorf("encountered error when executing dest setup statement: %d: %w", i+1, err)
			}
		}
		return nil
	})

	err = errgrp.Wait()
	if err != nil {
		return err
	}

	return nil
}
