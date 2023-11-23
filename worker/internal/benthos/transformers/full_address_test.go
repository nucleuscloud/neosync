package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateFullAddress(t *testing.T) {

	res := GenerateRandomFullAddress()

	assert.IsType(t, "", res, "The returned full address should be a string")
}

func Test_FullAddressTransformer(t *testing.T) {
	mapping := `root = generate_random_full_address()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned full address should be a string")
}
