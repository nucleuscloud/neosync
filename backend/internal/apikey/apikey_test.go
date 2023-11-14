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
}
