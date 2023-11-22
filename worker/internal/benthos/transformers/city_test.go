package transformers

import (
	"encoding/json"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCity(t *testing.T) {

	res, err := GenerateRandomCity()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.City, res, "The returned city should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	cityExists := false
	for _, address := range addresses {
		if address.City == res {
			cityExists = true
			break
		}
	}

	assert.True(t, cityExists, "The generated city should exist in the addresses array")
}

func TestCityTransformer(t *testing.T) {
	mapping := `root = citytransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the city transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned city should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	cityExists := false
	for _, address := range addresses {
		if address.City == res {
			cityExists = true
			break
		}
	}

	assert.True(t, cityExists, "The generated city should exist in the addresses array")
}
