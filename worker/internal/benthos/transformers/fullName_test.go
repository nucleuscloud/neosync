package neosync_transformers

import (
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessFullNamePreserveLengthTrue(t *testing.T) {

	name := "john doe"
	expectedLength := 8

	res, err := ProcessFullName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The full name output should be the same length as the input")
	assert.IsType(t, "", res, "The full name should be a string")
}

func TestProcessFullNamePreserveLengthFalse(t *testing.T) {

	name := "evis drenova"

	res, err := ProcessFullName(name, false)

	assert.NoError(t, err)
	assert.Equal(t, len(strings.Split(res, " ")), 2, "The full name should be more than 0 characters")
	assert.IsType(t, "", res, "The full name should be a string") // Check if the result is a string
}

func TestFullNameTransformer(t *testing.T) {
	mapping := `root = this.fullnametransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	testVal := "john smith"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated full name must be as long as input full name")
}
