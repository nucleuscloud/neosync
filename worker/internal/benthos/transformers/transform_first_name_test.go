package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var name = "evis"

func Test_TransformFirstNamePreserveLengthTrue(t *testing.T) {

	res, err := TransformFirstName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, len(name), len(*res), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFirstNamePreserveLengthTrueOOBValue(t *testing.T) {

	name := "hiuifuwenfiuwefniuefnw"

	res, err := TransformFirstName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(*res), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFirstNamePreserveLengthFalse(t *testing.T) {

	res, err := TransformFirstName(name, false)

	assert.NoError(t, err)
	assert.IsType(t, "", *res, "The first name should be a string")
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
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q,preserve_length:true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(testVal), "Generated first name must be as long as input first name")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformFirstNameTransformerWithEmptyValue(t *testing.T) {

	nilName := ""
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q,preserve_length:true)`, nilName)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
