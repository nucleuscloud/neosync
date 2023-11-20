package neosync_benthos_transformers_functions

import (
	"encoding/json"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestGenerateZipcode(t *testing.T) {

	res, err := GenerateRandomZipcode()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.Zipcode, res, "The returned zipcode should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	zipcodeExists := false
	for _, address := range addresses {
		if address.Zipcode == res {
			zipcodeExists = true
			break
		}
	}

	assert.True(t, zipcodeExists, "The generated zipcode should exist in the addresses array")
}

func TestZipcodeTransformer(t *testing.T) {
	mapping := `root = zipcodetransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the zipcode transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned zipcode should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	zipcodeExists := false
	for _, address := range addresses {
		if address.Zipcode == res {
			zipcodeExists = true
			break
		}
	}

	assert.True(t, zipcodeExists, "The generated city should exist in the addresses array")
}
