package neosync_benthos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuildBenthosTable(t *testing.T) {
	assert.Equal(t, BuildBenthosTable("public", "users"), "public.users", "Joins schema and table with a dot")
	assert.Equal(t, BuildBenthosTable("", "users"), "users", "Handles an empty schema")
}
