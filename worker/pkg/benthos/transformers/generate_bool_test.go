package transformers

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomBool(t *testing.T) {
	res := generateRandomBool(rng.New(time.Now().UnixNano()))
	assert.IsType(t, res, false)
}

func Test_GenerateRandomBoolTransformer(t *testing.T) {
	mapping := `root = generate_bool()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random bool transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	var test bool

	assert.IsType(t, res, test, "The actual value type should be a bool.")
}
