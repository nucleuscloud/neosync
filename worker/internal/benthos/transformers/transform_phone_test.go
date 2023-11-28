package transformers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testPhoneNumberHyphens = "183-849-2831"
var testPhoneNumberLong = "183849232831"
var testPhoneNumberNoHyphens = "1838492831"

func Test_GeneratePhoneNumberPreserveLengthHyphensError(t *testing.T) {

	_, err := TransformPhoneNumber(testPhoneNumberLong, true, true)

	assert.Error(t, err, "Can only preserve the length and include hyphens if the input phone number is 10 digits")

}

func Test_GeneratePhoneNumberPreserveLengthTrueHyphensTrue(t *testing.T) {

	res, err := TransformPhoneNumber(testPhoneNumberNoHyphens, true, true)

	assert.NoError(t, err)
	assert.Equal(t, len(testPhoneNumberHyphens), len(*res), "The actual value should be 10 digits long")
	assert.True(t, strings.Contains(*res, "-"), "The expecter value should have hyphens")

}

func Test_GeneratePhoneNumberPreserveLengthTrueHyphensFalse(t *testing.T) {

	res, err := TransformPhoneNumber(testPhoneNumberHyphens, true, false)

	assert.NoError(t, err)
	assert.False(t, strings.Contains(*res, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")
	assert.Equal(t, len(testPhoneNumberNoHyphens), len(*res), "The actual value should be 10 digits long")

}

func Test_GeneratePhoneNumberPreserveLengthFalseHyphensTrue(t *testing.T) {

	res, err := TransformPhoneNumber(testPhoneNumberHyphens, false, true)

	assert.NoError(t, err)
	assert.True(t, strings.Contains(*res, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")
	assert.Equal(t, len(testPhoneNumberHyphens), len(*res), "The actual value should be 10 digits long")

}

func Test_GeneratePhoneNumberPreserveLengthFalseHyphensFalse(t *testing.T) {

	res, err := TransformPhoneNumber(testPhoneNumberHyphens, false, false)

	assert.NoError(t, err)
	assert.False(t, strings.Contains(*res, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")
	assert.Equal(t, len(testPhoneNumberNoHyphens), len(*res), "The actual value should be 10 digits long")

}

func Test_PhoneNumberTransformerWithValue(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_phone(value:%q, preserve_length:true, include_hyphens:false,)`, testPhoneNumberNoHyphens)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(testPhoneNumberNoHyphens), "Generated phone number must be the same length as the input phone number")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}

}

func Test_TransformPhoneTransformerWithEmptyValue(t *testing.T) {

	nilNum := ""
	mapping := fmt.Sprintf(`root = transform_phone(value:%q, preserve_length:true, include_hyphens:false,)`, nilNum)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
