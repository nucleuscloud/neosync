package http_client

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_WithAuth(t *testing.T) {
	client := &http.Client{}
	client = WithBearerAuth(client, nil)
	require.NotNil(t, client)
}

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

func Test_headerTransport_RoundTrip_NilHttpHeader(t *testing.T) {
	mockRt := new(mockRoundTripper)
	mockRt.On("RoundTrip", mock.Anything).Return(&http.Response{}, nil)

	transport := &headerTransport{
		Transport: mockRt,
		Headers:   map[string]string{"Foo": "Bar", "Bar": "Baz"},
	}
	//nolint:bodyclose
	resp, err := transport.RoundTrip(&http.Request{
		Header: nil,
	})
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

type mockRoundTripper struct {
	mock.Mock
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func Test_NewWithAuth(t *testing.T) {
	client := NewWithBearerAuth(nil)
	assert.NotNil(t, client)

	empty := ""
	client = NewWithBearerAuth(&empty)
	assert.NotNil(t, client)

	token := "foo"
	client = NewWithBearerAuth(&token)
	assert.NotNil(t, client)
}

func Test_GetAuthHeaders(t *testing.T) {
	token := "foo"
	assert.Equal(
		t,
		GetBearerAuthHeaders(&token),
		map[string]string{"Authorization": "Bearer foo"},
	)
}

func TestMergeMaps(t *testing.T) {
	t.Run("merge empty maps", func(t *testing.T) {
		result := MergeMaps()
		expected := map[string]string{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("merge single map", func(t *testing.T) {
		input := map[string]string{"a": "1", "b": "2"}
		result := MergeMaps(input)
		if !reflect.DeepEqual(result, input) {
			t.Errorf("Expected %v, got %v", input, result)
		}
	})

	t.Run("merge two maps", func(t *testing.T) {
		map1 := map[string]string{"a": "1", "b": "2"}
		map2 := map[string]string{"c": "3", "d": "4"}
		result := MergeMaps(map1, map2)
		expected := map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("merge maps with overlapping keys", func(t *testing.T) {
		map1 := map[string]string{"a": "1", "b": "2"}
		map2 := map[string]string{"b": "3", "c": "4"}
		result := MergeMaps(map1, map2)
		expected := map[string]string{"a": "1", "b": "3", "c": "4"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("merge multiple maps", func(t *testing.T) {
		map1 := map[string]string{"a": "1"}
		map2 := map[string]string{"b": "2"}
		map3 := map[string]string{"c": "3"}
		result := MergeMaps(map1, map2, map3)
		expected := map[string]string{"a": "1", "b": "2", "c": "3"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("merge with nil map", func(t *testing.T) {
		map1 := map[string]string{"a": "1"}
		var nilMap map[string]string
		result := MergeMaps(map1, nilMap)
		expected := map[string]string{"a": "1"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
