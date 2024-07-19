package transformers

import (
	"testing"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateState(t *testing.T) {
	res := generateRandomState()

	assert.IsType(t, "", res, "The returned state should be a string")

	stateExists := false
	for _, address := range transformers_dataset.Addresses {
		if address.State == res {
			stateExists = true
			break
		}
	}

	assert.True(t, stateExists, "The generated state should exist in the addresses.go file")
}

func Test_StateTransformer(t *testing.T) {
	mapping := `root = generate_state()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the state transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned state should be a string")

	stateExists := false
	for _, address := range transformers_dataset.Addresses {
		if address.State == res {
			stateExists = true
			break
		}
	}

	assert.True(t, stateExists, "The generated state should exist in the addresses.go file")
}
