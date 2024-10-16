package transformers

import (
	"testing"
	"time"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateState(t *testing.T) {
	res := generateRandomState(rng.New(time.Now().UnixMilli()), false)

	assert.IsType(t, "", res, "The returned state should be a string")

	stateExists := false
	for _, address := range transformers_dataset.States {
		if address.Code == res {
			stateExists = true
			break
		}
	}

	assert.True(t, stateExists, "The generated state should exist in the satates.go file")
}

func Test_GenerateStateCodeLength(t *testing.T) {
	res := generateRandomState(rng.New(time.Now().UnixMilli()), false)

	assert.IsType(t, "", res, "The returned state should be a string")

	stateExists := false
	for _, state := range transformers_dataset.States {
		if state.Code == res {
			stateExists = true
			break
		}
	}

	assert.Len(t, res, 2)
	assert.True(t, stateExists, "The generated state should exist in the states.go file")
}

func Test_GenerateStateCodeFullName(t *testing.T) {
	res := generateRandomState(rng.New(time.Now().UnixMilli()), true)

	assert.IsType(t, "", res, "The returned state should be a string")

	stateExists := false
	for _, state := range transformers_dataset.States {
		if state.FullName == res {
			stateExists = true
			break
		}
	}

	assert.True(t, len(res) > 2)
	assert.True(t, stateExists, "The generated state should exist in the states.go file")
}

func Test_StateTransformer(t *testing.T) {
	mapping := `root = generate_state(generate_full_name:false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the state transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned state should be a string")

	stateExists := false
	for _, state := range transformers_dataset.States {
		if state.Code == res {
			stateExists = true
			break
		}
	}

	assert.Len(t, res, 2)
	assert.True(t, stateExists, "The generated state should exist in the states.go file")
}

func Test_StateTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_state()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the state transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
