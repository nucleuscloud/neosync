package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var fullName = "john smith"

func Test_TranformFullNameEmptyName(t *testing.T) {
	emptyName := ""

	res, err := TransformFullName(emptyName, true, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Nil(t, res, "The response should be nil")
}

func Test_TransformFullNamePreserveLengthTrue(t *testing.T) {
	nameLength := int64(len(fullName))

	res, err := TransformFullName(fullName, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFullNameMaxLengthBetween12And5(t *testing.T) {
	res, err := TransformFullName(fullName, false, 10)

	assert.NoError(t, err)
	assert.True(t, len(*res) >= 6, "The name should be greater than the min length name")
	assert.True(t, len(*res) <= 10, "The name should be less than the max character limit")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFullNameMaxLengthLessThan5(t *testing.T) {
	res, err := TransformFullName(fullName, false, 4)
	assert.NoError(t, err)
	assert.Equal(t, len(*res), 4, "The name should be greater than the min length name")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_GenerateullNamePreserveLengthFalse(t *testing.T) {
	res, err := TransformFullName(fullName, false, maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(*res) >= 6, "The name should be greater than the min length name")
	assert.True(t, len(*res) <= 13, "The name should be less than the max character limit")
	assert.IsType(t, "", *res, "The full name should be a string")
}

func Test_FullNameTransformerWithValue(t *testing.T) {
	fn := "john smith"
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true,max_length:%d)`, fn, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(fn)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(fn), "Generated full name must be as long as input full name")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformFullNamelTransformerWithEmptyValue(t *testing.T) {
	nilName := ""
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true,max_length:%d)`, nilName, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
