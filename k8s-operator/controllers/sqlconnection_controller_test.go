package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSha256Hash(t *testing.T) {
	input := "foo"
	assert.Equal(
		t,
		generateSha256Hash([]byte(input)),
		"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	)
}
