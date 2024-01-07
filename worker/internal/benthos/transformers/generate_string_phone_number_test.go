package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testStringPhone = "6183849282"
var testStringPhoneHyphens = "618-384-9282"

func Test_GeneratePhoneNumberHyphens(t *testing.T) {

	res, err := GenerateRandomPhoneNumber(true)

	assert.NoError(t, err)
	assert.Equal(t, len(testStringPhoneHyphens), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_GeneratePhoneNumberNoHyphens(t *testing.T) {

	res, err := GenerateRandomPhoneNumber(false)

	assert.NoError(t, err)
	assert.Equal(t, len(testStringPhone), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_GenerateRandomPhoneNumberHyphens(t *testing.T) {

	res, err := GenerateRandomPhoneNumberHyphens()
	assert.NoError(t, err)
	assert.Equal(t, len(testStringPhoneHyphens), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_GenerateRandomPhoneNumberNoHyphens(t *testing.T) {

	res, err := GenerateRandomPhoneNumberNoHyphens()

	assert.NoError(t, err)
	assert.Equal(t, len(testStringPhone), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_PhoneNumberTransformer(t *testing.T) {
	mapping := `root = generate_string_phone_number(include_hyphens:false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testStringPhone), "Generated phone number must be the same length as the input phone number")
}
