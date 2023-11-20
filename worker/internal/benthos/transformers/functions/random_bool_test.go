package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessRandomBool(t *testing.T) {

	res, err := GenerateRandomBool()

	assert.NoError(t, err)
	assert.IsType(t, res, false)

}

func TestRandomBoolTransformer(t *testing.T) {
	mapping := `root = randombooltransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random bool transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	var test bool

	assert.IsType(t, res, test)
}
