package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessLastNamePreserveLengthTrue(t *testing.T) {

	name := "jill"
	expectedLength := 4

	res, err := ProcessLastName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func TestProcessLastNamePreserveLengthFalse(t *testing.T) {

	name := "john"

	res, err := ProcessLastName(name, false)

	assert.NoError(t, err)
	assert.Greater(t, len(res), 0, "The first name should be more than 0 characters")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func TestLastNameTransformer(t *testing.T) {
	mapping := `root = this.lastnametransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	testVal := "johnson"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated last name must be as long as input last name")
}
