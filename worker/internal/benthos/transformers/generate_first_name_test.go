package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFirstName(t *testing.T) {
	res, err := GenerateRandomFirstName()

	assert.NoError(t, err)
	assert.Greater(t, len(res), 0, "The first name should be more than 0 characters")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameTransformer(t *testing.T) {
	mapping := `root = generate_first_name()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res.(string)), 2, "The first name should be more than 0 characters")
	assert.Less(t, len(res.(string)), 13, "The first name should be more than 0 characters")
	assert.IsType(t, "", res, "The first name should be a string")
}
