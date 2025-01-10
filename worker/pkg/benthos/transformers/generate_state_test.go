package transformers

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateState(t *testing.T) {
	res, err := generateRandomState(rng.New(time.Now().UnixMilli()), false)
	assert.NoError(t, err)
	assert.IsType(t, "", res, "The returned state should be a string")
}

func Test_GenerateStateCodeLength(t *testing.T) {
	res, err := generateRandomState(rng.New(time.Now().UnixMilli()), false)
	assert.NoError(t, err)
	assert.IsType(t, "", res, "The returned state should be a string")
	assert.Len(t, res, 2)
}

func Test_GenerateStateCodeFullName(t *testing.T) {
	res, err := generateRandomState(rng.New(time.Now().UnixMilli()), true)
	assert.NoError(t, err)
	assert.IsType(t, "", res, "The returned state should be a string")
	assert.True(t, len(res) > 2)
}

func Test_StateTransformer(t *testing.T) {
	mapping := `root = generate_state(generate_full_name:false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the state transformer")
	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func Test_StateTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_state()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the state transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
