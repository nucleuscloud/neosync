package v1alpha1_jobservice

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
)

const (
	anonymousUserId    = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider   = "test-provider"
	mockUserId         = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId      = "5629813e-1a35-4874-922c-9827d85f0378"
	mockConnectionName = "test-conn"
	mockConnectionId   = "884765c6-1708-488d-b03a-70a02b12c81e"
)

// GetJobs
func Test_GetJobs(t *testing.T) {
	m := createServiceMock(t, &Config{IsAuthEnabled: true})

	resp, err := m.Service.GetJobs(context.Background(), &connect.Request[mgmtv1alpha1.GetJobsRequest]{
		Msg: &mgmtv1alpha1.GetJobsRequest{
			AccountId: mockAccountId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

type serviceMocks struct {
	Service                     *Service
	DbtxMock                    *nucleusdb.MockDBTX
	QuerierMock                 *db_queries.MockQuerier
	UserAccountServiceMock      *mgmtv1alpha1connect.MockUserAccountServiceClient
	ConnectionServiceClientMock *mgmtv1alpha1connect.MockConnectionServiceClient
	TemporalWfManagerMock       clientmanager.TemporalClientManagerClient
}

func createServiceMock(t *testing.T, config *Config) *serviceMocks {
	mockDbtx := nucleusdb.NewMockDBTX(t)
	mockQuerier := db_queries.NewMockQuerier(t)
	mockUserAccountService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockConnectionService := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTemporalWfManager := clientmanager.NewMockTemporalClientManagerClient(t)

	service := New(config, nucleusdb.New(mockDbtx, mockQuerier), mockTemporalWfManager, mockConnectionService, mockUserAccountService)

	return &serviceMocks{
		Service:                     service,
		DbtxMock:                    mockDbtx,
		QuerierMock:                 mockQuerier,
		UserAccountServiceMock:      mockUserAccountService,
		ConnectionServiceClientMock: mockConnectionService,
		TemporalWfManagerMock:       mockTemporalWfManager,
	}
}
