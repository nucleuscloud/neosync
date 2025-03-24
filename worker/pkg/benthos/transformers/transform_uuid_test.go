package transformers

import (
	"fmt"
	"testing"
	"time"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_TransformUuidRandom(t *testing.T) {
	randomUuid := generateUuid(true)

	res := transformUuid(rng.New(time.Now().UnixNano()), randomUuid)

	assert.IsType(t, "", *res)

	assert.NotEqual(t, res, randomUuid, "The input UUID and the output UUID should be different")
	assert.True(t, isValidUuid(*res), "The UUID should have the right format and be valid")
}

func Test_TransformUuidSeeded(t *testing.T) {
	randomUuid := generateUuid(true)

	var checkVars []string

	//checks that the output UUID is the same everytime for given the same input since we're assigning it a specific seed value
	for i := 0; i < 5; i++ {
		randomizer := rng.New(1)
		res := transformUuid(randomizer, randomUuid)
		checkVars = append(checkVars, *res)
	}

	val := transformer_utils.ToSet(checkVars)
	assert.Len(t, val, 1, "The set should only contain one value")
}

func TestUUIDTransformerMapping(t *testing.T) {
	gen := generateUuid(true)
	mapping := fmt.Sprintf(`root = transform_uuid(value:%q)`, gen)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the uuid transformer")

	res, err := ex.Query(nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
