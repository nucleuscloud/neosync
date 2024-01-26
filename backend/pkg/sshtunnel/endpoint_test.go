package sshtunnel

import (
	"testing"

	"github.com/zeebo/assert"
)

func Test_NewEndpointWithUser(t *testing.T) {
	assert.Equal(
		t,
		NewEndpointWithUser("localhost", 5432, "nick"),
		&Endpoint{Host: "localhost", Port: 5432, User: "nick"},
	)
}

func Test_NewEndpoint(t *testing.T) {
	assert.Equal(
		t,
		NewEndpoint("localhost", 5432),
		&Endpoint{Host: "localhost", Port: 5432, User: ""},
	)
}

func Test_Endpoint_String(t *testing.T) {
	type testcase struct {
		name     string
		input    Endpoint
		expected string
	}
	tesstcases := []testcase{
		{name: "empty", input: Endpoint{}, expected: ":0"},
		{name: "host", input: Endpoint{Host: "foo"}, expected: "foo:0"},
		{name: "host+port", input: Endpoint{Host: "foo", Port: 4}, expected: "foo:4"},
		{name: "host+port+user, does not attach username", input: Endpoint{Host: "foo", Port: 4, User: "nick"}, expected: "foo:4"},
	}
	for _, tc := range tesstcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.input.String(), tc.expected)
		})
	}
}
