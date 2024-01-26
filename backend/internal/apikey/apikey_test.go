package apikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewV1AccountKey(t *testing.T) {
	assert.NotEmpty(t, NewV1AccountKey())
}

func Test_v1AccountKey(t *testing.T) {
	assert.Equal(
		t,
		v1AccountKey("foo-bar"),
		"neo_at_v1_foo-bar",
	)
}

func Test_IsValidV1AccountKey(t *testing.T) {
	assert.True(
		t,
		IsValidV1AccountKey(NewV1AccountKey()),
	)
	assert.False(
		t,
		IsValidV1AccountKey(NewV1WorkerKey()),
		"worker keys should not pass as valid account keys",
	)
}

func Test_NewV1WrokerKey(t *testing.T) {
	assert.NotEmpty(t, NewV1WorkerKey())
}

func Test_v1WorkerKey(t *testing.T) {
	assert.Equal(
		t,
		v1WorkerKey("foo-bar"),
		"neo_wt_v1_foo-bar",
	)
}

func Test_IsValidV1WorkerKey(t *testing.T) {
	assert.True(
		t,
		IsValidV1WorkerKey(NewV1WorkerKey()),
	)
	assert.False(
		t,
		IsValidV1WorkerKey(NewV1AccountKey()),
		"account keys should not pass as valid worker keys",
	)
}
