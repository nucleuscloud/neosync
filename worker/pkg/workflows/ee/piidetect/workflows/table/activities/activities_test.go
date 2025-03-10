package piidetect_table_activities

import (
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/openai/openai-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

func Test_New(t *testing.T) {
	mockConnClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockOpenAIClient := NewMockOpenAiCompletionsClient(t)
	mockConnDataBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)

	activities := New(mockConnClient, mockOpenAIClient, mockConnDataBuilder, mockJobClient)
	assert.NotNil(t, activities)
}

func Test_GetColumnData_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	mockConnClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockOpenAIClient := NewMockOpenAiCompletionsClient(t)
	mockConnData := connectiondata.NewMockConnectionDataService(t)
	mockConnDataBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)

	mockConnDataBuilder.EXPECT().NewDataConnection(mock.Anything, mock.Anything).Return(mockConnData, nil)
	mockConnData.EXPECT().GetTableSchema(mock.Anything, "public", "users").Return([]*mgmtv1alpha1.DatabaseColumn{
		{
			Schema:     "public",
			Table:      "users",
			Column:     "id",
			DataType:   "uuid",
			IsNullable: "NO",
		},
		{
			Schema:     "public",
			Table:      "users",
			Column:     "email",
			DataType:   "varchar",
			IsNullable: "YES",
		},
	}, nil)

	mockConnClient.EXPECT().GetConnection(mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
			Connection: &mgmtv1alpha1.Connection{
				Id: "test-conn",
			},
		}), nil)

	activities := New(mockConnClient, mockOpenAIClient, mockConnDataBuilder, mockJobClient)

	env.RegisterActivity(activities)

	val, err := env.ExecuteActivity(activities.GetColumnData, &GetColumnDataRequest{
		ConnectionId: "test-conn",
		TableSchema:  "public",
		TableName:    "users",
	})
	require.NoError(t, err)
	res := &GetColumnDataResponse{}
	err = val.Get(res)
	require.NoError(t, err)

	assert.NotNil(t, res)
	assert.Len(t, res.ColumnData, 2)
	assert.Equal(t, "id", res.ColumnData[0].Column)
	assert.Equal(t, "uuid", res.ColumnData[0].DataType)
	assert.False(t, res.ColumnData[0].IsNullable)
	assert.Equal(t, "email", res.ColumnData[1].Column)
	assert.Equal(t, "varchar", res.ColumnData[1].DataType)
	assert.True(t, res.ColumnData[1].IsNullable)
}

func Test_DetectPiiRegex_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	mockConnClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockOpenAIClient := NewMockOpenAiCompletionsClient(t)
	mockConnDataBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)

	activities := New(mockConnClient, mockOpenAIClient, mockConnDataBuilder, mockJobClient)

	env.RegisterActivity(activities)

	val, err := env.ExecuteActivity(activities.DetectPiiRegex, &DetectPiiRegexRequest{
		ColumnData: []*ColumnData{
			{
				Column:     "email",
				DataType:   "varchar",
				IsNullable: true,
			},
			{
				Column:     "created_at",
				DataType:   "timestamp",
				IsNullable: false,
			},
		},
	})
	require.NoError(t, err)
	res := &DetectPiiRegexResponse{}
	err = val.Get(res)
	require.NoError(t, err)

	assert.NotNil(t, res)
	assert.Len(t, res.PiiColumns, 1)
	assert.Equal(t, PiiCategoryContact, res.PiiColumns["email"])
}

func Test_DetectPiiLLM_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	mockConnClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockOpenAIClient := NewMockOpenAiCompletionsClient(t)
	mockConnData := connectiondata.NewMockConnectionDataService(t)
	mockConnDataBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)

	mockConnDataBuilder.EXPECT().NewDataConnection(mock.Anything, mock.Anything).Return(mockConnData, nil)
	mockConnData.EXPECT().SampleData(mock.Anything, mock.Anything, "public", "users", uint(5)).Return(nil)

	mockConnClient.EXPECT().GetConnection(mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id: "test-conn",
		},
	}), nil)

	mockOpenAIClient.EXPECT().New(mock.Anything, mock.Anything).Return(openai.ChatCompletion{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: "{\"output\": [{\"field_name\": \"email\", \"category\": \"contact\", \"confidence\": 0.95}]}",
				},
			},
		},
	}, nil)

	activities := New(mockConnClient, mockOpenAIClient, mockConnDataBuilder, mockJobClient)

	env.RegisterActivity(activities)

	val, err := env.ExecuteActivity(activities.DetectPiiLLM, &DetectPiiLLMRequest{
		TableSchema:  "public",
		TableName:    "users",
		ShouldSample: true,
		ConnectionId: "test-conn",
		UserPrompt:   "Detect PII in this table",
		ColumnData: []*ColumnData{
			{
				Column:     "email",
				DataType:   "varchar",
				IsNullable: true,
			},
		},
	})
	require.NoError(t, err)
	res := &DetectPiiLLMResponse{}
	err = val.Get(res)
	require.NoError(t, err)

	assert.NotNil(t, res)
	assert.Len(t, res.PiiColumns, 1)
	report, ok := res.PiiColumns["email"]
	assert.True(t, ok)
	assert.Equal(t, PiiCategoryContact, report.Category)
	assert.Equal(t, 0.95, report.Confidence)
}

func Test_SaveTablePiiDetectReport_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestActivityEnvironment()

	mockConnClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockOpenAIClient := NewMockOpenAiCompletionsClient(t)
	mockConnDataBuilder := connectiondata.NewMockConnectionDataBuilder(t)
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)

	mockJobClient.EXPECT().SetRunContext(mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)

	activities := New(mockConnClient, mockOpenAIClient, mockConnDataBuilder, mockJobClient)

	env.RegisterActivity(activities)

	parentRunId := "test-run"
	val, err := env.ExecuteActivity(activities.SaveTablePiiDetectReport, &SaveTablePiiDetectReportRequest{
		ParentRunId: &parentRunId,
		AccountId:   "test-account",
		TableSchema: "public",
		TableName:   "users",
		Report: map[string]CombinedPiiDetectReport{
			"email": {
				Regex: &[]PiiCategory{PiiCategoryContact}[0],
				LLM: &PiiDetectReport{
					Category:   PiiCategoryContact,
					Confidence: 0.95,
				},
			},
		},
	}, nil)

	require.NoError(t, err)
	res := &SaveTablePiiDetectReportResponse{}
	err = val.Get(res)
	require.NoError(t, err)

	assert.NotNil(t, res)
	assert.NotNil(t, res.Key)
}
