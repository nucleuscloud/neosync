package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomLastName(t *testing.T) {
	res, err := GenerateRandomLastName()

	assert.NoError(t, err)
	assert.Greater(t, len(res), 0, "The last name should be more than 0 characters")
	assert.IsType(t, "", res, "The last name should be a string")
}

func Test_GenerateRandomLastNameTransformer(t *testing.T) {
	mapping := `root = generate_random_last_name()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res.(string)), 2, "The last name should be more than 0 characters")
	assert.Less(t, len(res.(string)), 13, "The last name should be more than 0 characters")
	assert.IsType(t, "", res, "The last name should be a string")
}
