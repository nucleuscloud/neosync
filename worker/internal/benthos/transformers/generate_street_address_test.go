package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateStreetAddress(t *testing.T) {

	res := GenerateRandomStreetAddress()

	assert.IsType(t, "", res, "The returned street address should be a string")

	streetAddressExosts := false
	for _, address := range transformers_dataset.Addresses {
		if address.Address1 == res {
			streetAddressExosts = true
			break
		}
	}

	assert.True(t, streetAddressExosts, "The generated street address should exist in the addresses array")
}

func Test_StreetAddressTransformer(t *testing.T) {
	mapping := `root = generate_street_address()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")

	streetAddressExosts := false
	for _, address := range transformers_dataset.Addresses {
		if address.Address1 == res {
			streetAddressExosts = true
			break
		}
	}

	assert.True(t, streetAddressExosts, "The generated street address should exist in the addresses array")
}
