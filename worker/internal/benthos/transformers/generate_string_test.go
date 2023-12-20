package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomStringInRange(t *testing.T) {
	min := int64(2)
	max := int64(5)

	res, err := GenerateRandomString(min, max)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, int64(len(res)), min, "The output string should be greater than or equal to the min")
	assert.LessOrEqual(t, int64(len(res)), max, "The output string should be less than or equal to the max")

}

func Test_GenerateRandomStringTransformerWithValue(t *testing.T) {
	min := int64(2)
	max := int64(5)

	mapping := fmt.Sprintf(`root = generate_string(min:%d, max:%d)`, min, max)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the generate random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, res, "", "The actual value type should be a string")
	assert.GreaterOrEqual(t, int64(len(res.(string))), min, "The output string should be greater than or equal to the min")
	assert.LessOrEqual(t, int64(len(res.(string))), max, "The output string should be less than or equal to the max")

}
