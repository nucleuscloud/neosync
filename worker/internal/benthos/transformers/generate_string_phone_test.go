package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testStringPhone = "6183849282"
var testStringPhoneHyphens = "618-384-9282"
var e164TestValue = "+2393573894"

func Test_GeneratePhoneNumberE164Format(t *testing.T) {

	res, err := GenerateRandomPhoneNumber(true, false)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), ValidateE164(e164TestValue))
	assert.Equal(t, len(e164TestValue), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_GeneratePhoneNumberHyphens(t *testing.T) {

	res, err := GenerateRandomPhoneNumber(false, true)

	assert.NoError(t, err)
	assert.Equal(t, len(testStringPhoneHyphens), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_GeneratePhoneNumberNoHyphensNoE164(t *testing.T) {

	res, err := GenerateRandomPhoneNumber(false, false)

	assert.NoError(t, err)
	assert.Equal(t, len(testStringPhone), len(res), "The length of the output phone number should be the same as the input phone number")
}

func Test_GeneratePhoneNumberHyphensE164(t *testing.T) {

	_, err := GenerateRandomPhoneNumber(true, true)

	assert.Error(t, err)
}

func Test_PhoneNumberTransformer(t *testing.T) {
	mapping := `root = generate_string_phone(e164_format:false, include_hyphens:false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testStringPhone), "Generated phone number must be the same length as the input phone number")
}
