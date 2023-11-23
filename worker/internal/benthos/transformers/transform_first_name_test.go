package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_TransformFirstNamePreserveLengthTrue(t *testing.T) {

	name := "evis"

	res, err := TransformFirstName(name, true)

	fmt.Println("tes", res)

	assert.NoError(t, err)
	assert.Equal(t, len(name), len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_TransformFirstNamePreserveLengthTrueOOBValue(t *testing.T) {

	name := "hiuifuwenfiuwefniuefnw"

	res, err := TransformFirstName(name, true)

	fmt.Println("tes", res)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_TransformFirstNamePreserveLengthFalse(t *testing.T) {

	name := "evis"

	res, err := TransformFirstName(name, false)

	assert.NoError(t, err)
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateFirstNamePreserveLengthTrue(t *testing.T) {

	name := "chrissy"

	res, err := GenerateRandomFirstNameWithLength(name)

	assert.NoError(t, err)
	assert.Equal(t, len(name), len(res), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_FirstNameTransformer(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = transform_first_name(%q,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated first name must be as long as input first name")
}
