package integrationtest

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testcontainers/postgres"
	"golang.org/x/sync/errgroup"
)

type IntegrationTestSuite struct {
	neosyncApi *tcneosyncapi.NeosyncApiTestClient
	postgres   *tcpostgres.PostgresTestContainer
}

func newCliIntegrationTest(ctx context.Context, t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{}
	suite.SetupSuite(ctx, t)
	return suite
}

func (s *IntegrationTestSuite) SetupSuite(ctx context.Context, t *testing.T) {
	var neosyncApiTest *tcneosyncapi.NeosyncApiTestClient
	var postgresTest *tcpostgres.PostgresTestContainer

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		p, err := tcpostgres.NewPostgresTestContainer(ctx)
		if err != nil {
			return err
		}
		postgresTest = p
		return nil
	})

	errgrp.Go(func() error {
		api := tcneosyncapi.NewNeosyncApiTestClient(ctx, t)
		neosyncApiTest = api
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		panic(err)
	}

	s.postgres = postgresTest
	s.neosyncApi = neosyncApiTest
}

func (s *IntegrationTestSuite) TearDownSuite(ctx context.Context) {
	err := s.postgres.TearDown(ctx)
	if err != nil {
		panic(err)
	}
	err = s.neosyncApi.TearDown(ctx)
	if err != nil {
		panic(err)
	}
}

func shouldRun() bool {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
		return false
	}
	return true
}
