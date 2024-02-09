package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateSHA256Hash(t *testing.T) {
	res, err := GenerateRandomSHA256Hash()
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The actual value should be a string")
}

func Test_GenerateSHA256HashTransformer(t *testing.T) {
	mapping := `root = generate_sha256hash()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.IsType(t, "", res, "The actual value should be a string")
}
