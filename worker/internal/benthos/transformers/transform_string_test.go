package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testStringValue = "hello"

func Test_TransformStringPreserveLengthTrue(t *testing.T) {

	res, err := TransformString(testStringValue, true)

	assert.NoError(t, err)
	assert.Equal(t, len(testStringValue), len(res), "The output string should be as long as the input string")
}

func Test_TransformStringPreserveLengthFalse(t *testing.T) {

	res, err := TransformString(testStringValue, false)

	assert.NoError(t, err)
	assert.Equal(t, defaultStrLength, len(res), "The output string should be as long as the input string")
}

func Test_TransformStringTransformer(t *testing.T) {

	mapping := fmt.Sprintf(`root = transform_string(%q,true)`, testStringValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, len(testStringValue), len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, "", "The actual value type should be a string")
}
