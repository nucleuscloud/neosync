package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateUsername(t *testing.T) {

	res, err := GenerateUsername()
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The expected username should have a valid username")
}

func Test_RealisticmailTransformer(t *testing.T) {
	mapping := `root = generate_username()`
	ex, err := bloblang.Parse(mapping)

	assert.NoError(t, err, "failed to parse the realistic username transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The expected username should have a valid username")
}
