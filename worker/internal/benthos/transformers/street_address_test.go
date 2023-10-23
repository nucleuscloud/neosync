package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessStreetAddressPreserveLengthTrue(t *testing.T) {

	res, err := GenerateRandomStreetAddress()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
}

func TestStreetAddressTransformer(t *testing.T) {
	mapping := `root = streetaddresstransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
}
