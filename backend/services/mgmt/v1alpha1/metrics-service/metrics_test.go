package v1alpha1_metricsservice

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	promapiv1mock "github.com/nucleuscloud/neosync/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	anonymousUserId  = "00000000-0000-0000-0000-000000000000"
	mockAuthProvider = "test-provider"
	mockUserId       = "d5e29f1f-b920-458c-8b86-f3a180e06d98"
	mockAccountId    = "5629813e-1a35-4874-922c-9827d85f0378"
	mockJobId        = "884765c6-1708-488d-b03a-70a02b12c81e"
	mockJobRunId     = "004765c6-1708-488d-b03a-70a02b12c81e"
)

var (
	startTime = time.Date(2024, 3, 10, 00, 00, 00, 00, time.UTC)
	endTime   = time.Date(2024, 3, 10, 23, 59, 59, 00, time.UTC)

	startDate = mgmtv1alpha1.Date{Year: uint32(startTime.Year()), Month: uint32(startTime.Month()), Day: uint32(startTime.Day())}
	endDate   = mgmtv1alpha1.Date{Year: uint32(endTime.Year()), Month: uint32(endTime.Month()), Day: uint32(endTime.Day())}

	testVector = model.Vector{
		{
			Metric:    model.Metric{"foo": "bar"},
			Timestamp: 0,
			Value:     2,
		},
		{
			Metric:    model.Metric{"foo": "bar2"},
			Timestamp: 0,
			Value:     2,
		},
	}
)

func Test_GetMetricCount_Empty_Matrix(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(model.Vector{}, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   &endDate,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(0), resp.Msg.GetCount())
}

func Test_GetMetricCount_InvalidIdentifier(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay:   &startDate,
		EndDay:     &endDate,
		Metric:     mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: nil,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "must provide a valid identifier to proceed")
}

func Test_GetMetricCount_AccountId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(testVector, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   &endDate,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(4), resp.Msg.GetCount())
}

func Test_GetMetricCount_JobId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	m.JobServiceMock.On("GetJob", ctx, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				AccountId: mockAccountId,
				Id:        mockJobId,
			},
		}), nil)

	m.PromApiMock.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(testVector, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   &endDate,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_JobId{
			JobId: mockJobId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(4), resp.Msg.GetCount())
}

func Test_GetMetricCount_RunId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	m.JobServiceMock.On("GetJobRun", ctx, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{
			JobRun: &mgmtv1alpha1.JobRun{
				JobId: mockJobId,
				Id:    mockJobRunId,
			},
		}), nil)

	m.PromApiMock.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(testVector, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   &endDate,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_RunId{
			RunId: mockJobRunId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(4), resp.Msg.GetCount())
}

func Test_GetMetricCount_Bad_Times(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: nil,
		EndDay:   &endDate,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   nil,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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

	newStart := &mgmtv1alpha1.Date{
		Month: endDate.Month,
		Day:   endDate.Day + 1,
		Year:  endDate.Year,
	}

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: newStart,
		EndDay:   &endDate,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "start must not be before end")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   nil,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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

	newEndTime := startTime.Add(timeLimit + 1)
	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   &mgmtv1alpha1.Date{Year: uint32(newEndTime.Year()), Month: uint32(newEndTime.Month()), Day: uint32(newEndTime.Day())},
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "duration between start and end must not exceed 60 days")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		StartDay: &startDate,
		EndDay:   nil,
		Metric:   mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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
		StartDay: &startDate,
		EndDay:   &endDate,
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
	Service         *Service
	UserServiceMock *userdata.MockInterface
	JobServiceMock  *mgmtv1alpha1connect.MockJobServiceHandler
	PromApiMock     *promapiv1mock.MockAPI
}

func createServiceMock(t testing.TB, config *Config) *serviceMocks {
	t.Helper()

	mockUserService := userdata.NewMockInterface(t)
	mockJobService := mgmtv1alpha1connect.NewMockJobServiceHandler(t)
	mockPromApi := promapiv1mock.NewMockAPI(t)

	service := New(config, mockUserService, mockJobService, mockPromApi)
	return &serviceMocks{
		Service:         service,
		UserServiceMock: mockUserService,
		JobServiceMock:  mockJobService,
		PromApiMock:     mockPromApi,
	}
}

//nolint:unparam
func mockIsUserInAccount(userServiceMock *userdata.MockInterface, isInAccount bool) {

	// todo: fix this mock
	// userServiceMock.On("GetUser", mock.Anything).Return(userdata.User{
	// 	UserEntityEnforcer: ,
	// }, nil)
}

func Test_GetDailyMetricCount_Empty_Matrix(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(model.Vector{}, promv1.Warnings{}, nil)

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Msg.Results)
}

func Test_GetDailyMetricCount_InvalidIdentifier(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:      &startDate,
		End:        &endDate,
		Metric:     mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: nil,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "must provide a valid identifier to proceed")
}

func Test_GetDailyMetricCount_AccountId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(testVector, promv1.Warnings{}, nil)
	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	results := resp.Msg.GetResults()
	assert.Len(t, results, 1)
	assert.Equal(t, uint64(4), results[0].Count)
}

func Test_GetDailyMetricCount_JobId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	m.JobServiceMock.On("GetJob", ctx, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				AccountId: mockAccountId,
				Id:        mockJobId,
			},
		}), nil)

	m.PromApiMock.On("Query", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(testVector, promv1.Warnings{}, nil)

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_JobId{
			JobId: mockJobId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	results := resp.Msg.GetResults()
	assert.Len(t, results, 1)
	assert.Equal(t, uint64(4), results[0].Count)
}

func Test_GetDailyMetricCount_RunId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	m.JobServiceMock.On("GetJobRun", ctx, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobRunResponse{
			JobRun: &mgmtv1alpha1.JobRun{
				JobId: mockJobId,
				Id:    mockJobRunId,
			},
		}), nil)

	m.PromApiMock.On("Query", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(testVector, promv1.Warnings{}, nil)

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_RunId{
			RunId: mockJobRunId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	results := resp.Msg.GetResults()
	assert.Len(t, results, 1)
	assert.Equal(t, uint64(4), results[0].Count)
}

func Test_GetDailyMetricCount_MultipleDays(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(model.Vector{
			{
				Metric: model.Metric{"foo": "bar"},
				Value:  2,
			},
			{
				Metric: model.Metric{"foo": "bar2"},
				Value:  3,
			},
		}, promv1.Warnings{}, nil)

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	results := resp.Msg.GetResults()
	require.Len(t, results, 1)
	assert.Equal(t, uint64(5), results[0].Count)
}

func Test_GetDailyMetricCount_MultipleDays_Ordering(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", mock.MatchedBy(func(ctx context.Context) bool {
		return true
	}), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(model.Vector{
			{
				Metric: model.Metric{"foo": "bar2"},
				Value:  3,
			},
			{
				Metric: model.Metric{"foo": "bar"},
				Value:  2,
			},
		}, promv1.Warnings{}, nil)

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	results := resp.Msg.GetResults()
	require.Len(t, results, 1)
	assert.Equal(t, uint64(5), results[0].Count)
}

func Test_GetDailyMetricCount_Bad_Times(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  nil,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)

	resp, err = m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    nil,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)
}

func Test_GetDailyMetricCount_Swapped_Times(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	newStart := &mgmtv1alpha1.Date{
		Month: endDate.Month,
		Day:   endDate.Day + 1,
		Year:  endDate.Year,
	}
	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  newStart,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "start must not be before end")
	assert.Nil(t, resp)
}

func Test_GetDailyMetricCount_Time_Limit(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	newEndTime := startTime.Add(timeLimit + 1)
	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &startDate,
		End:    &mgmtv1alpha1.Date{Year: uint32(newEndTime.Year()), Month: uint32(newEndTime.Month()), Day: uint32(newEndTime.Day())},
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "duration between start and end must not exceed 60 days")
	assert.Nil(t, resp)
}

func Test_GetDailyMetricCount_No_Metric(t *testing.T) {
	m := createServiceMock(t, &Config{})

	ctx := context.Background()

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start: &startDate,
		End:   &endDate,
		// Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a metric name")
	assert.Nil(t, resp)
}
