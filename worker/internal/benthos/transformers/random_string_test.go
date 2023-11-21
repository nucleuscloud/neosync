package neosync_transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestRandomStringPreserveLengthTrue(t *testing.T) {

	val := "hellothe"
	expectedLength := 8

	res, err := GenerateRandomString(val, true, -1)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")
}

func TestRandomStringPreserveLengthFalse(t *testing.T) {

	val := "hello"
	expectedLength := 10

	res, err := GenerateRandomString(val, false, -1)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be a default 10 characters long")

}

func TestRandomStringPreserveLengthFalseStrLength(t *testing.T) {

	val := "hello"
	expectedLength := 14

	res, err := GenerateRandomString(val, false, int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")

}

func TestRandomStringTransformerWithValue(t *testing.T) {
	testVal := "testte"
	mapping := fmt.Sprintf(`root = randomstringtransformer(%q, false, 6)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, len(testVal), len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, testVal)
}

func TestRandomStringTransformerWithNoValue(t *testing.T) {
	mapping := `root = randomstringtransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, 10, len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, "")
}

func TestRandomStringTransformerWithNoValueAndLength(t *testing.T) {
	mapping := `root = randomstringtransformer(str_length:6)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, 6, len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, "")
}
