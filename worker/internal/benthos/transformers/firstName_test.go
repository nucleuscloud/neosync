package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessFirstNamePreserveLengthTrue(t *testing.T) {

	name := "evis"
	expectedLength := 4

	res, err := ProcessFirstName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func TestProcessFirstNamePreserveLengthFalse(t *testing.T) {

	name := "evis"

	res, err := ProcessFirstName(name, false)

	assert.NoError(t, err)
	assert.Greater(t, len(res), 0, "The first name should be more than 0 characters")
	assert.IsType(t, "", res, "The first name should be a string") // Check if the result is a string
}

func TestFirstNameTransformer(t *testing.T) {
	mapping := `root = this.firstnametransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	testVal := "evis"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated first name must be as long as input first name")
}
