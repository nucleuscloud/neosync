package neosync_transformers

import (
	"encoding/json"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestGenerateState(t *testing.T) {

	res, err := GenerateRandomState()

	assert.NoError(t, err)
	assert.IsType(t, Address{}.Zipcode, res, "The returned state should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	stateExists := false
	for _, address := range addresses {
		if address.State == res {
			stateExists = true
			break
		}
	}

	assert.True(t, stateExists, "The generated state should exist in the addresses array")
}

func TestStateTransformer(t *testing.T) {
	mapping := `root = statetransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the state transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned state should be a string")

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	stateExists := false
	for _, address := range addresses {
		if address.State == res {
			stateExists = true
			break
		}
	}

	assert.True(t, stateExists, "The generated state should exist in the addresses array")
}
