package integrationtests_test

import (
	"context"
	"testing"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	tcneosyncapi.NeosyncApiTestClient
	ctx context.Context
}

// TODO update service integration tests to not use testify suite
func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	api, err := tcneosyncapi.NewNeosyncApiTestClient(s.ctx, s.T(), tcneosyncapi.WithMigrationsDirectory("../../../backend/sql/postgresql/schema"))
	if err != nil {
		s.T().Fatalf("unable to create neosync api test client: %v", err)
	}
	s.NeosyncApiTestClient = *api
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	err := s.InitializeTest(s.ctx, s.T())
	if err != nil {
		s.T().Fatalf("unable to initialize test: %v", err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	err := s.CleanupTest(s.ctx)
	if err != nil {
		s.T().Fatalf("unable to cleanup test: %v", err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	err := s.TearDown(s.ctx)
	if err != nil {
		s.T().Fatalf("unable to cleanup test: %v", err)
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}
