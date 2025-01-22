package loki

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GetStreamsFromResponseData(t *testing.T) {
	actual, err := GetStreamsFromResponseData(nil)
	require.Nil(t, actual)
	require.Error(t, err)

	actual, err = GetStreamsFromResponseData(&QueryResponseData{})
	require.Nil(t, actual)
	require.Error(t, err)

	actual, err = GetStreamsFromResponseData(&QueryResponseData{ResultType: ResultTypeStream})
	require.Nil(t, actual)
	require.Error(t, err)

	actual, err = GetStreamsFromResponseData(&QueryResponseData{
		ResultType: ResultTypeStream,
		Result:     Streams{{Labels: LabelSet{"foo": "bar"}, Entries: []Entry{{Line: "foo-line"}}}},
	})
	require.NotNil(t, actual)
	require.NoError(t, err)
	require.Equal(
		t,
		Streams{{Labels: LabelSet{"foo": "bar"}, Entries: []Entry{{Line: "foo-line"}}}},
		actual,
	)
}

func Test_GetEntriesFromStreams(t *testing.T) {
	time := time.Now()
	actual := GetEntriesFromStreams(Streams{
		{Entries: []Entry{{Line: "foo-line", Timestamp: time}}, Labels: LabelSet{"Attempt": "1", "willberemoved": "true"}},
		{Entries: []Entry{{Line: "bar-line", Timestamp: time}, {Line: "baz-line", Timestamp: time}}, Labels: LabelSet{"Attempt": "2"}},
	})
	require.Equal(
		t,
		[]*LabeledEntry{
			{Entry: Entry{Line: "foo-line", Timestamp: time}, Labels: LabelSet{"Attempt": "1"}},
			{Entry: Entry{Line: "bar-line", Timestamp: time}, Labels: LabelSet{"Attempt": "2"}},
			{Entry: Entry{Line: "baz-line", Timestamp: time}, Labels: LabelSet{"Attempt": "2"}},
		},
		actual,
	)
}

func Test_New_LokiClient(t *testing.T) {
	require.NotNil(t, New("foo", http.DefaultClient))
}

func Test_LokiClient_QueryRange(t *testing.T) {
	mockHttpClient := NewMockLokiHttpClient(t)
	lokiClient := New("http://fake-url", mockHttpClient)

	queryResponse := QueryResponse{
		Status: "success",
		Data:   QueryResponseData{ResultType: ResultTypeStream, Result: Streams{}},
	}
	bits, err := json.Marshal(queryResponse)
	require.NoError(t, err)

	mockHttpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{Reader: strings.NewReader(string(bits))},
	}, nil)

	resp, err := lokiClient.QueryRange(context.Background(), &QueryRangeRequest{
		Query: "foo",
	}, slog.Default())
	require.NotNil(t, resp)
	require.NoError(t, err)
	require.Equal(t, *resp, queryResponse)
}

func Test_LokiClient_QueryRange_Non200(t *testing.T) {
	mockHttpClient := NewMockLokiHttpClient(t)
	lokiClient := New("http://fake-url", mockHttpClient)

	mockHttpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 400,
		Body:       mockReadCloser{Reader: strings.NewReader("bad request")},
	}, nil)

	resp, err := lokiClient.QueryRange(context.Background(), &QueryRangeRequest{
		Query: "foo",
	}, slog.Default())
	require.Nil(t, resp)
	require.Error(t, err)
}

type mockReadCloser struct {
	io.Reader
}

func (mrc mockReadCloser) Close() error {
	return nil // No-op, as there's nothing to close in our mock
}
