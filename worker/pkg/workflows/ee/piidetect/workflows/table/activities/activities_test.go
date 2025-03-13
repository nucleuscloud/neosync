package piidetect_table_activities

import (
	"bytes"
	"encoding/gob"
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

	mockOpenAIClient.EXPECT().New(mock.Anything, mock.Anything).Return(&openai.ChatCompletion{
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
				Regex: &RegexPiiDetectReport{
					Category: PiiCategoryContact,
				},
				LLM: &LLMPiiDetectReport{
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

func TestRecords_ToAssociativeArray(t *testing.T) {
	t.Run("empty records", func(t *testing.T) {
		records := Records{}
		maxRecords := uint(5)
		result := records.toAssociativeArray(maxRecords)
		assert.Nil(t, result)
	})

	t.Run("single record", func(t *testing.T) {
		records := Records{
			{"name": "John", "email": "john@example.com"},
		}
		maxRecords := uint(5)
		result := records.toAssociativeArray(maxRecords)
		expected := map[string][]any{
			"name":  {"John"},
			"email": {"john@example.com"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("multiple records", func(t *testing.T) {
		records := Records{
			{"name": "John", "email": "john@example.com"},
			{"name": "Jane", "email": "jane@example.com"},
			{"name": "Bob", "email": "bob@example.com"},
		}
		maxRecords := uint(5)
		result := records.toAssociativeArray(maxRecords)
		expected := map[string][]any{
			"name":  {"John", "Jane", "Bob"},
			"email": {"john@example.com", "jane@example.com", "bob@example.com"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("limit records", func(t *testing.T) {
		records := Records{
			{"name": "John", "email": "john@example.com"},
			{"name": "Jane", "email": "jane@example.com"},
			{"name": "Bob", "email": "bob@example.com"},
			{"name": "Alice", "email": "alice@example.com"},
			{"name": "Tom", "email": "tom@example.com"},
		}
		maxRecords := uint(2)
		result := records.toAssociativeArray(maxRecords)
		expected := map[string][]any{
			"name":  {"John", "Jane"},
			"email": {"john@example.com", "jane@example.com"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("skip empty maps", func(t *testing.T) {
		records := Records{
			{"name": "John", "email": map[string]any{}},
			{"name": "Jane", "email": "jane@example.com"},
		}
		maxRecords := uint(5)
		result := records.toAssociativeArray(maxRecords)
		expected := map[string][]any{
			"name":  {"John", "Jane"},
			"email": {"jane@example.com"},
		}
		assert.Equal(t, expected, result)
	})
}

func TestRecords_ToPromptString(t *testing.T) {
	t.Run("empty records", func(t *testing.T) {
		records := Records{}
		maxRecords := uint(5)
		result, err := records.ToPromptString(maxRecords)
		assert.NoError(t, err)
		assert.Equal(t, "{}", result)
	})

	t.Run("valid records", func(t *testing.T) {
		records := Records{
			{"name": "John", "email": "john@example.com"},
			{"name": "Jane", "email": "jane@example.com"},
		}
		maxRecords := uint(5)
		result, err := records.ToPromptString(maxRecords)
		assert.NoError(t, err)
		assert.Equal(t, `{"email":["john@example.com","jane@example.com"],"name":["John","Jane"]}`, result)
	})

	t.Run("limit records", func(t *testing.T) {
		records := Records{
			{"name": "John", "email": "john@example.com"},
			{"name": "Jane", "email": "jane@example.com"},
			{"name": "Bob", "email": "bob@example.com"},
		}
		maxRecords := uint(1)
		result, err := records.ToPromptString(maxRecords)
		assert.NoError(t, err)
		assert.Equal(t, `{"email":["john@example.com"],"name":["John"]}`, result)
	})
}

func Test_sampleDataStream(t *testing.T) {
	t.Run("Send and GetRecords", func(t *testing.T) {
		stream := sampleDataCollector()

		// Create test data
		record1 := map[string]any{"name": "John", "email": "john@example.com"}
		record2 := map[string]any{"name": "Jane", "email": "jane@example.com"}

		// Encode records
		var buf1, buf2 bytes.Buffer
		encoder1 := gob.NewEncoder(&buf1)
		encoder2 := gob.NewEncoder(&buf2)

		err := encoder1.Encode(record1)
		require.NoError(t, err)

		err = encoder2.Encode(record2)
		require.NoError(t, err)

		// Send records to stream
		err = stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{
			RowBytes: buf1.Bytes(),
		})
		require.NoError(t, err)

		err = stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{
			RowBytes: buf2.Bytes(),
		})
		require.NoError(t, err)

		// Get records from stream
		records := stream.GetRecords()

		// Verify records
		assert.Len(t, records, 2)
		assert.Equal(t, record1["name"], records[0]["name"])
		assert.Equal(t, record1["email"], records[0]["email"])
		assert.Equal(t, record2["name"], records[1]["name"])
		assert.Equal(t, record2["email"], records[1]["email"])

		// Verify deep copy
		records[0]["name"] = "Modified"
		assert.NotEqual(t, "Modified", stream.records[0]["name"])
	})

	t.Run("Send with invalid data", func(t *testing.T) {
		stream := sampleDataCollector()

		// Send invalid data
		err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{
			RowBytes: []byte("invalid gob data"),
		})

		assert.Error(t, err)
	})
}
