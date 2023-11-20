package neosync_benthos_transformers_functions

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestGenerateFullAddress(t *testing.T) {

	res, err := GenerateRandomFullAddress()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
}

func TestFullAddressTransformer(t *testing.T) {
	mapping := `root = fulladdresstransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
}
