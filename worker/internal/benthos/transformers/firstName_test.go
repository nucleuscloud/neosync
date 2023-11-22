package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateFirstName(t *testing.T) {

	name := "evis"
	expectedLength := 4

	res, err := GenerateFirstName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func Test_GenerateFirstNamePreserveLengthTrue(t *testing.T) {

	name := "evis"
	expectedLength := 4

	res, err := GenerateFirstNameWithLength(name)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func Test_GenerateFirstNamePreserveLengthFalse(t *testing.T) {
	res, err := GenerateFirstNameWithRandomLength()

	assert.NoError(t, err)
	assert.Greater(t, len(res), 0, "The first name should be more than 0 characters")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func Test_FirstNameTransformerWithValue(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = firstnametransformer(%q,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated first name must be as long as input first name")
}

func Test_FirstNameTransformerWithNoValue(t *testing.T) {
	mapping := `root = firstnametransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, res.(string), "", "Generated first name must be a string")
}
