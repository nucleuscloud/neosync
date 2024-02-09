package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var name = "evis"
var maxCharacterLimit = int64(20)

func Test_TranformFirstNameEmptyName(t *testing.T) {
	emptyName := ""

	res, err := TransformFirstName(emptyName, true, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Nil(t, res, "The response should be nil")
}

func Test_TransformFirstNamePreserveLengthTrue(t *testing.T) {
	nameLength := int64(len(name))

	res, err := TransformFirstName(name, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameInLengthRangeMinAndMaxSame(t *testing.T) {
	nameLength := int64(len(name))

	res, err := GenerateRandomFirstNameInLengthRange(nameLength, nameLength)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(res)), "The first name output should be the same length as the input")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameInLengthRangeMinAndMaxSameTooShort(t *testing.T) {
	nameLength := int64(len("a"))

	res, err := GenerateRandomFirstNameInLengthRange(nameLength, nameLength)

	assert.NoError(t, err)
	assert.Equal(t, len(res), 2, "The length of the name should be 2")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameInLengthRangeMinAndMaxSameTooLong(t *testing.T) {
	nameLength := int64(len("wkepofkwepofe"))

	res, err := GenerateRandomFirstNameInLengthRange(nameLength, nameLength)

	assert.NoError(t, err)
	assert.Equal(t, len(res), 12, "The length of the name should be 12")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_TransformFirstNamePreserveLengthFalse(t *testing.T) {
	res, err := TransformFirstName(name, false, maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(*res) >= int(minNameLength), "The name should be greater than the min length name")
	assert.True(t, len(*res) <= int(maxCharacterLimit), "The name should be less than the max character limit")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameInLengthRange(t *testing.T) {
	res, err := GenerateRandomFirstNameInLengthRange(int64(len(name)), maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(res) >= int(minNameLength), "The name should be greater than the min length name")
	assert.True(t, len(res) <= int(maxCharacterLimit), "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameInLengthRangeMaxCharLimitMedum(t *testing.T) {
	// tests where we have a low max char limit and we want to create a name that will fit in that eact max char limit

	res, err := GenerateRandomFirstNameInLengthRange(minNameLength, int64(8))

	assert.NoError(t, err)
	assert.True(t, len(res) >= int(minNameLength), "The name should be greater than the min length name")
	assert.True(t, len(res) <= int(maxCharacterLimit), "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFirstNameInLengthRangeLowMaxCharLimit(t *testing.T) {
	// tests where we have a very low max char limit and we want to create a name that will fit in that eact max char limit

	res, err := GenerateRandomFirstNameInLengthRange(minNameLength, int64(1))

	assert.NoError(t, err)
	assert.True(t, len(res) == 1, "The name should be greater than the min length name")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_FirstNameTransformer(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q,preserve_length:true,max_length:%d)`, testVal, maxCharacterLimit)
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
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q,preserve_length:true,max_length:%d)`, nilName, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
