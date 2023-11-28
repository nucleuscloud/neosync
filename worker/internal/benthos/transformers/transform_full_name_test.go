package transformers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateFullNamePreserveLengthTrue(t *testing.T) {

	name := "john doe"

	res, err := GenerateFullName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, len(name), len(res), "The full name output should be the same length as the input")
	assert.IsType(t, "", res, "The full name should be a string")
}

func Test_GenerateullNamePreserveLengthFalse(t *testing.T) {

	res, err := GenerateFullNameWithRandomLength()

	assert.NoError(t, err)
	assert.Equal(t, len(strings.Split(res, " ")), 2, "The full name should be more than 0 characters")
	assert.IsType(t, "", res, "The full name should be a string")
}

func Test_FullNameTransformerWithValue(t *testing.T) {
	testVal := "john smith"
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated full name must be as long as input full name")
}
