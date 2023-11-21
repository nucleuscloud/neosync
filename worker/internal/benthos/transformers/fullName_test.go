package neosync_transformers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateFullNamePreserveLengthTrue(t *testing.T) {

	name := "john doe"
	expectedLength := 8

	res, err := GenerateFullName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The full name output should be the same length as the input")
	assert.IsType(t, "", res, "The full name should be a string")
}

func Test_GenerateullNamePreserveLengthFalse(t *testing.T) {

	res, err := GenerateFullNameWithRandomLength()

	assert.NoError(t, err)
	assert.Equal(t, len(strings.Split(res, " ")), 2, "The full name should be more than 0 characters")
	assert.IsType(t, "", res, "The full name should be a string") // Check if the result is a string
}

func Test_FullNameTransformerWithValue(t *testing.T) {
	testVal := "john smith"
	mapping := fmt.Sprintf(`root = fullnametransformer(%q,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated full name must be as long as input full name")
}

func Test_FullNameTransformerWithNoValue(t *testing.T) {
	mapping := `root = fullnametransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, res.(string), "", "Generated first name must be a string")
	assert.NotEmpty(t, res.(string))
}
