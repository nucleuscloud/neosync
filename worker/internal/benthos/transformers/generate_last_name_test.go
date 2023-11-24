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

	lastName, ok := res.(string)
	assert.True(t, ok, "The result should be a string")
	assert.GreaterOrEqual(t, len(lastName), 2, "The last name should be at least 2 characters long")
	assert.Less(t, len(lastName), 13, "The last name should be less than 13 characters long")
}
