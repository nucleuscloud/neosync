package v1alpha1_jobservice

import (
	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	temporalmocks "go.temporal.io/sdk/mocks"
)

func (s *IntegrationTestSuite) Test_GetJobs_Empty() {
	accountId := s.createPersonalAccount(s.userclient)

	resp, err := s.jobsclient.GetJobs(s.ctx, connect.NewRequest(&mgmtv1alpha1.GetJobsRequest{
		AccountId: accountId,
	}))
	requireNoErrResp(s.T(), resp, err)
	jobs := resp.Msg.GetJobs()
	require.Empty(s.T(), jobs)
}

func (s *IntegrationTestSuite) Test_CreateJob_Ok() {
	accountId := s.createPersonalAccount(s.userclient)
	srcconn := s.createPostgresConnection(s.connclient, accountId, "source", "test")
	destconn := s.createPostgresConnection(s.connclient, accountId, "dest", "test2")

	mockScheduleClient := temporalmocks.NewScheduleClient(s.T())
	mockScheduleHandle := temporalmocks.NewScheduleHandle(s.T())
	s.mockTemporalClientMgr.
		On(
			"DoesAccountHaveTemporalWorkspace", mock.Anything, mock.Anything, mock.Anything,
		).
		Return(true, nil).
		Once()
	s.mockTemporalClientMgr.
		On("GetScheduleClientByAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(mockScheduleClient, nil).
		Once()
	s.mockTemporalClientMgr.
		On("GetTemporalConfigByAccount", mock.Anything, mock.Anything).
		Return(&pg_models.TemporalConfig{}, nil).
		Once()
	mockScheduleClient.
		On("Create", mock.Anything, mock.Anything).
		Return(mockScheduleHandle, nil).
		Once()
	mockScheduleHandle.
		On("GetID").
		Return("test-id").
		Once()

	resp, err := s.jobsclient.CreateJob(s.ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
		AccountId: accountId,
		JobName:   "test",
		Mappings:  []*mgmtv1alpha1.JobMapping{},
		Source: &mgmtv1alpha1.JobSource{
			Options: &mgmtv1alpha1.JobSourceOptions{
				Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
					Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
						ConnectionId: srcconn.GetId(),
					},
				},
			},
		},
		Destinations: []*mgmtv1alpha1.CreateJobDestination{
			{
				ConnectionId: destconn.GetId(),
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
						PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{},
					},
				}},
		},
		InitiateJobRun: false,
	}))
	requireNoErrResp(s.T(), resp, err)
	require.NotNil(s.T(), resp.Msg.GetJob())
}

func (s *IntegrationTestSuite) createPostgresConnection(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	accountId string,
	name string,
	pgurl string,
) *mgmtv1alpha1.Connection {
	resp, err := connclient.CreateConnection(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
			AccountId: accountId,
			Name:      name,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: pgurl,
						},
					},
				},
			},
		}),
	)
	requireNoErrResp(s.T(), resp, err)
	return resp.Msg.GetConnection()
}
