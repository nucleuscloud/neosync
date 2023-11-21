package neosync_transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateLastNamePreserveLengthTrue(t *testing.T) {

	name := "jill"
	expectedLength := 4

	res, err := GenerateLastNameWithLength(name)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func Test_GenerateLastNamePreserveLengthFalse(t *testing.T) {

	res, err := GenerateLastNameWithRandomLength()

	assert.NoError(t, err)
	assert.Greater(t, len(res), 0, "The first name should be more than 0 characters")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func Test_LastNameTransformerWithValue(t *testing.T) {
	testVal := "johnson"
	mapping := fmt.Sprintf(`root = lastnametransformer(%q,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated last name must be as long as input last name")
}

func Test_LastNameTransformerWithNoValue(t *testing.T) {
	mapping := `root = lastnametransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, res.(string), "", "Generated last name must be a string")
}
