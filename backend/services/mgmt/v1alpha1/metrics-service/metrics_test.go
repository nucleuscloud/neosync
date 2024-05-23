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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	mockPromV1 "github.com/nucleuscloud/neosync/backend/internal/mocks/github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
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
	startTime = timestamppb.New(time.Date(2024, 03, 10, 14, 14, 00, 00, time.Local))
	endTime   = timestamppb.New(time.Date(2024, 03, 11, 14, 14, 00, 00, time.Local))

	startDate = mgmtv1alpha1.Date{Year: uint32(startTime.AsTime().Year()), Month: uint32(startTime.AsTime().Month()), Day: uint32(startTime.AsTime().Day())}
	endDate   = mgmtv1alpha1.Date{Year: uint32(endTime.AsTime().Year()), Month: uint32(endTime.AsTime().Month()), Day: uint32(endTime.AsTime().Day())}

	testMatrix = model.Matrix{
		{
			Metric: model.Metric{"foo": "bar"},
			Values: []model.SamplePair{
				{Timestamp: 0, Value: 1},
				{Timestamp: 0, Value: 2},
			},
		},
		{
			Metric: model.Metric{"foo": "bar2"},
			Values: []model.SamplePair{
				{Timestamp: 0, Value: 1},
				{Timestamp: 0, Value: 2},
			},
		},
	}
)

func Test_GetMetricCount_Empty_Matrix(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(model.Matrix{}, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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
		Start:      &startDate,
		End:        &endDate,
		Metric:     mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: nil,
	}))
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "must provide a valid identifier to proceed")
}

func Test_GetMetricCount_AccountId(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(testMatrix, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(testMatrix, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(testMatrix, promv1.Warnings{}, nil)

	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
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
		Start:  nil,
		End:    &endDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "must provide a start and end time")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
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
		Start:  &endDate,
		End:    &startDate,
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "start must not be before end")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
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

	newEndTime := startTime.AsTime().Add(timeLimit + 1)
	resp, err := m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
		End:    &mgmtv1alpha1.Date{Year: uint32(newEndTime.Year()), Month: uint32(newEndTime.Month()), Day: uint32(newEndTime.Day())},
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Identifier: &mgmtv1alpha1.GetMetricCountRequest_AccountId{
			AccountId: mockAccountId,
		},
	}))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "duration between start and end must not exceed 60 days")
	assert.Nil(t, resp)

	resp, err = m.Service.GetMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetMetricCountRequest{
		Start:  &startDate,
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
		Start: &startDate,
		End:   &endDate,
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

//nolint:unparam
func mockIsUserInAccount(userAccountServiceMock *mgmtv1alpha1connect.MockUserAccountServiceClient, isInAccount bool) {
	userAccountServiceMock.On("IsUserInAccount", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.IsUserInAccountResponse{
		Ok: isInAccount,
	}), nil)
}

func Test_getPromQueryFromMetric(t *testing.T) {
	output, err := getPromQueryFromMetric(
		mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		metrics.MetricLabels{metrics.NewEqLabel("foo", "bar"), metrics.NewEqLabel("foo2", "bar2")},
		"1d",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(
		t,
		`input_received_total{foo="bar",foo2="bar2"}`,
		output,
	)
}

func Test_getPromQueryFromMetric_Invalid_Metric(t *testing.T) {
	output, err := getPromQueryFromMetric(
		mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_UNSPECIFIED,
		metrics.MetricLabels{metrics.NewEqLabel("foo", "bar"), metrics.NewEqLabel("foo2", "bar2")},
		"1d",
	)
	assert.Error(t, err)
	assert.Empty(t, output)
}

func Test_GetDailyMetricCount_Empty_Matrix(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(model.Matrix{}, promv1.Warnings{}, nil)

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

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(testMatrix, promv1.Warnings{}, nil)

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

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(testMatrix, promv1.Warnings{}, nil)

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

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(testMatrix, promv1.Warnings{}, nil)

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

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(model.Matrix{
			{
				Metric: model.Metric{"foo": "bar"},
				Values: []model.SamplePair{
					{Timestamp: model.Time(time.Date(2024, 10, 3, 0, 0, 0, 0, time.UTC).UnixMilli()), Value: 1},
					{Timestamp: model.Time(time.Date(2024, 10, 3, 0, 1, 0, 0, time.UTC).UnixMilli()), Value: 2},
				},
			},
			{
				Metric: model.Metric{"foo": "bar2"},
				Values: []model.SamplePair{
					{Timestamp: model.Time(time.Date(2024, 11, 3, 0, 0, 0, 0, time.UTC).UnixMilli()), Value: 1},
					{Timestamp: model.Time(time.Date(2024, 11, 3, 0, 1, 0, 0, time.UTC).UnixMilli()), Value: 3},
				},
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
	require.Len(t, results, 2)
	assert.Equal(t, uint64(2), results[0].Count)
	assert.Equal(t, uint64(3), results[1].Count)
}

func Test_GetDailyMetricCount_MultipleDays_Ordering(t *testing.T) {
	m := createServiceMock(t, &Config{})

	mockIsUserInAccount(m.UserAccountServiceMock, true)

	ctx := context.Background()

	m.PromApiMock.On("QueryRange", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).
		Return(model.Matrix{
			{
				Metric: model.Metric{"foo": "bar2"},
				Values: []model.SamplePair{
					{Timestamp: model.Time(time.Date(2024, 11, 3, 0, 0, 0, 0, time.UTC).UnixMilli()), Value: 1},
					{Timestamp: model.Time(time.Date(2024, 11, 3, 0, 1, 0, 0, time.UTC).UnixMilli()), Value: 3},
				},
			},
			{
				Metric: model.Metric{"foo": "bar"},
				Values: []model.SamplePair{
					{Timestamp: model.Time(time.Date(2024, 10, 3, 0, 0, 0, 0, time.UTC).UnixMilli()), Value: 1},
					{Timestamp: model.Time(time.Date(2024, 10, 3, 0, 1, 0, 0, time.UTC).UnixMilli()), Value: 2},
				},
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
	require.Len(t, results, 2)
	assert.Equal(t, uint64(2), results[0].Count)
	assert.Equal(t, uint32(10), results[0].Date.Month, "the expected month should be 10")
	assert.Equal(t, uint64(3), results[1].Count)
	assert.Equal(t, uint32(11), results[1].Date.Month, "the expected month should be 11")
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

	resp, err := m.Service.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Start:  &endDate,
		End:    &startDate,
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

	newEndTime := startTime.AsTime().Add(timeLimit + 1)
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
