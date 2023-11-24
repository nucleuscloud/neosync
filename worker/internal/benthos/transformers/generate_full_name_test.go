package transformers

import (
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GeneratRandomFullName(t *testing.T) {

	res, err := GenerateRandomFullName()

	assert.NoError(t, err)
	assert.Equal(t, len(strings.Split(res, " ")), 2, "The full name should have a first and a last name")
	assert.IsType(t, "", res, "The full name should be a string")
}

func Test_GenerateFullNameTransformer(t *testing.T) {
	mapping := `root = generate_random_full_name()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, len(strings.Split(res.(string), " ")), 2, "The full name should have a first and a last name")
	assert.IsType(t, "", res, "The full name should be a string")
}
