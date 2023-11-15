package http_client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewWithHeaders(t *testing.T) {
	client := NewWithHeaders(map[string]string{
		"Foo": "Bar",
		"Bar": "Baz",
	})
	assert.NotNil(t, client)
}

func Test_headerTransport_RoundTrip(t *testing.T) {
	mockRt := new(mockRoundTripper)
	mockRt.On("RoundTrip", mock.Anything).Return(&http.Response{}, nil)

	transport := &headerTransport{
		Transport: mockRt,
		Headers:   map[string]string{"Foo": "Bar", "Bar": "Baz"},
	}
	headers := http.Header{}
	//nolint:bodyclose
	resp, err := transport.RoundTrip(&http.Request{
		Header: headers,
	})
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(
		t,
		headers,
		http.Header{"Foo": []string{"Bar"}, "Bar": []string{"Baz"}},
	)
}

type mockRoundTripper struct {
	mock.Mock
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
