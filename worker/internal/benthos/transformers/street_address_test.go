package neosync_transformers

import (
	"encoding/json"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessStreetAddress(t *testing.T) {

	res, err := GenerateRandomStreetAddress()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	streetAddressExosts := false
	for _, address := range addresses {
		if address.Address1 == res {
			streetAddressExosts = true
			break
		}
	}

	assert.True(t, streetAddressExosts, "The generated street address should exist in the addresses array")
}

func TestStreetAddressTransformer(t *testing.T) {
	mapping := `root = streetaddresstransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	streetAddressExosts := false
	for _, address := range addresses {
		if address.Address1 == res {
			streetAddressExosts = true
			break
		}
	}

	assert.True(t, streetAddressExosts, "The generated street address should exist in the addresses array")
}
