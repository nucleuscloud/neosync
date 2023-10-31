package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestRandomStringPreserveLengthTrue(t *testing.T) {

	val := "hellothe"
	expectedLength := 8

	res, err := ProcessRandomString(val, true, -1)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")
}

func TestRandomStringPreserveLengthFalse(t *testing.T) {

	val := "hello"
	expectedLength := 10

	res, err := ProcessRandomString(val, false, -1)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be a default 10 characters long")

}

func TestRandomStringPreserveLengthFalseStrLength(t *testing.T) {

	val := "hello"
	expectedLength := 14

	res, err := ProcessRandomString(val, false, int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")

}

func TestRandomStringTransformer(t *testing.T) {
	mapping := `root = this.randomstringtransformer(true, -1)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	testVal := "testte"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Equal(t, len(testVal), len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, testVal)
}
