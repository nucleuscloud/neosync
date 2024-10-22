package integrationtest

import (
	"context"
	"fmt"
	"testing"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testcontainers/postgres"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	neosyncApi *tcneosyncapi.NeosyncApiTestClient

	postgres *tcpostgres.PostgresTestContainer
}

func (s *IntegrationTestSuite) SetupSuite() {
	fmt.Println("HERE")
	s.ctx = context.Background()

	var neosyncApiTest *tcneosyncapi.NeosyncApiTestClient
	var postgresTest *tcpostgres.PostgresTestContainer

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		p, err := tcpostgres.NewPostgresTestContainer(s.ctx)
		if err != nil {
			return err
		}
		postgresTest = p
		return nil
	})

	errgrp.Go(func() error {
		api := tcneosyncapi.NewNeosyncApiTestClient(s.ctx)
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

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	err := s.postgres.TearDown(s.ctx)
	if err != nil {
		panic(err)
	}
	err = s.neosyncApi.TearDown()
	if err != nil {
		panic(err)
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	t.Log("Running integration tests")
	// evkey := "INTEGRATION_TESTS_ENABLED"
	// shouldRun := os.Getenv(evkey)
	// if shouldRun != "1" {
	// 	slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
	// 	return
	// }
	suite.Run(t, new(IntegrationTestSuite))
}
