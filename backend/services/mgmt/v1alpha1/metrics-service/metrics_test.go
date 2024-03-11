package v1alpha1_metricsservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	mockPromV1 "github.com/nucleuscloud/neosync/backend/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	anonymousUserId    = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider   = "test-provider"
	mockUserId         = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId      = "5629813e-1a35-4874-922c-9827d85f0378"
	mockConnectionName = "test-conn"
	mockConnectionId   = "884765c6-1708-488d-b03a-70a02b12c81e"
)

var (
	startTime = timestamppb.New(time.Date(2024, 03, 10, 14, 14, 00, 00, time.Local))
	endTime   = timestamppb.New(time.Date(2024, 03, 11, 14, 14, 00, 00, time.Local))
)

func Test_GetMetricCount_Empty_Matrix(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(model.Matrix{}, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  startTime,
		End:    endTime,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(0), resp.Msg.GetCount())
}

func Test_GetMetricCount_Bad_Times(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  nil,
		End:    endTime,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  startTime,
		End:    nil,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)
}

func Test_GetMetricCount_Swapped_Times(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  endTime,
		End:    startTime,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "start must not be before end")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  startTime,
		End:    nil,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)
}

func Test_GetMetricCount_Time_Limit(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  startTime,
		End:    timestamppb.New(startTime.AsTime().Add(timeLimit + 1)),
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "duration between start and end must not exceed 60 days")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  startTime,
		End:    nil,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)
}

func Test_GetMetricCount_No_Metric(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start: startTime,
		End:   endTime,
		// Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a metric name")
	assert.Nil(t, resp)
}

type serviceMocks struct {
	Service                *Service
	UserAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient
	JobServiceMock         *mgmtv1alpha1connect.MockJobServiceHandler
	PromApiMock            *mockPromV1.MockAPI
}

func createServiceMock(t testing.TB, config *Config) *serviceMocks {
	t.Helper()

	mockUserAccService := mgmtv1alpha1connect.NewMockUserAccountServiceClient(t)
	mockJobService := mgmtv1alpha1connect.NewMockJobServiceHandler(t)
	mockPromApi := mockPromV1.NewMockAPI(t)

	service := New(config, mockUserAccService, mockJobService, mockPromApi)
	return &serviceMocks{
		Service:                service,
		UserAccountServiceMock: mockUserAccService,
		JobServiceMock:         mockJobService,
		PromApiMock:            mockPromApi,
	}
}

func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}
