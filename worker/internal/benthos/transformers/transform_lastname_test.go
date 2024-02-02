package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_TranformLastNameEmptyName(t *testing.T) {

	emptyName := ""

	res, err := TransformLastName(emptyName, true, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Nil(t, res, "The response should be nil")
}

func Test_TransformLastNamePreserveLengthTrue(t *testing.T) {

	nameLength := int64(len(name))

	res, err := TransformLastName(name, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The last name output should be the same length as the input")
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_GenerateRandomLastNameInLengthRangeMinAndMaxSame(t *testing.T) {

	nameLength := int64(len(name))

	res, err := GenerateRandomLastNameInLengthRange(nameLength, nameLength)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(res)), "The last name output should be the same length as the input")
	assert.IsType(t, "", res, "The last name should be a string")
}

func Test_GenerateRandomLastNameInLengthRangeMinAndMaxSameTooShort(t *testing.T) {

	nameLength := int64(len("a"))

	res, err := GenerateRandomLastNameInLengthRange(nameLength, nameLength)

	assert.NoError(t, err)
	assert.Equal(t, len(res), 2, "The length of the name should be two")
	assert.IsType(t, "", res, "The last name should be a string")
}

func Test_GenerateRandomLastNameInLengthRangeMinAndMaxSameTooLong(t *testing.T) {

	nameLength := int64(len("wkepofkwepofe"))

	res, err := GenerateRandomLastNameInLengthRange(nameLength, nameLength)

	assert.NoError(t, err)
	assert.Equal(t, len(res), 12, "The length of the name should be two")
	assert.IsType(t, "", res, "The last name should be a string")
}

func Test_TransformLastNamePreserveLengthFalse(t *testing.T) {

	res, err := TransformLastName(name, false, maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(*res) >= int(minNameLength), "The name should be greater than the min length name")
	assert.True(t, len(*res) <= int(maxCharacterLimit), "The name should be less than the max character limit")
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_GenerateRandomLastNameInLengthRange(t *testing.T) {

	res, err := GenerateRandomLastNameInLengthRange(int64(len(name)), maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(res) >= int(minNameLength), "The name should be greater than the min length name")
	assert.True(t, len(res) <= int(maxCharacterLimit), "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The last name should be a string")

}

func Test_GenerateRandomLastNameInLengthRangeMaxCharLimitMedum(t *testing.T) {
	// tests where we have a low max char limit and we want to create a name that will fit in that eact max char limit

	res, err := GenerateRandomLastNameInLengthRange(minNameLength, int64(8))

	assert.NoError(t, err)
	assert.True(t, len(res) >= int(minNameLength), "The name should be greater than the min length name")
	assert.True(t, len(res) <= int(maxCharacterLimit), "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The last name should be a string")

}

func Test_GenerateRandomLastNameInLengthRangeLowMaxCharLimit(t *testing.T) {

	// tests where we have a very low max char limit and we want to create a name that will fit in that eact max char limit

	res, err := GenerateRandomLastNameInLengthRange(minNameLength, int64(1))

	assert.NoError(t, err)
	assert.True(t, len(res) == 1, "The name should be greater than the min length name")
	assert.IsType(t, "", res, "The last name should be a string")

}

func Test_LastNameTransformer(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q,preserve_length:true,max_length:%d)`, testVal, maxCharacterLimit)
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
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q,preserve_length:true,max_length:%d)`, nilName, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
