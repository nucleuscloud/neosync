package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_TransformLastNamePreserveLengthTrue(t *testing.T) {

	name := "evis"
	expectedLength := 4

	res, err := TransformLastName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(*res), "The last name output should be the same length as the input")
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_TransformLastNamePreserveLengthTrueOOBValue(t *testing.T) {

	name := "hiuifuwenfiuwefniuefnw"

	res, err := TransformLastName(name, true)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(*res), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformLastNamePreserveLengthFalse(t *testing.T) {

	name := "evis"

	res, err := TransformLastName(name, false)

	assert.NoError(t, err)
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_GenerateLastNamePreserveLengthTrue(t *testing.T) {

	name := "chrissy"
	expectedLength := 7

	res, err := GenerateRandomLastNameWithLength(name)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, len(res), "The last name output should be the same length as the input")
	assert.IsType(t, "", res, "The last name should be a string")
}

func Test_LastNameTransformer(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q,preserve_length:true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(testVal), "Generated last name must be as long as input last name")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}

}

func Test_TransformLastNameTransformerWithEmptyValue(t *testing.T) {

	nilName := ""
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q,preserve_length:true)`, nilName)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
