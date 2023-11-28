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
	assert.Equal(t, len(name), len(*res), "The full name output should be the same length as the input")
	assert.IsType(t, "", *res, "The full name should be a string")
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

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {

		assert.Equal(t, len(*resStr), len(testVal), "Generated full name must be as long as input full name")

	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}

}

func Test_TransformFullNamelTransformerWithEmptyValue(t *testing.T) {

	nilName := ""
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true)`, nilName)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
