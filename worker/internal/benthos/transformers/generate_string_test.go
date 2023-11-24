package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomStringWithLength(t *testing.T) {

	expectedLength := 5

	res, err := GenerateRandomString(int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The output string should be as long as the input string")

}

func TestRandomStringTransformerWithValue(t *testing.T) {
	length := 6
	mapping := fmt.Sprintf(`root = generate_random_string(%d)`, length)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the generate random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, length, len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, "")
}
