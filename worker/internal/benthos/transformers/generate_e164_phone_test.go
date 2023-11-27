package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateE164FormatPhoneNumber(t *testing.T) {

	expectedLength := 11

	res, err := GenerateRandomE164Phone(int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.Equal(t, len(res), expectedLength+1, "The length of the output phone number should be the same as the input phone number")

}

func Test_GeneratePhoneNumberE164FormatPreserveLength(t *testing.T) {

	expectedLength := 11

	res, err := GenerateRandomE164Phone(int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.Equal(t, len(res), expectedLength+1, "The length of the output phone number should be the same as the input phone number")

}

func Test_GenerateE164PhoneNumberTransformer(t *testing.T) {
	testVal := int64(12)
	mapping := fmt.Sprintf(`root = generate_e164_number(length:%d)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res.(string)), true, "The actual value should be a valid e164 number")
	assert.Equal(t, int64(len(res.(string))), testVal+int64(1), "Generated phone number must be the same length as the input phone number")
}

func Test_ValidateE164True(t *testing.T) {

	val := "+6272636472"

	res := ValidateE164(val)

	assert.Equal(t, res, true, "The e164 number should have a plus sign at the 0th index and be 10 < x < 15 characters long.")
}

func Test_ValidateE164FalseTooLong(t *testing.T) {

	val := "627263647278439"

	res := ValidateE164(val)

	assert.Equal(t, res, false, "The e164 number should  be x < 15 characters long.")
}

func Test_ValidateE164FalseNoPlusSign(t *testing.T) {

	val := "6272636472784"

	res := ValidateE164(val)

	assert.Equal(t, res, false, "The e164 number should have a plus sign at the beginning.")
}

func Test_ValidateE164FalseTooshort(t *testing.T) {

	val := "627263"

	res := ValidateE164(val)

	assert.Equal(t, res, false, "The e164 number should  be 10 < x")
}
