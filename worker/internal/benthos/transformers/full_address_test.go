package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateFullAddress(t *testing.T) {

	res, err := GenerateRandomFullAddress()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
}

func Test_FullAddressTransformer(t *testing.T) {
	mapping := `root = fulladdresstransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
}
